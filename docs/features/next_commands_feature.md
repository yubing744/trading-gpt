# Next-Cycle Command Persistence Feature

## Overview

The Next-Cycle Command Persistence feature enables trading strategies to schedule commands for execution in the next decision cycle. This allows the AI to proactively collect indicator data, execute workflows, or perform auxiliary tasks before making the next trading decision.

## Key Benefits

- **Proactive Data Collection**: Schedule indicator collection commands instead of relying on manual configuration
- **Workflow Integration**: Execute external workflows (e.g., Coze sentiment analysis) as part of decision flow
- **Persistence**: Commands survive strategy restarts and are automatically retried on failure
- **Validation**: Commands are validated against entity capabilities before execution

## How It Works

### Architecture

```
Decision Cycle N:
  AI outputs Result with next_commands
  → Commands saved to memory-bank/commands.json

Decision Cycle N+1:
  Load pending commands from file
  → Execute each command via entity.HandleCommand()
  → Results emitted as events/indicators
  → AI makes decision with full context (market data + command results)
```

### Command Flow

1. **AI Schedules Commands**: During decision cycle N, AI outputs `next_commands` in JSON response
2. **Persistence**: Commands are saved to `memory-bank/commands.json` with status tracking
3. **Pre-Execution**: At the start of cycle N+1, pending commands are loaded and executed
4. **Result Integration**: Command results are available as events/indicators for AI analysis
5. **Status Update**: Completed/failed commands are archived for audit

## Usage

### AI Output Format

The AI can include `next_commands` in its JSON response:

```json
{
  "thoughts": {
    "plan": "...",
    "analyze": "...",
    "detail": "...",
    "reflection": "...",
    "speak": "..."
  },
  "action": {
    "name": "exchange.hold_position",
    "args": {}
  },
  "next_commands": [
    {
      "entity_id": "coze",
      "command_name": "market_sentiment_workflow",
      "args": {
        "workflow_id": "123456",
        "symbol": "BTCUSDT"
      }
    }
  ]
}
```

### How AI Determines entity_id

**Command Format in Prompt:**
All available commands are presented to the AI in the format `"entity_id.command_name"`, for example:
- `exchange.open_long_position`
- `coze.workflow_sentiment`
- `fng.refresh_index`
- `twitterapi.search_tweets`

**Extraction Process:**
The AI needs to split the command name at the dot (`.`) to determine:
1. **entity_id**: The part before the dot (e.g., `"exchange"`, `"coze"`, `"fng"`, `"twitterapi"`)
2. **command_name**: The part after the dot (e.g., `"open_long_position"`, `"refresh_index"`)

**Available Entities:**
- **exchange**: Trading operations
- **coze**: Coze workflow/bot execution
- **fng**: Fear & Greed Index data
- **twitterapi**: Twitter search functionality

**Examples:**
- `"exchange.open_long_position"` → `entity_id="exchange"`, `command_name="open_long_position"`
- `"fng.get_historical_index"` → `entity_id="fng"`, `command_name="get_historical_index"`
- `"twitterapi.search_tweets"` → `entity_id="twitterapi"`, `command_name="search_tweets"`

This information is clearly explained in the AI's prompt to ensure correct command scheduling.

### Command Structure

Each command in `next_commands` must include:

- **entity_id**: Target entity (e.g., "coze", "exchange")
- **command_name**: Command name matching entity's `Actions()` descriptor
- **args**: Map of string parameters required by the command

### Entity Command Declaration

All entities declare their available commands via the `Actions()` method. Here are examples from each entity:

**Coze Entity:**
```go
func (e *CozeEntity) Actions() []*types.ActionDesc {
    return []*types.ActionDesc{
        {
            Name:        "market_sentiment_workflow",
            Description: "Analyze market sentiment from social media",
            Args: []types.ArgmentDesc{
                {Name: "workflow_id", Description: "Coze workflow ID"},
                {Name: "symbol", Description: "Trading pair symbol"},
            },
        },
    }
}
```

**FNG Entity:**
```go
func (entity *FearAndGreedEntity) Actions() []*types.ActionDesc {
    return []*types.ActionDesc{
        {
            Name:        "refresh_index",
            Description: "Manually refresh the current Fear & Greed Index",
        },
        {
            Name:        "get_historical_index",
            Description: "Get historical Fear & Greed Index data",
            Args: []types.ArgmentDesc{
                {Name: "limit", Description: "Number of historical data points (default: 7, max: 30)"},
            },
        },
    }
}
```

