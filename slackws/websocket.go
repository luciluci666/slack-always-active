package slackws

import (
	"encoding/json"
	"fmt"
	"net/http"
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
	lastPingID   int
}

func NewSlackWebSocket(token, cookie string) *SlackWebSocket {
	return &SlackWebSocket{
		token:  token,
		cookie: cookie,
		pingID: 1,
	}
}

func (sws *SlackWebSocket) Connect() error {
	url := sws.reconnectURL
	if url == "" {
		url = fmt.Sprintf("wss://wss-primary.slack.com/?token=%s&sync_desync=1&slack_client=desktop&start_args=%%3Fagent%%3Dclient%%26org_wide_aware%%3Dtrue%%26agent_version%%3D1742552854%%26eac_cache_ts%%3Dtrue%%26cache_ts%%3D0%%26name_tagging%%3Dtrue%%26only_self_subteams%%3Dtrue%%26connect_only%%3Dtrue%%26ms_latest%%3Dtrue&no_query_on_subscribe=1&flannel=3&lazy_channels=1&gateway_server=T05N3TFM0RW-4&batch_presence_aware=1", sws.token)
	}

	// Create custom dialer with cookie header
	dialer := websocket.Dialer{
		HandshakeTimeout: 10 * time.Second,
		Subprotocols:     []string{"slack"},
	}

	// Create custom headers
	headers := http.Header{}
	headers.Add("Cookie", sws.cookie)

	// Connect with custom headers
	conn, _, err := dialer.Dial(url, headers)
	if err != nil {
		return fmt.Errorf("error connecting to websocket: %v", err)
	}

	sws.conn = conn
	return nil
}

func (sws *SlackWebSocket) Close() error {
	if sws.conn != nil {
		return sws.conn.Close()
	}
	return nil
}

func (sws *SlackWebSocket) sendPing() error {
	sws.lastPingID = sws.pingID
	ping := PingMessage{
		Type: "ping",
		ID:   sws.pingID,
	}

	message, err := json.Marshal(ping)
	if err != nil {
		return fmt.Errorf("error marshaling ping message: %v", err)
	}

	if err := sws.conn.WriteMessage(websocket.TextMessage, message); err != nil {
		return fmt.Errorf("error sending ping message: %v", err)
	}

	// Increment ping ID for next ping
	sws.pingID++
	return nil
}

func (sws *SlackWebSocket) ReadMessages() error {
	// Start ping goroutine
	go func() {
		for {
			if err := sws.sendPing(); err != nil {
				fmt.Printf("Error sending ping: %v\n", err)
				return
			}
			time.Sleep(5 * time.Second)
		}
	}()

	for {
		_, message, err := sws.conn.ReadMessage()
		if err != nil {
			return fmt.Errorf("error reading websocket message: %v", err)
		}

		// Try to parse as reconnect message
		var reconnectMsg ReconnectMessage
		if err := json.Unmarshal(message, &reconnectMsg); err == nil && reconnectMsg.Type == "reconnect_url" {
			sws.reconnectURL = reconnectMsg.URL
			fmt.Println("Received new reconnect URL")
			continue
		}

		// Try to parse as pong message
		var pongMsg PongMessage
		if err := json.Unmarshal(message, &pongMsg); err == nil && pongMsg.Type == "pong" {
			if pongMsg.ID == sws.lastPingID {
				fmt.Printf("Received matching pong with ID: %d\n", pongMsg.ID)
			} else {
				fmt.Printf("Warning: Received pong with mismatched ID. Expected: %d, Got: %d\n", sws.lastPingID, pongMsg.ID)
			}
			continue
		}

		fmt.Printf("Received message: %s\n", message)
	}
}
