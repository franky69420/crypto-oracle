package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/franko/crypto-oracle/pkg/utils/config"
	"github.com/franko/crypto-oracle/pkg/utils/logger"
	"github.com/franko/crypto-oracle/cmd/oracle/startup"
)

func main() {
	// Initialiser la configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Échec du chargement de la configuration: %v", err)
	}

	// Initialiser le logger
	l := logger.NewLogger(cfg.LogLevel)
	l.Info("🔮 Crypto Oracle Memecoin Detector v4.2 démarré")

	// Démarrer les composants du système
	app, err := startup.InitializeApplication(cfg, l)
	if err != nil {
		l.Fatal("Échec de l'initialisation de l'application", err)
	}

	// Démarrer l'application
	if err := app.Start(); err != nil {
		l.Fatal("Échec du démarrage de l'application", err)
	}

	// Attendre l'arrêt gracieux
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	sig := <-sigChan

	l.Info(fmt.Sprintf("Signal d'arrêt reçu: %s", sig.String()))

	// Arrêter l'application
	if err := app.Stop(); err != nil {
		l.Error("Problèmes lors de l'arrêt de l'application", err)
		os.Exit(1)
	}

	l.Info("Application arrêtée avec succès")
} 