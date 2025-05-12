package memory

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/franky69420/crypto-oracle/internal/storage/cache"
	"github.com/franky69420/crypto-oracle/internal/storage/db"
	"github.com/franky69420/crypto-oracle/pkg/models"
	"github.com/franky69420/crypto-oracle/pkg/utils/logger"
)

// Vérification statique que TrustNetwork implémente l'interface MemoryOfTrust
var _ MemoryOfTrust = (*TrustNetwork)(nil)

// TrustNetwork implémente le Memory of Trust, un système qui mémorise et évalue 
// la fiabilité des wallets basée sur leur historique d'interactions
type TrustNetwork struct {
	db          *db.Connection
	cache       *cache.Client
	logger      *logger.Logger
	trustGraph  *TrustGraph
	mutex       sync.RWMutex
	maintenanceInterval time.Duration
	stopCh      chan struct{}
	wg          sync.WaitGroup
}

// TrustGraph est la structure de données principale maintenant les relations entre wallets et tokens
type TrustGraph struct {
	Wallets     map[string]*WalletNode
	Tokens      map[string]*TokenNode
	LastUpdated time.Time
}

// WalletNode représente un nœud wallet dans le graphe de confiance
type WalletNode struct {
	Address      string
	TrustScore   float64
	Interactions []string // IDs d'interactions
	LastUpdated  time.Time
}

// TokenNode représente un nœud token dans le graphe de confiance
type TokenNode struct {
	Address      string
	Wallets      []string // Adresses des wallets
	Interactions []string // IDs d'interactions
	LastUpdated  time.Time
}

// NewTrustNetwork crée une nouvelle instance du réseau de confiance
func NewTrustNetwork(db *db.Connection, cache *cache.Client, logger *logger.Logger) *TrustNetwork {
	return &TrustNetwork{
		db:          db,
		cache:       cache,
		logger:      logger,
		trustGraph: &TrustGraph{
			Wallets:     make(map[string]*WalletNode),
			Tokens:      make(map[string]*TokenNode),
			LastUpdated: time.Now(),
		},
		maintenanceInterval: 6 * time.Hour, // Par défaut, maintenance toutes les 6 heures
		stopCh:              make(chan struct{}),
	}
}

// Start initialise et démarre le Memory of Trust
func (m *TrustNetwork) Start(ctx context.Context) error {
	m.logger.Info("Démarrage du Memory of Trust")
	
	// Charger les données du graphe de confiance depuis la base de données
	if err := m.loadTrustGraph(); err != nil {
		m.logger.Error("Échec du chargement du graphe de confiance", err)
		// Continuer avec un graphe vide plutôt que d'échouer complètement
	}
	
	// Démarrer la goroutine de maintenance
	m.wg.Add(1)
	go m.maintenanceRoutine(ctx)
	
	m.logger.Info("Memory of Trust démarré avec succès")
	return nil
}

// Stop arrête le Memory of Trust et libère les ressources
func (m *TrustNetwork) Stop() error {
	m.logger.Info("Arrêt du Memory of Trust")
	
	// Signaler l'arrêt aux goroutines
	close(m.stopCh)
	
	// Attendre la fin des goroutines
	m.wg.Wait()
	
	// Sauvegarder l'état du graphe de confiance
	if err := m.saveTrustGraph(); err != nil {
		m.logger.Error("Échec de la sauvegarde du graphe de confiance", err)
		// Continuer malgré l'erreur pour finaliser l'arrêt proprement
	}
	
	m.logger.Info("Memory of Trust arrêté avec succès")
	return nil
}

