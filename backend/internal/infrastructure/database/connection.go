package database

import (
	"fmt"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/terminator791/t-pos/internal/domain/entity"
	"github.com/terminator791/t-pos/internal/infrastructure/config"
)

// Connection holds database connection
type Connection struct {
	DB *gorm.DB
}

// NewConnection creates a new database connection
func NewConnection(cfg *config.Config) (*Connection, error) {
	dsn := cfg.GetDatabaseDSN()

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
		NowFunc: func() time.Time {
			return time.Now().UTC()
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Configure connection pool
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get database instance: %w", err)
	}

	// SetMaxIdleConns sets the maximum number of connections in the idle connection pool.
	sqlDB.SetMaxIdleConns(10)

	// SetMaxOpenConns sets the maximum number of open connections to the database.
	sqlDB.SetMaxOpenConns(100)

	// SetConnMaxLifetime sets the maximum amount of time a connection may be reused.
	sqlDB.SetConnMaxLifetime(time.Hour)

	return &Connection{DB: db}, nil
}

// Close closes the database connection
func (c *Connection) Close() error {
	sqlDB, err := c.DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

// Migrate runs database migrations
func (c *Connection) Migrate() error {
	return c.DB.AutoMigrate(
		&entity.User{},
		&entity.Customer{},
		&entity.Category{},
		&entity.Product{},
		&entity.Order{},
		&entity.OrderItem{},
		&entity.Payment{},
	)
}

// Seed seeds the database with initial data
func (c *Connection) Seed() error {
	// Check if admin user exists
	var adminUser entity.User
	if err := c.DB.Where("email = ?", "admin@tpos.com").First(&adminUser).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			// Create admin user
			admin := entity.User{
				Email:     "admin@tpos.com",
				Username:  "admin",
				Password:  "$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi", // password
				FirstName: "Admin",
				LastName:  "User",
				Role:      entity.RoleAdmin,
				IsActive:  true,
			}
			if err := c.DB.Create(&admin).Error; err != nil {
				return fmt.Errorf("failed to create admin user: %w", err)
			}
		} else {
			return fmt.Errorf("failed to check admin user: %w", err)
		}
	}

	// Check if default categories exist
	var categoryCount int64
	c.DB.Model(&entity.Category{}).Count(&categoryCount)
	if categoryCount == 0 {
		categories := []entity.Category{
			{Name: "Electronics", Description: "Electronic products"},
			{Name: "Clothing", Description: "Clothing and apparel"},
			{Name: "Food & Beverage", Description: "Food and drink products"},
			{Name: "Home & Garden", Description: "Home and garden products"},
			{Name: "Books", Description: "Books and magazines"},
		}
		
		for _, category := range categories {
			if err := c.DB.Create(&category).Error; err != nil {
				return fmt.Errorf("failed to create category %s: %w", category.Name, err)
			}
		}
	}

	return nil
}

// GetDB returns the database instance
func (c *Connection) GetDB() *gorm.DB {
	return c.DB
}

// Ping checks database connectivity
func (c *Connection) Ping() error {
	sqlDB, err := c.DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Ping()
}

// Transaction executes a function within a database transaction
func (c *Connection) Transaction(fn func(*gorm.DB) error) error {
	return c.DB.Transaction(fn)
}