**TwitterAPI Entity:**
```go
func (e *TwitterAPIEntity) Actions() []*types.ActionDesc {
    return []*types.ActionDesc{
        {
            Name:        "search_tweets",
            Description: "Search Twitter for tweets matching a query",
            Args: []types.ArgmentDesc{
                {Name: "query", Description: "Search query (required)"},
                {Name: "query_type", Description: "Query type: Top|Latest (default: Top)"},
                {Name: "max_results", Description: "Maximum results (default: 10, max: 100)"},
            },
        },
    }
}
```

**Exchange Entity:**
The Exchange entity has comprehensive command support for trading operations and technical indicator queries:

**Trading Commands:**
- `open_long_position` - Open long position with various order types (market/limit)
- `open_short_position` - Open short position with various order types (market/limit)
- `close_position` - Close position (full or partial)
- `update_position` - Update position stop-loss/take-profit
- `no_action` - No action to be taken

**Dynamic Indicator Query:**
```go
{
    Name:        "get_indicator",
    Description: "Dynamically calculate and retrieve technical indicator data for any timeframe",
    Args: []types.ArgmentDesc{
        {Name: "type", Description: "Indicator type (required): RSI, BOLL, SMA, EWMA, VWMA, ATR, ATRP, VR, EMV"},
        {Name: "interval", Description: "Time interval (default: 5m): 1m, 5m, 15m, 30m, 1h, 4h, 1d, etc."},
        {Name: "window_size", Description: "Window size for calculation (default varies by indicator type)"},
        {Name: "band_width", Description: "Band width for BOLL indicator (default: 2.0)"},
    },
}
```

## Configuration

### Enable Command System

Add to `bbgo.yaml`:

```yaml
exchangeStrategies:
- on: okex
  jarvis:
    # ... existing config ...

    commands:
      enabled: true
      command_path: "memory-bank/commands.json"  # optional, default shown
```

### Configuration Options

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `enabled` | bool | false | Enable/disable command system |
| `command_path` | string | `memory-bank/commands.json` | Path to command persistence file |

## File Format

Commands are persisted in JSON format:

```json
{
  "pending": [
    {
      "id": "uuid-1",
      "entity_id": "coze",
      "command_name": "market_sentiment_workflow",
      "args": {
        "workflow_id": "123456",
        "symbol": "BTCUSDT"
      },
      "status": "pending",
      "retry_count": 0,
      "max_retries": 1,
      "created_at": "2025-01-23T10:00:00Z",
      "updated_at": "2025-01-23T10:00:00Z"
    }
  ],
  "completed": [],
  "failed": []
}
```

## Error Handling

### Command Validation

Before executing, commands are validated:

1. **Entity Exists**: Checks if target entity is registered
2. **Command Supported**: Verifies command is in entity's `Actions()` list

Invalid commands are skipped with a warning message.

### Execution Errors

- **Timeout**: Commands have a 30-second execution timeout
- **Retry Logic**: Failed commands are retried once (configurable via `max_retries`)
- **Permanent Failure**: After max retries, commands are moved to `failed` list

### Status Transitions

```
pending → (execute) → completed
        ↓ (error)
        failed → (retry) → completed
              ↓ (max retries)
              permanently failed (archived)
```

## Dynamic Technical Indicator Queries

### Overview

The Exchange entity's `get_indicator` command enables AI to dynamically request technical indicator calculations for **any timeframe and parameter combination**, without being limited to pre-configured indicators. This provides unprecedented flexibility for multi-timeframe analysis and adaptive strategy development.

### Key Features

- **No Configuration Required**: Calculate indicators on-demand without pre-defining them in config files
- **Any Timeframe**: Support for all exchange timeframes (1m, 5m, 15m, 30m, 1h, 4h, 1d, etc.)
- **Flexible Parameters**: Adjust window sizes, band widths, and other parameters dynamically
- **Multi-Timeframe Analysis**: Request same indicator across different timeframes for trend confirmation
- **Parameter Optimization**: Test different parameter combinations to find optimal settings

### Supported Indicators

