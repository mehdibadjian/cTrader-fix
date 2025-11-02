package ctrader

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	BeginString  string
	SenderCompID string
	TargetCompID string
	TargetSubID  string
	SenderSubID  string
	Username     string
	Password     string
	HeartBeat    int
}

type ResponseMessage struct {
	message string
	fields  map[int][]string
}

func NewResponseMessage(message, delimiter string) *ResponseMessage {
	processedMessage := strings.ReplaceAll(message, delimiter, "|")
	fields := make(map[int][]string)
	
	parts := strings.Split(message, delimiter)
	for _, part := range parts {
		if part == "" {
			continue
		}
		if eqIndex := strings.Index(part, "="); eqIndex != -1 {
			fieldNumStr := part[:eqIndex]
			fieldValue := part[eqIndex+1:]
			if fieldNum, err := strconv.Atoi(fieldNumStr); err == nil {
				fields[fieldNum] = append(fields[fieldNum], fieldValue)
			}
		}
	}
	
	return &ResponseMessage{
		message: processedMessage,
		fields:  fields,
	}
}

func (rm *ResponseMessage) GetFieldValue(fieldNumber int) interface{} {
	values, exists := rm.fields[fieldNumber]
	if !exists {
		return nil
	}
	if len(values) == 1 {
		return values[0]
	}
	return values
}

func (rm *ResponseMessage) GetMessageType() string {
	if values, exists := rm.fields[35]; exists && len(values) > 0 {
		return values[0]
	}
	return ""
}

func (rm *ResponseMessage) GetMessage() string {
	return rm.message
}

type RequestMessageInterface interface {
	GetMessage(sequenceNumber int) string
	getBody() string
	getHeader(lenBody int, sequenceNumber int) string
	getTrailer(headerAndBody string) string
}

type RequestMessage struct {
	messageType string
	config      *Config
	delimiter   string
}

func NewRequestMessage(messageType string, config *Config) *RequestMessage {
	return &RequestMessage{
		messageType: messageType,
		config:      config,
		delimiter:   "\x01",
	}
}

func (rm *RequestMessage) GetMessage(sequenceNumber int) string {
	body := rm.getBody()
	var headerAndBody string
	if body != "" {
		header := rm.getHeader(len(body), sequenceNumber)
		headerAndBody = fmt.Sprintf("%s%s%s%s", header, rm.delimiter, body, rm.delimiter)
	} else {
		header := rm.getHeader(0, sequenceNumber)
		headerAndBody = fmt.Sprintf("%s%s", header, rm.delimiter)
	}
	trailer := rm.getTrailer(headerAndBody)
	return fmt.Sprintf("%s%s%s", headerAndBody, trailer, rm.delimiter)
}

func (rm *RequestMessage) getBody() string {
	return ""
}

func (rm *RequestMessage) getHeader(lenBody int, sequenceNumber int) string {
	var fields []string
	fields = append(fields, fmt.Sprintf("35=%s", rm.messageType))
	fields = append(fields, fmt.Sprintf("49=%s", rm.config.SenderCompID))
	fields = append(fields, fmt.Sprintf("56=%s", rm.config.TargetCompID))
	fields = append(fields, fmt.Sprintf("57=%s", rm.config.TargetSubID))
	fields = append(fields, fmt.Sprintf("50=%s", rm.config.SenderSubID))
	fields = append(fields, fmt.Sprintf("34=%d", sequenceNumber))
	fields = append(fields, fmt.Sprintf("52=%s", time.Now().UTC().Format("20060102-15:04:05.000")))
	
	fieldsJoined := strings.Join(fields, rm.delimiter)
	return fmt.Sprintf("8=%s%s9=%d%s%s", rm.config.BeginString, rm.delimiter, lenBody+len(fieldsJoined)+2, rm.delimiter, fieldsJoined)
}

func (rm *RequestMessage) getTrailer(headerAndBody string) string {
	messageBytes := []byte(headerAndBody)
	checksum := 0
	for _, b := range messageBytes {
		checksum += int(b)
	}
	checksum = checksum % 256
	return fmt.Sprintf("10=%03d", checksum)
}

