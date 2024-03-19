# Trading-GPT
Trading-GPT is a trading bot based on [bbgo](https://github.com/c9s/bbgo) and [ChatGPT](https://github.com/yubing744/chatgpt-go).

## Features
- Writing Trading Strategies Using Natural Language
- Supports multiple LLMs: Open AI, Claude AI and ollama
- Support for setting take profit and stop loss in the strategy
- Chat with strategy

## Example
* Moving average strategy
```
Trading strategy: Moving average strategy.
```

* Trend trading Strategies
```
Trading strategy: Trading on the right side, stop loss 3%, stop profit 10%.
```

* MACD divergence strategy
```
Trading strategy: Calculate the MACD indicator based on the K-line. The short period of the MACD indicator is 13, the long period is 34, the moving average period is 9, and the length of the ATR indicator is 13. Open a short order when the MACD indicator is bullish and there is a double top divergence. Open a long order when the MACD indicator is bearish and there is a double bottom divergence. Stop loss 1%, take profit 5%.
```


## Usage
Prepare your dotenv file .env.local and BBGO yaml config file bbgo.yaml

Config .env.local file
``` bash
# for OKEx exchange, if you have one
OKEX_API_KEY="your okex api key"
OKEX_API_SECRET="your okex api secret"
OKEX_API_PASSPHRASE="your okex api password"

# LLM OpenAI
LLM_OPENAI_TOKEN="your openai api token"

# LLM ANTHROPIC
LLM_ANTHROPIC_TOKEN="your claudeai api token"

```

Config bbgo.yaml file
``` yaml
persistence:
  json:
    directory: "./data/"

sessions:
  okex:
    exchange: okex
    envVarPrefix: okex
    margin: true
    isolatedMargin: false
    isolatedMarginSymbol: JUPUSDT

exchangeStrategies:
- on: okex
  jarvis:
    llm:
      openai:
        model: "gpt-4-1106-preview"
      anthropic:
        model: "claude-3-opus-20240229"
      ollama:
        server_url: "http://localhost:11434"
        model: "mistral:latest"
      primary: "ollama"
    env:
      exchange:
        window_size: 20
      include_events:
        - kline_changed
        - rsi_changed
        - boll_changed
        - fng_changed
        - position_changed
        - update_finish
    agent:
      trading:
        enabled: true
        name: "Trading AI"
        temperature: 0.1
        max_context_length: 4096
        backgroup: "I want you to act as an trading assistant. The trading assistant supports registering entities, analyzes market data provided by entities, and generates entity control commands. After receiving the command, the entity will report the result of the command execution. The goal of the transaction assistant is: to maximize returns by generating entity control commands."
    notify:
      feishu_hook:
        enabled: true
        url: "https://open.feishu.cn/open-apis/bot/v2/hook/e926c8b5-50e6-41e8-8f70-12a8631dfd93"
    symbol: JUPUSDT
    interval: 15m
    leverage: 3
    max_window_size: 20
    strategy: "Trading on the right side, trailing stop loss 3%, trailing stop profit 10%."
```

Run
``` bash
docker run --name trading-ai -d -v ${PWD}:/strategy yubing744/trading-gpt:latest run
```
