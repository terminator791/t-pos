package main

import (
	"log"

	"github.com/terminator791/t-pos/internal/infrastructure/config"
	"github.com/terminator791/t-pos/internal/infrastructure/database"
	"github.com/terminator791/t-pos/internal/infrastructure/logger"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize logger
	appLogger, err := logger.NewLogger(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}

	appLogger.Info("Starting database migration...")

	// Initialize database connection
	db, err := database.NewConnection(cfg)
	if err != nil {
		appLogger.Error("Failed to connect to database: %v", err)
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Run database migrations
	if err := db.Migrate(); err != nil {
		appLogger.Error("Failed to run database migrations: %v", err)
		log.Fatalf("Failed to run database migrations: %v", err)
	}

	appLogger.Info("Database migrations completed successfully")
}