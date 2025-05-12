package reactivation

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/franky69420/crypto-oracle/internal/token"
	"github.com/franky69420/crypto-oracle/internal/wallet"
	"github.com/franky69420/crypto-oracle/pkg/models"
	"github.com/sirupsen/logrus"
)

// System gère la détection des tokens réactivés
type System struct {
	tokenEngine *token.Engine
	walletEngine *wallet.Intelligence
	logger     *logrus.Logger
	running    bool
	interval   time.Duration
}

// NewSystem crée un nouveau système de détection de réactivation
func NewSystem(tokenEngine *token.Engine, walletEngine *wallet.Intelligence, logger *logrus.Logger) *System {
	return &System{
		tokenEngine: tokenEngine,
		walletEngine: walletEngine,
		logger:     logger,
		interval:   15 * time.Minute, // Intervalle par défaut pour vérifier les réactivations
	}
}

// Start démarre le système de réactivation
func (s *System) Start(ctx context.Context) error {
	s.logger.Info("Starting Reactivation System")
	s.running = true

	// Démarrer la routine de scan en arrière-plan
	go s.scanRoutine(ctx)

	return nil
}

// Shutdown arrête le système de réactivation
func (s *System) Shutdown(ctx context.Context) error {
	s.logger.Info("Shutting down Reactivation System")
	s.running = false
	return nil
}

// scanRoutine effectue le scan périodique des tokens dormants
func (s *System) scanRoutine(ctx context.Context) {
	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	for s.running {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			// Exécuter le scan
			candidates, err := s.ScanDormantTokens()
			if err != nil {
				s.logger.WithError(err).Error("Error scanning dormant tokens")
				continue
			}

			s.logger.WithField("count", len(candidates)).Info("Reactivation candidates found")

			// Traiter chaque candidat
			for _, candidate := range candidates {
				s.ProcessReactivationCandidate(candidate)
			}
		}
	}
}

