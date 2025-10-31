# 限价单功能说明

## 概述

本功能为 trading-gpt 添加了限价单支持，允许策略使用限价单来获得更好的入场价格。系统采用**周期自动清理**机制，在每个决策周期开始时自动取消未成交的限价单，确保AI每次都从干净的状态开始决策。

**核心特性：**
- 支持限价单和市价单两种订单类型
- 支持价格表达式（如 `last_close * 0.995`）
- 自动清理未成交限价单，防止堆积
- AI无需管理订单，保持决策简单

## 使用方法

### 基础用法

在交易命令中添加 `order_type` 和 `limit_price` 参数：

```json
{
  "action": {
    "name": "open_long_position",
    "args": {
      "stop_loss_trigger_price": "2800",
      "take_profit_trigger_price": "3200",
      "order_type": "limit",
      "limit_price": "2950"
    }
  }
}
```

### 价格表达式

支持使用表达式动态计算限价：

```json
{
  "action": {
    "name": "open_long_position",
    "args": {
      "order_type": "limit",
      "limit_price": "last_close * 0.995"
    }
  }
}
```

**可用变量：**
- `last_close` / `close` - 最新收盘价
- `last_high` - 最新最高价
- `last_low` - 最新最低价
- `last_open` - 最新开盘价
- `last_volume` - 最新成交量
- `prev_close` - 前一根K线收盘价

**表达式示例：**
```javascript
// 低于当前价0.5%买入
"last_close * 0.995"

// 高于当前价0.5%卖出
"last_close * 1.005"

// 在最低价上方1%买入
"last_low * 1.01"

// 复杂表达式
"(last_high + last_low) / 2"
```

### 高级参数

```json
{
  "action": {
    "name": "open_long_position",
    "args": {
      "order_type": "limit",
      "limit_price": "last_close * 0.995",
      "time_in_force": "GTC",    // GTC | IOC | FOK (默认GTC)
      "post_only": "false"        // 接受但不应用（bbgo限制）
    }
  }
}
```

### 支持的命令

所有交易命令都支持限价单参数：
- `open_long_position`
- `open_short_position`
- `update_position`

## 技术架构

### 工作流程

```
每个决策周期 (Kline更新时)：
┌─────────────────────────────────────────┐
│ 1. Kline更新                             │
│ 2. 【自动清理】取消所有未成交限价单         │
│ 3. 更新指标数据                           │
│ 4. AI分析市场                            │
│ 5. AI决策（市价单 or 限价单）             │
│ 6. 执行 → 等待下一个周期                  │
└─────────────────────────────────────────┘
```

### 核心组件

#### 1. 价格表达式解析器

**文件：** `pkg/utils/price.go`

```go
func ParsePrice(
    vm *goja.Runtime,
    klines *types.KLineWindow,
    closePrice fixedpoint.Value,
    expr string,
) (*fixedpoint.Value, error)
```

**功能：**
- 使用JavaScript VM解析表达式
- 设置K线数据上下文变量
- 返回计算后的价格

#### 2. 自动清理机制

**文件：** `pkg/env/exchange/exchange_entity.go`

```go
func (ent *ExchangeEntity) cleanupLimitOrders(ctx context.Context)
```

**清理规则：**
- 只清理限价单类型（`OrderTypeLimit`, `OrderTypeLimitMaker`）
- 保留止损止盈单（由position模块管理）
- 清理失败不阻塞流程，仅记录日志

**调用位置：**
```go
session.MarketDataStream.OnKLineClosed(types.KLineWith(ent.symbol, ent.interval, func(kline types.KLine) {
    // ... 更新K线和指标 ...

    // 在发送update_finish事件前清理
    ent.cleanupLimitOrders(ctx)

    // 触发AI决策
    ent.emitEvent(ch, ttypes.NewEvent("update_finish", nil))
}))
```

#### 3. 参数处理

**HandleCommand 流程：**

```go
// 1. 解析order_type
if orderType, ok := args["order_type"]; ok && orderType != "" {
    opts = append(opts, &OrderTypeOpt{
        Type: types.OrderType(strings.ToUpper(orderType)),
    })
}

// 2. 解析limit_price（支持表达式）
if limitPrice, ok := args["limit_price"]; ok && limitPrice != "" {
    price, err := utils.ParsePrice(ent.vm, ent.KLineWindow, closePrice, limitPrice)
    if err != nil {
        return errors.Wrapf(err, "invalid limit_price: %s", limitPrice)
    }
    opts = append(opts, &LimitPriceOpt{Value: *price})
}

// 3. 参数验证
if ot, ok := args["order_type"]; ok && strings.ToUpper(ot) == "LIMIT" {
    if lp, ok := args["limit_price"]; !ok || lp == "" {
        return errors.New("limit_price is required when order_type=limit")
    }
}
```

