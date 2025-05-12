package db

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/franko/crypto-oracle/pkg/models"
)

// GetEarlyTokenTransactions récupère les premières transactions d'un token
func (c *Connection) GetEarlyTokenTransactions(tokenAddress string, limit int) ([]models.WalletInteraction, error) {
	ctx := context.Background()
	
	query := `
		SELECT id, wallet_address, token_address, token_symbol, tx_hash, block_number,
			timestamp, action_type, amount, value, price, success, 
			related_buy_timestamp, token_risk_factor
		FROM wallet_interactions
		WHERE token_address = $1 AND action_type = 'buy'
		ORDER BY timestamp ASC
		LIMIT $2
	`
	
	rows, err := c.pool.Query(ctx, query, tokenAddress, limit)
	if err != nil {
		return nil, fmt.Errorf("échec de la récupération des transactions précoces: %w", err)
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
			return nil, fmt.Errorf("échec du scan des transactions: %w", err)
		}
		
		interactions = append(interactions, interaction)
	}
	
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("erreur pendant l'itération sur les résultats: %w", err)
	}
	
	return interactions, nil
}

// GetRecentInteractions récupère les interactions récentes (tous tokens confondus)
func (c *Connection) GetRecentInteractions(limit int) ([]models.WalletInteraction, error) {
	ctx := context.Background()
	
	query := `
		SELECT id, wallet_address, token_address, token_symbol, tx_hash, block_number,
			timestamp, action_type, amount, value, price, success, 
			related_buy_timestamp, token_risk_factor
		FROM wallet_interactions
		ORDER BY timestamp DESC
		LIMIT $1
	`
	
	rows, err := c.pool.Query(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("échec de la récupération des interactions récentes: %w", err)
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

// GetTokenTraders récupère les traders actifs sur un token
func (c *Connection) GetTokenTraders(tokenAddress string, limit int) ([]models.TokenTrader, error) {
	ctx := context.Background()
	
	query := `
		WITH trader_stats AS (
			SELECT 
				w.wallet_address,
				COUNT(*) as tx_count,
				SUM(w.value) as total_volume,
				(SELECT MIN(wi.timestamp) FROM wallet_interactions wi WHERE wi.token_address = $1 AND wi.wallet_address = w.wallet_address) as first_tx,
				(SELECT COUNT(*) FROM wallet_interactions) as total_tx_count,
				(SELECT SUM(wi.value) FROM wallet_interactions wi WHERE wi.token_address = $1) as token_total_volume
			FROM wallet_interactions w
			WHERE w.token_address = $1
			GROUP BY w.wallet_address
		)
		SELECT 
			wallet_address,
			$1 as token_address,
			(total_volume / NULLIF(token_total_volume, 0)) as relative_volume,
			CASE 
				WHEN first_tx IS NOT NULL THEN 
					1.0 - (EXTRACT(EPOCH FROM (first_tx - (SELECT MIN(timestamp) FROM wallet_interactions WHERE token_address = $1))) / 
					       NULLIF(EXTRACT(EPOCH FROM ((SELECT MAX(timestamp) FROM wallet_interactions WHERE token_address = $1) - 
					                                  (SELECT MIN(timestamp) FROM wallet_interactions WHERE token_address = $1))), 0))
				ELSE 0
			END as early_investor,
			tx_count
		FROM trader_stats
		ORDER BY relative_volume DESC
		LIMIT $2
	`
	
	rows, err := c.pool.Query(ctx, query, tokenAddress, limit)
	if err != nil {
		return nil, fmt.Errorf("échec de la récupération des traders: %w", err)
	}
	defer rows.Close()
	
	traders := make([]models.TokenTrader, 0)
	
	for rows.Next() {
		var trader models.TokenTrader
		
		err := rows.Scan(
			&trader.WalletAddress,
			&trader.TokenAddress,
			&trader.RelativeVolume,
			&trader.EarlyInvestor,
			&trader.TransactionCount,
		)
		
		if err != nil {
			return nil, fmt.Errorf("échec du scan des traders: %w", err)
		}
		
		traders = append(traders, trader)
	}
	
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("erreur pendant l'itération sur les résultats: %w", err)
	}
	
	return traders, nil
}

// GetWalletPriceImpact calcule l'impact des transactions d'un wallet sur le prix d'un token
func (c *Connection) GetWalletPriceImpact(walletAddress, tokenAddress string) (float64, error) {
	ctx := context.Background()
	
	query := `
		WITH price_changes AS (
			SELECT 
				w1.wallet_address,
				w1.timestamp,
				w1.price as price_before,
				(SELECT w2.price 
				 FROM wallet_interactions w2 
				 WHERE w2.token_address = $2 AND w2.timestamp > w1.timestamp 
				 ORDER BY w2.timestamp ASC 
				 LIMIT 1) as price_after
			FROM wallet_interactions w1
			WHERE w1.wallet_address = $1 AND w1.token_address = $2
			ORDER BY w1.timestamp
		)
		SELECT 
			AVG(ABS(COALESCE(price_after, 0) - price_before) / NULLIF(price_before, 0)) as avg_price_impact
		FROM price_changes
		WHERE price_after IS NOT NULL
	`
	
	var impact sql.NullFloat64
	err := c.pool.QueryRow(ctx, query, walletAddress, tokenAddress).Scan(&impact)
	if err != nil {
		return 0, fmt.Errorf("échec du calcul de l'impact sur le prix: %w", err)
	}
	
	if impact.Valid {
		return impact.Float64, nil
	}
	
	return 0, nil
}

// SaveTokenInfluencers enregistre les influenceurs d'un token
func (c *Connection) SaveTokenInfluencers(tokenAddress string, influencers []models.WalletInfluence) error {
	ctx := context.Background()
	tx, err := c.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("échec du démarrage de la transaction: %w", err)
	}
	defer tx.Rollback(ctx)
	
	// Supprimer les anciennes entrées
	deleteQuery := `
		DELETE FROM token_influencers
		WHERE token_address = $1
	`
	
	_, err = tx.Exec(ctx, deleteQuery, tokenAddress)
	if err != nil {
		return fmt.Errorf("échec de la suppression des anciens influenceurs: %w", err)
	}
	
	// Insérer les nouvelles entrées
	insertQuery := `
		INSERT INTO token_influencers (
			wallet_address, token_address, influence_score,
			volume_impact, timing_impact, price_impact, transaction_count
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7
		)
	`
	
	for _, influencer := range influencers {
		_, err = tx.Exec(ctx, insertQuery,
			influencer.WalletAddress,
			influencer.TokenAddress,
			influencer.InfluenceScore,
			influencer.VolumeImpact,
			influencer.TimingImpact,
			influencer.PriceImpact,
			influencer.TransactionCount,
		)
		
		if err != nil {
			return fmt.Errorf("échec de l'insertion d'un influenceur: %w", err)
		}
	}
	
	// Valider la transaction
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("échec de la validation de la transaction: %w", err)
	}
	
	return nil
}

