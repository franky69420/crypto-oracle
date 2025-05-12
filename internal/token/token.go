package token

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/franky69420/crypto-oracle/internal/memory"
	"github.com/franky69420/crypto-oracle/internal/pipeline"
	"github.com/franky69420/crypto-oracle/pkg/models"
	"github.com/sirupsen/logrus"
)

// Engine gère les opérations sur les tokens
type Engine struct {
	gmgn          interface {
		GetTokenInfo(tokenAddress string) (*models.Token, error)
		GetTokenStats(tokenAddress string) (*models.TokenStats, error)
		GetTokenTrades(tokenAddress string, limit int) ([]models.TokenTrade, error)
		GetTokenPrice(tokenAddress string) (*models.TokenPrice, error)
		GetWalletTokenTrades(walletAddress, tokenAddress string, limit int) ([]models.TokenTrade, error)
	}
	memoryOfTrust memory.MemoryOfTrust
	pipelineSvc   *pipeline.Pipeline
	logger        *logrus.Logger
	tokens        map[string]*models.Token // Cache en mémoire, à remplacer par Redis en prod
	metrics       map[string]*models.TokenMetrics
}

// NewEngine crée un nouveau moteur de token
func NewEngine(gmgn interface {
		GetTokenInfo(tokenAddress string) (*models.Token, error)
		GetTokenStats(tokenAddress string) (*models.TokenStats, error)
		GetTokenTrades(tokenAddress string, limit int) ([]models.TokenTrade, error)
		GetTokenPrice(tokenAddress string) (*models.TokenPrice, error)
		GetWalletTokenTrades(walletAddress, tokenAddress string, limit int) ([]models.TokenTrade, error)
	}, memoryOfTrust memory.MemoryOfTrust, pipelineSvc *pipeline.Pipeline, logger *logrus.Logger) *Engine {
	return &Engine{
		gmgn:          gmgn,
		memoryOfTrust: memoryOfTrust,
		pipelineSvc:   pipelineSvc,
		logger:        logger,
		tokens:        make(map[string]*models.Token),
		metrics:       make(map[string]*models.TokenMetrics),
	}
}

// Start initialise le moteur de token
func (e *Engine) Start(ctx context.Context) error {
	e.logger.Info("Starting Token Engine")
	return nil
}

// Shutdown arrête proprement le moteur de token
func (e *Engine) Shutdown(ctx context.Context) error {
	e.logger.Info("Shutting down Token Engine")
	return nil
}

// GetToken récupère les informations d'un token
func (e *Engine) GetToken(tokenAddress string) (*models.Token, error) {
	// Vérifier le cache en mémoire
	if token, ok := e.tokens[tokenAddress]; ok {
		return token, nil
	}

	// Récupérer via l'API GMGN
	tokenInfo, err := e.gmgn.GetTokenInfo(tokenAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to get token info: %w", err)
	}

	// Le token est déjà dans le bon format
	token := &models.Token{
		Address:     tokenAddress,
		Symbol:      tokenInfo.Symbol,
		Name:        tokenInfo.Name,
		TotalSupply: tokenInfo.TotalSupply,
		HolderCount: tokenInfo.HolderCount,
		Logo:        tokenInfo.Logo,
		Twitter:     tokenInfo.Twitter,
		Website:     tokenInfo.Website,
		Telegram:    tokenInfo.Telegram,
		CachedAt:    time.Now(),
	}

	// Mettre en cache
	e.tokens[tokenAddress] = token

	return token, nil
}

// GetTokenMetrics récupère les métriques d'un token
func (e *Engine) GetTokenMetrics(tokenAddress string) (*models.TokenMetrics, error) {
	// Récupérer les métriques via l'API GMGN
	tokenStats, err := e.gmgn.GetTokenStats(tokenAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to get token stats: %w", err)
	}

	// Créer l'objet métriques
	metrics := &models.TokenMetrics{
		TokenAddress:      tokenAddress,
		HolderCount:       tokenStats.HolderCount,
		Volume1h:          tokenStats.Volume1h,
		Volume24h:         tokenStats.Volume24h,
		Price:             tokenStats.Price,
		MarketCap:         tokenStats.MarketCap,
		PriceChange1h:     tokenStats.PriceChange1h,
		BuyCount1h:        tokenStats.BuyCount1h,
		SellCount1h:       tokenStats.SellCount1h,
		UpdatedAt:         time.Now(),
	}

	// Enrichir avec des données du Memory of Trust
	trustMetrics, err := e.memoryOfTrust.GetTokenTrustMetrics(tokenAddress)
	if err == nil {
		metrics.IntelligentHolders = trustMetrics.TrustedWallets
		metrics.AverageTrustScore = trustMetrics.AvgTrustScore
		metrics.SmartMoneyHolders = trustMetrics.SmartMoneyCount
	}

	return metrics, nil
}

