package db

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/franky69420/crypto-oracle/pkg/models"
)

// SaveWalletInteraction enregistre une interaction wallet-token dans la base de données
func (c *Connection) SaveWalletInteraction(interaction *models.WalletInteraction) error {
	ctx := context.Background()
	
	// Générer un ID si non spécifié
	if interaction.ID == "" {
		interaction.ID = fmt.Sprintf("%s:%s:%d", interaction.WalletAddress, interaction.TokenAddress, interaction.Timestamp.Unix())
	}
	
	query := `
		INSERT INTO wallet_interactions (
			id, wallet_address, token_address, token_symbol, tx_hash, block_number,
			timestamp, action_type, amount, value, price, success, 
			related_buy_timestamp, token_risk_factor
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14
		) ON CONFLICT (id) DO UPDATE SET
			token_symbol = $4,
			value = $10,
			price = $11,
			success = $12,
			related_buy_timestamp = $13,
			token_risk_factor = $14
	`
	
	_, err := c.pool.Exec(ctx, query,
		interaction.ID,
		interaction.WalletAddress,
		interaction.TokenAddress,
		interaction.TokenSymbol,
		interaction.TxHash,
		interaction.BlockNumber,
		interaction.Timestamp,
		interaction.ActionType,
		interaction.Amount,
		interaction.Value,
		interaction.Price,
		interaction.Success,
		interaction.RelatedBuyTimestamp,
		interaction.TokenRiskFactor,
	)
	
	if err != nil {
		return fmt.Errorf("échec de l'enregistrement de l'interaction wallet: %w", err)
	}
	
	return nil
}

// GetWalletInteractions récupère les interactions d'un wallet avec une limite
func (c *Connection) GetWalletInteractions(walletAddress string, limit int) ([]models.WalletInteraction, error) {
	ctx := context.Background()
	
	query := `
		SELECT id, wallet_address, token_address, token_symbol, tx_hash, block_number,
			timestamp, action_type, amount, value, price, success, 
			related_buy_timestamp, token_risk_factor
		FROM wallet_interactions
		WHERE wallet_address = $1
		ORDER BY timestamp DESC
		LIMIT $2
	`
	
	rows, err := c.pool.Query(ctx, query, walletAddress, limit)
	if err != nil {
		return nil, fmt.Errorf("échec de la récupération des interactions: %w", err)
	}
	defer rows.Close()
	
	interactions := make([]models.WalletInteraction, 0)
	
	for rows.Next() {
		var interaction models.WalletInteraction
		
		err := rows.Scan(
			&interaction.ID,
			&interaction.WalletAddress,
			&interaction.TokenAddress,
			&interaction.TokenSymbol,
			&interaction.TxHash,
			&interaction.BlockNumber,
			&interaction.Timestamp,
			&interaction.ActionType,
			&interaction.Amount,
			&interaction.Value,
			&interaction.Price,
			&interaction.Success,
			&interaction.RelatedBuyTimestamp,
			&interaction.TokenRiskFactor,
		)
		
		if err != nil {
			return nil, fmt.Errorf("échec du scan des interactions: %w", err)
		}
		
		interactions = append(interactions, interaction)
	}
	
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("erreur pendant l'itération sur les résultats: %w", err)
	}
	
	return interactions, nil
}

