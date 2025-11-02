package ctrader

import (
	"bufio"
	"crypto/tls"
	"context"
	"fmt"
	"net"
	"strings"
	"sync"
	"time"
)

type Client struct {
	host               string
	port               int
	ssl                bool
	delimiter          string
	config             *Config
	conn               net.Conn
	messageSequenceNum int
	isConnected        bool
	mu                 sync.RWMutex
	onConnected        func()
	onDisconnected     func(error)
	onMessage          func(*ResponseMessage)
	messageChan        chan *ResponseMessage
	errorChan          chan error
	stopChan           chan struct{}
	ctx                context.Context
	cancel             context.CancelFunc
	useTLS             bool
	tlsConfig          *tls.Config
}

type ClientOption func(*Client)

func NewClient(host string, port int, config *Config, opts ...ClientOption) *Client {
	ctx, cancel := context.WithCancel(context.Background())
	
	client := &Client{
		host:               host,
		port:               port,
		ssl:                false,
		delimiter:          "\x01",
		config:             config,
		messageSequenceNum: 0,
		messageChan:        make(chan *ResponseMessage, 100),
		errorChan:          make(chan error, 10),
		stopChan:           make(chan struct{}),
		ctx:                ctx,
		cancel:             cancel,
	}
	
	for _, opt := range opts {
		opt(client)
	}
	
	return client
}

func WithSSL(enabled bool) ClientOption {
	return func(c *Client) {
		c.ssl = enabled
	}
}

func WithDelimiter(delimiter string) ClientOption {
	return func(c *Client) {
		c.delimiter = delimiter
	}
}

func (c *Client) Connect() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	if c.isConnected {
		return fmt.Errorf("client is already connected")
	}
	
	address := fmt.Sprintf("%s:%d", c.host, c.port)
	
	var conn net.Conn
	var err error
	
	if c.ssl {
		// Create TLS configuration
		tlsConfig := &tls.Config{
			InsecureSkipVerify: true, // For demo/testing
			MinVersion:         tls.VersionTLS12,
		}
		
		// Connect with TLS
		conn, err = tls.DialWithDialer(&net.Dialer{Timeout: 10 * time.Second}, "tcp", address, tlsConfig)
		if err != nil {
			return fmt.Errorf("failed to connect with TLS to %s: %w", address, err)
		}
	} else {
		// Connect with plain TCP
		conn, err = net.DialTimeout("tcp", address, 10*time.Second)
		if err != nil {
			return fmt.Errorf("failed to connect to %s: %w", address, err)
		}
	}
	
	c.conn = conn
	c.isConnected = true
	c.messageSequenceNum = 0
	
	go c.readMessages()
	
	if c.onConnected != nil {
		go c.onConnected()
	}
	
	return nil
}

func (c *Client) Disconnect() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	if !c.isConnected {
		return nil
	}
	
	c.cancel()
	
	if c.conn != nil {
		c.conn.Close()
	}
	
	c.isConnected = false
	
	if c.onDisconnected != nil {
		go c.onDisconnected(fmt.Errorf("client disconnected"))
	}
	
	return nil
}

func (c *Client) IsConnected() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.isConnected
}

func (c *Client) Send(message interface{}) error {
	c.mu.RLock()
	defer c.mu.RUnlock()
	
	if !c.isConnected {
		return fmt.Errorf("client is not connected")
	}
	
	c.messageSequenceNum++
	var messageString string
	
	switch msg := message.(type) {
	case *LogonRequest:
		messageString = msg.GetMessage(c.messageSequenceNum)
	case *Heartbeat:
		messageString = msg.GetMessage(c.messageSequenceNum)
	case *TestRequest:
		messageString = msg.GetMessage(c.messageSequenceNum)
	case *LogoutRequest:
		messageString = msg.GetMessage(c.messageSequenceNum)
	case *OrderMsg:
		messageString = msg.GetMessage(c.messageSequenceNum)
	case *OrderCancelRequest:
		messageString = msg.GetMessage(c.messageSequenceNum)
	case *MarketDataRequest:
		messageString = msg.GetMessage(c.messageSequenceNum)
	case *SecurityListRequest:
		messageString = msg.GetMessage(c.messageSequenceNum)
	case *RequestForPositions:
		messageString = msg.GetMessage(c.messageSequenceNum)
	default:
		return fmt.Errorf("unsupported message type")
	}
	
	if !strings.HasSuffix(messageString, c.delimiter) {
		messageString += c.delimiter
	}
	
	_, err := c.conn.Write([]byte(messageString))
	if err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}
	
	return nil
}

func (c *Client) readMessages() {
	defer func() {
		if r := recover(); r != nil {
			c.errorChan <- fmt.Errorf("panic in readMessages: %v", r)
		}
	}()
	
	scanner := bufio.NewScanner(c.conn)
	var currentMessage strings.Builder
	
	for {
		select {
		case <-c.ctx.Done():
			return
		default:
			if !scanner.Scan() {
				if err := scanner.Err(); err != nil {
					c.errorChan <- fmt.Errorf("scanner error: %w", err)
				}
				c.handleDisconnection()
				return
			}
			
			data := scanner.Text()
			currentMessage.WriteString(data)
			
			messageStr := currentMessage.String()
			if strings.Contains(messageStr, "10=") && strings.HasSuffix(data, c.delimiter) {
				responseMessage := NewResponseMessage(messageStr, c.delimiter)
				
				select {
				case c.messageChan <- responseMessage:
				case <-c.ctx.Done():
					return
				default:
				}
				
				currentMessage.Reset()
			}
		}
	}
}

func (c *Client) handleDisconnection() {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	if c.isConnected {
		c.isConnected = false
		
		if c.onDisconnected != nil {
			go c.onDisconnected(fmt.Errorf("connection lost"))
		}
	}
}

func (c *Client) SetConnectedCallback(callback func()) {
	c.onConnected = callback
}

func (c *Client) SetDisconnectedCallback(callback func(error)) {
	c.onDisconnected = callback
}

func (c *Client) SetMessageCallback(callback func(*ResponseMessage)) {
	c.onMessage = callback
}

func (c *Client) Messages() <-chan *ResponseMessage {
	return c.messageChan
}

func (c *Client) Errors() <-chan error {
	return c.errorChan
}

func (c *Client) ChangeMessageSequenceNumber(newSeqNum int) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.messageSequenceNum = newSeqNum
}

func (c *Client) GetMessageSequenceNumber() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.messageSequenceNum
}
