# Technical Context: Trading-AI

## Technology Stack

### Programming Languages
- **Go** - Primary programming language (Go 1.20+)

### Core Libraries and Frameworks
- [bbgo](https://github.com/c9s/bbgo) - Foundation trading engine and exchange integrations
- [langchaingo](https://github.com/tmc/langchaingo) - Go-based LLM integration framework
- [logrus](https://github.com/sirupsen/logrus) - Structured logging
- [goja](https://github.com/dop251/goja) - JavaScript VM for price expression evaluation

### LLM Integration
- **OpenAI** - GPT models integration
- **Google AI** - Google's AI models integration
- **Anthropic** - Claude AI models integration
- **Ollama** - Local LLM deployment option

### Exchange Integrations
- Multiple cryptocurrency exchanges supported through bbgo

### Development Tools
- **Make** - Build automation
- **Docker** - Containerization
- **Go modules** - Dependency management

## Development Environment

### Prerequisites
- Go 1.20+ installed
- API keys for supported exchanges
- API keys for LLM providers (OpenAI, Google AI, Anthropic)
- Docker (optional for containerized deployment)

### Configuration
- `.env.local` file for environment configuration
- `bbgo.yaml` for bbgo-specific configuration

### Build Process
- `make build` - Build the application
- `make test` - Run tests
- Dockerfile provided for containerized builds

## Technical Constraints

### Performance Considerations
- LLM API response times can impact strategy execution speed
- Exchange API rate limits must be respected
- Real-time trading requires careful timing and execution

### Security Requirements
- Secure storage of exchange API credentials
- Secure handling of LLM API keys
- Protection against unauthorized trading actions

### Compatibility
- Must maintain compatibility with bbgo's trading engine
- Must support various LLM providers with different capabilities
- Must work with multiple exchange APIs

## Dependencies

### External Services
- Exchange APIs for trading
- LLM provider APIs for natural language processing
- Fear & Greed Index API for market sentiment

### Internal Dependencies
- Agent system depends on LLM manager
- Trading execution depends on environment system
- Notification depends on chat system

## Technical Debt and Limitations
- LLM context length limitations affect strategy complexity
- LLM interpretation quality varies by provider and model
- Exchange-specific features may not be universally available

## Tool Usage Patterns

### Configuration Pattern
```go
// Load configuration from environment and files
cfg := config.LoadConfig()
```

### LLM Integration Pattern
```go
// Initialize LLM
llmManager := llms.NewLLMManager(cfg.LLMs)
model := llmManager.GetModel("default")

// Generate response
result, err := model.GenerateContent(ctx, prompt, options...)
```

### Trading Pattern
```go
// Create trading strategy from natural language
strategy := trading.NewTradingAgent(cfg.Trading, llmModel)

// Execute trading actions
result, err := strategy.GenActions(ctx, session, messages)
```

### Price Expression Pattern
```go
// Parse price expression with dynamic variables
price, err := utils.ParsePrice(vm, klineWindow, closePrice, "last_close * 0.995")

// Variables available: last_close, last_high, last_low, last_open, prev_close, last_volume
```

### Limit Order Cleanup Pattern
```go
// Automatically cleanup unfilled limit orders at each decision cycle
// Called in Run() before emitting update_finish event
ent.cleanupLimitOrders(ctx)
```
