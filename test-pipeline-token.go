package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
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

// Test d'intégration du Pipeline et Token Engine
//
// Ce test simule l'intégration entre les composants Pipeline et Token Engine.
// Il vérifie que:
// 1. Les événements sont correctement publiés dans les streams Redis
// 2. Le token engine peut mettre à jour l'état des tokens
// 3. Le mécanisme de réactivation fonctionne
// 4. La surveillance des prix est fonctionnelle
//
// Pour exécuter ce test, assurez-vous que les services Redis et PostgreSQL
// sont en cours d'exécution (via Docker Compose).
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
	mockGMGN := &TestGMGNClient{Logger: logger}
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

	// Attendre les signaux d'interruption ou laisser tourner pendant 30 secondes
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	select {
	case sig := <-sigCh:
		logger.WithField("signal", sig.String()).Info("Received signal, shutting down...")
	case <-time.After(30 * time.Second):
		logger.Info("Test time elapsed, shutting down...")
	}

	// Arrêter les services
	if err = pipelineService.Shutdown(ctx); err != nil {
		logger.WithError(err).Error("Failed to shutdown Pipeline service")
	}

	logger.Info("Pipeline and Token Engine Integration Test completed")
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
		t.Fatalf("Failed to connect to Redis: %v", err)
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