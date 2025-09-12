package services

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
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
				Products: make([]entities.Product, 600),
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
	
	// Set small batch size for testing
	service.SetBatchSize(2)

	t.Run("Batch processing with multiple carts", func(t *testing.T) {
		carts := []entities.Cart{
			{ID: uuid.New(), ShopID: uuid.New(), ProductID: uuid.New(), UserID: uuid.New(), Quantity: 1},
			{ID: uuid.New(), ShopID: uuid.New(), ProductID: uuid.New(), UserID: uuid.New(), Quantity: 2},
			{ID: uuid.New(), ShopID: uuid.New(), ProductID: uuid.New(), UserID: uuid.New(), Quantity: 3},
		}

		// Test that batch size configuration works
		assert.Equal(t, 2, service.batchSize)
		
		// Test that we can set different batch sizes
		service.SetBatchSize(1)
		assert.Equal(t, 1, service.batchSize)
		
		// Restore original batch size
		service.SetBatchSize(2)
		assert.Equal(t, 2, service.batchSize)
		
		// We can't test the full pushCarts method without proper database mocking
		// but we can test the batch size logic and configuration
		assert.Len(t, carts, 3)
		assert.True(t, len(carts) > service.batchSize) // Would require batching
	})
}

func TestSyncService_ConfigurableSettings(t *testing.T) {
	service := createTestSyncService()

	t.Run("Set batch size", func(t *testing.T) {
		service.SetBatchSize(50)
		assert.Equal(t, 50, service.batchSize)

		// Test invalid values
		service.SetBatchSize(0)
		assert.Equal(t, 50, service.batchSize) // Should remain unchanged

		service.SetBatchSize(1000)
		assert.Equal(t, 50, service.batchSize) // Should remain unchanged (exceeds max)
	})

	t.Run("Set max entities per sync", func(t *testing.T) {
		service.SetMaxEntitiesPerSync(2000)
		assert.Equal(t, 2000, service.maxEntitiesPerSync)

		// Test invalid values
		service.SetMaxEntitiesPerSync(0)
		assert.Equal(t, 2000, service.maxEntitiesPerSync) // Should remain unchanged
	})

	t.Run("Set transaction timeout", func(t *testing.T) {
		service.SetTransactionTimeout(60 * time.Second)
		assert.Equal(t, 60*time.Second, service.transactionTimeout)

		// Test invalid values
		service.SetTransactionTimeout(0)
		assert.Equal(t, 60*time.Second, service.transactionTimeout) // Should remain unchanged
	})
}

func TestSyncService_AddDetailedError(t *testing.T) {
	service := createTestSyncService()
	response := &dto.SyncResponse{
		Errors: make([]dto.SyncError, 0),
	}

	entityID := uuid.New()
	details := map[string]interface{}{
		"operation": "create",
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
	return &SyncService{
		batchSize:              100,
		maxEntitiesPerSync:     1000,
		transactionTimeout:     30 * time.Second,
		conflictStrategy:       dto.LastWriteWins,
	}
}