// GetWalletInteractionsSince récupère les interactions d'un wallet depuis une date spécifique
func (c *Connection) GetWalletInteractionsSince(walletAddress string, since time.Time) ([]models.WalletInteraction, error) {
	ctx := context.Background()
	
	query := `
		SELECT id, wallet_address, token_address, token_symbol, tx_hash, block_number,
			timestamp, action_type, amount, value, price, success, 
			related_buy_timestamp, token_risk_factor
		FROM wallet_interactions
		WHERE wallet_address = $1 AND timestamp >= $2
		ORDER BY timestamp DESC
	`
	
	rows, err := c.pool.Query(ctx, query, walletAddress, since)
	if err != nil {
		return nil, fmt.Errorf("échec de la récupération des interactions: %w", err)
	}
	defer rows.Close()
	
	interactions := make([]models.WalletInteraction, 0)
	
	for rows.Next() {
		var interaction models.WalletInteraction
		
		err := rows.Scan(
			&interaction.ID,
			&interaction.WalletAddress,
			&interaction.TokenAddress,
			&interaction.TokenSymbol,
			&interaction.TxHash,
			&interaction.BlockNumber,
			&interaction.Timestamp,
			&interaction.ActionType,
			&interaction.Amount,
			&interaction.Value,
			&interaction.Price,
			&interaction.Success,
			&interaction.RelatedBuyTimestamp,
			&interaction.TokenRiskFactor,
		)
		
		if err != nil {
			return nil, fmt.Errorf("échec du scan des interactions: %w", err)
		}
		
		interactions = append(interactions, interaction)
	}
	
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("erreur pendant l'itération sur les résultats: %w", err)
	}
	
	return interactions, nil
}

// GetWalletTokenInteractions récupère les interactions entre un wallet et un token spécifique
func (c *Connection) GetWalletTokenInteractions(walletAddress, tokenAddress string, limit int) ([]models.WalletInteraction, error) {
	ctx := context.Background()
	
	query := `
		SELECT id, wallet_address, token_address, token_symbol, tx_hash, block_number,
			timestamp, action_type, amount, value, price, success, 
			related_buy_timestamp, token_risk_factor
		FROM wallet_interactions
		WHERE wallet_address = $1 AND token_address = $2
		ORDER BY timestamp DESC
		LIMIT $3
	`
	
	rows, err := c.pool.Query(ctx, query, walletAddress, tokenAddress, limit)
	if err != nil {
		return nil, fmt.Errorf("échec de la récupération des interactions wallet-token: %w", err)
	}
	defer rows.Close()
	
	interactions := make([]models.WalletInteraction, 0)
	
	for rows.Next() {
		var interaction models.WalletInteraction
		
		err := rows.Scan(
			&interaction.ID,
			&interaction.WalletAddress,
			&interaction.TokenAddress,
			&interaction.TokenSymbol,
			&interaction.TxHash,
			&interaction.BlockNumber,
			&interaction.Timestamp,
			&interaction.ActionType,
			&interaction.Amount,
			&interaction.Value,
			&interaction.Price,
			&interaction.Success,
			&interaction.RelatedBuyTimestamp,
			&interaction.TokenRiskFactor,
		)
		
		if err != nil {
			return nil, fmt.Errorf("échec du scan des interactions: %w", err)
		}
		
		interactions = append(interactions, interaction)
	}
	
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("erreur pendant l'itération sur les résultats: %w", err)
	}
	
	return interactions, nil
}

// SaveWalletTrustScore enregistre ou met à jour le score de confiance d'un wallet
func (c *Connection) SaveWalletTrustScore(walletAddress string, trustScore float64, lastUpdated time.Time) error {
	ctx := context.Background()
	
	query := `
		INSERT INTO wallet_trust_scores (wallet_address, trust_score, last_updated)
		VALUES ($1, $2, $3)
		ON CONFLICT (wallet_address) DO UPDATE SET
			trust_score = $2,
			last_updated = $3
	`
	
	_, err := c.pool.Exec(ctx, query, walletAddress, trustScore, lastUpdated)
	if err != nil {
		return fmt.Errorf("échec de l'enregistrement du score de confiance: %w", err)
	}
	
	return nil
}

// GetWalletTrustScore récupère le score de confiance d'un wallet
func (c *Connection) GetWalletTrustScore(walletAddress string) (float64, error) {
	ctx := context.Background()
	
	query := `
		SELECT trust_score
		FROM wallet_trust_scores
		WHERE wallet_address = $1
	`
	
	var trustScore float64
	err := c.pool.QueryRow(ctx, query, walletAddress).Scan(&trustScore)
	if err != nil {
		return 0, fmt.Errorf("échec de la récupération du score de confiance: %w", err)
	}
	
	return trustScore, nil
}

