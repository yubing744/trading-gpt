# Trading-AI 技术方案文档

## 1. 需求描述

### 1.1 项目背景

Trading-AI 是一个创新型交易机器人项目，基于 [bbgo](https://github.com/c9s/bbgo) 交易引擎和 [langchaingo](https://github.com/tmc/langchaingo) 语言模型集成框架构建。该项目旨在通过自然语言接口降低算法交易的门槛，使不具备编程技能的交易者也能创建和管理自动化交易策略。

### 1.2 核心需求

1. **自然语言交易策略**：用户能够使用自然语言描述交易策略，系统将其转换为可执行的交易逻辑。
2. **多LLM支持**：支持多种大语言模型，包括 OpenAI、Google AI、Claude AI 和 Ollama。
3. **风险管理**：支持设置止损和止盈参数，确保交易策略的风险可控。
4. **策略交互**：用户可以与策略进行对话，调整或了解策略的行为。
5. **技术指标支持**：支持多种技术指标，如MACD、RSI、布林带等。
6. **多交易所支持**：通过bbgo引擎支持多个加密货币交易所。
7. **通知系统**：提供交易更新和系统状态的通知功能。

### 1.3 项目目标

1. 创建一个稳健、可靠的交易系统，能够执行用自然语言定义的策略
2. 支持多个交易所和交易对
3. 提供全面的反馈和监控功能
4. 确保系统安全性，配备适当的风险管理功能

### 1.4 项目范围

- 初期聚焦于加密货币交易
- 保持与bbgo现有功能的兼容性
- 优先考虑可靠性和安全性，而非功能扩展

## 2. 需求分析

### 2.1 角色分析

#### 2.1.1 用户角色

1. **策略创建者**
   - **特点**：了解交易市场和策略，但可能缺乏编程技能
   - **需求**：希望使用自然语言创建和部署交易策略
   - **预期行为**：描述交易思路，设置风险参数，监控策略执行情况

2. **系统管理员**
   - **特点**：负责系统配置和维护
   - **需求**：需要监控系统健康状态，管理API密钥和配置
   - **预期行为**：配置系统参数，监控系统日志，处理异常情况

3. **交易分析师**
   - **特点**：专注于交易策略优化
   - **需求**：分析策略执行效果，调整策略参数
   - **预期行为**：查看交易记录，分析策略表现，提出改进建议

#### 2.1.2 系统角色

1. **Trading Agent**
   - **职责**：处理自然语言策略，生成交易动作
   - **交互对象**：用户，LLM，交易所

2. **Keeper Agent**
   - **职责**：维护系统状态，协调系统组件
   - **交互对象**：Trading Agent，环境系统

3. **LLM Manager**
   - **职责**：管理与LLM的通信，提供模型选择
   - **交互对象**：各类LLM提供商API

4. **环境系统**
   - **职责**：与外部系统接口，如交易所、市场数据源
   - **交互对象**：交易所API，市场数据源

5. **通知系统**
   - **职责**：向用户提供系统状态和交易更新
   - **交互对象**：用户，通知渠道

### 2.2 用例分析

#### 2.2.1 核心用例

1. **创建交易策略**
   - **参与者**：策略创建者，Trading Agent，LLM Manager
   - **前置条件**：用户已配置交易所API和LLM访问权限
   - **主要流程**：
     1. 用户用自然语言描述交易策略
     2. Trading Agent使用LLM解释策略
     3. 系统询问用户可能的澄清问题
     4. 系统创建可执行的交易逻辑
     5. 用户确认策略参数
     6. 系统部署策略
   - **后置条件**：交易策略被创建并准备执行
   - **异常流程**：
     1. LLM无法理解策略：系统请求用户重新描述或提供更多细节
     2. 参数无效：系统提示用户调整参数

2. **执行交易策略**
   - **参与者**：Trading Agent，环境系统
   - **前置条件**：交易策略已创建
   - **主要流程**：
     1. 系统监控市场条件
     2. 当满足策略条件时，生成交易动作
     3. 环境系统执行交易动作
     4. 系统记录交易结果
     5. 通知系统向用户发送交易更新
   - **后置条件**：交易执行完成，结果被记录
   - **异常流程**：
     1. 交易失败：系统记录错误，通知用户
     2. 市场条件变化：系统根据策略调整行为

3. **调整交易策略**
   - **参与者**：策略创建者，Trading Agent，LLM Manager
   - **前置条件**：交易策略已在执行中
   - **主要流程**：
     1. 用户通过聊天描述调整需求
     2. Trading Agent使用LLM理解调整内容
     3. 系统预览调整效果
     4. 用户确认调整
     5. 系统更新策略
   - **后置条件**：交易策略被更新
   - **异常流程**：
     1. 调整不兼容：系统解释问题并提供替代方案
     2. LLM无法理解调整：系统请求更清晰的说明

#### 2.2.2 辅助用例

1. **监控系统状态**
   - **参与者**：系统管理员，Keeper Agent
   - **主要流程**：
     1. Keeper Agent收集系统状态信息
     2. 系统生成状态报告
     3. 通知系统向管理员提供报告

2. **分析交易绩效**
   - **参与者**：交易分析师，Trading Agent
   - **主要流程**：
     1. 用户请求特定时间段的交易分析
     2. 系统收集交易数据
     3. Trading Agent使用LLM生成分析报告
     4. 系统呈现分析结果

## 3. 概要设计

### 3.1 系统架构图

```mermaid
flowchart TD
    User((用户)) <--> Chat["聊天系统"]
    
    subgraph Strategy["Trading-AI Strategy"]
        direction TB
        
        %% 核心事件驱动循环
        Environment["环境系统"] -- "1. 市场变化\n产生事件" --> EventCollector["事件收集器"]
        EventCollector -- "2. 事件触发\n策略评估" --> TradingLogic["交易逻辑"]
        TradingLogic -- "3. 请求决策" --> Agent["AI代理"]
        Agent -- "4. 生成交易决策" --> TradingLogic
        TradingLogic -- "5. 执行交易指令" --> Environment
        
        %% 关键组件
        subgraph Components["核心组件"]
            direction LR
            LLM["LLM系统"]
            Notify["通知系统"]
        end
        
        %% 组件关系
        Agent <--> LLM
        TradingLogic --> Notify
    end
    
    %% 外部连接
    Environment <--> External["外部系统\n(交易所/市场数据)"]
    Notify --> User
    
    %% BBGO集成
    BBGO["BBGO交易引擎"] <--> Strategy
```

系统采用事件驱动的架构设计，主要由以下关键部分组成：

1. **环境系统**：作为交易应用与外部世界(交易所、市场数据)的接口层，负责：
   - 监听市场数据变化
   - 接收交易指令并执行
   - 产生反映外部状态变化的事件

2. **事件收集器**：收集环境系统产生的各类事件，包括：
   - K线数据更新
   - 市场指标变化
   - 持仓状态变更
   - 恐慌贪婪指数变化

3. **交易逻辑**：核心决策层，负责：
   - 根据收集的事件评估当前市场状况
   - 决定何时需要AI代理参与决策
   - 协调交易执行和风险管理

4. **AI代理**：交易决策的智能中枢，负责：
   - 分析市场数据和用户意图
   - 通过自然语言理解用户策略
   - 生成具体交易决策和行动

5. **LLM系统**：AI代理的核心能力提供者，负责：
   - 自然语言理解和生成
   - 策略解析和转换
   - 支持多种LLM模型选择

6. **通知系统**：用户反馈渠道，负责：
   - 交易执行结果通知
   - 系统状态更新
   - 策略执行情况报告

### 3.2 领域模型

领域模型描述了 Trading-AI 系统的核心概念及其相互作用，构成了系统业务逻辑的基础。

```mermaid
graph TD
    subgraph TradingAISystem["Trading-AI 系统"]
        direction LR
        Strategy["交易策略 (用户定义)"]
        Agent["AI 代理"]
        Environment["环境 (市场/交易所)"]
        LLM["大语言模型"]
        Event["事件 (市场变化)"]
        Action["动作 (交易指令)"]
    end

    Environment -- "产生" --> Event
    Event -- "触发" --> Strategy
    Strategy -- "请求决策" --> Agent
    Agent -- "利用" --> LLM
    LLM -- "提供智能" --> Agent
    Agent -- "生成" --> Action
    Action -- "发送至" --> Environment
    Environment -- "执行" --> Action

    UserInterface["用户界面 (聊天)"] <--> Strategy
    UserInterface <--> Agent
```

**核心领域对象说明:**

1.  **交易策略 (Strategy)**:
    *   代表用户定义的交易逻辑和规则，通常通过自然语言输入。
    *   是系统决策的核心依据，消费市场事件并决定何时需要AI代理介入。

2.  **环境 (Environment)**:
    *   代表外部世界，主要是指交易所和市场数据源。
    *   负责监听外部变化（如价格、指标），产生相应的**事件**。
    *   负责执行由**AI代理**生成的**动作**（如买入/卖出）。

3.  **事件 (Event)**:
    *   代表环境中发生的变化，如K线更新、订单成交、指标变化等。
    *   是驱动**交易策略**进行评估和决策的触发器。

4.  **AI 代理 (Agent)**:
    *   系统的智能核心，负责解释**交易策略**和市场**事件**。
    *   利用**大语言模型 (LLM)** 的能力来理解上下文、用户意图，并生成交易**动作**。

5.  **大语言模型 (LLM)**:
    *   为**AI代理**提供自然语言理解、生成和推理能力。
    *   帮助解析用户输入的策略描述，并根据当前市场情况生成决策建议。

6.  **动作 (Action)**:
    *   代表**AI代理**生成的具体指令，通常是交易指令（如买入、卖出、设置止损）。
    *   由**环境**负责执行。

**核心交互流程:**

1.  **环境**监测到市场变化，产生**事件**。
2.  **事件**触发**交易策略**的评估逻辑。
3.  **交易策略**根据当前状态和事件信息，决定何时向**AI代理**请求决策。
4.  **AI代理**结合策略要求、当前事件和历史上下文，利用**LLM**进行分析和推理。
5.  **AI代理**生成具体的交易**动作**。
6.  **动作**被发送到**环境**中执行（例如，在交易所下单）。
7.  用户可以通过用户界面与**交易策略**和**AI代理**进行交互。

这个领域模型更清晰地展示了系统各核心概念之间的业务逻辑关系和信息流转，突出了事件驱动和AI决策的核心特点。

### 3.3 模块图

模块图展示了系统的分层架构和功能模块划分，清晰地呈现各模块间的依赖关系和系统的整体结构。该架构设计遵循了关注点分离原则，确保了系统各部分的独立性和可维护性。

```mermaid
flowchart TD
    subgraph Core[核心模块]
        Config[配置管理]
        Types[类型定义]
        Utils[工具函数]
    end

    subgraph Agents[代理模块]
        AgentI[代理接口]
        Trading[交易代理]
        Keeper[状态代理]
    end
    
    subgraph Env[环境模块]
        EnvI[环境接口]
        EnvSession[环境会话]
        Entity[实体接口]
        Exchange[交易所实体]
        Coze[Coze实体]
        FNG[恐慌贪婪指数实体]
    end
    
    subgraph LLMs[LLM模块]
        LLMManager[LLM管理器]
        OpenAI[OpenAI集成]
        GoogleAI[Google AI集成]
        Anthropic[Claude集成]
        Ollama[Ollama集成]
    end
    
    subgraph Chat[聊天模块]
        ChatProvider[聊天提供者接口]
        ChatSession[聊天会话]
        Feishu[飞书聊天]
    end
    
    subgraph Notify[通知模块]
        NotifyService[通知服务]
        FeishuNotify[飞书通知]
        FeishuHook[飞书Hook通知]
    end
    
    subgraph Prompt[提示词模块]
        PromptTemplate[提示词模板]
        XTemplate[模板引擎]
    end
    
    Core --> Agents
    Core --> Env
    Core --> LLMs
    Core --> Chat
    Core --> Notify
    
    Agents --> LLMs
    Agents --> Env
    Agents --> Prompt
    Agents --> Notify
    
    Chat --> Notify
    Env --> Types
```

模块图说明：

1. **核心模块（Core）**：
   - 为整个系统提供基础设施和共享功能
   - `Config`负责整个系统的配置管理，支持环境变量和配置文件的加载
   - `Types`定义了系统中使用的核心数据类型和接口
   - `Utils`提供了各类通用工具函数，如字符串处理、风险管理计算等

2. **代理模块（Agents）**：
   - 实现了系统的智能决策能力
   - `AgentI`定义了代理的接口规范
   - `Trading`实现了交易策略的解析和执行
   - `Keeper`负责系统状态的维护和监控

3. **环境模块（Env）**：
   - 提供了与外部系统交互的抽象层
   - `EnvI`和`Entity`定义了环境和实体的接口规范
   - `EnvSession`管理环境的会话状态
   - 包含多种实体实现，如`Exchange`、`Coze`和`FNG`等

4. **LLM模块（LLMs）**：
   - 提供了大语言模型的集成和管理
   - `LLMManager`统一管理不同的LLM提供商
   - 支持多种LLM实现，包括OpenAI、GoogleAI、Claude和Ollama
   - 提供了统一的模型访问接口

5. **聊天模块（Chat）**：
   - 实现了用户与系统的交互界面
   - `ChatProvider`定义了聊天提供者的接口规范
   - `ChatSession`管理用户的会话状态
   - 目前支持飞书作为聊天渠道

6. **通知模块（Notify）**：
   - 负责向用户推送系统状态和交易更新
   - `NotifyService`提供了统一的通知服务
   - 支持多种通知渠道，如飞书和飞书Hook

7. **提示词模块（Prompt）**：
   - 管理与LLM交互的提示词模板
   - `PromptTemplate`提供了标准化的提示词结构
   - `XTemplate`实现了模板渲染引擎

8. **模块依赖关系**：
   - 核心模块被所有其他模块依赖，提供基础能力
   - 代理模块依赖LLM、环境、提示词和通知模块，体现了其核心协调角色
   - 聊天模块依赖通知模块实现消息推送
   - 环境模块依赖核心类型定义，确保数据结构一致性

### 3.4 技术挑战

1. **自然语言理解的准确性**
   - **挑战**：确保LLM准确理解用户描述的交易策略
   - **解决方案**：
     - 使用结构化提示词模板
     - 实施澄清机制，在不确定时向用户提问
     - 支持多种LLM以提高理解的可靠性

2. **交易策略执行的安全性**
   - **挑战**：防止生成和执行有风险的交易策略
   - **解决方案**：
     - 实施风险管理约束（如止损和止盈）
     - 在执行交易前进行安全检查
     - 设置交易限额防止过度交易

3. **系统状态管理**
   - **挑战**：在分布式环境中维护一致的系统状态
   - **解决方案**：
     - 使用Keeper Agent监控和协调系统组件
     - 实施事件驱动架构，确保状态变更的一致传播
     - 使用环境会话管理状态和生命周期

4. **交易所API的可靠集成**
   - **挑战**：处理交易所API的限制和特性差异
   - **解决方案**：
     - 利用bbgo的交易引擎抽象不同交易所API
     - 实施重试机制和错误处理
     - 设计环境实体层以标准化交互

5. **LLM上下文限制**
   - **挑战**：处理复杂策略描述可能超出LLM上下文限制
   - **解决方案**：
     - 实现策略分解机制，将复杂策略分解为可管理的部分
     - 使用模板优化提示词结构
     - 探索不同LLM提供商的上下文容量权衡

## 4. 详细设计 (Detailed Design)

### 4.1 交易反思与记忆功能 (Trading Reflection and Memory Feature)

#### 4.1.1 功能概述

为了让 Trading-AI 具备学习和迭代能力，引入交易反思与记忆功能。当一个由特定策略管理的交易仓位关闭时，系统（由 `jarvis.go` 协调）将自动触发反思流程。该流程利用 LLM 分析本次交易的关键信息（入场、出场、盈亏、市场状况等），生成反思性总结。这个总结作为“记忆”被存储在 Markdown 文件中。在后续的交易决策周期中，系统（`jarvis.go`）会加载相关的历史“记忆”，并将其整合到发送给 `Trading Agent` 的上下文或提示词中，以期指导 LLM 做出更明智的决策。

#### 4.1.2 触发机制

1.  **事件源**: `pkg/env/exchange/exchange_entity.go` 中的交易所实体 (`Exchange Entity`) 负责监控仓位状态。
2.  **触发条件**: 当检测到与某个策略实例关联的仓位完全关闭时（无论是通过止盈、止损还是手动平仓），`Exchange Entity` 将触发一个新的事件。
3.  **事件类型**: 定义一个新的事件类型 `PositionClosedEvent` 在 `pkg/types/event.go` 中。该事件应包含必要的信息，例如：
    *   `StrategyID`: 关联的策略标识符。
    *   `Symbol`: 交易对。
    *   `EntryPrice`: 入场价格。
    *   `ExitPrice`: 出场价格。
    *   `Quantity`: 数量。
    *   `ProfitAndLoss`: 盈亏金额。
    *   `CloseReason`: 平仓原因（\"TakeProfit\", \"StopLoss\", \"Manual\", \"Liquidation\" 等）。
    *   `Timestamp`: 平仓时间戳。
    *   `RelatedMarketData`: （可选）平仓前后的关键市场数据快照（如K线、指标值）。
4.  **事件消费**: `pkg/jarvis.go` 中的核心事件循环将监听并处理 `PositionClosedEvent`。

#### 4.1.3 反思生成流程

```mermaid
sequenceDiagram
    participant ExchangeEntity as 环境系统 (交易所实体)
    participant Jarvis as Jarvis (核心协调器)
    participant TradingAgent as 交易代理
    participant LLMManager as LLM 管理器
    participant FileSystem as 文件系统 (Memory Bank)

    ExchangeEntity->>Jarvis: 触发 PositionClosedEvent(data)
    Jarvis->>Jarvis: 接收事件，提取 data
    Jarvis->>TradingAgent: 请求与 data.StrategyID 相关的上下文 (如初始策略描述)
    TradingAgent-->>Jarvis: 返回策略上下文
    Jarvis->>LLMManager: 构建反思提示词 (包含交易 data 和策略上下文)
    Note right of Jarvis: 提示词示例: \"反思以下已关闭的交易:\\n策略描述: [原始策略]\\n交易对: [Symbol]\\n入场价: [EntryPrice]\\n出场价: [ExitPrice]\\n盈亏: [PnL]\\n平仓原因: [CloseReason]\\n市场快照: [MarketData]\\n\\n总结本次交易的关键教训，并提出未来在类似情况下改进策略的建议。\"
    LLMManager->>LLMManager: 调用 LLM 生成反思内容
    LLMManager-->>Jarvis: 返回 LLM 生成的反思文本 (reflectionText)
    Jarvis->>FileSystem: 将 reflectionText 保存到 Markdown 文件
    Note right of Jarvis: 文件路径: memory-bank/reflections/<StrategyID>_<Timestamp>.md
```

1.  **事件处理**: `pkg/jarvis.go` 中的核心逻辑负责处理 `PositionClosedEvent`。
2.  **上下文收集**: `Jarvis` 接收事件数据，并可能向相关的 `Trading Agent` 实例请求额外的上下文信息（如原始策略描述）。
3.  **提示词构建**: `Jarvis` 使用 `pkg/prompt/prompt.go` 中的模板功能，构建一个专门用于交易反思的提示词，将收集到的上下文信息结构化地输入给 LLM。
4.  **LLM 调用**: `Jarvis` 通过 `LLM Manager` 调用配置的 LLM 服务，生成反思内容。
5.  **记忆存储**:\n    *   `Jarvis` 将 LLM 返回的反思文本保存到一个新的 Markdown 文件中。\n    *   **存储位置**: 在 `memory-bank/` 目录下创建一个新的子目录 `reflections/`。\n    *   **文件命名**: 采用 `<StrategyID>_<UnixTimestamp>.md` 的格式，确保文件唯一且易于关联和排序。例如：`macd_cross_strategy_1713580800.md`。\n    *   **文件内容**: 文件主体为 LLM 生成的反思文本。可以在文件头部添加一些元数据（如交易对、盈亏、时间戳）作为注释或 Front Matter。

#### 4.1.4 记忆检索与整合

```mermaid
sequenceDiagram
    participant Jarvis as Jarvis (核心协调器)
    participant FileSystem as 文件系统 (Memory Bank)
    participant TradingAgent as 交易代理
    participant LLMManager as LLM 管理器

    Jarvis->>Jarvis: 开始新的决策周期 (如收到市场事件)
    Jarvis->>FileSystem: 检查 memory-bank/reflections/ 目录下是否存在与 StrategyID 匹配的文件
    FileSystem-->>Jarvis: 返回匹配的文件列表
    alt 存在记忆文件
        Jarvis->>Jarvis: 读取最新的一个或多个记忆文件内容 (memoryContent)
        Jarvis->>Jarvis: 准备基础决策上下文 (baseContext)
        Jarvis->>TradingAgent: 传递包含记忆的上下文 (baseContext + memoryContent) 或构建包含记忆的提示词
        Note right of Jarvis: 整合方式取决于 Agent 设计，\n可以将记忆作为独立信息传递，\n或直接整合入提示词:\n\"...[基础决策上下文]...\\n\\n请参考以下过去的交易反思:\\n- [上次反思内容]\\n\\n基于以上信息和当前市场状况，请做出交易决策。\"
    else 无记忆文件
        Jarvis->>TradingAgent: 传递基础决策上下文 (baseContext) 或构建基础提示词
    end
    TradingAgent->>LLMManager: （使用收到的上下文/提示词）调用 LLM 获取决策
    LLMManager->>LLMManager: 调用 LLM
    LLMManager-->>TradingAgent: 返回 LLM 决策结果
    TradingAgent-->>Jarvis: 返回决策动作
```

1.  **检索时机**: 在 `pkg/jarvis.go` 的核心决策循环中，为特定策略准备调用 `Trading Agent` 之前。
2.  **检索逻辑**: `Jarvis` 检查 `memory-bank/reflections/` 目录下与该策略 `StrategyID` 匹配的所有 `.md` 文件。
3.  **内容选择**:\n    *   默认策略：读取最新创建的一个反思文件。\n    *   可选策略（未来扩展）：读取最近的 N 个反思文件，或者对多个反思进行摘要后再整合。
4.  **上下文/提示词整合**:\n    *   `Jarvis` 将检索到的反思内容 (`memoryContent`) 与当前的基础决策上下文 (`baseContext`) 结合。\n    *   整合后的信息被传递给 `Trading Agent`。这可以通过两种方式实现：\n        *   将 `memoryContent` 作为独立的上下文信息传递给 `Trading Agent`，由 Agent 负责将其整合到最终的 LLM 提示词中。\n        *   `Jarvis` 直接构建包含 `memoryContent` 的完整提示词，然后将此提示词传递给 `Trading Agent`，Agent 直接使用该提示词调用 LLM。\n    *   选择哪种方式取决于 `Trading Agent` 的接口设计和职责划分。
5.  **Agent 调用**: `Jarvis` 调用 `Trading Agent` 的决策方法，传递包含（或准备好的）历史记忆的上下文/提示词。

#### 4.1.5 组件和配置变更

*   **`pkg/jarvis.go`**:
    *   增加处理 `PositionClosedEvent` 的逻辑。
    *   实现反思上下文收集（可能需要与 Agent 交互）、LLM 调用、文件保存逻辑。
    *   实现记忆文件检索和加载逻辑。
    *   修改调用 `Trading Agent` 的逻辑，以传递整合了记忆的上下文或提示词。
*   **`pkg/env/exchange/exchange_entity.go`**: 增加仓位关闭检测逻辑，触发 `PositionClosedEvent`。
*   **`pkg/types/event.go`**: 定义 `PositionClosedEvent` 结构体。
*   **`pkg/agents/trading/trading_agent.go`**:
    *   可能需要调整接口，以接收包含历史记忆的上下文或提示词。
    *   可能需要提供方法供 `Jarvis` 获取策略相关的上下文信息。
*   **`pkg/prompt/prompt.go`**: 可能需要添加或修改提示词模板，以支持反思生成和记忆整合（无论是在 Jarvis 还是 Agent 中构建）。
*   **`pkg/config/config.go` (及相关子配置)**:\n    *   增加配置项以启用/禁用此记忆功能 (`memory.enabled: bool`)。\n    *   增加配置项指定记忆文件的存储路径 (`memory.path: string`)，默认为 `\"memory-bank/reflections\"`。\n    *   增加配置项控制加载记忆的数量或策略 (`memory.retrieval_count: int`)。

#### 4.1.6 技术考虑

*   **文件管理**: 需要考虑旧记忆文件的清理策略，避免无限增长。可以按时间或数量限制。
*   **性能**: 文件 I/O 操作应异步执行或在非关键路径上执行，避免阻塞交易决策流程。
*   **提示词长度**: 大量历史记忆可能导致提示词过长，超出 LLM 的上下文限制。需要实现截断、摘要或选择性加载机制。
*   **反思质量**: LLM 生成的反思质量直接影响其效果。需要精心设计反思提示词，并可能需要对结果进行后处理或筛选。
*   **错误处理**: 文件读写、LLM 调用失败等情况需要健壮的错误处理。
*   **并发**: 如果多个策略实例并发运行，需要确保文件读写和状态更新是并发安全的。
