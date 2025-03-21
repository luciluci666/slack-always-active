package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"github.com/lucy/slack-always-active/cache"
	"github.com/lucy/slack-always-active/logger"
	"github.com/lucy/slack-always-active/schedule"
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

	fmt.Printf("Connected as %s\n", userBoot.Self.RealName)

	return &userBoot, nil
}

func formatTimeWithOffset(t time.Time, offset int) string {
	// Adjust the time by the GMT offset
	adjustedTime := t.Add(time.Duration(offset) * time.Hour)

	// Format the time with the offset
	offsetStr := fmt.Sprintf("GMT%+d", offset)
	return fmt.Sprintf("%s (%s)", adjustedTime.Format("2006-01-02 15:04:05"), offsetStr)
}

func main() {
	// Initialize logger
	if err := logger.Init("logs/slack-always-active.log"); err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer logger.Close()

	// this requires rest token
	// _, err := checkSlackStatus(os.Getenv("SLACK_TOKEN"), os.Getenv("SLACK_COOKIE"))
	// if err != nil {
	// 	logger.Error("Failed to check Slack status:", err)
	// 	os.Exit(1)
	// }

	// Load environment variables
	if err := godotenv.Load(); err != nil {
		logger.Warn("Warning: .env file not found")
	}

	// Get required environment variables
	token := os.Getenv("SLACK_TOKEN")
	cookie := os.Getenv("SLACK_COOKIE")

	if token == "" || cookie == "" {
		logger.Error("Error: SLACK_TOKEN and SLACK_COOKIE must be set in .env file")
		os.Exit(1)
	}

	// Initialize cache
	cache, err := cache.NewCache("cache/cache")
	if err != nil {
		logger.Error("Failed to initialize cache:", err)
		os.Exit(1)
	}

	// Initialize schedule
	schedule, err := schedule.NewSchedule()
	if err != nil {
		logger.Error("Failed to initialize schedule:", err)
		os.Exit(1)
	}

	// Create context that can be cancelled
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Create WebSocket instance
	ws := slackws.NewSlackWebSocket(token, cookie, cache)

	// Start a goroutine to handle signals
	go func() {
		<-sigChan
		logger.Info("Received shutdown signal, cleaning up...")
		cancel()
	}()

	// Start a goroutine to check working hours and manage WebSocket connection
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				if schedule.IsWorkingTime() {
					// If we're in working hours, ensure WebSocket is connected
					if !ws.IsConnected() {
						logger.Info("Working hours started, connecting to Slack...")
						if err := ws.Connect(); err != nil {
							logger.Error("Failed to connect to Slack:", err)
							continue
						}
						logger.Info("Successfully connected to Slack")
					}
				} else {
					// If we're outside working hours, disconnect WebSocket
					if ws.IsConnected() {
						logger.Info("Working hours ended, disconnecting from Slack...")
						ws.Disconnect()
						logger.Info("Disconnected from Slack")
					}
					nextTime := schedule.GetNextWorkingTime()
					logger.Info(fmt.Sprintf("Outside working hours. Next working time: %s (GMT+%d)", nextTime.Format("2006-01-02 15:04:05"), schedule.GetOffset()))
				}
				// Check every minute
				time.Sleep(time.Minute)
			}
		}
	}()

	// Start reading messages in a separate goroutine
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				if ws.IsConnected() {
					if err := ws.ReadMessages(); err != nil {
						logger.Error("Error reading messages:", err)
						// Don't disconnect here, let the working hours check handle reconnection
					}
				}
				time.Sleep(time.Second)
			}
		}
	}()

	// Wait for context cancellation
	<-ctx.Done()

	// Cleanup
	if ws.IsConnected() {
		ws.Disconnect()
	}
	logger.Info("Application shutdown complete")
}
