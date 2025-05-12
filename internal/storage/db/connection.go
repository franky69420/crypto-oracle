package db

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/franky69420/crypto-oracle/pkg/utils/config"
	"github.com/franky69420/crypto-oracle/pkg/utils/logger"
)

// Connection représente une connexion à la base de données
type Connection struct {
	pool   *pgxpool.Pool
	logger *logger.Logger
	config *config.DatabaseConfig
}

// NewConnection crée une nouvelle connexion à la base de données
func NewConnection(cfg *config.DatabaseConfig, logger *logger.Logger) (*Connection, error) {
	connStr := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
		cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.Name, cfg.SSLMode)

	poolConfig, err := pgxpool.ParseConfig(connStr)
	if err != nil {
		return nil, fmt.Errorf("erreur lors de l'analyse de la configuration de la pool: %w", err)
	}

	// Configurer la pool de connexions
	poolConfig.MaxConns = int32(cfg.MaxConnections)
	poolConfig.MinConns = int32(cfg.MinConnections)
	poolConfig.MaxConnLifetime = time.Duration(cfg.MaxConnLifetime) * time.Second
	poolConfig.MaxConnIdleTime = time.Duration(cfg.MaxConnIdleTime) * time.Second
	poolConfig.HealthCheckPeriod = time.Duration(cfg.HealthCheckPeriod) * time.Second

	// Créer la pool de connexions
	pool, err := pgxpool.NewWithConfig(context.Background(), poolConfig)
	if err != nil {
		return nil, fmt.Errorf("erreur lors de la création de la pool de connexions: %w", err)
	}

	// Tester la connexion
	if err := pool.Ping(context.Background()); err != nil {
		return nil, fmt.Errorf("erreur lors du ping de la base de données: %w", err)
	}

	logger.Info("Connexion à la base de données établie avec succès")

	return &Connection{
		pool:   pool,
		logger: logger,
		config: cfg,
	}, nil
}

// Close ferme la connexion à la base de données
func (c *Connection) Close() {
	c.logger.Info("Fermeture de la connexion à la base de données")
	c.pool.Close()
}

// GetPool retourne la pool de connexions
func (c *Connection) GetPool() *pgxpool.Pool {
	return c.pool
}

// Begin démarre une nouvelle transaction
func (c *Connection) Begin(ctx context.Context) (pgx.Tx, error) {
	return c.pool.Begin(ctx)
}

// Exec exécute une requête SQL sans retour de résultats
func (c *Connection) Exec(ctx context.Context, sql string, args ...interface{}) (pgconn.CommandTag, error) {
	return c.pool.Exec(ctx, sql, args...)
}

// Query exécute une requête SQL et retourne les résultats
func (c *Connection) Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error) {
	return c.pool.Query(ctx, sql, args...)
}

// QueryRow exécute une requête SQL et retourne une seule ligne de résultats
func (c *Connection) QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row {
	return c.pool.QueryRow(ctx, sql, args...)
}

// OptimizeIndexes optimise les index de la base de données
func (c *Connection) OptimizeIndexes() error {
	ctx := context.Background()
	c.logger.Info("Optimisation des index de la base de données")
	
	// Analyser les tables pour mettre à jour les statistiques
	_, err := c.pool.Exec(ctx, "ANALYZE")
	if err != nil {
		return fmt.Errorf("erreur lors de l'optimisation des index: %w", err)
	}
	
	return nil
}

// SaveWalletSimilarity enregistre la similarité entre wallets
func (c *Connection) SaveWalletSimilarity(walletAddress string, similarWallet string, similarityScore float64) error {
	ctx := context.Background()
	
	query := `
		INSERT INTO wallet_similarities (
			wallet_address, similar_wallet_address, similarity_score, last_calculated
		) VALUES (
			$1, $2, $3, NOW()
		) ON CONFLICT (wallet_address, similar_wallet_address) DO UPDATE SET
			similarity_score = $3,
			last_calculated = NOW()
	`
	
	_, err := c.pool.Exec(ctx, query, walletAddress, similarWallet, similarityScore)
	if err != nil {
		return fmt.Errorf("failed to save wallet similarity: %w", err)
	}
	
	return nil
}

func getDSN() string {
	// Récupérer les variables d'environnement
	host := getEnv("POSTGRES_HOST", "localhost")
	port := getEnv("POSTGRES_PORT", "5432")
	user := getEnv("POSTGRES_USER", "crypto_oracle")
	password := getEnv("POSTGRES_PASSWORD", "crypto_oracle_pass")
	dbname := getEnv("POSTGRES_DB", "crypto_oracle")

	// Construire la chaîne de connexion
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname,
	)
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
} 