// GetTokenLastSnapshot récupère le dernier snapshot enregistré des métriques d'un token
func (e *Engine) GetTokenLastSnapshot(tokenAddress string) (*models.TokenMetrics, error) {
	// Logique à implémenter - ici on simule juste un retour
	// Dans une implémentation réelle, on récupérerait depuis TimescaleDB
	return &models.TokenMetrics{
		TokenAddress:  tokenAddress,
		HolderCount:   100,
		Volume1h:      1000.0,
		Volume24h:     24000.0,
		Price:         0.1,
		MarketCap:     100000.0,
		PriceChange1h: 0.05,
		UpdatedAt:     time.Now().Add(-24 * time.Hour),
	}, nil
}

// GetTokensByStates récupère les tokens par leur état de cycle de vie
func (e *Engine) GetTokensByStates(states []string) ([]models.Token, error) {
	// Simulation de récupération - à implémenter réellement avec DB
	var tokens []models.Token
	
	// Dans une vraie implémentation, on ferait une requête DB
	// SELECT * FROM tokens WHERE lifecycle_state IN (states)
	
	// Pour l'instant, on renvoie des données de test
	tokens = append(tokens, models.Token{
		Address:  "SoLDogMKjM9YMzSQzp7SuBYQCM9LCCgBkrysTNxMD3m",
		Symbol:   "SOLDOGE",
		Name:     "Solana Doge",
		HolderCount: 1200,
	})
	
	tokens = append(tokens, models.Token{
		Address:  "CATZwdqR8Prd2RRK1mXnQvh698GziRn4Tw8zKcQfNPdS",
		Symbol:   "CATZ",
		Name:     "Catz Token",
		HolderCount: 850,
	})
	
	return tokens, nil
}

// GetTokenRecentTrades récupère les trades récents d'un token
func (e *Engine) GetTokenRecentTrades(tokenAddress string, hours int) ([]models.TokenTrade, error) {
	// Récupérer via l'API GMGN
	trades, err := e.gmgn.GetTokenTrades(tokenAddress, 100)
	if err != nil {
		return nil, fmt.Errorf("failed to get token trades: %w", err)
	}
	
	// Convertir au format interne et filtrer par temps
	var recentTrades []models.TokenTrade
	cutoffTime := time.Now().Add(-time.Duration(hours) * time.Hour)
	
	for _, trade := range trades {
		if trade.Timestamp.After(cutoffTime) {
			recentTrades = append(recentTrades, models.TokenTrade{
				ID:            fmt.Sprintf("%s-%d", trade.TxHash, trade.BlockNumber),
				TokenAddress:  tokenAddress,
				WalletAddress: trade.WalletAddress,
				TradeType:     trade.TradeType,
				Amount:        trade.Amount,
				Price:         trade.Price,
				TotalValue:    trade.Amount * trade.Price,
				Timestamp:     trade.Timestamp,
				TxHash:        trade.TxHash,
				BlockNumber:   trade.BlockNumber,
			})
		}
	}
	
	return recentTrades, nil
}

// GetWalletTokenHistory récupère l'historique des interactions d'un wallet avec un token
func (e *Engine) GetWalletTokenHistory(walletAddress, tokenAddress string) ([]models.TokenTrade, error) {
	// Récupérer via l'API GMGN
	trades, err := e.gmgn.GetWalletTokenTrades(walletAddress, tokenAddress, 100)
	if err != nil {
		return nil, fmt.Errorf("failed to get wallet token trades: %w", err)
	}
	
	// Convertir au format interne
	var history []models.TokenTrade
	for _, trade := range trades {
		history = append(history, models.TokenTrade{
			ID:            fmt.Sprintf("%s-%d", trade.TxHash, trade.BlockNumber),
			TokenAddress:  tokenAddress,
			WalletAddress: walletAddress,
			TradeType:     trade.TradeType,
			Amount:        trade.Amount,
			Price:         trade.Price,
			TotalValue:    trade.Amount * trade.Price,
			Timestamp:     trade.Timestamp,
			TxHash:        trade.TxHash,
			BlockNumber:   trade.BlockNumber,
		})
	}
	
	return history, nil
}

