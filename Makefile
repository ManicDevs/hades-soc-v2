# Hades-V2 Enterprise Security Framework Makefile

.PHONY: help build test clean docker docker-build docker-run install uninstall dev prod check-secrets

# Variables
BINARY_NAME=hades
VERSION=$(shell git describe --tags --always --dirty)
BUILD_TIME=$(shell date -u '+%Y-%m-%d_%H:%M:%S')
LDFLAGS=-ldflags "-X main.version=$(VERSION) -X main.buildTime=$(BUILD_TIME)"
DOCKER_IMAGE=hades-v2
DOCKER_TAG=latest

# Help
help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-20s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

# Build
build: ## Build the binary
	@echo "Building $(BINARY_NAME)..."
	go build $(LDFLAGS) -o $(BINARY_NAME) ./cmd/hades

build-all: ## Build for all platforms
	@echo "Building for all platforms..."
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(BINARY_NAME)-linux-amd64 ./cmd/hades
	GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o $(BINARY_NAME)-linux-arm64 ./cmd/hades
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o $(BINARY_NAME)-darwin-amd64 ./cmd/hades
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o $(BINARY_NAME)-darwin-arm64 ./cmd/hades
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o $(BINARY_NAME)-windows-amd64.exe ./cmd/hades

# Development
dev: ## Run in development mode
	@echo "Starting development server..."
	./$(BINARY_NAME) web start --port 8443 --dev

dev-api: ## Run API server in development mode
	@echo "Starting API server..."
	./$(BINARY_NAME) auxiliary start api_server_fixed --port 8080 --token dev-token

# Testing
test: ## Run tests
	@echo "Running tests..."
	go test -v ./...

test-coverage: ## Run tests with coverage
	@echo "Running tests with coverage..."
	go test -cover -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

test-race: ## Run tests with race detector
	@echo "Running tests with race detector..."
	go test -race -v ./...

bench: ## Run benchmarks
	@echo "Running benchmarks..."
	go test -bench=. -benchmem ./...

# Quality
lint: ## Run linter
	@echo "Running linter..."
	golangci-lint run

fmt: ## Format code
	@echo "Formatting code..."
	go fmt ./...

vet: ## Run go vet
	@echo "Running go vet..."
	go vet ./...

mod-tidy: ## Tidy go modules
	@echo "Tidying modules..."
	go mod tidy

mod-verify: ## Verify go modules
	@echo "Verifying modules..."
	go mod verify

# Docker
docker-build: ## Build Docker image
	@echo "Building Docker image..."
	docker build -t $(DOCKER_IMAGE):$(DOCKER_TAG) .

docker-run: ## Run Docker container
	@echo "Running Docker container..."
	docker-compose up -d

docker-stop: ## Stop Docker containers
	@echo "Stopping Docker containers..."
	docker-compose down

docker-logs: ## Show Docker logs
	@echo "Showing Docker logs..."
	docker-compose logs -f

docker-clean: ## Clean Docker resources
	@echo "Cleaning Docker resources..."
	docker-compose down -v
	docker system prune -f

# Database
db-init: ## Initialize database
	@echo "Initializing database..."
	./$(BINARY_NAME) migrate init

db-migrate: ## Run database migrations
	@echo "Running migrations..."
	./$(BINARY_NAME) migrate up

db-reset: ## Reset database
	@echo "Resetting database..."
	./$(BINARY_NAME) migrate reset --force

# Configuration
config-wizard: ## Run configuration wizard
	@echo "Running configuration wizard..."
	./$(BINARY_NAME) config wizard

config-validate: ## Validate configuration
	@echo "Validating configuration..."
	./$(BINARY_NAME) config validate

# Users
user-admin: ## Create admin user
	@echo "Creating admin user..."
	./$(BINARY_NAME) user create --username admin --email admin@localhost --role admin --password admin123

# Installation
install: build ## Install binary to system
	@echo "Installing $(BINARY_NAME)..."
	sudo cp $(BINARY_NAME) /usr/local/bin/
	sudo chmod +x /usr/local/bin/$(BINARY_NAME)

uninstall: ## Uninstall binary from system
	@echo "Uninstalling $(BINARY_NAME)..."
	sudo rm -f /usr/local/bin/$(BINARY_NAME)

# Production
prod: build docker-build ## Build for production
	@echo "Production build completed"

