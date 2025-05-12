package models

import (
	"time"
)

// MarketCapCandle represents market cap candle data for a token
type MarketCapCandle struct {
	TokenAddress  string    `json:"token_address"`
	Timestamp     time.Time `json:"timestamp"`
	Open          float64   `json:"open"`
	High          float64   `json:"high"`
	Low           float64   `json:"low"`
	Close         float64   `json:"close"`
	Volume        float64   `json:"volume"`
	MarketCapOpen float64   `json:"market_cap_open"`
	MarketCapHigh float64   `json:"market_cap_high"`
	MarketCapLow  float64   `json:"market_cap_low"`
	MarketCapClose float64  `json:"market_cap_close"`
} 