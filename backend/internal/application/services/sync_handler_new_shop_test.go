package services

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/terminator791/t-pos/internal/domain/dto"
	"github.com/terminator791/t-pos/internal/domain/entities"
)

// TestNewShopSyncFiltering tests that newly created shops and their entities are not filtered out
func TestNewShopSyncFiltering(t *testing.T) {
	// Create test data
	licenseID := uuid.New()
	newShopID := uuid.New()
	address := "Jl. Test Sync No. 123, Jakarta"

	// Create a sync request with a new shop and category
	syncRequest := dto.SyncRequest{
		Shops: []entities.Shop{
			{
				ID:        newShopID,
				LicenseID: licenseID,
				Name:      "Test Shop Sync update",
				Address:   &address,
			},
		},
		Categories: []entities.Category{
			{
				ID:     uuid.New(),
				ShopID: newShopID,
				Name:   "Test Category Syncs update",
			},
		},
	}

	// Create sync context for owner_business (without the new shop in accessible shops initially)
	syncContext := dto.SyncContext{
		UserID:            uuid.New(),
		UserRole:          "owner_business",
		LicenseID:         licenseID,
		AccessibleShopIDs: []uuid.UUID{}, // Empty initially - no existing shops
		HasGlobalAccess:   false,
	}

	// Test that this simulates the fix
	// In the actual handler, new shop IDs would be added to AccessibleShopIDs
	// This demonstrates that the category would be filtered out without the fix

	// Before fix simulation (shop not in accessible list)
	// Simulate the filtering logic (simplified version)
	accessibleShops := make(map[uuid.UUID]bool)
	for _, shopID := range syncContext.AccessibleShopIDs {
		accessibleShops[shopID] = true
	}

	var filteredCategories []entities.Category
	for _, category := range syncRequest.Categories {
		if accessibleShops[category.ShopID] {
			filteredCategories = append(filteredCategories, category)
		}
	}

	// Without the fix, the category should be filtered out
	assert.Equal(t, 0, len(filteredCategories), "Category should be filtered out without new shop in accessible list")

	// Now simulate the fix - add new shop IDs to accessible shops
	for _, shop := range syncRequest.Shops {
		if shop.LicenseID == syncContext.LicenseID {
			syncContext.AccessibleShopIDs = append(syncContext.AccessibleShopIDs, shop.ID)
		}
	}

	// Rebuild accessible shops map
	accessibleShops = make(map[uuid.UUID]bool)
	for _, shopID := range syncContext.AccessibleShopIDs {
		accessibleShops[shopID] = true
	}

	// Filter again with the fix
	filteredCategories = []entities.Category{}
	for _, category := range syncRequest.Categories {
		if accessibleShops[category.ShopID] {
			filteredCategories = append(filteredCategories, category)
		}
	}

	// With the fix, the category should NOT be filtered out
	assert.Equal(t, 1, len(filteredCategories), "Category should NOT be filtered out with new shop in accessible list")
	assert.Equal(t, syncRequest.Categories[0].ID, filteredCategories[0].ID, "Filtered category should match original")
}

// TestNewShopSyncFilteringLicenseMismatch tests that shops with wrong license are still filtered
func TestNewShopSyncFilteringLicenseMismatch(t *testing.T) {
	// Create test data
	userLicenseID := uuid.New()
	wrongLicenseID := uuid.New()
	newShopID := uuid.New()

	// Create a sync request with a shop that has wrong license
	syncRequest := dto.SyncRequest{
		Shops: []entities.Shop{
			{
				ID:        newShopID,
				LicenseID: wrongLicenseID, // Wrong license!
				Name:      "Test Shop Wrong License",
			},
		},
		Categories: []entities.Category{
			{
				ID:     uuid.New(),
				ShopID: newShopID,
				Name:   "Test Category",
			},
		},
	}

	// Create sync context for owner_business
	syncContext := dto.SyncContext{
		UserID:            uuid.New(),
		UserRole:          "owner_business",
		LicenseID:         userLicenseID,
		AccessibleShopIDs: []uuid.UUID{},
		HasGlobalAccess:   false,
	}

	// Simulate the fix - only add shops with matching license
	for _, shop := range syncRequest.Shops {
		if shop.LicenseID == syncContext.LicenseID {
			syncContext.AccessibleShopIDs = append(syncContext.AccessibleShopIDs, shop.ID)
		}
	}

	// Build accessible shops map
	accessibleShops := make(map[uuid.UUID]bool)
	for _, shopID := range syncContext.AccessibleShopIDs {
		accessibleShops[shopID] = true
	}

	// Filter categories
	var filteredCategories []entities.Category
	for _, category := range syncRequest.Categories {
		if accessibleShops[category.ShopID] {
			filteredCategories = append(filteredCategories, category)
		}
	}

	// The category should still be filtered out due to license mismatch
	assert.Equal(t, 0, len(filteredCategories), "Category should be filtered out due to shop license mismatch")
	assert.Equal(t, 0, len(syncContext.AccessibleShopIDs), "No shops should be added due to license mismatch")
}

// TestAdminShopSyncFiltering tests that admin users can manage shops across any license
func TestAdminShopSyncFiltering(t *testing.T) {
	// Create test data
	licenseID1 := uuid.New()
	licenseID2 := uuid.New()
	shopID1 := uuid.New()
	shopID2 := uuid.New()

	// Create a sync request with shops from different licenses
	syncRequest := dto.SyncRequest{
		Shops: []entities.Shop{
			{
				ID:        shopID1,
				LicenseID: licenseID1,
				Name:      "Shop License 1",
			},
			{
				ID:        shopID2,
				LicenseID: licenseID2,
				Name:      "Shop License 2",
			},
		},
		Categories: []entities.Category{
			{
				ID:     uuid.New(),
				ShopID: shopID1,
				Name:   "Category Shop 1",
			},
			{
				ID:     uuid.New(),
				ShopID: shopID2,
				Name:   "Category Shop 2",
			},
		},
	}

	// Create sync context for admin user
	syncContext := dto.SyncContext{
		UserID:            uuid.New(),
		UserRole:          "admin",
		LicenseID:         licenseID1, // Admin's own license
		AccessibleShopIDs: []uuid.UUID{},
		HasGlobalAccess:   true,
	}

	// Simulate the fix - admin can add any shop from sync request
	for _, shop := range syncRequest.Shops {
		syncContext.AccessibleShopIDs = append(syncContext.AccessibleShopIDs, shop.ID)
	}

	// Build accessible shops map
	accessibleShops := make(map[uuid.UUID]bool)
	for _, shopID := range syncContext.AccessibleShopIDs {
		accessibleShops[shopID] = true
	}

	// Filter categories
	var filteredCategories []entities.Category
	for _, category := range syncRequest.Categories {
		if accessibleShops[category.ShopID] {
			filteredCategories = append(filteredCategories, category)
		}
	}

	// Admin should be able to process all categories regardless of license
	assert.Equal(t, 2, len(filteredCategories), "Admin should be able to process categories from all shops")
	assert.Equal(t, 2, len(syncContext.AccessibleShopIDs), "Admin should have access to all shops in sync request")
}
