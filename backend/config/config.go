package config

import (
	"os"
	"strconv"
	"time"
)

// Config holds application configuration
type Config struct {
	Database DatabaseConfig `json:"database"`
	Server   ServerConfig   `json:"server"`
	JWT      JWTConfig      `json:"jwt"`
	Sync     SyncConfig     `json:"sync"`
}

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
	Name     string `json:"name"`
	SSLMode  string `json:"ssl_mode"`
}

// ServerConfig holds server configuration
type ServerConfig struct {
	Host string `json:"host"`
	Port int    `json:"port"`
}

// JWTConfig holds JWT configuration
type JWTConfig struct {
	Secret     string `json:"secret"`
	ExpiryHour int    `json:"expiry_hour"`
}

// SyncConfig holds configuration for sync optimization
type SyncConfig struct {
	// Batch processing configuration
	BatchSize          int `yaml:"batch_size" json:"batch_size"`
	MaxEntitiesPerSync int `yaml:"max_entities_per_sync" json:"max_entities_per_sync"`

	// Transaction configuration
	TransactionTimeout time.Duration `yaml:"transaction_timeout" json:"transaction_timeout"`
	MaxTransactionSize int           `yaml:"max_transaction_size" json:"max_transaction_size"`

	// Retry configuration
	MaxRetries     int           `yaml:"max_retries" json:"max_retries"`
	BaseRetryDelay time.Duration `yaml:"base_retry_delay" json:"base_retry_delay"`

	// Performance monitoring
	EnablePerformanceLog bool    `yaml:"enable_performance_log" json:"enable_performance_log"`
	PerformanceThreshold float64 `yaml:"performance_threshold" json:"performance_threshold"`

	// Query optimization
	MaxResultsPerQuery int           `yaml:"max_results_per_query" json:"max_results_per_query"`
	QueryTimeout       time.Duration `yaml:"query_timeout" json:"query_timeout"`
}

// Load loads configuration from environment variables
func Load() *Config {
	return &Config{
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnvAsInt("DB_PORT", 5432),
			User:     getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", "password"),
			Name:     getEnv("DB_NAME", "tpos_db"),
			SSLMode:  getEnv("DB_SSL_MODE", "disable"),
		},
		Server: ServerConfig{
			Host: getEnv("SERVER_HOST", "localhost"),
			Port: getEnvAsInt("SERVER_PORT", 8080),
		},
		JWT: JWTConfig{
			Secret:     getEnv("JWT_SECRET", "your-secret-key"),
			ExpiryHour: getEnvAsInt("JWT_EXPIRY_HOUR", 720), // 30 days
		},
		Sync: loadSyncConfig(),
	}
}

// loadSyncConfig loads sync configuration from environment variables with defaults
func loadSyncConfig() SyncConfig {
	return SyncConfig{
		BatchSize:            getEnvAsInt("SYNC_BATCH_SIZE", 100),
		MaxEntitiesPerSync:   getEnvAsInt("SYNC_MAX_ENTITIES_PER_SYNC", 1000),
		TransactionTimeout:   getEnvAsDuration("SYNC_TRANSACTION_TIMEOUT", 30*time.Second),
		MaxTransactionSize:   getEnvAsInt("SYNC_MAX_TRANSACTION_SIZE", 500),
		MaxRetries:           getEnvAsInt("SYNC_MAX_RETRIES", 3),
		BaseRetryDelay:       getEnvAsDuration("SYNC_BASE_RETRY_DELAY", 100*time.Millisecond),
		EnablePerformanceLog: getEnvAsBool("SYNC_ENABLE_PERFORMANCE_LOG", true),
		PerformanceThreshold: getEnvAsFloat64("SYNC_PERFORMANCE_THRESHOLD", 10.0),
		MaxResultsPerQuery:   getEnvAsInt("SYNC_MAX_RESULTS_PER_QUERY", 1000),
		QueryTimeout:         getEnvAsDuration("SYNC_QUERY_TIMEOUT", 10*time.Second),
	}
}

// getEnv gets environment variable with default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvAsInt gets environment variable as integer with default value
func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

// getEnvAsBool gets environment variable as boolean with default value
func getEnvAsBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}

// getEnvAsFloat64 gets environment variable as float64 with default value
func getEnvAsFloat64(key string, defaultValue float64) float64 {
	if value := os.Getenv(key); value != "" {
		if floatValue, err := strconv.ParseFloat(value, 64); err == nil {
			return floatValue
		}
	}
	return defaultValue
}

// getEnvAsDuration gets environment variable as duration with default value
func getEnvAsDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}
