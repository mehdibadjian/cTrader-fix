package ctrader

import (
	"strings"
	"testing"
)

func TestConfig(t *testing.T) {
	config := &Config{
		BeginString:  "FIX.4.4",
		SenderCompID: "TEST_SENDER",
		TargetCompID: "cServer",
		TargetSubID:  "QUOTE",
		SenderSubID:  "QUOTE",
		Username:     "testuser",
		Password:     "testpass",
		HeartBeat:    30,
	}

	if config.BeginString != "FIX.4.4" {
		t.Errorf("Expected BeginString to be FIX.4.4, got %s", config.BeginString)
	}

	if config.HeartBeat != 30 {
		t.Errorf("Expected HeartBeat to be 30, got %d", config.HeartBeat)
	}
}

func TestResponseMessage(t *testing.T) {
	message := "8=FIX.4.4\x019=100\x0135=A\x0149=SENDER\x0156=TARGET\x0134=1\x0152=20231101-10:00:00\x0198=0\x01108=30\x01553=user\x01554=pass\x0110=123\x01"
	
	responseMsg := NewResponseMessage(message, "\x01")
	
	if msgType := responseMsg.GetMessageType(); msgType != "A" {
		t.Errorf("Expected message type A, got %s", msgType)
	}
	
	if sender := responseMsg.GetFieldValue(49); sender != "SENDER" {
		t.Errorf("Expected sender SENDER, got %v", sender)
	}
	
	if target := responseMsg.GetFieldValue(56); target != "TARGET" {
		t.Errorf("Expected target TARGET, got %v", target)
	}
	
	if nonExistent := responseMsg.GetFieldValue(999); nonExistent != nil {
		t.Errorf("Expected nil for non-existent field, got %v", nonExistent)
	}
}

func TestLogonRequest(t *testing.T) {
	config := &Config{
		BeginString:  "FIX.4.4",
		SenderCompID: "TEST_SENDER",
		TargetCompID: "cServer",
		TargetSubID:  "QUOTE",
		SenderSubID:  "QUOTE",
		Username:     "testuser",
		Password:     "testpass",
		HeartBeat:    30,
	}

	logonMsg := NewLogonRequest(config)
	logonMsg.ResetSeqNum = true
	
	message := logonMsg.GetMessage(1)
	
	if message == "" {
		t.Error("Expected non-empty message")
	}
	
	if !strings.Contains(message, "35=A") {
		t.Error("Message should contain MsgType=A")
	}
	
	if !strings.Contains(message, "553=testuser") {
		t.Error("Message should contain username")
	}
	
	if !strings.Contains(message, "554=testpass") {
		t.Error("Message should contain password")
	}
}

func TestHeartbeat(t *testing.T) {
	config := &Config{
		BeginString:  "FIX.4.4",
		SenderCompID: "TEST_SENDER",
		TargetCompID: "cServer",
		TargetSubID:  "QUOTE",
		SenderSubID:  "QUOTE",
		Username:     "testuser",
		Password:     "testpass",
		HeartBeat:    30,
	}

	heartbeat := NewHeartbeat(config)
	message := heartbeat.GetMessage(1)
	
	if message == "" {
		t.Error("Expected non-empty message")
	}
	
	if !strings.Contains(message, "35=0") {
		t.Error("Message should contain MsgType=0")
	}
	
	heartbeat.TestReqID = "TEST123"
	messageWithTestReqID := heartbeat.GetMessage(2)
	
	if !strings.Contains(messageWithTestReqID, "112=TEST123") {
		t.Error("Message should contain TestReqID")
	}
}

func TestTestRequest(t *testing.T) {
	config := &Config{
		BeginString:  "FIX.4.4",
		SenderCompID: "TEST_SENDER",
		TargetCompID: "cServer",
		TargetSubID:  "QUOTE",
		SenderSubID:  "QUOTE",
		Username:     "testuser",
		Password:     "testpass",
		HeartBeat:    30,
	}

	testReq := NewTestRequest(config)
	testReq.TestReqID = "TEST123"
	
	message := testReq.GetMessage(1)
	
	if message == "" {
		t.Error("Expected non-empty message")
	}
	
	if !strings.Contains(message, "35=1") {
		t.Error("Message should contain MsgType=1")
	}
	
	if !strings.Contains(message, "112=TEST123") {
		t.Error("Message should contain TestReqID")
	}
}

