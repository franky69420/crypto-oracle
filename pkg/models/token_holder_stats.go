package models

import (
	"time"
)

// TokenHolderStats represents holder statistics for a token
type TokenHolderStats struct {
	TokenAddress    string                 `json:"token_address"`
	TokenSymbol     string                 `json:"token_symbol"`
	TotalHolders    int                    `json:"total_holders"`
	HoldersByAmount map[string]int         `json:"holders_by_amount"`
	HoldersByTime   map[string]int         `json:"holders_by_time"`
	HoldersByValue  map[string]int         `json:"holders_by_value"`
	Distribution    map[string]float64     `json:"distribution"`
	BuyerCount      int                    `json:"buyer_count"`
	SellerCount     int                    `json:"seller_count"`
	ActiveWallets   int                    `json:"active_wallets"`
	UpdatedAt       time.Time              `json:"updated_at"`
} 