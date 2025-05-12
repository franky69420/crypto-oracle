package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	
	"github.com/franky69420/crypto-oracle/internal/api"
	"github.com/franky69420/crypto-oracle/internal/gateway/gmgn"
	"github.com/franky69420/crypto-oracle/internal/memory"
	"github.com/franky69420/crypto-oracle/internal/pipeline"
	"github.com/franky69420/crypto-oracle/internal/storage/cache"
	"github.com/franky69420/crypto-oracle/internal/storage/db"
	"github.com/franky69420/crypto-oracle/internal/token"
)

func main() {
	// Initialiser le logger
	logger := logrus.New()
	logger.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})
	
	// Définir le niveau de log
	logLevel := os.Getenv("LOG_LEVEL")
	if logLevel == "" {
		logLevel = "info"
	}
	level, err := logrus.ParseLevel(logLevel)
	if err != nil {
		logger.WithError(err).Warn("Invalid log level, using INFO")
		level = logrus.InfoLevel
	}
	logger.SetLevel(level)
	
	// Charger la configuration
	if err := loadConfig(); err != nil {
		logger.WithError(err).Fatal("Failed to load configuration")
	}
	
	// Créer le contexte principal avec annulation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	
	// Initialiser les composants
	redisClient, err := initRedis(logger)
	if err != nil {
		logger.WithError(err).Fatal("Failed to initialize Redis")
	}
	defer redisClient.Close()
	
	dbPool, err := initDatabase(logger)
	if err != nil {
		logger.WithError(err).Fatal("Failed to initialize database")
	}
	defer dbPool.Close()
	
	// Initialiser le pipeline
	pipelineSvc := pipeline.NewPipeline(redisClient, logger)
	if err := pipelineSvc.Start(ctx); err != nil {
		logger.WithError(err).Fatal("Failed to start pipeline service")
	}
	
	// Initialiser le client GMGN
	gmgnClient, err := initGMGNClient(logger)
	if err != nil {
		logger.WithError(err).Fatal("Failed to initialize GMGN client")
	}
	
	// Initialiser Memory of Trust
	memoryTrust, err := initMemoryOfTrust(ctx, redisClient, dbPool, logger)
	if err != nil {
		logger.WithError(err).Fatal("Failed to initialize Memory of Trust")
	}
	
	// Initialiser Token Engine
	tokenEngine, err := initTokenEngine(gmgnClient, memoryTrust, pipelineSvc, logger)
	if err != nil {
		logger.WithError(err).Fatal("Failed to initialize Token Engine")
	}
	
	// Initialiser l'API
	apiServer, err := initAPI(dbPool, redisClient, tokenEngine, memoryTrust, logger)
	if err != nil {
		logger.WithError(err).Fatal("Failed to initialize API server")
	}
	
	// Démarrer l'API en arrière-plan
	go func() {
		if err := apiServer.Start(); err != nil {
			logger.WithError(err).Fatal("API server error")
		}
	}()
	
	// Attendre les signaux d'interruption
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	
	<-sigCh
	logger.Info("Shutdown signal received")
	
	// Arrêter les services
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()
	
	if err := apiServer.Shutdown(shutdownCtx); err != nil {
		logger.WithError(err).Error("Failed to shutdown API server gracefully")
	}
	
	if err := pipelineSvc.Shutdown(shutdownCtx); err != nil {
		logger.WithError(err).Error("Failed to shutdown Pipeline service gracefully")
	}
	
	if err := memoryTrust.Stop(); err != nil {
		logger.WithError(err).Error("Failed to shutdown Memory of Trust gracefully")
	}
	
	logger.Info("Application shutdown complete")
}

// loadConfig charge la configuration depuis les fichiers et variables d'environnement
func loadConfig() error {
	viper.SetConfigName("default")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./config")
	
	// Charger le fichier de configuration par défaut
	if err := viper.ReadInConfig(); err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}
	
	// Charger le fichier de configuration spécifique à l'environnement
	env := os.Getenv("APP_ENV")
	if env == "" {
		env = "dev" // Environnement par défaut
	}
	
	viper.SetConfigName(env)
	if err := viper.MergeInConfig(); err != nil {
		// Ignorer l'erreur si le fichier n'existe pas
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return fmt.Errorf("failed to merge environment config: %w", err)
		}
	}
	
	// Lier les variables d'environnement
	viper.AutomaticEnv()
	
	return nil
}

// initRedis initialise la connexion Redis
func initRedis(logger *logrus.Logger) (*cache.Redis, error) {
	redisConfig := &config.RedisConfig{
		Host:     viper.GetString("redis.host"),
		Port:     viper.GetInt("redis.port"),
		Password: viper.GetString("redis.password"),
		DB:       viper.GetInt("redis.db"),
		PoolSize: viper.GetInt("redis.pool_size"),
	}
	
	return cache.NewRedisConnection(redisConfig, logger)
}