// GetAllWalletTrustScores récupère tous les scores de confiance des wallets
func (c *Connection) GetAllWalletTrustScores() ([]models.WalletTrustScore, error) {
	ctx := context.Background()
	
	query := `
		SELECT wallet_address, trust_score, last_updated
		FROM wallet_trust_scores
	`
	
	rows, err := c.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("échec de la récupération des scores de confiance: %w", err)
	}
	defer rows.Close()
	
	scores := make([]models.WalletTrustScore, 0)
	
	for rows.Next() {
		var score models.WalletTrustScore
		
		err := rows.Scan(
			&score.Address,
			&score.TrustScore,
			&score.LastUpdated,
		)
		
		if err != nil {
			return nil, fmt.Errorf("échec du scan des scores de confiance: %w", err)
		}
		
		scores = append(scores, score)
	}
	
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("erreur pendant l'itération sur les résultats: %w", err)
	}
	
	return scores, nil
}

// GetRecentlyActiveWallets récupère les wallets actifs récemment
func (c *Connection) GetRecentlyActiveWallets(days int, limit int) ([]models.WalletTrustScore, error) {
	ctx := context.Background()
	
	query := `
		SELECT DISTINCT w.wallet_address, ts.trust_score, ts.last_updated
		FROM wallet_interactions w
		JOIN wallet_trust_scores ts ON w.wallet_address = ts.wallet_address
		WHERE w.timestamp >= NOW() - INTERVAL '$1 days'
		GROUP BY w.wallet_address, ts.trust_score, ts.last_updated
		ORDER BY COUNT(*) DESC
		LIMIT $2
	`
	
	rows, err := c.pool.Query(ctx, query, days, limit)
	if err != nil {
		return nil, fmt.Errorf("échec de la récupération des wallets actifs: %w", err)
	}
	defer rows.Close()
	
	wallets := make([]models.WalletTrustScore, 0)
	
	for rows.Next() {
		var wallet models.WalletTrustScore
		
		err := rows.Scan(
			&wallet.Address,
			&wallet.TrustScore,
			&wallet.LastUpdated,
		)
		
		if err != nil {
			return nil, fmt.Errorf("échec du scan des wallets actifs: %w", err)
		}
		
		wallets = append(wallets, wallet)
	}
	
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("erreur pendant l'itération sur les résultats: %w", err)
	}
	
	return wallets, nil
}

// GetMostTrustedWallets récupère les wallets avec les meilleurs scores de confiance
func (c *Connection) GetMostTrustedWallets(limit int) ([]models.WalletTrustScore, error) {
	ctx := context.Background()
	
	query := `
		SELECT wallet_address, trust_score, last_updated
		FROM wallet_trust_scores
		ORDER BY trust_score DESC
		LIMIT $1
	`
	
	rows, err := c.pool.Query(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("échec de la récupération des wallets de confiance: %w", err)
	}
	defer rows.Close()
	
	wallets := make([]models.WalletTrustScore, 0)
	
	for rows.Next() {
		var wallet models.WalletTrustScore
		
		err := rows.Scan(
			&wallet.Address,
			&wallet.TrustScore,
			&wallet.LastUpdated,
		)
		
		if err != nil {
			return nil, fmt.Errorf("échec du scan des wallets de confiance: %w", err)
		}
		
		wallets = append(wallets, wallet)
	}
	
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("erreur pendant l'itération sur les résultats: %w", err)
	}
	
	return wallets, nil
}