// UpdateTokenState met à jour l'état d'un token et publie un événement
func (e *Engine) UpdateTokenState(tokenAddress, newState string) error {
	// Récupérer le token
	token, err := e.GetToken(tokenAddress)
	if err != nil {
		return fmt.Errorf("failed to get token: %w", err)
	}

	// Mettre à jour l'état (dans une implémentation réelle, il faudrait modifier la base de données)
	oldState := "unknown" // Dans une implémentation complète, récupérer l'ancien état

	e.logger.WithFields(logrus.Fields{
		"token_address": tokenAddress,
		"token_symbol":  token.Symbol,
		"old_state":     oldState,
		"new_state":     newState,
	}).Info("Updating token state")

	// Publier un événement si le pipeline est disponible
	if e.pipelineSvc != nil {
		event := pipeline.Message{
			Type:      "state_change",
			Timestamp: time.Now(),
			Payload: map[string]interface{}{
				"token_address": tokenAddress,
				"token_symbol":  token.Symbol,
				"old_state":     oldState,
				"new_state":     newState,
			},
		}

		if err := e.pipelineSvc.PublishMessage("token_events", event); err != nil {
			e.logger.WithError(err).Warn("Failed to publish state change event")
			// Continuer malgré l'erreur
		}
	}

	return nil
}

// SaveReactivationMetrics enregistre les métriques de réactivation et génère un événement
func (e *Engine) SaveReactivationMetrics(candidate models.ReactivationCandidate) error {
	e.logger.WithFields(logrus.Fields{
		"token_address":      candidate.TokenAddress,
		"token_symbol":       candidate.TokenSymbol,
		"reactivation_score": candidate.ReactivationScore,
	}).Info("Saving reactivation metrics")

	// Dans une implémentation réelle, sauvegarder dans la base de données

	// Publier un événement si le pipeline est disponible
	if e.pipelineSvc != nil {
		event := pipeline.Message{
			Type:      "reactivation",
			Timestamp: time.Now(),
			Payload: map[string]interface{}{
				"token_address":      candidate.TokenAddress,
				"token_symbol":       candidate.TokenSymbol,
				"reactivation_score": candidate.ReactivationScore,
				"changes":            candidate.Changes,
				"detected_at":        candidate.DetectedAt.Format(time.RFC3339),
				"smart_returns":      candidate.SmartReturns != nil,
			},
		}

		if err := e.pipelineSvc.PublishMessage("token_events", event); err != nil {
			e.logger.WithError(err).Warn("Failed to publish reactivation event")
			// Continuer malgré l'erreur
		}
	}

	return nil
}

// MonitorPriceMovements surveille les mouvements de prix et génère des événements
func (e *Engine) MonitorPriceMovements(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	e.logger.Info("Starting price movement monitoring")

	for {
		select {
		case <-ctx.Done():
			e.logger.Info("Stopping price movement monitoring")
			return
		case <-ticker.C:
			e.checkPriceMovements()
		}
	}
}

// checkPriceMovements vérifie les mouvements de prix significatifs
func (e *Engine) checkPriceMovements() {
	e.logger.Debug("Checking price movements")

	// Dans une implémentation réelle, récupérer les tokens à surveiller depuis la base de données
	// Pour l'exemple, utiliser les tokens en cache
	for addr, token := range e.tokens {
		// Récupérer les métriques actuelles
		currentMetrics, err := e.GetTokenMetrics(addr)
		if err != nil {
			e.logger.WithError(err).Warn("Failed to get token metrics")
			continue
		}

		// Récupérer les métriques précédentes depuis le cache
		var prevMetrics *models.TokenMetrics
		var ok bool
		if prevMetrics, ok = e.metrics[addr]; !ok {
			// Si pas de métriques précédentes, enregistrer les actuelles et continuer
			e.metrics[addr] = currentMetrics
			continue
		}

		// Calculer le changement de prix
		priceChange := 0.0
		if prevMetrics.Price > 0 {
			priceChange = (currentMetrics.Price - prevMetrics.Price) / prevMetrics.Price * 100
		}

		// Calculer le changement de volume
		volumeChange := 0.0
		if prevMetrics.Volume24h > 0 {
			volumeChange = (currentMetrics.Volume24h - prevMetrics.Volume24h) / prevMetrics.Volume24h * 100
		}

		// Mettre à jour les métriques en cache
		e.metrics[addr] = currentMetrics

		// Générer des événements si changements significatifs
		if math.Abs(priceChange) >= 5 {
			// Changement de prix de 5% ou plus
			e.publishPriceChangeEvent(token, priceChange)
		}

		if volumeChange >= 20 {
			// Augmentation de volume de 20% ou plus
			e.publishVolumeChangeEvent(token, volumeChange)
		}
	}
}

