package services

import (
	"fmt"
	"time"

	"github.com/terminator791/t-pos/config"
	"github.com/terminator791/t-pos/internal/domain/dto"
)

// SyncServiceConfig consolidates all sync service configuration
// This replaces scattered configuration parameters throughout the original service
type SyncServiceConfig struct {
	// Core Processing Configuration
	Processing ProcessingConfig `json:"processing"`

	// Security Configuration
	Security SecurityConfig `json:"security"`

	// Performance Configuration
	Performance PerformanceConfig `json:"performance"`

	// Error Handling Configuration
	ErrorHandling ErrorHandlingConfig `json:"error_handling"`

	// Memory Management Configuration
	Memory MemoryConfig `json:"memory"`
}

// ProcessingConfig contains entity processing settings
type ProcessingConfig struct {
	BatchSize                int           `json:"batch_size" default:"100"`
	MaxEntitiesPerSync       int           `json:"max_entities_per_sync" default:"1000"`
	TransactionTimeout       time.Duration `json:"transaction_timeout" default:"30s"`
	MaxTransactionTimeout    time.Duration `json:"max_transaction_timeout" default:"5m"`
	EnableSavepoints         bool          `json:"enable_savepoints" default:"true"`
	EnableBatchProcessing    bool          `json:"enable_batch_processing" default:"true"`
	OptimalBatchSize         int           `json:"optimal_batch_size" default:"200"`
	EnableParallelProcessing bool          `json:"enable_parallel_processing" default:"false"`
}

// SecurityConfig contains security-related settings
type SecurityConfig struct {
	EnableDistributedLocking      bool          `json:"enable_distributed_locking" default:"true"`
	EnableComprehensiveValidation bool          `json:"enable_comprehensive_validation" default:"true"`
	ValidationDepth               string        `json:"validation_depth" default:"comprehensive"`
	LockTimeout                   time.Duration `json:"lock_timeout" default:"30s"`
	LockCleanupInterval           time.Duration `json:"lock_cleanup_interval" default:"10s"`
	MaxConcurrentSyncsPerUser     int           `json:"max_concurrent_syncs_per_user" default:"1"`
	EnableEntityLocking           bool          `json:"enable_entity_locking" default:"false"`
	EnableAccessAuditLogging      bool          `json:"enable_access_audit_logging" default:"false"`
}

// PerformanceConfig contains performance optimization settings
type PerformanceConfig struct {
	EnableCaching           bool          `json:"enable_caching" default:"true"`
	EnableBulkValidation    bool          `json:"enable_bulk_validation" default:"true"`
	EnableQueryOptimization bool          `json:"enable_query_optimization" default:"true"`
	EnableIndexHints        bool          `json:"enable_index_hints" default:"true"`
	CacheTTL                time.Duration `json:"cache_ttl" default:"5m"`
	MaxCacheEntries         int           `json:"max_cache_entries" default:"10000"`
	CacheCleanupInterval    time.Duration `json:"cache_cleanup_interval" default:"2m"`
	MaxResultsPerQuery      int           `json:"max_results_per_query" default:"1000"`
	EnablePerformanceLog    bool          `json:"enable_performance_log" default:"false"`
	PerformanceThreshold    float64       `json:"performance_threshold" default:"100.0"`
}

// ErrorHandlingConfig contains error handling settings
type ErrorHandlingConfig struct {
	Policy                     string `json:"policy" default:"continue"`
	MaxEntityErrorsPerSync     int    `json:"max_entity_errors_per_sync" default:"50"`
	EnableDetailedErrorLogging bool   `json:"enable_detailed_error_logging" default:"true"`
	EnableErrorRetry           bool   `json:"enable_error_retry" default:"true"`
	MaxRetryAttempts           int    `json:"max_retry_attempts" default:"3"`
	RetryDelayMs               int    `json:"retry_delay_ms" default:"100"`
	EnableRetryBackoff         bool   `json:"enable_retry_backoff" default:"true"`
}

// MemoryConfig contains memory management settings
type MemoryConfig struct {
	MaxMemoryUsageMB       int64   `json:"max_memory_usage_mb" default:"100"`
	EntitySizeEstimateMB   float64 `json:"entity_size_estimate_mb" default:"0.001"`
	EnableMemoryMonitoring bool    `json:"enable_memory_monitoring" default:"true"`
	MemoryWarningThreshold float64 `json:"memory_warning_threshold" default:"0.8"`
	EnableGCOptimization   bool    `json:"enable_gc_optimization" default:"false"`
}