// loadTrustGraph charge le graphe de confiance depuis la base de données
func (m *TrustNetwork) loadTrustGraph() error {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	
	m.logger.Info("Chargement du graphe de confiance")
	
	// Récupérer les scores de confiance des wallets
	walletScores, err := m.db.GetAllWalletTrustScores()
	if err != nil {
		return fmt.Errorf("échec de la récupération des scores de confiance: %w", err)
	}
	
	// Ajouter les wallets au graphe
	for _, ws := range walletScores {
		m.trustGraph.Wallets[ws.Address] = &WalletNode{
			Address:      ws.Address,
			TrustScore:   ws.TrustScore,
			Interactions: []string{},
			LastUpdated:  ws.LastUpdated,
		}
	}
	
	// Récupérer les interactions récentes pour enrichir le graphe
	// Limiter à un nombre raisonnable pour éviter une surcharge de mémoire
	interactions, err := m.db.GetRecentInteractions(10000) // 10k interactions récentes
	if err != nil {
		return fmt.Errorf("échec de la récupération des interactions récentes: %w", err)
	}
	
	// Traiter les interactions pour construire le graphe
	for _, interaction := range interactions {
		// Ajouter le wallet s'il n'existe pas déjà
		if _, exists := m.trustGraph.Wallets[interaction.WalletAddress]; !exists {
			// Récupérer le score de confiance ou utiliser une valeur par défaut
			trustScore, err := m.db.GetWalletTrustScore(interaction.WalletAddress)
			if err != nil {
				trustScore = 50.0 // Score par défaut
			}
			
			m.trustGraph.Wallets[interaction.WalletAddress] = &WalletNode{
				Address:      interaction.WalletAddress,
				TrustScore:   trustScore,
				Interactions: []string{},
				LastUpdated:  time.Now(),
			}
		}
		
		// Ajouter le token s'il n'existe pas déjà
		if _, exists := m.trustGraph.Tokens[interaction.TokenAddress]; !exists {
			m.trustGraph.Tokens[interaction.TokenAddress] = &TokenNode{
				Address:      interaction.TokenAddress,
				Wallets:      []string{},
				Interactions: []string{},
				LastUpdated:  time.Now(),
			}
		}
		
		// Ajouter les relations
		interactionID := fmt.Sprintf("%s:%s:%s", interaction.TxHash, interaction.WalletAddress, interaction.TokenAddress)
		
		// Ajouter l'interaction au wallet
		walletNode := m.trustGraph.Wallets[interaction.WalletAddress]
		walletNode.Interactions = append(walletNode.Interactions, interactionID)
		
		// Ajouter le wallet au token s'il n'y est pas déjà
		tokenNode := m.trustGraph.Tokens[interaction.TokenAddress]
		// Vérifier si le wallet existe déjà dans la liste du token
		walletExists := false
		for _, addr := range tokenNode.Wallets {
			if addr == interaction.WalletAddress {
				walletExists = true
				break
			}
		}
		if !walletExists {
			tokenNode.Wallets = append(tokenNode.Wallets, interaction.WalletAddress)
		}
		
		// Ajouter l'interaction au token
		tokenNode.Interactions = append(tokenNode.Interactions, interactionID)
	}
	
	m.trustGraph.LastUpdated = time.Now()
	m.logger.Info("Graphe de confiance chargé avec succès", map[string]interface{}{
		"wallets_count": len(m.trustGraph.Wallets),
		"tokens_count":  len(m.trustGraph.Tokens),
	})
	
	return nil
}

// saveTrustGraph sauvegarde le graphe de confiance dans la base de données
func (m *TrustNetwork) saveTrustGraph() error {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	
	m.logger.Info("Sauvegarde des scores de confiance des wallets")
	
	// Sauvegarder uniquement les scores de confiance des wallets
	// Les relations sont déjà sauvegardées via les interactions
	for addr, wallet := range m.trustGraph.Wallets {
		err := m.db.SaveWalletTrustScore(addr, wallet.TrustScore, wallet.LastUpdated)
		if err != nil {
			m.logger.Error("Échec de la sauvegarde du score de confiance", err, map[string]interface{}{
				"wallet_address": addr,
			})
			// Continuer malgré l'erreur pour sauvegarder les autres wallets
		}
	}
	
	return nil
}

