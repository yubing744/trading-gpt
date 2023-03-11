# Trading-GPT
Trading-GPT is a trading bot based on [bbgo](https://github.com/c9s/bbgo) and [ChatGPT](https://github.com/yubing744/chatgpt-go).

## Features
- Writing Trading Strategies Using Natural Language
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

# Agent chatgpt
AGENT_OPENAI_TOKEN="your openai api token"
```

Config bbgo.yaml file
``` yaml
---
persistence:
  redis:
    host: 127.0.0.1  # The IP address or the hostname to your Redis server, 127.0.0.1 if same as BBGO  
    port: 6379  # Port to Redis server, default 6379
    db: 0  # DB number to use. You can set to another DB to avoid conflict if other applications are using Redis too.

sessions:
  okex:
    exchange: okex
    envVarPrefix: okex
    margin: true
    isolatedMargin: false
    isolatedMarginSymbol: OPUSDT

exchangeStrategies:
- on: okex
  jarvis:
    env:
      exchange:
        window_size: 20
      include_events:
        - kline_changed
        - position_changed
        - update_finish
    agent:
      openai:
        enabled: true
        name: "AI"
        model: "gpt-3.5-turbo-0301"
        temperature: 0.5
        backgroup: "I want you to act as an trading assistant. The trading assistant supports registering entities, analyzes market data provided by entities, and generates entity control commands. After receiving the command, the entity will report the result of the command execution. The goal of the transaction assistant is: to maximize returns by generating entity control commands."
        max_context_length: 4097
    notify:
      feishu_hook:
        enabled: true
        url: "your feishu group custom webhook url"
    symbol: OPUSDT
    interval: 5m
    leverage: 3
    max_window_size: 20
    prompt: "Your natural language strategy"
```

Install redis and config port 6379
```
apt install -y redis
systemctl status redis.service
```

Run
``` bash
docker run --name trading-gpt --net host -d -v ${PWD}:/strategy yubing744/trading-gpt:latest run
```
