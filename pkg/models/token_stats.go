package models

// TokenStats représente les statistiques d'un token
type TokenStats struct {
	HolderCount    int     `json:"holder_count"`
	Volume1h       float64 `json:"volume_1h"`
	Volume24h      float64 `json:"volume_24h"`
	Price          float64 `json:"price"`
	MarketCap      float64 `json:"market_cap"`
	PriceChange1h  float64 `json:"price_change_1h"`
	BuyCount1h     int     `json:"buy_count_1h"`
	SellCount1h    int     `json:"sell_count_1h"`
	LiquidityUSD   float64 `json:"liquidity_usd"`
	PoolAddress    string  `json:"pool_address,omitempty"`
	PoolTradesLast24h int  `json:"pool_trades_last_24h"`
} 