package main

import (
	"fmt"
	"log"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/terminator791/t-pos/config"
	"github.com/terminator791/t-pos/internal/application/services"
	"github.com/terminator791/t-pos/internal/domain/usecases"
	"github.com/terminator791/t-pos/internal/infrastructure/auth"
	"github.com/terminator791/t-pos/internal/infrastructure/casbin"
	"github.com/terminator791/t-pos/internal/infrastructure/database"
	"github.com/terminator791/t-pos/internal/infrastructure/repositories"
	"github.com/terminator791/t-pos/internal/infrastructure/seeders"
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
	roleRepo := repositories.NewRoleRepository(db)
	policyRepo := repositories.NewPolicyRepository(db)
	userDomainRepo := repositories.NewUserDomainRepository(db)
	categoryRepo := repositories.NewCategoryRepository(db)
	productRepo := repositories.NewProductRepository(db)
	shopRepo := repositories.NewShopRepository(db)
	transactionRepo := repositories.NewTransactionRepository(db)
	paymentRepo := repositories.NewPaymentRepository(db)
	cartRepo := repositories.NewCartRepository(db)
	expenseRepo := repositories.NewExpenseRepository(db)
	historyRepo := repositories.NewHistoryRepository(db)
	receiptRepo := repositories.NewReceiptRepository(db)
	stockHistoryRepo := repositories.NewStockHistoryRepository(db)
	licenseRepo := repositories.NewLicenseRepository(db)
	licenseLogRepo := repositories.NewLicenseLogRepository(db)
	transactionProductRepo := repositories.NewTransactionProductRepository(db)

	// Initialize JWT and Password services
	jwtService := auth.NewJWTService(cfg.JWT.Secret, "t-pos", cfg.JWT.ExpiryHour)
	passwordService := auth.NewPasswordService()

	// Initialize Casbin enforcer
	modelPath := filepath.Join("configs", "rbac_model.conf")
	enforcerService, err := casbin.NewEnforcerService(db, modelPath)
	if err != nil {
		log.Fatal("Failed to initialize Casbin enforcer:", err)
	}

	// Initialize middleware
	authMiddleware := auth.NewAuthMiddleware(jwtService, userRepo)
	authzMiddleware := casbin.NewAuthzMiddleware(enforcerService, shopRepo, userRepo)

	// Initialize seeders and seed data
	authSeeder := seeders.NewAuthSeeder(roleRepo, policyRepo, enforcerService)
	if err := authSeeder.SeedAll(); err != nil {
		log.Printf("Warning: Failed to seed auth data: %v", err)
	}

	initialDataSeeder := seeders.NewInitialDataSeeder(
		licenseRepo,
		userRepo,
		roleRepo,
		shopRepo,
		categoryRepo,
		productRepo,
		userDomainRepo,
		enforcerService,
	)
	if err := initialDataSeeder.SeedAll(); err != nil {
		log.Printf("Warning: Failed to seed initial data: %v", err)
	}

	// Initialize use cases
	productUseCase := usecases.NewProductUseCase(productRepo, categoryRepo, shopRepo)
	checkoutUseCase := usecases.NewCheckoutUseCase(transactionRepo, productRepo, shopRepo, userRepo, paymentRepo)
	categoryUseCase := usecases.NewCategoryUseCase(db, categoryRepo, shopRepo)
	cartUseCase := usecases.NewCartUseCase(db, cartRepo, productRepo, userRepo, shopRepo)
	transactionUseCase := usecases.NewTransactionUseCase(db, transactionRepo, productRepo, shopRepo, userRepo, paymentRepo, expenseRepo, historyRepo, receiptRepo)
	shopUseCase := usecases.NewShopUseCase(shopRepo, licenseRepo, userRepo)

	// Initialize services
	licenseService := services.NewLicenseService(licenseRepo, licenseLogRepo, userRepo, db)
	customerService := services.NewCustomerService(userRepo, roleRepo, licenseRepo, db)
	userManagementService := services.NewUserManagementService(userRepo, roleRepo, licenseRepo, db)

	// Initialize sync service
	syncService := services.NewSyncService(
		db,
		cartRepo,
		categoryRepo,
		expenseRepo,
		historyRepo,
		paymentRepo,
		productRepo,
		receiptRepo,
		shopRepo,
		stockHistoryRepo,
		transactionRepo,
		transactionProductRepo,
		userRepo,
		cfg.Sync,
	)

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(userRepo, userDomainRepo, roleRepo, licenseRepo, shopRepo, jwtService, passwordService, enforcerService, authSeeder)
	productHandler := handlers.NewProductHandler(productUseCase)
	checkoutHandler := handlers.NewCheckoutHandler(checkoutUseCase)
	categoryHandler := handlers.NewCategoryHandler(categoryUseCase)
	cartHandler := handlers.NewCartHandler(cartUseCase)
	transactionHandler := handlers.NewTransactionHandler(transactionUseCase)
	expenseHandler := handlers.NewExpenseHandler(expenseRepo)
	paymentHandler := handlers.NewPaymentHandler(paymentRepo)
	historyHandler := handlers.NewHistoryHandler(historyRepo)
	receiptHandler := handlers.NewReceiptHandler(receiptRepo)
	transactionProductHandler := handlers.NewTransactionProductHandler(transactionProductRepo)
	licenseHandler := handlers.NewLicenseHandler(licenseService)
	customerHandler := handlers.NewCustomerHandler(customerService)
	userManagementHandler := handlers.NewUserManagementHandler(userManagementService)
	roleHandler := handlers.NewRoleHandler(roleRepo)
	aclHandler := handlers.NewACLHandler(enforcerService, roleRepo, policyRepo)
	shopHandler := handlers.NewShopHandler(shopUseCase)
	syncHandler := handlers.NewSyncHandler(syncService, userRepo)

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
	routes.SetupRoutes(router, productHandler, checkoutHandler, authHandler, licenseHandler, customerHandler, userManagementHandler, roleHandler, categoryHandler, cartHandler, transactionHandler, expenseHandler, paymentHandler, historyHandler, receiptHandler, transactionProductHandler, aclHandler, shopHandler, syncHandler, authMiddleware, authzMiddleware)

	// Start server
	serverAddr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	log.Printf("Starting T-POS server on %s", serverAddr)
	log.Fatal(router.Run(serverAddr))
}
