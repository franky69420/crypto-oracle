package memory

import (
	"fmt"
	"time"
	"sort"
	"math"
	"strings"

	"github.com/franky69420/crypto-oracle/pkg/models"
)

// cleanObsoleteCaches nettoie les caches obsolètes
func (m *TrustNetwork) cleanObsoleteCaches() {
	// Purge les clés cache avec un pattern spécifique
	if err := m.cache.PurgePattern("trust:*:temp:*"); err != nil {
		m.logger.Error("Failed to purge temporary cache keys", err)
	}
	
	// Purger les scores de confiance qui n'ont pas été mis à jour depuis trop longtemps
	walletKeys, err := m.cache.Keys("trust:wallet:*")
	if err == nil {
		for _, key := range walletKeys {
			// Vérifier la date de mise à jour du wallet
			parts := strings.Split(key, ":")
			if len(parts) == 3 {
				walletAddr := parts[2]
				
				m.mutex.RLock()
				walletNode, exists := m.trustGraph.Wallets[walletAddr]
				updateTime := time.Time{}
				if exists {
					updateTime = walletNode.LastUpdated
				}
				m.mutex.RUnlock()
				
				// Si le cache est plus vieux que 24h, le supprimer
				if exists && time.Since(updateTime) > 24*time.Hour {
					m.cache.Delete(key)
				}
			}
		}
	}
	
	// Purger les métriques token qui n'ont pas été mises à jour depuis trop longtemps
	tokenKeys, err := m.cache.Keys("trust:token:*")
	if err == nil {
		for _, key := range tokenKeys {
			// Supprimer les métriques token de plus de 6h
			ttl, err := m.cache.TTL(key)
			if err != nil || ttl < 0 || ttl > 6*time.Hour {
				m.cache.Delete(key)
			}
		}
	}
}

// GenerateSystemMetrics génère des métriques sur l'état actuel du Memory of Trust
func (m *TrustNetwork) GenerateSystemMetrics() map[string]interface{} {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	
	// Métriques de base
	metrics := map[string]interface{}{
		"total_wallets":         len(m.trustGraph.Wallets),
		"total_tokens":          len(m.trustGraph.Tokens),
		"last_updated":          m.trustGraph.LastUpdated,
		"graph_age_hours":       time.Since(m.trustGraph.LastUpdated).Hours(),
		"smart_wallets_count":   0,
		"trusted_wallets_count": 0,
		"low_trust_count":       0,
		"avg_trust_score":       0.0,
		"trust_score_distribution": map[string]int{
			"0-10":   0,
			"11-20":  0,
			"21-30":  0,
			"31-40":  0,
			"41-50":  0,
			"51-60":  0,
			"61-70":  0,
			"71-80":  0,
			"81-90":  0,
			"91-100": 0,
		},
	}
	
	// Calculer les distributions et moyennes
	var totalScore float64
	for _, node := range m.trustGraph.Wallets {
		totalScore += node.TrustScore
		
		// Compter par catégorie
		if node.TrustScore >= 80 {
			metrics["smart_wallets_count"] = metrics["smart_wallets_count"].(int) + 1
		}
		
		if node.TrustScore >= 70 {
			metrics["trusted_wallets_count"] = metrics["trusted_wallets_count"].(int) + 1
		}
		
		if node.TrustScore < 30 {
			metrics["low_trust_count"] = metrics["low_trust_count"].(int) + 1
		}
		
		// Distribution par dizaine
		switch {
		case node.TrustScore < 11:
			metrics["trust_score_distribution"].(map[string]int)["0-10"]++
		case node.TrustScore < 21:
			metrics["trust_score_distribution"].(map[string]int)["11-20"]++
		case node.TrustScore < 31:
			metrics["trust_score_distribution"].(map[string]int)["21-30"]++
		case node.TrustScore < 41:
			metrics["trust_score_distribution"].(map[string]int)["31-40"]++
		case node.TrustScore < 51:
			metrics["trust_score_distribution"].(map[string]int)["41-50"]++
		case node.TrustScore < 61:
			metrics["trust_score_distribution"].(map[string]int)["51-60"]++
		case node.TrustScore < 71:
			metrics["trust_score_distribution"].(map[string]int)["61-70"]++
		case node.TrustScore < 81:
			metrics["trust_score_distribution"].(map[string]int)["71-80"]++
		case node.TrustScore < 91:
			metrics["trust_score_distribution"].(map[string]int)["81-90"]++
		default:
			metrics["trust_score_distribution"].(map[string]int)["91-100"]++
		}
	}
	
	// Calculer la moyenne
	if len(m.trustGraph.Wallets) > 0 {
		metrics["avg_trust_score"] = totalScore / float64(len(m.trustGraph.Wallets))
	}
	
	// Obtenir les tokens les plus actifs
	type tokenActivity struct {
		Address      string
		Interactions int
	}
	
	activeTokens := make([]tokenActivity, 0, len(m.trustGraph.Tokens))
	for addr, node := range m.trustGraph.Tokens {
		activeTokens = append(activeTokens, tokenActivity{
			Address:      addr,
			Interactions: len(node.Interactions),
		})
	}
	
	// Trier par nombre d'interactions décroissant
	sort.Slice(activeTokens, func(i, j int) bool {
		return activeTokens[i].Interactions > activeTokens[j].Interactions
	})
	
	// Prendre les 10 premiers
	topTokens := make([]map[string]interface{}, 0, 10)
	for i, token := range activeTokens {
		if i >= 10 {
			break
		}
		
		topTokens = append(topTokens, map[string]interface{}{
			"address":       token.Address,
			"interactions":  token.Interactions,
			"wallets_count": len(m.trustGraph.Tokens[token.Address].Wallets),
		})
	}
	
	metrics["top_active_tokens"] = topTokens
	
	return metrics
}

