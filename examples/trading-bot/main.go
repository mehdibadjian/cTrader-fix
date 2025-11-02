package main

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/pappi/ctrader-go/pkg/ctrader"
)

type TradingBot struct {
	client      *ctrader.Client
	config      *ctrader.Config
	orderID     int
	positionID  string
	symbol      string
	isRunning   bool
}

func NewTradingBot() *TradingBot {
	// Configuration for cTrader FIX API
	config := &ctrader.Config{
		BeginString:  "FIX.4.4",
		SenderCompID: getEnv("SENDER_COMP_ID", "demo.ctrader.5539991"),
		TargetCompID: getEnv("TARGET_COMP_ID", "cServer"),
		TargetSubID:  getEnv("TARGET_SUB_ID", "QUOTE"),
		SenderSubID:  getEnv("SENDER_SUB_ID", "QUOTE"),
		Username:     getEnv("CTRADER_USERNAME", "5539991"), // Use numeric login only
		Password:     getEnv("CTRADER_PASSWORD", "Test1234#"),
		HeartBeat:    30,
	}

	client := ctrader.NewClient("demo-uk-eqx-01.p.c-trader.com", 5211, config, ctrader.WithSSL(true))

	return &TradingBot{
		client:     client,
		config:     config,
		orderID:    1000,
		symbol:     getEnv("SYMBOL", "EURUSD"),
		isRunning:  false,
	}
}

func (bot *TradingBot) Start() error {
	// Set callbacks
	bot.client.SetConnectedCallback(bot.onConnected)
	bot.client.SetDisconnectedCallback(bot.onDisconnected)
	bot.client.SetMessageCallback(bot.onMessage)

	// Connect to server
	fmt.Println("Connecting to cTrader FIX server...")
	if err := bot.client.Connect(); err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}

	// Start message handling
	go bot.handleMessages()
	go bot.handleErrors()

	bot.isRunning = true
	fmt.Println("Trading bot started")

	// Wait for shutdown signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	fmt.Println("Shutting down trading bot...")
	bot.Stop()
	return nil
}

func (bot *TradingBot) Stop() {
	bot.isRunning = false
	
	if bot.client.IsConnected() {
		// Send logout message
		logoutMsg := ctrader.NewLogoutRequest(bot.config)
		bot.client.Send(logoutMsg)
		
		// Disconnect
		bot.client.Disconnect()
	}
	
	fmt.Println("Trading bot stopped")
}

func (bot *TradingBot) onConnected() {
	fmt.Println("Connected to cTrader FIX server")
	
	// Send logon message
	logonMsg := ctrader.NewLogonRequest(bot.config)
	logonMsg.ResetSeqNum = true
	
	if err := bot.client.Send(logonMsg); err != nil {
		log.Printf("Failed to send logon: %v", err)
	} else {
		fmt.Println("Logon message sent")
	}
}

func (bot *TradingBot) onDisconnected(err error) {
	fmt.Printf("Disconnected from server: %v\n", err)
	bot.isRunning = false
}

func (bot *TradingBot) onMessage(message *ctrader.ResponseMessage) {
	msgType := message.GetMessageType()
	fmt.Printf("Received message type: %s\n", msgType)

	switch msgType {
	case "A": // Logon
		fmt.Println("Logon successful")
		bot.startTrading()
		
	case "0": // Heartbeat
		// fmt.Println("Heartbeat received")
		
	case "1": // Test Request
		bot.handleTestRequest(message)
		
	case "8": // Execution Report
		bot.handleExecutionReport(message)
		
	case "AP": // Trade Capture Report
		bot.handleTradeCaptureReport(message)
		
	case "AO": // Position Report
		bot.handlePositionReport(message)
		
	default:
		// fmt.Printf("Unhandled message type: %s\n", msgType)
	}
}

func (bot *TradingBot) handleTestRequest(message *ctrader.ResponseMessage) {
	testReqID := message.GetFieldValue(112)
	fmt.Printf("Test request received: %v\n", testReqID)
	
	// Respond with heartbeat
	heartbeat := ctrader.NewHeartbeat(bot.config)
	heartbeat.TestReqID = fmt.Sprintf("%v", testReqID)
	bot.client.Send(heartbeat)
}

func (bot *TradingBot) handleSecurityList(message *ctrader.ResponseMessage) {
	fmt.Println("Security list received")
	// Process security list if needed
}

