package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// Config holds all configuration for the application
type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	JWT      JWTConfig
	Logger   LoggerConfig
}

// ServerConfig holds server configuration
type ServerConfig struct {
	Host         string
	Port         string
	Mode         string
	ReadTimeout  int
	WriteTimeout int
}

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
	TimeZone string
}

// JWTConfig holds JWT configuration
type JWTConfig struct {
	SecretKey       string
	ExpirationHours int
	RefreshHours    int
}

// LoggerConfig holds logger configuration
type LoggerConfig struct {
	Level    string
	Filename string
	MaxSize  int
	MaxAge   int
	Compress bool
}

// Load loads configuration from environment variables
func Load() (*Config, error) {
	// Load .env file if it exists
	_ = godotenv.Load()

	config := &Config{
		Server: ServerConfig{
			Host:         getEnv("SERVER_HOST", "localhost"),
			Port:         getEnv("SERVER_PORT", "8080"),
			Mode:         getEnv("GIN_MODE", "debug"),
			ReadTimeout:  getEnvAsInt("SERVER_READ_TIMEOUT", 30),
			WriteTimeout: getEnvAsInt("SERVER_WRITE_TIMEOUT", 30),
		},
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5432"),
			User:     getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", "password"),
			DBName:   getEnv("DB_NAME", "tpos"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
			TimeZone: getEnv("DB_TIMEZONE", "UTC"),
		},
		JWT: JWTConfig{
			SecretKey:       getEnv("JWT_SECRET", "your-secret-key"),
			ExpirationHours: getEnvAsInt("JWT_EXPIRATION_HOURS", 24),
			RefreshHours:    getEnvAsInt("JWT_REFRESH_HOURS", 168), // 7 days
		},
		Logger: LoggerConfig{
			Level:    getEnv("LOG_LEVEL", "info"),
			Filename: getEnv("LOG_FILENAME", "logs/app.log"),
			MaxSize:  getEnvAsInt("LOG_MAX_SIZE", 100),
			MaxAge:   getEnvAsInt("LOG_MAX_AGE", 30),
			Compress: getEnvAsBool("LOG_COMPRESS", true),
		},
	}

	return config, nil
}

// GetDatabaseDSN returns database connection string
func (c *Config) GetDatabaseDSN() string {
	return fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=%s",
		c.Database.Host,
		c.Database.User,
		c.Database.Password,
		c.Database.DBName,
		c.Database.Port,
		c.Database.SSLMode,
		c.Database.TimeZone,
	)
}

// GetServerAddress returns server address
func (c *Config) GetServerAddress() string {
	return fmt.Sprintf("%s:%s", c.Server.Host, c.Server.Port)
}

// Helper functions
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvAsBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}