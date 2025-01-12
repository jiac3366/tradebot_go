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

type BinanceClient struct {
	*base.Client
	ExID string
}

func NewBinanceClient(config *base.Config, accountType BinanceAccountType) *BinanceClient {
	baseURL := BinanceHttpURLs[accountType]
	baseClient := base.NewClient(config.BinanceFutureTestnet.APIKey, config.BinanceFutureTestnet.SecretKey, baseURL)
	return &BinanceClient{
		Client: baseClient,
		ExID:   "binance",
	}
}

// GetTradeList retrieves the account's trade list for a specific symbol
func (c *BinanceClient) GetFApiTradeList(params *TradeListParams) ([]BinanceTrade, error) {
	endpoint := "/fapi/v1/userTrades"

	// Create query parameters
	values := url.Values{}
	values.Add("symbol", params.Symbol)

	// Add optional parameters if they exist
	if params.OrderID != nil {
		values.Add("orderId", strconv.FormatInt(*params.OrderID, 10))
	}
	if params.StartTime != nil {
		values.Add("startTime", strconv.FormatInt(*params.StartTime, 10))
	}
	if params.EndTime != nil {
		values.Add("endTime", strconv.FormatInt(*params.EndTime, 10))
	}
	if params.FromID != nil {
		values.Add("fromId", strconv.FormatInt(*params.FromID, 10))
	}
	if params.Limit != nil {
		values.Add("limit", strconv.Itoa(*params.Limit))
	}
	if params.RecvWindow != nil {
		values.Add("recvWindow", strconv.FormatInt(*params.RecvWindow, 10))
	}

	resp, err := c.fetch(FetchRequest{
		Method:   http.MethodGet,
		Endpoint: endpoint,
		Payload:  &values,
		Signed:   true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get trade list: %w", err)
	}

	// Parse response
	var trades []BinanceTrade
	if err := json.Unmarshal(resp, &trades); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return trades, nil
}

// generateSignature creates HMAC SHA256 signature for the query string
func (c *BinanceClient) generateSignature(queryString string) string {
	mac := hmac.New(sha256.New, []byte(c.SecretKey))
	mac.Write([]byte(queryString))
	return hex.EncodeToString(mac.Sum(nil))
}

type FetchRequest struct {
	Method   string
	Endpoint string
	Payload  *url.Values
	Signed   bool
}

// Fetch sends a request to the API with optional signing
func (c *BinanceClient) fetch(req FetchRequest) ([]byte, error) {
	// Prepare payload with timestamp
	if req.Payload == nil {
		req.Payload = &url.Values{}
	}
	req.Payload.Set("timestamp", strconv.FormatInt(time.Now().UnixMilli(), 10))

	queryString := req.Payload.Encode()

	// Add signature if required
	if req.Signed {
		queryString += "&signature=" + c.generateSignature(queryString)
	}

	// Build and send request
	httpReq, err := c.BuildRequest(req.Method, req.Endpoint, queryString)
	if err != nil {
		return nil, fmt.Errorf("failed to build request: %w", err)
	}

	// Add API key header if signed request
	if req.Signed {
		httpReq.Header.Add("X-MBX-APIKEY", c.ApiKey)
	}
	httpReq.Header.Add("Content-Type", "application/json")
	httpReq.Header.Add("User-Agent", "TradingBot/1.0")

	// Send request and handle response
	var result json.RawMessage
	if err := c.SendRequest(httpReq, &result); err != nil {
		return nil, fmt.Errorf("fetch request failed: %w", err)
	}

	return result, nil
}