// RecordWalletInteraction enregistre une interaction wallet-token et met à jour le graphe
func (m *TrustNetwork) RecordWalletInteraction(interaction *models.WalletInteraction) error {
	// Vérifier l'intégrité des données
	if interaction.WalletAddress == "" || interaction.TokenAddress == "" {
		return fmt.Errorf("adresse wallet ou token manquante")
	}
	
	// Sauvegarder l'interaction dans la base de données
	if err := m.db.SaveWalletInteraction(interaction); err != nil {
		return fmt.Errorf("échec de l'enregistrement de l'interaction: %w", err)
	}
	
	// Mettre à jour le graphe de confiance
	m.mutex.Lock()
	defer m.mutex.Unlock()
	
	// Ajouter le wallet s'il n'existe pas déjà
	if _, exists := m.trustGraph.Wallets[interaction.WalletAddress]; !exists {
		// Récupérer le score de confiance ou utiliser une valeur par défaut
		trustScore, err := m.db.GetWalletTrustScore(interaction.WalletAddress)
		if err != nil {
			trustScore = 50.0 // Score par défaut
		}
		
		m.trustGraph.Wallets[interaction.WalletAddress] = &WalletNode{
			Address:      interaction.WalletAddress,
			TrustScore:   trustScore,
			Interactions: []string{},
			LastUpdated:  time.Now(),
		}
	}
	
	// Ajouter le token s'il n'existe pas déjà
	if _, exists := m.trustGraph.Tokens[interaction.TokenAddress]; !exists {
		m.trustGraph.Tokens[interaction.TokenAddress] = &TokenNode{
			Address:      interaction.TokenAddress,
			Wallets:      []string{},
			Interactions: []string{},
			LastUpdated:  time.Now(),
		}
	}
	
	// Ajouter les relations
	interactionID := fmt.Sprintf("%s:%s:%s", interaction.TxHash, interaction.WalletAddress, interaction.TokenAddress)
	
	// Ajouter l'interaction au wallet
	walletNode := m.trustGraph.Wallets[interaction.WalletAddress]
	walletNode.Interactions = append(walletNode.Interactions, interactionID)
	
	// Ajouter le wallet au token s'il n'y est pas déjà
	tokenNode := m.trustGraph.Tokens[interaction.TokenAddress]
	// Vérifier si le wallet existe déjà dans la liste du token
	walletExists := false
	for _, addr := range tokenNode.Wallets {
		if addr == interaction.WalletAddress {
			walletExists = true
			break
		}
	}
	if !walletExists {
		tokenNode.Wallets = append(tokenNode.Wallets, interaction.WalletAddress)
	}
	
	// Ajouter l'interaction au token
	tokenNode.Interactions = append(tokenNode.Interactions, interactionID)
	
	// Mettre à jour le score de confiance du wallet
	go m.updateWalletTrustScore(interaction.WalletAddress)
	
	return nil
}

// GetWalletTrustScore récupère le score de confiance d'un wallet
func (m *TrustNetwork) GetWalletTrustScore(walletAddress string) (float64, error) {
	// Essayer d'abord de récupérer depuis le cache
	cacheKey := fmt.Sprintf("wallet:trust:%s", walletAddress)
	score, err := m.cache.GetFloat64(cacheKey)
	if err == nil {
		// Score trouvé dans le cache
		return score, nil
	}
	
	// Ensuite, essayer de récupérer depuis le graphe de confiance en mémoire
	m.mutex.RLock()
	if wallet, exists := m.trustGraph.Wallets[walletAddress]; exists {
		score := wallet.TrustScore
		m.mutex.RUnlock()
		
		// Mettre en cache pour une utilisation future
		m.cache.SetFloat64(cacheKey, score, 1*time.Hour)
		
		return score, nil
	}
	m.mutex.RUnlock()
	
	// Enfin, essayer de récupérer depuis la base de données
	score, err = m.db.GetWalletTrustScore(walletAddress)
	if err != nil {
		// Si non trouvé, calculer le score et le sauvegarder
		m.updateWalletTrustScore(walletAddress)
		
		// Récupérer depuis le graphe après le calcul (peut encore être la valeur par défaut)
		m.mutex.RLock()
		if wallet, exists := m.trustGraph.Wallets[walletAddress]; exists {
			score = wallet.TrustScore
			m.mutex.RUnlock()
			return score, nil
		}
		m.mutex.RUnlock()
		
		// Si toujours pas trouvé, retourner une valeur par défaut
		return 50.0, nil
	}
	
	// Mettre à jour le graphe de confiance
	m.mutex.Lock()
	if _, exists := m.trustGraph.Wallets[walletAddress]; !exists {
		m.trustGraph.Wallets[walletAddress] = &WalletNode{
			Address:      walletAddress,
			TrustScore:   score,
			Interactions: []string{},
			LastUpdated:  time.Now(),
		}
	} else {
		m.trustGraph.Wallets[walletAddress].TrustScore = score
		m.trustGraph.Wallets[walletAddress].LastUpdated = time.Now()
	}
	m.mutex.Unlock()
	
	// Mettre en cache pour une utilisation future
	m.cache.SetFloat64(cacheKey, score, 1*time.Hour)
	
	return score, nil
}

