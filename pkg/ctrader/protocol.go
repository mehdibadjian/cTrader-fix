package ctrader

import (
	"fmt"
	"strconv"
	"strings"
)

type Protocol struct {
	delimiter string
}

func NewProtocol(delimiter string) *Protocol {
	if delimiter == "" {
		delimiter = "\x01"
	}
	return &Protocol{
		delimiter: delimiter,
	}
}

func (p *Protocol) ValidateMessage(message string) error {
	if message == "" {
		return fmt.Errorf("message is empty")
	}
	
	fields := p.parseFields(message)
	
	if _, exists := fields[8]; !exists {
		return fmt.Errorf("missing BeginString field (8)")
	}
	
	if _, exists := fields[9]; !exists {
		return fmt.Errorf("missing BodyLength field (9)")
	}
	
	if _, exists := fields[35]; !exists {
		return fmt.Errorf("missing MsgType field (35)")
	}
	
	if _, exists := fields[10]; !exists {
		return fmt.Errorf("missing Checksum field (10)")
	}
	
	if err := p.validateChecksum(message); err != nil {
		return fmt.Errorf("checksum validation failed: %w", err)
	}
	
	return nil
}

func (p *Protocol) parseFields(message string) map[int][]string {
	fields := make(map[int][]string)
	
	parts := strings.Split(message, p.delimiter)
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
	
	return fields
}

func (p *Protocol) validateChecksum(message string) error {
	checksumIndex := strings.LastIndex(message, p.delimiter+"10=")
	if checksumIndex == -1 {
		return fmt.Errorf("checksum field not found")
	}
	
	checksumStart := checksumIndex + 4
	checksumEnd := strings.Index(message[checksumStart:], p.delimiter)
	if checksumEnd == -1 {
		checksumEnd = len(message) - checksumStart
	} else {
		checksumEnd += checksumStart
	}
	
	checksumStr := message[checksumStart:checksumEnd]
	checksum, err := strconv.Atoi(checksumStr)
	if err != nil {
		return fmt.Errorf("invalid checksum format: %s", checksumStr)
	}
	
	// Calculate checksum on message up to and including the delimiter before checksum field
	messageBody := message[:checksumIndex+1]
	calculatedChecksum := p.calculateChecksum(messageBody)
	
	if calculatedChecksum != checksum {
		return fmt.Errorf("checksum mismatch: expected %d, got %d", calculatedChecksum, checksum)
	}
	
	return nil
}

func (p *Protocol) calculateChecksum(message string) int {
	checksum := 0
	for _, b := range []byte(message) {
		checksum += int(b)
	}
	return checksum % 256
}

func (p *Protocol) GetFieldNames() map[int]string {
	return map[int]string{
		8:   "BeginString",
		9:   "BodyLength",
		35:  "MsgType",
		49:  "SenderCompID",
		50:  "SenderSubID",
		56:  "TargetCompID",
		57:  "TargetSubID",
		34:  "MsgSeqNum",
		52:  "SendingTime",
		10:  "CheckSum",
		98:  "EncryptMethod",
		108: "HeartBtInt",
		141: "ResetSeqNumFlag",
		553: "Username",
		554: "Password",
		112: "TestReqID",
		7:   "BeginSeqNo",
		16:  "EndSeqNo",
		123: "GapFillFlag",
		36:  "NewSeqNo",
		262: "MDReqID",
		263: "SubscriptionRequestType",
		264: "MarketDepth",
		265: "MDUpdateType",
		267: "NoMDEntryTypes",
		269: "MDEntryType",
		146: "NoRelatedSym",
		55:  "Symbol",
		11:  "ClOrdID",
		54:  "Side",
		60:  "TransactTime",
		38:  "OrderQty",
		40:  "OrdType",
		44:  "Price",
		99:  "StopPx",
		126: "ExpireTime",
		721: "PosMaintRptID",
		494: "Designation",
		584: "MassStatusReqID",
		585: "MassStatusReqType",
		225: "IssueDate",
		710: "PosReqID",
		37:  "OrderID",
		41:  "OrigClOrdID",
		320: "SecurityReqID",
		559: "SecurityListRequestType",
	}
}

func (p *Protocol) GetMessageTypeName() map[string]string {
	return map[string]string{
		"0":  "Heartbeat",
		"1":  "TestRequest",
		"2":  "ResendRequest",
		"3":  "Reject",
		"4":  "SequenceReset",
		"5":  "Logout",
		"8":  "BusinessMessageReject",
		"A":  "Logon",
		"D":  "NewOrderSingle",
		"F":  "OrderCancelRequest",
		"G":  "OrderCancelReplaceRequest",
		"H":  "OrderStatusRequest",
		"J":  "AllocationInstruction",
		"K":  "AllocationInstructionAck",
		"L":  "AllocationReport",
		"V":  "MarketDataRequest",
		"W":  "MarketDataSnapshotFullRefresh",
		"X":  "MarketDataIncrementalRefresh",
		"Y":  "MarketDataRequestReject",
		"AF": "OrderMassStatusRequest",
		"AN": "RequestForPositions",
		"AO": "PositionReport",
		"AP": "TradeCaptureReportRequest",
		"AR": "TradeCaptureReport",
		"x":  "SecurityListRequest",
		"y":  "SecurityList",
		"z":  "SecurityListResponse",
	}
}

func (p *Protocol) FormatMessage(message string) string {
	fields := p.parseFields(message)
	fieldNames := p.GetFieldNames()
	messageTypes := p.GetMessageTypeName()
	
	var result strings.Builder
	
	if msgTypeValues, exists := fields[35]; exists && len(msgTypeValues) > 0 {
		if msgTypeName, exists := messageTypes[msgTypeValues[0]]; exists {
			result.WriteString(fmt.Sprintf("Message Type: %s (%s)\n", msgTypeName, msgTypeValues[0]))
		} else {
			result.WriteString(fmt.Sprintf("Message Type: %s\n", msgTypeValues[0]))
		}
	}
	
	for fieldNum, values := range fields {
		fieldName := fmt.Sprintf("Field%d", fieldNum)
		if name, exists := fieldNames[fieldNum]; exists {
			fieldName = name
		}
		
		for i, value := range values {
			if len(values) > 1 {
				result.WriteString(fmt.Sprintf("%s[%d]: %s\n", fieldName, i, value))
			} else {
				result.WriteString(fmt.Sprintf("%s: %s\n", fieldName, value))
			}
		}
	}
	
	return result.String()
}
