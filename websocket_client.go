package client

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type ClientWebSocketClient struct {
	apiKey               string
	wsUrl                string
	accessToken          string
	conn                 *websocket.Conn
	mu                   sync.Mutex
	isConnected          bool
	eventHandlers        map[string][]func(interface{})
	autoReconnect        bool
	reconnectInterval    time.Duration
	maxReconnectAttempts int
}

func NewClientWebSocketClient(apiKey, wsUrl string, autoReconnect bool, interval time.Duration, maxAttempts int) *ClientWebSocketClient {
	return &ClientWebSocketClient{
		apiKey:               apiKey,
		wsUrl:                wsUrl,
		autoReconnect:        autoReconnect,
		reconnectInterval:    interval,
		maxReconnectAttempts: maxAttempts,
		eventHandlers:        make(map[string][]func(interface{})),
	}
}

func (c *ClientWebSocketClient) Connect() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.isConnected {
		return nil
	}

	// Match JS SDK format: /ws?token=...
	url := fmt.Sprintf("%s/ws?token=%s", c.wsUrl, c.accessToken)

	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		return err
	}

	c.conn = conn
	c.isConnected = true

	go c.listen()

	return nil
}

func (c *ClientWebSocketClient) Disconnect() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.conn != nil {
		c.conn.Close()
		c.isConnected = false
	}
}

func (c *ClientWebSocketClient) listen() {
	defer func() {
		c.mu.Lock()
		c.isConnected = false
		c.mu.Unlock()
		if c.autoReconnect {
			// Simple reconnect logic here for now
			time.Sleep(c.reconnectInterval)
			c.Connect()
		}
	}()

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			log.Println("ws read error:", err)
			return
		}

		// First try to parse as generic event structure
		var eventData struct {
			Type string      `json:"type"`
			Data interface{} `json:"data"`
		}

		if err := json.Unmarshal(message, &eventData); err != nil {
			continue
		}

		// Handle different message types like JS SDK
		if eventData.Type == "message" {
			// Server sends: {type: 'message', source: '...', message: '...', hash: '...', conversationID: '...', timestamp: ...}
			// Parse the full message structure
			var msgData map[string]interface{}
			if err := json.Unmarshal(message, &msgData); err != nil {
				continue
			}

			// Transform to match JS SDK format
			transformedMsg := make(map[string]interface{})
			for k, v := range msgData {
				transformedMsg[k] = v
			}

			// Add aliases for convenience
			if source, ok := msgData["source"]; ok {
				transformedMsg["senderId"] = source
			}
			if convID, ok := msgData["conversationID"]; ok {
				transformedMsg["conversationId"] = convID
			}
			if msg, ok := msgData["message"]; ok {
				transformedMsg["content"] = msg
			}

			// Emit as message.create event
			c.mu.Lock()
			handlers := c.eventHandlers["message.create"]
			c.mu.Unlock()

			for _, handler := range handlers {
				go handler(transformedMsg)
			}

			// Also emit conversation-specific event
			if convID, ok := transformedMsg["conversationId"].(string); ok {
				eventKey := fmt.Sprintf("conversation:%s:message.create", convID)
				c.mu.Lock()
				convHandlers := c.eventHandlers[eventKey]
				c.mu.Unlock()

				for _, handler := range convHandlers {
					go handler(transformedMsg)
				}
			}
		} else if eventData.Type == "event" {
			// Handle standard event format
			c.mu.Lock()
			handlers := c.eventHandlers[eventData.Type]
			c.mu.Unlock()

			for _, handler := range handlers {
				go handler(eventData.Data)
			}
		} else if eventData.Type == "subscribed" {
			// Handle subscription confirmation
			log.Printf("Subscribed to events")
		} else if eventData.Type == "pong" {
			// Handle pong
		} else {
			// Generic handler for other event types
			c.mu.Lock()
			handlers := c.eventHandlers[eventData.Type]
			c.mu.Unlock()

			for _, handler := range handlers {
				go handler(eventData.Data)
			}
		}
	}
}

func (c *ClientWebSocketClient) Send(data interface{}) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.isConnected {
		return fmt.Errorf("websocket not connected")
	}

	return c.conn.WriteJSON(data)
}

func (c *ClientWebSocketClient) On(event string, callback func(interface{})) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.eventHandlers[event] = append(c.eventHandlers[event], callback)
}

func (c *ClientWebSocketClient) SetAccessToken(token string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.accessToken = token
}

func (c *ClientWebSocketClient) IsConnectedStatus() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.isConnected
}

func (c *ClientWebSocketClient) Subscribe(subscription interface{}) error {
	return c.Send(map[string]interface{}{
		"type": "subscribe",
		"data": subscription,
	})
}

func (c *ClientWebSocketClient) Emit(event string, data interface{}) {
	c.mu.Lock()
	handlers := c.eventHandlers[event]
	c.mu.Unlock()

	for _, handler := range handlers {
		go handler(data)
	}
}
