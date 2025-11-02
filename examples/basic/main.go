package main

import (
	"fmt"
	"log"
	"time"

	"github.com/pappi/ctrader-go/pkg/ctrader"
)

func main() {
	fmt.Println("cTrader FIX API Basic Example")
	fmt.Println("==============================")
	
	// Configuration for cTrader Demo - Following official Python specification
	config := &ctrader.Config{
		BeginString:  "FIX.4.4",
		SenderCompID: "demo.ctrader.5539991",
		TargetCompID: "cServer",  // FIXED: Must be "cServer" (lowercase 'c')
		TargetSubID:  "TRADE",    // FIXED: Use TRADE stream for trading
		SenderSubID:  "TRADE",    // FIXED: Must match TargetSubID
		Username:     "5539991",  // Numeric login only
		Password:     "Test1234#",
		HeartBeat:    30,
	}

	// Create client with SSL/TLS encryption
	client := ctrader.NewClient("demo-uk-eqx-01.p.c-trader.com", 5212, config, ctrader.WithSSL(true)) // FIXED: Port 5212 for TRADE

	// Set callbacks
	client.SetConnectedCallback(func() {
		fmt.Println("Connected to cTrader FIX server")
		
		// Send logon message
		logonMsg := ctrader.NewLogonRequest(config)
		logonMsg.ResetSeqNum = true
		
		if err := client.Send(logonMsg); err != nil {
			log.Printf("Failed to send logon: %v", err)
		} else {
			fmt.Println("Logon message sent")
		}
	})

	client.SetDisconnectedCallback(func(err error) {
		fmt.Printf("Disconnected from server: %v\n", err)
	})

	client.SetMessageCallback(func(message *ctrader.ResponseMessage) {
		fmt.Printf("Received message: %s\n", message.GetMessageType())
		
		// Handle different message types
		switch message.GetMessageType() {
		case "A": // Logon
			fmt.Println("Logon successful")
			
			// Send a test request
			testReq := ctrader.NewTestRequest(config)
			testReq.TestReqID = "TEST123"
			client.Send(testReq)
			
		case "0": // Heartbeat
			fmt.Println("Heartbeat received")
			
		case "1": // Test Request
			testReqID := message.GetFieldValue(112)
			fmt.Printf("Test request received: %v\n", testReqID)
			
			// Respond with heartbeat
			heartbeat := ctrader.NewHeartbeat(config)
			heartbeat.TestReqID = fmt.Sprintf("%v", testReqID)
			client.Send(heartbeat)
			
		default:
			fmt.Printf("Unhandled message type: %s\n", message.GetMessageType())
		}
	})

	// Connect to server
	fmt.Println("Connecting to cTrader FIX server...")
	if err := client.Connect(); err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}

	// Listen for messages and errors
	go func() {
		for message := range client.Messages() {
			protocol := ctrader.NewProtocol("\x01")
			fmt.Println("=== Received Message ===")
			fmt.Print(protocol.FormatMessage(message.GetMessage()))
			fmt.Println("========================")
		}
	}()

	go func() {
		for err := range client.Errors() {
			log.Printf("Error: %v", err)
		}
	}()

	// Keep the application running
	fmt.Println("Client is running. Press Ctrl+C to stop.")
	
	// Send periodic heartbeats if needed
	ticker := time.NewTicker(25 * time.Second)
	defer ticker.Stop()
	
	for {
		select {
		case <-ticker.C:
			if client.IsConnected() {
				heartbeat := ctrader.NewHeartbeat(config)
				if err := client.Send(heartbeat); err != nil {
					log.Printf("Failed to send heartbeat: %v", err)
				}
			}
		}
	}
}
