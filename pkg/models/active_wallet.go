package models

import (
	"time"
)

// ActiveWallet représente un wallet actif sur un token
type ActiveWallet struct {
	Address                 string    `json:"address"`
	FirstTransactionTimestamp time.Time `json:"first_transaction_timestamp"`
	EntryRank               int       `json:"entry_rank"`
	TransactionCount        int       `json:"transaction_count"`
	LastActive              time.Time `json:"last_active"`
	LastActivity            time.Time `json:"last_activity"`
	TrustScore              float64   `json:"trust_score,omitempty"`
	NetPosition             float64   `json:"net_position"` // différence entre achats et ventes
	BuyVolume               float64   `json:"buy_volume"`
	SellVolume              float64   `json:"sell_volume"`
	Categories              []string  `json:"categories,omitempty"` // smart, trusted, fresh, etc.
} 