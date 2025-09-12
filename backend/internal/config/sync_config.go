package config

import "time"

// SyncConfig holds configuration for sync optimization
type SyncConfig struct {
	// Batch processing configuration
	BatchSize           int           `yaml:"batch_size" json:"batch_size"`
	MaxEntitiesPerSync  int           `yaml:"max_entities_per_sync" json:"max_entities_per_sync"`
	
	// Transaction configuration
	TransactionTimeout  time.Duration `yaml:"transaction_timeout" json:"transaction_timeout"`
	MaxTransactionSize  int           `yaml:"max_transaction_size" json:"max_transaction_size"`
	
	// Retry configuration
	MaxRetries          int           `yaml:"max_retries" json:"max_retries"`
	BaseRetryDelay      time.Duration `yaml:"base_retry_delay" json:"base_retry_delay"`
	
	// Performance monitoring
	EnablePerformanceLog bool          `yaml:"enable_performance_log" json:"enable_performance_log"`
	PerformanceThreshold float64       `yaml:"performance_threshold" json:"performance_threshold"`
	
	// Query optimization
	MaxResultsPerQuery  int           `yaml:"max_results_per_query" json:"max_results_per_query"`
	QueryTimeout        time.Duration `yaml:"query_timeout" json:"query_timeout"`
}

// DefaultSyncConfig returns default sync configuration
func DefaultSyncConfig() *SyncConfig {
	return &SyncConfig{
		BatchSize:            100,
		MaxEntitiesPerSync:   1000,
		TransactionTimeout:   30 * time.Second,
		MaxTransactionSize:   500,
		MaxRetries:           3,
		BaseRetryDelay:       100 * time.Millisecond,
		EnablePerformanceLog: true,
		PerformanceThreshold: 10.0, // entities per second
		MaxResultsPerQuery:   1000,
		QueryTimeout:         10 * time.Second,
	}
}

// LoadSyncConfig loads sync configuration from environment or config file
func LoadSyncConfig() *SyncConfig {
	// For now, return default config
	// This can be extended to load from environment variables or config files
	return DefaultSyncConfig()
}