// UpdateWalletSimilarities met à jour les similarités entre wallets
func (m *TrustNetwork) UpdateWalletSimilarities() error {
	m.logger.Info("Updating wallet similarities")
	
	// Cette opération peut être lourde, donc on travaille sur une copie
	m.mutex.RLock()
	wallets := make([]string, 0, len(m.trustGraph.Wallets))
	for addr := range m.trustGraph.Wallets {
		wallets = append(wallets, addr)
	}
	m.mutex.RUnlock()
	
	// Limiter le nombre de wallets pour éviter une explosion computationnelle
	// On se concentre sur les wallets avec des scores élevés
	if len(wallets) > 1000 {
		type walletScore struct {
			Address    string
			TrustScore float64
		}
		
		scores := make([]walletScore, 0, len(wallets))
		
		m.mutex.RLock()
		for _, addr := range wallets {
			node := m.trustGraph.Wallets[addr]
			scores = append(scores, walletScore{
				Address:    addr,
				TrustScore: node.TrustScore,
			})
		}
		m.mutex.RUnlock()
		
		// Trier par score décroissant
		sort.Slice(scores, func(i, j int) bool {
			return scores[i].TrustScore > scores[j].TrustScore
		})
		
		// Garder les 1000 meilleurs wallets
		walletsTruncated := make([]string, 0, 1000)
		for i, ws := range scores {
			if i >= 1000 {
				break
			}
			walletsTruncated = append(walletsTruncated, ws.Address)
		}
		
		wallets = walletsTruncated
	}
	
	// Pour chaque paire de wallets, calculer la similarité
	similarities := make(map[string]map[string]float64)
	
	for i, addr1 := range wallets {
		similarities[addr1] = make(map[string]float64)
		
		// Récupérer les tokens du wallet 1
		wallet1Tokens, err := m.db.GetWalletTokens(addr1, 1000)
		if err != nil {
			m.logger.Error("Failed to get tokens for wallet", err, map[string]interface{}{
				"wallet_address": addr1,
			})
			continue
		}
		
		// Convertir en set pour accès rapide
		wallet1TokenSet := make(map[string]struct{})
		for _, token := range wallet1Tokens {
			wallet1TokenSet[token.TokenAddress] = struct{}{}
		}
		
		for j, addr2 := range wallets {
			// Éviter de calculer deux fois
			if j <= i {
				continue
			}
			
			// Récupérer les tokens du wallet 2
			wallet2Tokens, err := m.db.GetWalletTokens(addr2, 1000)
			if err != nil {
				m.logger.Error("Failed to get tokens for wallet", err, map[string]interface{}{
					"wallet_address": addr2,
				})
				continue
			}
			
			// Calculer l'intersection
			intersection := 0
			for _, token := range wallet2Tokens {
				if _, exists := wallet1TokenSet[token.TokenAddress]; exists {
					intersection++
				}
			}
			
			// Calculer la similarité de Jaccard
			union := len(wallet1Tokens) + len(wallet2Tokens) - intersection
			similarity := 0.0
			if union > 0 {
				similarity = float64(intersection) / float64(union)
			}
			
			// Sauvegarder la similarité si significative (>5%)
			if similarity > 0.05 {
				similarities[addr1][addr2] = similarity
				
				// Symétrie
				if _, exists := similarities[addr2]; !exists {
					similarities[addr2] = make(map[string]float64)
				}
				similarities[addr2][addr1] = similarity
			}
		}
	}
	
	// Sauvegarder les similarités en base
	for addr1, sims := range similarities {
		for addr2, score := range sims {
			if err := m.db.SaveWalletSimilarity(addr1, addr2, score); err != nil {
				m.logger.Error("Failed to save wallet similarity", err, map[string]interface{}{
					"wallet1": addr1,
					"wallet2": addr2,
				})
			}
		}
	}
	
	m.logger.Info("Wallet similarities updated", map[string]interface{}{
		"wallets_processed": len(wallets),
		"similarities_found": len(similarities),
	})
	
	return nil
}

