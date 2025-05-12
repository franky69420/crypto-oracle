package models

import (
	"time"
)

// WalletRanking represents a wallet in a ranking list
type WalletRanking struct {
	Rank           int       `json:"rank"`
	WalletAddress  string    `json:"wallet_address"`
	WalletName     string    `json:"wallet_name,omitempty"`
	TotalProfit    float64   `json:"total_profit"`
	TotalVolume    float64   `json:"total_volume"`
	WinRate        float64   `json:"win_rate"`
	TokenCount     int       `json:"token_count"`
	TrustScore     float64   `json:"trust_score"`
	Tags           []string  `json:"tags,omitempty"`
	TimePeriod     string    `json:"time_period"`
	LastActive     time.Time `json:"last_active"`
} 