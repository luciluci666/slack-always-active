package slackws

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type ReconnectMessage struct {
	Type string `json:"type"`
	URL  string `json:"url"`
}

type PingMessage struct {
	Type string `json:"type"`
	ID   int    `json:"id"`
}

type PongMessage struct {
	Type string `json:"type"`
	ID   int    `json:"reply_to"`
}

type SlackWebSocket struct {
	conn         *websocket.Conn
	token        string
	cookie       string
	reconnectURL string
	pingID       int
	mu           sync.Mutex
	stopChan     chan struct{}
}

func NewSlackWebSocket(token, cookie string) *SlackWebSocket {
	return &SlackWebSocket{
		token:    token,
		cookie:   cookie,
		stopChan: make(chan struct{}),
	}
}

func (s *SlackWebSocket) Connect() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Create WebSocket connection
	url := "wss://3dsellers.slack.com/api/rtm.connect"
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return fmt.Errorf("error creating request: %v", err)
	}

	// Add headers
	req.Header.Add("Cookie", s.cookie)
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", s.token))

	// Create WebSocket connection
	dialer := websocket.Dialer{
		EnableCompression: true,
		HandshakeTimeout:  10 * time.Second,
	}

	conn, _, err := dialer.Dial(req.URL.String(), nil)
	if err != nil {
		return fmt.Errorf("error connecting to WebSocket: %v", err)
	}

	s.conn = conn
	s.pingID = 0
	return nil
}

func (s *SlackWebSocket) Close() {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Signal all goroutines to stop
	close(s.stopChan)

	// Close the WebSocket connection
	if s.conn != nil {
		s.conn.Close()
		s.conn = nil
	}
}

func (s *SlackWebSocket) sendPing() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.conn == nil {
		return fmt.Errorf("websocket connection is closed")
	}

	s.pingID++
	pingMsg := fmt.Sprintf(`{"type":"ping","id":%d}`, s.pingID)
	return s.conn.WriteMessage(websocket.TextMessage, []byte(pingMsg))
}

func (s *SlackWebSocket) ReadMessages() error {
	pingTicker := time.NewTicker(30 * time.Second)
	defer pingTicker.Stop()

	// Start ping goroutine
	go func() {
		for {
			select {
			case <-s.stopChan:
				return
			case <-pingTicker.C:
				if err := s.sendPing(); err != nil {
					log.Printf("Error sending ping: %v\n", err)
					return
				}
			}
		}
	}()

	// Read messages
	for {
		select {
		case <-s.stopChan:
			return nil
		default:
			s.mu.Lock()
			if s.conn == nil {
				s.mu.Unlock()
				return fmt.Errorf("websocket connection is closed")
			}
			s.mu.Unlock()

			_, message, err := s.conn.ReadMessage()
			if err != nil {
				return fmt.Errorf("error reading message: %v", err)
			}

			// Parse message
			var msg map[string]interface{}
			if err := json.Unmarshal(message, &msg); err != nil {
				log.Printf("Error parsing message: %v\n", err)
				continue
			}

			// Handle different message types
			switch msg["type"] {
			case "pong":
				if id, ok := msg["id"].(float64); ok {
					if int(id) != s.pingID {
						log.Printf("Ping ID mismatch: expected %d, got %d\n", s.pingID, int(id))
					}
				}
			case "reconnect_url":
				// Handle reconnect URL if needed
			}
		}
	}
}
