package wallet

import (
	"context"
	"fmt"
	"time"

	"github.com/franko/crypto-oracle/internal/memory"
	"github.com/franko/crypto-oracle/pkg/models"

	"github.com/sirupsen/logrus"
)

// Intelligence est le service principal pour l'analyse et le traitement des wallets
type Intelligence struct {
	analyzer      *Analyzer
	memoryOfTrust memory.MemoryOfTrust
	logger        *logrus.Logger
}

// NewIntelligence crée une nouvelle instance du service d'intelligence des wallets
func NewIntelligence(memoryOfTrust memory.MemoryOfTrust, logger *logrus.Logger) *Intelligence {
	return &Intelligence{
		memoryOfTrust: memoryOfTrust,
		logger:        logger,
	}
}

// SetAnalyzer permet de définir l'analyseur de wallets
func (i *Intelligence) SetAnalyzer(analyzer *Analyzer) {
	i.analyzer = analyzer
}

// AnalyzeWallet effectue une analyse complète d'un wallet
func (i *Intelligence) AnalyzeWallet(walletAddress string) (*models.WalletProfile, error) {
	i.logger.Info("Analyzing wallet", logrus.Fields{
		"wallet_address": walletAddress,
	})

	profile, err := i.analyzer.GetWalletProfile(walletAddress)
	if err != nil {
		return nil, err
	}

	// Enrichir le profil avec des métadonnées supplémentaires
	isSniper, sniperScore, err := i.analyzer.IsSniperWallet(walletAddress)
	if err == nil && isSniper {
		profile.Tags = append(profile.Tags, "sniper")
	}

	isSmart, smartScore, err := i.analyzer.IsSmartMoneyWallet(walletAddress)
	if err == nil && isSmart {
		profile.Tags = append(profile.Tags, "smart_money")
	}

	// Mettre à jour les données dans Memory of Trust
	i.memoryOfTrust.UpdateWalletTrustScore(walletAddress, profile.TrustScore)

	i.logger.Info("Wallet analysis completed", logrus.Fields{
		"wallet_address": walletAddress,
		"is_sniper":      isSniper,
		"sniper_score":   sniperScore,
		"is_smart":       isSmart,
		"smart_score":    smartScore,
		"trust_score":    profile.TrustScore,
	})

	return profile, nil
}

// DetectRelatedWallets trouve les wallets potentiellement liés au wallet spécifié
func (i *Intelligence) DetectRelatedWallets(walletAddress string) ([]string, error) {
	// Implémentation simplifiée de la détection des wallets liés
	// Dans une implémentation réelle, cela analyserait les patterns de transactions,
	// les comportements de trading coordonnés, et d'autres signaux
	return []string{}, nil
}

// EvaluateTokenHolders évalue la qualité des détenteurs d'un token
func (i *Intelligence) EvaluateTokenHolders(tokenAddress string) (*models.HolderQualityReport, error) {
	// Récupérer l'analyse des wallets pour ce token
	walletAnalysis, err := i.analyzer.AnalyzeTokenWallets(tokenAddress)
	if err != nil {
		return nil, err
	}

	// Calculer le score de qualité basé sur les métriques d'analyse
	smartMoneyRatio := walletAnalysis.TrustMetrics.SmartMoneyRatio
	earlyTrustedRatio := walletAnalysis.TrustMetrics.EarlyTrustedRatio
	sniperRatio := walletAnalysis.SniperRatio

	// Calculer le score final (0-100)
	qualityScore := (smartMoneyRatio*0.5 + earlyTrustedRatio*0.3) * 100
	
	// Pénalité pour trop de snipers
	if sniperRatio > 0.1 {
		qualityScore -= (sniperRatio - 0.1) * 200 // Pénalité proportionnelle
	}

	// Limiter le score entre 0 et 100
	if qualityScore < 0 {
		qualityScore = 0
	} else if qualityScore > 100 {
		qualityScore = 100
	}

	// Déterminer la distribution des catégories
	categoryDistribution := make(map[string]float64)
	categoryDistribution["smart"] = float64(walletAnalysis.WalletCategories.Smart) / float64(walletAnalysis.TotalWallets)
	categoryDistribution["trusted"] = float64(walletAnalysis.WalletCategories.Trusted) / float64(walletAnalysis.TotalWallets)
	categoryDistribution["fresh"] = float64(walletAnalysis.WalletCategories.Fresh) / float64(walletAnalysis.TotalWallets)
	categoryDistribution["bot"] = float64(walletAnalysis.WalletCategories.Bot) / float64(walletAnalysis.TotalWallets)
	categoryDistribution["sniper"] = float64(walletAnalysis.WalletCategories.Sniper) / float64(walletAnalysis.TotalWallets)
	categoryDistribution["bluechip"] = float64(walletAnalysis.WalletCategories.Bluechip) / float64(walletAnalysis.TotalWallets)

	// Créer le rapport final
	report := &models.HolderQualityReport{
		TokenAddress:         tokenAddress,
		TotalHolders:         walletAnalysis.TotalWallets,
		QualityScore:         qualityScore,
		SmartMoneyRatio:      smartMoneyRatio,
		SmartMoneyCount:      walletAnalysis.TrustMetrics.SmartMoneyCount,
		EarlyTrustedRatio:    earlyTrustedRatio,
		SniperRatio:          sniperRatio,
		SniperCount:          walletAnalysis.SniperCount,
		CategoryDistribution: categoryDistribution,
		Timestamp:            time.Now(),
	}

	i.logger.Info("Token holder quality evaluation completed", logrus.Fields{
		"token_address":    tokenAddress,
		"quality_score":    qualityScore,
		"total_holders":    walletAnalysis.TotalWallets,
		"smart_money_count": walletAnalysis.TrustMetrics.SmartMoneyCount,
	})

	return report, nil
}

// Start démarre les processus d'intelligence des wallets
func (i *Intelligence) Start(ctx context.Context) error {
	i.logger.Info("Starting Wallet Intelligence service")

	// Aucun processus à démarrer dans cette implémentation simple
	return nil
}

// Shutdown arrête les processus d'intelligence des wallets
func (i *Intelligence) Shutdown(ctx context.Context) error {
	i.logger.Info("Shutting down Wallet Intelligence service")

	// Aucun processus à arrêter dans cette implémentation simple
	return nil
} 