| Indicator | Type Code | Default Window | Parameters |
|-----------|-----------|----------------|------------|
| Relative Strength Index | `RSI` | 14 | `window_size` |
| Bollinger Bands | `BOLL` | 20 | `window_size`, `band_width` (default: 2.0) |
| Simple Moving Average | `SMA` | 20 | `window_size` |
| Exponential Moving Average | `EWMA` | 20 | `window_size` |
| Volume Weighted MA | `VWMA` | 20 | `window_size` |
| Average True Range | `ATR` | 14 | `window_size` |
| ATR Percentage | `ATRP` | 14 | `window_size` |
| Volume Ratio | `VR` | 14 | `window_size` |
| Ease of Movement | `EMV` | 14 | `window_size` |

### Command Structure

```json
{
  "entity_id": "exchange",
  "command_name": "get_indicator",
  "args": {
    "type": "RSI",           // Required: indicator type
    "interval": "15m",       // Optional: default 5m
    "window_size": "14",     // Optional: default varies by type
    "band_width": "2.0"      // Optional: BOLL only, default 2.0
  }
}
```

### Usage Scenarios

#### Scenario 1: Multi-Timeframe Trend Confirmation

AI wants to confirm a trend by checking RSI across short, medium, and long timeframes:

```json
{
  "action": {"name": "exchange.hold_position", "args": {}},
  "next_commands": [
    {
      "entity_id": "exchange",
      "command_name": "get_indicator",
      "args": {
        "type": "RSI",
        "interval": "5m",
        "window_size": "14"
      }
    },
    {
      "entity_id": "exchange",
      "command_name": "get_indicator",
      "args": {
        "type": "RSI",
        "interval": "1h",
        "window_size": "14"
      }
    },
    {
      "entity_id": "exchange",
      "command_name": "get_indicator",
      "args": {
        "type": "RSI",
        "interval": "4h",
        "window_size": "14"
      }
    }
  ]
}
```

#### Scenario 2: Adaptive Parameter Testing

AI detects high volatility and wants tighter Bollinger Bands:

```json
{
  "action": {"name": "exchange.hold_position", "args": {}},
  "next_commands": [
    {
      "entity_id": "exchange",
      "command_name": "get_indicator",
      "args": {
        "type": "BOLL",
        "interval": "15m",
        "window_size": "20",
        "band_width": "2.0"
      }
    },
    {
      "entity_id": "exchange",
      "command_name": "get_indicator",
      "args": {
        "type": "BOLL",
        "interval": "15m",
        "window_size": "20",
        "band_width": "3.0"
      }
    }
  ]
}
```

#### Scenario 3: Fast vs Slow Moving Average Crossover

AI wants to check for golden cross/death cross signals:

```json
{
  "action": {"name": "exchange.hold_position", "args": {}},
  "next_commands": [
    {
      "entity_id": "exchange",
      "command_name": "get_indicator",
      "args": {
        "type": "SMA",
        "interval": "1h",
        "window_size": "10"
      }
    },
    {
      "entity_id": "exchange",
      "command_name": "get_indicator",
      "args": {
        "type": "SMA",
        "interval": "1h",
        "window_size": "50"
      }
    }
  ]
}
```

#### Scenario 4: Volatility Analysis

AI wants to assess market volatility before opening position:

```json
{
  "action": {"name": "exchange.hold_position", "args": {}},
  "next_commands": [
    {
      "entity_id": "exchange",
      "command_name": "get_indicator",
      "args": {
        "type": "ATR",
        "interval": "1h",
        "window_size": "14"
      }
    },
    {
      "entity_id": "exchange",
      "command_name": "get_indicator",
      "args": {
        "type": "ATRP",
        "interval": "1h",
        "window_size": "14"
      }
    },
    {
      "entity_id": "exchange",
      "command_name": "get_indicator",
      "args": {
        "type": "BOLL",
        "interval": "1h"
      }
    }
  ]
}
```

#### Scenario 5: Sensitive vs Conservative Indicators

AI wants to compare fast-reacting vs slow-reacting indicators:

```json
{
  "action": {"name": "exchange.hold_position", "args": {}},
  "next_commands": [
    {
      "entity_id": "exchange",
      "command_name": "get_indicator",
      "args": {
        "type": "RSI",
        "interval": "5m",
        "window_size": "7"
      }
    },
    {
      "entity_id": "exchange",
      "command_name": "get_indicator",
      "args": {
        "type": "RSI",
        "interval": "5m",
        "window_size": "21"
      }
    }
  ]
}
```

