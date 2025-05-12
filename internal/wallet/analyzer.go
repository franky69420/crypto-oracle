package wallet

import (
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/franko/crypto-oracle/internal/gateway/gmgn"
	"github.com/franko/crypto-oracle/internal/memory"
	"github.com/franko/crypto-oracle/pkg/models"

	"github.com/sirupsen/logrus"
)

// Analyzer est responsable de l'analyse des wallets pour déterminer leur comportement et fiabilité
type Analyzer struct {
	gateway       gmgn.Client
	memoryOfTrust memory.MemoryOfTrust
	logger        *logrus.Logger
}

// NewAnalyzer crée une nouvelle instance de l'analyseur de wallets
func NewAnalyzer(gateway gmgn.Client, memoryOfTrust memory.MemoryOfTrust, logger *logrus.Logger) *Analyzer {
	return &Analyzer{
		gateway:       gateway,
		memoryOfTrust: memoryOfTrust,
		logger:        logger,
	}
}

// AnalyzeTokenWallets effectue une analyse complète des wallets pour un token spécifique
func (a *Analyzer) AnalyzeTokenWallets(tokenAddress string) (*models.WalletAnalysis, error) {
	a.logger.Info("Analyzing wallets for token", logrus.Fields{
		"token_address": tokenAddress,
	})

	// Initialiser les statistiques
	stats := &models.WalletAnalysis{
		TokenAddress: tokenAddress,
		Timestamp:    time.Now(),
	}

	// Récupérer tous les traders du token
	traders, err := a.gateway.GetAllTokenTraders(tokenAddress)
	if err != nil {
		return nil, fmt.Errorf("échec de la récupération des traders: %w", err)
	}

	// Récupérer les statistiques des wallets du token
	holderStat, err := a.gateway.GetTokenHolderStat(tokenAddress)
	if err != nil {
		a.logger.Warn("Failed to get token holder stats", logrus.Fields{
			"token_address": tokenAddress,
			"error":         err.Error(),
		})
		// Continuer sans les stats
	}

	// Récupérer les tags des wallets
	walletTags, err := a.gateway.GetTokenWalletTagsStat(tokenAddress)
	if err != nil {
		a.logger.Warn("Failed to get wallet tags stats", logrus.Fields{
			"token_address": tokenAddress,
			"error":         err.Error(),
		})
		// Continuer sans les tags
	} else {
		// Remplir les catégories de wallets basées sur les distributions disponibles
		for _, dist := range walletTags.Distributions {
			switch dist.Tag {
			case "smart_money":
				stats.WalletCategories.Smart = dist.Count
			case "fresh_wallet":
				stats.WalletCategories.Fresh = dist.Count
			case "sniper":
				stats.WalletCategories.Sniper = dist.Count
				stats.SniperCount = dist.Count
			case "bluechip":
				stats.WalletCategories.Bluechip = dist.Count
			case "bundler":
				stats.WalletCategories.Bundler = dist.Count
			case "bot":
				stats.WalletCategories.Bot = dist.Count
			}
		}
	}

	// Initialiser les structures pour l'analyse
	walletDetails := make([]models.WalletDetail, 0, len(traders))
	trustScores := make([]float64, 0, len(traders))
	totalWallets := len(traders)
	stats.TotalWallets = totalWallets

	// Analyser chaque wallet
	for _, trader := range traders {
		// Obtenir le trust score depuis le Memory of Trust
		trustScore, err := a.memoryOfTrust.GetWalletTrustScore(trader.Address)
		if err != nil {
			a.logger.Warn("Failed to get trust score", logrus.Fields{
				"wallet_address": trader.Address,
				"error":          err.Error(),
			})
			trustScore = 50.0 // Score par défaut
		}
		
		trustScores = append(trustScores, trustScore)
		
		// Déterminer les catégories du wallet
		categories := a.categorizeWallet(trader, trustScore)
		
		// Ajouter aux détails des wallets
		walletDetails = append(walletDetails, models.WalletDetail{
			Address:    trader.Address,
			TrustScore: trustScore,
			Categories: categories,
			EntryRank:  0, // À déterminer
			EntryTime:  trader.LastTrade,
			Volume:     trader.BuyVolume + trader.SellVolume,
			Buys:       trader.TradeCount / 2, // Approximation
			Sells:      trader.TradeCount / 2, // Approximation
		})
	}
	
	// Trier les wallets par timestamp d'entrée (early investors first)
	sort.Slice(walletDetails, func(i, j int) bool {
		return walletDetails[i].EntryTime.Before(walletDetails[j].EntryTime)
	})
	
	// Mettre à jour les rangs d'entrée après le tri
	for i := range walletDetails {
		walletDetails[i].EntryRank = i + 1
	}
	
	// Calculer les métriques globales
	if len(trustScores) > 0 {
		// Calculer le score moyen
		var totalScore float64
		for _, score := range trustScores {
			totalScore += score
		}
		stats.TrustMetrics.AvgTrustScore = totalScore / float64(len(trustScores))
		
		// Compter les wallets smart
		var smartCount int
		for _, score := range trustScores {
			if score >= 70 {
				smartCount++
			}
		}
		stats.TrustMetrics.SmartMoneyCount = smartCount
		stats.TrustMetrics.TotalWallets = len(trustScores)
		
		if len(trustScores) > 0 {
			stats.TrustMetrics.SmartMoneyRatio = float64(smartCount) / float64(len(trustScores))
		}
	}
	
	// Détection early trusted
	earlyWallets := walletDetails
	if len(earlyWallets) > 10 {
		earlyWallets = earlyWallets[:10] // Prendre les 10 premiers wallets
	}
	
	var earlyTrustedCount int
	for _, w := range earlyWallets {
		if w.TrustScore > 70 {
			earlyTrustedCount++
		}
	}
	
	if len(earlyWallets) > 0 {
		stats.TrustMetrics.EarlyTrustedRatio = float64(earlyTrustedCount) / float64(len(earlyWallets))
	}
	
	// Calculer la métrique d'activité Smart Money
	// C'est une métrique composite qui reflète l'implication des wallets smart
	var recentActivity float64
	var totalBuyVolume float64
	var smartBuyVolume float64
	
	// Parcourir les traders pour calculer les volumes d'achat récents
	for _, trader := range traders {
		// Ne prendre en compte que les transactions récentes (dernières 24h)
		if time.Since(trader.LastTrade).Hours() < 24 {
			buyVolume := trader.BuyVolume
			totalBuyVolume += buyVolume
			
			// Vérifier si c'est un wallet smart
			for _, detail := range walletDetails {
				if detail.Address == trader.Address && detail.TrustScore >= 70 {
					smartBuyVolume += buyVolume
					break
				}
			}
		}
	}
	
	// Calculer le ratio d'activité smart
	if totalBuyVolume > 0 {
		recentActivity = (smartBuyVolume / totalBuyVolume) * 100 // En pourcentage
	}
	
	stats.TrustMetrics.SmartMoneyActivity = recentActivity
	
	// Calculer les métriques de trades
	var totalBuys, totalSells int
	for _, trader := range traders {
		// Approximation: répartir les transactions
		txCount := trader.TradeCount
		totalBuys += txCount / 2
		totalSells += txCount / 2
	}
	
	stats.TradePatterns.BuyOrders = totalBuys
	stats.TradePatterns.SellOrders = totalSells
	
	if totalSells > 0 {
		stats.TradePatterns.BuySellRatio = float64(totalBuys) / float64(totalSells)
	}
	
	// Stocker les détails des wallets
	stats.WalletDetails = walletDetails
	
	// Calculer le ratio de snipers
	if totalWallets > 0 {
		stats.SniperRatio = float64(stats.SniperCount) / float64(totalWallets)
	}
	
	a.logger.Info("Wallet analysis completed", logrus.Fields{
		"token_address":     tokenAddress,
		"total_wallets":     stats.TotalWallets,
		"smart_money_ratio": stats.TrustMetrics.SmartMoneyRatio,
		"smart_money_count": stats.TrustMetrics.SmartMoneyCount,
		"sniper_count":      stats.SniperCount,
		"buy_sell_ratio":    stats.TradePatterns.BuySellRatio,
	})
	
	return stats, nil
}