// GetTokenTrustMetrics récupère les métriques de confiance pour un token
func (m *TrustNetwork) GetTokenTrustMetrics(tokenAddress string) (*models.TokenTrustMetrics, error) {
	// Essayer d'abord de récupérer depuis le cache
	cacheKey := fmt.Sprintf("token:trust:%s", tokenAddress)
	var metrics models.TokenTrustMetrics
	err := m.cache.GetStruct(cacheKey, &metrics)
	if err == nil {
		// Métriques trouvées dans le cache
		return &metrics, nil
	}
	
	// Préparer les métriques de base
	metrics = models.TokenTrustMetrics{
		TokenAddress:          tokenAddress,
		TrustScoreDistribution: make(map[string]int),
	}
	
	// Récupérer les données du token depuis le graphe
	m.mutex.RLock()
	tokenNode, exists := m.trustGraph.Tokens[tokenAddress]
	if !exists {
		m.mutex.RUnlock()
		return &metrics, nil // Retourner des métriques vides si le token n'existe pas
	}
	
	// Copier les wallets pour éviter de garder le mutex verrouillé pendant traitement
	wallets := make([]string, len(tokenNode.Wallets))
	copy(wallets, tokenNode.Wallets)
	m.mutex.RUnlock()
	
	// Initialiser les compteurs
	metrics.ActiveWallets = len(wallets)
	totalTrustScore := 0.0
	trustedCount := 0
	
	// Analyser chaque wallet
	for _, walletAddr := range wallets {
		// Récupérer le score de confiance
		score, err := m.GetWalletTrustScore(walletAddr)
		if err != nil {
			continue // Ignorer ce wallet en cas d'erreur
		}
		
		// Mettre à jour les métriques
		totalTrustScore += score
		
		// Catégoriser le score
		var category string
		switch {
		case score >= 90:
			category = "excellent"
			trustedCount++
		case score >= 75:
			category = "high"
			trustedCount++
		case score >= 60:
			category = "good"
			trustedCount++
		case score >= 40:
			category = "average"
		case score >= 25:
			category = "low"
		default:
			category = "poor"
		}
		
		metrics.TrustScoreDistribution[category]++
	}
	
	// Calculer le score moyen si des wallets existent
	if metrics.ActiveWallets > 0 {
		metrics.AvgTrustScore = totalTrustScore / float64(metrics.ActiveWallets)
		metrics.TrustedWallets = trustedCount
	}
	
	// Calculer le ratio de confiance précoce
	earlyWallets, err := m.getEarlyWallets(tokenAddress, 50) // Limite aux 50 premiers wallets
	if err == nil && len(earlyWallets) > 0 {
		trustedEarlyCount := 0
		for _, walletAddr := range earlyWallets {
			score, err := m.GetWalletTrustScore(walletAddr)
			if err == nil && score >= 60 { // Score considéré comme "trusted"
				trustedEarlyCount++
			}
		}
		
		if len(earlyWallets) > 0 {
			metrics.EarlyTrustRatio = float64(trustedEarlyCount) / float64(len(earlyWallets))
		}
	}
	
	// Mettre en cache pour une utilisation future
	m.cache.SetStruct(cacheKey, metrics, 30*time.Minute)
	
	return &metrics, nil
}

