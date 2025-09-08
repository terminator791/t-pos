package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/terminator791/t-pos/internal/delivery/http/router"
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

	appLogger.Info("Starting t-pos backend server...")

	// Initialize database connection
	db, err := database.NewConnection(cfg)
	if err != nil {
		appLogger.Error("Failed to connect to database: %v", err)
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	appLogger.Info("Database connected successfully")

	// Run database migrations
	if err := db.Migrate(); err != nil {
		appLogger.Error("Failed to run database migrations: %v", err)
		log.Fatalf("Failed to run database migrations: %v", err)
	}

	appLogger.Info("Database migrations completed successfully")

	// Seed database with initial data
	if err := db.Seed(); err != nil {
		appLogger.Error("Failed to seed database: %v", err)
		log.Fatalf("Failed to seed database: %v", err)
	}

	appLogger.Info("Database seeded successfully")

	// Test database connection
	if err := db.Ping(); err != nil {
		appLogger.Error("Database ping failed: %v", err)
		log.Fatalf("Database ping failed: %v", err)
	}

	appLogger.Info("Database ping successful")

	// Initialize router
	appRouter := router.NewRouter(cfg, appLogger)
	ginRouter := appRouter.SetupRoutes()

	// Create HTTP server
	server := &http.Server{
		Addr:         cfg.GetServerAddress(),
		Handler:      ginRouter,
		ReadTimeout:  time.Duration(cfg.Server.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(cfg.Server.WriteTimeout) * time.Second,
	}

	// Start server in a goroutine
	go func() {
		appLogger.Info("Server starting on %s", cfg.GetServerAddress())
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			appLogger.Error("Failed to start server: %v", err)
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	appLogger.Info("t-pos backend server started successfully on %s", cfg.GetServerAddress())
	appLogger.Info("API documentation available at http://%s/health", cfg.GetServerAddress())

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	appLogger.Info("Server is shutting down...")

	// Create a deadline for shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Gracefully shutdown the server
	if err := server.Shutdown(ctx); err != nil {
		appLogger.Error("Server forced to shutdown: %v", err)
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	appLogger.Info("Server exited")
}
