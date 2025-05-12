package models

import (
	"time"
)

// PumpRankings represents tokens currently pumping
type PumpRankings struct {
	Timeframe   string       `json:"timeframe"`
	UpdatedAt   time.Time    `json:"updated_at"`
	TotalTokens int          `json:"total_tokens"`
	Rankings    []PumpToken  `json:"rankings"`
}

// PumpToken represents a token in the pump rankings
type PumpToken struct {
	Rank          int       `json:"rank"`
	TokenAddress  string    `json:"token_address"`
	TokenSymbol   string    `json:"token_symbol"`
	TokenName     string    `json:"token_name"`
	Price         float64   `json:"price"`
	PriceChange   float64   `json:"price_change"`
	Volume        float64   `json:"volume"`
	VolumeChange  float64   `json:"volume_change"`
	MarketCap     float64   `json:"market_cap"`
	HolderCount   int       `json:"holder_count"`
	CreateTime    time.Time `json:"create_time"`
	UpdatedAt     time.Time `json:"updated_at"`
} 