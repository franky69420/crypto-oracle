package gmgn

import (
	"fmt"
	"time"

	"github.com/franky69420/crypto-oracle/pkg/models"
)

// Adapter wraps a GMGN client to implement the interface required by token engine
type Adapter struct {
	client Client
	useMock bool // Flag to use mock data instead of real API calls
}

// NewAdapter creates a new GMGN adapter
func NewAdapter(client Client) *Adapter {
	return &Adapter{
		client: client,
		useMock: false,
	}
}

// EnableMock enables the use of mock data instead of real API calls
func (a *Adapter) EnableMock() {
	a.useMock = true
}

// GetTokenInfo retrieves token information
func (a *Adapter) GetTokenInfo(tokenAddress string) (*models.Token, error) {
	if a.useMock {
		return a.getMockTokenInfo(tokenAddress)
	}

	// Get token stats which contains basic token info
	stats, err := a.client.GetTokenStat(tokenAddress)
	if err != nil {
		return nil, err
	}

	return &models.Token{
		Address:     tokenAddress,
		Symbol:      stats.Symbol,
		Name:        stats.Name,
		HolderCount: stats.Holders,
		Logo:        stats.Logo,
		CachedAt:    time.Now(),
	}, nil
}

// getMockTokenInfo returns mock token information
func (a *Adapter) getMockTokenInfo(tokenAddress string) (*models.Token, error) {
	// Map of known tokens for demo
	mockTokens := map[string]struct {
		Symbol string
		Name   string
	}{
		"SoLDogMKjM9YMzSQzp7SuBYQCM9LCCgBkrysTNxMD3m": {Symbol: "SDOGE", Name: "Solana Doge"},
		"CATZwdqR8Prd2RRK1mXnQvh698GziRn4Tw8zKcQfNPdS": {Symbol: "CATZ", Name: "Catz Token"},
		"DogXWzRCxXu7KcdAHuXtw8JP5YqwofHxjVdys3nwYzz": {Symbol: "DOG", Name: "DOG"},
		"PRT88RkA4Kg5z7pKnezeNH4mzoAjJ8Q8pCVUxnPQRmK": {Symbol: "PRT", Name: "Parrot Protocol"},
		"CUvVMqXAcyUMqJX14rMCQsXWHK19z5GVWu6MdwB5EiFa": {Symbol: "JITOSOL", Name: "JitoSol"},
	}

	token, exists := mockTokens[tokenAddress]
	if !exists {
		// Default for unknown tokens
		return &models.Token{
			Address:     tokenAddress,
			Symbol:      "UNKNOWN",
			Name:        "Unknown Token",
			HolderCount: 100,
			CachedAt:    time.Now(),
		}, nil
	}

	return &models.Token{
		Address:     tokenAddress,
		Symbol:      token.Symbol,
		Name:        token.Name,
		HolderCount: 500 + (len(token.Symbol) * 100), // Just to create some variation
		CachedAt:    time.Now(),
	}, nil
}

// GetTokenPrice retrieves token price information
func (a *Adapter) GetTokenPrice(tokenAddress string) (*models.TokenPrice, error) {
	if a.useMock {
		return a.getMockTokenPrice(tokenAddress)
	}

	klineData, err := a.client.GetTokenPrice(tokenAddress, "5m")
	if err != nil {
		return nil, err
	}

	// Get the most recent price data point
	var price float64
	if len(klineData.C) > 0 {
		price = klineData.C[len(klineData.C)-1]
	}

	return &models.TokenPrice{
		TokenAddress: tokenAddress,
		Price:        price,
		UpdatedAt:    time.Now(),
	}, nil
}

// getMockTokenPrice returns mock price data
func (a *Adapter) getMockTokenPrice(tokenAddress string) (*models.TokenPrice, error) {
	// Generate a price based on token address hash for consistency
	hashBase := 0
	for i, c := range tokenAddress {
		if i < 8 {
			hashBase += int(c)
		}
	}
	// Base price between $0.01 and $10.00
	price := (float64(hashBase % 1000) + 1.0) / 100.0

	return &models.TokenPrice{
		TokenAddress: tokenAddress,
		Price:        price,
		Change1h:     (float64(hashBase % 21) - 10) / 100.0, // -10% to +10%
		Change24h:    (float64(hashBase % 41) - 20) / 100.0, // -20% to +20%
		Volume24h:    float64(100000 + (hashBase % 900000)),
		MarketCap:    price * float64(1000000+(hashBase%9000000)),
		UpdatedAt:    time.Now(),
	}, nil
}

// GetTokenStats retrieves token statistics
func (a *Adapter) GetTokenStats(tokenAddress string) (*models.TokenStats, error) {
	if a.useMock {
		return a.getMockTokenStats(tokenAddress)
	}

	stats, err := a.client.GetTokenStat(tokenAddress)
	if err != nil {
		return nil, err
	}

	// Map GMGN stats to our internal model
	return &models.TokenStats{
		HolderCount:   stats.Holders,
		Price:         stats.Price,
		MarketCap:     stats.Mcap,
		Volume24h:     stats.Volume,
		// These fields might not be directly available from the API
		// For now, set reasonable placeholder values
		Volume1h:      stats.Volume / 24.0, // Approximate hourly volume
		PriceChange1h: stats.PriceChange / 24.0, // Approximate hourly change
		BuyCount1h:    10, // Placeholder
		SellCount1h:   8,  // Placeholder
	}, nil
}

// getMockTokenStats returns mock token statistics
func (a *Adapter) getMockTokenStats(tokenAddress string) (*models.TokenStats, error) {
	// Get the mock price for consistency
	priceData, _ := a.getMockTokenPrice(tokenAddress)

	// Generate hash-based variations
	hashBase := 0
	for i, c := range tokenAddress {
		if i < 8 {
			hashBase += int(c)
		}
	}

	return &models.TokenStats{
		HolderCount:   500 + (hashBase % 5000),
		Price:         priceData.Price,
		MarketCap:     priceData.MarketCap,
		Volume24h:     priceData.Volume24h,
		Volume1h:      priceData.Volume24h / 24.0,
		PriceChange1h: priceData.Change1h,
		BuyCount1h:    10 + (hashBase % 40),
		SellCount1h:   8 + (hashBase % 30),
		LiquidityUSD:  priceData.MarketCap * 0.1,
	}, nil
}