type LogonRequest struct {
	*RequestMessage
	EncryptionScheme int
	ResetSeqNum      bool
}

func NewLogonRequest(config *Config) *LogonRequest {
	return &LogonRequest{
		RequestMessage:  NewRequestMessage("A", config),
		EncryptionScheme: 0,
		ResetSeqNum:      false,
	}
}

func (lr *LogonRequest) GetMessage(sequenceNumber int) string {
	body := lr.GetBody()
	var headerAndBody string
	if body != "" {
		header := lr.getHeader(len(body), sequenceNumber)
		headerAndBody = fmt.Sprintf("%s%s%s%s", header, lr.delimiter, body, lr.delimiter)
	} else {
		header := lr.getHeader(0, sequenceNumber)
		headerAndBody = fmt.Sprintf("%s%s", header, lr.delimiter)
	}
	trailer := lr.getTrailer(headerAndBody)
	return fmt.Sprintf("%s%s%s", headerAndBody, trailer, lr.delimiter)
}

func (lr *LogonRequest) GetBody() string {
	var fields []string
	fields = append(fields, fmt.Sprintf("98=%d", lr.EncryptionScheme))
	fields = append(fields, fmt.Sprintf("108=%d", lr.config.HeartBeat))
	if lr.ResetSeqNum {
		fields = append(fields, "141=Y")
	}
	fields = append(fields, fmt.Sprintf("553=%s", lr.config.Username))
	fields = append(fields, fmt.Sprintf("554=%s", lr.config.Password))
	return strings.Join(fields, lr.delimiter)
}

type Heartbeat struct {
	*RequestMessage
	TestReqID string
}

func NewHeartbeat(config *Config) *Heartbeat {
	return &Heartbeat{
		RequestMessage: NewRequestMessage("0", config),
	}
}

func (h *Heartbeat) GetMessage(sequenceNumber int) string {
	body := h.GetBody()
	var headerAndBody string
	if body != "" {
		header := h.getHeader(len(body), sequenceNumber)
		headerAndBody = fmt.Sprintf("%s%s%s%s", header, h.delimiter, body, h.delimiter)
	} else {
		header := h.getHeader(0, sequenceNumber)
		headerAndBody = fmt.Sprintf("%s%s", header, h.delimiter)
	}
	trailer := h.getTrailer(headerAndBody)
	return fmt.Sprintf("%s%s%s", headerAndBody, trailer, h.delimiter)
}

func (h *Heartbeat) GetBody() string {
	if h.TestReqID == "" {
		return ""
	}
	return fmt.Sprintf("112=%s", h.TestReqID)
}

type TestRequest struct {
	*RequestMessage
	TestReqID string
}

func NewTestRequest(config *Config) *TestRequest {
	return &TestRequest{
		RequestMessage: NewRequestMessage("1", config),
	}
}

func (tr *TestRequest) GetMessage(sequenceNumber int) string {
	body := tr.GetBody()
	var headerAndBody string
	if body != "" {
		header := tr.getHeader(len(body), sequenceNumber)
		headerAndBody = fmt.Sprintf("%s%s%s%s", header, tr.delimiter, body, tr.delimiter)
	} else {
		header := tr.getHeader(0, sequenceNumber)
		headerAndBody = fmt.Sprintf("%s%s", header, tr.delimiter)
	}
	trailer := tr.getTrailer(headerAndBody)
	return fmt.Sprintf("%s%s%s", headerAndBody, trailer, tr.delimiter)
}

func (tr *TestRequest) GetBody() string {
	return fmt.Sprintf("112=%s", tr.TestReqID)
}

type LogoutRequest struct {
	*RequestMessage
}

func (lr *LogoutRequest) GetMessage(sequenceNumber int) string {
	body := lr.GetBody()
	var headerAndBody string
	if body != "" {
		header := lr.getHeader(len(body), sequenceNumber)
		headerAndBody = fmt.Sprintf("%s%s%s%s", header, lr.delimiter, body, lr.delimiter)
	} else {
		header := lr.getHeader(0, sequenceNumber)
		headerAndBody = fmt.Sprintf("%s%s", header, lr.delimiter)
	}
	trailer := lr.getTrailer(headerAndBody)
	return fmt.Sprintf("%s%s%s", headerAndBody, trailer, lr.delimiter)
}

