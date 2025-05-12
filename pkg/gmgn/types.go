package gmgn

// Response est la structure de base pour toutes les réponses de l'API GMGN
type Response struct {
	Code    int         `json:"code"`
	Msg     string      `json:"msg"`
	Reason  string      `json:"reason"`
	Data    interface{} `json:"data"`
}

// Wallet info structures
type WalletInfoResponse struct {
	Address     string       `json:"address"`
	Balance     string       `json:"balance"`
	Tokens      []TokenInfo  `json:"tokens"`
	Statistics  WalletStats  `json:"statistics"`
}

type TokenInfo struct {
	Symbol    string `json:"symbol"`
	Address   string `json:"address"`
	Balance   string `json:"balance"`
	Price     string `json:"price"`
	ValueUSD  string `json:"value_usd"`
}

type WalletStats struct {
	TotalValue       string `json:"total_value"`
	PnL24h           string `json:"pnl_24h"`
	PnLPercentage24h string `json:"pnl_percentage_24h"`
}

// Wallet statistics
type WalletStatResponse struct {
	Buy               int     `json:"buy"`
	Buy1d             int     `json:"buy_1d"`
	Buy7d             int     `json:"buy_7d"`
	Buy30d            int     `json:"buy_30d"`
	Sell              int     `json:"sell"`
	Sell1d            int     `json:"sell_1d"`
	Sell7d            int     `json:"sell_7d"`
	Sell30d           int     `json:"sell_30d"`
	Pnl               float64 `json:"pnl"`
	Pnl1d             float64 `json:"pnl_1d"`
	Pnl7d             float64 `json:"pnl_7d"`
	Pnl30d            float64 `json:"pnl_30d"`
	AllPnl            float64 `json:"all_pnl"`
	RealizedProfit    float64 `json:"realized_profit"`
	RealizedProfit1d  float64 `json:"realized_profit_1d"`
	RealizedProfit7d  float64 `json:"realized_profit_7d"`
	RealizedProfit30d float64 `json:"realized_profit_30d"`
	UnrealizedProfit  float64 `json:"unrealized_profit"`
	UnrealizedPnl     float64 `json:"unrealized_pnl"`
	TotalProfit       float64 `json:"total_profit"`
	TotalProfitPnl    float64 `json:"total_profit_pnl"`
	Balance           string  `json:"balance"`
	TotalValue        float64 `json:"total_value"`
	Winrate           float64 `json:"winrate"`
	TokenNum          int     `json:"token_num"`
	ProfitNum         int     `json:"profit_num"`
	Tags              []string `json:"tags"`
	LastActiveTimestamp int64  `json:"last_active_timestamp"`
	Risk              struct {
		TokenActive        string  `json:"token_active"`
		TokenHoneypot      string  `json:"token_honeypot"`
		TokenHoneypotRatio float64 `json:"token_honeypot_ratio"`
		NoBuyHold          string  `json:"no_buy_hold"`
		NoBuyHoldRatio     float64 `json:"no_buy_hold_ratio"`
		SellPassBuy        string  `json:"sell_pass_buy"`
		SellPassBuyRatio   float64 `json:"sell_pass_buy_ratio"`
		FastTx             string  `json:"fast_tx"`
		FastTxRatio        float64 `json:"fast_tx_ratio"`
	} `json:"risk"`
	AvgHoldingPeriod   float64 `json:"avg_holding_peroid"`
}

// Wallet Holdings
type Holding struct {
	Token struct {
		Address        string `json:"address"`
		TokenAddress   string `json:"token_address"`
		Symbol         string `json:"symbol"`
		Name           string `json:"name"`
		Decimals       int    `json:"decimals"`
		Logo           string `json:"logo"`
		PriceChange6h  string `json:"price_change_6h"`
		IsShowAlert    bool   `json:"is_show_alert"`
		IsHoneypot     *bool  `json:"is_honeypot"`
	} `json:"token"`
	Balance            string  `json:"balance"`
	UsdValue           string  `json:"usd_value"`
	RealizedProfit30d  string  `json:"realized_profit_30d"`
	RealizedProfit     string  `json:"realized_profit"`
	RealizedPnl        string  `json:"realized_pnl"`
	RealizedPnl30d     string  `json:"realized_pnl_30d"`
	UnrealizedProfit   string  `json:"unrealized_profit"`
	UnrealizedPnl      string  `json:"unrealized_pnl"`
	TotalProfit        string  `json:"total_profit"`
	TotalProfitPnl     string  `json:"total_profit_pnl"`
	AvgCost            string  `json:"avg_cost"`
	AvgSold            string  `json:"avg_sold"`
	Buy30d             int     `json:"buy_30d"`
	Sell30d            int     `json:"sell_30d"`
	Sells              int     `json:"sells"`
	Price              string  `json:"price"`
	Cost               string  `json:"cost"`
	PositionPercent    string  `json:"position_percent"`
	LastActiveTimestamp int64   `json:"last_active_timestamp"`
	Liquidity          string  `json:"liquidity"`
	TotalSupply        string  `json:"total_supply"`
}

