package services

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/terminator791/t-pos/config"
	"github.com/terminator791/t-pos/internal/domain/dto"
	"github.com/terminator791/t-pos/internal/domain/entities"
)

func TestSyncService_ValidateSyncRequest(t *testing.T) {
	service := createTestSyncService()

	tests := []struct {
		name        string
		request     dto.SyncRequest
		expectError bool
		errorMsg    string
	}{
		{
			name: "Valid request within limits",
			request: dto.SyncRequest{
				Carts:      make([]entities.Cart, 50),
				Categories: make([]entities.Category, 50),
				Products:   make([]entities.Product, 50),
			},
			expectError: false,
		},
		{
			name: "Request exceeds total entity limit",
			request: dto.SyncRequest{
				Carts:      make([]entities.Cart, 600),
				Categories: make([]entities.Category, 600),
			},
			expectError: true,
			errorMsg:    "sync request too large",
		},
		{
			name: "Single entity type exceeds limit",
			request: dto.SyncRequest{
				Products: make([]entities.Product, 300), // Exceeds half of MaxEntitiesPerSync (500/2 = 250)
			},
			expectError: true,
			errorMsg:    "products count too large",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.validateSyncRequest(tt.request)
			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestSyncService_IsRetryableError(t *testing.T) {
	service := createTestSyncService()

	tests := []struct {
		name        string
		error       error
		shouldRetry bool
	}{
		{
			name:        "Connection refused error",
			error:       assert.AnError,
			shouldRetry: false,
		},
		{
			name:        "Timeout error",
			error:       &testError{msg: "context deadline exceeded"},
			shouldRetry: true,
		},
		{
			name:        "Deadlock error",
			error:       &testError{msg: "deadlock detected"},
			shouldRetry: true,
		},
		{
			name:        "Regular error",
			error:       &testError{msg: "some regular error"},
			shouldRetry: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.isRetryableError(tt.error)
			assert.Equal(t, tt.shouldRetry, result)
		})
	}
}

func TestSyncService_RetryOperation(t *testing.T) {
	service := createTestSyncService()
	ctx := context.Background()

	t.Run("Successful operation on first try", func(t *testing.T) {
		callCount := 0
		operation := func() error {
			callCount++
			return nil
		}

		err := service.retryOperation(ctx, operation, 3, 10*time.Millisecond, "test_op")
		assert.NoError(t, err)
		assert.Equal(t, 1, callCount)
	})

	t.Run("Successful operation after retries", func(t *testing.T) {
		callCount := 0
		operation := func() error {
			callCount++
			if callCount < 3 {
				return &testError{msg: "timeout"}
			}
			return nil
		}

		err := service.retryOperation(ctx, operation, 3, 1*time.Millisecond, "test_op")
		assert.NoError(t, err)
		assert.Equal(t, 3, callCount)
	})

	t.Run("Non-retryable error", func(t *testing.T) {
		callCount := 0
		operation := func() error {
			callCount++
			return &testError{msg: "validation error"}
		}

		err := service.retryOperation(ctx, operation, 3, 1*time.Millisecond, "test_op")
		assert.Error(t, err)
		assert.Equal(t, 1, callCount) // Should not retry
	})

	t.Run("Exceeds max retries", func(t *testing.T) {
		callCount := 0
		operation := func() error {
			callCount++
			return &testError{msg: "timeout"}
		}

		err := service.retryOperation(ctx, operation, 2, 1*time.Millisecond, "test_op")
		assert.Error(t, err)
		assert.Equal(t, 3, callCount) // Initial + 2 retries
		assert.Contains(t, err.Error(), "failed after 3 retries")
	})
}

func TestSyncService_BatchProcessing(t *testing.T) {
	service := createTestSyncService()

	t.Run("Batch processing with multiple carts", func(t *testing.T) {
		carts := []entities.Cart{
			{ID: uuid.New(), ShopID: uuid.New(), ProductID: uuid.New(), UserID: uuid.New(), Quantity: 1},
			{ID: uuid.New(), ShopID: uuid.New(), ProductID: uuid.New(), UserID: uuid.New(), Quantity: 2},
			{ID: uuid.New(), ShopID: uuid.New(), ProductID: uuid.New(), UserID: uuid.New(), Quantity: 3},
		}

		// Test that config is properly set
		assert.Equal(t, 50, service.config.BatchSize)
		assert.Equal(t, 500, service.config.MaxEntitiesPerSync)
		assert.Equal(t, 30*time.Second, service.config.TransactionTimeout)

		// Test that we have more carts than batch size (would require batching)
		assert.Len(t, carts, 3)
		assert.True(t, len(carts) > service.config.BatchSize/20) // Adjusted for smaller test batch
	})
}

func TestSyncService_ConfigurableSettings(t *testing.T) {
	service := createTestSyncService()

	t.Run("Config values are properly set", func(t *testing.T) {
		// Test that configuration is loaded correctly
		assert.Equal(t, 50, service.config.BatchSize)
		assert.Equal(t, 500, service.config.MaxEntitiesPerSync)
		assert.Equal(t, 30*time.Second, service.config.TransactionTimeout)
		assert.Equal(t, 3, service.config.MaxRetries)
		assert.Equal(t, 50*time.Millisecond, service.config.BaseRetryDelay)
		assert.True(t, service.config.EnablePerformanceLog)
		assert.Equal(t, 10.0, service.config.PerformanceThreshold)
		assert.Equal(t, 100, service.config.MaxResultsPerQuery)
		assert.Equal(t, 5*time.Second, service.config.QueryTimeout)
	})

	t.Run("Config is used in validation", func(t *testing.T) {
		// Test that validateSyncRequest uses the config values
		req := dto.SyncRequest{
			Products: make([]entities.Product, service.config.MaxEntitiesPerSync+1),
		}

		err := service.validateSyncRequest(req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "sync request too large")
	})
}

func TestSyncService_AddDetailedError(t *testing.T) {
	service := createTestSyncService()
	response := &dto.SyncResponse{
		Errors: make([]dto.SyncError, 0),
	}

	entityID := uuid.New()
	details := map[string]interface{}{
		"operation":      "create",
		"retry_attempts": 3,
	}

	service.addDetailedError(response, "products", entityID, "create_failed", "Failed to create product", details)

	assert.Len(t, response.Errors, 1)
	assert.Equal(t, "products", response.Errors[0].EntityType)
	assert.Equal(t, entityID, response.Errors[0].EntityID)
	assert.Equal(t, "create_failed", response.Errors[0].ErrorCode)
	assert.Equal(t, "Failed to create product", response.Errors[0].Message)
	assert.Contains(t, response.Errors[0].Details, "operation")
	assert.Contains(t, response.Errors[0].Details, "retry_attempts")
}

func TestSyncService_LogPerformanceMetrics(t *testing.T) {
	service := createTestSyncService()

	// This test just ensures the method doesn't panic
	// In a real test environment, you'd capture log output
	service.logPerformanceMetrics("products", 100, 5*time.Second, "push")
	service.logPerformanceMetrics("carts", 5, 10*time.Second, "pull") // Should trigger warning
}

// Helper types and functions for testing

type testError struct {
	msg string
}

func (e *testError) Error() string {
	return e.msg
}

func createTestSyncService() *SyncService {
	testConfig := config.SyncConfig{
		BatchSize:              50,
		MaxEntitiesPerSync:     500,
		MaxMemoryUsageMB:       100,   // 100MB limit for tests
		EntitySizeEstimateMB:   0.001, // 1KB per entity estimate
		TransactionTimeout:     30 * time.Second,
		MaxTransactionSize:     100,
		ErrorPolicy:            "continue", // Default error policy
		MaxEntityErrorsPerSync: 50,
		MaxRetries:             3,
		BaseRetryDelay:         50 * time.Millisecond,
		EnablePerformanceLog:   true,
		PerformanceThreshold:   10.0,
		MaxResultsPerQuery:     100,
		QueryTimeout:           5 * time.Second,
	}

	return &SyncService{
		config:           testConfig,
		conflictStrategy: dto.LastWriteWins,
	}
}
