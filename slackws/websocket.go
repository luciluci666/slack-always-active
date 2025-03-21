package slackws

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/lucy/slack-always-active/cache"
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
	conn        *websocket.Conn
	token       string
	cookie      string
	pingID      int
	lastPingID  int
	mu          sync.Mutex
	stopChan    chan struct{}
	closed      bool
	isConnected bool
	cache       *cache.Cache
}

func NewSlackWebSocket(token, cookie string, cache *cache.Cache) *SlackWebSocket {
	return &SlackWebSocket{
		token:       token,
		cookie:      cookie,
		pingID:      1,
		stopChan:    make(chan struct{}),
		closed:      false,
		isConnected: false,
		cache:       cache,
	}
}

func (s *SlackWebSocket) Connect() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Reset state for new connection
	s.pingID = 1
	s.lastPingID = 0
	s.closed = false
	s.isConnected = false
	s.stopChan = make(chan struct{})

	// Use default WebSocket URL
	url := fmt.Sprintf("wss://wss-primary.slack.com/?token=%s&sync_desync=1&slack_client=desktop&start_args=%%3Fagent%%3Dclient%%26org_wide_aware%%3Dtrue%%26agent_version%%3D1742552854%%26eac_cache_ts%%3Dtrue%%26cache_ts%%3D0%%26name_tagging%%3Dtrue%%26only_self_subteams%%3Dtrue%%26connect_only%%3Dtrue%%26ms_latest%%3Dtrue&no_query_on_subscribe=1&flannel=3&lazy_channels=1&gateway_server=T05N3TFM0RW-4&batch_presence_aware=1", s.token)

	// Create custom dialer with cookie header
	dialer := websocket.Dialer{
		EnableCompression: true,
		HandshakeTimeout:  10 * time.Second,
		Subprotocols:      []string{"slack"},
	}

	// Create custom headers
	headers := http.Header{}
	headers.Add("Cookie", s.cookie)

	// Connect with custom headers
	conn, _, err := dialer.Dial(url, headers)
	if err != nil {
		return fmt.Errorf("error connecting to websocket: %v", err)
	}

	s.conn = conn
	s.isConnected = true
	return nil
}

func (s *SlackWebSocket) Close() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed {
		return
	}

	// Signal all goroutines to stop
	close(s.stopChan)
	s.closed = true
	s.isConnected = false

	// Close the WebSocket connection
	if s.conn != nil {
		s.conn.Close()
		s.conn = nil
	}
}

func (s *SlackWebSocket) IsConnected() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.isConnected
}

func (s *SlackWebSocket) sendPing() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.conn == nil || s.closed || !s.isConnected {
		return fmt.Errorf("websocket connection is closed")
	}

	s.lastPingID = s.pingID
	ping := PingMessage{
		Type: "ping",
		ID:   s.pingID,
	}

	message, err := json.Marshal(ping)
	if err != nil {
		return fmt.Errorf("error marshaling ping message: %v", err)
	}

	if err := s.conn.WriteMessage(websocket.TextMessage, message); err != nil {
		s.isConnected = false
		return fmt.Errorf("error sending ping message: %v", err)
	}

	// Increment ping ID for next ping
	s.pingID++
	return nil
}

func (s *SlackWebSocket) ReadMessages() error {
	pingTicker := time.NewTicker(5 * time.Second)
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
			if s.conn == nil || s.closed || !s.isConnected {
				s.mu.Unlock()
				return fmt.Errorf("websocket connection is closed")
			}
			s.mu.Unlock()

			_, message, err := s.conn.ReadMessage()
			if err != nil {
				s.mu.Lock()
				s.isConnected = false
				s.mu.Unlock()
				return fmt.Errorf("error reading message: %v", err)
			}

			// Try to parse as pong message
			var pongMsg PongMessage
			if err := json.Unmarshal(message, &pongMsg); err == nil && pongMsg.Type == "pong" {
				if pongMsg.ID == s.lastPingID {
					log.Printf("Received matching pong with ID: %d\n", pongMsg.ID)
				} else {
					log.Printf("Warning: Received pong with mismatched ID. Expected: %d, Got: %d\n", s.lastPingID, pongMsg.ID)
				}
				continue
			}

			// Try to parse as reconnect message
			var reconnectMsg ReconnectMessage
			if err := json.Unmarshal(message, &reconnectMsg); err == nil && reconnectMsg.Type == "reconnect_url" {
				log.Printf("Received new reconnect URL\n")
				s.cache.SetWebSocketURL(reconnectMsg.URL)
				continue
			}

			// Log other messages for debugging
			log.Printf("Received message: %s\n", message)
		}
	}
}

// Disconnect closes the WebSocket connection
func (ws *SlackWebSocket) Disconnect() {
	ws.mu.Lock()
	defer ws.mu.Unlock()

	if ws.conn != nil {
		ws.conn.Close()
		ws.conn = nil
		ws.isConnected = false
	}
}