// NewSyncServiceConfig creates a new sync service configuration with defaults
func NewSyncServiceConfig() SyncServiceConfig {
	return SyncServiceConfig{
		Processing: ProcessingConfig{
			BatchSize:                100,
			MaxEntitiesPerSync:       1000,
			TransactionTimeout:       30 * time.Second,
			MaxTransactionTimeout:    5 * time.Minute,
			EnableSavepoints:         true,
			EnableBatchProcessing:    true,
			OptimalBatchSize:         200,
			EnableParallelProcessing: false,
		},
		Security: SecurityConfig{
			EnableDistributedLocking:      true,
			EnableComprehensiveValidation: true,
			ValidationDepth:               "comprehensive",
			LockTimeout:                   30 * time.Second,
			LockCleanupInterval:           10 * time.Second,
			MaxConcurrentSyncsPerUser:     1,
			EnableEntityLocking:           false,
			EnableAccessAuditLogging:      false,
		},
		Performance: PerformanceConfig{
			EnableCaching:           true,
			EnableBulkValidation:    true,
			EnableQueryOptimization: true,
			EnableIndexHints:        true,
			CacheTTL:                5 * time.Minute,
			MaxCacheEntries:         10000,
			CacheCleanupInterval:    2 * time.Minute,
			MaxResultsPerQuery:      1000,
			EnablePerformanceLog:    false,
			PerformanceThreshold:    100.0,
		},
		ErrorHandling: ErrorHandlingConfig{
			Policy:                     "continue",
			MaxEntityErrorsPerSync:     50,
			EnableDetailedErrorLogging: true,
			EnableErrorRetry:           true,
			MaxRetryAttempts:           3,
			RetryDelayMs:               100,
			EnableRetryBackoff:         true,
		},
		Memory: MemoryConfig{
			MaxMemoryUsageMB:       100,
			EntitySizeEstimateMB:   0.001,
			EnableMemoryMonitoring: true,
			MemoryWarningThreshold: 0.8,
			EnableGCOptimization:   false,
		},
	}
}

// LoadFromConfig converts the existing config.SyncConfig to the new structured format
func (c *SyncServiceConfig) LoadFromConfig(syncConfig config.SyncConfig) {
	// Processing configuration
	c.Processing.BatchSize = syncConfig.BatchSize
	c.Processing.MaxEntitiesPerSync = syncConfig.MaxEntitiesPerSync
	c.Processing.TransactionTimeout = time.Duration(syncConfig.TransactionTimeout) * time.Second
	c.Processing.OptimalBatchSize = syncConfig.OptimalBatchSize
	c.Processing.EnableBatchProcessing = syncConfig.EnableBatchProcessing

	// Security configuration
	c.Security.EnableDistributedLocking = syncConfig.EnableDistributedLocking
	c.Security.EnableComprehensiveValidation = syncConfig.EnableComprehensiveValidation
	c.Security.ValidationDepth = syncConfig.ValidationDepth
	c.Security.LockTimeout = time.Duration(syncConfig.LockTimeout) * time.Second
	c.Security.LockCleanupInterval = time.Duration(syncConfig.LockCleanupInterval) * time.Second
	c.Security.MaxConcurrentSyncsPerUser = syncConfig.MaxConcurrentSyncsPerUser

	// Performance configuration
	c.Performance.EnableCaching = syncConfig.EnableCaching
	c.Performance.EnableBulkValidation = syncConfig.EnableBulkValidation
	c.Performance.EnableQueryOptimization = syncConfig.EnableQueryOptimization
	c.Performance.EnableIndexHints = syncConfig.EnableIndexHints
	c.Performance.CacheTTL = syncConfig.CacheTTL
	c.Performance.MaxCacheEntries = syncConfig.MaxCacheEntries
	c.Performance.CacheCleanupInterval = syncConfig.CacheCleanupInterval
	c.Performance.MaxResultsPerQuery = syncConfig.MaxResultsPerQuery
	c.Performance.EnablePerformanceLog = syncConfig.EnablePerformanceLog
	c.Performance.PerformanceThreshold = syncConfig.PerformanceThreshold

	// Error handling configuration
	c.ErrorHandling.Policy = syncConfig.ErrorPolicy
	c.ErrorHandling.MaxEntityErrorsPerSync = syncConfig.MaxEntityErrorsPerSync

	// Memory configuration
	c.Memory.MaxMemoryUsageMB = syncConfig.MaxMemoryUsageMB
	c.Memory.EntitySizeEstimateMB = syncConfig.EntitySizeEstimateMB
}

// ToEntityProcessingConfig converts to the entity processing configuration
func (c *SyncServiceConfig) ToEntityProcessingConfig() EntityProcessingConfig {
	return EntityProcessingConfig{
		BatchSize:        c.Processing.BatchSize,
		EnableValidation: c.Security.EnableComprehensiveValidation,
		EnableSavepoints: c.Processing.EnableSavepoints,
		ErrorPolicy:      dto.ParseSyncErrorPolicy(c.ErrorHandling.Policy),
		MaxRetries:       c.ErrorHandling.MaxRetryAttempts,
		RetryDelay:       time.Duration(c.ErrorHandling.RetryDelayMs) * time.Millisecond,
	}
}

