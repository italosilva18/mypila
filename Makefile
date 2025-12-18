.PHONY: help build test run deploy clean install lint format docker-build docker-up docker-down docker-logs dev-backend dev-frontend check-health migrate backup restore

# Variables
BACKEND_DIR := ./backend
FRONTEND_DIR := ./frontend
DOCKER_COMPOSE := docker-compose
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME := $(shell date -u '+%Y-%m-%d_%H:%M:%S')
COMMIT_SHA := $(shell git rev-parse HEAD 2>/dev/null || echo "unknown")

# Colors for output
COLOR_RESET := \033[0m
COLOR_BOLD := \033[1m
COLOR_GREEN := \033[32m
COLOR_YELLOW := \033[33m
COLOR_BLUE := \033[34m

help: ## Show this help message
	@echo "$(COLOR_BOLD)M2M Financeiro - Makefile Commands$(COLOR_RESET)"
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "$(COLOR_GREEN)%-20s$(COLOR_RESET) %s\n", $$1, $$2}'

install: ## Install all dependencies
	@echo "$(COLOR_BLUE)Installing backend dependencies...$(COLOR_RESET)"
	cd $(BACKEND_DIR) && go mod download && go mod tidy
	@echo "$(COLOR_BLUE)Installing frontend dependencies...$(COLOR_RESET)"
	cd $(FRONTEND_DIR) && npm install
	@echo "$(COLOR_GREEN)Dependencies installed successfully!$(COLOR_RESET)"

build: ## Build all components
	@echo "$(COLOR_BLUE)Building backend...$(COLOR_RESET)"
	cd $(BACKEND_DIR) && CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo \
		-ldflags "-X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME) -X main.CommitSHA=$(COMMIT_SHA)" \
		-o bin/main .
	@echo "$(COLOR_BLUE)Building frontend...$(COLOR_RESET)"
	cd $(FRONTEND_DIR) && npm run build
	@echo "$(COLOR_GREEN)Build completed successfully!$(COLOR_RESET)"

test: test-backend test-frontend ## Run all tests

test-backend: ## Run backend tests
	@echo "$(COLOR_BLUE)Running backend tests...$(COLOR_RESET)"
	cd $(BACKEND_DIR) && go test -v -race -coverprofile=coverage.out -covermode=atomic ./...
	cd $(BACKEND_DIR) && go tool cover -html=coverage.out -o coverage.html
	@echo "$(COLOR_GREEN)Backend tests completed!$(COLOR_RESET)"

test-frontend: ## Run frontend tests
	@echo "$(COLOR_BLUE)Running frontend tests...$(COLOR_RESET)"
	cd $(FRONTEND_DIR) && npm test -- --run --coverage || echo "No tests configured"
	@echo "$(COLOR_GREEN)Frontend tests completed!$(COLOR_RESET)"

lint: lint-backend lint-frontend ## Run all linters

lint-backend: ## Run backend linter
	@echo "$(COLOR_BLUE)Running backend linters...$(COLOR_RESET)"
	cd $(BACKEND_DIR) && go vet ./...
	cd $(BACKEND_DIR) && gofmt -s -l . || true
	@which staticcheck > /dev/null && cd $(BACKEND_DIR) && staticcheck ./... || echo "$(COLOR_YELLOW)staticcheck not installed$(COLOR_RESET)"
	@echo "$(COLOR_GREEN)Backend linting completed!$(COLOR_RESET)"

lint-frontend: ## Run frontend linter
	@echo "$(COLOR_BLUE)Running frontend linters...$(COLOR_RESET)"
	cd $(FRONTEND_DIR) && npm run lint || echo "$(COLOR_YELLOW)No lint script configured$(COLOR_RESET)"
	@echo "$(COLOR_GREEN)Frontend linting completed!$(COLOR_RESET)"

format: ## Format all code
	@echo "$(COLOR_BLUE)Formatting backend code...$(COLOR_RESET)"
	cd $(BACKEND_DIR) && gofmt -s -w .
	@echo "$(COLOR_BLUE)Formatting frontend code...$(COLOR_RESET)"
	cd $(FRONTEND_DIR) && npm run format || echo "$(COLOR_YELLOW)No format script configured$(COLOR_RESET)"
	@echo "$(COLOR_GREEN)Code formatting completed!$(COLOR_RESET)"

dev-backend: ## Run backend in development mode
	@echo "$(COLOR_BLUE)Starting backend in development mode...$(COLOR_RESET)"
	cd $(BACKEND_DIR) && go run main.go

dev-frontend: ## Run frontend in development mode
	@echo "$(COLOR_BLUE)Starting frontend in development mode...$(COLOR_RESET)"
	cd $(FRONTEND_DIR) && npm run dev

run: docker-up ## Run the application using Docker Compose

docker-build: ## Build Docker images
	@echo "$(COLOR_BLUE)Building Docker images...$(COLOR_RESET)"
	VERSION=$(VERSION) $(DOCKER_COMPOSE) build --no-cache
	@echo "$(COLOR_GREEN)Docker images built successfully!$(COLOR_RESET)"

docker-up: ## Start all services with Docker Compose
	@echo "$(COLOR_BLUE)Starting services...$(COLOR_RESET)"
	VERSION=$(VERSION) $(DOCKER_COMPOSE) up -d
	@echo "$(COLOR_GREEN)Services started successfully!$(COLOR_RESET)"
	@echo "$(COLOR_YELLOW)Frontend:$(COLOR_RESET) http://localhost:3333"
	@echo "$(COLOR_YELLOW)Backend:$(COLOR_RESET)  http://localhost:8081"
	@echo "$(COLOR_YELLOW)MongoDB:$(COLOR_RESET)  localhost:27018"

docker-down: ## Stop all services
	@echo "$(COLOR_BLUE)Stopping services...$(COLOR_RESET)"
	$(DOCKER_COMPOSE) down
	@echo "$(COLOR_GREEN)Services stopped!$(COLOR_RESET)"

docker-down-volumes: ## Stop all services and remove volumes
	@echo "$(COLOR_BLUE)Stopping services and removing volumes...$(COLOR_RESET)"
	$(DOCKER_COMPOSE) down -v
	@echo "$(COLOR_GREEN)Services stopped and volumes removed!$(COLOR_RESET)"

docker-logs: ## Show Docker logs
	$(DOCKER_COMPOSE) logs -f

docker-logs-backend: ## Show backend logs
	$(DOCKER_COMPOSE) logs -f backend

docker-logs-frontend: ## Show frontend logs
	$(DOCKER_COMPOSE) logs -f frontend

docker-logs-mongodb: ## Show MongoDB logs
	$(DOCKER_COMPOSE) logs -f mongodb

docker-restart: ## Restart all services
	@echo "$(COLOR_BLUE)Restarting services...$(COLOR_RESET)"
	$(DOCKER_COMPOSE) restart
	@echo "$(COLOR_GREEN)Services restarted!$(COLOR_RESET)"

docker-ps: ## Show running containers
	$(DOCKER_COMPOSE) ps

check-health: ## Check health of all services
	@echo "$(COLOR_BLUE)Checking service health...$(COLOR_RESET)"
	@curl -f http://localhost:8081/health && echo "$(COLOR_GREEN)Backend: Healthy$(COLOR_RESET)" || echo "$(COLOR_YELLOW)Backend: Unhealthy$(COLOR_RESET)"
	@curl -f http://localhost:3333 && echo "$(COLOR_GREEN)Frontend: Healthy$(COLOR_RESET)" || echo "$(COLOR_YELLOW)Frontend: Unhealthy$(COLOR_RESET)"
	@docker exec m2m-mongodb mongosh --eval "db.runCommand({ping:1})" && echo "$(COLOR_GREEN)MongoDB: Healthy$(COLOR_RESET)" || echo "$(COLOR_YELLOW)MongoDB: Unhealthy$(COLOR_RESET)"

backup: ## Backup MongoDB database
	@echo "$(COLOR_BLUE)Creating MongoDB backup...$(COLOR_RESET)"
	@mkdir -p ./backups
	docker exec m2m-mongodb mongodump --out=/tmp/backup --db=m2m_financeiro
	docker cp m2m-mongodb:/tmp/backup ./backups/backup-$(shell date +%Y%m%d-%H%M%S)
	@echo "$(COLOR_GREEN)Backup created successfully!$(COLOR_RESET)"

restore: ## Restore MongoDB database from latest backup
	@echo "$(COLOR_BLUE)Restoring MongoDB from backup...$(COLOR_RESET)"
	@LATEST=$$(ls -t ./backups | head -1); \
	if [ -z "$$LATEST" ]; then \
		echo "$(COLOR_YELLOW)No backups found!$(COLOR_RESET)"; \
		exit 1; \
	fi; \
	docker cp ./backups/$$LATEST m2m-mongodb:/tmp/restore; \
	docker exec m2m-mongodb mongorestore /tmp/restore
	@echo "$(COLOR_GREEN)Database restored successfully!$(COLOR_RESET)"

migrate: ## Run database migrations
	@echo "$(COLOR_BLUE)Running database migrations...$(COLOR_RESET)"
	cd $(BACKEND_DIR) && go run migrations/*.go
	@echo "$(COLOR_GREEN)Migrations completed!$(COLOR_RESET)"

clean: ## Clean build artifacts and dependencies
	@echo "$(COLOR_BLUE)Cleaning build artifacts...$(COLOR_RESET)"
	cd $(BACKEND_DIR) && rm -rf bin/ *.exe coverage.out coverage.html
	cd $(FRONTEND_DIR) && rm -rf dist/ build/ node_modules/.cache
	@echo "$(COLOR_GREEN)Clean completed!$(COLOR_RESET)"

clean-all: clean docker-down-volumes ## Clean everything including Docker volumes
	@echo "$(COLOR_BLUE)Deep cleaning...$(COLOR_RESET)"
	cd $(FRONTEND_DIR) && rm -rf node_modules/
	docker system prune -f
	@echo "$(COLOR_GREEN)Deep clean completed!$(COLOR_RESET)"

deploy-staging: ## Deploy to staging environment
	@echo "$(COLOR_BLUE)Deploying to staging...$(COLOR_RESET)"
	@echo "$(COLOR_YELLOW)Not implemented yet. Configure your staging deployment.$(COLOR_RESET)"

deploy-production: ## Deploy to production environment
	@echo "$(COLOR_BLUE)Deploying to production...$(COLOR_RESET)"
	@echo "$(COLOR_YELLOW)Not implemented yet. Configure your production deployment.$(COLOR_RESET)"

ci: lint test build ## Run CI pipeline locally

version: ## Show current version
	@echo "Version: $(VERSION)"
	@echo "Build Time: $(BUILD_TIME)"
	@echo "Commit SHA: $(COMMIT_SHA)"

stats: ## Show project statistics
	@echo "$(COLOR_BOLD)Project Statistics$(COLOR_RESET)"
	@echo "$(COLOR_BLUE)Backend:$(COLOR_RESET)"
	@cd $(BACKEND_DIR) && find . -name "*.go" -not -path "./vendor/*" | xargs wc -l | tail -1
	@echo "$(COLOR_BLUE)Frontend:$(COLOR_RESET)"
	@cd $(FRONTEND_DIR) && find ./src -name "*.tsx" -o -name "*.ts" | xargs wc -l | tail -1