// GetTokenTrades retrieves token trade history
func (a *Adapter) GetTokenTrades(tokenAddress string, limit int) ([]models.TokenTrade, error) {
	if a.useMock {
		return a.getMockTokenTrades(tokenAddress, limit)
	}

	tradeResponse, err := a.client.GetTokenTrades(tokenAddress, limit, "")
	if err != nil {
		return nil, err
	}

	// Map GMGN trades to our internal model
	var result []models.TokenTrade
	for _, trade := range tradeResponse.List {
		result = append(result, models.TokenTrade{
			ID:            trade.ID,
			TokenAddress:  tokenAddress,
			WalletAddress: trade.WalletFrom, // Use WalletFrom as the main wallet
			TradeType:     trade.Type,
			Amount:        trade.TokenAmount,
			Price:         trade.Price,
			TotalValue:    trade.UsdAmount,
			Timestamp:     trade.Time,
			TxHash:        trade.TxHash,
			BlockNumber:   uint64(trade.BlockHeight),
		})
	}

	return result, nil
}

// getMockTokenTrades returns mock trade history
func (a *Adapter) getMockTokenTrades(tokenAddress string, limit int) ([]models.TokenTrade, error) {
	// Get mock token info and price for consistency
	tokenInfo, _ := a.getMockTokenInfo(tokenAddress)
	priceData, _ := a.getMockTokenPrice(tokenAddress)

	var result []models.TokenTrade
	now := time.Now()

	// Generate some mock wallets
	mockWallets := []string{
		"8xxa7L8dDT6vfJcKCGQJUyJU1xMFxGjkEcCXRRfLqQpr",
		"EWWKpWP65qzRDE7glTveiSCQe5jWX9NFxnkuZZutSLfT",
		"6FxYJn7ZQwCRyB4JcUfYZ9NPk1bBUJKyFwjXVxkEQyQg",
		"DYgCXwQ6KA3ZqTyL1vHSxmysR9Fpy5YfsCBBrMrLXTuF",
		"B2MvKUXL8FQjxJ4xqkiHQzA8LTMY6F7LTHjV8oHQkUvy",
	}

	// Generate mock trades
	for i := 0; i < limit; i++ {
		// Alternate buy/sell
		tradeType := "buy"
		if i%2 == 1 {
			tradeType = "sell"
		}

		// Randomize amounts based on token address hash
		hashBase := 0
		for j, c := range tokenAddress {
			if j < 8 {
				hashBase += int(c)
			}
		}
		
		// Calculate values as integers first, then convert to float64
		hashPlusI := hashBase + i
		amountInt := 100 + (hashPlusI % 10000)
		amount := float64(amountInt)
		
		priceVariation := 0.95 + (float64(hashPlusI % 10) / 100.0)
		price := priceData.Price * priceVariation
		
		result = append(result, models.TokenTrade{
			ID:            fmt.Sprintf("tx-%s-%d", tokenAddress[:8], i),
			TokenAddress:  tokenAddress,
			TokenSymbol:   tokenInfo.Symbol,
			WalletAddress: mockWallets[i%len(mockWallets)],
			TradeType:     tradeType,
			Amount:        amount,
			Price:         price,
			TotalValue:    amount * price,
			Timestamp:     now.Add(-time.Duration(i*5) * time.Minute),
			TxHash:        fmt.Sprintf("0x%x", hashBase+i),
			BlockNumber:   uint64(15000000 + i),
			Success:       true,
		})
	}

	return result, nil
}

// GetWalletTokenTrades retrieves trades for a specific wallet and token
func (a *Adapter) GetWalletTokenTrades(walletAddress string, tokenAddress string, limit int) ([]models.TokenTrade, error) {
	if a.useMock {
		// For mock, just return a subset of all trades that match this wallet
		allTrades, err := a.getMockTokenTrades(tokenAddress, limit*3)
		if err != nil {
			return nil, err
		}
		
		var walletTrades []models.TokenTrade
		for _, trade := range allTrades {
			if trade.WalletAddress == walletAddress {
				walletTrades = append(walletTrades, trade)
				if len(walletTrades) >= limit {
					break
				}
			}
		}
		
		// If we don't have enough matching trades, create some specifically for this wallet
		if len(walletTrades) < limit {
			tokenInfo, _ := a.getMockTokenInfo(tokenAddress)
			priceData, _ := a.getMockTokenPrice(tokenAddress)
			now := time.Now()
			
			for i := len(walletTrades); i < limit; i++ {
				tradeType := "buy"
				if i%2 == 1 {
					tradeType = "sell"
				}
				
				// Calculate values as integers first, then convert to float64
				amountInt := 100 + (i % 5000)
				amount := float64(amountInt)
				
				priceVariation := 0.97 + (float64(i % 6) / 100.0)
				price := priceData.Price * priceVariation
				
				walletTrades = append(walletTrades, models.TokenTrade{
					ID:            fmt.Sprintf("wtx-%s-%d", tokenAddress[:8], i),
					TokenAddress:  tokenAddress,
					TokenSymbol:   tokenInfo.Symbol,
					WalletAddress: walletAddress,
					TradeType:     tradeType,
					Amount:        amount,
					Price:         price,
					TotalValue:    amount * price,
					Timestamp:     now.Add(-time.Duration(i*15) * time.Minute),
					TxHash:        fmt.Sprintf("0x%x", i+1000),
					BlockNumber:   uint64(15000000 + i*10),
					Success:       true,
				})
			}
		}
		
		return walletTrades, nil
	}

	// Get all token trades
	allTrades, err := a.GetTokenTrades(tokenAddress, limit*10) // Get more trades to ensure we have enough wallet-specific ones
	if err != nil {
		return nil, err
	}

	// Filter for the specific wallet
	var walletTrades []models.TokenTrade
	for _, trade := range allTrades {
		if trade.WalletAddress == walletAddress {
			walletTrades = append(walletTrades, trade)
			if len(walletTrades) >= limit {
				break
			}
		}
	}

	return walletTrades, nil
}

// GetWalletHoldings retrieves a wallet's token holdings
func (a *Adapter) GetWalletHoldings(walletAddress string, limit int, showSmall bool) ([]models.WalletHolding, error) {
	if a.useMock {
		return a.getMockWalletHoldings(walletAddress, limit)
	}
	
	// This would call the GMGN API endpoint: /api/v1/wallet_holdings/sol/{wallet_address}
	// with appropriate pagination
	return nil, fmt.Errorf("not implemented")
}

