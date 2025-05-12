package models

import (
	"time"
)

// TokenHolders represents information about token holders
type TokenHolders struct {
	TokenAddress   string        `json:"token_address"`
	TokenSymbol    string        `json:"token_symbol"`
	TotalHolders   int           `json:"total_holders"`
	TopBuyers      []TokenHolder `json:"top_buyers"`
	TopHolders     []TokenHolder `json:"top_holders"`
	UpdatedAt      time.Time     `json:"updated_at"`
}

// TokenHolder represents a holder of a token
type TokenHolder struct {
	WalletAddress  string    `json:"wallet_address"`
	WalletName     string    `json:"wallet_name,omitempty"`
	Balance        float64   `json:"balance"`
	Value          float64   `json:"value"`
	PercentOwned   float64   `json:"percent_owned"`
	BuyAmount      float64   `json:"buy_amount"`
	SellAmount     float64   `json:"sell_amount"`
	BuyCount       int       `json:"buy_count"`
	SellCount      int       `json:"sell_count"`
	FirstBuy       time.Time `json:"first_buy"`
	LastAction     time.Time `json:"last_action"`
	TrustScore     float64   `json:"trust_score,omitempty"`
	Tags           []string  `json:"tags,omitempty"`
} 