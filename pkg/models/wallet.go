package models

import (
	"time"
)

// WalletProfile représente le profil complet d'un wallet
type WalletProfile struct {
	Address           string           `json:"address"`
	Name              string           `json:"name,omitempty"`
	TwitterUsername   string           `json:"twitter_username,omitempty"`
	TwitterName       string           `json:"twitter_name,omitempty"`
	Avatar            string           `json:"avatar,omitempty"`
	CreatedAt         time.Time        `json:"created_at"`
	LastActive        time.Time        `json:"last_active"`
	TrustScore        float64          `json:"trust_score"`
	Tags              []string         `json:"tags"`
	TotalTransactions int              `json:"total_transactions"`
	WinRate           float64          `json:"win_rate"`
	Holdings          []WalletHolding  `json:"holdings"`
	RiskFactors       WalletRiskFactors `json:"risk_factors"`
}

// WalletHolding représente un token détenu par un wallet
type WalletHolding struct {
	TokenAddress     string    `json:"token_address"`
	TokenSymbol      string    `json:"token_symbol"`
	Balance          string    `json:"balance"`
	Value            string    `json:"value"`
	UnrealizedProfit float64   `json:"unrealized_profit"`
	BuyCount         int       `json:"buy_count"`
	SellCount        int       `json:"sell_count"`
	LastActive       time.Time `json:"last_active"`
}

// WalletRiskFactors représente les facteurs de risque d'un wallet
type WalletRiskFactors struct {
	WalletAddress       string    `json:"wallet_address"`
	RiskScore           float64   `json:"risk_score"`
	FalseFlaggedTokens  int       `json:"false_flagged_tokens"`
	RugpullExitRate     float64   `json:"rugpull_exit_rate"`
	FastSellRate        float64   `json:"fast_sell_rate"`
	LongHoldRate        float64   `json:"long_hold_rate"`
	UpdatedAt           time.Time `json:"updated_at"`
}

// WalletInteraction représente une interaction entre un wallet et un token
type WalletInteraction struct {
	ID                 string    `json:"id"`
	WalletAddress      string    `json:"wallet_address"`
	TokenAddress       string    `json:"token_address"`
	TokenSymbol        string    `json:"token_symbol,omitempty"`
	TxHash             string    `json:"tx_hash"`
	BlockNumber        int64     `json:"block_number"`
	Timestamp          time.Time `json:"timestamp"`
	ActionType         string    `json:"action_type"` // "buy", "sell", "transfer", etc.
	Amount             float64   `json:"amount"`
	Value              float64   `json:"value"`
	Price              float64   `json:"price"`
	Success            bool      `json:"success"`
	RelatedBuyTimestamp time.Time `json:"related_buy_timestamp,omitempty"`
	TokenRiskFactor    float64   `json:"token_risk_factor,omitempty"`
}

// WalletToken représente la relation entre un wallet et un token
type WalletToken struct {
	WalletAddress        string    `json:"wallet_address"`
	TokenAddress         string    `json:"token_address"`
	TokenSymbol          string    `json:"token_symbol,omitempty"`
	TokenName            string    `json:"token_name,omitempty"`
	FirstInteractionTime time.Time `json:"first_interaction_time"`
	LastInteractionTime  time.Time `json:"last_interaction_time"`
	TransactionCount     int       `json:"transaction_count"`
	TotalVolume          float64   `json:"total_volume"`
	CurrentBalance       float64   `json:"current_balance,omitempty"`
	NetProfit            float64   `json:"net_profit,omitempty"`
}

// HolderQualityReport représente un rapport sur la qualité des détenteurs d'un token
type HolderQualityReport struct {
	TokenAddress         string             `json:"token_address"`
	TotalHolders         int                `json:"total_holders"`
	QualityScore         float64            `json:"quality_score"`
	SmartMoneyRatio      float64            `json:"smart_money_ratio"`
	SmartMoneyCount      int                `json:"smart_money_count"`
	EarlyTrustedRatio    float64            `json:"early_trusted_ratio"`
	SniperRatio          float64            `json:"sniper_ratio"`
	SniperCount          int                `json:"sniper_count"`
	CategoryDistribution map[string]float64 `json:"category_distribution"`
	Timestamp            time.Time          `json:"timestamp"`
} 