// getMockWalletHoldings returns mock wallet holdings data
func (a *Adapter) getMockWalletHoldings(walletAddress string, limit int) ([]models.WalletHolding, error) {
	var holdings []models.WalletHolding
	mockTokens := []struct {
		Address string
		Symbol  string
		Name    string
	}{
		{Address: "SoLDogMKjM9YMzSQzp7SuBYQCM9LCCgBkrysTNxMD3m", Symbol: "SDOGE", Name: "Solana Doge"},
		{Address: "CATZwdqR8Prd2RRK1mXnQvh698GziRn4Tw8zKcQfNPdS", Symbol: "CATZ", Name: "Catz Token"},
		{Address: "DogXWzRCxXu7KcdAHuXtw8JP5YqwofHxjVdys3nwYzz", Symbol: "DOG", Name: "DOG"},
		{Address: "PRT88RkA4Kg5z7pKnezeNH4mzoAjJ8Q8pCVUxnPQRmK", Symbol: "PRT", Name: "Parrot Protocol"},
		{Address: "CUvVMqXAcyUMqJX14rMCQsXWHK19z5GVWu6MdwB5EiFa", Symbol: "JITOSOL", Name: "JitoSol"},
	}
	
	for i := 0; i < limit && i < len(mockTokens); i++ {
		token := mockTokens[i]
		
		// Create deterministic but variable mock data based on wallet address
		hashBase := 0
		for j, c := range walletAddress {
			if j < 8 {
				hashBase += int(c)
			}
		}
		balance := float64(1000 + ((hashBase + i) % 10000))
		price := (float64(hashBase % 1000) + 1.0) / 100.0
		value := balance * price
		
		holdings = append(holdings, models.WalletHolding{
			TokenAddress:     token.Address,
			TokenSymbol:      token.Symbol,
			Balance:          fmt.Sprintf("%.2f", balance),
			Value:            fmt.Sprintf("%.2f", value),
			UnrealizedProfit: value * 0.2 * float64(1+((hashBase+i)%5))/5.0,
			BuyCount:         3 + (hashBase+i)%8,
			SellCount:        1 + (hashBase+i)%5,
			LastActive:       time.Now().Add(-time.Duration((hashBase+i)%72) * time.Hour),
		})
	}
	
	return holdings, nil
}

// GetWalletDailyProfit retrieves daily profit for a wallet
func (a *Adapter) GetWalletDailyProfit(walletAddress string, period string) ([]models.DailyProfit, error) {
	if a.useMock {
		return a.getMockWalletDailyProfit(walletAddress, period)
	}
	
	// This would call the GMGN API endpoint: /api/v1/daily_profit/sol/{wallet_address}/{period}
	return nil, fmt.Errorf("not implemented")
}

// getMockWalletDailyProfit returns mock daily profit data
func (a *Adapter) getMockWalletDailyProfit(walletAddress string, period string) ([]models.DailyProfit, error) {
	var profits []models.DailyProfit
	
	// Determine days based on period
	days := 30
	if period == "7d" {
		days = 7
	} else if period == "30d" {
		days = 30
	} else if period == "90d" {
		days = 90
	}
	
	// Create mock daily profits
	for i := 0; i < days; i++ {
		date := time.Now().Add(-time.Duration(i) * 24 * time.Hour)
		
		// Generate deterministic but variable data based on wallet address and date
		hashBase := 0
		for j, c := range walletAddress {
			if j < 8 {
				hashBase += int(c)
			}
		}
		dateHash := int(date.Day() + int(date.Month())*31)
		combined := hashBase + dateHash
		
		// Some days positive, some negative
		profitMultiplier := 1.0
		if combined%3 == 0 {
			profitMultiplier = -0.5
		}
		
		profit := float64(50 + (combined % 950)) * profitMultiplier
		volume := float64(500 + (combined % 9500))
		txCount := 5 + (combined % 20)
		
		profits = append(profits, models.DailyProfit{
			Date:            date,
			Profit:          profit,
			Volume:          volume,
			TransactionCount: txCount,
			WalletAddress:   walletAddress,
		})
	}
	
	return profits, nil
}

// GetWalletStats retrieves statistics for a wallet
func (a *Adapter) GetWalletStats(walletAddress string, period string) (*models.WalletStats, error) {
	if a.useMock {
		return a.getMockWalletStats(walletAddress, period)
	}
	
	// This would call the GMGN API endpoint: /api/v1/wallet_stat/sol/{wallet_address}/{period}
	return nil, fmt.Errorf("not implemented")
}

// getMockWalletStats returns mock wallet statistics
func (a *Adapter) getMockWalletStats(walletAddress string, period string) (*models.WalletStats, error) {
	// Generate deterministic but variable data based on wallet address
	hashBase := 0
	for i, c := range walletAddress {
		if i < 8 {
			hashBase += int(c)
		}
	}
	
	// Get mock daily profits to sum up total profit and volume
	dailyProfits, _ := a.getMockWalletDailyProfit(walletAddress, period)
	totalProfit := 0.0
	totalVolume := 0.0
	for _, profit := range dailyProfits {
		totalProfit += profit.Profit
		totalVolume += profit.Volume
	}
	
	// Calculate other stats based on wallet address hash
	tokenCount := 5 + (hashBase % 20)
	txCount := 50 + (hashBase % 200)
	winRate := 50.0 + float64(hashBase%40)
	avgHoldTime := 24.0 + float64(hashBase%72)
	biggestWin := totalProfit * 0.3 * (1.0 + float64(hashBase%5)/5.0)
	biggestLoss := totalProfit * -0.15 * (1.0 + float64(hashBase%5)/5.0)
	
	// First and last transaction times
	now := time.Now()
	days := 90 + (hashBase % 275)
	firstTx := now.Add(-time.Duration(days) * 24 * time.Hour)
	lastTx := now.Add(-time.Duration(hashBase%24) * time.Hour)
	
	return &models.WalletStats{
		WalletAddress:     walletAddress,
		TotalProfit:       totalProfit,
		TotalVolume:       totalVolume,
		WinRate:           winRate,
		TokenCount:        tokenCount,
		TransactionCount:  txCount,
		FirstTransaction:  firstTx,
		LastTransaction:   lastTx,
		AverageHoldTime:   avgHoldTime,
		BiggestWin:        biggestWin,
		BiggestLoss:       biggestLoss,
		UpdatedAt:         time.Now(),
	}, nil
}

