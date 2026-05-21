# Chomolungma

A Go-based cryptocurrency exchange data sync and trade execution module, supporting **Binance** and **HuoBi**. It provides two core capabilities:

- **Market data sync** — synchronises K-line data into MongoDB *(legacy; superseded by [ChomoSyncer](https://github.com/GeekChomolungma/ChomoSyncer))*
- **Trade execution** — order placement, cancellation, account queries, and real-time order push *(primary focus of this module)*

---

## Directory Structure

```text
Chomolungma/
├── config/          # Configuration loading (INI file)
├── db/              # MongoDB connection layer
│   └── mongoInc/    # Generic MetaCollection[T] wrapper (new mongo-driver)
├── dtos/            # Data transfer objects (DTOs)
├── engine/
│   ├── engine.go    # TradeEngine main engine
│   ├── binance/     # Binance data sync (legacy)
│   └── huobi/       # HuoBi trade execution + data sync (legacy)
│       ├── clients/
│       │   ├── accountclient.go          # Account HTTP client
│       │   ├── orderclient.go            # Order HTTP client
│       │   ├── marketwebsocketclient/    # Market-data WebSocket clients
│       │   └── orderwebsocketclient/     # Order-update WebSocket V2 client
│       ├── internal/
│       │   └── requestbuilder/           # HMAC-SHA256 URL signer
│       └── model/                        # Request / response models
└── build.sh
```

---

## Engine Architecture

The main `TradeEngine` uses a **Cylinder** abstraction — each exchange maps to one cylinder:

```go
// engine/engine.go
type Cylinder interface {
    Ignite()    // Start up: open connections, subscribe to streams
    Flush()     // Periodic flush: persist sync state
    Flameout()  // Graceful shutdown: unsubscribe, close connections
}
```

The Binance cylinder is currently active; the HuoBi cylinder is commented out but fully implemented:

```go
func (te *TradeEngine) Load() {
    te.Cylinders["binance"] = &binance.BinanCylinder{}
    // te.Cylinders["huobi"] = &huobi.HuoBiCylinder{}
}
```

---

## Trade Execution (Primary Focus)

Trade execution lives under `engine/huobi/`. The core design routes all authenticated private API requests **through a Gateway server** rather than calling the exchange directly.

### Overall Request Flow

```text
Business logic
  │
  ├─ AccountClient / OrderClient
  │     └─ PrivateUrlBuilder (builds fully signed URL with HMAC-SHA256)
  │           └─ Wrapped as BaseReqModel { AimSite, Method, Url, Body }
  │                 └─ HTTP POST → Gateway Server
  │                       └─ /api/v1/Chomolungma/entrypoint
  │                             └─ Forwarded to HuoBi Exchange API
  └─ Parse Data field from BaseRspModel { Code, Msg, Data }
```

All private interfaces (account, order placement, cancellation) never call the exchange directly — the Gateway handles that, decoupling signature logic from network policy.

---

### Signature Authentication

File: [engine/huobi/internal/requestbuilder/privateurlbuilder.go](engine/huobi/internal/requestbuilder/privateurlbuilder.go)

`PrivateUrlBuilder` follows the HuoBi API v2 signature specification, automatically appending auth parameters and computing an HMAC-SHA256 signature on every request:

| Parameter | Description |
| --- | --- |
| `AccessKeyId` | API access key |
| `SignatureMethod` | Fixed value: `HmacSHA256` |
| `SignatureVersion` | Fixed value: `2` |
| `Timestamp` | UTC time in `2006-01-02T15:04:05` format |
| `Signature` | HMAC-SHA256 over method + host + path + sorted params (URL-encoded) |

Usage:

```go
builder := new(requestbuilder.PrivateUrlBuilder).Init(accessKey, secretKey, "api.huobi.pro")
signedURL := builder.Build("POST", "/v1/order/orders/place", nil)
```

---

### Account Operations (AccountClient)

File: [engine/huobi/clients/accountclient.go](engine/huobi/clients/accountclient.go)

```go
client := new(clients.AccountClient).Init(gatewayHost, accessKey, secretKey, apiHost)
```

| Method | Description | HuoBi API |
| --- | --- | --- |
| `GetAccountInfo()` | List all accounts owned by this API user | `GET /v1/account/accounts` |
| `GetAccountBalance(id)` | Get balance for a specific account ID | `GET /v1/account/accounts/{id}/balance` |
| `GetAccountAssetValuation(type, currency, subUid)` | Total asset valuation in BTC or fiat | `GET /v2/account/asset-valuation` |
| `GetAccountHistory(id, opts)` | Account asset change records | `GET /v1/account/history` |
| `GetAccountLedger(id, opts)` | Account financial ledger | `GET /v2/account/ledger` |
| `FuturesTransfer(req)` | Transfer between spot and futures accounts | `POST /v1/futures/transfer` |
| `TransferAccount(req)` | Transfer assets between accounts | `POST /v1/account/transfer` |
| `GetPointBalance(subUid)` | Get point balance | `GET /v2/point/account` |
| `TransferPoint(req)` | Transfer points | `POST /v2/point/transfer` |

**Example — query tradable balance for a specific currency:**

```go
// Resolve account ID list first
accounts, err := client.GetAccountInfo()

// Fetch balance by account ID
balance, err := client.GetAccountBalance(accounts[0].Id)
for _, b := range balance.List {
    fmt.Printf("Currency: %s, Type: %s, Balance: %s\n", b.Currency, b.Type, b.Balance)
}
```

---

### Order Operations (OrderClient)

File: [engine/huobi/clients/orderclient.go](engine/huobi/clients/orderclient.go)

```go
client := new(clients.OrderClient).Init(gatewayHost, accessKey, secretKey, apiHost)
```

#### 1. Place an order

```go
// engine/huobi/model/order/placeorderrequest.go
req := &order.PlaceOrderRequest{
    AccountId: "12345678",
    Symbol:    "btcusdt",
    Type:      "buy-limit",      // buy-market / sell-market / buy-limit / sell-limit
    Amount:    "0.001",
    Price:     "30000",          // omit for market orders
    Source:    "spot-api",       // spot-api / margin-api / super-margin-api / c2c-margin-api
}

resp, err := client.PlaceOrder(req)
// resp.Data holds the new order ID on success
```

**Supported order types (`Type` field):**

| Type | Description |
| --- | --- |
| `buy-market` | Market buy |
| `sell-market` | Market sell |
| `buy-limit` | Limit buy |
| `sell-limit` | Limit sell |
| `buy-stop-limit` | Stop-limit buy |
| `sell-stop-limit` | Stop-limit sell |

#### 2. Place multiple orders (up to 10)

```go
reqs := []order.PlaceOrderRequest{req1, req2}
resp, err := client.PlaceOrders(reqs)
```

#### 3. Cancel an order

```go
// By order ID (routed through Gateway)
resp, err := client.CancelOrderById("1234567890")

// By client order ID (direct to exchange)
resp, err := client.CancelOrderByClientOrderId("my-order-001")

// Batch cancel by criteria
batchReq := &order.CancelOrdersByCriteriaRequest{
    AccountId: "12345678",
    Symbol:    "btcusdt",
}
resp, err := client.CancelOrdersByCriteria(batchReq)

// Batch cancel by ID list
idsReq := &order.CancelOrdersByIdsRequest{
    OrderIds: []string{"111", "222", "333"},
}
resp, err := client.CancelOrdersByIds(idsReq)
```

#### 4. Query orders

```go
// Single order by ID (routed through Gateway)
orderResp, err := client.GetOrderById("1234567890")

// Single order by criteria
getReq := new(model.GetRequest).Init()
getReq.AddParam("clientOrderId", "my-order-001")
orderResp, err := client.GetOrderByCriteria(getReq)

// Trade fills for an order
matchResp, err := client.GetMatchResultsById("1234567890")

// Historical orders
histReq := new(model.GetRequest).Init()
histReq.AddParam("symbol", "btcusdt")
histReq.AddParam("states", "filled")
histResp, err := client.GetHistoryOrders(histReq)

// Orders within the last 48 hours (routed through Gateway)
recent48Req := new(model.GetRequest).Init()
recent48Req.AddParam("symbol", "btcusdt")
resp, err := client.GetLast48hOrders(recent48Req)

// Current open orders
openReq := new(model.GetRequest).Init()
openReq.AddParam("symbol", "btcusdt")
openResp, err := client.GetOpenOrders(openReq)
```

#### 5. Query transaction fee rate

```go
feeReq := new(model.GetRequest).Init()
feeReq.AddParam("symbols", "btcusdt,ethusdt")
feeResp, err := client.GetTransactFeeRate(feeReq)
```

---

### Real-time Order Push (WebSocket V2)

File: [engine/huobi/clients/orderwebsocketclient/subscribeorderwebsocketv2client.go](engine/huobi/clients/orderwebsocketclient/subscribeorderwebsocketv2client.go)  
Base: [engine/huobi/clients/websocketclientbase/websocketv2clientbase.go](engine/huobi/clients/websocketclientbase/websocketv2clientbase.go)

Order push uses HuoBi's WebSocket V2 protocol with mandatory authentication. `HuoBiCylinder` subscribes to all configured symbols at startup:

```go
// engine/huobi/huobiCylinder.go — inside Ignite()
for accountID := range config.AccountMap {
    subOrderUpdateV2(config.OrderSymbols, accountID)
}
```

The `WebSocketV2ClientBase` handles the full lifecycle:

- Connects to `ws://<gatewayHost>/ws/v2` and immediately sends an HMAC-SHA256 auth frame
- Responds to server `ping` with `pong`; auto-reconnects when no data is received beyond `ReconnectWaitV2Second`
- Dispatches three message actions: `req` (auth response), `sub` (subscription ACK), `push` (data event)

**Subscription example:**

```go
wsClient := new(orderwebsocketclient.SubscribeOrderWebSocketV2Client).Init(
    gatewayHost, accessKey, secretKey, apiHost,
)

wsClient.SetHandler(
    func(resp *auth.WebSocketV2AuthenticationResponse) {
        // authentication success callback
    },
    func(resp interface{}) {
        orderResp := resp.(order.SubscribeOrderV2Response)
        info := orderResp.Data
        fmt.Printf("Order update: symbol=%s, status=%s, orderId=%d\n",
            info.Symbol, info.OrderStatus, info.OrderId)
    },
)

wsClient.Connect(true)                      // true = enable auto-reconnect
wsClient.Subscribe("btcusdt", "clientId-001")
```

**`OrderInfo` push payload fields:**

| Field | Description |
| --- | --- |
| `EventType` | `creation` / `trade` / `cancellation` |
| `Symbol` | Trading pair, e.g. `btcusdt` |
| `AccountId` | Account ID |
| `OrderId` | Exchange order ID |
| `OrderSide` | `buy` or `sell` |
| `OrderPrice` | Limit price |
| `OrderSize` | Order quantity |
| `OrderStatus` | `submitted` / `partial-filled` / `filled` / `canceled` |
| `TradePrice` | Fill price (on `trade` events) |
| `TradeVolume` | Fill quantity |
| `RemainAmt` | Remaining unfilled quantity |
| `LastActTime` | Timestamp of last state change (ms) |
| `ErrorCode` | Non-zero on error |

---

## DTO Reference

### External order entry (`dtos/climberModel.go`)

Defines the payload the upstream strategy layer (Climber) sends into Chomolungma:

```go
type OrderPlace struct {
    Symbol string // e.g. "btcusdt"
    Model  string // buy-market / sell-market / buy-limit / sell-limit
    Amount string // order quantity (quote amount for market orders)
    Price  string // limit price; omitted for market orders
    Source string // spot-api / margin-api / super-margin-api / c2c-margin-api
}

type OrderCancel struct {
    OrderID string
}

type OrderQuery struct {
    OrderID string
}

type CurrencyBalanceReq struct {
    Currency string
}
```

### Gateway communication (`dtos/baseModel.go`)

```go
// Outbound to Gateway
type BaseReqModel struct {
    AimSite string // target exchange, e.g. "HuoBi"
    Method  string // "GET" or "POST"
    Url     string // fully signed URL built by PrivateUrlBuilder
    Body    string // JSON-encoded POST body
}

// Inbound from Gateway
type BaseRspModel struct {
    Code int    // 0 = OK; see dtos/errors.go for other codes
    Msg  string
    Data string // raw JSON response from the exchange
}
```

---

## Configuration (`my.ini`)

```ini
[mongo]
Uri = mongodb://localhost:27017

[server]
Host = 0.0.0.0:8080

[gateway]
GatewayHost     = 127.0.0.1:9090
GatewayTcpHost  = 127.0.0.1:9091

[huobi]
ApiServerHost = api.huobi.pro
AccessKey     = ak1,ak2
AccountId     = id1,id2
SecretKey     = sk1,sk2

[binance]
AccessKey = your-binance-ak
SecretKey = your-binance-sk

[marketsub]
Symbols = BTCUSDT,ETHUSDT,ETCUSDT
Periods = 1m,5m,1h,1d

[validate]
MarketUrl = http://127.0.0.1:8081
Open      = true
```

> Multiple HuoBi accounts are supported — `AccessKey`, `AccountId`, and `SecretKey` are comma-separated and index-aligned.  
> Hot-reload: call `config.ReloadKeys()` to refresh API keys at runtime without restarting the process.

---

## Market Data Sync (Brief)

> ⚠️ This module has been rewritten in a separate repository. The code here is the original implementation, kept for reference.

- **Binance**: `BinanCylinder.Ignite()` launches two goroutines per symbol/interval:
  - `SyncHistoricalKline()` — paginates the REST API in 499-bar windows with checkpoint-based resume (stored in the `marketSyncFlag` MongoDB collection)
  - `SubscribeKlineStream()` — real-time WebSocket K-line updates; deduplicates against MongoDB and auto-resubscribes on error

- **HuoBi**: `HuoBiCylinder.Ignite()` subscribes to K-line WebSocket streams; `Flush` persists sync timestamps to MongoDB every hour (`FlushDuration = 3600 s`)

---

## Dependencies

| Package | Purpose |
| --- | --- |
| `github.com/binance/binance-connector-go` | Binance official SDK |
| `github.com/gorilla/websocket` | WebSocket connection management |
| `go.mongodb.org/mongo-driver` | MongoDB client (Binance data, generic `MetaCollection[T]`) |
| `gopkg.in/mgo.v2` | MongoDB client (HuoBi legacy data) |
| `github.com/gin-gonic/gin` | HTTP server framework |
| `github.com/go-ini/ini` | INI config file parsing |
| `github.com/json-iterator/go` | High-performance JSON serialisation |
| `go.uber.org/zap` | Structured logging |
| `github.com/shopspring/decimal` | Precise decimal arithmetic |