// categorizeWallet détermine les catégories d'un wallet basées sur son comportement
func (a *Analyzer) categorizeWallet(trader gmgn.Trader, trustScore float64) []string {
	categories := make([]string, 0)
	
	// Catégorisation basée sur le trust score
	if trustScore >= 70 {
		categories = append(categories, "smart")
	}
	
	if trustScore >= 60 {
		categories = append(categories, "trusted")
	}
	
	// Vérifier les tags
	for _, tag := range trader.Tags {
		switch tag.Name {
		case "fresh_wallet", "new":
			categories = append(categories, "fresh")
		case "sniper":
			categories = append(categories, "sniper")
		case "bundler":
			categories = append(categories, "bundler")
		case "bluechip", "whale":
			categories = append(categories, "bluechip")
		case "bot", "dex_bot":
			categories = append(categories, "bot")
		}
	}
	
	return categories
}

// GetWalletProfile récupère le profil complet d'un wallet
func (a *Analyzer) GetWalletProfile(walletAddress string) (*models.WalletProfile, error) {
	// Initialiser le profil
	profile := &models.WalletProfile{
		Address:           walletAddress,
		CreatedAt:         time.Now(), // Approximatif
		LastActive:        time.Now(), // À remplacer par la vraie date
		TrustScore:        50.0,       // Valeur par défaut
		Tags:              []string{},
		TotalTransactions: 0,
		WinRate:           0.0,
		AvgProfitPerTrade: 0.0,
	}
	
	// Récupérer les informations du wallet
	walletInfo, err := a.gateway.GetWalletInfo(walletAddress)
	if err != nil {
		return profile, fmt.Errorf("échec de la récupération des infos wallet: %w", err)
	}
	
	// Convertir les tags en strings
	for _, tag := range walletInfo.Tags {
		profile.Tags = append(profile.Tags, tag.Name)
	}
	
	// Récupérer les statistiques du wallet
	walletStat, err := a.gateway.GetWalletStat(walletAddress, "all")
	if err != nil {
		a.logger.Warn("Failed to get wallet stats", logrus.Fields{
			"wallet_address": walletAddress,
			"error":         err.Error(),
		})
		// Continuer sans les stats
	} else {
		// Remplir le profil avec les statistiques approximatives
		profile.TotalTransactions = walletStat.TotalTrades
		profile.WinRate = float64(walletStat.WinningTrades) / float64(walletStat.TotalTrades)
		profile.AvgProfitPerTrade = walletStat.TotalProfit / float64(walletStat.TotalTrades)
		
		// Définir les facteurs de risque (simplifiés)
		profile.RiskFactors = models.WalletRiskFactors{
			FastTxRatio:        0.1, // Valeurs par défaut simplifiées
			SellPassBuyRatio:   0.1,
			NoBuyHoldRatio:     0.1,
			TokenHoneypotRatio: 0.1,
		}
	}
	
	// Récupérer le score de confiance depuis le Memory of Trust
	trustScore, err := a.memoryOfTrust.GetWalletTrustScore(walletAddress)
	if err != nil {
		a.logger.Warn("Failed to get trust score for profile", logrus.Fields{
			"wallet_address": walletAddress,
			"error":          err.Error(),
		})
		// Utiliser score par défaut
	} else {
		profile.TrustScore = trustScore
	}
	
	// Récupérer les holdings du wallet
	holdings, err := a.gateway.GetAllWalletHoldings(walletAddress)
	if err != nil {
		a.logger.Warn("Failed to get wallet holdings", logrus.Fields{
			"wallet_address": walletAddress,
			"error":          err.Error(),
		})
		// Continuer sans les holdings
	} else {
		// Convertir les holdings en modèle interne
		for _, holding := range holdings {
			// Convertir les valeurs numériques en chaînes pour les nouveaux champs
			balance := fmt.Sprintf("%.6f", holding.Amount)
			value := fmt.Sprintf("%.2f", holding.UsdValue)
			
			profile.Holdings = append(profile.Holdings, models.WalletHolding{
				TokenAddress:     holding.TokenAddress,
				TokenSymbol:      holding.TokenSymbol,
				Balance:          balance,
				Value:            value,
				UnrealizedProfit: holding.Price * holding.Amount * 0.1, // Simplifié
				BuyCount:         5, // Valeur par défaut
				SellCount:        2, // Valeur par défaut
				LastActive:       time.Now().Add(-24 * time.Hour), // Valeur par défaut
			})
		}
	}
	
	return profile, nil
}

