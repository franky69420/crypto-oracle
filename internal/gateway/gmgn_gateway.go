package gateway

import (
	"time"

	"github.com/franko/crypto-oracle/pkg/gmgn"
	"github.com/franko/crypto-oracle/pkg/models"
	"github.com/franko/crypto-oracle/pkg/utils/config"
	"github.com/franko/crypto-oracle/pkg/utils/logger"
)

// GMGNGateway est la passerelle vers l'API GMGN
type GMGNGateway struct {
	client gmgn.Client
	logger *logger.Logger
	config *config.GMGNConfig
}

// NewGMGNGateway crée une nouvelle passerelle GMGN
func NewGMGNGateway(config *config.GMGNConfig, logger *logger.Logger) *GMGNGateway {
	// Convertir la configuration en format attendu par le client GMGN
	clientConfig := gmgn.ClientConfig{
		BaseURL:        config.BaseURL,
		DeviceID:       config.DeviceID,
		ClientID:       config.ClientID,
		AppVer:         config.AppVer,
		TzName:         config.TzName,
		TzOffset:       config.TzOffset,
		AppLang:        config.AppLang,
		FromApp:        config.FromApp,
		RequestTimeout: time.Duration(config.RequestTimeout) * time.Second,
		RateLimitDelay: time.Duration(config.RateLimitDelay) * time.Millisecond,
	}

	// Créer le client GMGN
	client, err := gmgn.NewClient(clientConfig)
	if err != nil {
		logger.Error("Failed to create GMGN client", err)
		// Utilisez une implémentation mock ou retournez nil dans un cas réel
		// Ici, on utilise panic pour simplifier
		panic(err)
	}

	return &GMGNGateway{
		client: client,
		logger: logger,
		config: config,
	}
}

// GetTokenInfo récupère les informations d'un token
func (g *GMGNGateway) GetTokenInfo(tokenAddress string) (*models.Token, error) {
	g.logger.Debug("Getting token info", map[string]interface{}{
		"token_address": tokenAddress,
	})

	// Récupérer les stats token
	tokenStat, err := g.client.GetTokenStat(tokenAddress)
	if err != nil {
		g.logger.Error("Failed to get token stats", err, map[string]interface{}{
			"token_address": tokenAddress,
		})
		return nil, err
	}

	// Créer l'objet token
	token := &models.Token{
		Address:    tokenAddress,
		HolderCount: tokenStat.HolderCount,
	}

	// Enrichir avec les autres données
	completedTokens, err := g.client.GetCompletedCoins("100", "completed_at", "desc")
	if err == nil {
		// Chercher le token dans les tokens complétés
		for _, ct := range completedTokens.Data.Rank {
			if ct.Address == tokenAddress {
				token.Symbol = ct.Symbol
				token.Name = ct.Name
				token.CompletedTimestamp = ct.CompletedAt
				token.TotalSupply = ct.TotalSupply
				token.Logo = ct.Logo
				token.Twitter = ct.Twitter
				token.Website = ct.Website
				token.Telegram = ct.Telegram
				break
			}
		}
	} else {
		g.logger.Warning("Failed to get completed tokens", map[string]interface{}{
			"error": err.Error(),
		})
	}

	// Si le token n'a pas été trouvé dans les complétés, essayer les trending
	if token.Symbol == "" {
		trendingTokens, err := g.client.GetTrending("1h", "", "", nil)
		if err == nil {
			// Chercher le token dans les tokens trending
			for _, tt := range trendingTokens.Data.Rank {
				if tt.Address == tokenAddress {
					token.Symbol = tt.Symbol
					token.Name = tt.Name
					token.CreatedTimestamp = tt.CreatedTimestamp
					token.LastTradeTimestamp = tt.LastTradeTimestamp
					token.TotalSupply = int64(tt.TotalSupply)
					token.Logo = tt.Logo
					token.Twitter = tt.Twitter
					token.Website = tt.Website
					token.Telegram = tt.Telegram
					break
				}
			}
		} else {
			g.logger.Warning("Failed to get trending tokens", map[string]interface{}{
				"error": err.Error(),
			})
		}
	}

	return token, nil
}