// GetTopWallets retrieves top wallets by various metrics
func (a *Adapter) GetTopWallets(period string, orderBy string, direction string, limit int) ([]models.WalletRanking, error) {
	if a.useMock {
		return a.getMockTopWallets(period, orderBy, limit)
	}
	
	// This would call the GMGN API endpoint: /defi/quotation/v1/rank/sol/wallets/{period}
	return nil, fmt.Errorf("not implemented")
}

// getMockTopWallets returns mock top wallets data
func (a *Adapter) getMockTopWallets(period string, orderBy string, limit int) ([]models.WalletRanking, error) {
	var wallets []models.WalletRanking
	
	// Mock wallet addresses
	mockWallets := []string{
		"8xxa7L8dDT6vfJcKCGQJUyJU1xMFxGjkEcCXRRfLqQpr",
		"EWWKpWP65qzRDE7glTveiSCQe5jWX9NFxnkuZZutSLfT",
		"6FxYJn7ZQwCRyB4JcUfYZ9NPk1bBUJKyFwjXVxkEQyQg",
		"DYgCXwQ6KA3ZqTyL1vHSxmysR9Fpy5YfsCBBrMrLXTuF",
		"B2MvKUXL8FQjxJ4xqkiHQzA8LTMY6F7LTHjV8oHQkUvy",
		"FVDt1RxkMfSrPJ4u4rvHG6YZvxwCVmcMoKVP5Fua1zKT",
		"GMV5TkRQZBUQqQpwXEUKLQFdXzWtgx3yLdB4TqEWaPzt",
		"HneEB7oNTbWBZ2JNX5kTpkYFu7sNgdmeNz8VCNH7pNQZ",
		"JLKMYcHZa9y9aTVKPHQg2RWAyouLbpdE6wVYhfuH6kxj",
		"KPqzXtVKbJJFkHaQSZJgLEpzSy5RqyShZejvjiJcAJK2",
	}
	
	for i := 0; i < limit && i < len(mockWallets); i++ {
		walletAddr := mockWallets[i]
		
		// Generate deterministic but variable data based on wallet address
		hashBase := 0
		for j, c := range walletAddr {
			if j < 8 {
				hashBase += int(c)
			}
		}
		
		// Mock stats
		totalProfit := float64(10000 + (hashBase % 990000))
		totalVolume := float64(100000 + (hashBase % 9900000))
		winRate := 50.0 + float64(hashBase%40)
		tokenCount := 10 + (hashBase % 90)
		trustScore := 60.0 + float64(hashBase%30)
		
		// Tags based on hash
		var tags []string
		if hashBase%4 == 0 {
			tags = append(tags, "smart_money")
		}
		if hashBase%3 == 0 {
			tags = append(tags, "early_investor") 
		}
		if hashBase%5 == 0 {
			tags = append(tags, "blue_chip")
		}
		
		// Last active time
		daysAgo := hashBase % 10
		lastActive := time.Now().Add(-time.Duration(daysAgo) * 24 * time.Hour)
		
		wallets = append(wallets, models.WalletRanking{
			Rank:          i + 1,
			WalletAddress: walletAddr,
			WalletName:    fmt.Sprintf("Wallet%d", i+1),
			TotalProfit:   totalProfit,
			TotalVolume:   totalVolume,
			WinRate:       winRate,
			TokenCount:    tokenCount,
			TrustScore:    trustScore,
			Tags:          tags,
			TimePeriod:    period,
			LastActive:    lastActive,
		})
	}
	
	return wallets, nil
}

// GetPumpRankings retrieves tokens currently pumping
func (a *Adapter) GetPumpRankings(timeframe string, limit int) (*models.PumpRankings, error) {
	if a.useMock {
		return a.getMockPumpRankings(timeframe, limit)
	}
	
	// This would call the GMGN API endpoint: /defi/quotation/v1/rank/sol/pump_ranks/{timeframe}
	return nil, fmt.Errorf("not implemented")
}

// getMockPumpRankings returns mock pump rankings data
func (a *Adapter) getMockPumpRankings(timeframe string, limit int) (*models.PumpRankings, error) {
	// Mock token data
	mockTokens := []struct {
		Address string
		Symbol  string
		Name    string
	}{
		{Address: "SoLDogMKjM9YMzSQzp7SuBYQCM9LCCgBkrysTNxMD3m", Symbol: "SDOGE", Name: "Solana Doge"},
		{Address: "CATZwdqR8Prd2RRK1mXnQvh698GziRn4Tw8zKcQfNPdS", Symbol: "CATZ", Name: "Catz Token"},
		{Address: "DogXWzRCxXu7KcdAHuXtw8JP5YqwofHxjVdys3nwYzz", Symbol: "DOG", Name: "DOG"},
		{Address: "PRT88RkA4Kg5z7pKnezeNH4mzoAjJ8Q8pCVUxnPQRmK", Symbol: "PRT", Name: "Parrot Protocol"},
		{Address: "CUvVMqXAcyUMqJX14rMCQsXWHK19z5GVWu6MdwB5EiFa", Symbol: "JITOSOL", Name: "JitoSol"},
		{Address: "EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v", Symbol: "USDC", Name: "USD Coin"},
		{Address: "7Vbe8fNJJnpBE2SJQQbSraV4EzKNr8rkQmmYsHXCgkSg", Symbol: "BONK", Name: "Bonk"},
		{Address: "So11111111111111111111111111111111111111112", Symbol: "SOL", Name: "Wrapped SOL"},
		{Address: "MNDEFzGvMt87ueuHvVU9VcTqsAP5b3fTGPsHuuPA5ey", Symbol: "MNDE", Name: "Marinade"},
		{Address: "9vMJfxuKxXBoEa7rM12mYLMwTacLMLDJqHozw96WQL8i", Symbol: "UST", Name: "UST (Wormhole)"},
	}
	
	var rankings []models.PumpToken
	now := time.Now()
	
	// Convert timeframe for calculation
	var timeframeMultiplier float64
	switch timeframe {
	case "1h":
		timeframeMultiplier = 0.3
	case "7d":
		timeframeMultiplier = 2.5
	default:
		timeframeMultiplier = 1.0 // Default to 24h
	}
	
	for i := 0; i < limit && i < len(mockTokens); i++ {
		token := mockTokens[i]
		
		// Generate deterministic but variable data based on token address
		hashBase := 0
		for j, c := range token.Address {
			if j < 8 {
				hashBase += int(c)
			}
		}
		
		// Mock price and changes based on timeframe
		basePrice := (float64(hashBase%1000) + 1.0) / 100.0
		
		// More volatility for lower ranked tokens
		rankMultiplier := float64(limit-i) / float64(limit)
		
		priceChange := float64(10+hashBase%90) * timeframeMultiplier * rankMultiplier
		volumeChange := float64(20+hashBase%180) * timeframeMultiplier * rankMultiplier
		
		// Volume and market cap based on price
		volume := basePrice * float64(100000+hashBase%900000) * timeframeMultiplier
		marketCap := basePrice * float64(1000000+hashBase%9000000)
		holderCount := 1000 + (hashBase % 9000)
		
		// Create time is older for higher ranked tokens
		createTime := now.Add(-time.Duration(30+i*10+hashBase%30) * 24 * time.Hour)
		
		rankings = append(rankings, models.PumpToken{
			Rank:         i + 1,
			TokenAddress: token.Address,
			TokenSymbol:  token.Symbol,
			TokenName:    token.Name,
			Price:        basePrice,
			PriceChange:  priceChange,
			Volume:       volume,
			VolumeChange: volumeChange,
			MarketCap:    marketCap,
			HolderCount:  holderCount,
			CreateTime:   createTime,
			UpdatedAt:    now,
		})
	}
	
	return &models.PumpRankings{
		Timeframe:   timeframe,
		UpdatedAt:   now,
		TotalTokens: len(rankings),
		Rankings:    rankings,
	}, nil
}

