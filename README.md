# cTrader Go FIX API Client

A pure Go implementation of the cTrader FIX protocol client, inspired by the Python cTraderFixPy library. This library provides a clean, idiomatic Go interface for connecting to cTrader's FIX API for trading and market data operations.

## Features

- **Pure Go Implementation**: No external dependencies, built with standard library only
- **Full FIX Protocol Support**: Complete implementation of cTrader FIX messages
- **SSL/TLS Encryption**: Secure connections supported
- **Connection Management**: Automatic reconnection and heartbeat handling
- **Type Safety**: Strongly typed message structures
- **Concurrent Safe**: Thread-safe client implementation
- **Easy to Use**: Simple, intuitive API design

## Installation

```bash
go get github.com/pappi/ctrader-go
```

## Quick Start

```go
package main

import (
    "fmt"
    "log"
    "os"
    
    "github.com/pappi/ctrader-go/pkg/ctrader"
)

func main() {
    // Configure your cTrader FIX connection using environment variables
    config := &ctrader.Config{
        BeginString:  "FIX.4.4",
        SenderCompID: os.Getenv("SENDER_COMP_ID"), // Set this environment variable
        TargetCompID: "cServer",
        TargetSubID:  "QUOTE",
        SenderSubID:  "QUOTE",
        Username:     os.Getenv("CTRADER_USERNAME"), // Set this environment variable
        Password:     os.Getenv("CTRADER_PASSWORD"), // Set this environment variable
        HeartBeat:    30,
    }

    // Create client with SSL/TLS encryption
    client := ctrader.NewClient("demo-uk-eqx-01.p.c-trader.com", 5211, config, ctrader.WithSSL(true))

    // Set callbacks
    client.SetConnectedCallback(func() {
        fmt.Println("Connected!")
        
        // Send logon
        logon := ctrader.NewLogonRequest(config)
        logon.ResetSeqNum = true // Required for demo connections
        client.Send(logon)
    })

    client.SetMessageCallback(func(msg *ctrader.ResponseMessage) {
        fmt.Printf("Received: %s\n", msg.GetMessageType())
    })

    // Connect
    if err := client.Connect(); err != nil {
        log.Fatal(err)
    }
}
```

## Configuration

The client requires a `Config` struct with the following fields:

- `BeginString`: FIX protocol version (usually "FIX.4.4")
- `SenderCompID`: Your sender identifier
- `TargetCompID`: cTrader server identifier (usually "cServer")
- `TargetSubID`: Target sub-ID (usually "QUOTE")
- `SenderSubID`: Sender sub-ID (usually "QUOTE")
- `Username`: Your cTrader username
- `Password`: Your cTrader password
- `HeartBeat`: Heartbeat interval in seconds

### Security Best Practices

⚠️ **Never hardcode credentials in your code!** Use environment variables instead:

1. Copy the example environment file:
```bash
cp .env.example .env
```

2. Edit the `.env` file with your actual credentials:
```bash
export CTRADER_USERNAME="your_actual_username"
export CTRADER_PASSWORD="your_actual_password"
export SENDER_COMP_ID="demo.ctrader.your_actual_id"
```

3. Load the environment variables in your application:
```bash
source .env
```

Then use them in your Go code:

```go
config := &ctrader.Config{
    SenderCompID: os.Getenv("SENDER_COMP_ID"),
    Username:     os.Getenv("CTRADER_USERNAME"),
    Password:     os.Getenv("CTRADER_PASSWORD"),
    // ... other fields
}
```

Or use the helper functions with defaults:

```go
import "github.com/pappi/ctrader-go/examples/trading-bot"

config := &ctrader.Config{
    SenderCompID: getEnv("SENDER_COMP_ID", "demo.ctrader.YOUR_ID"),
    Username:     getEnv("CTRADER_USERNAME", "YOUR_USERNAME"),
    Password:     getEnv("CTRADER_PASSWORD", "YOUR_PASSWORD"),
    // ... other fields
}
```

## Message Types

### Authentication Messages

- **LogonRequest** (`MsgType=A`): Authenticate with the server
- **LogoutRequest** (`MsgType=5`): Terminate the session
- **Heartbeat** (`MsgType=0`): Respond to server heartbeats
- **TestRequest** (`MsgType=1`): Test connectivity

### Trading Messages

- **NewOrderSingle** (`MsgType=D`): Place a new order
- **OrderCancelRequest** (`MsgType=F`): Cancel an existing order
- **OrderCancelReplaceRequest** (`MsgType=G`): Modify an existing order
- **OrderStatusRequest** (`MsgType=H`): Request order status
- **OrderMassStatusRequest** (`MsgType=AF`): Request status for multiple orders

