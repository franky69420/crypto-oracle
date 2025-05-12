# Crypto Oracle Memecoin Detector v4.2

A sophisticated AI-powered system for detecting, analyzing, and monitoring cryptocurrency tokens with a focus on memecoins.

## Features

- **Alert Manager**: Token monitoring and alert generation system
- **Pipeline Service**: Data processing pipeline with Redis streams
- **Reactivation System**: Detection of dormant tokens showing signs of activity
- **Token Engine**: X-Score calculation and anti-dump pattern detection
- **Memory of Trust**: Advanced wallet reputation tracking system

## Technologies

- Go (Golang)
- Redis for caching and stream processing
- PostgreSQL for persistent storage
- Docker for containerization

## Getting Started

### Prerequisites

- Go 1.18+
- Docker and Docker Compose
- PostgreSQL
- Redis

### Installation

1. Clone the repository
```bash
git clone https://github.com/franky69420/crypto-oracle.git
cd crypto-oracle
```

2. Start the required services with Docker Compose
```bash
docker-compose up -d
```

3. Build the application
```bash
go build -o crypto-oracle ./cmd/oracle
```

4. Run the application
```bash
./crypto-oracle
```

### Testing

To run the core services test:
```bash
go build -o test-core test-modules.go
./test-core
```

## Project Structure

- `/cmd/oracle`: Entry point for the application
- `/internal`: Internal packages and implementation
  - `/alerting`: Alert management system
  - `/pipeline`: Data processing pipeline
  - `/reactivation`: Token reactivation detection
  - `/storage`: Database and cache implementations
  - `/token`: Token analysis engine
  - `/wallet`: Wallet intelligence system
- `/pkg`: Public packages and models
- `/config`: Configuration files
- `/scripts`: Utility scripts

## License

This project is licensed under the MIT License - see the LICENSE file for details 