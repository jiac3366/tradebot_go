package constants

type Order struct {
	Exchange        string
	Symbol          string
	Status          OrderStatus
	Id              string
	ClientOrderId   string
	Timestamp       int64
	Type            OrderType
	Side            OrderSide
	TimeInForce     TimeInForce
	Price           float64
	Average         float64
	LastFilledPrice float64
	Amount          float64
	Filled          float64
	LastFilled      float64
	Remaining       float64
	Fee             float64
	FeeCurrency     string
	Cost            float64
	CumCost         float64
	ReduceOnly      bool
	PositionSide    PositionSide
	Success         bool
}

type OrderStatus string

const (
	// LOCAL
	OrderStatusFailed OrderStatus = "FAILED"

	// IN-FLOW
	OrderStatusPending   OrderStatus = "PENDING"
	OrderStatusCanceling OrderStatus = "CANCELING"

	// OPEN
	OrderStatusAccepted        OrderStatus = "ACCEPTED"
	OrderStatusPartiallyFilled OrderStatus = "PARTIALLY_FILLED"

	// CLOSED
	OrderStatusFilled   OrderStatus = "FILLED"
	OrderStatusCanceled OrderStatus = "CANCELED"
	OrderStatusExpired  OrderStatus = "EXPIRED"
)

type OrderType string

const (
	OrderTypeMarket OrderType = "MARKET"
	OrderTypeLimit  OrderType = "LIMIT"
)

type OrderSide string

const (
	OrderSideBuy  OrderSide = "BUY"
	OrderSideSell OrderSide = "SELL"
)

type TimeInForce string

const (
	TimeInForceGTC TimeInForce = "GTC"
	TimeInForceIOC TimeInForce = "IOC"
	TimeInForceFOK TimeInForce = "FOK"
)

type PositionSide string

const (
	PositionSideLong  PositionSide = "LONG"
	PositionSideShort PositionSide = "SHORT"
	PositionSideFlat  PositionSide = "FLAT"
)

type BinanceAccountType string

type OkxAccountType string

const (
	OkxAccountTypeLive OkxAccountType = "0"
	OkxAccountTypeAws  OkxAccountType = "1"
	OkxAccountTypeDemo OkxAccountType = "2"
)

type BybitAccountType string

const (
	BybitAccountTypeSpot           BybitAccountType = "0"
	BybitAccountTypeLinear         BybitAccountType = "1"
	BybitAccountTypeInverse        BybitAccountType = "2"
	BybitAccountTypeOption         BybitAccountType = "3"
	BybitAccountTypeSpotTestnet    BybitAccountType = "4"
	BybitAccountTypeLinearTestnet  BybitAccountType = "5"
	BybitAccountTypeInverseTestnet BybitAccountType = "6"
	BybitAccountTypeOptionTestnet  BybitAccountType = "7"
)

type AccountType string
