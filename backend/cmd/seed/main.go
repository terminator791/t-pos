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

	appLogger.Info("Starting database seeding...")

	// Initialize database connection
	db, err := database.NewConnection(cfg)
	if err != nil {
		appLogger.Error("Failed to connect to database: %v", err)
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Seed database with initial data
	if err := db.Seed(); err != nil {
		appLogger.Error("Failed to seed database: %v", err)
		log.Fatalf("Failed to seed database: %v", err)
	}

	appLogger.Info("Database seeded successfully")
}