func (lr *LogoutRequest) GetBody() string {
	return ""
}

func NewLogoutRequest(config *Config) *LogoutRequest {
	return &LogoutRequest{
		RequestMessage: NewRequestMessage("5", config),
	}
}

type OrderMsg struct {
	*RequestMessage
	ClOrdID  string
	Symbol   string
	Side     string
	OrderQty float64
	OrdType  string
	Price    float64
}

func NewOrderMsg(config *Config) *OrderMsg {
	return &OrderMsg{
		RequestMessage: NewRequestMessage("D", config),
	}
}

func (nos *OrderMsg) GetMessage(sequenceNumber int) string {
	body := nos.GetBody()
	var headerAndBody string
	if body != "" {
		header := nos.getHeader(len(body), sequenceNumber)
		headerAndBody = fmt.Sprintf("%s%s%s%s", header, nos.delimiter, body, nos.delimiter)
	} else {
		header := nos.getHeader(0, sequenceNumber)
		headerAndBody = fmt.Sprintf("%s%s", header, nos.delimiter)
	}
	trailer := nos.getTrailer(headerAndBody)
	return fmt.Sprintf("%s%s%s", headerAndBody, trailer, nos.delimiter)
}

func (nos *OrderMsg) GetBody() string {
	var fields []string
	fields = append(fields, fmt.Sprintf("11=%s", nos.ClOrdID))
	fields = append(fields, fmt.Sprintf("55=%s", nos.Symbol))
	fields = append(fields, fmt.Sprintf("54=%s", nos.Side))
	fields = append(fields, fmt.Sprintf("60=%s", time.Now().UTC().Format("20060102-15:04:05")))
	fields = append(fields, fmt.Sprintf("38=%.2f", nos.OrderQty))
	fields = append(fields, fmt.Sprintf("40=%s", nos.OrdType))
	if nos.Price != 0 {
		fields = append(fields, fmt.Sprintf("44=%.5f", nos.Price))
	}
	return strings.Join(fields, nos.delimiter)
}

type OrderCancelRequest struct {
	*RequestMessage
	OrigClOrdID string
	OrderID     string
	ClOrdID     string
}

func NewOrderCancelRequest(config *Config) *OrderCancelRequest {
	return &OrderCancelRequest{
		RequestMessage: NewRequestMessage("F", config),
	}
}

func (ocr *OrderCancelRequest) GetMessage(sequenceNumber int) string {
	body := ocr.GetBody()
	var headerAndBody string
	if body != "" {
		header := ocr.getHeader(len(body), sequenceNumber)
		headerAndBody = fmt.Sprintf("%s%s%s%s", header, ocr.delimiter, body, ocr.delimiter)
	} else {
		header := ocr.getHeader(0, sequenceNumber)
		headerAndBody = fmt.Sprintf("%s%s", header, ocr.delimiter)
	}
	trailer := ocr.getTrailer(headerAndBody)
	return fmt.Sprintf("%s%s%s", headerAndBody, trailer, ocr.delimiter)
}

func (ocr *OrderCancelRequest) GetBody() string {
	var fields []string
	fields = append(fields, fmt.Sprintf("41=%s", ocr.OrigClOrdID))
	if ocr.OrderID != "" {
		fields = append(fields, fmt.Sprintf("37=%s", ocr.OrderID))
	}
	fields = append(fields, fmt.Sprintf("11=%s", ocr.ClOrdID))
	return strings.Join(fields, ocr.delimiter)
}

type MarketDataRequest struct {
	*RequestMessage
	MDReqID                 string
	SubscriptionRequestType string
	MarketDepth             int
	NoMDEntryTypes          int
	MDEntryType             string
	NoRelatedSym            int
	Symbol                  string
}

func NewMarketDataRequest(config *Config) *MarketDataRequest {
	return &MarketDataRequest{
		RequestMessage: NewRequestMessage("V", config),
	}
}

