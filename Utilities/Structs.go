package Utilities

import "time"

type SignalType int
type OrderState int
type ExecuteOrderCount int

const (
	BUY  SignalType = 0
	SELL SignalType = 1
)

const (
	RECEIVED  OrderState = 0
	ORDERED   OrderState = 1
	CANCELLED OrderState = 2
)

const (
	EXECUTIONCOUNT ExecuteOrderCount = 2
)

type Order struct {
	Id            string
	Signal        string
	State         OrderState
	Symbol        string
	OrderType     string
	PositionSize  int
	Strategy      string
	EachSLPercent string
	DaySLPercent  string
	Broker        string
	Secret        string
	RecvTime      time.Time
	OrderedTime   time.Time
	CancelledTime time.Time
	Product       string
	Exchange      string
}

type ConfigJson struct {
	LogPath     string `json:"LogPath"`
	Port        string `json:"Port"`
	ApiKey      string `json:"ApiKey"`
	AccessToken string `json:"AccessToken"`
	CancelMins  string `json:"CancelMins"`
}

type StopLoss struct {
	BuyPrice        float64 `json:"buy-price"`
	StopLossPercent string  `json:"stoploss-percent"`
	StopLossPrice   float64 `json:"stoploss-price"`
}

type SellStopLoss struct {
	SellPrice       float64 `json:"sell-price"`
	StopLossPercent string  `json:"stoploss-percent"`
	StopLossPrice   float64 `json:"stoploss-price"`
}

type CustomGTTParams struct {
	Tradingsymbol   string  `json:"tradingsymbol"`
	Exchange        string  `json:"exchange"`
	TransactionType string  `json:"transaction_type"`
	Quantity        int     `json:"quantity"`
	OrderType       string  `json:"order_type"`
	Product         string  `json:"product"`
	Price           float64 `json:"price"`
}

type GTTOrerResponse struct {
	Status string `json:"status"`
	Data   struct {
		TriggerID int `json:"trigger_id"`
	} `json:"data"`
}

// GTTOrder represents the structure of the JSON response
type GTTOrder struct {
	Status string       `json:"status"`
	Data   []GTTOrderData `json:"data"`
}

// GTTOrderData represents the nested data within the JSON response
type GTTOrderData struct {
	ID            int               `json:"id"`
	UserID        string            `json:"user_id"`
	ParentTrigger interface{}       `json:"parent_trigger"`
	Type          string            `json:"type"`
	CreatedAt     string            `json:"created_at"`
	UpdatedAt     string            `json:"updated_at"`
	ExpiresAt     string            `json:"expires_at"`
	OrderStatus   string            `json:"status"`
	Condition     GTTOrderCondition `json:"condition"`
	Orders        []GTTOrderOrder   `json:"orders"`
}

// GTTOrderCondition represents the condition field within the JSON response
type GTTOrderCondition struct {
	Exchange        string `json:"exchange"`
	LastPrice       int    `json:"last_price"`
	TradingSymbol   string `json:"tradingsymbol"`
	TriggerValues   []int  `json:"trigger_values"`
	InstrumentToken int    `json:"instrument_token"`
}

// GTTOrderOrder represents the orders field within the JSON response
type GTTOrderOrder struct {
	Exchange        string `json:"exchange"`
	TradingSymbol   string `json:"tradingsymbol"`
	Product         string `json:"product"`
	OrderType       string `json:"order_type"`
	TransactionType string `json:"transaction_type"`
	Quantity        int    `json:"quantity"`
	Price           int    `json:"price"`
	// Add more fields as needed based on the actual response structure
}