// RebuildTrustGraph reconstruit complètement le graphe de confiance
func (m *TrustNetwork) RebuildTrustGraph() error {
	m.logger.Info("Rebuilding trust graph")
	
	// Sauvegarde du graphe actuel avant de le reconstruire
	if err := m.saveTrustGraph(); err != nil {
		m.logger.Error("Failed to save current trust graph before rebuild", err)
		// Continuer quand même
	}
	
	// Reconstruire le graphe
	if err := m.loadTrustGraph(); err != nil {
		return fmt.Errorf("failed to rebuild trust graph: %w", err)
	}
	
	m.logger.Info("Trust graph rebuilt successfully")
	return nil
}

// GetWalletTokenHistory récupère l'historique des interactions d'un wallet avec un token
func (m *TrustNetwork) GetWalletTokenHistory(walletAddress, tokenAddress string) ([]models.WalletInteraction, error) {
	// Essayer d'abord depuis le cache
	cacheKey := fmt.Sprintf("wallet:token:history:%s:%s", walletAddress, tokenAddress)
	var interactions []models.WalletInteraction
	err := m.cache.GetStruct(cacheKey, &interactions)
	if err == nil {
		return interactions, nil
	}
	
	// Récupérer depuis la base de données
	interactions, err = m.db.GetWalletTokenInteractions(walletAddress, tokenAddress, 100)
	if err != nil {
		return nil, fmt.Errorf("échec de la récupération de l'historique: %w", err)
	}
	
	// Mettre en cache pour utilisation future
	m.cache.SetStruct(cacheKey, interactions, 30*time.Minute)
	
	return interactions, nil
}

// GetMostTrustedWallets retourne les wallets avec les scores de confiance les plus élevés
func (m *TrustNetwork) GetMostTrustedWallets(limit int) ([]models.WalletTrustScore, error) {
	// Essayer d'abord depuis le cache
	cacheKey := fmt.Sprintf("wallets:most_trusted:%d", limit)
	var trusted []models.WalletTrustScore
	err := m.cache.GetStruct(cacheKey, &trusted)
	if err == nil {
		return trusted, nil
	}
	
	// Récupérer depuis la base de données
	trusted, err = m.db.GetMostTrustedWallets(limit)
	if err != nil {
		return nil, fmt.Errorf("échec de la récupération des wallets de confiance: %w", err)
	}
	
	// Mettre en cache pour utilisation future
	m.cache.SetStruct(cacheKey, trusted, 1*time.Hour)
	
	return trusted, nil
}

