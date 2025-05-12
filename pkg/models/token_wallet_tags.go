package models

import (
	"time"
)

// TokenWalletTagsStats represents wallet tag statistics for a token
type TokenWalletTagsStats struct {
	TokenAddress   string                   `json:"token_address"`
	TokenSymbol    string                   `json:"token_symbol"`
	TotalHolders   int                      `json:"total_holders"`
	TagsCount      map[string]int           `json:"tags_count"`
	TagsPercentage map[string]float64       `json:"tags_percentage"`
	TagsDetails    map[string][]TaggedWallet `json:"tags_details"`
	UpdatedAt      time.Time                `json:"updated_at"`
}

// TaggedWallet represents a wallet with tags
type TaggedWallet struct {
	WalletAddress  string    `json:"wallet_address"`
	Tags           []string  `json:"tags"`
	Balance        float64   `json:"balance"`
	Value          float64   `json:"value"`
	FirstBuy       time.Time `json:"first_buy"`
	LastAction     time.Time `json:"last_action"`
} 