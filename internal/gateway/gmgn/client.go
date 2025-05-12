package gmgn

import (
	"time"
)

// Client est l'interface pour interagir avec l'API GMGN
type Client interface {
	GetTokenStat(tokenAddress string) (*TokenStatResponse, error)
	GetTokenTrades(tokenAddress string, limit int, tag string) (*TradeHistoryResponse, error)
	GetTokenPrice(tokenAddress string, timeframe string) (*KlineDataResponse, error)
	GetAllTokenTraders(tokenAddress string) ([]Trader, error)
	GetTokenHolderStat(tokenAddress string) (*TokenHolderStatResponse, error)
	GetTokenWalletTagsStat(tokenAddress string) (*TokenWalletTagsStatResponse, error)
	GetWalletInfo(walletAddress string) (*WalletInfoResponse, error)
	GetAllWalletHoldings(walletAddress string) ([]Holding, error)
	GetWalletStat(walletAddress string, period string) (*WalletStatResponse, error)
	GetTrending(timeframe string, orderBy string, direction string, filters []string) (*TrendingResponse, error)
	GetCompletedCoins(limit string, orderBy string, direction string) (*CompletedTokensResponse, error)
}

// ClientConfig contient la configuration pour le client GMGN
type ClientConfig struct {
	BaseURL        string        // URL de base de l'API
	DeviceID       string        // ID du dispositif
	ClientID       string        // ID du client
	FromApp        string        // Nom de l'application
	AppVer         string        // Version de l'application
	TzName         string        // Nom du fuseau horaire
	TzOffset       string        // Décalage du fuseau horaire
	AppLang        string        // Langue de l'application
	RequestTimeout time.Duration // Timeout en secondes
	RateLimitDelay time.Duration // Délai entre les requêtes en millisecondes
}

// NewClient crée un nouveau client GMGN
func NewClient(config ClientConfig) Client {
	return newClientImpl(config)
}

// Response est la structure de base pour les réponses de l'API
type Response struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
} 