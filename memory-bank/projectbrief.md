# Project Brief: Trading-AI

## Overview
Trading-AI is a sophisticated trading bot built on top of [bbgo](https://github.com/c9s/bbgo) and [langchaingo](https://github.com/tmc/langchaingo). It integrates AI language models to enable users to create and manage trading strategies using natural language.

## Core Purpose
To democratize algorithmic trading by allowing users to express trading strategies in natural language, which are then translated into executable code through AI.

## Key Features
- Writing trading strategies using natural language
- Support for multiple LLMs: Google AI, Open AI, Claude AI, and Ollama
- Setting take profit and stop loss parameters in strategies
- **Limit order support with price expressions** (e.g., "last_close * 0.995")
- **Automatic limit order cleanup** to maintain clean decision state
- Chatting with strategies to adjust or understand their behavior
- Support for various technical indicators (MACD, RSI, Bollinger Bands, etc.)

## Architecture Overview
The project follows a modular architecture with:
- Agent system for handling different tasks (trading, keeping)
- Environment system for interfacing with external entities (exchanges, LLMs)
- Chat/notification system for user interaction
- LLM integration for natural language processing

## Development Goals
1. Create a robust, reliable trading system that can execute strategies defined in natural language
2. Support multiple exchanges and trading pairs
3. Provide comprehensive feedback and monitoring capabilities
4. Ensure system safety with appropriate risk management features

## Project Boundaries
- Focus on cryptocurrency trading initially
- Maintain compatibility with bbgo's existing capabilities
- Prioritize reliability and safety over feature expansion

## Success Metrics
- Successful execution of trading strategies defined in natural language
- Accuracy in interpreting user intent into trading logic
- System stability and error handling
- Trading performance metrics (when applicable)