func (bot *TradingBot) handleExecutionReport(message *ctrader.ResponseMessage) {
	orderID := message.GetFieldValue(11)
	orderStatus := message.GetFieldValue(39)
	symbol := message.GetFieldValue(55)
	side := message.GetFieldValue(54)
	orderQty := message.GetFieldValue(38)
	price := message.GetFieldValue(44)
	
	fmt.Printf("Execution Report - Order: %v, Status: %v, Symbol: %v, Side: %v, Qty: %v, Price: %v\n",
		orderID, orderStatus, symbol, side, orderQty, price)
}

func (bot *TradingBot) handleTradeCaptureReport(message *ctrader.ResponseMessage) {
	symbol := message.GetFieldValue(55)
	side := message.GetFieldValue(54)
	orderQty := message.GetFieldValue(32)
	price := message.GetFieldValue(31)
	
	fmt.Printf("Trade Capture - Symbol: %v, Side: %v, Qty: %v, Price: %v\n",
		symbol, side, orderQty, price)
}

func (bot *TradingBot) handlePositionReport(message *ctrader.ResponseMessage) {
	symbol := message.GetFieldValue(55)
	longQty := message.GetFieldValue(704)
	shortQty := message.GetFieldValue(705)
	
	fmt.Printf("Position Report - Symbol: %v, Long: %v, Short: %v\n",
		symbol, longQty, shortQty)
}

func (bot *TradingBot) startTrading() {
	fmt.Println("Starting trading logic...")
	
	// Request security list
	securityReq := ctrader.NewSecurityListRequest(bot.config)
	securityReq.SecurityReqID = "SEC_REQ_001"
	securityReq.SecurityListRequestType = "4" // All securities
	securityReq.Symbol = bot.symbol
	bot.client.Send(securityReq)
	
	// Request positions
	posReq := ctrader.NewRequestForPositions(bot.config)
	posReq.PosReqID = "POS_REQ_001"
	bot.client.Send(posReq)
	
	// Start periodic trading
	go bot.tradingLoop()
}

func (bot *TradingBot) tradingLoop() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for bot.isRunning && bot.client.IsConnected() {
		select {
		case <-ticker.C:
			bot.placeRandomOrder()
		}
	}
}

func (bot *TradingBot) placeRandomOrder() {
	bot.orderID++
	clOrdID := fmt.Sprintf("ORDER_%d", bot.orderID)
	
	// Random order parameters
	sides := []string{"1", "2"} // 1=Buy, 2=Sell
	side := sides[rand.Intn(len(sides))]
	
	orderTypes := []string{"1", "2"} // 1=Market, 2=Limit
	orderType := orderTypes[rand.Intn(len(orderTypes))]
	
	orderQty := 0.01 + rand.Float64()*0.09 // Random quantity between 0.01 and 0.1
	
	var price float64
	if orderType == "2" { // Limit order
		// Set a random price (in real implementation, you'd get current market price)
		price = 1.0500 + rand.Float64()*0.0200 // Random price between 1.0500 and 1.0700
	}
	
	// Create order
	order := ctrader.NewOrderMsg(bot.config)
	order.ClOrdID = clOrdID
	order.Symbol = bot.symbol
	order.Side = side
	order.OrderQty = orderQty
	order.OrdType = orderType
	
	if orderType == "2" {
		order.Price = price
	}
	
	fmt.Printf("Placing order: %s %s %.2f @ %.5f\n", 
		bot.getSideName(side), bot.symbol, orderQty, price)
	
	if err := bot.client.Send(order); err != nil {
		log.Printf("Failed to place order: %v", err)
	}
}

func (bot *TradingBot) getSideName(side string) string {
	switch side {
	case "1":
		return "BUY"
	case "2":
		return "SELL"
	default:
		return "UNKNOWN"
	}
}

func (bot *TradingBot) handleMessages() {
	for message := range bot.client.Messages() {
		// Messages are already handled by the callback
		_ = message
	}
}

func (bot *TradingBot) handleErrors() {
	for err := range bot.client.Errors() {
		log.Printf("Client error: %v", err)
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func main() {
	fmt.Println("cTrader Trading Bot")
	fmt.Println("===================")
	
	bot := NewTradingBot()
	
	if err := bot.Start(); err != nil {
		log.Fatalf("Failed to start trading bot: %v", err)
	}
}
