package models

import (
	"time"
)

// Constantes pour les états du cycle de vie des tokens
const (
	LifecycleStateCompleted       = "COMPLETED"
	LifecycleStateDiscovered      = "DISCOVERED"
	LifecycleStateValidated       = "VALIDATED"
	LifecycleStateHyped           = "HYPED"
	LifecycleStateSleepMode       = "SLEEP_MODE"
	LifecycleStateMonitoringLight = "MONITORING_LIGHT"
	LifecycleStateReactivated     = "REACTIVATED"
)

// Token représente un token avec ses métadonnées
type Token struct {
	Address            string    `json:"address"`
	Symbol             string    `json:"symbol"`
	Name               string    `json:"name"`
	TotalSupply        int64     `json:"total_supply"`
	HolderCount        int       `json:"holder_count"`
	CreatedTimestamp   int64     `json:"created_timestamp,omitempty"`
	CompletedTimestamp int64     `json:"completed_timestamp,omitempty"`
	LastTradeTimestamp int64     `json:"last_trade_timestamp,omitempty"`
	Logo               string    `json:"logo,omitempty"`
	Twitter            string    `json:"twitter,omitempty"`
	Website            string    `json:"website,omitempty"`
	Telegram           string    `json:"telegram,omitempty"`
	CachedAt           time.Time `json:"cached_at,omitempty"`
}

