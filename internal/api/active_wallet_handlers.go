package api

import (
	"encoding/json"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/franky69420/crypto-oracle/internal/memory"
	"github.com/franky69420/crypto-oracle/pkg/utils/logger"
	"github.com/franky69420/crypto-oracle/pkg/models"
)

// ActiveWalletHandler gère les requêtes API relatives aux wallets actifs
type ActiveWalletHandler struct {
	trustNetwork memory.MemoryOfTrust
	logger       *logger.Logger
}

// NewActiveWalletHandler crée un nouveau gestionnaire de wallets actifs
func NewActiveWalletHandler(trustNetwork memory.MemoryOfTrust, logger *logger.Logger) *ActiveWalletHandler {
	return &ActiveWalletHandler{
		trustNetwork: trustNetwork,
		logger:       logger,
	}
}

// RegisterRoutes enregistre les routes de l'API pour les wallets actifs
func (h *ActiveWalletHandler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/api/tokens/{tokenAddress}/active-wallets", h.GetTokenActiveWallets).Methods("GET")
	router.HandleFunc("/api/tokens/{tokenAddress}/active-wallets/count", h.GetTokenActiveWalletsCount).Methods("GET")
	router.HandleFunc("/api/tokens/{tokenAddress}/active-wallets/trusted", h.GetTrustedActiveWallets).Methods("GET")
	router.HandleFunc("/api/tokens/{tokenAddress}/active-wallets/search", h.SearchActiveWallets).Methods("GET")
	router.HandleFunc("/api/tokens/{tokenAddress}/active-wallets/recent", h.GetRecentActiveWallets).Methods("GET")
}

// GetTokenActiveWallets retourne la liste des wallets actifs sur un token
func (h *ActiveWalletHandler) GetTokenActiveWallets(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	tokenAddress := vars["tokenAddress"]

	// Paramètres optionnels
	limitStr := r.URL.Query().Get("limit")
	limit := 100 // Valeur par défaut
	if limitStr != "" {
		parsedLimit, err := strconv.Atoi(limitStr)
		if err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	// Récupérer les wallets actifs
	activeWallets, err := h.trustNetwork.GetTokenActiveWallets(tokenAddress, 0, limit)
	if err != nil {
		h.logger.Error("Échec de la récupération des wallets actifs", err, map[string]interface{}{
			"token_address": tokenAddress,
		})
		http.Error(w, "Erreur lors de la récupération des wallets actifs", http.StatusInternalServerError)
		return
	}

	// Répondre avec les wallets actifs
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"token_address":  tokenAddress,
		"active_wallets": activeWallets,
		"count":          len(activeWallets),
	})
}

// GetTokenActiveWalletsCount retourne le nombre de wallets actifs sur un token
func (h *ActiveWalletHandler) GetTokenActiveWalletsCount(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	tokenAddress := vars["tokenAddress"]

	// Récupérer le nombre de wallets actifs
	count, err := h.trustNetwork.GetActiveWalletsCount(tokenAddress)
	if err != nil {
		h.logger.Error("Échec du comptage des wallets actifs", err, map[string]interface{}{
			"token_address": tokenAddress,
		})
		http.Error(w, "Erreur lors du comptage des wallets actifs", http.StatusInternalServerError)
		return
	}

	// Répondre avec le nombre de wallets actifs
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"token_address": tokenAddress,
		"count":         count,
	})
}

// GetTrustedActiveWallets retourne la liste des wallets actifs de confiance sur un token
func (h *ActiveWalletHandler) GetTrustedActiveWallets(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	tokenAddress := vars["tokenAddress"]

	// Paramètres optionnels
	limitStr := r.URL.Query().Get("limit")
	limit := 50 // Valeur par défaut
	if limitStr != "" {
		parsedLimit, err := strconv.Atoi(limitStr)
		if err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	minScoreStr := r.URL.Query().Get("min_score")
	minScore := 70.0 // Valeur par défaut pour les wallets de confiance
	if minScoreStr != "" {
		parsedScore, err := strconv.ParseFloat(minScoreStr, 64)
		if err == nil && parsedScore > 0 {
			minScore = parsedScore
		}
	}

	// Récupérer les wallets actifs de confiance
	trustedWallets, err := h.trustNetwork.GetTokenActiveWallets(tokenAddress, minScore, limit)
	if err != nil {
		h.logger.Error("Échec de la récupération des wallets actifs de confiance", err, map[string]interface{}{
			"token_address": tokenAddress,
			"min_score":     minScore,
		})
		http.Error(w, "Erreur lors de la récupération des wallets actifs de confiance", http.StatusInternalServerError)
		return
	}

	// Répondre avec les wallets actifs de confiance
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"token_address":    tokenAddress,
		"min_trust_score":  minScore,
		"trusted_wallets":  trustedWallets,
		"count":            len(trustedWallets),
	})
}