### How It Works

1. **Request**: AI schedules `get_indicator` command with desired parameters
2. **Data Retrieval**: System fetches historical kline data for specified interval from MarketDataStore
3. **Calculation**: Indicator is dynamically created and calculated using BBGO's StandardIndicatorSet
4. **Event Emission**: Calculated indicator data is emitted as `indicator_changed` event
5. **AI Analysis**: AI receives indicator data in next cycle and makes informed decision

### Benefits vs Pre-Configured Indicators

| Aspect | Pre-Configured | Dynamic `get_indicator` |
|--------|----------------|-------------------------|
| **Flexibility** | ❌ Limited to config | ✅ Unlimited combinations |
| **Setup** | ❌ Must edit config file | ✅ Zero configuration |
| **Timeframes** | ❌ Few pre-defined | ✅ All available timeframes |
| **Parameters** | ❌ Fixed at startup | ✅ Adjust on-the-fly |
| **Multi-timeframe** | ❌ Need many configs | ✅ Request as needed |
| **Experimentation** | ❌ Requires restart | ✅ Instant testing |

### Performance Considerations

- **Caching**: Indicator calculations are performed on-demand; BBGO's MarketDataStore provides kline caching
- **Computation**: Indicators use historical data already stored in memory
- **Latency**: Typical calculation time: 10-50ms depending on window size and data availability
- **Resource Usage**: Minimal; indicators are garbage collected after use

### Best Practices

1. **Start with Defaults**: Use default parameters first, then adjust based on results
2. **Limit Requests**: Request only indicators needed for current decision (3-5 per cycle recommended)
3. **Progressive Refinement**: Start with coarse timeframes (1h, 4h), drill down to finer (5m, 15m) if needed
4. **Combine Indicators**: Use multiple indicator types for confirmation (e.g., RSI + BOLL + ATR)
5. **Document Findings**: Have AI log which parameter combinations work best for different market conditions

### Error Handling

The command will fail if:
- **Invalid Indicator Type**: Type not in supported list
- **No Data Available**: Exchange hasn't provided data for requested interval
- **Invalid Parameters**: Non-numeric window_size or band_width

Error messages clearly indicate the issue for AI to adjust the request.

## Example Use Cases

### 1. Sentiment Analysis with Coze Workflow

```json
{
  "action": {"name": "exchange.hold_position", "args": {}},
  "next_commands": [
    {
      "entity_id": "coze",
      "command_name": "social_sentiment",
      "args": {
        "workflow_id": "sentiment_workflow_id",
        "symbol": "BTCUSDT"
      }
    }
  ]
}
```

### 2. Fear & Greed Index Analysis

```json
{
  "action": {"name": "exchange.hold_position", "args": {}},
  "next_commands": [
    {
      "entity_id": "fng",
      "command_name": "get_historical_index",
      "args": {
        "limit": "14"
      }
    }
  ]
}
```

### 3. Twitter Sentiment Analysis

```json
{
  "action": {"name": "exchange.hold_position", "args": {}},
  "next_commands": [
    {
      "entity_id": "twitterapi",
      "command_name": "search_tweets",
      "args": {
        "query": "Bitcoin OR BTC lang:en",
        "query_type": "Top",
        "max_results": "20"
      }
    }
  ]
}
```

### 4. Multi-Source Data Collection

```json
{
  "action": {"name": "exchange.hold_position", "args": {}},
  "next_commands": [
    {
      "entity_id": "fng",
      "command_name": "refresh_index",
      "args": {}
    },
    {
      "entity_id": "twitterapi",
      "command_name": "search_tweets",
      "args": {
        "query": "$BTC",
        "max_results": "15"
      }
    },
    {
      "entity_id": "coze",
      "command_name": "market_analysis",
      "args": {"workflow_id": "analysis_workflow"}
    }
  ]
}
```

### 5. Combined Technical and Sentiment Analysis

AI wants comprehensive market analysis by combining technical indicators with sentiment data:

