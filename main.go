package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/lucy/slack-always-active/logger"
	"github.com/lucy/slack-always-active/slackws"
)

type UserBootResponse struct {
	Ok      bool   `json:"ok"`
	Error   string `json:"error,omitempty"`
	Self    User   `json:"self"`
	Team    Team   `json:"team"`
	CacheTs int    `json:"cache_ts"`
}

type User struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	RealName string `json:"real_name"`
	Email    string `json:"email"`
}

type Team struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func checkSlackStatus(token, cookie string) (*UserBootResponse, error) {
	url := "https://3dsellers.slack.com/api/client.userBoot"

	// Create form data
	formData := strings.NewReader(fmt.Sprintf("token=%s", token))

	// Create request
	req, err := http.NewRequest("POST", url, formData)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %v", err)
	}

	// Add headers
	req.Header.Add("Cookie", cookie)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	// Send request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error sending request: %v", err)
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response: %v", err)
	}

	// Parse response
	var userBoot UserBootResponse
	if err := json.Unmarshal(body, &userBoot); err != nil {
		return nil, fmt.Errorf("error parsing response: %v", err)
	}

	// Check if response is ok
	if !userBoot.Ok {
		return nil, fmt.Errorf("slack API error: %s", userBoot.Error)
	}

	return &userBoot, nil
}

func main() {
	// Initialize logger with local logs directory
	logPath := "logs/slack-always-active.log"
	if err := logger.Init(logPath); err != nil {
		fmt.Printf("Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	defer logger.Close()

	// Load environment variables
	if err := godotenv.Load(); err != nil {
		logger.Printf("Warning: .env file not found: %v\n", err)
	}

	// Get credentials from environment variables
	token := os.Getenv("SLACK_TOKEN")
	cookie := os.Getenv("SLACK_COOKIE")

	if token == "" || cookie == "" {
		logger.Error("SLACK_TOKEN and SLACK_COOKIE environment variables must be set")
		os.Exit(1)
	}

	// Check Slack status
	logger.Printf("Checking Slack status...")
	userBoot, err := checkSlackStatus(token, cookie)
	if err != nil {
		logger.Error("Error checking Slack status: %v\n", err)
		os.Exit(1)
	}

	logger.Printf("Connected as: %s\n", userBoot.Self.RealName)

	// Create and connect WebSocket
	logger.Printf("Connecting to Slack WebSocket...")
	ws := slackws.NewSlackWebSocket(token, cookie)
	for {
		if err := ws.Connect(); err != nil {
			logger.Error("WebSocket connection error: %v\n", err)
			logger.Printf("Reconnecting in 5 seconds...\n")
			time.Sleep(5 * time.Second)
			continue
		}

		if err := ws.ReadMessages(); err != nil {
			logger.Error("WebSocket read error: %v\n", err)
			ws.Close()
			logger.Printf("Reconnecting in 5 seconds...\n")
			time.Sleep(5 * time.Second)
			continue
		}
	}
}
