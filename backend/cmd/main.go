package main

import (
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/terminator791/t-pos/config"
	"github.com/terminator791/t-pos/internal/domain/usecases"
	"github.com/terminator791/t-pos/internal/infrastructure/database"
	"github.com/terminator791/t-pos/internal/infrastructure/repositories"
	"github.com/terminator791/t-pos/internal/interfaces/http/handlers"
	"github.com/terminator791/t-pos/internal/interfaces/http/routes"
)

func main() {
	// Load environment variables
	err := godotenv.Load()
	if err != nil {
		log.Println("Warning: .env file not found, using environment variables")
	}

	// Load configuration
	cfg := config.Load()

	// Connect to database
	err = database.Connect(&cfg.Database)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Run database migrations
	err = database.Migrate()
	if err != nil {
		log.Fatal("Failed to migrate database:", err)
	}

	// Initialize repositories
	db := database.GetDB()
	userRepo := repositories.NewUserRepository(db)
	categoryRepo := repositories.NewCategoryRepository(db)
	productRepo := repositories.NewProductRepository(db)
	shopRepo := repositories.NewShopRepository(db)
	transactionRepo := repositories.NewTransactionRepository(db)
	paymentRepo := repositories.NewPaymentRepository(db)

	// Initialize use cases
	productUseCase := usecases.NewProductUseCase(productRepo, categoryRepo, shopRepo)
	checkoutUseCase := usecases.NewCheckoutUseCase(transactionRepo, productRepo, shopRepo, userRepo, paymentRepo)

	// Initialize handlers
	productHandler := handlers.NewProductHandler(productUseCase)
	checkoutHandler := handlers.NewCheckoutHandler(checkoutUseCase)

	// Initialize Gin router
	router := gin.Default()

	// Add CORS middleware
	router.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
		
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(200)
			return
		}
		
		c.Next()
	})

	// Setup routes
	routes.SetupRoutes(router, productHandler, checkoutHandler)

	// Start server
	serverAddr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	log.Printf("Starting T-POS server on %s", serverAddr)
	log.Fatal(router.Run(serverAddr))
}