// updateWalletTrustScore met à jour le score de confiance d'un wallet
func (m *TrustNetwork) updateWalletTrustScore(walletAddress string) {
	// Calculer le nouveau score
	newScore := m.calculateWalletTrustScore(walletAddress)
	
	// Mettre à jour le graphe
	m.mutex.Lock()
	if wallet, exists := m.trustGraph.Wallets[walletAddress]; exists {
		// Vérifier si le changement est significatif (plus de 5%)
		if abs(wallet.TrustScore-newScore) > 5.0 {
			wallet.TrustScore = newScore
			wallet.LastUpdated = time.Now()
			
			// Sauvegarder dans la base de données
			m.db.SaveWalletTrustScore(walletAddress, newScore, wallet.LastUpdated)
			
			// Mettre à jour le cache
			cacheKey := fmt.Sprintf("wallet:trust:%s", walletAddress)
			m.cache.SetFloat64(cacheKey, newScore, 1*time.Hour)
		}
	} else {
		// Nouveau wallet
		m.trustGraph.Wallets[walletAddress] = &WalletNode{
			Address:      walletAddress,
			TrustScore:   newScore,
			Interactions: []string{},
			LastUpdated:  time.Now(),
		}
		
		// Sauvegarder dans la base de données
		m.db.SaveWalletTrustScore(walletAddress, newScore, time.Now())
		
		// Mettre à jour le cache
		cacheKey := fmt.Sprintf("wallet:trust:%s", walletAddress)
		m.cache.SetFloat64(cacheKey, newScore, 1*time.Hour)
	}
	m.mutex.Unlock()
}

// calculateWalletTrustScore calcule le score de confiance d'un wallet
func (m *TrustNetwork) calculateWalletTrustScore(walletAddress string) float64 {
	// Récupérer l'historique des interactions
	interactions, err := m.db.GetWalletInteractions(walletAddress, 1000) // Limiter à 1000 interactions
	if err != nil || len(interactions) == 0 {
		// Pas d'historique, retourner un score par défaut
		return 50.0
	}
	
	// Calculer le score basé sur différents facteurs
	// L'algorithme ci-dessous est simplifié et devrait être enrichi
	// avec des analyses plus complexes dans une implémentation réelle
	
	// 1. Performance basée sur les profits
	profitScore := m.evaluateProfitPerformance(interactions)
	
	// 2. Performance basée sur le timing
	timingScore := m.evaluateTimingPerformance(interactions)
	
	// 3. Performance basée sur le volume
	volumeScore := m.evaluateVolumePerformance(interactions)
	
	// 4. Performance basée sur le réseau
	networkScore := m.evaluateNetworkPerformance(walletAddress)
	
	// Pondération des facteurs
	// 40% profit, 25% timing, 15% volume, 20% réseau
	finalScore := (profitScore * 0.40) + (timingScore * 0.25) + (volumeScore * 0.15) + (networkScore * 0.20)
	
	// Borner le score entre 0 et 100
	if finalScore < 0 {
		finalScore = 0
	} else if finalScore > 100 {
		finalScore = 100
	}
	
	return finalScore
}

// evaluateProfitPerformance évalue la performance du wallet basée sur ses profits
func (m *TrustNetwork) evaluateProfitPerformance(interactions []models.WalletInteraction) float64 {
	// Cette fonction est très simplifiée
	// Une implémentation réelle analyserait:
	// - Le ratio de trades gagnants vs perdants
	// - L'ampleur des gains et pertes
	// - La constance des performances
	
	// Pour cet exemple, on suppose un score moyen
	return 50.0
}

