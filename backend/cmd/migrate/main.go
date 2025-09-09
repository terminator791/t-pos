package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
	"github.com/terminator791/t-pos/config"
	"github.com/terminator791/t-pos/internal/infrastructure/casbin"
	"github.com/terminator791/t-pos/internal/infrastructure/database"
	"github.com/terminator791/t-pos/internal/infrastructure/repositories"
	"github.com/terminator791/t-pos/internal/infrastructure/seeders"
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
	defer database.Close()

	// Parse command line arguments
	var command string
	if len(os.Args) > 1 {
		command = os.Args[1]
	} else {
		printUsage()
		return
	}

	// Parse flags
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	flag.Parse()

	// Execute commands
	switch command {
	case "up":
		if err := database.MigrateUp(); err != nil {
			log.Fatal("Migration up failed:", err)
		}
		log.Println("Migration up completed successfully")

	case "down":
		if err := database.MigrateDown(); err != nil {
			log.Fatal("Migration down failed:", err)
		}
		log.Println("Migration down completed successfully")

	case "fresh", "refresh":
		if err := database.RefreshDatabase(); err != nil {
			log.Fatal("Database refresh failed:", err)
		}
		log.Println("Database refresh completed successfully")

	case "drop":
		if err := database.DropAllTables(); err != nil {
			log.Fatal("Drop tables failed:", err)
		}
		log.Println("All tables dropped successfully")

	case "status":
		checkStatus()

	case "seed":
		if err := runSeeder(); err != nil {
			log.Fatal("Seeding failed:", err)
		}
		log.Println("Seeding completed successfully")

	default:
		log.Printf("Unknown command: %s", command)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	log.Println("T-POS Database Migration Tool")
	log.Println("")
	log.Println("Usage:")
	log.Println("  go run cmd/migrate/main.go <command>")
	log.Println("")
	log.Println("Available commands:")
	log.Println("  up       Run pending migrations")
	log.Println("  down     Drop all tables (rollback)")
	log.Println("  fresh    Drop all tables and re-run migrations")
	log.Println("  refresh  Alias for fresh")
	log.Println("  drop     Drop all tables")
	log.Println("  status   Check migration status")
	log.Println("  seed     Run database seeders")
	log.Println("")
	log.Println("Examples:")
	log.Println("  go run cmd/migrate/main.go up")
	log.Println("  go run cmd/migrate/main.go fresh")
	log.Println("  go run cmd/migrate/main.go down")
}

func runSeeder() error {
	db := database.GetDB()
	if db == nil {
		return fmt.Errorf("database not connected")
	}

	// Initialize repositories
	userRepo := repositories.NewUserRepository(db)
	roleRepo := repositories.NewRoleRepository(db)
	policyRepo := repositories.NewPolicyRepository(db)
	userDomainRepo := repositories.NewUserDomainRepository(db)
	categoryRepo := repositories.NewCategoryRepository(db)
	productRepo := repositories.NewProductRepository(db)
	shopRepo := repositories.NewShopRepository(db)
	licenseRepo := repositories.NewLicenseRepository(db)

	// Initialize Casbin enforcer for auth seeder
	modelPath := filepath.Join("configs", "rbac_model.conf")
	enforcerService, err := casbin.NewEnforcerService(db, modelPath)
	if err != nil {
		return fmt.Errorf("failed to initialize Casbin enforcer: %v", err)
	}

	// Run auth seeder first
	authSeeder := seeders.NewAuthSeeder(roleRepo, policyRepo, enforcerService)
	if err := authSeeder.SeedAll(); err != nil {
		return fmt.Errorf("failed to seed auth data: %v", err)
	}

	// Run initial data seeder
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
		return fmt.Errorf("failed to seed initial data: %v", err)
	}

	return nil
}

func checkStatus() {
	db := database.GetDB()
	if db == nil {
		log.Fatal("Database not connected")
	}

	tables := []string{
		"licenses", "users", "roles", "policies", "license_logs", 
		"shops", "categories", "products", "carts", "transactions", "transaction_products", 
		"payments", "receipts", "histories", "stock_histories", "expenses", "logs", 
		"user_domains", "casbin_rule",
	}

	log.Println("Checking table status...")
	log.Println("=========================")

	for _, table := range tables {
		var exists bool
		query := `SELECT EXISTS (
			SELECT FROM information_schema.tables 
			WHERE table_schema = 'public' 
			AND table_name = ?
		)`
		
		if err := db.Raw(query, table).Scan(&exists).Error; err != nil {
			log.Printf("❌ %s: Error checking - %v", table, err)
			continue
		}

		if exists {
			var count int64
			if err := db.Table(table).Count(&count).Error; err != nil {
				log.Printf("✅ %s: Exists (count error: %v)", table, err)
			} else {
				log.Printf("✅ %s: Exists (%d records)", table, count)
			}
		} else {
			log.Printf("❌ %s: Not found", table)
		}
	}
}
