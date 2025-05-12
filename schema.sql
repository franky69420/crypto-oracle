-- Schéma SQL pour la base de données Crypto Oracle

-- Table des tokens
CREATE TABLE IF NOT EXISTS tokens (
    address VARCHAR(255) PRIMARY KEY,
    symbol VARCHAR(50) NOT NULL,
    name VARCHAR(255) NOT NULL,
    total_supply BIGINT,
    holder_count INTEGER,
    created_timestamp BIGINT,
    completed_timestamp BIGINT,
    last_trade_timestamp BIGINT,
    logo TEXT,
    twitter VARCHAR(255),
    website VARCHAR(255),
    telegram VARCHAR(255),
    cached_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Table des prix des tokens
CREATE TABLE IF NOT EXISTS token_prices (
    token_address VARCHAR(255) REFERENCES tokens(address),
    price DOUBLE PRECISION NOT NULL,
    change_1h DOUBLE PRECISION,
    change_24h DOUBLE PRECISION,
    change_7d DOUBLE PRECISION,
    volume_24h DOUBLE PRECISION,
    market_cap DOUBLE PRECISION,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    PRIMARY KEY (token_address, updated_at)
);

-- Table des métriques des tokens
CREATE TABLE IF NOT EXISTS token_metrics (
    token_address VARCHAR(255) REFERENCES tokens(address),
    holder_count INTEGER,
    intelligent_holders INTEGER,
    average_hold_time DOUBLE PRECISION,
    creator_wallet_addr VARCHAR(255),
    creator_trust_score DOUBLE PRECISION,
    dev_trust_score DOUBLE PRECISION,
    smart_money_holders INTEGER,
    average_trust_score DOUBLE PRECISION,
    risk_factor DOUBLE PRECISION,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    PRIMARY KEY (token_address, updated_at)
);

-- Table des trades de tokens
CREATE TABLE IF NOT EXISTS token_trades (
    id VARCHAR(255) PRIMARY KEY,
    token_address VARCHAR(255) REFERENCES tokens(address),
    wallet_address VARCHAR(255) NOT NULL,
    trade_type VARCHAR(10) NOT NULL,
    amount DOUBLE PRECISION NOT NULL,
    price DOUBLE PRECISION NOT NULL,
    total_value DOUBLE PRECISION NOT NULL,
    timestamp TIMESTAMP WITH TIME ZONE NOT NULL,
    tx_hash VARCHAR(255) NOT NULL,
    block_number BIGINT NOT NULL
);

-- Table des alertes sur les tokens
CREATE TABLE IF NOT EXISTS token_alerts (
    id VARCHAR(255) PRIMARY KEY,
    token_address VARCHAR(255) REFERENCES tokens(address),
    token_symbol VARCHAR(50) NOT NULL,
    alert_type VARCHAR(50) NOT NULL,
    severity VARCHAR(20) NOT NULL,
    message TEXT NOT NULL,
    detected_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    confirmation_count INTEGER DEFAULT 0,
    is_confirmed BOOLEAN DEFAULT FALSE
);

-- Table des métriques historiques des tokens
CREATE TABLE IF NOT EXISTS token_historical_metrics (
    token_address VARCHAR(255) REFERENCES tokens(address),
    date TIMESTAMP WITH TIME ZONE NOT NULL,
    price DOUBLE PRECISION,
    volume DOUBLE PRECISION,
    market_cap DOUBLE PRECISION,
    holder_count INTEGER,
    intelligent_ratio DOUBLE PRECISION,
    trust_score DOUBLE PRECISION,
    social_score DOUBLE PRECISION,
    PRIMARY KEY (token_address, date)
);

-- Table des points de prix des tokens
CREATE TABLE IF NOT EXISTS token_price_points (
    token_address VARCHAR(255) REFERENCES tokens(address),
    timestamp TIMESTAMP WITH TIME ZONE NOT NULL,
    open DOUBLE PRECISION NOT NULL,
    high DOUBLE PRECISION NOT NULL,
    low DOUBLE PRECISION NOT NULL,
    close DOUBLE PRECISION NOT NULL,
    volume DOUBLE PRECISION NOT NULL,
    PRIMARY KEY (token_address, timestamp)
);

-- Table des wallets
CREATE TABLE IF NOT EXISTS wallets (
    address VARCHAR(255) PRIMARY KEY,
    name VARCHAR(255),
    twitter_username VARCHAR(255),
    twitter_name VARCHAR(255),
    avatar TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    last_active TIMESTAMP WITH TIME ZONE
);

-- Table des scores de confiance des wallets
CREATE TABLE IF NOT EXISTS wallet_trust_scores (
    wallet_address VARCHAR(255) PRIMARY KEY,
    trust_score DOUBLE PRECISION NOT NULL,
    last_updated TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Table des facteurs de risque des wallets
CREATE TABLE IF NOT EXISTS wallet_risk_factors (
    wallet_address VARCHAR(255) PRIMARY KEY,
    risk_score DOUBLE PRECISION NOT NULL,
    false_flagged_tokens INTEGER,
    rugpull_exit_rate DOUBLE PRECISION,
    fast_sell_rate DOUBLE PRECISION,
    long_hold_rate DOUBLE PRECISION,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Table des interactions wallet-token
CREATE TABLE IF NOT EXISTS wallet_interactions (
    id VARCHAR(255) PRIMARY KEY,
    wallet_address VARCHAR(255) NOT NULL,
    token_address VARCHAR(255) REFERENCES tokens(address),
    token_symbol VARCHAR(50),
    tx_hash VARCHAR(255) NOT NULL,
    block_number BIGINT NOT NULL,
    timestamp TIMESTAMP WITH TIME ZONE NOT NULL,
    action_type VARCHAR(20) NOT NULL,
    amount DOUBLE PRECISION NOT NULL,
    value DOUBLE PRECISION NOT NULL,
    price DOUBLE PRECISION NOT NULL,
    success BOOLEAN NOT NULL,
    related_buy_timestamp TIMESTAMP WITH TIME ZONE,
    token_risk_factor DOUBLE PRECISION
);

-- Table des similarités entre wallets
CREATE TABLE IF NOT EXISTS wallet_similarities (
    wallet_address VARCHAR(255) NOT NULL,
    similar_wallet_address VARCHAR(255) NOT NULL,
    similarity_score DOUBLE PRECISION NOT NULL,
    common_tokens INTEGER NOT NULL,
    timing_score DOUBLE PRECISION NOT NULL,
    position_score DOUBLE PRECISION NOT NULL,
    trade_frequency DOUBLE PRECISION NOT NULL,
    PRIMARY KEY (wallet_address, similar_wallet_address)
);

-- Table des tokens détenus par les wallets
CREATE TABLE IF NOT EXISTS wallet_holdings (
    wallet_address VARCHAR(255) NOT NULL,
    token_address VARCHAR(255) REFERENCES tokens(address),
    token_symbol VARCHAR(50) NOT NULL,
    token_name VARCHAR(255) NOT NULL,
    token_logo TEXT,
    balance DOUBLE PRECISION NOT NULL,
    value DOUBLE PRECISION NOT NULL,
    purchase_price DOUBLE PRECISION,
    current_price DOUBLE PRECISION,
    profit_loss DOUBLE PRECISION,
    profit_loss_percent DOUBLE PRECISION,
    entry_timestamp TIMESTAMP WITH TIME ZONE,
    last_update_time TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    PRIMARY KEY (wallet_address, token_address)
);

-- Table des token traders
CREATE TABLE IF NOT EXISTS token_traders (
    wallet_address VARCHAR(255) NOT NULL,
    token_address VARCHAR(255) REFERENCES tokens(address),
    relative_volume DOUBLE PRECISION NOT NULL,
    early_investor DOUBLE PRECISION NOT NULL,
    transaction_count INTEGER NOT NULL,
    PRIMARY KEY (wallet_address, token_address)
);

-- Table des influenceurs de tokens
CREATE TABLE IF NOT EXISTS token_influencers (
    wallet_address VARCHAR(255) NOT NULL,
    token_address VARCHAR(255) REFERENCES tokens(address),
    influence_score DOUBLE PRECISION NOT NULL,
    volume_impact DOUBLE PRECISION NOT NULL,
    timing_impact DOUBLE PRECISION NOT NULL,
    price_impact DOUBLE PRECISION NOT NULL,
    transaction_count INTEGER NOT NULL,
    PRIMARY KEY (wallet_address, token_address)
);

-- Index pour les performances
CREATE INDEX IF NOT EXISTS idx_wallet_interactions_wallet ON wallet_interactions(wallet_address);
CREATE INDEX IF NOT EXISTS idx_wallet_interactions_token ON wallet_interactions(token_address);
CREATE INDEX IF NOT EXISTS idx_wallet_interactions_timestamp ON wallet_interactions(timestamp DESC);
CREATE INDEX IF NOT EXISTS idx_wallet_interactions_action_type ON wallet_interactions(action_type);
CREATE INDEX IF NOT EXISTS idx_token_trades_token ON token_trades(token_address);
CREATE INDEX IF NOT EXISTS idx_token_trades_wallet ON token_trades(wallet_address);
CREATE INDEX IF NOT EXISTS idx_token_trades_timestamp ON token_trades(timestamp DESC);
CREATE INDEX IF NOT EXISTS idx_token_metrics_updated ON token_metrics(updated_at DESC);
CREATE INDEX IF NOT EXISTS idx_wallet_trust_scores_score ON wallet_trust_scores(trust_score DESC);

-- Vues pour les requêtes fréquentes
CREATE OR REPLACE VIEW token_recent_metrics AS
SELECT DISTINCT ON (token_address) *
FROM token_metrics
ORDER BY token_address, updated_at DESC;

CREATE OR REPLACE VIEW active_wallets AS
SELECT DISTINCT wallet_address, MAX(timestamp) as last_active
FROM wallet_interactions
GROUP BY wallet_address;

CREATE OR REPLACE VIEW token_active_wallets AS
SELECT token_address, COUNT(DISTINCT wallet_address) as active_wallets_count
FROM wallet_interactions
WHERE timestamp > (NOW() - INTERVAL '7 days')
GROUP BY token_address; 