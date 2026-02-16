package session

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// ConsoleClient is an HTTP client for the console server API.
type ConsoleClient struct {
	baseURL string
	http    *http.Client
}

// NewConsoleClient creates a client pointed at the given base URL.
func NewConsoleClient(baseURL string) *ConsoleClient {
	return &ConsoleClient{
		baseURL: baseURL,
		http: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// DefaultConsoleClient creates a client for localhost on the given port.
func DefaultConsoleClient(port int) *ConsoleClient {
	return NewConsoleClient(fmt.Sprintf("http://localhost:%d", port))
}

// BaseURL returns the base URL of the console server.
func (c *ConsoleClient) BaseURL() string {
	return c.baseURL
}

// Post sends a JSON POST request to the given path.
func (c *ConsoleClient) Post(path string, body any) (*http.Response, error) {
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(body); err != nil {
		return nil, fmt.Errorf("encode body: %w", err)
	}
	return c.http.Post(c.baseURL+path, "application/json", &buf)
}

// Get sends a GET request to the given path.
func (c *ConsoleClient) Get(path string) (*http.Response, error) {
	return c.http.Get(c.baseURL + path)
}
