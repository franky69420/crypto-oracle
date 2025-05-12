package gmgn

import "time"

// TokenStatResponse contient les statistiques d'un token
type TokenStatResponse struct {
	Address      string  `json:"address"`
	Symbol       string  `json:"symbol"`
	Name         string  `json:"name"`
	Logo         string  `json:"logo"`
	Price        float64 `json:"price"`
	PriceChange  float64 `json:"price_change"`
	Volume       float64 `json:"volume"`
	VolumeChange float64 `json:"volume_change"`
	Mcap         float64 `json:"mcap"`
	McapChange   float64 `json:"mcap_change"`
	Holders      int     `json:"holders"`
	HoldersChange int    `json:"holders_change"`
	Tags         []Tag   `json:"tags"`
}

// Tag représente un tag pour un token ou un wallet
type Tag struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// TradeHistoryResponse contient l'historique des transactions d'un token
type TradeHistoryResponse struct {
	List []Trade `json:"list"`
	Next string  `json:"next"`
}

// Trade représente une transaction
type Trade struct {
	ID          string    `json:"id"`
	Timestamp   int64     `json:"timestamp"`
	Time        time.Time `json:"time"`
	BlockHeight int64     `json:"block_height"`
	TxHash      string    `json:"tx_hash"`
	Type        string    `json:"type"`
	TokenAmount float64   `json:"token_amount"`
	UsdAmount   float64   `json:"usd_amount"`
	Price       float64   `json:"price"`
	PoolAddress string    `json:"pool_address"`
	PoolName    string    `json:"pool_name"`
	WalletFrom  string    `json:"wallet_from"`
	WalletTo    string    `json:"wallet_to"`
	Tags        []Tag     `json:"tags"`
}

// KlineDataResponse contient les données de prix d'un token
type KlineDataResponse struct {
	S []int64   `json:"s"` // Status
	T []int64   `json:"t"` // Timestamp
	O []float64 `json:"o"` // Open
	H []float64 `json:"h"` // High
	L []float64 `json:"l"` // Low
	C []float64 `json:"c"` // Close
	V []float64 `json:"v"` // Volume
}

// Trader représente un trader d'un token
type Trader struct {
	Address      string    `json:"address"`
	Nickname     string    `json:"nickname"`
	BuyVolume    float64   `json:"buy_volume"`
	SellVolume   float64   `json:"sell_volume"`
	NetVolume    float64   `json:"net_volume"`
	UsdNetVolume float64   `json:"usd_net_volume"`
	TradeCount   int       `json:"trade_count"`
	LastTrade    time.Time `json:"last_trade"`
	TokenBalance float64   `json:"token_balance"`
	UsdBalance   float64   `json:"usd_balance"`
	Tags         []Tag     `json:"tags"`
}

// TokenHolderStatResponse contient les statistiques des détenteurs d'un token
type TokenHolderStatResponse struct {
	Total       int `json:"total"`
	Distribution struct {
		Dolphin  int `json:"dolphin"`
		Shark    int `json:"shark"`
		Whale    int `json:"whale"`
		Wallaby  int `json:"wallaby"`
		Fish     int `json:"fish"`
		Shrimp   int `json:"shrimp"`
		Crab     int `json:"crab"`
		Octopus  int `json:"octopus"`
		Lobster  int `json:"lobster"`
		Stingray int `json:"stingray"`
	} `json:"distribution"`
}

// TokenWalletTagsStatResponse contient les statistiques des tags de wallets pour un token
type TokenWalletTagsStatResponse struct {
	Total         int                    `json:"total"`
	Distributions []TagDistributionEntry `json:"distributions"`
}

// TagDistributionEntry représente une entrée dans la distribution des tags
type TagDistributionEntry struct {
	Tag   string `json:"tag"`
	Count int    `json:"count"`
}

// WalletInfoResponse contient les informations d'un wallet
type WalletInfoResponse struct {
	Address          string  `json:"address"`
	Nickname         string  `json:"nickname"`
	HoldingsValue    float64 `json:"holdings_value"`
	PortfolioHistory []struct {
		Timestamp int64   `json:"timestamp"`
		Value     float64 `json:"value"`
	} `json:"portfolio_history"`
	Tags []Tag `json:"tags"`
}

// Holding représente un token détenu par un wallet
type Holding struct {
	TokenAddress string  `json:"token_address"`
	TokenSymbol  string  `json:"token_symbol"`
	TokenName    string  `json:"token_name"`
	TokenLogo    string  `json:"token_logo"`
	Amount       float64 `json:"amount"`
	UsdValue     float64 `json:"usd_value"`
	Price        float64 `json:"price"`
	PriceChange  float64 `json:"price_change"`
}

// WalletStatResponse contient les statistiques d'un wallet
type WalletStatResponse struct {
	TotalTrades      int     `json:"total_trades"`
	WinningTrades    int     `json:"winning_trades"`
	LosingTrades     int     `json:"losing_trades"`
	AverageHoldTime  float64 `json:"average_hold_time"`
	AverageGain      float64 `json:"average_gain"`
	AverageLoss      float64 `json:"average_loss"`
	TotalVolume      float64 `json:"total_volume"`
	BuyVolume        float64 `json:"buy_volume"`
	SellVolume       float64 `json:"sell_volume"`
	TotalProfit      float64 `json:"total_profit"`
	AverageTradeSize float64 `json:"average_trade_size"`
	WinRate          float64 `json:"win_rate"`
}

// TrendingResponse contient les tokens en tendance
type TrendingResponse struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data struct {
		Rank []TrendingToken `json:"rank"`
	} `json:"data"`
}

// TrendingToken représente un token en tendance
type TrendingToken struct {
	Address      string  `json:"address"`
	Symbol       string  `json:"symbol"`
	Name         string  `json:"name"`
	Logo         string  `json:"logo"`
	Price        float64 `json:"price"`
	PriceChange  float64 `json:"price_change"`
	Volume       float64 `json:"volume"`
	VolumeChange float64 `json:"volume_change"`
	Mcap         float64 `json:"mcap"`
	McapChange   float64 `json:"mcap_change"`
}

// CompletedTokensResponse contient les tokens complétés
type CompletedTokensResponse struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data struct {
		Rank []CompletedToken `json:"rank"`
	} `json:"data"`
}

// CompletedToken représente un token complété
type CompletedToken struct {
	Address     string  `json:"address"`
	Symbol      string  `json:"symbol"`
	Name        string  `json:"name"`
	Logo        string  `json:"logo"`
	MaxPrice    float64 `json:"max_price"`
	MaxMcap     float64 `json:"max_mcap"`
	MaxVolume   float64 `json:"max_volume"`
	CreatedTime int64   `json:"created_time"`
	MaxPriceTs  int64   `json:"max_price_ts"`
} 