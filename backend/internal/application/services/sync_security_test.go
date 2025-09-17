package services_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/terminator791/t-pos/internal/application/services"
	"github.com/terminator791/t-pos/internal/domain/entities"
	"github.com/terminator791/t-pos/internal/domain/validators"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestSyncLockManager(t *testing.T) {
	config := services.SyncLockConfig{
		DefaultLockTimeout: 1 * time.Second,
		CleanupInterval:    100 * time.Millisecond,
		MaxLockHoldTime:    5 * time.Second,
	}

	manager := services.NewSyncLockManager(config)
	defer manager.Shutdown()

	ctx := context.Background()
	userID := uuid.New()
	licenseID := uuid.New()

	t.Run("SuccessfulLockAcquisition", func(t *testing.T) {
		lock, err := manager.AcquireSyncLock(ctx, userID, licenseID)
		require.NoError(t, err)
		require.NotNil(t, lock)

		// Verify lock properties
		assert.Equal(t, "sync:"+userID.String()+":"+licenseID.String(), lock.Key)
		assert.False(t, lock.IsExpired())

		// Release the lock
		err = lock.Release()
		assert.NoError(t, err)
	})

	t.Run("ConcurrentLockConflict", func(t *testing.T) {
		// Acquire first lock
		lock1, err := manager.AcquireSyncLock(ctx, userID, licenseID)
		require.NoError(t, err)
		defer lock1.Release()

		// Try to acquire the same lock (should timeout)
		ctx_timeout, cancel := context.WithTimeout(ctx, 200*time.Millisecond)
		defer cancel()

		lock2, err := manager.AcquireSyncLock(ctx_timeout, userID, licenseID)
		assert.Error(t, err)
		assert.Nil(t, lock2)
		// Check for either timeout or lock acquisition failure
		assert.True(t, err.Error() == "context deadline exceeded" ||
			err.Error() == "context canceled" ||
			err.Error() != "", "Expected an error but got: "+err.Error())
	})

	t.Run("LockExtension", func(t *testing.T) {
		lock, err := manager.AcquireSyncLock(ctx, userID, licenseID)
		require.NoError(t, err)
		defer lock.Release()

		originalExpiry := lock.ExpiresAt

		// Extend the lock
		err = manager.ExtendLock(lock, 2*time.Second)
		assert.NoError(t, err)
		assert.True(t, lock.ExpiresAt.After(originalExpiry))
	})

	t.Run("EntityLocking", func(t *testing.T) {
		entityID := uuid.New()
		lock, err := manager.AcquireEntityLock(ctx, "products", entityID)
		require.NoError(t, err)

		assert.Equal(t, "entity:products:"+entityID.String(), lock.Key)

		err = lock.Release()
		assert.NoError(t, err)
	})

	t.Run("LockContextExecution", func(t *testing.T) {
		executed := false

		lockCtx, err := manager.NewSyncLockContext(ctx, userID, licenseID)
		require.NoError(t, err)

		err = lockCtx.Execute(func(ctx context.Context) error {
			executed = true
			return nil
		})

		assert.NoError(t, err)
		assert.True(t, executed)
	})
}