// Token stat structures
type TokenStatResponse struct {
	HolderCount            int     `json:"holder_count"`
	BluechipOwnerCount     int     `json:"bluechip_owner_count"`
	BluechipOwnerPercentage string  `json:"bluechip_owner_percentage"`
	SignalCount            int     `json:"signal_count"`
	DegenCallCount         int     `json:"degen_call_count"`
	TopRatTraderPercentage string  `json:"top_rat_trader_percentage"`
}

// Token wallet tags statistics
type TokenWalletTagsStatResponse struct {
	Chain           string `json:"chain"`
	TokenAddress    string `json:"token_address"`
	SmartWallets    int    `json:"smart_wallets"`
	FreshWallets    int    `json:"fresh_wallets"`
	RenownedWallets int    `json:"renowned_wallets"`
	CreatorWallets  int    `json:"creator_wallets"`
	SniperWallets   int    `json:"sniper_wallets"`
	RatTraderWallets int   `json:"rat_trader_wallets"`
	WhaleWallets    int    `json:"whale_wallets"`
	TopWallets      int    `json:"top_wallets"`
	FollowingWallets int   `json:"following_wallets"`
	BundlerWallets  int    `json:"bundler_wallets"`
}

// Token holder statistics
type TokenHolderStatResponse struct {
	SmartDegenCount     int `json:"smart_degen_count"`
	RenownedCount       int `json:"renowned_count"`
	FreshWalletCount    int `json:"fresh_wallet_count"`
	DexBotCount         int `json:"dex_bot_count"`
	InsiderCount        int `json:"insider_count"`
	FollowingCount      int `json:"following_count"`
	DevCount            int `json:"dev_count"`
	BluechipOwnerCount  int `json:"bluechip_owner_count"`
	BundlerCount        int `json:"bundler_count"`
}

// Token price/kline structures
type KlineDataResponse struct {
	List []KlineData `json:"list"`
}

type KlineData struct {
	Time   string `json:"time"`
	Open   string `json:"open"`
	Close  string `json:"close"`
	High   string `json:"high"`
	Low    string `json:"low"`
	Volume string `json:"volume"`
}

// Trade history structures
type TradeHistoryResponse struct {
	History []TradeHistory `json:"history"`
}

type TradeHistory struct {
	Maker            string   `json:"maker"`
	BaseAmount       string   `json:"base_amount"`
	QuoteAmount      string   `json:"quote_amount"`
	QuoteSymbol      string   `json:"quote_symbol"`
	QuoteAddress     string   `json:"quote_address"`
	AmountUsd        string   `json:"amount_usd"`
	Timestamp        int64    `json:"timestamp"`
	Event            string   `json:"event"`
	TxHash           string   `json:"tx_hash"`
	PriceUsd         string   `json:"price_usd"`
	TotalTrade       int      `json:"total_trade"`
	ID               string   `json:"id"`
	IsFollowing      int      `json:"is_following"`
	IsOpenOrClose    int      `json:"is_open_or_close"`
	MakerTags        []string `json:"maker_tags"`
	TokenAddress     string   `json:"token_address"`
}

// Token traders structures
type Trader struct {
	Address             string  `json:"address"`
	AccountAddress      string  `json:"account_address"`
	AddrType            int     `json:"addr_type"`
	AmountCur           float64 `json:"amount_cur"`
	UsdValue            float64 `json:"usd_value"`
	CostCur             float64 `json:"cost_cur"`
	SellAmountCur       float64 `json:"sell_amount_cur"`
	SellAmountPercentage float64 `json:"sell_amount_percentage"`
	SellVolumeCur       float64 `json:"sell_volume_cur"`
	BuyVolumeCur        float64 `json:"buy_volume_cur"`
	BuyAmountCur        float64 `json:"buy_amount_cur"`
	NetflowUsd          float64 `json:"netflow_usd"`
	NetflowAmount       float64 `json:"netflow_amount"`
	BuyTxCountCur       int     `json:"buy_tx_count_cur"`
	SellTxCountCur      int     `json:"sell_tx_count_cur"`
	WalletTagV2         string  `json:"wallet_tag_v2"`
	NativeBalance       string  `json:"native_balance"`
	Balance             float64 `json:"balance"`
	Profit              float64 `json:"profit"`
	RealizedProfit      float64 `json:"realized_profit"`
	ProfitChange        float64 `json:"profit_change"`
	AmountPercentage    float64 `json:"amount_percentage"`
	UnrealizedProfit    float64 `json:"unrealized_profit"`
	AvgCost             float64 `json:"avg_cost"`
	AvgSold             float64 `json:"avg_sold"`
	LastActiveTimestamp int64   `json:"last_active_timestamp"`
	Tags                []string `json:"tags"`
}