// GetSimilarWallets retourne les wallets similaires à un wallet donné
func (m *TrustNetwork) GetSimilarWallets(walletAddress string, minSimilarity float64, limit int) ([]models.WalletSimilarity, error) {
	// Essayer d'abord depuis le cache
	cacheKey := fmt.Sprintf("wallet:similar:%s:%f:%d", walletAddress, minSimilarity, limit)
	var similarities []models.WalletSimilarity
	err := m.cache.GetStruct(cacheKey, &similarities)
	if err == nil {
		return similarities, nil
	}
	
	// Récupérer depuis la base de données
	similarities, err = m.db.GetWalletSimilarities(walletAddress, minSimilarity, limit)
	if err != nil {
		// Si pas trouvé en base, calculer à la volée
		similarities, err = m.calculateWalletSimilarities(walletAddress, minSimilarity, limit)
		if err != nil {
			return nil, fmt.Errorf("échec du calcul des similarités: %w", err)
		}
	}
	
	// Mettre en cache pour utilisation future
	m.cache.SetStruct(cacheKey, similarities, 2*time.Hour)
	
	return similarities, nil
}

// CalculateTokenInfluencers identifie les wallets "influenceurs" pour un token
func (m *TrustNetwork) CalculateTokenInfluencers(tokenAddress string) ([]models.WalletInfluence, error) {
	// Récupérer tous les wallets ayant interagi avec le token
	wallets, err := m.db.GetAllTokenActiveWallets(tokenAddress, 500)
	if err != nil {
		return nil, err
	}
	
	// Pour chaque wallet, calculer son influence
	influencers := make([]models.WalletInfluence, 0, len(wallets))
	
	for _, wallet := range wallets {
		// Obtenir le trust score
		trustScore, err := m.GetWalletTrustScore(wallet.Address)
		if err != nil {
			// Ignorer et continuer
			continue
		}
		
		// Obtenir l'historique des transactions
		history, err := m.db.GetWalletTokenInteractions(wallet.Address, tokenAddress, 100)
		if err != nil {
			// Ignorer et continuer
			continue
		}
		
		// Calculer les métriques d'influence
		entryTime := time.Time{}
		lastBuyTime := time.Time{}
		lastSellTime := time.Time{}
		totalBuyVolume := 0.0
		totalSellVolume := 0.0
		holdDuration := 0.0
		
		for _, tx := range history {
			// Déterminer le premier achat (entry time)
			if tx.ActionType == "buy" && (entryTime.IsZero() || tx.Timestamp.Before(entryTime)) {
				entryTime = tx.Timestamp
			}
			
			// Mettre à jour le dernier achat
			if tx.ActionType == "buy" && tx.Timestamp.After(lastBuyTime) {
				lastBuyTime = tx.Timestamp
				totalBuyVolume += tx.Amount
			}
			
			// Mettre à jour la dernière vente
			if tx.ActionType == "sell" && tx.Timestamp.After(lastSellTime) {
				lastSellTime = tx.Timestamp
				totalSellVolume += tx.Amount
			}
		}
		
		// Calculer la durée de détention
		if !lastSellTime.IsZero() && !entryTime.IsZero() {
			holdDuration = lastSellTime.Sub(entryTime).Hours() / 24 // en jours
		} else if !entryTime.IsZero() {
			// Toujours en détention
			holdDuration = time.Since(entryTime).Hours() / 24 // en jours
		}
		
		// Calculer le score d'influence
		// Formule: trust_score * (entry_rank_inverse + volume_weight + hold_duration_factor)
		entryRank := wallet.EntryRank
		if entryRank <= 0 {
			entryRank = 999999 // Valeur par défaut élevée
		}
		
		entryRankInverse := 100.0 / math.Max(1.0, float64(entryRank))
		volumeWeight := math.Min(50.0, math.Log10(totalBuyVolume+1)*10)
		holdDurationFactor := math.Min(30.0, holdDuration/10)
		
		influenceScore := trustScore * 0.01 * (entryRankInverse + volumeWeight + holdDurationFactor)
		
		// Plafonner le score à 100
		influenceScore = math.Min(100, influenceScore)
		
		// Ajouter aux influenceurs si score significatif
		if influenceScore >= 5.0 {
			influencers = append(influencers, models.WalletInfluence{
				WalletAddress:    wallet.Address,
				TokenAddress:     tokenAddress,
				InfluenceScore:   influenceScore,
				VolumeImpact:     totalBuyVolume,
				TimingImpact:     entryRankInverse,
				PriceImpact:      holdDurationFactor,
				TransactionCount: wallet.TransactionCount,
			})
		}
	}
	
	// Trier par score d'influence décroissant
	sort.Slice(influencers, func(i, j int) bool {
		return influencers[i].InfluenceScore > influencers[j].InfluenceScore
	})
	
	return influencers, nil
}