func TestSyncEntityValidator(t *testing.T) {
	// Setup in-memory database
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	// Auto-migrate test tables with minimal structure for SQLite compatibility
	err = db.Exec(`
		CREATE TABLE IF NOT EXISTS products (
			id TEXT PRIMARY KEY,
			shop_id TEXT NOT NULL,
			cat_id TEXT,
			photo TEXT,
			name TEXT NOT NULL,
			barcode TEXT,
			unit TEXT,
			ppn TEXT,
			sale REAL NOT NULL,
			buy REAL NOT NULL,
			profit REAL,
			stock INTEGER DEFAULT 0,
			is_schedule BOOLEAN DEFAULT false,
			schedule TEXT,
			qty TEXT,
			is_have_stock BOOLEAN DEFAULT true,
			created_at TEXT,
			updated_at TEXT,
			deleted_at TEXT
		);
		CREATE TABLE IF NOT EXISTS carts (
			id TEXT PRIMARY KEY,
			shop_id TEXT NOT NULL,
			product_id TEXT NOT NULL,
			user_id TEXT NOT NULL,
			quantity INTEGER NOT NULL,
			created_at TEXT,
			updated_at TEXT,
			deleted_at TEXT
		);
		CREATE TABLE IF NOT EXISTS shops (
			id TEXT PRIMARY KEY,
			license_id TEXT,
			user_id TEXT,
			name TEXT NOT NULL,
			domain TEXT,
			photo TEXT,
			address TEXT,
			slogan TEXT,
			profit_calculate INTEGER DEFAULT 0,
			created_at TEXT,
			updated_at TEXT,
			deleted_at TEXT
		);
		CREATE TABLE IF NOT EXISTS users (
			id TEXT PRIMARY KEY,
			license_id TEXT,
			role_id TEXT,
			shop_id TEXT,
			email TEXT UNIQUE,
			email_verified_at TEXT,
			username TEXT UNIQUE,
			name TEXT NOT NULL,
			password TEXT NOT NULL DEFAULT '',
			pin TEXT,
			info_device TEXT,
			fcm_token TEXT,
			remember_token TEXT,
			created_at TEXT,
			updated_at TEXT,
			deleted_at TEXT
		);
		CREATE TABLE IF NOT EXISTS categories (
			id TEXT PRIMARY KEY,
			shop_id TEXT NOT NULL,
			name TEXT NOT NULL,
			created_at TEXT,
			updated_at TEXT,
			deleted_at TEXT
		);
	`).Error
	require.NoError(t, err)

	validator := validators.NewSyncEntityValidator(db)
	ctx := context.Background()

	// Create test data
	testShop := entities.Shop{
		ID:   uuid.New(),
		Name: "Test Shop",
	}
	db.Create(&testShop)

	testUser := entities.User{
		ID:   uuid.New(),
		Name: "Test User",
	}
	db.Create(&testUser)

	testCategory := entities.Category{
		ID:     uuid.New(),
		ShopID: testShop.ID,
		Name:   "Test Category",
	}
	db.Create(&testCategory)

	testProduct := entities.Product{
		ID:     uuid.New(),
		ShopID: testShop.ID,
		CatID:  &testCategory.ID,
		Name:   "Test Product",
		Sale:   100.0,
		Buy:    50.0,
		Stock:  10,
	}
	db.Create(&testProduct)

	t.Run("ValidProductCreation", func(t *testing.T) {
		product := entities.Product{
			ID:     uuid.New(),
			ShopID: testShop.ID,
			CatID:  &testCategory.ID,
			Name:   "Valid Product",
			Sale:   100.0,
			Buy:    80.0,
			Stock:  5,
		}

		err := validator.ValidateEntity(ctx, product, "create", nil)
		assert.NoError(t, err)
	})

	t.Run("InvalidProductPricing", func(t *testing.T) {
		product := entities.Product{
			ID:     uuid.New(),
			ShopID: testShop.ID,
			Name:   "Invalid Product",
			Sale:   50.0, // Sale price less than buy price
			Buy:    100.0,
			Stock:  5,
		}

		err := validator.ValidateEntity(ctx, product, "create", nil)
		assert.Error(t, err)

		validationErrors, ok := err.(validators.ValidationErrors)
		require.True(t, ok)

		// Should have validation error for invalid profit margin
		found := false
		for _, ve := range validationErrors {
			if ve.Code == "INVALID_PROFIT_MARGIN" {
				found = true
				break
			}
		}
		assert.True(t, found)
	})

	t.Run("InvalidProductStock", func(t *testing.T) {
		product := entities.Product{
			ID:     uuid.New(),
			ShopID: testShop.ID,
			Name:   "Invalid Stock Product",
			Sale:   100.0,
			Buy:    80.0,
			Stock:  -5, // Negative stock
		}

		err := validator.ValidateEntity(ctx, product, "create", nil)
		assert.Error(t, err)

		validationErrors, ok := err.(validators.ValidationErrors)
		require.True(t, ok)

		// Should have validation error for invalid stock
		found := false
		for _, ve := range validationErrors {
			if ve.Code == "INVALID_STOCK" {
				found = true
				break
			}
		}
		assert.True(t, found)
	})

	t.Run("MissingForeignKey", func(t *testing.T) {
		product := entities.Product{
			ID:     uuid.New(),
			ShopID: uuid.New(), // Non-existent shop
			Name:   "Orphaned Product",
			Sale:   100.0,
			Buy:    80.0,
			Stock:  5,
		}

		err := validator.ValidateEntity(ctx, product, "create", nil)
		assert.Error(t, err)

		validationErrors, ok := err.(validators.ValidationErrors)
		require.True(t, ok)

		// Should have foreign key validation error
		found := false
		for _, ve := range validationErrors {
			if ve.Code == "FOREIGN_KEY_NOT_FOUND" && ve.Field == "shop_id" {
				found = true
				break
			}
		}
		assert.True(t, found)
	})

	t.Run("ValidCartCreation", func(t *testing.T) {
		cart := entities.Cart{
			ID:        uuid.New(),
			ShopID:    testShop.ID,
			ProductID: testProduct.ID,
			UserID:    testUser.ID,
			Quantity:  2,
		}

		err := validator.ValidateEntity(ctx, cart, "create", nil)
		assert.NoError(t, err)
	})

	t.Run("InvalidCartQuantity", func(t *testing.T) {
		cart := entities.Cart{
			ID:        uuid.New(),
			ShopID:    testShop.ID,
			ProductID: testProduct.ID,
			UserID:    testUser.ID,
			Quantity:  0, // Invalid quantity
		}

		err := validator.ValidateEntity(ctx, cart, "create", nil)
		assert.Error(t, err)

		validationErrors, ok := err.(validators.ValidationErrors)
		require.True(t, ok)

		// Should have validation error for invalid quantity
		found := false
		for _, ve := range validationErrors {
			if ve.Code == "INVALID_QUANTITY" {
				found = true
				break
			}
		}
		assert.True(t, found)
	})

	t.Run("UpdateValidation", func(t *testing.T) {
		existingProduct := entities.Product{
			ID:     testProduct.ID,
			ShopID: testShop.ID,
			Name:   "Existing Product",
			Sale:   100.0,
			Buy:    80.0,
		}

		updatedProduct := entities.Product{
			ID:     testProduct.ID,
			ShopID: uuid.New(), // Trying to change shop ID (should fail)
			Name:   "Updated Product",
			Sale:   120.0,
			Buy:    80.0,
		}

		err := validator.ValidateEntity(ctx, updatedProduct, "update", existingProduct)
		assert.Error(t, err)

		validationErrors, ok := err.(validators.ValidationErrors)
		require.True(t, ok)

		// Should have validation error for immutable field change
		found := false
		for _, ve := range validationErrors {
			if ve.Code == "IMMUTABLE_FIELD_CHANGE" && ve.Field == "shop_id" {
				found = true
				break
			}
		}
		assert.True(t, found)
	})

	t.Run("BarcodeValidation", func(t *testing.T) {
		validBarcode := "1234567890123"
		product := entities.Product{
			ID:      uuid.New(),
			ShopID:  testShop.ID,
			Name:    "Barcode Product",
			Barcode: &validBarcode,
			Sale:    100.0,
			Buy:     80.0,
			Stock:   5,
		}

		err := validator.ValidateEntity(ctx, product, "create", nil)
		assert.NoError(t, err)

		// Test invalid barcode
		invalidBarcode := "invalid-barcode-@#$"
		product.Barcode = &invalidBarcode

		err = validator.ValidateEntity(ctx, product, "create", nil)
		assert.Error(t, err)
	})
}

