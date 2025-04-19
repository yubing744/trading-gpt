# System Patterns: Trading-AI

## Architecture Overview
Trading-AI follows a modular architecture pattern with clear separation of concerns:

```
Trading-AI
├── Agents System
│   ├── Trading Agent - Executes trading strategies using LLMs
│   └── Keeper Agent - Maintains system state and coordination
├── Environment System
│   ├── Exchange Entity - Interfaces with trading exchanges
│   ├── Coze Entity - Manages Coze integration
│   └── FNG Entity - Handles Fear & Greed index data
├── LLM Manager - Coordinates with various LLM providers
├── Chat System - Handles user interactions
└── Notification System - Provides updates and alerts
```

## Core Design Patterns

### Agent Pattern
The system uses an agent-based architecture where specialized agents handle specific responsibilities:
- `IAgent` interface defines the core agent capabilities
- Each agent implements `GetName()` and `GenActions()` methods
- Agents use LLMs to generate actions based on context and user input

### Environment Pattern
The environment system provides abstractions for external systems:
- `IEnvironment` interface defines the core environment capabilities
- `IEntity` represents components within environments
- Environment sessions manage state and lifecycle

### Command Pattern
Commands are used for interactions between system components:
- Actions are represented as commands with clear inputs and outputs
- Components process commands and return results

### Observer Pattern
The system uses event-driven communication:
- Components can register callbacks for specific events
- Events propagate through the system allowing loose coupling

### Template Pattern
Templates are used for generating prompts and structuring LLM interactions:
- Standardized prompt formats for consistent LLM interactions
- Template rendering for dynamic content generation

## Key Implementation Paths

### Strategy Execution Flow
1. Environment system detects market changes and generates events.
2. Events trigger evaluation within the Trading Logic/Strategy.
3. Trading Logic requests decisions from the Trading Agent when necessary.
4. Trading Agent processes context and events, potentially using LLM, to generate trading actions.
5. Actions are sent to the Environment system for execution on the exchange.
6. Results and system status are reported back through the Notification system.

### Risk Management Flow
1. Strategy includes stop loss and take profit parameters
2. These parameters are extracted and validated
3. Orders are placed with appropriate risk management
4. Positions are monitored for trigger conditions
5. Actions are taken when conditions are met

### Communication Flow
1. User inputs come through Chat system
2. Inputs are processed by appropriate Agent
3. Responses are generated using LLM
4. Responses are delivered through Notification system

## Critical Technical Decisions
1. Using bbgo as the foundational trading engine
2. Adopting langchaingo for LLM integration
3. Supporting multiple LLM providers for flexibility
4. Implementing a modular architecture for extensibility
5. Using natural language as the primary interface
