package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/franky69420/crypto-oracle/internal/pipeline"
	"github.com/franky69420/crypto-oracle/internal/storage/cache"
	"github.com/franky69420/crypto-oracle/internal/token"
	"github.com/franky69420/crypto-oracle/pkg/models"
	"github.com/franky69420/crypto-oracle/pkg/utils/config"
	"github.com/sirupsen/logrus"
)

// MockGMGNClient implémente l'interface gmgn.Client pour les tests
type MockGMGNClient struct {
	Logger *logrus.Logger
}

// GetTokenInfo implémente la méthode de l'interface gmgn.Client
func (m *MockGMGNClient) GetTokenInfo(tokenAddress string) (*models.Token, error) {
	return &models.Token{
		Address:    tokenAddress,
		Symbol:     "TEST",
		Name:       "Test Token",
		TotalSupply: 1000000,
		HolderCount: 100,
	}, nil
}

// GetTokenStats implémente la méthode de l'interface gmgn.Client
func (m *MockGMGNClient) GetTokenStats(tokenAddress string) (*models.TokenStats, error) {
	return &models.TokenStats{
		HolderCount: 100,
		Volume1h:    1000,
		Volume24h:   24000,
		Price:       1.5,
		MarketCap:   1500000,
		PriceChange1h: 2.5,
		BuyCount1h:  50,
		SellCount1h: 20,
	}, nil
}

// Implémentation des autres méthodes requises de l'interface gmgn.Client
func (m *MockGMGNClient) GetTokenTrades(tokenAddress string, limit int) ([]models.TokenTrade, error) {
	return []models.TokenTrade{}, nil
}

func (m *MockGMGNClient) GetTokenPrice(tokenAddress string) (*models.TokenPrice, error) {
	return &models.TokenPrice{
		Price: 1.5,
		Change24h: 2.5,
	}, nil
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
		TokenAddress:  tokenAddress,
		ActiveWallets: 100,
		TrustedWallets: 30,
		AvgTrustScore: 65.0,
	}, nil
}

func (m *MockMemoryOfTrust) GetWalletTokenHistory(walletAddress, tokenAddress string) ([]models.WalletInteraction, error) {
	return []models.WalletInteraction{}, nil
}

func (m *MockMemoryOfTrust) GetSimilarWallets(walletAddress string, minSimilarity float64, limit int) ([]models.WalletSimilarity, error) {
	return []models.WalletSimilarity{}, nil
}

func (m *MockMemoryOfTrust) GetMostTrustedWallets(limit int) ([]models.WalletTrustScore, error) {
	return []models.WalletTrustScore{}, nil
}

func (m *MockMemoryOfTrust) GetWalletRiskFactors(walletAddress string) (*models.WalletRiskFactors, error) {
	return nil, nil
}

func (m *MockMemoryOfTrust) GetTokenInfluencers(tokenAddress string, limit int) ([]models.WalletInfluence, error) {
	return []models.WalletInfluence{}, nil
}

func (m *MockMemoryOfTrust) GetWalletTokens(walletAddress string, limit int) ([]models.WalletToken, error) {
	return []models.WalletToken{}, nil
}

func (m *MockMemoryOfTrust) UpdateWalletSimilarities() error {
	return nil
}

func (m *MockMemoryOfTrust) GetTokenActiveWallets(tokenAddress string, minTrustScore float64, limit int) ([]models.ActiveWallet, error) {
	return []models.ActiveWallet{}, nil
}

func (m *MockMemoryOfTrust) GetActiveWalletsCount(tokenAddress string) (int, error) {
	return 100, nil
}

// PipelineTestProcessor est un processeur simple pour afficher les messages du pipeline
type PipelineTestProcessor struct {
	name   string
	logger *logrus.Logger
}

func NewPipelineTestProcessor(logger *logrus.Logger) *PipelineTestProcessor {
	return &PipelineTestProcessor{
		name:   "test_processor",
		logger: logger,
	}
}

func (p *PipelineTestProcessor) Process(message pipeline.Message) error {
	p.logger.WithFields(logrus.Fields{
		"id":        message.ID,
		"type":      message.Type,
		"timestamp": message.Timestamp,
		"payload":   message.Payload,
	}).Info("Received message")
	return nil
}

func (p *PipelineTestProcessor) GetName() string {
	return p.name
}

func main() {
	// Initialiser le logger
	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)
	logger.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})

	logger.Info("Starting Pipeline and Token Engine Integration Test")

	// Créer une configuration Redis
	redisConfig := &config.RedisConfig{
		Host:     "localhost",
		Port:     6379,
		Password: "",
		DB:       0,
		PoolSize: 10,
	}

	// Créer un client Redis
	redisClient, err := cache.NewRedisConnection(redisConfig, logger)
	if err != nil {
		logger.WithError(err).Fatal("Failed to connect to Redis")
	}
	defer redisClient.Close()

	// Créer et démarrer le service pipeline
	pipelineService := pipeline.NewPipeline(redisClient, logger)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err = pipelineService.Start(ctx)
	if err != nil {
		logger.WithError(err).Fatal("Failed to start Pipeline service")
	}

	// Créer et enregistrer le processeur de test
	testProcessor := NewPipelineTestProcessor(logger)
	pipelineService.RegisterProcessor(testProcessor)

	// Créer les mocks
	mockGMGN := &MockGMGNClient{Logger: logger}
	mockMemory := &MockMemoryOfTrust{Logger: logger}

	// Créer le token engine
	tokenEngine := token.NewEngine(mockGMGN, mockMemory, pipelineService, logger)
	tokenEngine.Start(ctx)

	// Test 1: Mise à jour d'état
	logger.Info("Test 1: État du token")
	testTokenAddress := "0x1234567890abcdef1234567890abcdef12345678"
	
	err = tokenEngine.UpdateTokenState(testTokenAddress, "active")
	if err != nil {
		logger.WithError(err).Error("Failed to update token state")
	}

	time.Sleep(1 * time.Second) // Attendre que l'événement soit traité

	// Test 2: Réactivation
	logger.Info("Test 2: Réactivation du token")
	reactivationCandidate := models.ReactivationCandidate{
		TokenAddress:      testTokenAddress,
		TokenSymbol:       "TEST",
		ReactivationScore: 85.5,
		Changes:           map[string]float64{"volume_1h_change": 3.5, "price_change": 0.15},
		DetectedAt:        time.Now(),
	}

	err = tokenEngine.SaveReactivationMetrics(reactivationCandidate)
	if err != nil {
		logger.WithError(err).Error("Failed to save reactivation metrics")
	}

	time.Sleep(1 * time.Second) // Attendre que l'événement soit traité

	// Test 3: Tester la surveillance des prix
	logger.Info("Test 3: Starting price monitoring...")
	go tokenEngine.MonitorPriceMovements(ctx, 2*time.Second)

	// Attendre les signaux d'interruption ou laisser tourner pendant 15 secondes
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	select {
	case sig := <-sigCh:
		logger.WithField("signal", sig.String()).Info("Received signal, shutting down...")
	case <-time.After(15 * time.Second):
		logger.Info("Test time elapsed, shutting down...")
	}

	// Arrêter les services
	if err = pipelineService.Shutdown(ctx); err != nil {
		logger.WithError(err).Error("Failed to shutdown Pipeline service")
	}

	logger.Info("Pipeline and Token Engine Integration Test completed")
} 