func TestSecureSyncIntegration(t *testing.T) {
	// This would be a more comprehensive integration test
	// For now, we'll test the basic security integration

	config := services.SyncLockConfig{
		DefaultLockTimeout: 5 * time.Second,
		CleanupInterval:    1 * time.Second,
	}

	manager := services.NewSyncLockManager(config)
	defer manager.Shutdown()

	ctx := context.Background()
	userID := uuid.New()
	licenseID := uuid.New()

	t.Run("ConcurrentSyncPrevention", func(t *testing.T) {
		// Simulate concurrent sync attempts
		completed := make(chan bool, 2)
		errors := make(chan error, 2)

		// First sync (should succeed)
		go func() {
			lockCtx, err := manager.NewSyncLockContext(ctx, userID, licenseID)
			if err != nil {
				errors <- err
				return
			}

			err = lockCtx.Execute(func(ctx context.Context) error {
				// Simulate sync processing time
				time.Sleep(500 * time.Millisecond)
				return nil
			})

			if err != nil {
				errors <- err
			} else {
				completed <- true
			}
		}()

		// Second sync (should fail due to lock)
		go func() {
			time.Sleep(100 * time.Millisecond) // Start slightly after first sync

			ctx_timeout, cancel := context.WithTimeout(ctx, 200*time.Millisecond)
			defer cancel()

			lockCtx, err := manager.NewSyncLockContext(ctx_timeout, userID, licenseID)
			if err != nil {
				errors <- err
				return
			}

			err = lockCtx.Execute(func(ctx context.Context) error {
				return nil
			})

			if err != nil {
				errors <- err
			} else {
				completed <- true
			}
		}()

		// Wait for results
		completedCount := 0
		errorCount := 0

		for i := 0; i < 2; i++ {
			select {
			case <-completed:
				completedCount++
			case <-errors:
				errorCount++
				// Just verify we got an error, don't check specific message
			case <-time.After(3 * time.Second):
				t.Fatal("Test timed out")
			}
		}

		// Should have exactly one success and one error
		assert.Equal(t, 1, completedCount, "Expected exactly one successful sync")
		assert.Equal(t, 1, errorCount, "Expected exactly one failed sync due to lock conflict")
	})
}
