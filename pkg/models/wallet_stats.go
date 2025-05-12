package models

import (
	"time"
)

// WalletStats represents statistics for a wallet
type WalletStats struct {
	WalletAddress   string    `json:"wallet_address"`
	TotalProfit     float64   `json:"total_profit"`
	TotalVolume     float64   `json:"total_volume"`
	WinRate         float64   `json:"win_rate"`
	TokenCount      int       `json:"token_count"`
	TransactionCount int      `json:"transaction_count"`
	FirstTransaction time.Time `json:"first_transaction"`
	LastTransaction  time.Time `json:"last_transaction"`
	AverageHoldTime  float64   `json:"average_hold_time"`
	BiggestWin       float64   `json:"biggest_win"`
	BiggestLoss      float64   `json:"biggest_loss"`
	UpdatedAt        time.Time `json:"updated_at"`
} 