// publishPriceChangeEvent publie un événement de changement de prix
func (e *Engine) publishPriceChangeEvent(token *models.Token, priceChange float64) {
	if e.pipelineSvc == nil {
		return
	}

	e.logger.WithFields(logrus.Fields{
		"token_address": token.Address,
		"token_symbol":  token.Symbol,
		"price_change":  priceChange,
	}).Info("Significant price change detected")

	event := pipeline.Message{
		Type:      "price_change",
		Timestamp: time.Now(),
		Payload: map[string]interface{}{
			"token_address": token.Address,
			"token_symbol":  token.Symbol,
			"price_change":  priceChange,
			"is_positive":   priceChange > 0,
		},
	}

	if err := e.pipelineSvc.PublishMessage("token_events", event); err != nil {
		e.logger.WithError(err).Warn("Failed to publish price change event")
	}
}

// publishVolumeChangeEvent publie un événement de changement de volume
func (e *Engine) publishVolumeChangeEvent(token *models.Token, volumeChange float64) {
	if e.pipelineSvc == nil {
		return
	}

	e.logger.WithFields(logrus.Fields{
		"token_address": token.Address,
		"token_symbol":  token.Symbol,
		"volume_change": volumeChange,
	}).Info("Significant volume change detected")

	event := pipeline.Message{
		Type:      "volume_spike",
		Timestamp: time.Now(),
		Payload: map[string]interface{}{
			"token_address": token.Address,
			"token_symbol":  token.Symbol,
			"volume_change": volumeChange,
		},
	}

	if err := e.pipelineSvc.PublishMessage("token_events", event); err != nil {
		e.logger.WithError(err).Warn("Failed to publish volume change event")
	}
}

// CalculateXScore calcule le X-Score pour un token
func (e *Engine) CalculateXScore(tokenAddress string, walletAnalysis *models.WalletAnalysis) (*models.XScoreResult, error) {
	// Récupérer les métriques du token
	metrics, err := e.GetTokenMetrics(tokenAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to get token metrics: %w", err)
	}
	
	// Récupérer le token
	token, err := e.GetToken(tokenAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to get token: %w", err)
	}
	
	// Si l'analyse des wallets n'est pas fournie, en faire une
	if walletAnalysis == nil {
		// À implémenter avec le service wallet
		walletAnalysis = &models.WalletAnalysis{
			TokenAddress: tokenAddress,
			TotalWallets: metrics.HolderCount,
		}
	}
	
	// Initialiser les composants du score
	components := make(map[string]float64)
	
	// 1. Qualité Token (20%)
	tokenQuality := e.calculateTokenQuality(token, metrics)
	components["token_quality"] = tokenQuality * 0.20
	
	// 2. Wallet Quality (25%)
	walletQuality := e.calculateWalletQuality(walletAnalysis)
	components["wallet_quality"] = walletQuality * 0.25
	
	// 3. Memory of Trust (20%)
	trustFactor := e.calculateTrustFactor(token, walletAnalysis)
	components["trust_factor"] = trustFactor * 0.20
	
	// 4. Market Dynamics (15%)
	marketFactor := e.calculateMarketDynamics(metrics)
	components["market_factor"] = marketFactor * 0.15
	
	// 5. Temporal Patterns (10%)
	temporalFactor := e.calculateTemporalPatterns(metrics)
	components["temporal_factor"] = temporalFactor * 0.10
	
	// 6. Reactivation Boost (10%)
	reactivationFactor := e.calculateReactivationFactor(token, metrics)
	components["reactivation_factor"] = reactivationFactor * 0.10
	
	// NOUVEAU: Bonus Sniper Wallets
	sniperCount := walletAnalysis.SniperCount
	sniperBonus := 5 * math.Min(1.0, float64(sniperCount) / 3)
	components["sniper_bonus"] = sniperBonus
	
	// NOUVEAU: Pondération price_change × smart_money_ratio
	priceChange := metrics.PriceChange1h
	smartMoneyRatio := walletAnalysis.TrustMetrics.SmartMoneyRatio
	
	// Boost significatif si le prix augmente ET que les smart money sont présents
	priceSmartBoost := priceChange * smartMoneyRatio * 10
	components["price_smart_boost"] = priceSmartBoost
	
	// Score de base (somme des composantes)
	baseScore := 0.0
	for _, value := range components {
		baseScore += value
	}
	
	// Anti-Dump Check
	antiDump := e.checkAntiDumpPattern(tokenAddress, walletAnalysis)
	
	// Application pénalité dump si détecté
	finalScore := baseScore
	if antiDump.Detected {
		// Pénalité proportionnelle à la sévérité
		dumpPenalty := math.Min(0.90, antiDump.Severity / 100)
		finalScore = baseScore * (1.0 - dumpPenalty)
		components["anti_dump_penalty"] = -baseScore * dumpPenalty
	}
	
	// Range final 0-100
	finalScore = math.Max(0, math.Min(100, finalScore))
	
	return &models.XScoreResult{
		TokenAddress: tokenAddress,
		XScore:       finalScore,
		BaseScore:    baseScore,
		Components:   components,
		AntiDump:     antiDump,
		CalculatedAt: time.Now(),
	}, nil
}

