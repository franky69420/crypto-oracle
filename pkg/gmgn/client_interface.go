package gmgn

import (
	"fmt"
	"time"
)

// Client est l'interface pour les interactions avec l'API GMGN
type Client interface {
	// Méthodes pour les tokens
	GetTokenStat(tokenAddress string) (*TokenStatResponse, error)
	GetTokenTrades(tokenAddress string, limit int, tag string) (*TradeHistoryResponse, error)
	GetTokenPrice(tokenAddress string, timeframe string) (*KlineDataResponse, error)
	GetAllTokenTraders(tokenAddress string) ([]Trader, error)
	GetTokenHolderStat(tokenAddress string) (*TokenHolderStatResponse, error)
	GetTokenWalletTagsStat(tokenAddress string) (*TokenWalletTagsStatResponse, error)

	// Méthodes pour les wallets
	GetWalletInfo(walletAddress string) (*WalletInfoResponse, error)
	GetAllWalletHoldings(walletAddress string) ([]Holding, error)
	GetWalletStat(walletAddress string, period string) (*WalletStatResponse, error)

	// Méthodes pour la découverte
	GetTrending(timeframe string, orderBy string, direction string, filters []string) (*TrendingResponse, error)
	GetCompletedCoins(limit string, orderBy string, direction string) (*CompletedTokensResponse, error)
}

// Configuration du client
type ClientConfig struct {
	BaseURL        string
	DeviceID       string
	ClientID       string
	AppVer         string
	TzName         string
	TzOffset       string
	AppLang        string
	FromApp        string
	RequestTimeout time.Duration
	RateLimitDelay time.Duration
}

// DefaultConfig retourne une configuration par défaut pour le client GMGN
func DefaultConfig() ClientConfig {
	return ClientConfig{
		BaseURL:        "https://gmgn.ai",
		DeviceID:       "411fa5e2-ade9-4058-9fef-90147baf61fe",
		ClientID:       "gmgn_web_2025.0128.214338",
		AppVer:         "2025.0128.214338",
		TzName:         "Africa/Casablanca",
		TzOffset:       "3600",
		AppLang:        "en",
		FromApp:        "gmgn",
		RequestTimeout: 30 * time.Second,
		RateLimitDelay: 300 * time.Millisecond, // 300ms entre les requêtes (environ 200 req/min)
	}
}

// NewClient crée une nouvelle instance du client GMGN avec la configuration spécifiée
func NewClient(config ClientConfig) (Client, error) {
	// Valider la configuration
	if config.BaseURL == "" {
		return nil, fmt.Errorf("URL de base requis")
	}

	// Créer et retourner l'implémentation du client
	return newClientImpl(config), nil
} 