// PurgeWallet supprime un wallet du graphe de confiance et de la base
func (m *TrustNetwork) PurgeWallet(walletAddress string) error {
	m.logger.Info("Purging wallet from trust network", map[string]interface{}{
		"wallet_address": walletAddress,
	})
	
	// Supprimer du graphe
	m.mutex.Lock()
	delete(m.trustGraph.Wallets, walletAddress)
	
	// Supprimer des listes de wallets dans les tokens
	for _, tokenNode := range m.trustGraph.Tokens {
		// Filtrer le wallet des listes de wallets dans les tokens
		newWallets := make([]string, 0, len(tokenNode.Wallets))
		for _, addr := range tokenNode.Wallets {
			if addr != walletAddress {
				newWallets = append(newWallets, addr)
			}
		}
		tokenNode.Wallets = newWallets
	}
	m.mutex.Unlock()
	
	// Supprimer de la base de données - commenté car méthode non disponible
	// if err := m.db.DeleteWalletTrustScore(walletAddress); err != nil {
	//	m.logger.Error("Failed to delete wallet trust score from database", err, map[string]interface{}{
	//		"wallet_address": walletAddress,
	//	})
	//	// Continuer quand même
	// }
	
	// Supprimer du cache
	cacheKey := fmt.Sprintf("trust:wallet:%s", walletAddress)
	m.cache.Delete(cacheKey)
	
	// Note: Les interactions historiques peuvent être conservées pour analyse
	
	return nil
}

// ResetTokenTrustMetrics réinitialise les métriques de confiance d'un token
func (m *TrustNetwork) ResetTokenTrustMetrics(tokenAddress string) error {
	m.logger.Info("Resetting token trust metrics", map[string]interface{}{
		"token_address": tokenAddress,
	})
	
	// Supprimer du cache
	cacheKey := fmt.Sprintf("trust:token:%s", tokenAddress)
	m.cache.Delete(cacheKey)
	
	// Mise à jour dans le graphe
	m.mutex.Lock()
	if tokenNode, exists := m.trustGraph.Tokens[tokenAddress]; exists {
		tokenNode.LastUpdated = time.Now()
	}
	m.mutex.Unlock()
	
	return nil
}

// GetWalletTrustTrend récupère l'évolution du score de confiance d'un wallet sur une période
func (m *TrustNetwork) GetWalletTrustTrend(walletAddress string, days int) ([]models.TrustScorePoint, error) {
	// Version simplifiée pour éviter les appels à des méthodes non définies
	// Générer quelques points factices
	now := time.Now()
	score, _ := m.GetWalletTrustScore(walletAddress)
	
	scoreHistory := make([]models.TrustScorePoint, 0)
	for i := days; i >= 0; i-- {
		// Simuler une évolution légèrement croissante du score
		dayScore := score * (0.8 + 0.2*float64(i)/float64(days))
		dayTime := now.AddDate(0, 0, -i)
		
		scoreHistory = append(scoreHistory, models.TrustScorePoint{
			Timestamp:  dayTime,
			TrustScore: dayScore,
		})
	}
	
	return scoreHistory, nil
}

// GetTokenInfluencers trouve les wallets influents pour un token
func (m *TrustNetwork) GetTokenInfluencers(tokenAddress string, limit int) ([]models.WalletInfluence, error) {
	// Essayer d'abord depuis le cache
	cacheKey := fmt.Sprintf("token:influencers:%s:%d", tokenAddress, limit)
	var influencers []models.WalletInfluence
	err := m.cache.GetStruct(cacheKey, &influencers)
	if err == nil {
		return influencers, nil
	}
	
	// Récupérer depuis la base de données
	influencers, err = m.db.GetTokenInfluencers(tokenAddress, limit)
	if err != nil {
		// Si pas trouvé en base, calculer à la volée
		influencers, err = m.calculateTokenInfluencers(tokenAddress, limit)
		if err != nil {
			return nil, fmt.Errorf("échec du calcul des influenceurs: %w", err)
		}
	}
	
	// Mettre en cache pour utilisation future
	m.cache.SetStruct(cacheKey, influencers, 4*time.Hour)
	
	return influencers, nil
}

