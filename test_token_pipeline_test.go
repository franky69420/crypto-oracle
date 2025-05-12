package main

import (
	"context"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/franky69420/crypto-oracle/internal/pipeline"
	"github.com/franky69420/crypto-oracle/internal/storage/cache"
	"github.com/franky69420/crypto-oracle/internal/token"
	"github.com/franky69420/crypto-oracle/pkg/models"
	"github.com/franky69420/crypto-oracle/pkg/utils/config"
)

// TokenEngineClient définit l'interface minimale pour les tests
type TokenEngineClient interface {
	GetTokenInfo(tokenAddress string) (*models.Token, error)
	GetTokenStats(tokenAddress string) (*models.TokenStats, error)
	GetTokenTrades(tokenAddress string, limit int) ([]models.TokenTrade, error)
	GetTokenPrice(tokenAddress string) (*models.TokenPrice, error)
	GetWalletTokenTrades(walletAddress, tokenAddress string, limit int) ([]models.TokenTrade, error)
}

// TestGMGNClient implémente l'interface TokenEngineClient pour les tests
type TestGMGNClient struct {
	Logger *logrus.Logger
}

// GetTokenInfo implémente la méthode de l'interface TokenEngineClient
func (m *TestGMGNClient) GetTokenInfo(tokenAddress string) (*models.Token, error) {
	return &models.Token{
		Address:     tokenAddress,
		Symbol:      "TEST",
		Name:        "Test Token",
		TotalSupply: 1000000,
		HolderCount: 100,
	}, nil
}

// GetTokenStats implémente la méthode de l'interface TokenEngineClient
func (m *TestGMGNClient) GetTokenStats(tokenAddress string) (*models.TokenStats, error) {
	return &models.TokenStats{
		HolderCount:       100,
		Volume1h:          1000,
		Volume24h:         24000,
		Price:             1.5,
		MarketCap:         1500000,
		PriceChange1h:     2.5,
		BuyCount1h:        50,
		SellCount1h:       20,
		LiquidityUSD:      800000,
		PoolAddress:       "0xpool123456789",
		PoolTradesLast24h: 150,
	}, nil
}

// GetTokenTrades implémente la méthode de l'interface TokenEngineClient
func (m *TestGMGNClient) GetTokenTrades(tokenAddress string, limit int) ([]models.TokenTrade, error) {
	return []models.TokenTrade{}, nil
}

// GetTokenPrice implémente la méthode de l'interface TokenEngineClient
func (m *TestGMGNClient) GetTokenPrice(tokenAddress string) (*models.TokenPrice, error) {
	return &models.TokenPrice{
		Price:     1.5,
		Change24h: 2.5,
	}, nil
}

// GetWalletTokenTrades implémente la méthode de l'interface TokenEngineClient
func (m *TestGMGNClient) GetWalletTokenTrades(walletAddress, tokenAddress string, limit int) ([]models.TokenTrade, error) {
	return []models.TokenTrade{}, nil
}

// MockMemoryOfTrust implémente l'interface memory.MemoryOfTrust pour les tests
type MockMemoryOfTrust struct {
	Logger *logrus.Logger
}

func (m *MockMemoryOfTrust) Start(ctx context.Context) error {
	return nil
}

func (m *MockMemoryOfTrust) Stop() error {
	return nil
}

func (m *MockMemoryOfTrust) GetWalletTrustScore(walletAddress string) (float64, error) {
	return 75.0, nil
}

func (m *MockMemoryOfTrust) RecordWalletInteraction(interaction *models.WalletInteraction) error {
	return nil
}

func (m *MockMemoryOfTrust) GetTokenTrustMetrics(tokenAddress string) (*models.TokenTrustMetrics, error) {
	return &models.TokenTrustMetrics{
		TokenAddress:          tokenAddress,
		ActiveWallets:         100,
		TrustedWallets:        30,
		AvgTrustScore:         65.0,
		TrustScoreDistribution: map[string]int{
			"high":   25,
			"medium": 45,
			"low":    30,
		},
		EarlyTrustRatio:       0.35,
		SmartMoneyCount:       15,
		SmartMoneyRatio:       0.15,
		SmartMoneyActivity:    35.0,
		EarlyTrustedRatio:     0.4,
	}, nil
}

func (m *MockMemoryOfTrust) GetWalletTokenHistory(walletAddress, tokenAddress string) ([]models.WalletInteraction, error) {
	return []models.WalletInteraction{}, nil
}

func (m *MockMemoryOfTrust) GetSimilarWallets(walletAddress string, minSimilarity float64, limit int) ([]models.WalletSimilarity, error) {
	return nil, nil
}

func (m *MockMemoryOfTrust) GetMostTrustedWallets(limit int) ([]models.WalletTrustScore, error) {
	return nil, nil
}

func (m *MockMemoryOfTrust) GetWalletRiskFactors(walletAddress string) (*models.WalletRiskFactors, error) {
	return nil, nil
}

func (m *MockMemoryOfTrust) GetTokenInfluencers(tokenAddress string, limit int) ([]models.WalletInfluence, error) {
	return nil, nil
}

func (m *MockMemoryOfTrust) GetWalletTokens(walletAddress string, limit int) ([]models.WalletToken, error) {
	return nil, nil
}

func (m *MockMemoryOfTrust) UpdateWalletSimilarities() error {
	return nil
}

func (m *MockMemoryOfTrust) GetTokenActiveWallets(tokenAddress string, minTrustScore float64, limit int) ([]models.ActiveWallet, error) {
	return nil, nil
}

func (m *MockMemoryOfTrust) GetActiveWalletsCount(tokenAddress string) (int, error) {
	return 100, nil
}

// TestTokenPipelineIntegration tests the integration between token engine and pipeline
func TestTokenPipelineIntegration(t *testing.T) {
	// Setup logger
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)
	
	// Create Redis client for testing
	redisConfig := &config.RedisConfig{
		Host: "localhost",
		Port: 6379,
	}
	redisClient, err := cache.NewRedisConnection(redisConfig, logger)
	if err != nil {
		t.Skipf("Skipping test due to Redis connection error: %v", err)
		return
	}
	defer redisClient.Close()
	
	// Create pipeline
	pipelineSvc := pipeline.NewPipeline(redisClient, logger)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	
	err = pipelineSvc.Start(ctx)
	if err != nil {
		t.Fatalf("Failed to start pipeline: %v", err)
	}
	
	// Create mock implementations
	mockGMGN := &TestGMGNClient{Logger: logger}
	mockMemory := &MockMemoryOfTrust{Logger: logger}
	
	// Create token engine
	tokenEngine := token.NewEngine(mockGMGN, mockMemory, pipelineSvc, logger)
	
	// Test token state update
	err = tokenEngine.UpdateTokenState("0xTestToken123", "trending")
	if err != nil {
		t.Errorf("Failed to update token state: %v", err)
	}
	
	// Test reactivation metrics
	candidate := models.ReactivationCandidate{
		TokenAddress:      "0xTestToken123",
		TokenSymbol:       "TEST",
		ReactivationScore: 75.0,
		Changes:           map[string]float64{"volume_1h_change": 3.5, "price_change": 0.15},
		DetectedAt:        time.Now(),
	}
	
	err = tokenEngine.SaveReactivationMetrics(candidate)
	if err != nil {
		t.Errorf("Failed to save reactivation metrics: %v", err)
	}
	
	// Allow time for events to be published to Redis streams
	time.Sleep(500 * time.Millisecond)
	
	// Test successful
	t.Log("Token engine and pipeline integration test successful")
} 