package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/pappi/ctrader-go/pkg/ctrader"
)

func main() {
	fmt.Println("💰 cTrader Trade Operations Example")
	fmt.Println("===================================")
	
	// TRADE session configuration
	config := &ctrader.Config{
		BeginString:  "FIX.4.4",
		SenderCompID: os.Getenv("SENDER_COMP_ID"),
		TargetCompID: "cServer",
		TargetSubID:  "TRADE",
		SenderSubID:  "TRADE",
		Username:     os.Getenv("CTRADER_USERNAME"),
		Password:     os.Getenv("CTRADER_PASSWORD"),
		HeartBeat:    30,
	}

	client := ctrader.NewClient("demo-uk-eqx-01.p.c-trader.com", 5212, config, ctrader.WithSSL(true))

	client.SetConnectedCallback(func() {
		fmt.Println("✅ Connected to TRADE server")
		
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
		fmt.Printf("📨 Trade message: %s\n", msgType)
		
		switch msgType {
		case "A": // Logon
			fmt.Println("✅ Trade logon successful!")
			
			// Start trade operations after successful logon
			go func() {
				time.Sleep(2 * time.Second)
				startTradeOperations(client, config)
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
			
		case "8": // Execution Report
			handleExecutionReport(message)
			
		case "3": // Order Reject
			handleOrderReject(message)
			
		case "AO": // Position Report
			handlePositionReport(message)
			
		case "AP": // Trade Capture Report
			handleTradeCaptureReport(message)
		}
	})

	fmt.Println("🔌 Connecting to TRADE server...")
	if err := client.Connect(); err != nil {
		log.Fatalf("❌ Failed to connect: %v", err)
	}

	// Keep running for trade operations
	fmt.Println("💰 Trade operations active. Press Ctrl+C to stop.")
	
	// Status ticker
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	
	for {
		select {
		case <-ticker.C:
			if client.IsConnected() {
				fmt.Printf("💓 Trade connection active - %s\n", time.Now().Format("15:04:05"))
			} else {
				fmt.Println("❌ Connection lost")
				return
			}
		}
	}
}

func startTradeOperations(client *ctrader.Client, config *ctrader.Config) {
	fmt.Println("🚀 Starting trade operations...")
	
	// 1. Request positions
	requestPositions(client, config)
	
	// 2. Place a test order (small size)
	go func() {
		time.Sleep(3 * time.Second)
		placeTestOrder(client, config)
	}()
}

func requestPositions(client *ctrader.Client, config *ctrader.Config) {
	fmt.Println("📋 Requesting open positions...")
	
	posReq := ctrader.NewRequestForPositions(config)
	posReq.PosReqID = "POS_REQ_001"
	
	if err := client.Send(posReq); err != nil {
		fmt.Printf("❌ Failed to request positions: %v\n", err)
	} else {
		fmt.Println("✅ Positions request sent")
	}
}

func placeTestOrder(client *ctrader.Client, config *ctrader.Config) {
	fmt.Println("📈 Placing test BUY order...")
	
	order := ctrader.NewOrderMsg(config)
	order.ClOrdID = "TEST_BUY_001"
	order.Symbol = "1" // EURUSD
	order.Side = "1"   // Buy
	order.OrderQty = 0.001 // Micro lot (1000 units)
	order.OrdType = "1"   // Market order
	
	if err := client.Send(order); err != nil {
		fmt.Printf("❌ Failed to place order: %v\n", err)
	} else {
		fmt.Println("✅ Test BUY order sent")
	}
}

func handleExecutionReport(message *ctrader.ResponseMessage) {
	orderID := message.GetFieldValue(11)
	orderStatus := message.GetFieldValue(39)
	symbol := message.GetFieldValue(55)
	side := message.GetFieldValue(54)
	
	fmt.Printf("📋 Execution Report:\n")
	fmt.Printf("   OrderID: %v\n", orderID)
	fmt.Printf("   Symbol: %v\n", symbol)
	fmt.Printf("   Side: %v\n", side)
	fmt.Printf("   Status: %v\n", orderStatus)
	
	if filledQty := message.GetFieldValue(32); filledQty != nil {
		fmt.Printf("   Filled Qty: %v\n", filledQty)
	}
	
	if avgPx := message.GetFieldValue(6); avgPx != nil {
		fmt.Printf("   Avg Price: %v\n", avgPx)
	}
}

func handleOrderReject(message *ctrader.ResponseMessage) {
	orderID := message.GetFieldValue(11)
	reason := message.GetFieldValue(58)
	
	fmt.Printf("❌ Order Rejected:\n")
	fmt.Printf("   OrderID: %v\n", orderID)
	fmt.Printf("   Reason: %v\n", reason)
}

func handlePositionReport(message *ctrader.ResponseMessage) {
	symbol := message.GetFieldValue(55)
	posQty := message.GetFieldValue(703)
	
	fmt.Printf("📊 Position Report:\n")
	fmt.Printf("   Symbol: %v\n", symbol)
	fmt.Printf("   Quantity: %v\n", posQty)
}

func handleTradeCaptureReport(message *ctrader.ResponseMessage) {
	tradeID := message.GetFieldValue(1003)
	symbol := message.GetFieldValue(55)
	side := message.GetFieldValue(54)
	
	fmt.Printf("💰 Trade Capture Report:\n")
	fmt.Printf("   TradeID: %v\n", tradeID)
	fmt.Printf("   Symbol: %v\n", symbol)
	fmt.Printf("   Side: %v\n", side)
}