// IsSniperWallet détermine si un wallet présente des caractéristiques d'un sniper
func (a *Analyzer) IsSniperWallet(walletAddress string) (bool, float64, error) {
	// Un sniper est généralement un wallet qui:
	// 1. A un historique de trades rapides après le lancement des tokens
	// 2. Présente un ratio achat/vente élevé dans des fenêtres courtes
	// 3. Montre un pattern d'entrée très tôt dans le cycle de vie des tokens
	
	// Récupérer le profil du wallet
	profile, err := a.GetWalletProfile(walletAddress)
	if err != nil {
		return false, 0.0, err
	}
	
	// Calculer le score sniper simplifié
	sniperScore := 0.0
	
	// Vérifier si déjà taggé comme sniper dans les tags API
	for _, tag := range profile.Tags {
		if tag == "sniper" {
			sniperScore += 60 // Bonus si déjà identifié comme sniper
			break
		}
	}
	
	// Facteur basé sur risk factors (simplifiés)
	sniperScore += profile.RiskFactors.FastTxRatio * 40 // Max 40 points
	
	// Décision finale
	isSniper := sniperScore >= 50 // Seuil de 50 points
	
	return isSniper, sniperScore, nil
}

// IsSmartMoneyWallet détermine si un wallet est considéré comme "smart money"
func (a *Analyzer) IsSmartMoneyWallet(walletAddress string) (bool, float64, error) {
	// Smart money est généralement un wallet qui:
	// 1. A un historique de profit consistant
	// 2. Choisit des tokens qui performent bien
	// 3. A un timing d'entrée/sortie optimal
	
	// Récupérer le profil du wallet
	profile, err := a.GetWalletProfile(walletAddress)
	if err != nil {
		return false, 0.0, err
	}
	
	// Calculer le score smart money
	smartScore := 0.0
	
	// Facteur 1: Win rate (50 points max)
	if profile.WinRate > 0.5 {
		smartScore += (profile.WinRate - 0.5) * 100 // 0-50 points
	}
	
	// Facteur 2: Trust score from Memory of Trust
	trustScore := profile.TrustScore
	if trustScore > 60 {
		smartScore += (trustScore - 60) * 0.5 // 0-20 points bonus
	}
	
	// Vérifier les tags
	for _, tag := range profile.Tags {
		if tag == "smart" || tag == "smart_money" {
			smartScore += 30
			break
		}
	}
	
	// Plafonner le score à 100
	if smartScore > 100 {
		smartScore = 100
	}
	
	// Décision finale
	isSmart := smartScore >= 70 // Seuil de 70 points
	
	return isSmart, smartScore, nil
} 