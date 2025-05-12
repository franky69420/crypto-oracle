package models

import (
	"time"
)

// DailyProfit represents daily profit data for a wallet
type DailyProfit struct {
	Date          time.Time `json:"date"`
	Profit        float64   `json:"profit"`
	Volume        float64   `json:"volume"`
	TransactionCount int    `json:"transaction_count"`
	WalletAddress string    `json:"wallet_address"`
} 