// GetTokenTopBuyers retrieves top buyers for a token
func (a *Adapter) GetTokenTopBuyers(tokenAddress string) (*models.TokenHolders, error) {
	if a.useMock {
		return a.getMockTokenTopBuyers(tokenAddress)
	}
	
	// This would call the GMGN API endpoint: /defi/quotation/v1/tokens/top_buyers/sol/{token_address}
	return nil, fmt.Errorf("not implemented")
}

// getMockTokenTopBuyers returns mock token top buyers data
func (a *Adapter) getMockTokenTopBuyers(tokenAddress string) (*models.TokenHolders, error) {
	// Get token info for consistency
	tokenInfo, _ := a.getMockTokenInfo(tokenAddress)
	
	// Generate mock wallets
	mockWallets := []string{
		"8xxa7L8dDT6vfJcKCGQJUyJU1xMFxGjkEcCXRRfLqQpr",
		"EWWKpWP65qzRDE7glTveiSCQe5jWX9NFxnkuZZutSLfT",
		"6FxYJn7ZQwCRyB4JcUfYZ9NPk1bBUJKyFwjXVxkEQyQg",
		"DYgCXwQ6KA3ZqTyL1vHSxmysR9Fpy5YfsCBBrMrLXTuF",
		"B2MvKUXL8FQjxJ4xqkiHQzA8LTMY6F7LTHjV8oHQkUvy",
		"FVDt1RxkMfSrPJ4u4rvHG6YZvxwCVmcMoKVP5Fua1zKT",
		"GMV5TkRQZBUQqQpwXEUKLQFdXzWtgx3yLdB4TqEWaPzt",
		"HneEB7oNTbWBZ2JNX5kTpkYFu7sNgdmeNz8VCNH7pNQZ",
		"JLKMYcHZa9y9aTVKPHQg2RWAyouLbpdE6wVYhfuH6kxj",
		"KPqzXtVKbJJFkHaQSZJgLEpzSy5RqyShZejvjiJcAJK2",
	}
	
	// Mock price for calculations
	priceInfo, _ := a.getMockTokenPrice(tokenAddress)
	
	var topBuyers []models.TokenHolder
	var topHolders []models.TokenHolder
	
	now := time.Now()
	
	for i := 0; i < len(mockWallets); i++ {
		wallet := mockWallets[i]
		
		// Generate deterministic but variable data based on wallet and token
		hashBase := 0
		for j, c := range wallet {
			if j < 4 {
				hashBase += int(c)
			}
		}
		for j, c := range tokenAddress {
			if j < 4 {
				hashBase += int(c)
			}
		}
		
		// Top holders have more balance
		ranking := len(mockWallets) - i
		balance := float64(10000*ranking + hashBase%90000)
		value := balance * priceInfo.Price
		percentOwned := value / priceInfo.MarketCap * 100
		
		buyAmount := balance * 1.2
		sellAmount := balance * 0.2
		buyCount := 5 + (hashBase % 20)
		sellCount := 2 + (hashBase % 10)
		
		// Times
		buyDaysAgo := 1 + (hashBase % 30)
		actionDaysAgo := hashBase % 7
		firstBuy := now.Add(-time.Duration(buyDaysAgo) * 24 * time.Hour)
		lastAction := now.Add(-time.Duration(actionDaysAgo) * 24 * time.Hour)
		
		// Trust score and tags
		trustScore := 50.0 + float64(hashBase%40)
		
		var tags []string
		if hashBase%4 == 0 {
			tags = append(tags, "smart_money")
		}
		if hashBase%3 == 0 {
			tags = append(tags, "early_investor") 
		}
		if hashBase%5 == 0 {
			tags = append(tags, "blue_chip")
		}
		
		holder := models.TokenHolder{
			WalletAddress: wallet,
			WalletName:    fmt.Sprintf("Wallet%d", i+1),
			Balance:       balance,
			Value:         value,
			PercentOwned:  percentOwned,
			BuyAmount:     buyAmount,
			SellAmount:    sellAmount,
			BuyCount:      buyCount,
			SellCount:     sellCount,
			FirstBuy:      firstBuy,
			LastAction:    lastAction,
			TrustScore:    trustScore,
			Tags:          tags,
		}
		
		// Add to appropriate lists
		if i < 5 {
			topBuyers = append(topBuyers, holder)
		}
		topHolders = append(topHolders, holder)
	}
	
	return &models.TokenHolders{
		TokenAddress: tokenAddress,
		TokenSymbol:  tokenInfo.Symbol,
		TotalHolders: 1000 + len(mockWallets),
		TopBuyers:    topBuyers,
		TopHolders:   topHolders,
		UpdatedAt:    now,
	}, nil
}

