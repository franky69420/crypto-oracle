package memory

import (
	"context"

	"github.com/franko/crypto-oracle/pkg/models"
)

// MemoryOfTrust définit l'interface pour le système de confiance des wallets
type MemoryOfTrust interface {
	// Gestion du cycle de vie
	Start(ctx context.Context) error
	Stop() error
	
	// Opérations de base
	GetWalletTrustScore(walletAddress string) (float64, error)
	RecordWalletInteraction(interaction *models.WalletInteraction) error
	GetTokenTrustMetrics(tokenAddress string) (*models.TokenTrustMetrics, error)
	
	// Analyses de wallets
	GetWalletTokenHistory(walletAddress, tokenAddress string) ([]models.WalletInteraction, error)
	GetSimilarWallets(walletAddress string, minSimilarity float64, limit int) ([]models.WalletSimilarity, error)
	GetMostTrustedWallets(limit int) ([]models.WalletTrustScore, error)
	GetWalletRiskFactors(walletAddress string) (*models.WalletRiskFactors, error)
	
	// Analyses de tokens
	GetTokenInfluencers(tokenAddress string, limit int) ([]models.WalletInfluence, error)
	GetWalletTokens(walletAddress string, limit int) ([]models.WalletToken, error)
	
	// Gestion des similarités
	UpdateWalletSimilarities() error

	// Gestion des wallets actifs
	GetTokenActiveWallets(tokenAddress string, minTrustScore float64, limit int) ([]models.ActiveWallet, error)
	GetActiveWalletsCount(tokenAddress string) (int, error)
} 