# Crypto Oracle Memecoin Detector

A sophisticated system for detecting, filtering, and monitoring promising memecoin tokens on the Solana blockchain, with a focus on early detection and smart money analysis.

## Overview

Crypto Oracle is designed to identify tokens that have recently been "completed" (migrated to an AMM like Raydium or Jupiter) and analyze them using a multi-stage filtering process and a proprietary X-Score algorithm. The system tracks wallet behaviors, identifies smart money patterns, and generates alerts for tokens with high potential.

## Key Features

- **Automated Token Detection**: Constantly monitors for newly completed tokens on Solana
- **Sophisticated Filtering**: Multi-level filtering based on market cap, holders, creator balance, etc.
- **Smart Money Analysis**: Identifies and tracks wallets with consistent successful trading patterns
- **X-Score Algorithm**: Proprietary scoring system to evaluate token potential based on multiple factors
- **Memory of Trust**: Network of trust that tracks relationships between wallets and tokens over time
- **Lifecycle Management**: Tokens progress through different states based on their potential and activity
- **Reactivation Detection**: Monitors dormant tokens for signs of renewed activity
- **Notification System**: Configurable alerts for different token states and events

## System Components

- **Token Engine**: Core system for token detection, analysis, and lifecycle management
- **Memory of Trust**: Tracks wallet interactions and calculates trust scores
- **GMGN Adapter**: Connects to the GMGN API to retrieve token and wallet data
- **Notification Pipeline**: Routes alerts to configured destinations (console, Telegram, webhooks)
- **Filtering System**: Applies strict criteria to identify high-quality tokens
- **X-Score Calculation**: Computes potential score based on multiple weighted factors

## System Architecture

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│  GMGN Adapter   │───▶│   Token Engine   │───▶│  Notifications  │
└─────────────────┘    └────────┬────────┘    └─────────────────┘
                               │
                      ┌────────▼────────┐
                      │  Memory of Trust │
                      └────────┬────────┘
                               │
                      ┌────────▼────────┐
                      │  Storage/Cache  │
                      └─────────────────┘
```

## Token Lifecycle States

| State           | Description                        | TTL    | Polling Interval |
|-----------------|------------------------------------| -------|------------------|
| DISCOVERED      | Initial scan completed             | 6h     | 15min            |
| VALIDATED       | X-Score > 60, promising            | 24h    | 5min             |
| HYPED           | X-Score > 80, high priority        | 48h    | 1min             |
| SLEEP_MODE      | Reduced activity                   | 30d    | 1h               |
| MONITORING_LIGHT| Potential for reactivation         | 30d    | 15min            |
| REACTIVATED     | Renewed activity                   | 48h    | 1min             |

## X-Score Components

The X-Score is a weighted combination of these factors:

- **Token Quality (20%)**: Market cap, holder count, creator balance, top holder concentration
- **Wallet Quality (25%)**: Smart money ratio, trusted wallets, bot ratio, fresh wallet rate
- **Trust Factor (20%)**: Average trust score, early trusted wallets, smart money presence
- **Market Dynamics (15%)**: Volume, price change, buy/sell ratio, holder growth
- **Temporal Patterns (10%)**: Time-based activity patterns
- **Reactivation Factor (10%)**: Signs of renewed activity after dormancy
- **Special Bonuses**: Sniper wallet presence, smart money combined with price increase

## Installation

1. Clone the repository:
```
git clone https://github.com/franky69420/crypto-oracle.git
cd crypto-oracle
```

2. Install dependencies:
```
go mod download
```

3. Create configuration file:
```
cp config.example.yaml config.yaml
```

4. Edit the configuration file with your settings

## Usage

### Live Monitoring Mode

```
go run cmd/detector/detector.go --mode=live
```

### Token Scanner

The Token Scanner is a simpler version of the Crypto Oracle detector that focuses specifically on detecting newly completed tokens and tracking their metrics. It's useful for testing the token detection engine without requiring a full database setup.

To run the token scanner:

```
# Using make
make run-token-scan

# Or directly with Go
go run cmd/token-scan/main.go
```

To build the token scanner binary:

```
make build-token-scan
```

The token scanner connects directly to the GMGN API and implements simplified versions of the Memory of Trust and notification components.

### Configuration Options

```
--config string      Path to configuration file
--log-level string   Log level (debug, info, warn, error)
--console            Output logs to console
--log-file string    Output logs to file
--mode string        Run mode (live, backtest)
```

## Requirements

- Go 1.16+
- Redis
- PostgreSQL
- GMGN API access

## License

Copyright © 2023 Franky69420

This software is proprietary and confidential. Unauthorized copying or distribution of this software, via any medium, is strictly prohibited. 