// calculateTokenQuality calcule le score de qualité du token
func (e *Engine) calculateTokenQuality(token *models.Token, metrics *models.TokenMetrics) float64 {
	quality := 50.0 // Score de base
	
	// Facteurs positifs
	if token.HolderCount > 1000 {
		quality += 10.0
	} else if token.HolderCount > 500 {
		quality += 5.0
	}
	
	if metrics.MarketCap > 1000000 {
		quality += 10.0
	} else if metrics.MarketCap > 500000 {
		quality += 5.0
	}
	
	// Présence sociale
	if token.Website != "" {
		quality += 5.0
	}
	if token.Twitter != "" {
		quality += 5.0
	}
	if token.Telegram != "" {
		quality += 5.0
	}
	
	// Facteurs négatifs
	volumeMcapRatio := 0.0
	if metrics.MarketCap > 0 {
		volumeMcapRatio = metrics.Volume1h / metrics.MarketCap
	}
	
	if volumeMcapRatio > 0.5 {
		quality -= 20.0 // Potentiel wash trading
	} else if volumeMcapRatio > 0.3 {
		quality -= 10.0
	}
	
	// Normaliser entre 0-100
	return math.Max(0, math.Min(100, quality))
}

// calculateWalletQuality calcule le score de qualité des wallets
func (e *Engine) calculateWalletQuality(walletAnalysis *models.WalletAnalysis) float64 {
	quality := 50.0 // Score de base
	
	// Facteurs liés aux wallets "indésirables"
	totalWallets := walletAnalysis.TotalWallets
	if totalWallets > 0 {
		// Pénalité pour présence excessive de wallets frais ou bots
		freshRatio := float64(walletAnalysis.WalletCategories.Fresh) / float64(totalWallets)
		botRatio := float64(walletAnalysis.WalletCategories.Bot) / float64(totalWallets)
		
		if freshRatio > 0.7 {
			quality -= 30.0
		} else if freshRatio > 0.5 {
			quality -= 15.0
		}
		
		if botRatio > 0.4 {
			quality -= 20.0
		} else if botRatio > 0.2 {
			quality -= 10.0
		}
		
		// Bonus pour présence de wallets de qualité
		blueChipRatio := float64(walletAnalysis.WalletCategories.Bluechip) / float64(totalWallets)
		
		if blueChipRatio > 0.1 {
			quality += 20.0
		} else if blueChipRatio > 0.05 {
			quality += 10.0
		}
	}
	
	// Facteurs liés au ratio buy/sell
	buySellRatio := walletAnalysis.TradePatterns.BuySellRatio
	if buySellRatio > 3.0 {
		quality += 15.0 // Fort consensus à l'achat
	} else if buySellRatio > 2.0 {
		quality += 10.0
	} else if buySellRatio < 0.5 {
		quality -= 20.0 // Forte pression de vente
	} else if buySellRatio < 0.8 {
		quality -= 10.0
	}
	
	// Normaliser entre 0-100
	return math.Max(0, math.Min(100, quality))
}