// TokenPrice représente les données de prix d'un token
type TokenPrice struct {
	TokenAddress string    `json:"token_address"`
	Price        float64   `json:"price"`
	Change1h     float64   `json:"change_1h"`
	Change24h    float64   `json:"change_24h"`
	Change7d     float64   `json:"change_7d"`
	Volume24h    float64   `json:"volume_24h"`
	MarketCap    float64   `json:"market_cap"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// TokenMetrics représente les métriques d'analyse d'un token
type TokenMetrics struct {
	TokenAddress       string    `json:"token_address"`
	HolderCount        int       `json:"holder_count"`
	IntelligentHolders int       `json:"intelligent_holders"`
	AverageHoldTime    float64   `json:"average_hold_time"`
	CreatorWalletAddr  string    `json:"creator_wallet_addr"`
	CreatorTrustScore  float64   `json:"creator_trust_score"`
	DevTrustScore      float64   `json:"dev_trust_score"`
	SmartMoneyHolders  int       `json:"smart_money_holders"`
	AverageTrustScore  float64   `json:"average_trust_score"`
	RiskFactor         float64   `json:"risk_factor"`
	Volume1h           float64   `json:"volume_1h"`
	Volume24h          float64   `json:"volume_24h"`
	Price              float64   `json:"price"`
	MarketCap          float64   `json:"market_cap"`
	PriceChange1h      float64   `json:"price_change_1h"`
	BuyCount1h         int       `json:"buy_count_1h"`
	SellCount1h        int       `json:"sell_count_1h"`
	UpdatedAt          time.Time `json:"updated_at"`
}

// TokenTrade représente une transaction sur un token
type TokenTrade struct {
	ID            string    `json:"id"`
	TxHash        string    `json:"tx_hash"`
	BlockNumber   uint64    `json:"block_number"`
	Timestamp     time.Time `json:"timestamp"`
	TokenAddress  string    `json:"token_address"`
	TokenSymbol   string    `json:"token_symbol"`
	WalletAddress string    `json:"wallet_address"`
	ActionType    string    `json:"action_type"` // buy, sell
	TradeType     string    `json:"trade_type"` // buy, sell
	Amount        float64   `json:"amount"`
	Price         float64   `json:"price"`
	Value         float64   `json:"value"`
	TotalValue    float64   `json:"total_value"`
	Success       bool      `json:"success"`
}

// TokenAlert représente une alerte sur un token
type TokenAlert struct {
	ID              string    `json:"id"`
	TokenAddress    string    `json:"token_address"`
	TokenSymbol     string    `json:"token_symbol"`
	AlertType       string    `json:"alert_type"` // "rugpull", "pump", "dump", etc.
	Severity        string    `json:"severity"` // "high", "medium", "low"
	Message         string    `json:"message"`
	DetectedAt      time.Time `json:"detected_at"`
	ConfirmationCount int     `json:"confirmation_count"`
	IsConfirmed      bool     `json:"is_confirmed"`
	RelatedWallets  []string  `json:"related_wallets,omitempty"`
}

// TokenHistoricalMetrics représente des métriques historiques pour un token
type TokenHistoricalMetrics struct {
	TokenAddress     string    `json:"token_address"`
	Date             time.Time `json:"date"`
	Price            float64   `json:"price"`
	Volume           float64   `json:"volume"`
	MarketCap        float64   `json:"market_cap"`
	HolderCount      int       `json:"holder_count"`
	IntelligentRatio float64   `json:"intelligent_ratio"` // Proportion de holders "intelligents"
	TrustScore       float64   `json:"trust_score"`       // Score global de confiance pour le token
	SocialScore      float64   `json:"social_score"`      // Score d'activité sociale
}

// TokenPricePoint représente un point de prix pour un token
type TokenPricePoint struct {
	TokenAddress string    `json:"token_address"`
	Timestamp    time.Time `json:"timestamp"`
	Open         float64   `json:"open"`
	High         float64   `json:"high"`
	Low          float64   `json:"low"`
	Close        float64   `json:"close"`
	Volume       float64   `json:"volume"`
}

// TokenTrader représente un trader actif sur un token
type TokenTrader struct {
	WalletAddress     string  `json:"wallet_address"`
	TokenAddress      string  `json:"token_address"`
	RelativeVolume    float64 `json:"relative_volume"`
	EarlyInvestor     float64 `json:"early_investor"`
	TransactionCount  int     `json:"transaction_count"`
}

// TokenCreatorInfo représente les informations sur le créateur d'un token
type TokenCreatorInfo struct {
	TokenAddress       string    `json:"token_address"`
	CreatorAddress     string    `json:"creator_address"`
	CreatorTrustScore  float64   `json:"creator_trust_score"`
	CreationTimestamp  time.Time `json:"creation_timestamp"`
	OtherTokensCreated int       `json:"other_tokens_created"`
	PrevRugpulls       int       `json:"prev_rugpulls"`
	CreatorTags        []string  `json:"creator_tags"`
}

// TradeHistory représente une transaction historique
type TradeHistory struct {
	WalletAddress string    `json:"wallet_address"`
	Amount        float64   `json:"amount"`
	Type          string    `json:"type"` // buy ou sell
	Timestamp     time.Time `json:"timestamp"`
	TxHash        string    `json:"tx_hash"`
}

// WalletAnalysis contient l'analyse des wallets pour un token
type WalletAnalysis struct {
	TokenAddress string `json:"token_address"`
	Timestamp    time.Time `json:"timestamp"`
	TotalWallets int    `json:"total_wallets"`
	WalletCategories struct {
		Smart    int `json:"smart"`
		Trusted  int `json:"trusted"`
		Fresh    int `json:"fresh"`
		Bot      int `json:"bot"`
		Sniper   int `json:"sniper"`
		Bluechip int `json:"bluechip"`
		Bundler  int `json:"bundler"`
	} `json:"wallet_categories"`
	TrustMetrics struct {
		AvgTrustScore     float64 `json:"avg_trust_score"`
		SmartMoneyRatio   float64 `json:"smart_money_ratio"`
		SmartMoneyCount   int     `json:"smart_money_count"`
		TotalWallets      int     `json:"total_wallets"`
		EarlyTrustedRatio float64 `json:"early_trusted_ratio"`
		SmartMoneyActivity float64 `json:"smart_money_activity"`
	} `json:"trust_metrics"`
	TradePatterns struct {
		BuyOrders    int     `json:"buy_orders"`
		SellOrders   int     `json:"sell_orders"`
		BuySellRatio float64 `json:"buy_sell_ratio"`
		AvgHoldTime  float64 `json:"avg_hold_time"`
	} `json:"trade_patterns"`
	SniperCount int `json:"sniper_count"`
	SniperRatio float64 `json:"sniper_ratio"`
	WalletDetails []WalletDetail `json:"wallet_details"`
}

// WalletDetail contient les détails d'un wallet impliqué dans un token
type WalletDetail struct {
	Address    string   `json:"address"`
	TrustScore float64  `json:"trust_score"`
	Categories []string `json:"categories"`
	EntryRank  int      `json:"entry_rank"`
	EntryTime  time.Time `json:"entry_time"`
	Volume     float64  `json:"volume"`
	Buys       int      `json:"buys"`
	Sells      int      `json:"sells"`
}

// XScoreResult contient le résultat du calcul du X-Score
type XScoreResult struct {
	TokenAddress string             `json:"token_address"`
	XScore       float64            `json:"x_score"`
	BaseScore    float64            `json:"base_score"`
	Components   map[string]float64 `json:"components"`
	AntiDump     *AntiDumpResult    `json:"anti_dump"`
	CalculatedAt time.Time          `json:"calculated_at"`
}

// AntiDumpResult contient le résultat de l'analyse anti-dump
type AntiDumpResult struct {
	Detected bool          `json:"detected"`
	Severity float64       `json:"severity"`
	Clusters []DumpCluster `json:"clusters"`
}

// DumpCluster représente un groupe de ventes coordonnées
type DumpCluster struct {
	TimestampStart   time.Time `json:"timestamp_start"`
	TimestampEnd     time.Time `json:"timestamp_end"`
	DurationSeconds  float64   `json:"duration_seconds"`
	TransactionCount int       `json:"transaction_count"`
	UniqueWallets    int       `json:"unique_wallets"`
	SmartWallets     int       `json:"smart_wallets"`
	TotalVolume      float64   `json:"total_volume"`
	Severity         float64   `json:"severity"`
}

// SmartWalletReturns contient les informations sur le retour de wallets smart
type SmartWalletReturns struct {
	Detected          bool                `json:"detected"`
	Wallets           []string            `json:"wallets"`
	ReturnTimestamp   time.Time           `json:"return_timestamp"`
	InitialExitTimestamp time.Time        `json:"initial_exit_timestamp"`
	ReturningTotalVolume float64          `json:"returning_total_volume"`
	Severity          float64             `json:"severity"`
}

// ReactivationCandidate représente un token candidat à la réactivation
type ReactivationCandidate struct {
	TokenAddress      string                `json:"token_address"`
	TokenSymbol       string                `json:"token_symbol"`
	ReactivationScore float64               `json:"reactivation_score"`
	Changes           map[string]float64    `json:"changes"`
	SmartReturns      *SmartWalletReturns   `json:"smart_returns"`
	CurrentMetrics    *TokenMetrics         `json:"current_metrics"`
	DetectedAt        time.Time             `json:"detected_at"`
}