func TestOrderMsg(t *testing.T) {
	config := &Config{
		BeginString:  "FIX.4.4",
		SenderCompID: "TEST_SENDER",
		TargetCompID: "cServer",
		TargetSubID:  "QUOTE",
		SenderSubID:  "QUOTE",
		Username:     "testuser",
		Password:     "testpass",
		HeartBeat:    30,
	}

	order := NewOrderMsg(config)
	order.ClOrdID = "ORDER_123"
	order.Symbol = "EURUSD"
	order.Side = "1"
	order.OrderQty = 0.1
	order.OrdType = "1"
	
	message := order.GetMessage(1)
	
	if message == "" {
		t.Error("Expected non-empty message")
	}
	
	if !strings.Contains(message, "35=D") {
		t.Error("Message should contain MsgType=D")
	}
	
	if !strings.Contains(message, "11=ORDER_123") {
		t.Error("Message should contain ClOrdID")
	}
	
	if !strings.Contains(message, "55=EURUSD") {
		t.Error("Message should contain Symbol")
	}
	
	if !strings.Contains(message, "54=1") {
		t.Error("Message should contain Side")
	}
	
	if !strings.Contains(message, "38=0.10") {
		t.Error("Message should contain OrderQty")
	}
	
	if !strings.Contains(message, "40=1") {
		t.Error("Message should contain OrdType")
	}
}

func TestOrderMsgWithLimit(t *testing.T) {
	config := &Config{
		BeginString:  "FIX.4.4",
		SenderCompID: "TEST_SENDER",
		TargetCompID: "cServer",
		TargetSubID:  "QUOTE",
		SenderSubID:  "QUOTE",
		Username:     "testuser",
		Password:     "testpass",
		HeartBeat:    30,
	}

	order := NewOrderMsg(config)
	order.ClOrdID = "LIMIT_ORDER_456"
	order.Symbol = "EURUSD"
	order.Side = "2"
	order.OrderQty = 0.1
	order.OrdType = "2"
	order.Price = 1.10500
	
	message := order.GetMessage(1)
	
	if !strings.Contains(message, "44=1.10500") {
		t.Error("Message should contain Price")
	}
}

func TestOrderCancelRequest(t *testing.T) {
	config := &Config{
		BeginString:  "FIX.4.4",
		SenderCompID: "TEST_SENDER",
		TargetCompID: "cServer",
		TargetSubID:  "QUOTE",
		SenderSubID:  "QUOTE",
		Username:     "testuser",
		Password:     "testpass",
		HeartBeat:    30,
	}

	cancelReq := NewOrderCancelRequest(config)
	cancelReq.OrigClOrdID = "ORDER_123"
	cancelReq.ClOrdID = "CANCEL_456"
	
	message := cancelReq.GetMessage(1)
	
	if message == "" {
		t.Error("Expected non-empty message")
	}
	
	if !strings.Contains(message, "35=F") {
		t.Error("Message should contain MsgType=F")
	}
	
	if !strings.Contains(message, "41=ORDER_123") {
		t.Error("Message should contain OrigClOrdID")
	}
	
	if !strings.Contains(message, "11=CANCEL_456") {
		t.Error("Message should contain ClOrdID")
	}
}

func TestMarketDataRequest(t *testing.T) {
	config := &Config{
		BeginString:  "FIX.4.4",
		SenderCompID: "TEST_SENDER",
		TargetCompID: "cServer",
		TargetSubID:  "QUOTE",
		SenderSubID:  "QUOTE",
		Username:     "testuser",
		Password:     "testpass",
		HeartBeat:    30,
	}

	mdReq := NewMarketDataRequest(config)
	mdReq.MDReqID = "MD_REQ_001"
	mdReq.SubscriptionRequestType = "1"
	mdReq.MarketDepth = 0
	mdReq.NoMDEntryTypes = 1
	mdReq.MDEntryType = "0"
	mdReq.NoRelatedSym = 1
	mdReq.Symbol = "EURUSD"
	
	message := mdReq.GetMessage(1)
	
	if message == "" {
		t.Error("Expected non-empty message")
	}
	
	if !strings.Contains(message, "35=V") {
		t.Error("Message should contain MsgType=V")
	}
	
	if !strings.Contains(message, "262=MD_REQ_001") {
		t.Error("Message should contain MDReqID")
	}
	
	if !strings.Contains(message, "55=EURUSD") {
		t.Error("Message should contain Symbol")
	}
}