// calculateTrustFactor calcule le facteur de confiance basé sur le Memory of Trust
func (e *Engine) calculateTrustFactor(token *models.Token, walletAnalysis *models.WalletAnalysis) float64 {
	trust := 50.0 // Score de base
	
	// Facteurs basés sur la présence de wallets smart et trusted
	if walletAnalysis.TotalWallets > 0 {
		smartMoneyRatio := walletAnalysis.TrustMetrics.SmartMoneyRatio
		
		if smartMoneyRatio > 0.2 {
			trust += 30.0
		} else if smartMoneyRatio > 0.1 {
			trust += 20.0
		} else if smartMoneyRatio > 0.05 {
			trust += 10.0
		}
		
		// L'importance des early trusted wallets
		if walletAnalysis.TrustMetrics.EarlyTrustedRatio > 0.5 {
			trust += 20.0
		} else if walletAnalysis.TrustMetrics.EarlyTrustedRatio > 0.3 {
			trust += 10.0
		}
	}
	
	// Facteur basé sur l'activité récente des smart wallets
	smartMoneyActivity := walletAnalysis.TrustMetrics.SmartMoneyActivity
	if smartMoneyActivity > 50 {
		trust += 15.0
	} else if smartMoneyActivity > 30 {
		trust += 10.0
	}
	
	// Normaliser entre 0-100
	return math.Max(0, math.Min(100, trust))
}

// calculateMarketDynamics calcule le facteur de dynamique de marché
func (e *Engine) calculateMarketDynamics(metrics *models.TokenMetrics) float64 {
	dynamics := 50.0 // Score de base
	
	// Facteurs basés sur le volume
	if metrics.Volume1h > 100000 {
		dynamics += 20.0
	} else if metrics.Volume1h > 50000 {
		dynamics += 15.0
	} else if metrics.Volume1h > 10000 {
		dynamics += 10.0
	}
	
	// Facteurs basés sur les variations de prix
	if metrics.PriceChange1h > 0.2 {
		dynamics += 15.0
	} else if metrics.PriceChange1h > 0.1 {
		dynamics += 10.0
	} else if metrics.PriceChange1h < -0.2 {
		dynamics -= 15.0
	} else if metrics.PriceChange1h < -0.1 {
		dynamics -= 10.0
	}
	
	// Ratio buy/sell count
	buySellRatio := 1.0
	if metrics.SellCount1h > 0 {
		buySellRatio = float64(metrics.BuyCount1h) / float64(metrics.SellCount1h)
	}
	
	if buySellRatio > 2.0 {
		dynamics += 15.0
	} else if buySellRatio > 1.5 {
		dynamics += 10.0
	} else if buySellRatio < 0.5 {
		dynamics -= 15.0
	} else if buySellRatio < 0.8 {
		dynamics -= 10.0
	}
	
	// Normaliser entre 0-100
	return math.Max(0, math.Min(100, dynamics))
}

// calculateTemporalPatterns calcule le facteur de patterns temporels
func (e *Engine) calculateTemporalPatterns(metrics *models.TokenMetrics) float64 {
	// Pour une implémentation complète, on analyserait les patterns temporels
	// des transactions sur plusieurs heures/jours
	
	// Simplification pour le moment
	return 60.0 // Score fixe pour démo
}

// calculateReactivationFactor calcule le facteur de réactivation
func (e *Engine) calculateReactivationFactor(token *models.Token, metrics *models.TokenMetrics) float64 {
	// Pour une implémentation complète, on vérifierait si le token montre des signes
	// de réactivation après une période d'inactivité
	
	// Simplification pour le moment
	return 0.0 // Par défaut pas de bonus réactivation
}