// Validate validates the configuration and returns any errors
func (c *SyncServiceConfig) Validate() error {
	// Validate processing configuration
	if c.Processing.BatchSize <= 0 {
		return fmt.Errorf("processing.batch_size must be positive")
	}
	if c.Processing.MaxEntitiesPerSync <= 0 {
		return fmt.Errorf("processing.max_entities_per_sync must be positive")
	}
	if c.Processing.TransactionTimeout <= 0 {
		return fmt.Errorf("processing.transaction_timeout must be positive")
	}

	// Validate security configuration
	if c.Security.LockTimeout <= 0 {
		return fmt.Errorf("security.lock_timeout must be positive")
	}
	if c.Security.MaxConcurrentSyncsPerUser <= 0 {
		return fmt.Errorf("security.max_concurrent_syncs_per_user must be positive")
	}

	// Validate performance configuration
	if c.Performance.CacheTTL <= 0 {
		return fmt.Errorf("performance.cache_ttl must be positive")
	}
	if c.Performance.MaxCacheEntries <= 0 {
		return fmt.Errorf("performance.max_cache_entries must be positive")
	}

	// Validate error handling configuration
	validPolicies := []string{"continue", "abort", "retry"}
	valid := false
	for _, policy := range validPolicies {
		if c.ErrorHandling.Policy == policy {
			valid = true
			break
		}
	}
	if !valid {
		return fmt.Errorf("error_handling.policy must be one of: %v", validPolicies)
	}

	// Validate memory configuration
	if c.Memory.MaxMemoryUsageMB <= 0 {
		return fmt.Errorf("memory.max_memory_usage_mb must be positive")
	}
	if c.Memory.EntitySizeEstimateMB <= 0 {
		return fmt.Errorf("memory.entity_size_estimate_mb must be positive")
	}

	return nil
}

// GetConfigSummary returns a summary of the current configuration
func (c *SyncServiceConfig) GetConfigSummary() map[string]interface{} {
	return map[string]interface{}{
		"processing": map[string]interface{}{
			"batch_size":              c.Processing.BatchSize,
			"max_entities_per_sync":   c.Processing.MaxEntitiesPerSync,
			"transaction_timeout":     c.Processing.TransactionTimeout.String(),
			"enable_savepoints":       c.Processing.EnableSavepoints,
			"enable_batch_processing": c.Processing.EnableBatchProcessing,
		},
		"security": map[string]interface{}{
			"enable_distributed_locking":      c.Security.EnableDistributedLocking,
			"enable_comprehensive_validation": c.Security.EnableComprehensiveValidation,
			"validation_depth":                c.Security.ValidationDepth,
			"max_concurrent_syncs_per_user":   c.Security.MaxConcurrentSyncsPerUser,
		},
		"performance": map[string]interface{}{
			"enable_caching":            c.Performance.EnableCaching,
			"enable_bulk_validation":    c.Performance.EnableBulkValidation,
			"enable_query_optimization": c.Performance.EnableQueryOptimization,
			"cache_ttl":                 c.Performance.CacheTTL.String(),
			"max_cache_entries":         c.Performance.MaxCacheEntries,
		},
		"error_handling": map[string]interface{}{
			"policy":                     c.ErrorHandling.Policy,
			"max_entity_errors_per_sync": c.ErrorHandling.MaxEntityErrorsPerSync,
			"max_retry_attempts":         c.ErrorHandling.MaxRetryAttempts,
			"enable_error_retry":         c.ErrorHandling.EnableErrorRetry,
		},
		"memory": map[string]interface{}{
			"max_memory_usage_mb":      c.Memory.MaxMemoryUsageMB,
			"entity_size_estimate_mb":  c.Memory.EntitySizeEstimateMB,
			"enable_memory_monitoring": c.Memory.EnableMemoryMonitoring,
		},
	}
}

// OptimizeForEnvironment adjusts configuration based on environment type
func (c *SyncServiceConfig) OptimizeForEnvironment(env string) {
	switch env {
	case "development":
		c.Performance.EnablePerformanceLog = true
		c.ErrorHandling.EnableDetailedErrorLogging = true
		c.Memory.EnableMemoryMonitoring = true
		c.Security.EnableAccessAuditLogging = true

	case "testing":
		c.Processing.BatchSize = 10
		c.Performance.EnableCaching = false
		c.Security.EnableDistributedLocking = false
		c.Memory.MaxMemoryUsageMB = 50

	case "production":
		c.Performance.EnablePerformanceLog = false
		c.ErrorHandling.EnableDetailedErrorLogging = false
		c.Processing.EnableParallelProcessing = true
		c.Performance.EnableIndexHints = true
		c.Memory.EnableGCOptimization = true

	case "staging":
		c.Performance.EnablePerformanceLog = true
		c.ErrorHandling.EnableDetailedErrorLogging = true
		c.Processing.BatchSize = 50
		c.Memory.MaxMemoryUsageMB = 75
	}
}
