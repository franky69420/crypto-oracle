package models

import (
	"time"
)

// SimilarCoinAnalysis represents similar coin analysis
type SimilarCoinAnalysis struct {
	TokenAddress     string         `json:"token_address"`
	TokenSymbol      string         `json:"token_symbol"`
	TokenName        string         `json:"token_name"`
	Timeframe        string         `json:"timeframe"`
	MaximumGains     []SimilarCoin  `json:"maximum_gains"`
	EarliestCoins    []SimilarCoin  `json:"earliest_coins"`
	UpdatedAt        time.Time      `json:"updated_at"`
}

// SimilarCoin represents a token similar to the analyzed token
type SimilarCoin struct {
	TokenAddress     string    `json:"token_address"`
	TokenSymbol      string    `json:"token_symbol"`
	TokenName        string    `json:"token_name"`
	Similarity       float64   `json:"similarity"`
	InitialPrice     float64   `json:"initial_price"`
	PeakPrice        float64   `json:"peak_price"`
	PriceMultiplier  float64   `json:"price_multiplier"`
	TimeToMultiplier Duration  `json:"time_to_multiplier"`
	LaunchDate       time.Time `json:"launch_date"`
	PeakDate         time.Time `json:"peak_date"`
}

// Duration is a custom duration type to handle string duration representation
type Duration string 