```json
{
  "action": {"name": "exchange.hold_position", "args": {}},
  "next_commands": [
    {
      "entity_id": "exchange",
      "command_name": "get_indicator",
      "args": {
        "type": "RSI",
        "interval": "1h",
        "window_size": "14"
      }
    },
    {
      "entity_id": "exchange",
      "command_name": "get_indicator",
      "args": {
        "type": "BOLL",
        "interval": "1h",
        "window_size": "20"
      }
    },
    {
      "entity_id": "fng",
      "command_name": "get_historical_index",
      "args": {
        "limit": "7"
      }
    },
    {
      "entity_id": "twitterapi",
      "command_name": "search_tweets",
      "args": {
        "query": "Bitcoin OR BTC",
        "max_results": "10"
      }
    }
  ]
}
```

This gives AI access to:
- 1-hour RSI (technical momentum)
- 1-hour Bollinger Bands (volatility and price position)
- 7-day Fear & Greed Index history (market sentiment trend)
- Recent Bitcoin tweets (real-time social sentiment)

## Implementation Details

### Key Components

1. **`pkg/types/result.go`**: Extended with `NextCommand` type
2. **`pkg/memory/command_memory.go`**: Command persistence manager
3. **`pkg/config/config.go`**: Configuration structures
4. **`pkg/jarvis.go`**: Command execution orchestration
5. **`pkg/env/environment.go`**: Entity management and command routing

### Execution Timeline

```
14:00:00 - Cycle N completes, AI outputs next_commands
14:00:01 - Commands saved to commands.json
...
14:15:00 - Cycle N+1 starts
14:15:01 - Load pending commands (2 commands found)
14:15:02 - Execute command 1: coze.market_sentiment
14:15:05 - Command 1 completed successfully
14:15:06 - Execute command 2: coze.on_chain_metrics
14:15:09 - Command 2 completed successfully
14:15:10 - Collect market data (kline, indicators, position)
14:15:11 - Send all data (market + command results) to AI
14:15:15 - AI makes decision based on complete context
```

## Monitoring and Debugging

### Log Messages

The system provides detailed logging:

- **Command Scheduling**: `📝 Scheduling N commands for next cycle...`
- **Command Execution**: `📋 Executing N pending commands...`
- **Success**: `✅ Command executed successfully: entity.command`
- **Failure**: `❌ Command failed permanently: entity.command - error`
- **Validation**: `⚠️ Skipping command: 'command' not supported by entity 'entity_id'`

### Checking Command Status

View the command file directly:

```bash
cat memory-bank/commands.json | jq
```

### Testing Command Execution

Test entity commands manually:

```go
// In Go code or test
entity := world.GetEntity("coze")
err := entity.HandleCommand(ctx, "market_sentiment", map[string]string{
    "workflow_id": "123456",
    "symbol": "BTCUSDT",
})
```

## Limitations

1. **Sequential Execution**: Commands execute one at a time (future: parallel execution)
2. **String Parameters Only**: All command arguments must be strings
3. **No Scheduling Priority**: Commands execute in the order they appear
4. **Single Instance**: Assumes single Jarvis instance (no distributed coordination)

## Best Practices

1. **Keep Commands Simple**: Each command should be focused and fast (< 30s)
2. **Validate Entity Support**: Check entity `Actions()` before scheduling
3. **Provide Fallbacks**: AI should handle cases where commands fail
4. **Monitor Failures**: Regularly review failed commands in logs
5. **Limit Command Count**: Avoid scheduling too many commands per cycle (impacts latency)

## Troubleshooting

### Commands Not Executing

- Check `commands.enabled: true` in config
- Verify entity is registered and running
- Check logs for validation errors

### Commands Timing Out

- Reduce workflow complexity
- Increase timeout in code (default: 30s)
- Check external API performance

### Commands Failing Validation

- Verify command name matches entity's `Actions()`
- Check entity ID is correct
- Review entity implementation of `Actions()`

## Future Enhancements

- Parallel command execution
- Command priority and scheduling
- Database-backed persistence
- UI for command status monitoring
- Command execution metrics
- Advanced retry strategies (exponential backoff)

## Related Documentation

- [Memory System](memory-system-feature.md) - Similar file-based persistence pattern
- [Limit Order Feature](limit-order-feature.md) - Another AI-driven feature
- [Entity Development Guide](entity-development.md) - How to create new entities with commands