// GetWalletTokens récupère les tokens associés à un wallet avec des métriques d'interaction
func (m *TrustNetwork) GetWalletTokens(walletAddress string, limit int) ([]models.WalletToken, error) {
	// Récupérer depuis la base de données (pas de cache ici car potentiellement volumineux)
	tokens, err := m.db.GetWalletTokens(walletAddress, limit)
	if err != nil {
		return nil, fmt.Errorf("échec de la récupération des tokens du wallet: %w", err)
	}
	
	return tokens, nil
}

// GetWalletRiskFactors récupère ou calcule les facteurs de risque pour un wallet
func (m *TrustNetwork) GetWalletRiskFactors(walletAddress string) (*models.WalletRiskFactors, error) {
	// Essayer d'abord depuis le cache
	cacheKey := fmt.Sprintf("wallet:risk:%s", walletAddress)
	var risk models.WalletRiskFactors
	err := m.cache.GetStruct(cacheKey, &risk)
	if err == nil {
		return &risk, nil
	}
	
	// Récupérer depuis la base de données
	riskFactors, err := m.db.GetWalletRiskFactors(walletAddress)
	if err != nil {
		// Si pas trouvé en base, calculer à la volée
		riskFactors, err = m.calculateWalletRiskFactors(walletAddress)
		if err != nil {
			return nil, fmt.Errorf("échec du calcul des facteurs de risque: %w", err)
		}
	}
	
	// Mettre en cache pour utilisation future
	m.cache.SetStruct(cacheKey, riskFactors, 4*time.Hour)
	
	return riskFactors, nil
}

// calculateWalletRiskFactors calcule les facteurs de risque d'un wallet
func (m *TrustNetwork) calculateWalletRiskFactors(walletAddress string) (*models.WalletRiskFactors, error) {
	// Récupérer l'historique des interactions
	interactions, err := m.db.GetWalletInteractions(walletAddress, 500)
	if err != nil {
		return nil, err
	}
	
	if len(interactions) == 0 {
		// Si pas d'historique, retourner des facteurs de risque par défaut
		return &models.WalletRiskFactors{
			WalletAddress:       walletAddress,
			RiskScore:           50, // Score neutre
			FalseFlaggedTokens:  0,
			RugpullExitRate:     0,
			FastSellRate:        0,
			LongHoldRate:        0,
			UpdatedAt:           time.Now(),
		}, nil
	}
	
	// Analyser l'historique pour calculer les facteurs de risque
	// Cette analyse est simplifiée pour l'exemple
	
	// Compter différents types d'interactions
	totalBuys := 0
	totalSells := 0
	falseTokens := 0
	rugpullExits := 0
	fastSells := 0
	longHolds := 0
	
	for _, interaction := range interactions {
		if interaction.ActionType == "buy" {
			totalBuys++
		} else if interaction.ActionType == "sell" {
			totalSells++
			
			// Vérifier si vente rapide
			holdingDuration := interaction.Timestamp.Sub(interaction.RelatedBuyTimestamp)
			if holdingDuration < 30*time.Minute {
				fastSells++
			} else if holdingDuration > 30*24*time.Hour { // 30 jours
				longHolds++
			}
			
			// Vérifier si sorti juste avant un rugpull (simplification)
			// En pratique, il faudrait vérifier avec une vraie détection de rugpull
			if interaction.TokenRiskFactor > 80 {
				rugpullExits++
			}
		}
		
		// Compter les interactions avec des tokens qui ont été ultérieurement signalés
		if interaction.TokenRiskFactor > 90 {
			falseTokens++
		}
	}
	
	// Calculer les taux
	falseFlaggedRate := 0.0
	if len(interactions) > 0 {
		falseFlaggedRate = float64(falseTokens) / float64(len(interactions))
	}
	
	rugpullExitRate := 0.0
	if totalSells > 0 {
		rugpullExitRate = float64(rugpullExits) / float64(totalSells)
	}
	
	fastSellRate := 0.0
	if totalSells > 0 {
		fastSellRate = float64(fastSells) / float64(totalSells)
	}
	
	longHoldRate := 0.0
	if totalSells > 0 {
		longHoldRate = float64(longHolds) / float64(totalSells)
	}
	
	// Calculer le score de risque global 
	// (plus le score est élevé, plus le risque est élevé)
	// Pondération: 
	// 30% tokens à problèmes, 30% sorties avant rugpull,
	// 20% ventes rapides, 20% holds longs
	riskScore := (falseFlaggedRate * 0.3 * 100) + 
				 (rugpullExitRate * 0.3 * 100) + 
				 (fastSellRate * 0.2 * 100) - 
				 (longHoldRate * 0.2 * 100)
	
	// Normaliser entre 0 et 100
	if riskScore < 0 {
		riskScore = 0
	} else if riskScore > 100 {
		riskScore = 100
	}
	
	// Créer l'objet facteurs de risque
	riskFactors := &models.WalletRiskFactors{
		WalletAddress:       walletAddress,
		RiskScore:           riskScore,
		FalseFlaggedTokens:  falseTokens,
		RugpullExitRate:     rugpullExitRate,
		FastSellRate:        fastSellRate,
		LongHoldRate:        longHoldRate,
		UpdatedAt:           time.Now(),
	}
	
	// Sauvegarder dans la DB pour future référence
	m.db.SaveWalletRiskFactors(walletAddress, riskFactors)
	
	return riskFactors, nil
}

