package binance

import (
	"crypto/hmac"
	"crypto/sha256"

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
	baseClient *base.Client
}

func NewBinanceClient(config *base.Config, accountType BinanceAccountType) *BinanceClient {
	baseURL := BinanceHttpURLs[accountType]
	baseClient := base.NewClient(config.BinanceFutureTestnet.APIKey, config.BinanceFutureTestnet.SecretKey, baseURL)
	return &BinanceClient{
		baseClient: baseClient,
	}
}

// GetTradeList retrieves the account's trade list for a specific symbol
func (c *BinanceClient) GetFApiTradeList(params TradeListParams) ([]BinanceTrade, error) {
	endpoint := "/fapi/v1/userTrades"

	// Create query parameters
	values := url.Values{}
	values.Set("symbol", params.Symbol)
	values.Set("timestamp", strconv.FormatInt(time.Now().UnixMilli(), 10))

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

	// Generate signature from the query string
	queryString := values.Encode()
	queryString += "&signature=" + c.generateSignature(queryString)

	// Create request
	req, err := c.baseClient.BuildRequest(http.MethodGet, endpoint, queryString)
	if err != nil {
		return nil, fmt.Errorf("failed to build request: %w", err)
	}

	// Add API key header
	req.Header.Set("X-MBX-APIKEY", c.baseClient.ApiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "TradingBot/1.0")

	// Execute request
	var trades []BinanceTrade
	if err := c.baseClient.SendRequest(req, &trades); err != nil {
		return nil, fmt.Errorf("failed to get trade list: %w", err)
	}

	return trades, nil
}

// generateSignature creates HMAC SHA256 signature for the query string
func (c *BinanceClient) generateSignature(queryString string) string {
	mac := hmac.New(sha256.New, []byte(c.baseClient.SecretKey))
	mac.Write([]byte(queryString))
	return fmt.Sprintf("%x", mac.Sum(nil))
}
