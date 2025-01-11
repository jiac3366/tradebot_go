package base

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"time"
)

// Client represents a Binance API client
type Client struct {
	baseURL   string
	ApiKey    string
	SecretKey string
	client    *http.Client
}

// NewClient creates a new Binance API client
func NewClient(apiKey, secretKey string, baseURL string) *Client {
	return &Client{
		baseURL:   baseURL,
		ApiKey:    apiKey,
		SecretKey: secretKey,
		client:    &http.Client{Timeout: 10 * time.Second},
	}
}

// buildRequestx creates an HTTP request with the given parameters
func (c *Client) BuildRequest(method, endpoint, queryString string) (*http.Request, error) {
	u, err := url.Parse(c.baseURL + endpoint)
	if err != nil {
		return nil, err
	}

	if queryString != "" {
		u.RawQuery = queryString
	}

	req, err := http.NewRequest(method, u.String(), nil)
	if err != nil {
		return nil, err
	}

	return req, nil
}

// sendRequest sends an HTTP request and decodes the response into the result interface
func (c *Client) SendRequest(req *http.Request, result interface{}) error {
	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Check if the status code indicates an error
	if resp.StatusCode != http.StatusOK {
		var apiErr struct {
			Code    int    `json:"code"`
			Message string `json:"msg"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&apiErr); err != nil {
			return fmt.Errorf("http status %d: failed to decode error response: %w", resp.StatusCode, err)
		}
		return fmt.Errorf("api error: code=%d, message=%s", apiErr.Code, apiErr.Message)
	}

	// Decode the successful response
	if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	return nil
}
