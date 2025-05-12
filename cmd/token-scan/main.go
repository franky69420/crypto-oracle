package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/franky69420/crypto-oracle/internal/gateway/gmgn"
	"github.com/franky69420/crypto-oracle/internal/pipeline"
	"github.com/franky69420/crypto-oracle/internal/token"
	"github.com/franky69420/crypto-oracle/pkg/models"
	"github.com/sirupsen/logrus"
)

// Basic in-memory implementation of MemoryOfTrust
type SimpleMemoryOfTrust struct {
	logger *logrus.Logger
}

func (m *SimpleMemoryOfTrust) GetWalletTrustScore(walletAddress string) (float64, error) {
	return 50.0, nil // Default neutral score
}

func (m *SimpleMemoryOfTrust) RecordWalletInteraction(interaction *models.WalletInteraction) error {
	m.logger.WithFields(logrus.Fields{
		"wallet": interaction.WalletAddress,
		"token":  interaction.TokenAddress,
		"type":   interaction.ActionType,
	}).Debug("Recorded wallet interaction")
	return nil
}

func (m *SimpleMemoryOfTrust) GetTokenTrustMetrics(tokenAddress string) (*models.TokenTrustMetrics, error) {
	return &models.TokenTrustMetrics{
		TokenAddress:    tokenAddress,
		ActiveWallets:   100,
		TrustedWallets:  20,
		AvgTrustScore:   60.0,
		EarlyTrustRatio: 0.15,
	}, nil
}

func (m *SimpleMemoryOfTrust) GetWalletTokenHistory(walletAddress, tokenAddress string) ([]models.WalletInteraction, error) {
	return []models.WalletInteraction{}, nil
}

func (m *SimpleMemoryOfTrust) GetSimilarWallets(walletAddress string, minSimilarity float64, limit int) ([]models.WalletSimilarity, error) {
	return []models.WalletSimilarity{}, nil
}

func (m *SimpleMemoryOfTrust) GetMostTrustedWallets(limit int) ([]models.WalletTrustScore, error) {
	return []models.WalletTrustScore{}, nil
}

func (m *SimpleMemoryOfTrust) GetWalletRiskFactors(walletAddress string) (*models.WalletRiskFactors, error) {
	return &models.WalletRiskFactors{}, nil
}

func (m *SimpleMemoryOfTrust) GetTokenInfluencers(tokenAddress string, limit int) ([]models.WalletInfluence, error) {
	return []models.WalletInfluence{}, nil
}

func (m *SimpleMemoryOfTrust) GetWalletTokens(walletAddress string, limit int) ([]models.WalletToken, error) {
	return []models.WalletToken{}, nil
}

func (m *SimpleMemoryOfTrust) UpdateWalletSimilarities() error {
	return nil
}

func (m *SimpleMemoryOfTrust) GetTokenActiveWallets(tokenAddress string, minTrustScore float64, limit int) ([]models.ActiveWallet, error) {
	return []models.ActiveWallet{}, nil
}

func (m *SimpleMemoryOfTrust) GetActiveWalletsCount(tokenAddress string) (int, error) {
	return 100, nil
}

func (m *SimpleMemoryOfTrust) Start(ctx context.Context) error {
	m.logger.Info("Starting simple memory of trust")
	return nil
}

func (m *SimpleMemoryOfTrust) Stop() error {
	m.logger.Info("Stopping simple memory of trust")
	return nil
}

// SimplePipeline implements NotificationPipeline interface
type SimplePipeline struct {
	logger   *logrus.Logger
	handlers []pipeline.NotificationHandler
}

func (p *SimplePipeline) Publish(msg pipeline.NotificationMessage) error {
	p.logger.WithField("message", fmt.Sprintf("%+v", msg)).Info("Published notification")
	return nil
}

func (p *SimplePipeline) AddHandler(handler pipeline.NotificationHandler) {
	p.handlers = append(p.handlers, handler)
	p.logger.WithField("handler", handler.Name()).Info("Added notification handler")
}

func (p *SimplePipeline) RemoveHandler(handlerName string) {
	for i, h := range p.handlers {
		if h.Name() == handlerName {
			p.handlers = append(p.handlers[:i], p.handlers[i+1:]...)
			p.logger.WithField("handler", handlerName).Info("Removed notification handler")
			return
		}
	}
}

func main() {
	// Initialize logger
	logger := logrus.New()
	logger.SetOutput(os.Stdout)
	logger.SetLevel(logrus.DebugLevel)
	logger.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})

	logger.Info("Starting Token Scanner")

	// Initialize components
	clientConfig := gmgn.ClientConfig{
		BaseURL:        "https://gmgn.ai",
		DeviceID:       "411fa5e2-ade9-4058-9fef-90147baf61fe",
		ClientID:       "gmgn_web_2025.0128.214338",
		FromApp:        "web",
		AppVer:         "2025.0128.214338",
		TzName:         "Europe/Paris",
		TzOffset:       "+0100",
		AppLang:        "en",
		RequestTimeout: 60 * time.Second,
		RateLimitDelay: 500 * time.Millisecond,
	}

	// Create GMGN client and adapter with real data
	client := gmgn.NewClientWithLogger(clientConfig, logger)
	adapter := gmgn.NewRealAdapter(client, logger)

	// Create memory of trust and notification pipeline
	memoryOfTrust := &SimpleMemoryOfTrust{logger: logger}
	notificationPipeline := &SimplePipeline{
		logger:   logger,
		handlers: []pipeline.NotificationHandler{},
	}

	// Add a console handler to show notifications
	notificationPipeline.AddHandler(pipeline.NewConsoleHandler(logger))

	// Create token engine
	tokenEngine := token.NewEngine(adapter, memoryOfTrust, notificationPipeline, logger)

	// Set up context with cancellation for clean shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start token engine
	if err := tokenEngine.Start(ctx); err != nil {
		logger.WithError(err).Fatal("Failed to start token engine")
	}

	// Wait for termination signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	<-sigChan
	logger.Info("Shutting down...")

	// Clean shutdown
	if err := tokenEngine.Stop(); err != nil {
		logger.WithError(err).Error("Error shutting down token engine")
	}

	logger.Info("Shutdown complete")
}
