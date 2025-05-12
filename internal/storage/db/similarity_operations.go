package db

import (
	"context"
	"fmt"

	"github.com/franky69420/crypto-oracle/pkg/models"
)

// SaveWalletSimilarities enregistre les similarités entre wallets
func (c *Connection) SaveWalletSimilarities(walletAddress string, similarities []models.WalletSimilarity) error {
	ctx := context.Background()
	tx, err := c.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("échec du démarrage de la transaction: %w", err)
	}
	defer tx.Rollback(ctx)
	
	// Supprimer les anciennes entrées
	deleteQuery := `
		DELETE FROM wallet_similarities
		WHERE wallet_address = $1
	`
	
	_, err = tx.Exec(ctx, deleteQuery, walletAddress)
	if err != nil {
		return fmt.Errorf("échec de la suppression des anciennes similarités: %w", err)
	}
	
	// Insérer les nouvelles entrées
	insertQuery := `
		INSERT INTO wallet_similarities (
			wallet_address, similar_wallet_address, similarity_score,
			common_tokens, timing_score, position_score, trade_frequency
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7
		)
	`
	
	for _, similarity := range similarities {
		_, err = tx.Exec(ctx, insertQuery,
			walletAddress,
			similarity.WalletAddress,
			similarity.Score,
			similarity.CommonTokens,
			similarity.TimingScore,
			similarity.PositionScore,
			similarity.TradeFrequency,
		)
		
		if err != nil {
			return fmt.Errorf("échec de l'insertion d'une similarité: %w", err)
		}
	}
	
	// Valider la transaction
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("échec de la validation de la transaction: %w", err)
	}
	
	return nil
}

// GetWalletSimilarities récupère les wallets similaires à un wallet donné
func (c *Connection) GetWalletSimilarities(walletAddress string, minSimilarity float64, limit int) ([]models.WalletSimilarity, error) {
	ctx := context.Background()
	
	query := `
		SELECT s.wallet_address, s.similar_wallet_address, s.similarity_score,
		       s.common_tokens, s.timing_score, s.position_score, s.trade_frequency,
		       t.trust_score
		FROM wallet_similarities s
		LEFT JOIN wallet_trust_scores t ON s.similar_wallet_address = t.wallet_address
		WHERE s.wallet_address = $1 AND s.similarity_score >= $2
		ORDER BY s.similarity_score DESC
		LIMIT $3
	`
	
	rows, err := c.pool.Query(ctx, query, walletAddress, minSimilarity, limit)
	if err != nil {
		return nil, fmt.Errorf("échec de la récupération des similarités: %w", err)
	}
	defer rows.Close()
	
	similarities := make([]models.WalletSimilarity, 0)
	
	for rows.Next() {
		var similarity models.WalletSimilarity
		var sourceWallet string // Variable temporaire pour stocker l'adresse du wallet source
		
		err := rows.Scan(
			&sourceWallet, // wallet_address (source)
			&similarity.WalletAddress, // similar_wallet_address (cible)
			&similarity.Score,
			&similarity.CommonTokens,
			&similarity.TimingScore,
			&similarity.PositionScore,
			&similarity.TradeFrequency,
			&similarity.TrustScore,
		)
		
		if err != nil {
			return nil, fmt.Errorf("échec du scan des similarités: %w", err)
		}
		
		similarities = append(similarities, similarity)
	}
	
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("erreur pendant l'itération sur les résultats: %w", err)
	}
	
	return similarities, nil
} 