// initDatabase initialise la connexion à la base de données
func initDatabase(logger *logrus.Logger) (*db.Pool, error) {
	dbConfig := &config.DatabaseConfig{
		Host:     viper.GetString("database.host"),
		Port:     viper.GetInt("database.port"),
		User:     viper.GetString("database.user"),
		Password: viper.GetString("database.password"),
		DBName:   viper.GetString("database.dbname"),
		SSLMode:  viper.GetString("database.sslmode"),
		PoolSize: viper.GetInt("database.pool_size"),
	}
	
	return db.NewDatabasePool(dbConfig, logger)
}

// initGMGNClient initialise le client GMGN
func initGMGNClient(logger *logrus.Logger) (gmgn.Client, error) {
	gmgnConfig := gmgn.ClientConfig{
		BaseURL:        viper.GetString("gmgn.base_url"),
		DeviceID:       viper.GetString("gmgn.device_id"),
		ClientID:       viper.GetString("gmgn.client_id"),
		FromApp:        viper.GetString("gmgn.from_app"),
		AppVer:         viper.GetString("gmgn.app_ver"),
		TzName:         viper.GetString("gmgn.tz_name"),
		TzOffset:       viper.GetString("gmgn.tz_offset"),
		AppLang:        viper.GetString("gmgn.app_lang"),
		RequestTimeout: viper.GetDuration("gmgn.request_timeout"),
		RateLimitDelay: viper.GetDuration("gmgn.rate_limit_delay"),
	}
	
	return gmgn.NewClient(gmgnConfig), nil
}

// initMemoryOfTrust initialise le système Memory of Trust
func initMemoryOfTrust(ctx context.Context, redisClient *cache.Redis, dbPool *db.Pool, logger *logrus.Logger) (memory.MemoryOfTrust, error) {
	memoryConfig := &memory.Config{
		UpdateInterval:      viper.GetDuration("memory.update_interval"),
		TrustScoreThreshold: viper.GetFloat64("memory.trust_score_threshold"),
		CacheTTL:            viper.GetDuration("memory.cache_ttl"),
	}
	
	memoryTrust := memory.NewTrustNetwork(dbPool, redisClient, memoryConfig, logger)
	if err := memoryTrust.Start(ctx); err != nil {
		return nil, fmt.Errorf("failed to start Memory of Trust: %w", err)
	}
	
	return memoryTrust, nil
}

// initTokenEngine initialise et démarre le moteur de token
func initTokenEngine(gmgnClient gmgn.Client, memoryTrust memory.MemoryOfTrust, pipelineSvc *pipeline.Pipeline, logger *logrus.Logger) (*token.Engine, error) {
	logger.Info("Initializing Token Engine")
	
	// Charger la configuration
	tokenEngineConfig := viper.Sub("token_engine")
	if tokenEngineConfig == nil {
		return nil, fmt.Errorf("token_engine configuration not found")
	}
	
	// Créer le moteur de token
	tokenEngine := token.NewEngine(gmgnClient, memoryTrust, pipelineSvc, logger)
	
	// Démarrer le moteur
	ctx := context.Background()
	if err := tokenEngine.Start(ctx); err != nil {
		return nil, fmt.Errorf("failed to start token engine: %w", err)
	}
	
	// Démarrer la surveillance des prix si activée
	if tokenEngineConfig.GetBool("enable_price_monitoring") {
		monitoringInterval := tokenEngineConfig.GetDuration("price_monitoring_interval")
		if monitoringInterval == 0 {
			monitoringInterval = 60 * time.Second // Valeur par défaut: 60 secondes
		}
		
		go tokenEngine.MonitorPriceMovements(ctx, monitoringInterval)
		logger.WithField("interval", monitoringInterval).Info("Started price monitoring")
	}
	
	return tokenEngine, nil
}

// initAPI initialise le serveur API
func initAPI(dbPool *db.Pool, redisClient *cache.Redis, tokenEngine *token.Engine, memoryTrust memory.MemoryOfTrust, logger *logrus.Logger) (*api.Server, error) {
	apiConfig := &api.Config{
		Host:            viper.GetString("api.host"),
		Port:            viper.GetInt("api.port"),
		AllowedOrigins:  viper.GetStringSlice("api.allowed_origins"),
		ReadTimeout:     viper.GetDuration("api.read_timeout"),
		WriteTimeout:    viper.GetDuration("api.write_timeout"),
		EnableSwagger:   viper.GetBool("api.enable_swagger"),
		EnableMetrics:   viper.GetBool("api.enable_metrics"),
		EnableRateLimit: viper.GetBool("api.enable_rate_limit"),
	}
	
	return api.NewServer(apiConfig, dbPool, redisClient, tokenEngine, memoryTrust, logger)
} 