# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Development Commands

### Build & Test Commands
```bash
# Build the application
make build              # Builds to ./build/bbgo
make build-linux        # Cross-compile for Linux
make clean             # Clean build artifacts

# Run the application
make run               # Build and run with bbgo.yaml config
./build/bbgo run --dotenv .env.local --config bbgo.yaml --lightweight false --no-sync false

# Testing
make unit-test         # Run all unit tests (go test ./pkg/...)
go test ./pkg/agents/... -v    # Test specific modules
go test -run TestSpecificFunction  # Run specific test

# Docker development
make docker-build      # Build Docker image
make docker-start      # Run in container with volume mount
make docker-logs       # View container logs
make docker-stop       # Stop and remove container
```

### BBGO Library Commands
```bash
# From libs/bbgo directory:
make bbgo             # Build with web interface
make bbgo-slim        # Build without web interface  
make static           # Build frontend assets
make migrations       # Compile database migrations
```

## Architecture Overview

### Event-Driven Trading System
The core architecture follows an event-driven pattern where:
1. **Environment System** monitors market data and triggers events
2. **Event Collection** aggregates market changes (K-line, indicators, positions)
3. **AI Agent** processes events using LLM to generate trading decisions
4. **Trading Execution** carries out AI-generated commands through BBGO

Key orchestration happens in `pkg/jarvis.go` (main event loop) which coordinates between:
- Trading Agent (`pkg/agents/trading/`)
- Environment entities (`pkg/env/exchange/`, `pkg/env/coze/`)
- LLM system (`pkg/llms/`)
- Notification system (`pkg/notify/`)

### Multi-LLM Integration
- Supports OpenAI, Google AI, Claude, and Ollama through unified interface
- LLM Manager (`pkg/llms/llm_manager.go`) handles provider switching and fallback
- Structured prompts in `pkg/prompt/` for consistent AI interactions
- Configurable primary/secondary LLM selection in bbgo.yaml

### Memory & Learning System
- **Trade Reflection**: Automatically analyzes closed positions using LLM
- **Memory Bank**: Stores trade analysis in `memory-bank/reflections/` as Markdown files
- **Learning Integration**: Past reflections inform future trading decisions
- Triggered by `PositionClosedEvent` and managed in `pkg/jarvis.go`

## Configuration Setup

### Environment Variables (.env.local)
```bash
# Exchange API credentials
OKEX_API_KEY="your_okex_api_key"
OKEX_API_SECRET="your_okex_api_secret"
OKEX_API_PASSPHRASE="your_okex_api_password"

# LLM API tokens
LLM_GOOGLEAI_APIKEY="your_googleai_api_key"
LLM_OPENAI_TOKEN="your_openai_api_token"
LLM_ANTHROPIC_TOKEN="your_anthropic_api_token"

# Optional services
COZE_API_KEY="your_coze_api_key"
```

### Strategy Configuration (bbgo.yaml)
- Natural language strategy definition in `strategy:` field
- LLM configuration with primary/secondary selection
- Environment settings for indicators and market data
- Agent configuration with temperature and context limits
- Notification channels (Feishu hook integration)

## Code Organization Patterns

### Agent Interface Pattern
All agents implement `IAgent` interface with:
- `Start()` and `Stop()` lifecycle methods
- `GenActions()` for processing messages and generating actions
- Dependency injection through configuration

### Environment & Entity Abstraction
- `Environment` manages external system interactions
- `Entity` interface for data sources (Exchange, Coze, FNG index)
- Clean separation between trading logic and external APIs

### Configuration-Driven Architecture
- Hierarchical configuration in `pkg/config/`
- Type-specific configs (agent, LLM, chat, environment)
- YAML-based with environment variable overrides

## Testing Patterns

### Framework & Structure
- Uses testify framework (`assert`, `require`)
- Mock objects follow interface pattern (`types.NewMockSession`)
- Integration tests use real exchange connections
- Test configuration via `.env.local`

### Test Organization
```bash
pkg/                    # Unit tests alongside source
test/integration/       # Integration tests
test/e2e/              # End-to-end tests
```

## Key Files to Understand

- `pkg/jarvis.go` - Main strategy orchestrator and event loop
- `pkg/agents/trading/trading_agent.go` - AI trading decision logic
- `pkg/env/exchange/exchange_entity.go` - Exchange interface and trading execution
- `pkg/llms/llm_manager.go` - Multi-LLM provider management
- `pkg/prompt/prompt.go` - Structured LLM prompt templates
- `pkg/config/config.go` - Configuration loading and validation

## Safety & Risk Management

- Emergency position closure on agent errors
- Validation of LLM-generated commands before execution
- Configurable stop-loss and take-profit enforcement
- Retry mechanisms with configurable attempts
- Trade size and frequency limits

## Natural Language Strategy Development

Users define trading strategies in plain English within the YAML config. The system:
1. Parses natural language strategy descriptions
2. Uses LLM to understand market conditions and user intent
3. Generates specific trading actions (buy/sell orders)
4. Executes through BBGO's exchange integrations
5. Learns from outcomes via reflection system

## Memory Bank Usage

The `memory-bank/` directory contains:
- `reflections/` - Auto-generated trade analysis
- `activeContext.md` - Current market context
- `systemPatterns.md` - Observed trading patterns
- `techContext.md` - Technical analysis insights

These files inform future AI decisions and help the system learn from trading history.