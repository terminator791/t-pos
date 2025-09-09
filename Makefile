.PHONY: run run-backend run-frontend build-backend test-backend clean help migrate-up migrate-down migrate-fresh migrate-status

# Default target
help:
	@echo "T-POS Development Commands:"
	@echo "  make run            - Run both backend and frontend concurrently"
	@echo "  make run-backend    - Run only the backend server"
	@echo "  make run-frontend   - Run only the frontend development server"
	@echo "  make build-backend  - Build the backend application"
	@echo "  make test-backend   - Run backend tests"
	@echo "  make clean          - Clean build artifacts"
	@echo ""
	@echo "Database Migration Commands:"
	@echo "  make migrate-up     - Run pending migrations"
	@echo "  make migrate-down   - Drop all tables (rollback)"
	@echo "  make migrate-fresh  - Drop all tables and re-run migrations"
	@echo "  make migrate-status - Check migration status"
	@echo "  make seed           - Run database seeders"

# Run both backend and frontend concurrently
run:
	@echo "Starting T-POS development environment..."
	@make -j2 run-backend run-frontend

# Run backend server
run-backend:
	@echo "Starting T-POS backend server..."
	@cd backend && go run cmd/main.go

# Run frontend development server
run-frontend:
	@echo "Starting T-POS frontend development server..."
	@cd frontend && yarn dev

# Build backend application
build-backend:
	@echo "Building T-POS backend..."
	@cd backend && go build -o bin/tpos cmd/main.go

# Run backend tests
test-backend:
	@echo "Running T-POS backend tests..."
	@cd backend && go test ./...

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	@cd backend && rm -rf bin/
	@cd frontend && rm -rf dist/

# Install dependencies
deps:
	@echo "Installing dependencies..."
	@cd backend && go mod tidy
	@cd frontend && yarn install

# Database setup (requires PostgreSQL running)
db-setup:
	@echo "Setting up database..."
	@createdb tpos_db || echo "Database might already exist"

# Docker commands (if you want to add Docker later)
docker-build:
	@echo "Building Docker containers..."
	@docker-compose build

docker-up:
	@echo "Starting Docker containers..."
	@docker-compose up -d

docker-down:
	@echo "Stopping Docker containers..."
	@docker-compose down

# Database Migration Commands
migrate-up:
	@echo "Running database migrations..."
	@cd backend && go run cmd/migrate/main.go up

migrate-down:
	@echo "Rolling back database migrations..."
	@cd backend && go run cmd/migrate/main.go down

migrate-fresh:
	@echo "Refreshing database (drop + migrate)..."
	@cd backend && go run cmd/migrate/main.go fresh

migrate-status:
	@echo "Checking migration status..."
	@cd backend && go run cmd/migrate/main.go status

seed:
	@echo "Running database seeders..."
	@cd backend && go run cmd/migrate/main.go seed