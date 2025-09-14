package handlers

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/terminator791/t-pos/internal/domain/dto"
	"github.com/terminator791/t-pos/internal/domain/entities"
)

// Helper function to create test UUIDs
func createTestUUID() uuid.UUID {
	id, _ := uuid.NewRandom()
	return id
}

func TestSyncHandler_InjectShopIDIntoEntities(t *testing.T) {
	// Create sync handler (we'll test the method directly)
	syncHandler := &SyncHandler{}
	
	// Setup test data
	shopID := createTestUUID()
	productID := createTestUUID()
	cartID := createTestUUID()
	
	// Create sync request with entities that don't have shop_id set
	syncReq := dto.SyncRequest{
		Products: []entities.Product{
			{
				ID:   productID,
				Name: "Test Product",
				// ShopID is not set
			},
		},
		Carts: []entities.Cart{
			{
				ID:       cartID,
				Quantity: 5,
				// ShopID is not set
			},
		},
	}
	
	// Test the injection
	syncHandler.injectShopIDIntoEntities(&syncReq, shopID)
	
	// Assertions
	assert.Equal(t, shopID, syncReq.Products[0].ShopID)
	assert.Equal(t, shopID, syncReq.Carts[0].ShopID)
}

func TestSyncHandler_ValidateCashierEntitiesShopID_Success(t *testing.T) {
	// Create sync handler
	syncHandler := &SyncHandler{}
	
	// Setup test data
	shopID := createTestUUID()
	productID := createTestUUID()
	
	// Create sync request with entities that have correct shop_id
	syncReq := dto.SyncRequest{
		Products: []entities.Product{
			{
				ID:     productID,
				Name:   "Test Product",
				ShopID: shopID, // Correct shop ID
			},
		},
	}
	
	// Test validation - should pass
	err := syncHandler.validateCashierEntitiesShopID(&syncReq, shopID)
	
	// Assertions
	assert.NoError(t, err)
}

func TestSyncHandler_ValidateCashierEntitiesShopID_Failure(t *testing.T) {
	// Create sync handler
	syncHandler := &SyncHandler{}
	
	// Setup test data
	correctShopID := createTestUUID()
	wrongShopID := createTestUUID()
	productID := createTestUUID()
	
	// Create sync request with entities that have wrong shop_id
	syncReq := dto.SyncRequest{
		Products: []entities.Product{
			{
				ID:     productID,
				Name:   "Test Product",
				ShopID: wrongShopID, // Wrong shop ID
			},
		},
	}
	
	// Test validation - should fail
	err := syncHandler.validateCashierEntitiesShopID(&syncReq, correctShopID)
	
	// Assertions
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid shop_id detected")
	assert.Contains(t, err.Error(), wrongShopID.String())
	assert.Contains(t, err.Error(), correctShopID.String())
}

func TestSyncHandler_HandleShopIDRequirements_Cashier_Success(t *testing.T) {
	// Create sync handler
	syncHandler := &SyncHandler{}
	
	// Setup test data
	shopID := createTestUUID()
	userID := createTestUUID()
	productID := createTestUUID()
	
	// Create cashier user
	cashierUser := &entities.User{
		ID:     userID,
		Name:   "Test Cashier",
		ShopID: &shopID,
	}
	
	// Create sync request without shop_id (should be auto-injected)
	syncReq := dto.SyncRequest{
		Products: []entities.Product{
			{
				ID:   productID,
				Name: "Test Product",
				// ShopID is not set - should be auto-injected
			},
		},
	}
	
	// Test handling
	err := syncHandler.handleShopIDRequirements(&syncReq, "cashier", cashierUser)
	
	// Assertions
	assert.NoError(t, err)
	assert.Equal(t, shopID, syncReq.Products[0].ShopID)
}

func TestSyncHandler_HandleShopIDRequirements_Cashier_NoShopAssigned(t *testing.T) {
	// Create sync handler
	syncHandler := &SyncHandler{}
	
	// Setup test data
	userID := createTestUUID()
	
	// Create cashier user without shop assignment
	cashierUser := &entities.User{
		ID:     userID,
		Name:   "Test Cashier",
		ShopID: nil, // No shop assigned
	}
	
	// Create sync request
	syncReq := dto.SyncRequest{
		Products: []entities.Product{{ID: createTestUUID(), Name: "Test Product"}},
	}
	
	// Test handling - should fail
	err := syncHandler.handleShopIDRequirements(&syncReq, "cashier", cashierUser)
	
	// Assertions
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not assigned to a shop")
}

func TestSyncHandler_HandleShopIDRequirements_Cashier_DomainViolation(t *testing.T) {
	// Create sync handler
	syncHandler := &SyncHandler{}
	
	// Setup test data
	correctShopID := createTestUUID()
	wrongShopID := createTestUUID()
	userID := createTestUUID()
	productID := createTestUUID()
	
	// Create cashier user
	cashierUser := &entities.User{
		ID:     userID,
		Name:   "Test Cashier",
		ShopID: &correctShopID,
	}
	
	// Create sync request with wrong shop_id
	syncReq := dto.SyncRequest{
		Products: []entities.Product{
			{
				ID:     productID,
				Name:   "Test Product",
				ShopID: wrongShopID, // Wrong shop ID
			},
		},
	}
	
	// Test handling - should fail
	err := syncHandler.handleShopIDRequirements(&syncReq, "cashier", cashierUser)
	
	// Assertions
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "domain validation failed")
}

func TestSyncHandler_HandleShopIDRequirements_UnknownRole(t *testing.T) {
	// Create sync handler
	syncHandler := &SyncHandler{}
	
	// Setup test data
	userID := createTestUUID()
	
	// Create user
	user := &entities.User{
		ID:   userID,
		Name: "Test User",
	}
	
	// Create sync request
	syncReq := dto.SyncRequest{}
	
	// Test handling with unknown role
	err := syncHandler.handleShopIDRequirements(&syncReq, "unknown_role", user)
	
	// Assertions
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown user role: unknown_role")
}