### Market Data Messages

- **MarketDataRequest** (`MsgType=V`): Subscribe to market data
- **SecurityListRequest** (`MsgType=x`): Request symbol information

### Position Messages

- **RequestForPositions** (`MsgType=AN`): Request position information

## Usage Examples

### Placing a Market Order

```go
order := ctrader.NewOrderSingle(config)
order.ClOrdID = "ORDER_123"
order.Symbol = "EURUSD"
order.Side = "1"        // 1=Buy, 2=Sell
order.OrderQty = 0.1    // Lot size
order.OrdType = "1"     // 1=Market

if err := client.Send(order); err != nil {
    log.Printf("Failed to place order: %v", err)
}
```

### Placing a Limit Order

```go
order := ctrader.NewOrderSingle(config)
order.ClOrdID = "LIMIT_ORDER_456"
order.Symbol = "EURUSD"
order.Side = "2"        // Sell
order.OrderQty = 0.1
order.OrdType = "2"     // 2=Limit
order.Price = 1.10500   // Limit price

client.Send(order)
```

### Subscribing to Market Data

```go
mdReq := ctrader.NewMarketDataRequest(config)
mdReq.MDReqID = "MD_REQ_001"
mdReq.SubscriptionRequestType = "1"  // 1=Snapshot+Updates
mdReq.MarketDepth = 0                // 0=Full book
mdReq.NoMDEntryTypes = 1
mdReq.MDEntryType = "0"              // 0=Bid
mdReq.NoRelatedSym = 1
mdReq.Symbol = "EURUSD"

client.Send(mdReq)
```

### Requesting Positions

```go
posReq := ctrader.NewRequestForPositions(config)
posReq.PosReqID = "POS_REQ_001"
client.Send(posReq)
```

## Message Handling

The client provides two ways to handle incoming messages:

### 1. Callback Approach

```go
client.SetMessageCallback(func(msg *ctrader.ResponseMessage) {
    switch msg.GetMessageType() {
    case "8": // Execution Report
        orderID := msg.GetFieldValue(11)
        status := msg.GetFieldValue(39)
        fmt.Printf("Order %s status: %s\n", orderID, status)
    }
})
```

### 2. Channel Approach

```go
go func() {
    for msg := range client.Messages() {
        fmt.Printf("Received: %s\n", msg.GetMessageType())
    }
}()
```

## Error Handling

```go
go func() {
    for err := range client.Errors() {
        log.Printf("Client error: %v", err)
    }
}()
```

## Connection Management

The client handles connection lifecycle automatically:

- **Automatic Reconnection**: The client will attempt to reconnect if the connection is lost
- **Heartbeat Management**: Automatic heartbeat responses to maintain the connection
- **Graceful Shutdown**: Proper cleanup when disconnecting

```go
// Check connection status
if client.IsConnected() {
    fmt.Println("Client is connected")
}

// Manual disconnect
client.Disconnect()

// Change message sequence number
client.ChangeMessageSequenceNumber(100)
```

## Message Validation

The protocol package provides message validation:

```go
protocol := ctrader.NewProtocol("\x01")
if err := protocol.ValidateMessage(message); err != nil {
    log.Printf("Invalid message: %v", err)
}
```

## Field Reference

### Common FIX Fields

| Field | Name | Description |
|-------|------|-------------|
| 8 | BeginString | FIX protocol version |
| 9 | BodyLength | Message body length |
| 35 | MsgType | Message type |
| 49 | SenderCompID | Sender ID |
| 56 | TargetCompID | Target ID |
| 34 | MsgSeqNum | Message sequence number |
| 52 | SendingTime | Message timestamp |
| 10 | CheckSum | Message checksum |

### Order Fields

| Field | Name | Description |
|-------|------|-------------|
| 11 | ClOrdID | Client order ID |
| 55 | Symbol | Trading symbol |
| 54 | Side | Order side (1=Buy, 2=Sell) |
| 38 | OrderQty | Order quantity |
| 40 | OrdType | Order type (1=Market, 2=Limit) |
| 44 | Price | Limit price |

## Examples

The repository includes several examples:

- **Basic Example**: Simple connection and message handling
- **Trading Bot**: Automated trading with position management

Run examples:

```bash
go run examples/basic/main.go
go run examples/trading-bot/main.go
```

## Testing

Run the test suite:

```bash
go test ./...
```

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests for new functionality
5. Submit a pull request

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Disclaimer

This software is for educational and development purposes. Use at your own risk when connecting to live trading environments. Always test thoroughly in demo environments before using with real funds.

## Support

For issues and questions:
- Open an issue on GitHub
- Check the examples directory for usage patterns
- Review the cTrader FIX API documentation for protocol specifications
