package cache

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

type Cache struct {
	WebSocketURL string `json:"websocket_url"`
	mu           sync.RWMutex
	cacheFile    string
}

func NewCache(cacheDir string) (*Cache, error) {
	// Create cache directory if it doesn't exist
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create cache directory: %v", err)
	}

	cacheFile := filepath.Join(cacheDir, "websocket_cache.json")
	cache := &Cache{
		cacheFile: cacheFile,
	}

	// Load existing cache if it exists
	if err := cache.load(); err != nil {
		// If file doesn't exist, that's okay
		if !os.IsNotExist(err) {
			return nil, fmt.Errorf("failed to load cache: %v", err)
		}
	}

	return cache, nil
}

func (c *Cache) load() error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	data, err := os.ReadFile(c.cacheFile)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, c)
}

func (c *Cache) save() error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal cache: %v", err)
	}

	return os.WriteFile(c.cacheFile, data, 0644)
}

func (c *Cache) GetWebSocketURL() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.WebSocketURL
}

func (c *Cache) SetWebSocketURL(url string) error {
	c.mu.Lock()
	c.WebSocketURL = url
	c.mu.Unlock()
	return c.save()
}
