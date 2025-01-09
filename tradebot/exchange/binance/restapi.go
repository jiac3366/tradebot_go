package binance

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"tradebot_go/tradebot/base"
)

// Trade represents a single trade from the account trade list
type BinanceTrade struct {
	Buyer           bool   `json:"buyer"`
	Commission      string `json:"commission"`
	CommissionAsset string `json:"commissionAsset"`
	ID              int64  `json:"id"`
	Maker           bool   `json:"maker"`
	OrderID         int64  `json:"orderId"`
	Price           string `json:"price"`
	Qty             string `json:"qty"`
	QuoteQty        string `json:"quoteQty"`
	RealizedPnl     string `json:"realizedPnl"`
	Side            string `json:"side"`
	PositionSide    string `json:"positionSide"`
	Symbol          string `json:"symbol"`
	Time            int64  `json:"time"`
}

// TradeListParams represents the parameters for GetTradeList
type TradeListParams struct {
	Symbol     string
	OrderID    *int64
	StartTime  *int64
	EndTime    *int64
	FromID     *int64
	Limit      *int
	RecvWindow *int64
}

// Client represents a Binance API client
type Client struct {
	baseURL   string
	apiKey    string
	secretKey string
	client    *http.Client
}

// NewClient creates a new Binance API client
func NewClient(apiKey, secretKey string) *Client {
	return &Client{
		baseURL:   "https://fapi.binance.com",
		apiKey:    apiKey,
		secretKey: secretKey,
		client:    &http.Client{Timeout: 10 * time.Second},
	}
}

// NewClientWithConfig creates a new Binance API client using configuration from file
func NewClientWithConfig(config *base.Config) (*Client, error) {
	return NewClient(config.BinanceFutureTestnet.APIKey,
		config.BinanceFutureTestnet.SecretKey), nil
}

// GetTradeList retrieves the account's trade list for a specific symbol
func (c *Client) GetTradeList(params TradeListParams) ([]BinanceTrade, error) {
	endpoint := "/fapi/v1/userTrades"

	// Build query parameters
	query := url.Values{}
	query.Add("symbol", params.Symbol)
	query.Add("timestamp", strconv.FormatInt(time.Now().UnixMilli(), 10))

	if params.OrderID != nil {
		query.Add("orderId", strconv.FormatInt(*params.OrderID, 10))
	}
	if params.StartTime != nil {
		query.Add("startTime", strconv.FormatInt(*params.StartTime, 10))
	}
	if params.EndTime != nil {
		query.Add("endTime", strconv.FormatInt(*params.EndTime, 10))
	}
	if params.FromID != nil {
		query.Add("fromId", strconv.FormatInt(*params.FromID, 10))
	}
	if params.Limit != nil {
		query.Add("limit", strconv.Itoa(*params.Limit))
	}
	if params.RecvWindow != nil {
		query.Add("recvWindow", strconv.FormatInt(*params.RecvWindow, 10))
	}

	// Sign the request
	signature := c.generateSignature(query.Encode())
	query.Add("signature", signature)

	// Create request
	req, err := c.buildRequest(http.MethodGet, endpoint, query)
	if err != nil {
		return nil, fmt.Errorf("failed to build request: %w", err)
	}

	// Add API key header
	req.Header.Add("X-MBX-APIKEY", c.apiKey)

	// Execute request
	var trades []BinanceTrade
	if err := c.sendRequest(req, &trades); err != nil {
		return nil, fmt.Errorf("failed to get trade list: %w", err)
	}

	return trades, nil
}

// generateSignature creates HMAC SHA256 signature for the query string
func (c *Client) generateSignature(queryString string) string {
	mac := hmac.New(sha256.New, []byte(c.secretKey))
	mac.Write([]byte(queryString))
	return hex.EncodeToString(mac.Sum(nil))
}

// buildRequest creates an HTTP request with the given parameters
func (c *Client) buildRequest(method, endpoint string, query url.Values) (*http.Request, error) {
	u, err := url.Parse(c.baseURL + endpoint)
	if err != nil {
		return nil, err
	}

	if query != nil {
		u.RawQuery = query.Encode()
	}

	req, err := http.NewRequest(method, u.String(), nil)
	if err != nil {
		return nil, err
	}

	return req, nil
}

// sendRequest sends an HTTP request and decodes the response into the result interface
func (c *Client) sendRequest(req *http.Request, result interface{}) error {
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
