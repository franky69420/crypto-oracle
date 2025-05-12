package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/franko/crypto-oracle/internal/alerting"
	"github.com/franko/crypto-oracle/internal/pipeline"
	"github.com/franko/crypto-oracle/internal/storage/cache"
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

	logger.Info("Starting Crypto Oracle Core Services Test")

	// Créer une configuration Redis temporaire
	redisConfig := &config.RedisConfig{
		Host:     "localhost",
		Port:     6379,
		Password: "",
		DB:       0,
		PoolSize: 10,
	}

	// Créer un client Redis temporaire
	redisClient, err := cache.NewRedisConnection(redisConfig, logger)
	if err != nil {
		logger.WithError(err).Fatal("Failed to connect to Redis")
	}
	defer redisClient.Close()

	// Créer les composants
	alertManager := alerting.NewManager(logger)
	pipelineService := pipeline.NewPipeline(redisClient, logger)

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

	// Créer un exemple d'alerte
	alertManager.CreateAlert(
		"9xDUcv5rTQQkn9eWYnhJ2YHvy91aeE9w5FvPsHxTt6BB",
		"SOLMEME",
		"HIGH_SCORE",
		"URGENT",
		"Test alert for SOLMEME with high X-Score",
	)

	// Tester le pipeline en ajoutant un message
	msg := map[string]interface{}{
		"token_address": "9xDUcv5rTQQkn9eWYnhJ2YHvy91aeE9w5FvPsHxTt6BB",
		"token_symbol":  "SOLMEME",
		"timestamp":     time.Now().Unix(),
		"event_type":    "price_change",
		"data": map[string]interface{}{
			"price_change": 25.5,
			"volume":       1500000.0,
		},
	}

	err = pipelineService.GetRedisClient().XAdd("token_events", msg)
	if err != nil {
		logger.WithError(err).Error("Failed to add event to pipeline")
	} else {
		logger.Info("Successfully added event to pipeline")
	}

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
	err = pipelineService.Shutdown(ctx)
	if err != nil {
		logger.WithError(err).Error("Failed to shutdown Pipeline")
	}

	err = alertManager.Shutdown(ctx)
	if err != nil {
		logger.WithError(err).Error("Failed to shutdown Alert Manager")
	}

	logger.Info("Crypto Oracle core services test shutdown complete")
} 