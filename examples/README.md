# cTrader Go Examples

This directory contains separate examples for different cTrader FIX API operations.

## Setup

Make sure your `.env` file is configured with your cTrader credentials:

```bash
CTRADER_USERNAME=your_username
CTRADER_PASSWORD=your_password
SENDER_COMP_ID=demo.ctrader.your_id
```

## Examples

### 1. Logon Example (`examples/logon/`)

Basic connection and authentication example.

**Purpose:** Demonstrates how to connect to cTrader and perform a successful logon.

**Features:**
- Connect to QUOTE server (port 5211)
- Send logon message with credentials
- Handle logon acknowledgment
- Respond to test requests
- Graceful logout

**Run:**
```bash
export $(cat .env | grep -v '^#' | xargs) && go run examples/logon/main.go
```

### 2. Quote & Market Data Subscription (`examples/quote-subscribe/`)

Market data subscription and real-time price updates.

**Purpose:** Shows how to subscribe to and receive market data for trading symbols.

**Features:**
- Connect to QUOTE server
- Subscribe to EURUSD market data
- Receive real-time bid/ask prices
- Handle market data updates
- Maintain connection with heartbeat responses

**Run:**
```bash
export $(cat .env | grep -v '^#' | xargs) && go run examples/quote-subscribe/main.go
```

### 3. Trade Operations (`examples/trade/`)

Trading operations including position management and order placement.

**Purpose:** Demonstrates trading functionality with the TRADE session.

**Features:**
- Connect to TRADE server (port 5212)
- Request open positions
- Place market orders
- Handle execution reports
- Process order rejections
- Manage position reports

**Run:**
```bash
export $(cat .env | grep -v '^#' | xargs) && go run examples/trade/main.go
```

## Session Types

cTrader requires separate sessions for different operations:

- **QUOTE Session (Port 5211):** Market data, price subscriptions, security lists
- **TRADE Session (Port 5212):** Order placement, position management, trade execution

## Important Notes

1. **Environment Variables:** Always load environment variables before running examples
2. **Separate Sessions:** QUOTE and TRADE operations require separate connections
3. **Heartbeat Handling:** Both sessions must respond to test requests to maintain connection
4. **Symbol IDs:** Use numeric symbol IDs (e.g., "1" for EURUSD)
5. **Order Sizes:** Use micro lots (0.001) for forex trading on demo accounts

## Troubleshooting

If connections drop:
- Verify environment variables are loaded correctly
- Check network connectivity to cTrader servers
- Ensure proper heartbeat responses are implemented
- Use correct ports: 5211 for QUOTE, 5212 for TRADE
