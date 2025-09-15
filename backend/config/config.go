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

	// Memory management configuration
	MaxMemoryUsageMB  int64 `yaml:"max_memory_usage_mb" json:"max_memory_usage_mb"`
	EntitySizeEstimateMB float64 `yaml:"entity_size_estimate_mb" json:"entity_size_estimate_mb"`

	// Transaction configuration
	TransactionTimeout time.Duration `yaml:"transaction_timeout" json:"transaction_timeout"`
	MaxTransactionSize int           `yaml:"max_transaction_size" json:"max_transaction_size"`

	// Error handling configuration
	ErrorPolicy            string `yaml:"error_policy" json:"error_policy"` // "continue", "abort", "retry"
	MaxEntityErrorsPerSync int    `yaml:"max_entity_errors_per_sync" json:"max_entity_errors_per_sync"`

	// Retry configuration
	MaxRetries     int           `yaml:"max_retries" json:"max_retries"`
	BaseRetryDelay time.Duration `yaml:"base_retry_delay" json:"base_retry_delay"`

	// Performance monitoring
	EnablePerformanceLog bool    `yaml:"enable_performance_log" json:"enable_performance_log"`
	PerformanceThreshold float64 `yaml:"performance_threshold" json:"performance_threshold"`

	// Query optimization
	MaxResultsPerQuery int           `yaml:"max_results_per_query" json:"max_results_per_query"`
	QueryTimeout       time.Duration `yaml:"query_timeout" json:"query_timeout"`

	// Security enhancements (Session 2)
	EnableDistributedLocking    bool          `yaml:"enable_distributed_locking" json:"enable_distributed_locking"`
	LockTimeout                 time.Duration `yaml:"lock_timeout" json:"lock_timeout"`
	LockCleanupInterval         time.Duration `yaml:"lock_cleanup_interval" json:"lock_cleanup_interval"`
	EnableComprehensiveValidation bool        `yaml:"enable_comprehensive_validation" json:"enable_comprehensive_validation"`
	ValidationDepth             string        `yaml:"validation_depth" json:"validation_depth"` // "basic", "standard", "comprehensive"
	EnableEntityLocking         bool          `yaml:"enable_entity_locking" json:"enable_entity_locking"`
	MaxConcurrentSyncsPerUser   int           `yaml:"max_concurrent_syncs_per_user" json:"max_concurrent_syncs_per_user"`

	// Performance optimizations (Session 3)
	EnableBulkValidation       bool          `yaml:"enable_bulk_validation" json:"enable_bulk_validation"`
	EnableQueryOptimization    bool          `yaml:"enable_query_optimization" json:"enable_query_optimization"`
	EnableCaching              bool          `yaml:"enable_caching" json:"enable_caching"`
	CacheTTL                   time.Duration `yaml:"cache_ttl" json:"cache_ttl"`
	MaxCacheEntries            int           `yaml:"max_cache_entries" json:"max_cache_entries"`
	CacheCleanupInterval       time.Duration `yaml:"cache_cleanup_interval" json:"cache_cleanup_interval"`
	EnableBatchProcessing      bool          `yaml:"enable_batch_processing" json:"enable_batch_processing"`
	OptimalBatchSize           int           `yaml:"optimal_batch_size" json:"optimal_batch_size"`
	EnableAsyncProcessing      bool          `yaml:"enable_async_processing" json:"enable_async_processing"`
	EnableIndexHints           bool          `yaml:"enable_index_hints" json:"enable_index_hints"`
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
		BatchSize:              getEnvAsInt("SYNC_BATCH_SIZE", 100),
		MaxEntitiesPerSync:     getEnvAsInt("SYNC_MAX_ENTITIES_PER_SYNC", 1000),
		MaxMemoryUsageMB:       getEnvAsInt64("SYNC_MAX_MEMORY_USAGE_MB", 100),
		EntitySizeEstimateMB:   getEnvAsFloat64("SYNC_ENTITY_SIZE_ESTIMATE_MB", 0.001), // 1KB average
		TransactionTimeout:     getEnvAsDuration("SYNC_TRANSACTION_TIMEOUT", 30*time.Second),
		MaxTransactionSize:     getEnvAsInt("SYNC_MAX_TRANSACTION_SIZE", 500),
		ErrorPolicy:            getEnv("SYNC_ERROR_POLICY", "continue"),
		MaxEntityErrorsPerSync: getEnvAsInt("SYNC_MAX_ENTITY_ERRORS_PER_SYNC", 50),
		MaxRetries:             getEnvAsInt("SYNC_MAX_RETRIES", 3),
		BaseRetryDelay:         getEnvAsDuration("SYNC_BASE_RETRY_DELAY", 100*time.Millisecond),
		EnablePerformanceLog:   getEnvAsBool("SYNC_ENABLE_PERFORMANCE_LOG", true),
		PerformanceThreshold:   getEnvAsFloat64("SYNC_PERFORMANCE_THRESHOLD", 10.0),
		MaxResultsPerQuery:     getEnvAsInt("SYNC_MAX_RESULTS_PER_QUERY", 1000),
		QueryTimeout:           getEnvAsDuration("SYNC_QUERY_TIMEOUT", 10*time.Second),
		
		// Security enhancements (Session 2)
		EnableDistributedLocking:      getEnvAsBool("SYNC_ENABLE_DISTRIBUTED_LOCKING", true),
		LockTimeout:                   getEnvAsDuration("SYNC_LOCK_TIMEOUT", 30*time.Second),
		LockCleanupInterval:           getEnvAsDuration("SYNC_LOCK_CLEANUP_INTERVAL", 10*time.Second),
		EnableComprehensiveValidation: getEnvAsBool("SYNC_ENABLE_COMPREHENSIVE_VALIDATION", true),
		ValidationDepth:               getEnv("SYNC_VALIDATION_DEPTH", "comprehensive"), // basic, standard, comprehensive
		EnableEntityLocking:           getEnvAsBool("SYNC_ENABLE_ENTITY_LOCKING", false), // Disabled by default for performance
		MaxConcurrentSyncsPerUser:     getEnvAsInt("SYNC_MAX_CONCURRENT_SYNCS_PER_USER", 1),

		// Performance optimizations (Session 3)
		EnableBulkValidation:    getEnvAsBool("SYNC_ENABLE_BULK_VALIDATION", true),
		EnableQueryOptimization: getEnvAsBool("SYNC_ENABLE_QUERY_OPTIMIZATION", true),
		EnableCaching:           getEnvAsBool("SYNC_ENABLE_CACHING", true),
		CacheTTL:                getEnvAsDuration("SYNC_CACHE_TTL", 5*time.Minute),
		MaxCacheEntries:         getEnvAsInt("SYNC_MAX_CACHE_ENTRIES", 10000),
		CacheCleanupInterval:    getEnvAsDuration("SYNC_CACHE_CLEANUP_INTERVAL", 2*time.Minute),
		EnableBatchProcessing:   getEnvAsBool("SYNC_ENABLE_BATCH_PROCESSING", true),
		OptimalBatchSize:        getEnvAsInt("SYNC_OPTIMAL_BATCH_SIZE", 200),
		EnableAsyncProcessing:   getEnvAsBool("SYNC_ENABLE_ASYNC_PROCESSING", false), // Disabled by default
		EnableIndexHints:        getEnvAsBool("SYNC_ENABLE_INDEX_HINTS", true),
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

// getEnvAsInt64 gets environment variable as int64 with default value
func getEnvAsInt64(key string, defaultValue int64) int64 {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.ParseInt(value, 10, 64); err == nil {
			return intValue
		}
	}
	return defaultValue
}
