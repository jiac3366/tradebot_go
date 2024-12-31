# TradeBot

A WebSocket client for subscribing to Binance trade streams.

## Prerequisites

- Go 1.22 or higher
- Git

## Quick Start

1. Install dependencies

```bash
go mod tidy
```

2. Run the benchmark

```bash
cd benchmark
go run benchmark.go
```


## Plan

### (1) Public Connector

- `ExchangeManager` to fetch meta data for exchange
- `WSManager` to manage websocket connection 
- `WSClient` to subscribe to websocket streams
- `MsgBus` to publish and subscribe to messages
- `PublicConnector` be combined with `MsgBus` to push market data to `MsgBus`


### (2) Private Connector