// Trending token structures
type TrendingResponse struct {
	Code    int    `json:"code"`
	Msg     string `json:"msg"`
	Data    struct {
		Rank []TrendingToken `json:"rank"`
	} `json:"data"`
}

type TrendingToken struct {
	Address              string  `json:"address"`
	Name                 string  `json:"name"`
	Symbol               string  `json:"symbol"`
	Twitter              string  `json:"twitter"`
	Website              string  `json:"website"`
	Telegram             string  `json:"telegram"`
	CreatedTimestamp     int64   `json:"created_timestamp"`
	LastTradeTimestamp   int64   `json:"last_trade_timestamp"`
	UsdMarketCap         string  `json:"usd_market_cap"`
	Price                float64 `json:"price"`
	UpdatedAt            int64   `json:"updated_at"`
	Logo                 string  `json:"logo"`
	Volume1m             string  `json:"volume_1m"`
	Volume5m             string  `json:"volume_5m"`
	Volume1h             string  `json:"volume_1h"`
	Volume6h             string  `json:"volume_6h"`
	Volume24h            string  `json:"volume_24h"`
	Swaps1m              int     `json:"swaps_1m"`
	Swaps5m              int     `json:"swaps_5m"`
	Swaps1h              int     `json:"swaps_1h"`
	Swaps6h              int     `json:"swaps_6h"`
	Swaps24h             int     `json:"swaps_24h"`
	Buys1m               int     `json:"buys_1m"`
	Buys5m               int     `json:"buys_5m"`
	Buys1h               int     `json:"buys_1h"`
	Buys6h               int     `json:"buys_6h"`
	Buys24h              int     `json:"buys_24h"`
	Sells1m              int     `json:"sells_1m"`
	Sells5m              int     `json:"sells_5m"`
	Sells1h              int     `json:"sells_1h"`
	Sells6h              int     `json:"sells_6h"`
	Sells24h             int     `json:"sells_24h"`
	PriceChangePercent1m string  `json:"price_change_percent1m"`
	PriceChangePercent5m string  `json:"price_change_percent5m"`
	MarketCap1m          string  `json:"market_cap_1m"`
	MarketCap5m          string  `json:"market_cap_5m"`
	HolderCount          int     `json:"holder_count"`
	TotalSupply          int64   `json:"total_supply"`
	BotDegenCount        string  `json:"bot_degen_count"`
	Top10HolderRate      float64 `json:"top_10_holder_rate"`
	CreatorTokenStatus   string  `json:"creator_token_status"`
	CreatorClose         bool    `json:"creator_close"`
	CreatorTokenBalance  string  `json:"creator_token_balance"`
	KothDuration         int     `json:"koth_duration"`
	TimeSinceKoth        int     `json:"time_since_koth"`
	IsWashTrading        bool    `json:"is_wash_trading"`
}

// Completed token structures
type CompletedTokensResponse struct {
	Code    int    `json:"code"`
	Msg     string `json:"msg"`
	Data    struct {
		Rank []CompletedToken `json:"rank"`
	} `json:"data"`
}

type CompletedToken struct {
	Address             string  `json:"address"`
	Name                string  `json:"name"`
	Symbol              string  `json:"symbol"`
	Price               float64 `json:"price,string"`
	MarketCap           string  `json:"market_cap"`
	Volume              string  `json:"volume"`
	TotalSupply         float64 `json:"total_supply,string"`
	HolderCount         int     `json:"holder_count"`
	Status              string  `json:"status"`
	CompletedAt         int64   `json:"completed_at"`
	Twitter             string  `json:"twitter,omitempty"`
	Website             string  `json:"website,omitempty"`
	Telegram            string  `json:"telegram,omitempty"`
	Logo                string  `json:"logo,omitempty"`
	PriceChange         string  `json:"price_change,omitempty"`
	Top10HolderRate     float64 `json:"top_10_holder_rate"`
	CreatorTokenStatus  string  `json:"creator_token_status"`
	BotDegenCount       string  `json:"bot_degen_count"`
	IsWashTrading       bool    `json:"is_wash_trading"`
} 