#### 4. 选项结构体

```go
type OrderTypeOpt struct {
    Type types.OrderType
}

type LimitPriceOpt struct {
    Value fixedpoint.Value
}

type TimeInForceOpt struct {
    Value types.TimeInForce
}

type PostOnlyOpt struct {
    Enabled bool
}
```

## 参数说明

### order_type

**类型：** string
**可选值：** `market` | `limit`
**默认值：** `market`
**说明：** 订单类型

### limit_price

**类型：** string
**必需条件：** 当 `order_type=limit` 时必须提供
**格式：** 绝对值或表达式
**说明：** 限价价格

**示例：**
- 绝对值：`"3000"`
- 表达式：`"last_close * 0.995"`

### time_in_force

**类型：** string
**可选值：** `GTC` | `IOC` | `FOK`
**默认值：** `GTC`
**说明：** 订单有效期

- `GTC` (Good Till Cancel): 一直有效直到成交或取消
- `IOC` (Immediate Or Cancel): 立即成交否则取消
- `FOK` (Fill Or Kill): 全部成交否则取消

### post_only

**类型：** string
**可选值：** `true` | `false`
**默认值：** `false`
**说明：** 仅做maker（当前bbgo不支持，参数被接受但不应用）

## 设计决策

### 为什么自动清理？

**问题：** 如果不清理，未成交的限价单会不断堆积

**方案对比：**

| 方案 | AI复杂度 | 挂单堆积 | 实现复杂度 |
|------|---------|---------|-----------|
| 不清理 | 简单 | ❌ 会堆积 | 简单 |
| 手动清理 | ❌ 复杂 | ✅ 不堆积 | 复杂 |
| **自动清理** | ✅ 简单 | ✅ 不堆积 | 简单 |

**选择自动清理的原因：**
1. AI无需学习订单管理
2. 符合周期性决策模式
3. 每次都是全新决策，更符合AI思维
4. 实现简单，维护成本低

### 为什么只清理限价单？

**清理规则：**
- ✅ 清理：`OrderTypeLimit`, `OrderTypeLimitMaker`
- ❌ 保留：止损止盈单（由position模块管理）

**原因：**
- 止损止盈是风险管理的一部分，应该持续有效
- 限价单是"主动挂出"的订单，与周期决策相关
- 避免误清理重要的风控单

## 测试

### 单元测试

**文件：** `pkg/utils/price_test.go`

```bash
go test -v ./pkg/utils -run TestParsePrice
```

**测试覆盖：**
- ✅ 绝对值解析
- ✅ 表达式计算
- ✅ K线数据上下文
- ✅ 边界情况处理
- ✅ 错误处理

### 手动测试

1. **创建限价单**
```bash
# 设置策略使用限价单
# 观察日志确认订单创建
```

2. **验证自动清理**
```bash
# 观察下一个周期开始时的日志
# 应该看到：auto cleanup limit orders before new decision cycle
```

3. **测试表达式**
```bash
# 使用不同表达式创建限价单
# 验证价格计算正确
```

## 常见问题

### Q: 限价单不成交怎么办？

A: 下一个决策周期开始时会自动取消，AI会重新决策是否下单。

### Q: 如何让限价单持续有效？

A: 当前设计不支持。如需此功能，建议使用更专业的交易工具。

### Q: post_only参数为什么不生效？

A: bbgo的SubmitOrder结构体不支持PostOnly字段。参数被接受但不会应用到订单。

### Q: 可以手动管理限价单吗？

A: 当前不支持。系统会在每个周期自动清理。如需此功能，需要扩展实现（参考已删除的方案C）。

### Q: 表达式支持哪些运算？

A: 支持JavaScript的所有数学运算，包括：
- 四则运算：`+`, `-`, `*`, `/`
- 括号：`()`
- 函数：`Math.min()`, `Math.max()` 等

## 相关文件

- `pkg/env/exchange/exchange_entity.go` - 主要逻辑
- `pkg/utils/price.go` - 价格解析器
- `pkg/utils/price_test.go` - 单元测试
- Issue: [#58](https://github.com/yubing744/trading-gpt/issues/58)

## 更新历史

- **2025-01-22**: 初始实现，采用方案D（创建+自动清理）
