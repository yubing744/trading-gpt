# trading-bot
Trading bot based on bbgo and ChatGPT.

## Usage
Prepare your dotenv file .env.local and BBGO yaml config file bbgo.yaml

Config .env.local file
``` bash
# for OKEx exchange, if you have one
OKEX_API_KEY="your okex api key"
OKEX_API_SECRET="your okex api secret"
OKEX_API_PASSPHRASE="your okex api password"

# Notify
NOTIFY_FEISHU_APP_ID="your feishu app id"
NOTIFY_FEISHU_APP_SECRET="your feishu app secret"

# Chat Feishu
CHAT_FEISHU_APP_ID="your feishu app id"
CHAT_FEISHU_APP_SECRET="your feishu app secret"
CHAT_FEISHU_EVENT_ENCRYPT_KEY="your feishu event encrypt key"
CHAT_FEISHU_VERIFICATION_TOKEN="your feishu verification token"

# Agent OpenAI
AGENT_OPENAI_TOKEN="your openai api token"

# Agent chatgpt
AGENT_CHATGPT_EMAIL="your chat gpt account"
AGENT_CHATGPT_PASSWORD="your chat gpt password"
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
    chat:
      feishu:
        enabled: true
        server_port: 3000
    notify:
      feishu:
        enabled: false
        tenant_key: "your feishu tenant key"
        receive_id_type: "your feishu receive id type, like chat_id"
        receive_id: "your feishu receive id"
    agent:
      openai:
        enabled: false
        name: "your bot name"
        max_context_length: 4097
      chatgpt:
        enabled: true
        name: "your bot name"
        max_context_length: 4097
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
    symbol: OPUSDT
    interval: 1h
    leverage: 3
    max_window_size: 20
```

Run
``` bash
go run ./cmd/bbgo.go run --dotenv .env.local --config bbgo.yaml
```