func TestProtocolValidation(t *testing.T) {
	protocol := NewProtocol("\x01")
	
	// Create a valid message with correct checksum
	config := &Config{
		BeginString:  "FIX.4.4",
		SenderCompID: "SENDER",
		TargetCompID: "TARGET",
		TargetSubID:  "QUOTE",
		SenderSubID:  "QUOTE",
		Username:     "user",
		Password:     "pass",
		HeartBeat:    30,
	}
	
	logonMsg := NewLogonRequest(config)
	validMessage := logonMsg.GetMessage(1)
	
	if err := protocol.ValidateMessage(validMessage); err != nil {
		t.Errorf("Expected valid message to pass validation, got error: %v", err)
	}
	
	if err := protocol.ValidateMessage(""); err == nil {
		t.Error("Expected empty message to fail validation")
	}
	
	invalidMessage := "35=A\x0149=SENDER\x01"
	if err := protocol.ValidateMessage(invalidMessage); err == nil {
		t.Error("Expected message without required fields to fail validation")
	}
}

func TestProtocolFieldNames(t *testing.T) {
	protocol := NewProtocol("\x01")
	fieldNames := protocol.GetFieldNames()
	
	if len(fieldNames) == 0 {
		t.Error("Expected field names map to not be empty")
	}
	
	if fieldNames[35] != "MsgType" {
		t.Errorf("Expected field 35 to be MsgType, got %s", fieldNames[35])
	}
	
	if fieldNames[49] != "SenderCompID" {
		t.Errorf("Expected field 49 to be SenderCompID, got %s", fieldNames[49])
	}
}

func TestProtocolMessageTypes(t *testing.T) {
	protocol := NewProtocol("\x01")
	messageTypes := protocol.GetMessageTypeName()
	
	if len(messageTypes) == 0 {
		t.Error("Expected message types map to not be empty")
	}
	
	if messageTypes["A"] != "Logon" {
		t.Errorf("Expected message type A to be Logon, got %s", messageTypes["A"])
	}
	
	if messageTypes["0"] != "Heartbeat" {
		t.Errorf("Expected message type 0 to be Heartbeat, got %s", messageTypes["0"])
	}
}

func TestProtocolFormatMessage(t *testing.T) {
	protocol := NewProtocol("\x01")
	message := "8=FIX.4.4\x019=100\x0135=A\x0149=SENDER\x0156=TARGET\x0134=1\x0152=20231101-10:00:00\x0198=0\x01108=30\x01553=user\x01554=pass\x0110=123\x01"
	
	formatted := protocol.FormatMessage(message)
	
	if formatted == "" {
		t.Error("Expected formatted message to not be empty")
	}
	
	if !strings.Contains(formatted, "Message Type:") {
		t.Error("Formatted message should contain message type")
	}
}

func TestMessageSequenceNumber(t *testing.T) {
	config := &Config{
		BeginString:  "FIX.4.4",
		SenderCompID: "TEST_SENDER",
		TargetCompID: "cServer",
		TargetSubID:  "QUOTE",
		SenderSubID:  "QUOTE",
		Username:     "testuser",
		Password:     "testpass",
		HeartBeat:    30,
	}

	logonMsg := NewLogonRequest(config)
	
	msg1 := logonMsg.GetMessage(1)
	msg2 := logonMsg.GetMessage(2)
	
	if msg1 == msg2 {
		t.Error("Messages with different sequence numbers should be different")
	}
	
	if !strings.Contains(msg1, "34=1") {
		t.Error("First message should contain sequence number 1")
	}
	
	if !strings.Contains(msg2, "34=2") {
		t.Error("Second message should contain sequence number 2")
	}
}

func TestMessageTimestamp(t *testing.T) {
	config := &Config{
		BeginString:  "FIX.4.4",
		SenderCompID: "TEST_SENDER",
		TargetCompID: "cServer",
		TargetSubID:  "QUOTE",
		SenderSubID:  "QUOTE",
		Username:     "testuser",
		Password:     "testpass",
		HeartBeat:    30,
	}

	logonMsg := NewLogonRequest(config)
	
	message := logonMsg.GetMessage(1)
	
	if !strings.Contains(message, "52=") {
		t.Error("Message should contain timestamp field (52)")
	}
	
	if len(message) < 20 {
		t.Error("Message should be long enough to contain timestamp")
	}
}
