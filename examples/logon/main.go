package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/pappi/ctrader-go/pkg/ctrader"
)

func main() {
	fmt.Println("🔐 cTrader Logon Example")
	fmt.Println("========================")
	
	// Load environment variables
	config := &ctrader.Config{
		BeginString:  "FIX.4.4",
		SenderCompID: os.Getenv("SENDER_COMP_ID"),
		TargetCompID: "cServer",
		TargetSubID:  "QUOTE", // Using QUOTE for basic logon test
		SenderSubID:  "QUOTE",
		Username:     os.Getenv("CTRADER_USERNAME"),
		Password:     os.Getenv("CTRADER_PASSWORD"),
		HeartBeat:    30,
	}

	fmt.Printf("Configuration:\n")
	fmt.Printf("  SenderCompID: %s\n", config.SenderCompID)
	fmt.Printf("  Username: %s\n", config.Username)
	fmt.Printf("  TargetSubID: %s\n", config.TargetSubID)
	fmt.Println()

	client := ctrader.NewClient("demo-uk-eqx-01.p.c-trader.com", 5211, config, ctrader.WithSSL(true))

	client.SetConnectedCallback(func() {
		fmt.Println("✅ Connected to cTrader server")
		
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
			
		case "5": // Logout
			fmt.Println("👋 Logout received")
		}
	})

	fmt.Println("🔌 Connecting to server...")
	if err := client.Connect(); err != nil {
		log.Fatalf("❌ Failed to connect: %v", err)
	}

	// Wait for logon completion
	fmt.Println("⏳ Waiting for logon...")
	time.Sleep(5 * time.Second)
	
	if client.IsConnected() {
		fmt.Println("✅ Logon example completed successfully")
		
		// Logout gracefully
		logoutMsg := ctrader.NewLogoutRequest(config)
		client.Send(logoutMsg)
		time.Sleep(1 * time.Second)
	}
	
	client.Disconnect()
	fmt.Println("🔌 Disconnected")
}
