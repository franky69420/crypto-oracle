package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/franko/crypto-oracle/internal/alerting"
	"github.com/franko/crypto-oracle/internal/pipeline"
	"github.com/franko/crypto-oracle/internal/reactivation"
	"github.com/franko/crypto-oracle/internal/storage/cache"
	"github.com/franko/crypto-oracle/pkg/models"
	"github.com/franko/crypto-oracle/pkg/utils/config"
	"github.com/sirupsen/logrus"
)

func main() {
	// Initialiser le logger
	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)
	logger.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})

	logger.Info("Starting Crypto Oracle Memecoin Detector v4.2 (Test)")

	// Créer une configuration Redis temporaire
	redisConfig := &config.RedisConfig{
		Host:     "localhost",
		Port:     6379,
		Password: "",
		DB:       0,
		PoolSize: 10,
	}

	// Créer un client Redis temporaire
	redisClient, err := cache.NewRedisConnection(redisConfig)
	if err != nil {
		logger.WithError(err).Fatal("Failed to connect to Redis")
	}
	defer redisClient.Close()

	// Créer les composants
	alertManager := alerting.NewManager(logger)
	pipelineService := pipeline.NewPipeline(redisClient, logger)
	
	// Initialiser le token engine et le wallet intelligence temporairement
	// Ces services ne font pas partie de notre implémentation actuelle
	tokenEngine := &MockTokenEngine{}
	walletEngine := &MockWalletIntelligence{}

	// Créer le système de réactivation
	reactivationSystem := reactivation.NewSystem(tokenEngine, walletEngine, logger)

	// Créer un contexte avec annulation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Démarrer les services
	err = alertManager.Start(ctx)
	if err != nil {
		logger.WithError(err).Fatal("Failed to start Alert Manager")
	}

	err = pipelineService.Start(ctx)
	if err != nil {
		logger.WithError(err).Fatal("Failed to start Pipeline")
	}

	err = reactivationSystem.Start(ctx)
	if err != nil {
		logger.WithError(err).Fatal("Failed to start Reactivation System")
	}

	// Créer un exemple d'alerte
	alertManager.CreateAlert(
		"9xDUcv5rTQQkn9eWYnhJ2YHvy91aeE9w5FvPsHxTt6BB",
		"SOLMEME",
		"HIGH_SCORE",
		"URGENT",
		"Test alert for SOLMEME with high X-Score",
	)

	// Attendre les signaux d'interruption pour un arrêt propre
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	// Attendre le signal ou laisser le programme tourner pendant 10 secondes
	select {
	case sig := <-sigCh:
		logger.WithField("signal", sig.String()).Info("Received signal, shutting down...")
	case <-time.After(10 * time.Second):
		logger.Info("Test time elapsed, shutting down...")
	}

	// Arrêter les services
	err = reactivationSystem.Shutdown(ctx)
	if err != nil {
		logger.WithError(err).Error("Failed to shutdown Reactivation System")
	}

	err = pipelineService.Shutdown(ctx)
	if err != nil {
		logger.WithError(err).Error("Failed to shutdown Pipeline")
	}

	err = alertManager.Shutdown(ctx)
	if err != nil {
		logger.WithError(err).Error("Failed to shutdown Alert Manager")
	}

	logger.Info("Crypto Oracle test shutdown complete")
}

// Mock structures temporaires pour les tests

type MockTokenEngine struct{}

func (m *MockTokenEngine) GetTokensByStates(states []string) ([]models.Token, error) {
	return []models.Token{}, nil
}

func (m *MockTokenEngine) GetTokenMetrics(tokenAddress string) (*models.TokenMetrics, error) {
	return nil, nil
}

func (m *MockTokenEngine) GetTokenLastSnapshot(tokenAddress string) (*models.TokenMetrics, error) {
	return nil, nil
}

func (m *MockTokenEngine) GetTokenRecentTrades(tokenAddress string, hours int) ([]models.TokenTrade, error) {
	return []models.TokenTrade{}, nil
}

func (m *MockTokenEngine) GetWalletTokenHistory(wallet, token string) ([]models.TokenTrade, error) {
	return []models.TokenTrade{}, nil
}

func (m *MockTokenEngine) UpdateTokenState(tokenAddress, newState string) error {
	return nil
}

func (m *MockTokenEngine) SaveReactivationMetrics(candidate models.ReactivationCandidate) error {
	return nil
}

type MockWalletIntelligence struct{}

func (m *MockWalletIntelligence) IsSmartMoneyWallet(walletAddress string) (bool, float64, error) {
	return false, 0, nil
}

func (m *MockWalletIntelligence) AnalyzeWallet(walletAddress string) (*models.WalletProfile, error) {
	return nil, nil
}

func (m *MockWalletIntelligence) EvaluateTokenHolders(tokenAddress string) (*models.HolderQualityReport, error) {
	return nil, nil
}

func (m *MockWalletIntelligence) Start(ctx context.Context) error {
	return nil
}

func (m *MockWalletIntelligence) Shutdown(ctx context.Context) error {
	return nil
} 