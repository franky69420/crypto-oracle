.PHONY: build run clean test docker-build docker-run docker-up docker-down build-token-scan run-token-scan

# Application name
APP_NAME = crypto-oracle

# Main variables
GO = go
DOCKER = docker
DOCKER_COMPOSE = docker-compose
CONFIG_PATH = config/dev.yaml

# Build flags
BUILD_FLAGS = -ldflags="-s -w"

# Build the application
build:
	$(GO) build $(BUILD_FLAGS) -o bin/$(APP_NAME) ./cmd/detector/detector.go

# Run the application
run:
	$(GO) run ./cmd/detector/detector.go -config $(CONFIG_PATH) -log-level debug

# Build token scan application
build-token-scan:
	$(GO) build $(BUILD_FLAGS) -o bin/token-scan ./cmd/token-scan/main.go

# Run token scan application
run-token-scan:
	$(GO) run ./cmd/token-scan/main.go -log-level debug

# Clean build artifacts
clean:
	rm -rf bin/

# Run tests
test:
	$(GO) test -v ./...

# Build docker image
docker-build:
	$(DOCKER) build -t crypto-oracle:latest .

# Run docker image
docker-run:
	$(DOCKER) run --rm -p 3000:3000 crypto-oracle:latest

# Start all services with docker-compose
docker-up:
	$(DOCKER_COMPOSE) up -d

# Stop all services
docker-down:
	$(DOCKER_COMPOSE) down

# Start only dependencies (Redis and PostgreSQL)
docker-deps:
	$(DOCKER_COMPOSE) up -d postgres redis

# Apply database schema
db-schema:
	psql -h localhost -U crypto_oracle -d crypto_oracle -f schema.sql

# Generate mock data for testing
db-mock:
	$(GO) run ./cmd/tools/mockdata.go

# Show logs
logs:
	$(DOCKER_COMPOSE) logs -f

# Start complete dev environment and show logs
dev: docker-deps
	sleep 5
	$(MAKE) db-schema
	$(MAKE) run 