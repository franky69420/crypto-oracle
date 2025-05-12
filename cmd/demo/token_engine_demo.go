package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/franky69420/crypto-oracle/internal/pipeline"
	"github.com/franky69420/crypto-oracle/internal/storage/cache"
	"github.com/franky69420/crypto-oracle/internal/token"
	"github.com/franky69420/crypto-oracle/pkg/models"
	"github.com/franky69420/crypto-oracle/pkg/utils/config"
)

// MockGMGNClient is a simple implementation for demo
type MockGMGNClient struct {
	Logger *logrus.Logger
}

func (m *MockGMGNClient) GetTokenInfo(tokenAddress string) (*models.Token, error) {
	return &models.Token{
		Address:     tokenAddress,
		Symbol:      "TEST",
		Name:        "Test Token",
		TotalSupply: 1000000,
		HolderCount: 100,
	}, nil
}

func (m *MockGMGNClient) GetTokenStats(tokenAddress string) (*models.TokenStats, error) {
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

func (m *MockGMGNClient) GetTokenTrades(tokenAddress string, limit int) ([]models.TokenTrade, error) {
	return []models.TokenTrade{}, nil
}

func (m *MockGMGNClient) GetTokenPrice(tokenAddress string) (*models.TokenPrice, error) {
	return &models.TokenPrice{
		Price:     1.5,
		Change24h: 2.5,
	}, nil
}

func (m *MockGMGNClient) GetWalletTokenTrades(walletAddress, tokenAddress string, limit int) ([]models.TokenTrade, error) {
	return []models.TokenTrade{}, nil
}

// MockMemoryOfTrust is a simple implementation for demo
type MockMemoryOfTrust struct {
	Logger *logrus.Logger
}

func (m *MockMemoryOfTrust) Start(ctx context.Context) error {
	m.Logger.Info("MockMemoryOfTrust started")
	return nil
}

func (m *MockMemoryOfTrust) Stop() error {
	m.Logger.Info("MockMemoryOfTrust stopped")
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

// PipelineProcessor demonstrates handling token events
type PipelineProcessor struct {
	name   string
	logger *logrus.Logger
}

func NewPipelineProcessor(logger *logrus.Logger) *PipelineProcessor {
	return &PipelineProcessor{
		name:   "demo_processor",
		logger: logger,
	}
}

func (p *PipelineProcessor) Process(message pipeline.Message) error {
	p.logger.WithFields(logrus.Fields{
		"id":        message.ID,
		"type":      message.Type,
		"timestamp": message.Timestamp,
		"payload":   message.Payload,
	}).Info("Processed token event")
	return nil
}

func (p *PipelineProcessor) GetName() string {
	return p.name
}

func main() {
	// Initialize logger
	logger := logrus.New()
	logger.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})
	logger.SetLevel(logrus.InfoLevel)

	logger.Info("Starting Token Engine Demo")

	// Create Redis connection
	redisConfig := &config.RedisConfig{
		Host: "localhost",
		Port: 6379,
	}
	redisClient, err := cache.NewRedisConnection(redisConfig, logger)
	if err != nil {
		logger.WithError(err).Fatal("Failed to connect to Redis")
	}
	defer redisClient.Close()

	// Create and start pipeline
	pipelineSvc := pipeline.NewPipeline(redisClient, logger)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err = pipelineSvc.Start(ctx)
	if err != nil {
		logger.WithError(err).Fatal("Failed to start pipeline")
	}

	// Register pipeline processor
	processor := NewPipelineProcessor(logger)
	pipelineSvc.RegisterProcessor(processor)

	// Create mocks
	mockGMGN := &MockGMGNClient{Logger: logger}
	mockMemory := &MockMemoryOfTrust{Logger: logger}

	// Create token engine
	tokenEngine := token.NewEngine(mockGMGN, mockMemory, pipelineSvc, logger)
	tokenEngine.Start(ctx)

	// Start price monitoring in background
	go tokenEngine.MonitorPriceMovements(ctx, 5*time.Second)

	// Demo token operations
	tokenAddress := "0xDemoToken123456789"
	
	logger.Info("Updating token state to 'trending'")
	if err := tokenEngine.UpdateTokenState(tokenAddress, "trending"); err != nil {
		logger.WithError(err).Error("Failed to update token state")
	}

	time.Sleep(1 * time.Second)

	logger.Info("Saving token reactivation metrics")
	reactivation := models.ReactivationCandidate{
		TokenAddress:      tokenAddress,
		TokenSymbol:       "DEMO",
		ReactivationScore: 82.5,
		Changes:           map[string]float64{"volume_1h_change": 5.2, "price_change": 1.8},
		DetectedAt:        time.Now(),
	}
	
	if err := tokenEngine.SaveReactivationMetrics(reactivation); err != nil {
		logger.WithError(err).Error("Failed to save reactivation metrics")
	}

	// Wait for interrupt signal
	logger.Info("Demo is running. Press Ctrl+C to stop.")
	
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh
	
	logger.Info("Shutting down demo")
	pipelineSvc.Shutdown(ctx)
	
	logger.Info("Demo completed")
} 