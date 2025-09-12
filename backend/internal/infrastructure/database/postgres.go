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

	// First, migrate base entities (no dependencies)
	err := DB.AutoMigrate(
		&entities.License{},
		&entities.Role{},
		&entities.Policy{},
	)
	if err != nil {
		return fmt.Errorf("failed to migrate base entities: %w", err)
	}

	// Create User table without Shop foreign key constraint to avoid circular dependency
	err = createUserTableWithoutShopFK()
	if err != nil {
		return fmt.Errorf("failed to create users table: %w", err)
	}

	// Migrate remaining entities that depend on User and License
	err = DB.AutoMigrate(
		&entities.LicenseLog{},
		&entities.UserDomain{},
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
	)
	if err != nil {
		return fmt.Errorf("failed to migrate remaining entities: %w", err)
	}

	// Now add missing foreign key constraints for User-Shop relationship
	err = addUserShopConstraint()
	if err != nil {
		log.Printf("Warning: Failed to add User-Shop foreign key constraint: %v", err)
	}

	log.Println("Database migration completed successfully")
	return nil
}

// createUserTableWithoutShopFK creates the users table without the shop_id foreign key constraint
func createUserTableWithoutShopFK() error {
	createSQL := `
	CREATE TABLE IF NOT EXISTS "users" (
		"id" uuid DEFAULT uuid_generate_v4(),
		"license_id" uuid,
		"role_id" uuid,
		"shop_id" uuid,
		"email" varchar(255),
		"email_verified_at" timestamptz,
		"username" varchar(255) NOT NULL,
		"name" varchar(255),
		"password" varchar(255) NOT NULL,
		"pin" varchar(255),
		"info_device" varchar(255),
		"fcm_token" varchar(255),
		"remember_token" varchar(100),
		"created_at" timestamptz,
		"updated_at" timestamptz,
		"deleted_at" timestamptz,
		PRIMARY KEY ("id"),
		CONSTRAINT "fk_users_role" FOREIGN KEY ("role_id") REFERENCES "roles"("id") ON DELETE SET NULL,
		CONSTRAINT "fk_licenses_users" FOREIGN KEY ("license_id") REFERENCES "licenses"("id") ON DELETE CASCADE
	)`

	err := DB.Exec(createSQL).Error
	if err != nil {
		return err
	}

	// Create indexes
	indexes := []string{
		`CREATE INDEX IF NOT EXISTS "idx_users_deleted_at" ON "users" ("deleted_at")`,
		`CREATE UNIQUE INDEX IF NOT EXISTS "idx_users_email" ON "users" ("email")`,
		`CREATE UNIQUE INDEX IF NOT EXISTS "idx_users_username" ON "users" ("username")`,
	}

	for _, indexSQL := range indexes {
		if err := DB.Exec(indexSQL).Error; err != nil {
			log.Printf("Warning: Failed to create index: %v", err)
		}
	}

	return nil
}

// addUserShopConstraint adds the foreign key constraint between users and shops
func addUserShopConstraint() error {
	constraintSQL := `
	ALTER TABLE users 
	ADD CONSTRAINT IF NOT EXISTS fk_users_assigned_shop 
	FOREIGN KEY (shop_id) REFERENCES shops(id) ON DELETE SET NULL`

	return DB.Exec(constraintSQL).Error
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

	// Drop tables in reverse dependency order to respect foreign key constraints
	tables := []string{
		"casbin_rule", // Casbin table

		// History and logs (depends on Shop and User)
		"logs",
		"stock_histories",
		"expenses",
		"histories",

		// Transaction related entities (depends on User, Shop, Product)
		"receipts",
		"payments",
		"transaction_products",
		"transactions",
		"carts",

		// Product related entities (depends on Shop)
		"products",
		"categories",

		// Shop entities (depends on License and User)
		"shops",

		// User entities (depends on License and Role)
		"user_domains",
		"license_logs",
		"users",

		// Base entities (no dependencies)
		"policies",
		"roles",
		"licenses",
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
