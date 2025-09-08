# t-pos Makefile
# Point of Sale System - Development Commands

.PHONY: run run-backend run-frontend build-backend test-backend clean help

# Default target
help: ## Show this help message
	@echo "Available commands:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-15s\033[0m %s\n", $$1, $$2}'

# Run both backend and frontend concurrently
run: ## Run both backend and frontend services concurrently
	@echo "Starting both backend and frontend..."
	@$(MAKE) -j2 run-backend run-frontend

# Run only the backend service
run-backend: ## Compile and run the Golang backend application
	@echo "Starting backend service..."
	@cd backend && go mod tidy && go run cmd/main.go

# Run only the frontend service
run-frontend: ## Navigate to frontend directory and run yarn dev
	@echo "Starting frontend service..."
	@cd frontend && yarn dev

# Build the backend application
build-backend: ## Build the Go backend application
	@echo "Building backend..."
	@cd backend && go mod tidy && go build -o bin/t-pos cmd/main.go

# Test the backend application
test-backend: ## Run backend tests
	@echo "Running backend tests..."
	@cd backend && go test ./...

# Clean build artifacts
clean: ## Clean build artifacts
	@echo "Cleaning build artifacts..."
	@rm -rf backend/bin/
	@rm -rf frontend/dist/
	@rm -rf frontend/node_modules/.cache/

# Setup development environment
setup: ## Setup development environment
	@echo "Setting up development environment..."
	@cd backend && go mod tidy
	@cd frontend && yarn install

# Database operations
db-migrate: ## Run database migrations
	@echo "Running database migrations..."
	@cd backend && go run cmd/migrate/main.go

db-seed: ## Seed database with sample data
	@echo "Seeding database..."
	@cd backend && go run cmd/seed/main.go

# Development helpers
dev-backend: ## Run backend in development mode with hot reload
	@echo "Starting backend in development mode..."
	@cd backend && go run cmd/main.go

logs-backend: ## Show backend logs
	@echo "Showing backend logs..."
	@cd backend && tail -f logs/app.log

# Build for production
build: ## Build both backend and frontend for production
	@echo "Building for production..."
	@$(MAKE) build-backend
	@cd frontend && yarn build