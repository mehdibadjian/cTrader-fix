package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/pappi/ctrader-go/pkg/ctrader"
)

func main() {
	fmt.Println("📊 cTrader Quote Session Example")
	fmt.Println("=================================")
	
	// Configuration for QUOTE session only
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
		fmt.Println("✅ Connected to cTrader QUOTE server")
		
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
		fmt.Printf("📨 Received: %s\n", msgType)
		
		switch msgType {
		case "A": // Logon
			fmt.Println("✅ Logon successful!")
			
			// Wait a moment then request market data
			go func() {
				time.Sleep(1 * time.Second)
				requestMarketData(client, config)
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
			fmt.Println("📊 Market data received")
			bid := message.GetFieldValue(126)
			ask := message.GetFieldValue(127)
			fmt.Printf("   Bid: %v, Ask: %v\n", bid, ask)
		}
	})

	fmt.Println("🔌 Connecting to QUOTE server...")
	if err := client.Connect(); err != nil {
		log.Fatalf("❌ Failed to connect: %v", err)
	}

	// Handle messages and errors
	go func() {
		for message := range client.Messages() {
			// Messages handled by callback
			_ = message
		}
	}()

	go func() {
		for err := range client.Errors() {
			fmt.Printf("❌ Error: %v\n", err)
		}
	}()

	// Keep running
	fmt.Println("⏳ Quote session active. Press Ctrl+C to stop.")
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	
	for {
		select {
		case <-ticker.C:
			if client.IsConnected() {
				fmt.Printf("💓 Connection stable - %s\n", time.Now().Format("15:04:05"))
			} else {
				fmt.Println("❌ Connection lost")
				return
			}
		}
	}
}

func requestMarketData(client *ctrader.Client, config *ctrader.Config) {
	fmt.Println("📊 Requesting EURUSD market data...")
	
	mdReq := ctrader.NewMarketDataRequest(config)
	mdReq.MDReqID = "MD_REQ_EURUSD"
	mdReq.SubscriptionRequestType = "1" // Snapshot + Updates
	mdReq.MarketDepth = 0
	mdReq.NoMDEntryTypes = 2
	mdReq.MDEntryType = "0"  // Bid
	mdReq.NoRelatedSym = 1
	mdReq.Symbol = "1" // EURUSD symbol ID
	
	if err := client.Send(mdReq); err != nil {
		fmt.Printf("❌ Failed to request market data: %v\n", err)
	} else {
		fmt.Println("✅ Market data request sent")
	}
}