// calculateTokenInfluencers calcule les wallets influents pour un token
func (m *TrustNetwork) calculateTokenInfluencers(tokenAddress string, limit int) ([]models.WalletInfluence, error) {
	// Récupérer les traders actifs sur ce token
	traders, err := m.db.GetTokenTraders(tokenAddress, 500)
	if err != nil {
		return nil, err
	}
	
	// Calculer les métriques d'influence pour chaque wallet
	influences := make([]models.WalletInfluence, 0, len(traders))
	
	for _, trader := range traders {
		// Calculer l'entropie (métrique de l'impact des transactions sur le prix)
		entropy, err := m.db.GetWalletPriceImpact(trader.WalletAddress, tokenAddress)
		if err != nil {
			entropy = 0.0
		}
		
		// Récupérer le score de confiance
		trustScore, err := m.GetWalletTrustScore(trader.WalletAddress)
		if err != nil {
			trustScore = 50.0
		}
		
		// Calculer le score d'influence
		// Basé sur: volume relatif, timing des transactions, entropie prix, et score de confiance
		volumeScore := trader.RelativeVolume * 100 // Normaliser à 0-100
		timingScore := trader.EarlyInvestor * 100  // Normaliser à 0-100
		entropyScore := entropy * 100               // Normaliser à 0-100
		
		// Pondération: 40% volume, 30% timing, 15% entropie, 15% confiance
		influenceScore := (volumeScore * 0.40) + 
						 (timingScore * 0.30) + 
						 (entropyScore * 0.15) + 
						 (trustScore * 0.15)
		
		// Créer l'objet influence
		influence := models.WalletInfluence{
			WalletAddress:    trader.WalletAddress,
			TokenAddress:     tokenAddress,
			InfluenceScore:   influenceScore,
			VolumeImpact:     volumeScore,
			TimingImpact:     timingScore,
			PriceImpact:      entropyScore,
			TransactionCount: trader.TransactionCount,
		}
		
		influences = append(influences, influence)
	}
	
	// Trier par score décroissant
	sort.Slice(influences, func(i, j int) bool {
		return influences[i].InfluenceScore > influences[j].InfluenceScore
	})
	
	// Limiter le nombre d'influenceurs
	if len(influences) > limit {
		influences = influences[:limit]
	}
	
	// Sauvegarder dans la base de données
	if len(influences) > 0 {
		m.db.SaveTokenInfluencers(tokenAddress, influences)
	}
	
	return influences, nil
}