// checkAntiDumpPattern vérifie les patterns de dump coordonnés
func (e *Engine) checkAntiDumpPattern(tokenAddress string, walletAnalysis *models.WalletAnalysis) *models.AntiDumpResult {
	// Récupérer transactions récentes (24h)
	transactions, err := e.GetTokenRecentTrades(tokenAddress, 24)
	if err != nil || len(transactions) == 0 {
		return &models.AntiDumpResult{
			Detected: false,
			Severity: 0,
			Clusters: []models.DumpCluster{},
		}
	}
	
	// Filtrer ventes uniquement
	var sellTransactions []models.TokenTrade
	for _, tx := range transactions {
		if tx.TradeType == "sell" {
			sellTransactions = append(sellTransactions, tx)
		}
	}
	
	// Si peu de ventes, pas de pattern
	if len(sellTransactions) < 5 {
		return &models.AntiDumpResult{
			Detected: false,
			Severity: 0,
			Clusters: []models.DumpCluster{},
		}
	}
	
	// Détection des clusters temporels (ventes rapprochées)
	var clusters [][]models.TokenTrade
	var currentCluster []models.TokenTrade
	
	// Trier par timestamp
	// Note: simplification ici, normalement on trierait
	
	for i, tx := range sellTransactions {
		if i == 0 {
			currentCluster = append(currentCluster, tx)
			continue
		}
		
		lastTx := currentCluster[len(currentCluster)-1]
		timeDiff := tx.Timestamp.Sub(lastTx.Timestamp).Seconds()
		
		// Si vente dans fenêtre 5min, ajouter au cluster
		if timeDiff <= 300 { // 5 minutes
			currentCluster = append(currentCluster, tx)
		} else {
			// Enregistrer cluster si significatif (3+ ventes)
			if len(currentCluster) >= 3 {
				clusters = append(clusters, currentCluster)
			}
			currentCluster = []models.TokenTrade{tx}
		}
	}
	
	// Ajouter dernier cluster si significatif
	if len(currentCluster) >= 3 {
		clusters = append(clusters, currentCluster)
	}
	
	// Si pas de clusters significatifs
	if len(clusters) == 0 {
		return &models.AntiDumpResult{
			Detected: false,
			Severity: 0,
			Clusters: []models.DumpCluster{},
		}
	}
	
	// Analyser chaque cluster
	var analyzedClusters []models.DumpCluster
	highestSeverity := 0.0
	
	for _, cluster := range clusters {
		// Extraire wallets vendeurs uniques
		walletMap := make(map[string]struct{})
		for _, tx := range cluster {
			walletMap[tx.WalletAddress] = struct{}{}
		}
		uniqueWallets := len(walletMap)
		
		// Calculer volume total vendu
		totalVolume := 0.0
		for _, tx := range cluster {
			totalVolume += tx.TotalValue
		}
		
		// Vérifier si wallets smart sont impliqués
		smartSellerCount := 0
		if walletAnalysis != nil {
			smartWallets := make(map[string]struct{})
			for _, detail := range walletAnalysis.WalletDetails {
				for _, category := range detail.Categories {
					if category == "smart" {
						smartWallets[detail.Address] = struct{}{}
						break
					}
				}
			}
			
			for wallet := range walletMap {
				if _, ok := smartWallets[wallet]; ok {
					smartSellerCount++
				}
			}
		}
		
		// Calculer gravité du cluster
		var severity float64
		if smartSellerCount > 0 {
			// Plus grave si wallets smart impliqués
			severity = math.Min(100, float64(smartSellerCount*20) + (totalVolume/100))
		} else {
			// Moins grave si wallets non smart
			severity = math.Min(60, float64(uniqueWallets*10) + (totalVolume/200))
		}
		
		// Marquer le cluster
		clusterInfo := models.DumpCluster{
			TimestampStart:  cluster[0].Timestamp,
			TimestampEnd:    cluster[len(cluster)-1].Timestamp,
			DurationSeconds: cluster[len(cluster)-1].Timestamp.Sub(cluster[0].Timestamp).Seconds(),
			TransactionCount: len(cluster),
			UniqueWallets:   uniqueWallets,
			SmartWallets:    smartSellerCount,
			TotalVolume:     totalVolume,
			Severity:        severity,
		}
		
		analyzedClusters = append(analyzedClusters, clusterInfo)
		if severity > highestSeverity {
			highestSeverity = severity
		}
	}
	
	// Résultat final
	return &models.AntiDumpResult{
		Detected: highestSeverity >= 30, // Seuil de détection
		Severity: highestSeverity,
		Clusters: analyzedClusters,
	}
} 