// evaluateTimingPerformance évalue la performance du wallet basée sur son timing
func (m *TrustNetwork) evaluateTimingPerformance(interactions []models.WalletInteraction) float64 {
	// Cette fonction est très simplifiée
	// Une implémentation réelle analyserait:
	// - L'entrée précoce sur les tokens prometteurs
	// - La capacité à vendre près des sommets
	// - La capacité à éviter les rugpulls
	
	// Pour cet exemple, on suppose un score moyen
	// Influencé par le nombre d'interactions (plus d'expérience = meilleur score)
	baseScore := 50.0
	
	// Bonus basé sur l'expérience
	if len(interactions) > 500 {
		baseScore += 20.0
	} else if len(interactions) > 200 {
		baseScore += 15.0
	} else if len(interactions) > 100 {
		baseScore += 10.0
	} else if len(interactions) > 50 {
		baseScore += 5.0
	}
	
	return min(baseScore, 100.0)
}

// evaluateVolumePerformance évalue la performance du wallet basée sur ses volumes
func (m *TrustNetwork) evaluateVolumePerformance(interactions []models.WalletInteraction) float64 {
	// Cette fonction est très simplifiée
	// Une implémentation réelle analyserait:
	// - Le volume total tradé
	// - La constance des volumes
	// - Le rapport entre volume et profit
	
	// Pour cet exemple, on suppose un score moyen
	return 50.0
}

// evaluateNetworkPerformance évalue la performance du wallet basée sur son réseau
func (m *TrustNetwork) evaluateNetworkPerformance(walletAddress string) float64 {
	// Cette fonction est très simplifiée
	// Une implémentation réelle analyserait:
	// - Les connections avec d'autres wallets fiables
	// - La participation à des tokens de qualité
	// - L'influence dans l'écosystème
	
	// Récupérer les wallets similaires pour évaluer le réseau
	similarWallets, err := m.GetSimilarWallets(walletAddress, 0.2, 10)
	if err != nil || len(similarWallets) == 0 {
		// Pas de données réseau, score moyen
		return 50.0
	}
	
	// Calculer le score réseau basé sur les scores de confiance des wallets similaires
	totalScore := 0.0
	for _, similar := range similarWallets {
		totalScore += similar.TrustScore
	}
	
	// Score moyen des wallets similaires
	networkScore := totalScore / float64(len(similarWallets))
	
	// Pondérer: 30% score propre du wallet, 70% score du réseau
	return networkScore
}

// getEarlyWallets récupère les wallets entrés tôt sur un token
func (m *TrustNetwork) getEarlyWallets(tokenAddress string, limit int) ([]string, error) {
	// Récupérer les premières transactions pour ce token
	earlyTransactions, err := m.db.GetEarlyTokenTransactions(tokenAddress, limit)
	if err != nil {
		return nil, err
	}
	
	// Extraire les adresses de wallet uniques
	walletMap := make(map[string]bool)
	for _, tx := range earlyTransactions {
		walletMap[tx.WalletAddress] = true
	}
	
	// Convertir en slice
	wallets := make([]string, 0, len(walletMap))
	for wallet := range walletMap {
		wallets = append(wallets, wallet)
	}
	
	return wallets, nil
}

// maintenanceRoutine exécute des tâches de maintenance périodiques
func (m *TrustNetwork) maintenanceRoutine(ctx context.Context) {
	defer m.wg.Done()
	
	ticker := time.NewTicker(m.maintenanceInterval)
	defer ticker.Stop()
	
	m.logger.Info("Démarrage de la routine de maintenance du Memory of Trust")
	
	for {
		select {
		case <-m.stopCh:
			m.logger.Info("Arrêt de la routine de maintenance")
			return
		case <-ctx.Done():
			m.logger.Info("Contexte terminé, arrêt de la routine de maintenance")
			return
		case <-ticker.C:
			// Exécuter les tâches de maintenance
			m.logger.Info("Exécution des tâches de maintenance du Memory of Trust")
			
			// 1. Sauvegarder le graphe de confiance
			if err := m.saveTrustGraph(); err != nil {
				m.logger.Error("Échec de la sauvegarde du graphe de confiance", err)
			}
			
			// 2. Mettre à jour les similarités entre wallets
			if err := m.UpdateWalletSimilarities(); err != nil {
				m.logger.Error("Échec de la mise à jour des similarités", err)
			}
			
			// 3. Nettoyer les caches obsolètes
			m.cleanObsoleteCaches()
			
			// 4. Optimiser les index pour les requêtes futures
			if err := m.db.OptimizeIndexes(); err != nil {
				m.logger.Error("Échec de l'optimisation des index", err)
			}
			
			m.logger.Info("Maintenance du Memory of Trust terminée")
		}
	}
}

