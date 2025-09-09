package database

import (
	"fmt"
	"log"

	"github.com/terminator791/t-pos/config"
	"github.com/terminator791/t-pos/internal/domain/entities"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// DB is the global database instance
var DB *gorm.DB

// Connect establishes database connection
func Connect(cfg *config.DatabaseConfig) error {
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.Name, cfg.SSLMode)

	var err error
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})

	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	log.Println("Database connected successfully")
	return nil
}

// Migrate runs database migrations
func Migrate() error {
	if DB == nil {
		return fmt.Errorf("database not connected")
	}

	err := DB.AutoMigrate(
		&entities.License{},
		&entities.User{},
		&entities.Role{},
		&entities.UserRole{},
		&entities.Policy{},
		&entities.LicenseLog{},
		&entities.Shop{},
		&entities.Category{},
		&entities.Product{},
		&entities.Cart{},
		&entities.Transaction{},
		&entities.TransactionProduct{},
		&entities.Payment{},
		&entities.Receipt{},
		&entities.History{},
		&entities.StockHistory{},
		&entities.Expense{},
		&entities.Log{},
		&entities.UserDomain{},
	)

	if err != nil {
		return fmt.Errorf("failed to migrate database: %w", err)
	}

	log.Println("Database migration completed successfully")
	return nil
}

// Close closes the database connection
func Close() error {
	if DB == nil {
		return nil
	}

	sqlDB, err := DB.DB()
	if err != nil {
		return err
	}

	return sqlDB.Close()
}

// GetDB returns the database instance
func GetDB() *gorm.DB {
	return DB
}

// DropAllTables drops all tables in the database (like migrate:fresh)
func DropAllTables() error {
	if DB == nil {
		return fmt.Errorf("database not connected")
	}

	// Get all table names
	tables := []string{
		"casbin_rule", // Casbin table
		"logs",
		"stock_histories", 
		"expenses",
		"histories",
		"receipts",
		"payments",
		"transaction_products",
		"transactions",
		"carts",
		"products",
		"categories",
		"shops",
		"policies",
		"user_roles",
		"roles",
		"license_logs",
		"users",
		"licenses",
		"user_domains",
	}

	// Drop tables in reverse order to respect foreign key constraints
	for _, table := range tables {
		if err := DB.Exec(fmt.Sprintf("DROP TABLE IF EXISTS %s CASCADE;", table)).Error; err != nil {
			log.Printf("Warning: Failed to drop table %s: %v", table, err)
		} else {
			log.Printf("Dropped table: %s", table)
		}
	}

	log.Println("All tables dropped successfully")
	return nil
}

// RefreshDatabase drops all tables and re-runs migrations (like migrate:fresh)
func RefreshDatabase() error {
	log.Println("Refreshing database...")
	
	if err := DropAllTables(); err != nil {
		return fmt.Errorf("failed to drop tables: %w", err)
	}

	if err := Migrate(); err != nil {
		return fmt.Errorf("failed to migrate after drop: %w", err)
	}

	log.Println("Database refreshed successfully")
	return nil
}

// MigrateDown simulates rolling back migrations by dropping and recreating
func MigrateDown() error {
	log.Println("Rolling back migrations...")
	return DropAllTables()
}

// MigrateUp runs the migrations
func MigrateUp() error {
	log.Println("Running migrations...")
	return Migrate()
}