// ScanDormantTokens scanne les tokens dormants pour détecter des signes de réactivation
func (s *System) ScanDormantTokens() ([]models.ReactivationCandidate, error) {
	// Récupérer tokens en SLEEP_MODE ou MONITORING_LIGHT
	dormantTokens, err := s.tokenEngine.GetTokensByStates([]string{
		models.LifecycleStateSleepMode,
		models.LifecycleStateMonitoringLight,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get dormant tokens: %w", err)
	}

	var candidates []models.ReactivationCandidate

	// Analyser chaque token dormant
	for _, token := range dormantTokens {
		// Récupérer métriques récentes
		metrics, err := s.tokenEngine.GetTokenMetrics(token.Address)
		if err != nil {
			s.logger.WithError(err).WithField("token", token.Address).
				Warn("Failed to get token metrics")
			continue
		}

		// Récupérer snapshot précédent pour comparaison
		previousMetrics, err := s.tokenEngine.GetTokenLastSnapshot(token.Address)
		if err != nil {
			s.logger.WithError(err).WithField("token", token.Address).
				Warn("Failed to get token last snapshot")
			// Continuer sans snapshot précédent
		}

		// Calculer les changements
		changes := s.calculateMetricChanges(metrics, previousMetrics)

		// Vérifier retour de wallets smart
		smartReturns, err := s.detectSmartWalletReturns(token.Address)
		if err != nil {
			s.logger.WithError(err).WithField("token", token.Address).
				Warn("Failed to detect smart wallet returns")
			// Continuer sans smart returns
		}

		// Calculer le score de réactivation
		reactivationScore := s.calculateReactivationScore(token, changes, smartReturns)

		// Si score suffisant, marquer comme candidat
		if reactivationScore >= 60 {
			candidate := models.ReactivationCandidate{
				TokenAddress:      token.Address,
				TokenSymbol:       token.Symbol,
				ReactivationScore: reactivationScore,
				Changes:           changes,
				SmartReturns:      smartReturns,
				CurrentMetrics:    metrics,
				DetectedAt:        time.Now(),
			}

			candidates = append(candidates, candidate)

			s.logger.WithFields(logrus.Fields{
				"token_address":      token.Address,
				"token_symbol":       token.Symbol,
				"reactivation_score": reactivationScore,
			}).Info("Reactivation candidate detected")
		}
	}

	return candidates, nil
}

// calculateMetricChanges calcule les changements entre les métriques actuelles et précédentes
func (s *System) calculateMetricChanges(current, previous *models.TokenMetrics) map[string]float64 {
	changes := make(map[string]float64)

	// Si pas de métriques précédentes, retourner des changements par défaut
	if previous == nil {
		return map[string]float64{
			"volume_1h_change": 0,
			"price_change":     0,
			"holder_growth":    0,
		}
	}

	// Calculer les changements
	if current.Volume1h > 0 && previous.Volume1h > 0 {
		changes["volume_1h_change"] = current.Volume1h / previous.Volume1h
	} else if current.Volume1h > 0 {
		changes["volume_1h_change"] = 10.0 // Valeur arbitraire élevée si volume précédent était 0
	} else {
		changes["volume_1h_change"] = 0
	}

	if current.Price > 0 && previous.Price > 0 {
		changes["price_change"] = (current.Price - previous.Price) / previous.Price
	} else {
		changes["price_change"] = 0
	}

	if current.HolderCount > 0 && previous.HolderCount > 0 {
		changes["holder_growth"] = float64(current.HolderCount-previous.HolderCount) / float64(previous.HolderCount)
	} else {
		changes["holder_growth"] = 0
	}

	return changes
}

// detectSmartWalletReturns détecte le retour de wallets smart sur un token dormant
func (s *System) detectSmartWalletReturns(tokenAddress string) (*models.SmartWalletReturns, error) {
	// Récupérer les transactions récentes (48h)
	recentTrades, err := s.tokenEngine.GetTokenRecentTrades(tokenAddress, 48)
	if err != nil {
		return nil, err
	}

	result := &models.SmartWalletReturns{
		Detected: false,
		Wallets:  []string{},
	}

	if len(recentTrades) == 0 {
		return result, nil
	}

	// Filtrer achats uniquement
	buyTrades := make([]models.TokenTrade, 0)
	for _, trade := range recentTrades {
		if trade.TradeType == "buy" {
			buyTrades = append(buyTrades, trade)
		}
	}

	// Collecter des adresses uniques d'acheteurs
	buyerMap := make(map[string]struct{})
	for _, trade := range buyTrades {
		buyerMap[trade.WalletAddress] = struct{}{}
	}

	// Convertir en slice
	buyers := make([]string, 0, len(buyerMap))
	for addr := range buyerMap {
		buyers = append(buyers, addr)
	}

	// Vérifier chaque wallet
	smartReturns := make([]string, 0)
	totalReturnVolume := 0.0

	for _, addr := range buyers {
		// Vérifier si le wallet est "smart"
		// Dans une implémentation réelle, on vérifierait avec le Memory of Trust
		isSmart, smartScore, err := s.walletEngine.IsSmartMoneyWallet(addr)
		if err != nil {
			s.logger.WithError(err).WithField("wallet", addr).
				Warn("Failed to check if wallet is smart")
			continue
		}

		if !isSmart || smartScore < 70 {
			continue
		}

		// Vérifier l'historique du wallet avec ce token
		history, err := s.tokenEngine.GetWalletTokenHistory(addr, tokenAddress)
		if err != nil || len(history) == 0 {
			continue
		}

		// Séparer les trades récents et anciens
		var recentBuys, pastSells []models.TokenTrade
		for _, trade := range history {
			if trade.TradeType == "sell" && trade.Timestamp.Before(time.Now().Add(-72*time.Hour)) {
				pastSells = append(pastSells, trade)
			} else if trade.TradeType == "buy" && trade.Timestamp.After(time.Now().Add(-48*time.Hour)) {
				recentBuys = append(recentBuys, trade)
			}
		}

		// Vérifier s'il y a eu des ventes dans le passé et des achats récents
		if len(pastSells) > 0 && len(recentBuys) > 0 {
			smartReturns = append(smartReturns, addr)
			
			// Calculer le volume de retour
			for _, buy := range recentBuys {
				totalReturnVolume += buy.TotalValue
			}
		}
	}

	// Définir les résultats
	result.Detected = len(smartReturns) >= 2 // Au moins 2 wallets smart qui reviennent
	result.Wallets = smartReturns
	result.ReturningTotalVolume = totalReturnVolume

	if result.Detected {
		// Trouver les timestamps
		result.ReturnTimestamp = time.Now().Add(-24 * time.Hour) // Approximatif
		result.InitialExitTimestamp = time.Now().Add(-7 * 24 * time.Hour) // Approximatif
		
		// Calculer la sévérité basée sur le nombre de wallets et le volume
		walletFactor := math.Min(1.0, float64(len(smartReturns))/5.0)
		volumeFactor := math.Min(1.0, totalReturnVolume/1000.0)
		result.Severity = (walletFactor*0.7 + volumeFactor*0.3) * 100
	}

	return result, nil
}

// calculateReactivationScore calcule un score de réactivation
func (s *System) calculateReactivationScore(token models.Token, changes map[string]float64, smartReturns *models.SmartWalletReturns) float64 {
	// Facteurs de réactivation
	volumeFactor := math.Min(1.0, changes["volume_1h_change"]/5.0)  // 5x max
	priceFactor := math.Min(1.0, changes["price_change"]/0.3)       // +30% max
	holdersFactor := math.Min(1.0, changes["holder_growth"]/0.1)    // +10% max
	
	// Score basé sur les métriques
	baseScore := (
		volumeFactor * 0.5 +
		priceFactor * 0.3 +
		holdersFactor * 0.2
	) * 100
	
	// Bonus pour smart wallet returns
	smartWalletBonus := 0.0
	if smartReturns != nil && smartReturns.Detected {
		// Bonus proportionnel au nombre de wallets qui reviennent
		returnCountFactor := math.Min(1.0, float64(len(smartReturns.Wallets))/5.0)  // Max pour 5 wallets
		
		// Bonus proportionnel au volume de retour
		volumeFactor := math.Min(1.0, smartReturns.ReturningTotalVolume/500.0)  // Max pour 500 SOL
		
		// Calcul bonus final
		smartWalletBonus = (
			returnCountFactor * 0.7 +
			volumeFactor * 0.3
		) * 30  // Max 30 points bonus
	}
	
	// Score final
	reactivationScore := baseScore + smartWalletBonus
	
	// Limiter à 0-100
	return math.Max(0, math.Min(100, reactivationScore))
}

// ProcessReactivationCandidate traite un candidat à la réactivation
func (s *System) ProcessReactivationCandidate(candidate models.ReactivationCandidate) error {
	// Mettre à jour l'état du token
	err := s.tokenEngine.UpdateTokenState(candidate.TokenAddress, models.LifecycleStateReactivated)
	if err != nil {
		return fmt.Errorf("failed to update token state: %w", err)
	}
	
	// Sauvegarder les métriques de réactivation
	err = s.tokenEngine.SaveReactivationMetrics(candidate)
	if err != nil {
		s.logger.WithError(err).WithField("token", candidate.TokenAddress).
			Error("Failed to save reactivation metrics")
		// Continuer malgré l'erreur
	}
	
	// Générer une alerte
	// Dans une implémentation réelle, on utiliserait le alertManager
	s.logger.WithFields(logrus.Fields{
		"token_address":      candidate.TokenAddress,
		"token_symbol":       candidate.TokenSymbol,
		"reactivation_score": candidate.ReactivationScore,
		"smart_wallets":      len(candidate.SmartReturns.Wallets),
	}).Info("Token reactivation detected and processed")
	
	return nil
} 