// GetSimilarCoinAnalysis retrieves similar coin analysis
func (a *Adapter) GetSimilarCoinAnalysis(symbol string, address string) (*models.SimilarCoinAnalysis, error) {
	if a.useMock {
		return a.getMockSimilarCoinAnalysis(symbol, address)
	}
	
	// This would call the GMGN API endpoint: /vas/api/v1/similar_coin_max_and_earliest
	return nil, fmt.Errorf("not implemented")
}

// getMockSimilarCoinAnalysis returns mock similar coin analysis data
func (a *Adapter) getMockSimilarCoinAnalysis(symbol string, address string) (*models.SimilarCoinAnalysis, error) {
	// Get token info for consistency
	tokenInfo, _ := a.getMockTokenInfo(address)
	
	// Mock similar tokens
	mockSimilarTokens := []struct {
		Address string
		Symbol  string
		Name    string
	}{
		{Address: "7Vbe8fNJJnpBE2SJQQbSraV4EzKNr8rkQmmYsHXCgkSg", Symbol: "BONK", Name: "Bonk"},
		{Address: "F9qCnm5z1f4TVLjZS8yxGWiYrJECFvUfPEskpXv4DxWF", Symbol: "BOOK", Name: "Book Token"},
		{Address: "3oTHPpKMiQnZUPR9iELs1pLv4ZWtJfGV8EprVASSLajL", Symbol: "MONG", Name: "Mong"},
		{Address: "A2P1owGNgRRGZ5KGVLmCgmWAeRjw8iZZhr6NqPnbZQ5U", Symbol: "KONG", Name: "Kong"},
		{Address: "3eZQcJQUhrNJ7bJJiS1uGAgUryvXtkRjNFyoVJdSSnxs", Symbol: "ELON", Name: "Elonium"},
		{Address: "2HeykdKjpEXkVq46j2xPeXcMqGqHEfPRgxjnDv6dB7MZ", Symbol: "MOON", Name: "MoonToken"},
		{Address: "HLLihvavVr6nB1JwZxoKk8GFvJ8MJc1jzD1LTxv8mQAU", Symbol: "ROCKS", Name: "Rocks"},
		{Address: "Eh4V2SRWNWu8VrNGEb5n2gwaCvkMwNP8XGp5ZNXsWEjg", Symbol: "WIF", Name: "WIF"},
	}
	
	var maxGains []models.SimilarCoin
	var earliest []models.SimilarCoin
	now := time.Now()
	
	for i, token := range mockSimilarTokens {
		// Generate deterministic but variable data based on token addresses
		hashBase := 0
		for j, c := range token.Address {
			if j < 4 {
				hashBase += int(c)
			}
		}
		for j, c := range address {
			if j < 4 {
				hashBase += int(c)
			}
		}
		
		// Calculate similarity score (60-95%)
		similarity := 60.0 + float64(hashBase%35)
		
		// Initial price and peak price
		initialPrice := float64(1+hashBase%99) / 10000.0
		multiplier := float64(10 + hashBase%990)
		peakPrice := initialPrice * multiplier
		
		// Dates
		// Earlier coins launched earlier
		launchDaysAgo := 60 + (hashBase % 120) + (i * 15)
		peakDaysAgo := 10 + (hashBase % 50)
		launchDate := now.Add(-time.Duration(launchDaysAgo) * 24 * time.Hour)
		peakDate := now.Add(-time.Duration(peakDaysAgo) * 24 * time.Hour)
		
		// Time to reach multiplier (hours to days)
		timeToMultiplier := models.Duration(fmt.Sprintf("%dh", 12+(hashBase%240)))
		
		similarCoin := models.SimilarCoin{
			TokenAddress:    token.Address,
			TokenSymbol:     token.Symbol,
			TokenName:       token.Name,
			Similarity:      similarity,
			InitialPrice:    initialPrice,
			PeakPrice:       peakPrice,
			PriceMultiplier: multiplier,
			TimeToMultiplier: timeToMultiplier,
			LaunchDate:      launchDate,
			PeakDate:        peakDate,
		}
		
		// Add to appropriate lists - first half to max gains, second half to earliest
		if i < len(mockSimilarTokens)/2 {
			maxGains = append(maxGains, similarCoin)
		} else {
			earliest = append(earliest, similarCoin)
		}
	}
	
	return &models.SimilarCoinAnalysis{
		TokenAddress: address,
		TokenSymbol:  tokenInfo.Symbol,
		TokenName:    tokenInfo.Name,
		Timeframe:    "all",
		MaximumGains: maxGains,
		EarliestCoins: earliest,
		UpdatedAt:    now,
	}, nil
}

// GetTokenWalletTagsStats retrieves wallet tag statistics for a token
func (a *Adapter) GetTokenWalletTagsStats(tokenAddress string) (*models.TokenWalletTagsStats, error) {
	if a.useMock {
		return a.getMockTokenWalletTagsStats(tokenAddress)
	}
	
	// This would call the GMGN API endpoint: /api/v1/token_wallet_tags_stat/sol/{token_address}
	return nil, fmt.Errorf("not implemented")
}