// SaveWalletRiskFactors enregistre les facteurs de risque d'un wallet
func (c *Connection) SaveWalletRiskFactors(walletAddress string, riskFactors *models.WalletRiskFactors) error {
	ctx := context.Background()
	
	query := `
		INSERT INTO wallet_risk_factors (
			wallet_address, risk_score, false_flagged_tokens, rugpull_exit_rate,
			fast_sell_rate, long_hold_rate, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7
		) ON CONFLICT (wallet_address) DO UPDATE SET
			risk_score = $2,
			false_flagged_tokens = $3,
			rugpull_exit_rate = $4,
			fast_sell_rate = $5,
			long_hold_rate = $6,
			updated_at = $7
	`
	
	_, err := c.pool.Exec(ctx, query,
		riskFactors.WalletAddress,
		riskFactors.RiskScore,
		riskFactors.FalseFlaggedTokens,
		riskFactors.RugpullExitRate,
		riskFactors.FastSellRate,
		riskFactors.LongHoldRate,
		riskFactors.UpdatedAt,
	)
	
	if err != nil {
		return fmt.Errorf("échec de l'enregistrement des facteurs de risque: %w", err)
	}
	
	return nil
}

// GetWalletRiskFactors récupère les facteurs de risque d'un wallet
func (c *Connection) GetWalletRiskFactors(walletAddress string) (*models.WalletRiskFactors, error) {
	ctx := context.Background()
	
	query := `
		SELECT wallet_address, risk_score, false_flagged_tokens, rugpull_exit_rate,
			fast_sell_rate, long_hold_rate, updated_at
		FROM wallet_risk_factors
		WHERE wallet_address = $1
	`
	
	var riskFactors models.WalletRiskFactors
	err := c.pool.QueryRow(ctx, query, walletAddress).Scan(
		&riskFactors.WalletAddress,
		&riskFactors.RiskScore,
		&riskFactors.FalseFlaggedTokens,
		&riskFactors.RugpullExitRate,
		&riskFactors.FastSellRate,
		&riskFactors.LongHoldRate,
		&riskFactors.UpdatedAt,
	)
	
	if err != nil {
		return nil, fmt.Errorf("échec de la récupération des facteurs de risque: %w", err)
	}
	
	return &riskFactors, nil
}

// GetWalletTokens récupère les tokens associés à un wallet
func (c *Connection) GetWalletTokens(walletAddress string, limit int) ([]models.WalletToken, error) {
	ctx := context.Background()
	
	query := `
		SELECT w.wallet_address, w.token_address, t.symbol, t.name,
			MIN(w.timestamp) as first_interaction,
			MAX(w.timestamp) as last_interaction,
			COUNT(*) as tx_count,
			SUM(CASE WHEN w.action_type = 'buy' THEN w.value ELSE 0 END) -
			SUM(CASE WHEN w.action_type = 'sell' THEN w.value ELSE 0 END) as net_profit,
			SUM(w.value) as total_volume
		FROM wallet_interactions w
		LEFT JOIN tokens t ON w.token_address = t.address
		WHERE w.wallet_address = $1
		GROUP BY w.wallet_address, w.token_address, t.symbol, t.name
		ORDER BY last_interaction DESC
		LIMIT $2
	`
	
	rows, err := c.pool.Query(ctx, query, walletAddress, limit)
	if err != nil {
		return nil, fmt.Errorf("échec de la récupération des tokens du wallet: %w", err)
	}
	defer rows.Close()
	
	tokens := make([]models.WalletToken, 0)
	
	for rows.Next() {
		var token models.WalletToken
		var netProfit sql.NullFloat64
		
		err := rows.Scan(
			&token.WalletAddress,
			&token.TokenAddress,
			&token.TokenSymbol,
			&token.TokenName,
			&token.FirstInteractionTime,
			&token.LastInteractionTime,
			&token.TransactionCount,
			&netProfit,
			&token.TotalVolume,
		)
		
		if err != nil {
			return nil, fmt.Errorf("échec du scan des tokens du wallet: %w", err)
		}
		
		if netProfit.Valid {
			token.NetProfit = netProfit.Float64
		}
		
		tokens = append(tokens, token)
	}
	
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("erreur pendant l'itération sur les résultats: %w", err)
	}
	
	return tokens, nil
} 