// GetTokenInfluencers récupère les influenceurs d'un token
func (c *Connection) GetTokenInfluencers(tokenAddress string, limit int) ([]models.WalletInfluence, error) {
	ctx := context.Background()
	
	query := `
		SELECT wallet_address, token_address, influence_score,
		       volume_impact, timing_impact, price_impact, transaction_count
		FROM token_influencers
		WHERE token_address = $1
		ORDER BY influence_score DESC
		LIMIT $2
	`
	
	rows, err := c.pool.Query(ctx, query, tokenAddress, limit)
	if err != nil {
		return nil, fmt.Errorf("échec de la récupération des influenceurs: %w", err)
	}
	defer rows.Close()
	
	influencers := make([]models.WalletInfluence, 0)
	
	for rows.Next() {
		var influencer models.WalletInfluence
		
		err := rows.Scan(
			&influencer.WalletAddress,
			&influencer.TokenAddress,
			&influencer.InfluenceScore,
			&influencer.VolumeImpact,
			&influencer.TimingImpact,
			&influencer.PriceImpact,
			&influencer.TransactionCount,
		)
		
		if err != nil {
			return nil, fmt.Errorf("échec du scan des influenceurs: %w", err)
		}
		
		influencers = append(influencers, influencer)
	}
	
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("erreur pendant l'itération sur les résultats: %w", err)
	}
	
	return influencers, nil
}

// GetActiveWalletsByTrustScore récupère les wallets actifs sur un token triés par score de confiance
func (c *Connection) GetActiveWalletsByTrustScore(tokenAddress string, minTrustScore float64, limit int) ([]models.ActiveWallet, error) {
	ctx := context.Background()
	
	query := `
		WITH wallet_stats AS (
			SELECT 
				wallet_address,
				COUNT(*) as tx_count,
				MAX(timestamp) as last_activity,
				SUM(CASE WHEN action_type = 'buy' THEN value ELSE 0 END) as buy_volume,
				SUM(CASE WHEN action_type = 'sell' THEN value ELSE 0 END) as sell_volume
			FROM wallet_interactions
			WHERE token_address = $1
			GROUP BY wallet_address
		)
		SELECT
			w.wallet_address as address,
			COALESCE(ts.trust_score, 50) as trust_score,
			w.tx_count as transaction_count,
			w.last_activity,
			(w.buy_volume - w.sell_volume) as net_position,
			w.buy_volume,
			w.sell_volume
		FROM wallet_stats w
		LEFT JOIN wallet_trust_scores ts ON w.wallet_address = ts.wallet_address
		WHERE COALESCE(ts.trust_score, 50) >= $2
		ORDER BY ts.trust_score DESC, w.last_activity DESC
		LIMIT $3
	`
	
	rows, err := c.pool.Query(ctx, query, tokenAddress, minTrustScore, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get wallets by trust score: %w", err)
	}
	defer rows.Close()
	
	wallets := make([]models.ActiveWallet, 0)
	
	for rows.Next() {
		var wallet models.ActiveWallet
		
		err := rows.Scan(
			&wallet.Address,
			&wallet.TrustScore,
			&wallet.TransactionCount,
			&wallet.LastActivity,
			&wallet.NetPosition,
			&wallet.BuyVolume,
			&wallet.SellVolume,
		)
		
		if err != nil {
			return nil, fmt.Errorf("failed to scan wallet by trust score: %w", err)
		}
		
		wallets = append(wallets, wallet)
	}
	
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error during iteration of wallets by trust score: %w", err)
	}
	
	return wallets, nil
}

