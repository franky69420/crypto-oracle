package startup

import (
	"context"
	"fmt"
	"time"

	"github.com/franky69420/crypto-oracle/internal/api"
	"github.com/franky69420/crypto-oracle/internal/alerting"
	"github.com/franky69420/crypto-oracle/internal/gateway/gmgn"
	"github.com/franky69420/crypto-oracle/internal/memory"
	"github.com/franky69420/crypto-oracle/internal/pipeline"
	"github.com/franky69420/crypto-oracle/internal/reactivation"
	"github.com/franky69420/crypto-oracle/internal/storage/cache"
	"github.com/franky69420/crypto-oracle/internal/storage/db"
	"github.com/franky69420/crypto-oracle/internal/token"
	"github.com/franky69420/crypto-oracle/internal/wallet"
	"github.com/franky69420/crypto-oracle/pkg/utils/config"

	"github.com/sirupsen/logrus"
)

// Application représente l'application complète avec tous ses composants
type Application struct {
	cfg           *config.Config
	logger        *logrus.Logger
	db            *db.Database
	redis         *cache.Redis
	gmgnGateway   *gmgn.Client
	memoryOfTrust *memory.MemoryOfTrust
	tokenEngine   *token.Engine
	walletEngine  *wallet.Intelligence
	reactivation  *reactivation.System
	pipeline      *pipeline.Pipeline
	alertManager  *alerting.Manager
	apiServer     *api.Server
	ctx           context.Context
	cancel        context.CancelFunc
}

// InitializeApplication initialise tous les composants de l'application
func InitializeApplication(cfg *config.Config, logger *logrus.Logger) (*Application, error) {
	// Créer un contexte avec annulation
	ctx, cancel := context.WithCancel(context.Background())

	// Initialiser la base de données
	database, err := db.NewDatabaseConnection(cfg.Database)
	if err != nil {
		return nil, fmt.Errorf("échec de la connexion à la base de données: %w", err)
	}

	// Initialiser Redis
	redisClient, err := cache.NewRedisConnection(cfg.Redis)
	if err != nil {
		database.Close()
		return nil, fmt.Errorf("échec de la connexion à Redis: %w", err)
	}

	// Configuration du client GMGN
	gmgnConfig := gmgn.ClientConfig{
		BaseURL:        "https://gmgn.ai",
		DeviceID:       "web",
		ClientID:       "web",
		FromApp:        "web",
		AppVer:         "1.0.0",
		TzName:         "Europe/Paris",
		TzOffset:       "3600",
		AppLang:        "fr",
		RequestTimeout: 30,
		RateLimitDelay: 2000,
	}

	// Initialiser les composants principaux
	gmgnClient := gmgn.NewClient(gmgnConfig)
	memoryTrust := memory.NewMemoryOfTrust(database, redisClient, logger)
	tokenEng := token.NewEngine(gmgnClient, memoryTrust, logger)
	walletEng := wallet.NewIntelligence(memoryTrust, logger)
	reactivationSys := reactivation.NewSystem(tokenEng, walletEng, logger)
	pipelineSys := pipeline.NewPipeline(redisClient, logger)
	alertMgr := alerting.NewManager(logger)

	// Initialiser le serveur API
	apiSrv := api.NewServer(cfg.API, tokenEng, walletEng, memoryTrust, pipelineSys, alertMgr, logger)

	return &Application{
		cfg:           cfg,
		logger:        logger,
		db:            database,
		redis:         redisClient,
		gmgnGateway:   gmgnClient,
		memoryOfTrust: memoryTrust,
		tokenEngine:   tokenEng,
		walletEngine:  walletEng,
		reactivation:  reactivationSys,
		pipeline:      pipelineSys,
		alertManager:  alertMgr,
		apiServer:     apiSrv,
		ctx:           ctx,
		cancel:        cancel,
	}, nil
}

// Start démarre l'application
func (app *Application) Start() error {
	// Démarrer le pipeline de traitement
	if err := app.pipeline.Start(app.ctx); err != nil {
		return fmt.Errorf("échec du démarrage du pipeline: %w", err)
	}

	// Démarrer le système de réactivation
	if err := app.reactivation.Start(app.ctx); err != nil {
		return fmt.Errorf("échec du démarrage du système de réactivation: %w", err)
	}

	// Démarrer le gestionnaire d'alertes
	if err := app.alertManager.Start(app.ctx); err != nil {
		return fmt.Errorf("échec du démarrage du gestionnaire d'alertes: %w", err)
	}

	// Démarrer le serveur API
	go func() {
		if err := app.apiServer.Start(); err != nil {
			app.logger.Errorf("Erreur du serveur API: %v", err)
			app.cancel()
		}
	}()

	app.logger.Info("Tous les composants ont démarré avec succès")
	return nil
}

// Stop arrête l'application
func (app *Application) Stop() error {
	// Annuler le contexte pour signaler à tous les composants de s'arrêter
	app.cancel()

	// Arrêter les composants dans l'ordre inverse
	if err := app.apiServer.Shutdown(app.ctx); err != nil {
		app.logger.Errorf("Erreur lors de l'arrêt du serveur API: %v", err)
	}

	if err := app.alertManager.Shutdown(app.ctx); err != nil {
		app.logger.Errorf("Erreur lors de l'arrêt du gestionnaire d'alertes: %v", err)
	}

	if err := app.pipeline.Shutdown(app.ctx); err != nil {
		app.logger.Errorf("Erreur lors de l'arrêt du pipeline: %v", err)
	}

	if err := app.reactivation.Shutdown(app.ctx); err != nil {
		app.logger.Errorf("Erreur lors de l'arrêt du système de réactivation: %v", err)
	}

	app.redis.Close()
	app.db.Close()

	return nil
} 