// getMockTokenWalletTagsStats returns mock token wallet tags statistics
func (a *Adapter) getMockTokenWalletTagsStats(tokenAddress string) (*models.TokenWalletTagsStats, error) {
	// Get token info for consistency
	tokenInfo, _ := a.getMockTokenInfo(tokenAddress)
	
	// Available tags
	tags := []string{"smart_money", "early_investor", "blue_chip", "whale", "degen", "arbitrage_bot", "mev_bot", "sniper", "developer", "team"}
	
	// Generate deterministic but variable data based on token address
	hashBase := 0
	for i, c := range tokenAddress {
		if i < 8 {
			hashBase += int(c)
		}
	}
	
	// Total holders
	totalHolders := 1000 + (hashBase % 9000)
	
	// Generate counts and percentages for each tag
	tagsCount := make(map[string]int)
	tagsPercentage := make(map[string]float64)
	tagsDetails := make(map[string][]models.TaggedWallet)
	
	for i, tag := range tags {
		// Each tag gets a different count based on token and tag index
		tagHash := hashBase + i
		count := totalHolders / 10 + (tagHash % (totalHolders / 5))
		percentage := float64(count) / float64(totalHolders) * 100
		
		tagsCount[tag] = count
		tagsPercentage[tag] = percentage
		
		// Generate some sample wallets for each tag (max 3)
		var taggedWallets []models.TaggedWallet
		
		maxWallets := 3
		if count < 3 {
			maxWallets = count
		}
		
		for j := 0; j < maxWallets; j++ {
			// Create wallet address based on tag and index
			walletAddr := fmt.Sprintf("%s%d%s", 
				tokenAddress[:8], 
				j, 
				tag[:4])
			
			// Generate mock data for this wallet
			balance := float64(10000 + (tagHash+j) % 90000)
			value := balance * (float64(hashBase%100) / 100.0 + 0.5) // $0.50-$1.50 per token
			
			// First buy and last action times
			buyDaysAgo := 10 + ((tagHash+j) % 90)
			actionDaysAgo := (tagHash+j) % 10
			
			firstBuy := time.Now().Add(-time.Duration(buyDaysAgo) * 24 * time.Hour)
			lastAction := time.Now().Add(-time.Duration(actionDaysAgo) * 24 * time.Hour)
			
			// This wallet's tags (include the main tag plus potentially 1-2 others)
			walletTags := []string{tag}
			if (tagHash+j)%3 == 0 && len(tags) > i+1 {
				walletTags = append(walletTags, tags[i+1])
			}
			if (tagHash+j)%5 == 0 && i > 0 {
				walletTags = append(walletTags, tags[i-1])
			}
			
			taggedWallets = append(taggedWallets, models.TaggedWallet{
				WalletAddress: walletAddr,
				Tags:          walletTags,
				Balance:       balance,
				Value:         value,
				FirstBuy:      firstBuy,
				LastAction:    lastAction,
			})
		}
		
		tagsDetails[tag] = taggedWallets
	}
	
	return &models.TokenWalletTagsStats{
		TokenAddress:   tokenAddress,
		TokenSymbol:    tokenInfo.Symbol,
		TotalHolders:   totalHolders,
		TagsCount:      tagsCount,
		TagsPercentage: tagsPercentage,
		TagsDetails:    tagsDetails,
		UpdatedAt:      time.Now(),
	}, nil
}

// GetTokenTraders retrieves traders for a token
func (a *Adapter) GetTokenTraders(tokenAddress string, limit int, orderBy string, direction string) ([]models.TokenTrader, error) {
	if a.useMock {
		return a.getMockTokenTraders(tokenAddress, limit)
	}
	
	// This would call the GMGN API endpoint: /vas/api/v1/token_traders/sol/{token_address}
	return nil, fmt.Errorf("not implemented")
}

// getMockTokenTraders returns mock token traders data
func (a *Adapter) getMockTokenTraders(tokenAddress string, limit int) ([]models.TokenTrader, error) {
	var traders []models.TokenTrader
	
	// Mock wallet addresses
	mockWallets := []string{
		"8xxa7L8dDT6vfJcKCGQJUyJU1xMFxGjkEcCXRRfLqQpr",
		"EWWKpWP65qzRDE7glTveiSCQe5jWX9NFxnkuZZutSLfT",
		"6FxYJn7ZQwCRyB4JcUfYZ9NPk1bBUJKyFwjXVxkEQyQg",
		"DYgCXwQ6KA3ZqTyL1vHSxmysR9Fpy5YfsCBBrMrLXTuF",
		"B2MvKUXL8FQjxJ4xqkiHQzA8LTMY6F7LTHjV8oHQkUvy",
		"FVDt1RxkMfSrPJ4u4rvHG6YZvxwCVmcMoKVP5Fua1zKT",
		"GMV5TkRQZBUQqQpwXEUKLQFdXzWtgx3yLdB4TqEWaPzt",
		"HneEB7oNTbWBZ2JNX5kTpkYFu7sNgdmeNz8VCNH7pNQZ",
		"JLKMYcHZa9y9aTVKPHQg2RWAyouLbpdE6wVYhfuH6kxj",
		"KPqzXtVKbJJFkHaQSZJgLEpzSy5RqyShZejvjiJcAJK2",
	}
	
	for i := 0; i < limit && i < len(mockWallets); i++ {
		wallet := mockWallets[i]
		
		// Generate deterministic but variable data based on wallet and token
		hashBase := 0
		for j, c := range wallet {
			if j < 4 {
				hashBase += int(c)
			}
		}
		for j, c := range tokenAddress {
			if j < 4 {
				hashBase += int(c)
			}
		}
		
		// Calculate relative volume (higher for first few wallets)
		relativeVolume := float64(10 - i + (hashBase % 10)) / 10.0
		
		// Early investor score (0-1)
		earlyInvestor := float64(10 - i + (hashBase % 5)) / 10.0
		
		// Transaction count
		txCount := 10 + (hashBase % 90) - (i * 5)
		if txCount < 5 {
			txCount = 5
		}
		
		traders = append(traders, models.TokenTrader{
			WalletAddress:    wallet,
			TokenAddress:     tokenAddress,
			RelativeVolume:   relativeVolume,
			EarlyInvestor:    earlyInvestor,
			TransactionCount: txCount,
		})
	}
	
	return traders, nil
}

// GetTokenHolderStats retrieves holder statistics for a token
func (a *Adapter) GetTokenHolderStats(tokenAddress string) (*models.TokenHolderStats, error) {
	if a.useMock {
		return a.getMockTokenHolderStats(tokenAddress)
	}
	
	// This would call the GMGN API endpoint: /vas/api/v1/token_holder_stat/sol/{token_address}
	return nil, fmt.Errorf("not implemented")
}