// GetTokenStat récupère les statistiques d'un token
func (g *GMGNGateway) GetTokenStat(tokenAddress string) (*gmgn.TokenStatResponse, error) {
	g.logger.Debug("Getting token stats", map[string]interface{}{
		"token_address": tokenAddress,
	})

	return g.client.GetTokenStat(tokenAddress)
}

// GetTokenTrades récupère les transactions d'un token
func (g *GMGNGateway) GetTokenTrades(tokenAddress string, limit int, tag string) (*gmgn.TradeHistoryResponse, error) {
	g.logger.Debug("Getting token trades", map[string]interface{}{
		"token_address": tokenAddress,
		"limit":         limit,
		"tag":           tag,
	})

	return g.client.GetTokenTrades(tokenAddress, limit, tag)
}

// GetTokenPrice récupère les données de prix d'un token
func (g *GMGNGateway) GetTokenPrice(tokenAddress string, timeframe string) (*gmgn.KlineDataResponse, error) {
	g.logger.Debug("Getting token price", map[string]interface{}{
		"token_address": tokenAddress,
		"timeframe":     timeframe,
	})

	return g.client.GetTokenPrice(tokenAddress, timeframe)
}

// GetAllTokenTraders récupère tous les traders d'un token avec pagination
func (g *GMGNGateway) GetAllTokenTraders(tokenAddress string) ([]gmgn.Trader, error) {
	g.logger.Debug("Getting all token traders", map[string]interface{}{
		"token_address": tokenAddress,
	})

	return g.client.GetAllTokenTraders(tokenAddress)
}

// GetTokenHolderStat récupère les statistiques des détenteurs d'un token
func (g *GMGNGateway) GetTokenHolderStat(tokenAddress string) (*gmgn.TokenHolderStatResponse, error) {
	g.logger.Debug("Getting token holder stats", map[string]interface{}{
		"token_address": tokenAddress,
	})

	return g.client.GetTokenHolderStat(tokenAddress)
}

// GetTokenWalletTagsStat récupère les statistiques des tags de wallets pour un token
func (g *GMGNGateway) GetTokenWalletTagsStat(tokenAddress string) (*gmgn.TokenWalletTagsStatResponse, error) {
	g.logger.Debug("Getting token wallet tags stats", map[string]interface{}{
		"token_address": tokenAddress,
	})

	return g.client.GetTokenWalletTagsStat(tokenAddress)
}

// GetWalletInfo récupère les informations d'un wallet
func (g *GMGNGateway) GetWalletInfo(walletAddress string) (*gmgn.WalletInfoResponse, error) {
	g.logger.Debug("Getting wallet info", map[string]interface{}{
		"wallet_address": walletAddress,
	})

	return g.client.GetWalletInfo(walletAddress)
}

// GetAllWalletHoldings récupère tous les holdings d'un wallet avec pagination
func (g *GMGNGateway) GetAllWalletHoldings(walletAddress string) ([]gmgn.Holding, error) {
	g.logger.Debug("Getting all wallet holdings", map[string]interface{}{
		"wallet_address": walletAddress,
	})

	return g.client.GetAllWalletHoldings(walletAddress)
}

// GetWalletStat récupère les statistiques d'un wallet
func (g *GMGNGateway) GetWalletStat(walletAddress string, period string) (*gmgn.WalletStatResponse, error) {
	g.logger.Debug("Getting wallet stats", map[string]interface{}{
		"wallet_address": walletAddress,
		"period":         period,
	})

	return g.client.GetWalletStat(walletAddress, period)
}

// GetTrending récupère les tokens en tendance
func (g *GMGNGateway) GetTrending(timeframe string, orderBy string, direction string, filters []string) (*gmgn.TrendingResponse, error) {
	g.logger.Debug("Getting trending tokens", map[string]interface{}{
		"timeframe":  timeframe,
		"orderBy":    orderBy,
		"direction":  direction,
		"filters":    filters,
	})

	return g.client.GetTrending(timeframe, orderBy, direction, filters)
}

// GetCompletedCoins récupère les tokens complétés
func (g *GMGNGateway) GetCompletedCoins(limit string, orderBy string, direction string) (*gmgn.CompletedTokensResponse, error) {
	g.logger.Debug("Getting completed coins", map[string]interface{}{
		"limit":     limit,
		"orderBy":   orderBy,
		"direction": direction,
	})

	return g.client.GetCompletedCoins(limit, orderBy, direction)
} 