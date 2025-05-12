package models

import (
	"time"
)

// WalletTrustScore représente le score de confiance d'un wallet
type WalletTrustScore struct {
	Address     string    `json:"address"`
	TrustScore  float64   `json:"trust_score"`
	LastUpdated time.Time `json:"last_updated"`
}

// TokenTrustMetrics représente les métriques de confiance pour un token
type TokenTrustMetrics struct {
	TokenAddress           string         `json:"token_address"`
	ActiveWallets          int            `json:"active_wallets"`
	TrustedWallets         int            `json:"trusted_wallets"`
	AvgTrustScore          float64        `json:"avg_trust_score"`
	TrustScoreDistribution map[string]int `json:"trust_score_distribution"`
	EarlyTrustRatio        float64        `json:"early_trust_ratio"`
	SmartMoneyCount        int            `json:"smart_money_count"`
	SmartMoneyRatio        float64        `json:"smart_money_ratio"`
	SmartMoneyActivity     float64        `json:"smart_money_activity"`
	EarlyTrustedRatio      float64        `json:"early_trusted_ratio"`
}

// WalletSimilarity représente la similarité entre deux wallets
type WalletSimilarity struct {
	WalletAddress  string  `json:"wallet_address"`
	Score          float64 `json:"score"`
	CommonTokens   int     `json:"common_tokens"`
	TimingScore    float64 `json:"timing_score"`
	PositionScore  float64 `json:"position_score"`
	TrustScore     float64 `json:"trust_score,omitempty"`
	TradeFrequency float64 `json:"trade_frequency"`
}

// WalletInfluence représente l'influence d'un wallet sur un token
type WalletInfluence struct {
	WalletAddress    string  `json:"wallet_address"`
	TokenAddress     string  `json:"token_address"`
	InfluenceScore   float64 `json:"influence_score"`
	VolumeImpact     float64 `json:"volume_impact"`
	TimingImpact     float64 `json:"timing_impact"`
	PriceImpact      float64 `json:"price_impact"`
	TransactionCount int     `json:"transaction_count"`
}

// TrustScorePoint représente un point sur un graphique d'évolution du score de confiance
type TrustScorePoint struct {
	Timestamp  time.Time `json:"timestamp"`
	TrustScore float64   `json:"trust_score"`
}

// TrustSystemMetrics représente les métriques de système de confiance
type TrustSystemMetrics struct {
	TotalWallets        int       `json:"total_wallets"`
	TotalTokens         int       `json:"total_tokens"`
	TotalInteractions   int       `json:"total_interactions"`
	AvgTrustScore       float64   `json:"avg_trust_score"`
	TrustedWalletsRatio float64   `json:"trusted_wallets_ratio"`
	LastUpdated         time.Time `json:"last_updated"`
}

// TrustScoreUpdate représente une mise à jour du score de confiance
type TrustScoreUpdate struct {
	WalletAddress  string    `json:"wallet_address"`
	OldScore       float64   `json:"old_score"`
	NewScore       float64   `json:"new_score"`
	ChangeReason   string    `json:"change_reason,omitempty"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// TrustNetworkMetrics représente les métriques globales du réseau de confiance
type TrustNetworkMetrics struct {
	TotalWallets         int                    `json:"total_wallets"`
	TotalTokens          int                    `json:"total_tokens"`
	AvgTrustScore        float64                `json:"avg_trust_score"`
	SmartWalletsCount    int                    `json:"smart_wallets_count"`
	TrustedWalletsCount  int                    `json:"trusted_wallets_count"`
	LowTrustCount        int                    `json:"low_trust_count"`
	ScoreDistribution    map[string]int         `json:"score_distribution"`
	TopActiveTokens      []map[string]interface{} `json:"top_active_tokens"`
	TopInfluencerWallets []WalletTrustScore     `json:"top_influencer_wallets"`
	LastUpdated          time.Time              `json:"last_updated"`
}

// ConnectionStrength représente la force de la connexion entre deux wallets
type ConnectionStrength struct {
	WalletA       string  `json:"wallet_a"`
	WalletB       string  `json:"wallet_b"`
	Strength      float64 `json:"strength"`      // Entre 0 et 1
	SharedTokens  int     `json:"shared_tokens"`
	FirstSeen     time.Time `json:"first_seen"`
	LastSeen      time.Time `json:"last_seen"`
	Interactions  int     `json:"interactions"`
}

// WalletActivity représente une activité d'un wallet
type WalletActivity struct {
	WalletAddress string    `json:"wallet_address"`
	TokenAddress  string    `json:"token_address"`
	ActionType    string    `json:"action_type"`
	Timestamp     time.Time `json:"timestamp"`
	Success       bool      `json:"success"`
}

// WalletCluster représente un groupe de wallets similaires
type WalletCluster struct {
	ID            string   `json:"id"`
	WalletAddresses []string `json:"wallet_addresses"`
	AvgTrustScore float64  `json:"avg_trust_score"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
	Tags          []string `json:"tags"`
	Description   string   `json:"description"`
} 