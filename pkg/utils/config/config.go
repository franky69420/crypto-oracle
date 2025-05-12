package config

import (
	"fmt"
	"os"

	"github.com/spf13/viper"
)

// Config est la structure principale de configuration de l'application
type Config struct {
	LogLevel  string         `mapstructure:"log_level"`
	API       *APIConfig     `mapstructure:"api"`
	Database  *DatabaseConfig `mapstructure:"database"`
	Redis     *RedisConfig    `mapstructure:"redis"`
	GMGN      *GMGNConfig     `mapstructure:"gmgn"`
}

// APIConfig contient la configuration du serveur API
type APIConfig struct {
	Host           string `mapstructure:"host"`
	Port           int    `mapstructure:"port"`
	ReadTimeout    int    `mapstructure:"read_timeout"`
	WriteTimeout   int    `mapstructure:"write_timeout"`
	MaxHeaderBytes int    `mapstructure:"max_header_bytes"`
}

// DatabaseConfig contient la configuration de la base de données
type DatabaseConfig struct {
	Host               string `mapstructure:"host"`
	Port               int    `mapstructure:"port"`
	User               string `mapstructure:"user"`
	Password           string `mapstructure:"password"`
	Database           string `mapstructure:"database"`
	Name               string `mapstructure:"name"`
	SSLMode            string `mapstructure:"ssl_mode"`
	MaxConnections     int    `mapstructure:"max_connections"`
	MinConnections     int    `mapstructure:"min_connections"`
	MaxConnLifetime    int    `mapstructure:"max_conn_lifetime"`
	MaxConnIdleTime    int    `mapstructure:"max_conn_idle_time"`
	HealthCheckPeriod  int    `mapstructure:"health_check_period"`
}

// RedisConfig contient la configuration de Redis
type RedisConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
	PoolSize int    `mapstructure:"pool_size"`
}

// GMGNConfig contient la configuration pour l'API GMGN
type GMGNConfig struct {
	BaseURL        string `mapstructure:"base_url"`
	DeviceID       string `mapstructure:"device_id"`
	ClientID       string `mapstructure:"client_id"`
	AppVer         string `mapstructure:"app_ver"`
	TzName         string `mapstructure:"tz_name"`
	TzOffset       string `mapstructure:"tz_offset"`
	AppLang        string `mapstructure:"app_lang"`
	FromApp        string `mapstructure:"from_app"`
	RequestTimeout int    `mapstructure:"request_timeout"`
	RateLimitDelay int    `mapstructure:"rate_limit_delay"`
}

// Load charge la configuration à partir d'un fichier
func Load() (*Config, error) {
	// Régler les valeurs par défaut
	setDefaults()

	// Déterminer l'environnement
	env := os.Getenv("APP_ENV")
	if env == "" {
		env = "development"
	}

	// Configurer Viper
	viper.SetConfigName("config")         // nom du fichier de configuration
	viper.SetConfigType("yaml")           // format du fichier de configuration
	viper.AddConfigPath(".")              // chercher dans le répertoire courant
	viper.AddConfigPath("./config")       // chercher dans ./config
	viper.AddConfigPath("../config")      // chercher dans ../config
	viper.AddConfigPath("/etc/crypto-oracle") // chercher dans /etc/crypto-oracle

	// Permettre la surcharge par les variables d'environnement
	viper.AutomaticEnv()

	// Lire la configuration
	if err := viper.ReadInConfig(); err != nil {
		// Si le fichier de configuration n'existe pas, c'est OK, on utilise les valeurs par défaut
		// et les variables d'environnement
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("erreur lors de la lecture du fichier de configuration: %w", err)
		}
	}

	// Charger la configuration spécifique à l'environnement
	envConfigFile := fmt.Sprintf("config.%s", env)
	viper.SetConfigName(envConfigFile)
	if err := viper.MergeInConfig(); err != nil {
		// Ignorer si le fichier spécifique à l'environnement n'existe pas
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("erreur lors de la lecture du fichier de configuration d'environnement: %w", err)
		}
	}

	// Charger la configuration dans la structure
	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("impossible de décoder la configuration: %w", err)
	}

	return &config, nil
}

// setDefaults définit les valeurs par défaut pour la configuration
func setDefaults() {
	// Valeurs par défaut générales
	viper.SetDefault("log_level", "info")

	// Valeurs par défaut pour l'API
	viper.SetDefault("api.host", "0.0.0.0")
	viper.SetDefault("api.port", 8080)
	viper.SetDefault("api.read_timeout", 30)
	viper.SetDefault("api.write_timeout", 30)
	viper.SetDefault("api.max_header_bytes", 1048576) // 1MB

	// Valeurs par défaut pour la base de données
	viper.SetDefault("database.host", "localhost")
	viper.SetDefault("database.port", 5432)
	viper.SetDefault("database.user", "postgres")
	viper.SetDefault("database.password", "postgres")
	viper.SetDefault("database.database", "crypto_oracle")
	viper.SetDefault("database.name", "crypto_oracle")
	viper.SetDefault("database.ssl_mode", "disable")
	viper.SetDefault("database.max_connections", 20)
	viper.SetDefault("database.min_connections", 5)
	viper.SetDefault("database.max_conn_lifetime", 3600)
	viper.SetDefault("database.max_conn_idle_time", 1800)
	viper.SetDefault("database.health_check_period", 60)

	// Valeurs par défaut pour Redis
	viper.SetDefault("redis.host", "localhost")
	viper.SetDefault("redis.port", 6379)
	viper.SetDefault("redis.password", "")
	viper.SetDefault("redis.db", 0)
	viper.SetDefault("redis.pool_size", 10)

	// Valeurs par défaut pour GMGN
	viper.SetDefault("gmgn.base_url", "https://gmgn.ai")
	viper.SetDefault("gmgn.device_id", "411fa5e2-ade9-4058-9fef-90147baf61fe")
	viper.SetDefault("gmgn.client_id", "gmgn_web_2025.0128.214338")
	viper.SetDefault("gmgn.app_ver", "2025.0128.214338")
	viper.SetDefault("gmgn.tz_name", "Africa/Casablanca")
	viper.SetDefault("gmgn.tz_offset", "3600")
	viper.SetDefault("gmgn.app_lang", "en")
	viper.SetDefault("gmgn.from_app", "gmgn")
	viper.SetDefault("gmgn.request_timeout", 30)
	viper.SetDefault("gmgn.rate_limit_delay", 300) // 300ms entre les requêtes
} 