prod-deploy: ## Deploy to production
	@echo "Deploying to production..."
	docker-compose -f docker-compose.prod.yml up -d

# Cleanup
clean: ## Clean build artifacts
	@echo "Cleaning..."
	rm -f $(BINARY_NAME)
	rm -f $(BINARY_NAME)-*
	rm -f coverage.out coverage.html
	go clean -cache

clean-all: clean docker-clean ## Clean everything
	@echo "All clean!"

# Quick start
quick-start: build config-wizard db-init db-migrate user-admin ## Quick start setup
	@echo "Quick start completed!"
	@echo "Run 'make dev' to start the development server"
	@echo "Access the dashboard at http://localhost:8443"

# CI/CD
ci: fmt vet lint test-coverage ## Run CI checks
	@echo "CI checks completed"

# Release
release: clean build-all ## Build release binaries
	@echo "Release binaries built:"
	@ls -la $(BINARY_NAME)-*

# Security
security-scan: ## Run security scan
	@echo "Running security scan..."
	gosec ./...

check-secrets: ## Scan repository for secrets
	@echo "🔍 Scanning repository for secrets..."
	@echo "Scanning for JWT_SECRET, PASSWORD, *_KEY, *_TOKEN, TAILSCALE_AUTHKEY patterns..."
	@SECRETS_FOUND=false; \
	for file in $$(find . -name "*.go" -not -path "./vendor/*" -not -path "./.git/*"); do \
		if grep -E "JWT_SECRET=[^[:space:]]+[^[:space:]]|PASSWORD=[^[:space:]]+[^[:space:]]|.*_KEY=[^[:space:]]+[^[:space:]]|.*_TOKEN=[^[:space:]]+[^[:space:]]|TAILSCALE_AUTHKEY=[^[:space:]]+[^[:space:]]" "$$file" >/dev/null 2>&1; then \
			echo "❌ SECRET DETECTED in $$file"; \
			SECRETS_FOUND=true; \
		fi; \
	done; \
	for file in $$(find . -name "*.yml" -o -name "*.yaml" -o -name "*.json" -o -name "*.env*" -not -name "*.env.*.example" -not -path "./vendor/*" -not -path "./.git/*"); do \
		if grep -E "JWT_SECRET=[^[:space:]]+[^[:space:]]|PASSWORD=[^[:space:]]+[^[:space:]]|.*_KEY=[^[:space:]]+[^[:space:]]|.*_TOKEN=[^[:space:]]+[^[:space:]]|TAILSCALE_AUTHKEY=[^[:space:]]+[^[:space:]]" "$$file" >/dev/null 2>&1; then \
			echo "❌ SECRET DETECTED in $$file"; \
			SECRETS_FOUND=true; \
		fi; \
	done; \
	for file in $$(find ./web -name "*.js" -o -name "*.jsx" -o -name "*.ts" -o -name "*.tsx" -o -name "*.json"); do \
		if grep -E "JWT_SECRET=[^[:space:]]+[^[:space:]]|PASSWORD=[^[:space:]]+[^[:space:]]|.*_KEY=[^[:space:]]+[^[:space:]]|.*_TOKEN=[^[:space:]]+[^[:space:]]|TAILSCALE_AUTHKEY=[^[:space:]]+[^[:space:]]" "$$file" >/dev/null 2>&1; then \
			echo "❌ SECRET DETECTED in $$file"; \
			SECRETS_FOUND=true; \
		fi; \
	done; \
	if [ "$$SECRETS_FOUND" = true ]; then \
		echo "❌ SECRETS DETECTED! Please remove or replace with placeholder values."; \
		exit 1; \
	else \
		echo "✅ No secrets detected in repository."; \
		exit 0; \
	fi

deps-check: ## Check for vulnerable dependencies
	@echo "Checking dependencies..."
	go list -json -m all | nancy sleuth

# Monitoring
health: ## Check system health
	@echo "Checking system health..."
	curl -f http://localhost:8443/api/health || echo "Service not running"

logs: ## Show application logs
	@echo "Showing logs..."
	tail -f logs/hades.log

# Backup
backup: ## Backup data
	@echo "Creating backup..."
	mkdir -p backups
	tar -czf backups/hades-backup-$(shell date +%Y%m%d-%H%M%S).tar.gz data/ config/

# Default
.DEFAULT_GOAL := help
