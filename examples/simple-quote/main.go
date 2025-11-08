package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/pappi/ctrader-go/pkg/ctrader"
)

func main() {
	fmt.Println("📊 cTrader Simple Market Data Subscription")
	fmt.Println("==========================================")
	
	// QUOTE session configuration
	config := &ctrader.Config{
		BeginString:  "FIX.4.4",
		SenderCompID: os.Getenv("SENDER_COMP_ID"),
		TargetCompID: "cServer",
		TargetSubID:  "QUOTE",
		SenderSubID:  "QUOTE",
		Username:     os.Getenv("CTRADER_USERNAME"),
		Password:     os.Getenv("CTRADER_PASSWORD"),
		HeartBeat:    30,
	}

	client := ctrader.NewClient("demo-uk-eqx-01.p.c-trader.com", 5211, config, ctrader.WithSSL(true))

	client.SetConnectedCallback(func() {
		fmt.Println("✅ Connected to QUOTE server")
		
		logonMsg := ctrader.NewLogonRequest(config)
		logonMsg.ResetSeqNum = true
		
		if err := client.Send(logonMsg); err != nil {
			log.Printf("❌ Failed to send logon: %v", err)
		} else {
			fmt.Println("✅ Logon message sent")
		}
	})

	client.SetDisconnectedCallback(func(err error) {
		fmt.Printf("❌ Disconnected: %v\n", err)
	})

	client.SetMessageCallback(func(message *ctrader.ResponseMessage) {
		msgType := message.GetMessageType()
		fmt.Printf("📨 Quote message: %s\n", msgType)
		
		switch msgType {
		case "A": // Logon
			fmt.Println("✅ Quote logon successful!")
			
			// Subscribe to market data after logon
			go func() {
				time.Sleep(2 * time.Second)
				subscribeToMarketData(client, config)
			}()
			
		case "0": // Heartbeat
			fmt.Println("💓 Heartbeat received")
			
		case "1": // Test Request
			testReqID := message.GetFieldValue(112)
			fmt.Printf("🧪 Test request: %v\n", testReqID)
			
			// Respond with heartbeat
			heartbeat := ctrader.NewHeartbeat(config)
			heartbeat.TestReqID = fmt.Sprintf("%v", testReqID)
			if err := client.Send(heartbeat); err != nil {
				fmt.Printf("❌ Failed to send heartbeat: %v\n", err)
			} else {
				fmt.Println("✅ Heartbeat response sent")
			}
			
		case "W": // Market Data
			handleMarketData(message)
		}
	})

	fmt.Println("🔌 Connecting to QUOTE server...")
	if err := client.Connect(); err != nil {
		log.Fatalf("❌ Failed to connect: %v", err)
	}

	// Keep running to receive market data
	fmt.Println("📊 Subscribing to market data. Press Ctrl+C to stop.")
	
	// Status ticker
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	
	for {
		select {
		case <-ticker.C:
			if client.IsConnected() {
				fmt.Printf("💓 Quote connection active - %s\n", time.Now().Format("15:04:05"))
			} else {
				fmt.Println("❌ Connection lost")
				return
			}
		}
	}
}

func subscribeToMarketData(client *ctrader.Client, config *ctrader.Config) {
	// Known symbol IDs for cTrader demo:
	// "1" = EURUSD
	// "2" = GBPUSD  
	// "3" = USDJPY
	// Note: Crypto symbols like BTCUSD may not be available on demo
	
	symbolID := "1" // EURUSD
	symbolName := "EURUSD"
	
	fmt.Printf("📊 Subscribing to %s market data with SymbolID: %s\n", symbolName, symbolID)
	
	mdReq := ctrader.NewMarketDataRequest(config)
	mdReq.MDReqID = "MD_" + symbolName + "_001"
	mdReq.SubscriptionRequestType = "1" // Snapshot + Updates
	mdReq.MarketDepth = 0
	mdReq.NoMDEntryTypes = 2 // Bid and Ask
	mdReq.MDEntryType = "0"  // Bid
	mdReq.MDEntryType = "1"  // Ask
	mdReq.NoRelatedSym = 1
	mdReq.Symbol = symbolID // Use the known symbol ID
	
	if err := client.Send(mdReq); err != nil {
		fmt.Printf("❌ Failed to subscribe: %v\n", err)
	} else {
		fmt.Println("✅ Market data subscription sent")
	}
}

func handleMarketData(message *ctrader.ResponseMessage) {
	mdReqID := message.GetFieldValue(262)
	
	if bid := message.GetFieldValue(126); bid != nil {
		fmt.Printf("📈 EURUSD [%v] Bid: %v\n", mdReqID, bid)
	}
	
	if ask := message.GetFieldValue(127); ask != nil {
		fmt.Printf("📉 EURUSD [%v] Ask: %v\n", mdReqID, ask)
	}
	
	// Show spread if both bid and ask are available
	bid := message.GetFieldValue(126)
	ask := message.GetFieldValue(127)
	if bid != nil && ask != nil {
		spread := ask.(float64) - bid.(float64)
		fmt.Printf("📊 EURUSD Spread: %.5f\n", spread)
	}
}