// getMockTokenHolderStats returns mock token holder statistics
func (a *Adapter) getMockTokenHolderStats(tokenAddress string) (*models.TokenHolderStats, error) {
	// Generate deterministic but variable data based on token address
	hashBase := 0
	for i, c := range tokenAddress {
		if i < 8 {
			hashBase += int(c)
		}
	}
	
	// Get token info for consistency
	tokenInfo, _ := a.getMockTokenInfo(tokenAddress)
	totalHolders := 1000 + (hashBase % 9000)
	
	// Create holder distribution by amount buckets
	holdersByAmount := map[string]int{
		"0-100":      totalHolders / 2 + (hashBase % 100),
		"100-1K":     totalHolders / 4 + (hashBase % 50),
		"1K-10K":     totalHolders / 10 + (hashBase % 30),
		"10K-100K":   totalHolders / 20 + (hashBase % 20),
		"100K-1M":    totalHolders / 100 + (hashBase % 10),
		"1M+":        totalHolders / 200 + (hashBase % 5),
	}
	
	// Create holder distribution by time buckets
	holdersByTime := map[string]int{
		"<1d":        totalHolders / 10 + (hashBase % 50),
		"1d-7d":      totalHolders / 5 + (hashBase % 100),
		"7d-30d":     totalHolders / 3 + (hashBase % 150),
		"30d-90d":    totalHolders / 4 + (hashBase % 100),
		"90d+":       totalHolders / 6 + (hashBase % 50),
	}
	
	// Create holder distribution by value buckets
	holdersByValue := map[string]int{
		"<$10":       totalHolders / 3 + (hashBase % 100),
		"$10-$100":   totalHolders / 3 + (hashBase % 100),
		"$100-$1K":   totalHolders / 6 + (hashBase % 50),
		"$1K-$10K":   totalHolders / 15 + (hashBase % 30),
		"$10K-$100K": totalHolders / 50 + (hashBase % 10),
		"$100K+":     totalHolders / 100 + (hashBase % 5),
	}
	
	// Create token ownership distribution
	distribution := map[string]float64{
		"top_1":        float64(15 + (hashBase % 10)),
		"top_10":       float64(40 + (hashBase % 20)),
		"top_50":       float64(70 + (hashBase % 15)),
		"top_100":      float64(85 + (hashBase % 10)),
		"remaining":    float64(15 - (hashBase % 10)),
	}
	
	// Activity counts
	buyerCount := totalHolders / 3 * 2 + (hashBase % 100)
	sellerCount := totalHolders / 6 + (hashBase % 50)
	activeWallets := totalHolders / 2 + (hashBase % (totalHolders / 4))
	
	return &models.TokenHolderStats{
		TokenAddress:    tokenAddress,
		TokenSymbol:     tokenInfo.Symbol,
		TotalHolders:    totalHolders,
		HoldersByAmount: holdersByAmount,
		HoldersByTime:   holdersByTime,
		HoldersByValue:  holdersByValue,
		Distribution:    distribution,
		BuyerCount:      buyerCount,
		SellerCount:     sellerCount,
		ActiveWallets:   activeWallets,
		UpdatedAt:       time.Now(),
	}, nil
}

// GetTokenMarketCapCandles retrieves market cap candle data for a token
func (a *Adapter) GetTokenMarketCapCandles(tokenAddress string, resolution string, from int64, to int64, limit int) ([]models.MarketCapCandle, error) {
	if a.useMock {
		return a.getMockTokenMarketCapCandles(tokenAddress, resolution, limit)
	}
	
	// This would call the GMGN API endpoint: /api/v1/token_mcap_candles/sol/{token_address}
	return nil, fmt.Errorf("not implemented")
}

// getMockTokenMarketCapCandles returns mock token market cap candles data
func (a *Adapter) getMockTokenMarketCapCandles(tokenAddress string, resolution string, limit int) ([]models.MarketCapCandle, error) {
	var candles []models.MarketCapCandle
	
	// Convert resolution to duration
	var interval time.Duration
	switch resolution {
	case "5m":
		interval = 5 * time.Minute
	case "15m":
		interval = 15 * time.Minute
	case "1h":
		interval = time.Hour
	case "4h":
		interval = 4 * time.Hour
	case "1d":
		interval = 24 * time.Hour
	default:
		interval = time.Hour // Default to 1h
	}
	
	// Get token price for base values
	priceInfo, _ := a.getMockTokenPrice(tokenAddress)
	basePrice := priceInfo.Price
	baseMarketCap := priceInfo.MarketCap
	
	// Generate deterministic but variable data based on token address
	hashBase := 0
	for i, c := range tokenAddress {
		if i < 8 {
			hashBase += int(c)
		}
	}
	
	// Generate candles going backward from now
	now := time.Now()
	
	for i := 0; i < limit; i++ {
		timestamp := now.Add(-time.Duration(i) * interval)
		
		// Create some variability based on token address and timestamp
		timeHash := int(timestamp.Unix() % 100)
		variability := float64(1.0 + float64((hashBase+timeHash)%20-10)/100.0)
		
		// Price generally trends downward as we go back in time (newer prices are higher)
		trendFactor := 1.0 - float64(i)/float64(limit) * 0.3 // max 30% lower at start
		price := basePrice * variability * trendFactor
		
		// Calculate OHLC with small variations
		open := price * (1.0 - float64((hashBase+timeHash)%10)/100.0)
		high := price * (1.0 + float64((hashBase+timeHash)%15)/100.0)
		low := price * (1.0 - float64((hashBase+timeHash)%15)/100.0)
		close := price
		
		// Volume fluctuates more
		volume := basePrice * float64(10000+((hashBase+timeHash)%90000)) * variability
		
		// Market cap follows price pattern
		marketCap := baseMarketCap * variability * trendFactor
		marketCapOpen := marketCap * (1.0 - float64((hashBase+timeHash)%10)/100.0)
		marketCapHigh := marketCap * (1.0 + float64((hashBase+timeHash)%15)/100.0)
		marketCapLow := marketCap * (1.0 - float64((hashBase+timeHash)%15)/100.0)
		marketCapClose := marketCap
		
		candles = append(candles, models.MarketCapCandle{
			TokenAddress:   tokenAddress,
			Timestamp:      timestamp,
			Open:           open,
			High:           high,
			Low:            low,
			Close:          close,
			Volume:         volume,
			MarketCapOpen:  marketCapOpen,
			MarketCapHigh:  marketCapHigh,
			MarketCapLow:   marketCapLow,
			MarketCapClose: marketCapClose,
		})
	}
	
	return candles, nil
}

// GetLaunchpadLPProviders retrieves launchpad LP providers
func (a *Adapter) GetLaunchpadLPProviders() (map[string]string, error) {
	if a.useMock {
		return a.getMockLaunchpadLPProviders()
	}
	
	// This would call the GMGN API endpoint: /defi/quotation/v1/launchpad/sol/lp_provider
	return nil, fmt.Errorf("not implemented")
}

// getMockLaunchpadLPProviders returns mock launchpad LP providers data
func (a *Adapter) getMockLaunchpadLPProviders() (map[string]string, error) {
	providers := map[string]string{
		"Pump.fun": "39azUYFWPz3VHgKCf3VChUwbpURdCHRxjWVowf5jUJjg",
		"Moonshot": "CGsqR7CTqTwbmAUTPnfg9Bj9GLJgkrUD9rhjh3vHEYvh",
		"MakeNow": "BVR2swLp4DoGUSAcduvhPeVRpkRXQYcQdbnFKdKTNcDn",
	}
	return providers, nil
} 