// GetActiveWalletsPaginated récupère les wallets actifs sur un token avec pagination
func (c *Connection) GetActiveWalletsPaginated(tokenAddress string, offset, limit int) ([]models.ActiveWallet, error) {
	ctx := context.Background()
	
	query := `
		WITH wallet_stats AS (
			SELECT 
				wallet_address,
				COUNT(*) as tx_count,
				MAX(timestamp) as last_activity,
				SUM(CASE WHEN action_type = 'buy' THEN value ELSE 0 END) as buy_volume,
				SUM(CASE WHEN action_type = 'sell' THEN value ELSE 0 END) as sell_volume
			FROM wallet_interactions
			WHERE token_address = $1
			GROUP BY wallet_address
		)
		SELECT
			w.wallet_address as address,
			COALESCE(ts.trust_score, 50) as trust_score,
			w.tx_count as transaction_count,
			w.last_activity,
			(w.buy_volume - w.sell_volume) as net_position,
			w.buy_volume,
			w.sell_volume
		FROM wallet_stats w
		LEFT JOIN wallet_trust_scores ts ON w.wallet_address = ts.wallet_address
		ORDER BY w.last_activity DESC
		LIMIT $2 OFFSET $3
	`
	
	rows, err := c.pool.Query(ctx, query, tokenAddress, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get wallets paginated: %w", err)
	}
	defer rows.Close()
	
	wallets := make([]models.ActiveWallet, 0)
	
	for rows.Next() {
		var wallet models.ActiveWallet
		
		err := rows.Scan(
			&wallet.Address,
			&wallet.TrustScore,
			&wallet.TransactionCount,
			&wallet.LastActivity,
			&wallet.NetPosition,
			&wallet.BuyVolume,
			&wallet.SellVolume,
		)
		
		if err != nil {
			return nil, fmt.Errorf("failed to scan wallet paginated: %w", err)
		}
		
		wallets = append(wallets, wallet)
	}
	
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error during iteration of wallets paginated: %w", err)
	}
	
	return wallets, nil
}

// GetActiveWalletsCount récupère le nombre total de wallets actifs sur un token
func (c *Connection) GetActiveWalletsCount(tokenAddress string) (int, error) {
	ctx := context.Background()
	
	query := `
		SELECT COUNT(DISTINCT wallet_address) as wallet_count
		FROM wallet_interactions
		WHERE token_address = $1
	`
	
	var count int
	err := c.pool.QueryRow(ctx, query, tokenAddress).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to get active wallets count: %w", err)
	}
	
	return count, nil
}

// GetAllTokenActiveWallets récupère tous les wallets actifs détenant un token
func (c *Connection) GetAllTokenActiveWallets(tokenAddress string, limit int) ([]models.ActiveWallet, error) {
	ctx := context.Background()
	
	query := `
		WITH wallet_stats AS (
			SELECT 
				wallet_address,
				COUNT(*) as tx_count,
				MAX(timestamp) as last_activity,
				SUM(CASE WHEN action_type = 'buy' THEN value ELSE 0 END) as buy_volume,
				SUM(CASE WHEN action_type = 'sell' THEN value ELSE 0 END) as sell_volume
			FROM wallet_interactions
			WHERE token_address = $1
			GROUP BY wallet_address
		)
		SELECT
			w.wallet_address as address,
			COALESCE(ts.trust_score, 50) as trust_score,
			w.tx_count as transaction_count,
			w.last_activity,
			(w.buy_volume - w.sell_volume) as net_position,
			w.buy_volume,
			w.sell_volume
		FROM wallet_stats w
		LEFT JOIN wallet_trust_scores ts ON w.wallet_address = ts.wallet_address
		ORDER BY w.last_activity DESC
		LIMIT $2
	`
	
	rows, err := c.pool.Query(ctx, query, tokenAddress, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get token active wallets: %w", err)
	}
	defer rows.Close()
	
	wallets := make([]models.ActiveWallet, 0)
	
	for rows.Next() {
		var wallet models.ActiveWallet
		
		err := rows.Scan(
			&wallet.Address,
			&wallet.TrustScore,
			&wallet.TransactionCount,
			&wallet.LastActivity,
			&wallet.NetPosition,
			&wallet.BuyVolume,
			&wallet.SellVolume,
		)
		
		if err != nil {
			return nil, fmt.Errorf("failed to scan token active wallet: %w", err)
		}
		
		wallets = append(wallets, wallet)
	}
	
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error during iteration of token active wallets: %w", err)
	}
	
	return wallets, nil
} 