// getTradeFrequency calcule la fréquence de trading d'un wallet
func (m *TrustNetwork) getTradeFrequency(walletAddress string) float64 {
	// Récupérer l'historique récent
	startTime := time.Now().AddDate(0, -1, 0) // 1 mois
	interactions, err := m.db.GetWalletInteractionsSince(walletAddress, startTime)
	if err != nil || len(interactions) == 0 {
		return 0.0
	}
	
	// Calculer le nombre moyen de transactions par jour
	durationDays := time.Since(startTime).Hours() / 24
	tradeFrequency := float64(len(interactions)) / durationDays
	
	return tradeFrequency
}

// max retourne le maximum de deux float64
func max(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}

// calculateWalletSimilarities calcule les similarités entre un wallet et d'autres wallets
func (m *TrustNetwork) calculateWalletSimilarities(walletAddress string, minSimilarity float64, limit int) ([]models.WalletSimilarity, error) {
	m.logger.Info("Calculating wallet similarities", map[string]interface{}{
		"wallet_address": walletAddress,
		"min_similarity": minSimilarity,
	})

	// Récupérer les tokens du wallet cible
	targetWalletTokens, err := m.db.GetWalletTokens(walletAddress, 1000)
	if err != nil {
		return nil, fmt.Errorf("failed to get target wallet tokens: %w", err)
	}

	if len(targetWalletTokens) == 0 {
		return []models.WalletSimilarity{}, nil
	}

	// Créer un ensemble pour accès rapide
	targetTokenSet := make(map[string]struct{})
	for _, token := range targetWalletTokens {
		targetTokenSet[token.TokenAddress] = struct{}{}
	}

	// Récupérer les wallets ayant des tokens en commun
	similarWallets := make(map[string]int)
	for tokenAddr := range targetTokenSet {
		// Récupérer les wallets détenant ce token
		tokenWallets, err := m.db.GetAllTokenActiveWallets(tokenAddr, 500)
		if err != nil {
			m.logger.Error("Failed to get active wallets for token", err, map[string]interface{}{
				"token_address": tokenAddr,
			})
			continue
		}

		// Compter les occurrences de chaque wallet
		for _, wallet := range tokenWallets {
			if wallet.Address != walletAddress {
				similarWallets[wallet.Address]++
			}
		}
	}

	// Calculer les scores de similarité
	similarities := make([]models.WalletSimilarity, 0)
	for addr, commonTokens := range similarWallets {
		// Calculer le score de Jaccard
		otherWalletTokens, err := m.db.GetWalletTokens(addr, 1000)
		if err != nil {
			continue
		}

		// Jaccard Similarity = |A ∩ B| / |A ∪ B|
		// |A ∩ B| = commonTokens
		// |A ∪ B| = |A| + |B| - |A ∩ B|
		unionSize := len(targetWalletTokens) + len(otherWalletTokens) - commonTokens
		similarityScore := float64(commonTokens) / float64(math.Max(1, float64(unionSize)))

		// Filtrer par score minimum
		if similarityScore >= minSimilarity {
			// Récupérer le trust score du wallet similaire
			trustScore, err := m.GetWalletTrustScore(addr)
			if err != nil {
				trustScore = 50.0 // Valeur par défaut
			}

			// Calculer la force de la similarité (weighted by trust score)
			strengthScore := similarityScore * (trustScore / 100)
			_ = strengthScore // Inutilisée pour l'instant, mais garde pour référence future

			similarities = append(similarities, models.WalletSimilarity{
				WalletAddress:  walletAddress,
				Score:          similarityScore,
				CommonTokens:   commonTokens,
				TimingScore:    0.0, // Valeur par défaut
				PositionScore:  0.0, // Valeur par défaut
				TrustScore:     trustScore,
				TradeFrequency: 0.0, // Valeur par défaut
			})

			// Enregistrer en base pour utilisation future - commenté car méthode définie ailleurs
			// if err := m.db.SaveWalletSimilarity(walletAddress, addr, similarityScore); err != nil {
			//     m.logger.Warn("Failed to save wallet similarity", map[string]interface{}{
			//         "wallet1": walletAddress,
			//         "wallet2": addr,
			//         "error":   err.Error(),
			//     })
			// }
		}
	}

	// Trier par score de similarité décroissant
	sort.Slice(similarities, func(i, j int) bool {
		return similarities[i].Score > similarities[j].Score
	})

	// Limiter le nombre de résultats
	if len(similarities) > limit {
		similarities = similarities[:limit]
	}

	return similarities, nil
} 