// SearchActiveWallets recherche parmi les wallets actifs d'un token
func (h *ActiveWalletHandler) SearchActiveWallets(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	tokenAddress := vars["tokenAddress"]

	// Paramètres de recherche
	query := r.URL.Query().Get("query")
	minTrustScoreStr := r.URL.Query().Get("min_trust_score")
	maxTrustScoreStr := r.URL.Query().Get("max_trust_score")
	minTransactionsStr := r.URL.Query().Get("min_transactions")
	limitStr := r.URL.Query().Get("limit")

	// Valeurs par défaut et parsing
	limit := 50
	if limitStr != "" {
		parsedLimit, err := strconv.Atoi(limitStr)
		if err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	minTrustScore := 0.0
	if minTrustScoreStr != "" {
		parsedScore, err := strconv.ParseFloat(minTrustScoreStr, 64)
		if err == nil {
			minTrustScore = parsedScore
		}
	}

	maxTrustScore := 100.0
	if maxTrustScoreStr != "" {
		parsedScore, err := strconv.ParseFloat(maxTrustScoreStr, 64)
		if err == nil {
			maxTrustScore = parsedScore
		}
	}

	minTransactions := 0
	if minTransactionsStr != "" {
		parsedMin, err := strconv.Atoi(minTransactionsStr)
		if err == nil && parsedMin > 0 {
			minTransactions = parsedMin
		}
	}

	// Récupérer tous les wallets actifs
	allWallets, err := h.trustNetwork.GetTokenActiveWallets(tokenAddress, 0, 0)
	if err != nil {
		h.logger.Error("Échec de la récupération des wallets actifs pour la recherche", err, map[string]interface{}{
			"token_address": tokenAddress,
		})
		http.Error(w, "Erreur lors de la recherche des wallets actifs", http.StatusInternalServerError)
		return
	}

	// Filtrer selon les critères
	filteredWallets := make([]map[string]interface{}, 0)
	for _, wallet := range allWallets {
		// Vérifier les critères de filtrage
		if wallet.TrustScore < minTrustScore || wallet.TrustScore > maxTrustScore {
			continue
		}

		if wallet.TransactionCount < minTransactions {
			continue
		}

		// Recherche par adresse si une requête est spécifiée
		if query != "" && !strings.Contains(strings.ToLower(wallet.Address), strings.ToLower(query)) {
			continue
		}

		// Ajouter à la liste filtrée avec des informations supplémentaires
		walletInfo := map[string]interface{}{
			"address":                   wallet.Address,
			"first_transaction_timestamp": wallet.FirstTransactionTimestamp,
			"entry_rank":                wallet.EntryRank,
			"transaction_count":         wallet.TransactionCount,
			"last_active":               wallet.LastActive,
			"trust_score":               wallet.TrustScore,
			"days_active":               time.Since(wallet.FirstTransactionTimestamp).Hours() / 24,
		}

		filteredWallets = append(filteredWallets, walletInfo)
	}

	// Limiter le nombre de résultats
	if len(filteredWallets) > limit {
		filteredWallets = filteredWallets[:limit]
	}

	// Répondre avec les wallets filtrés
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"token_address": tokenAddress,
		"wallets":        filteredWallets,
		"count":         len(filteredWallets),
		"filters": map[string]interface{}{
			"query":            query,
			"min_trust_score":  minTrustScore,
			"max_trust_score":  maxTrustScore,
			"min_transactions": minTransactions,
		},
	})
}

// GetRecentActiveWallets retourne les wallets récemment actifs sur un token
func (h *ActiveWalletHandler) GetRecentActiveWallets(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	tokenAddress := vars["tokenAddress"]

	// Paramètres optionnels
	hoursStr := r.URL.Query().Get("hours")
	hours := 24 // Valeur par défaut: 24 dernières heures
	if hoursStr != "" {
		parsedHours, err := strconv.Atoi(hoursStr)
		if err == nil && parsedHours > 0 {
			hours = parsedHours
		}
	}

	limitStr := r.URL.Query().Get("limit")
	limit := 50 // Valeur par défaut
	if limitStr != "" {
		parsedLimit, err := strconv.Atoi(limitStr)
		if err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	// Récupérer les wallets récemment actifs
	// Puisque MemoryOfTrust n'a pas de méthode dédiée pour les wallets récents,
	// on va récupérer tous les wallets et les filtrer par leur dernier timestamp d'activité
	allWallets, err := h.trustNetwork.GetTokenActiveWallets(tokenAddress, 0, 0)
	if err != nil {
		h.logger.Error("Échec de la récupération des wallets récemment actifs", err, map[string]interface{}{
			"token_address": tokenAddress,
			"hours":        hours,
		})
		http.Error(w, "Erreur lors de la récupération des wallets récemment actifs", http.StatusInternalServerError)
		return
	}
	
	// Filtrer par dernière activité
	cutoffTime := time.Now().Add(-time.Duration(hours) * time.Hour)
	recentWallets := make([]models.ActiveWallet, 0)
	
	for _, wallet := range allWallets {
		if wallet.LastActive.After(cutoffTime) {
			recentWallets = append(recentWallets, wallet)
		}
	}
	
	// Trier par timestamp d'activité décroissant (les plus récents d'abord)
	sort.Slice(recentWallets, func(i, j int) bool {
		return recentWallets[i].LastActive.After(recentWallets[j].LastActive)
	})
	
	// Limiter le nombre de résultats
	if len(recentWallets) > limit {
		recentWallets = recentWallets[:limit]
	}

	// Répondre avec les wallets récemment actifs
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"token_address":   tokenAddress,
		"hours":          hours,
		"active_wallets": recentWallets,
		"count":          len(recentWallets),
	})
} 