// abs retourne la valeur absolue d'un float64
func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

// min retourne le minimum de deux float64
func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

// GetTokenActiveWallets récupère les wallets actifs sur un token avec un filtre de score de confiance minimal
func (m *TrustNetwork) GetTokenActiveWallets(tokenAddress string, minTrustScore float64, limit int) ([]models.ActiveWallet, error) {
	// Essayer d'abord depuis le cache
	cacheKey := fmt.Sprintf("token:%s:active_wallets:trust_score:%f", tokenAddress, minTrustScore)
	var activeWallets []models.ActiveWallet
	err := m.cache.GetStruct(cacheKey, &activeWallets)
	if err == nil && len(activeWallets) > 0 {
		// Limiter le nombre de résultats si nécessaire
		if limit > 0 && len(activeWallets) > limit {
			return activeWallets[:limit], nil
		}
		return activeWallets, nil
	}

	// Récupérer depuis la base de données
	if minTrustScore > 0 {
		// Si un score minimum est requis
		activeWallets, err = m.db.GetActiveWalletsByTrustScore(tokenAddress, minTrustScore, limit)
	} else {
		// Sinon, récupérer tous les wallets actifs
		if limit > 0 {
			activeWallets, err = m.db.GetActiveWalletsPaginated(tokenAddress, 0, limit)
		} else {
			activeWallets, err = m.db.GetAllTokenActiveWallets(tokenAddress, 1000)
		}
	}

	if err != nil {
		return nil, fmt.Errorf("échec de la récupération des wallets actifs: %w", err)
	}

	// Mettre en cache pour utilisation future
	cacheDuration := 30 * time.Minute
	m.cache.CacheActiveWalletsByTrustScore(tokenAddress, minTrustScore, activeWallets, cacheDuration)

	// Limiter le nombre de résultats si nécessaire
	if limit > 0 && len(activeWallets) > limit {
		return activeWallets[:limit], nil
	}

	return activeWallets, nil
}

// GetActiveWalletsCount récupère le nombre total de wallets actifs sur un token
func (m *TrustNetwork) GetActiveWalletsCount(tokenAddress string) (int, error) {
	// Essayer d'abord depuis le cache
	count, err := m.cache.GetCachedActiveWalletsCount(tokenAddress)
	if err == nil {
		return count, nil
	}

	// Récupérer depuis la base de données
	count, err = m.db.GetActiveWalletsCount(tokenAddress)
	if err != nil {
		return 0, fmt.Errorf("échec du comptage des wallets actifs: %w", err)
	}

	// Mettre en cache pour utilisation future
	cacheDuration := 1 * time.Hour
	m.cache.CacheActiveWalletsCount(tokenAddress, count, cacheDuration)

	return count, nil
}

// GetTokenWallets récupère les wallets actifs sur un token
func (m *TrustNetwork) GetTokenWallets(tokenAddress string, limit int) ([]models.ActiveWallet, error) {
	// Essayer d'abord depuis le cache
	cacheKey := fmt.Sprintf("token:%s:wallets", tokenAddress)
	var wallets []models.ActiveWallet
	err := m.cache.GetStruct(cacheKey, &wallets)
	if err == nil && len(wallets) > 0 {
		// Limiter si nécessaire
		if limit > 0 && len(wallets) > limit {
			return wallets[:limit], nil
		}
		return wallets, nil
	}
	
	// Récupérer depuis la base de données
	wallets, err = m.db.GetAllTokenActiveWallets(tokenAddress, 1000)
	if err != nil {
		return nil, fmt.Errorf("failed to get token wallets: %w", err)
	}
	
	// Mettre en cache
	m.cache.SetStruct(cacheKey, wallets, 30*time.Minute)
	
	// Limiter si nécessaire
	if limit > 0 && len(wallets) > limit {
		return wallets[:limit], nil
	}
	
	return wallets, nil
} 