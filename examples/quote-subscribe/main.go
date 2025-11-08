package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/pappi/ctrader-go/pkg/ctrader"
)

// To change symbols, modify these functions:
// 1. requestSecurityList() - Change the Symbol field to your desired symbol (e.g., "BTCUSD", "GBPUSD")
// 2. handleSecurityListResponse() - Update the success message
// 3. subscribeToMarketData() - Update the MDReqID and success message
// 4. handleMarketData() - Update the price display message
//
// Common forex symbols: EURUSD, GBPUSD, USDJPY, AUDUSD
// Crypto symbols may not be available on demo servers

var messageSequence int = 1

func main() {
	fmt.Println("📊 cTrader Quote & Market Data Subscription Example")
	fmt.Println("====================================================")
	
	// QUOTE session configuration (reverted - TRADE doesn't respond to security list either)
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

	var securityID string // Store the security ID we get from the server

	client := ctrader.NewClient("demo-uk-eqx-01.p.c-trader.com", 5211, config, ctrader.WithSSL(true))

	client.SetConnectedCallback(func() {
		fmt.Println("✅ Connected to QUOTE server")
		
		logonMsg := ctrader.NewLogonRequest(config)
		logonMsg.ResetSeqNum = true
		
		// Log the raw logon message being sent
		protocol := ctrader.NewProtocol("\x01")
		rawLogon := protocol.FormatMessage(logonMsg.GetMessage(messageSequence))
		fmt.Printf("🔤 SENDING Logon Message (Seq: %d):\n%s\n", messageSequence, rawLogon)
		messageSequence++
		
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
		
		// Log raw FIX message for all responses
		protocol := ctrader.NewProtocol("\x01")
		rawMessage := protocol.FormatMessage(message.GetMessage())
		fmt.Printf("🔤 RECEIVED Raw FIX Message:\n%s\n", rawMessage)
		
		switch msgType {
		case "A": // Logon
			fmt.Println("✅ Quote logon successful!")
			
			// Wait a moment then request security list
			go func() {
				time.Sleep(1 * time.Second)
				requestSecurityList(client, config)
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
			
		case "y": // Security List Response
			fmt.Println("📋 Security list received")
			securityID = handleSecurityListResponse(message)
			if securityID != "" {
				// Subscribe to market data with the correct security ID
				go func() {
					time.Sleep(1 * time.Second)
					subscribeToMarketData(client, config, securityID)
				}()
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
	fmt.Println("📊 Waiting for security list, then subscribing to market data. Press Ctrl+C to stop.")
	
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

func requestSecurityList(client *ctrader.Client, config *ctrader.Config) {
	fmt.Println("📋 Requesting security list for EURUSD (common forex pair)...")
	
	securityReq := ctrader.NewSecurityListRequest(config)
	securityReq.SecurityReqID = "SEC_REQ_EURUSD"
	securityReq.SecurityListRequestType = "0" // Symbol
	securityReq.Symbol = "EURUSD" // Request by symbol name
	
	// Log the raw FIX message being sent
	protocol := ctrader.NewProtocol("\x01")
	rawMessage := protocol.FormatMessage(securityReq.GetMessage(messageSequence))
	fmt.Printf("🔤 SENDING Security List Request (Seq: %d):\n%s\n", messageSequence, rawMessage)
	messageSequence++
	
	if err := client.Send(securityReq); err != nil {
		fmt.Printf("❌ Failed to send security list: %v\n", err)
	} else {
		fmt.Println("✅ Security list request sent")
	}
}

func handleSecurityListResponse(message *ctrader.ResponseMessage) string {
	securityReqID := message.GetFieldValue(320)
	symbol := message.GetFieldValue(55)
	securityID := message.GetFieldValue(48)
	
	fmt.Printf("📋 Security List Response:\n")
	fmt.Printf("   RequestID: %v\n", securityReqID)
	fmt.Printf("   Symbol: %v\n", symbol)
	fmt.Printf("   SecurityID: %v\n", securityID)
	
	// Convert securityID to string if it's not already
	var secID string
	if securityID != nil {
		secID = fmt.Sprintf("%v", securityID)
		fmt.Printf("✅ Found EURUSD SecurityID: %s\n", secID)
		return secID
	}
	
	fmt.Println("❌ Could not find EURUSD SecurityID")
	return ""
}

func subscribeToMarketData(client *ctrader.Client, config *ctrader.Config, securityID string) {
	fmt.Printf("📊 Subscribing to EURUSD market data with SecurityID: %s\n", securityID)
	
	mdReq := ctrader.NewMarketDataRequest(config)
	mdReq.MDReqID = "MD_EURUSD_001"
	mdReq.SubscriptionRequestType = "1" // Snapshot + Updates
	mdReq.MarketDepth = 0
	mdReq.NoMDEntryTypes = 2 // Bid and Ask
	mdReq.MDEntryType = "0"  // Bid
	mdReq.MDEntryType = "1"  // Ask
	mdReq.NoRelatedSym = 1
	mdReq.Symbol = securityID // Use the security ID from the server
	
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
}