func (mdr *MarketDataRequest) GetMessage(sequenceNumber int) string {
	body := mdr.GetBody()
	var headerAndBody string
	if body != "" {
		header := mdr.getHeader(len(body), sequenceNumber)
		headerAndBody = fmt.Sprintf("%s%s%s%s", header, mdr.delimiter, body, mdr.delimiter)
	} else {
		header := mdr.getHeader(0, sequenceNumber)
		headerAndBody = fmt.Sprintf("%s%s", header, mdr.delimiter)
	}
	trailer := mdr.getTrailer(headerAndBody)
	return fmt.Sprintf("%s%s%s", headerAndBody, trailer, mdr.delimiter)
}

func (mdr *MarketDataRequest) GetBody() string {
	var fields []string
	fields = append(fields, fmt.Sprintf("262=%s", mdr.MDReqID))
	fields = append(fields, fmt.Sprintf("263=%s", mdr.SubscriptionRequestType))
	fields = append(fields, fmt.Sprintf("264=%d", mdr.MarketDepth))
	fields = append(fields, fmt.Sprintf("267=%d", mdr.NoMDEntryTypes))
	fields = append(fields, fmt.Sprintf("269=%s", mdr.MDEntryType))
	fields = append(fields, fmt.Sprintf("146=%d", mdr.NoRelatedSym))
	fields = append(fields, fmt.Sprintf("55=%s", mdr.Symbol))
	return strings.Join(fields, mdr.delimiter)
}

type SecurityListRequest struct {
	*RequestMessage
	SecurityReqID           string
	SecurityListRequestType string
	Symbol                  string
}

func NewSecurityListRequest(config *Config) *SecurityListRequest {
	return &SecurityListRequest{
		RequestMessage: NewRequestMessage("x", config),
	}
}

func (slr *SecurityListRequest) GetMessage(sequenceNumber int) string {
	body := slr.GetBody()
	var headerAndBody string
	if body != "" {
		header := slr.getHeader(len(body), sequenceNumber)
		headerAndBody = fmt.Sprintf("%s%s%s%s", header, slr.delimiter, body, slr.delimiter)
	} else {
		header := slr.getHeader(0, sequenceNumber)
		headerAndBody = fmt.Sprintf("%s%s", header, slr.delimiter)
	}
	trailer := slr.getTrailer(headerAndBody)
	return fmt.Sprintf("%s%s%s", headerAndBody, trailer, slr.delimiter)
}

func (slr *SecurityListRequest) GetBody() string {
	var fields []string
	fields = append(fields, fmt.Sprintf("320=%s", slr.SecurityReqID))
	fields = append(fields, fmt.Sprintf("559=%s", slr.SecurityListRequestType))
	if slr.Symbol != "" {
		fields = append(fields, fmt.Sprintf("55=%s", slr.Symbol))
	}
	return strings.Join(fields, slr.delimiter)
}

type RequestForPositions struct {
	*RequestMessage
	PosReqID      string
	PosMaintRptID string
}

func NewRequestForPositions(config *Config) *RequestForPositions {
	return &RequestForPositions{
		RequestMessage: NewRequestMessage("AN", config),
	}
}

func (rfp *RequestForPositions) GetMessage(sequenceNumber int) string {
	body := rfp.GetBody()
	var headerAndBody string
	if body != "" {
		header := rfp.getHeader(len(body), sequenceNumber)
		headerAndBody = fmt.Sprintf("%s%s%s%s", header, rfp.delimiter, body, rfp.delimiter)
	} else {
		header := rfp.getHeader(0, sequenceNumber)
		headerAndBody = fmt.Sprintf("%s%s", header, rfp.delimiter)
	}
	trailer := rfp.getTrailer(headerAndBody)
	return fmt.Sprintf("%s%s%s", headerAndBody, trailer, rfp.delimiter)
}

func (rfp *RequestForPositions) GetBody() string {
	var fields []string
	fields = append(fields, fmt.Sprintf("710=%s", rfp.PosReqID))
	if rfp.PosMaintRptID != "" {
		fields = append(fields, fmt.Sprintf("721=%s", rfp.PosMaintRptID))
	}
	return strings.Join(fields, rfp.delimiter)
}
