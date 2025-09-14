package services

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/terminator791/t-pos/config"
	"github.com/terminator791/t-pos/internal/domain/entities"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestSyncPerformanceOptimizations(t *testing.T) {
	// Setup test database with simpler configuration for SQLite
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
	})
	require.NoError(t, err)

	// Manually create simple tables for testing (avoiding UUID generation issues in SQLite)
	err = db.Exec("CREATE TABLE shops (id TEXT PRIMARY KEY, license_id TEXT, name TEXT)").Error
	require.NoError(t, err)
	
	err = db.Exec("CREATE TABLE products (id TEXT PRIMARY KEY, shop_id TEXT, name TEXT, updated_at DATETIME)").Error
	require.NoError(t, err)
	
	err = db.Exec("CREATE TABLE transactions (id TEXT PRIMARY KEY, shop_id TEXT, total_price REAL, updated_at DATETIME)").Error
	require.NoError(t, err)
	
	err = db.Exec("CREATE TABLE stock_histories (id TEXT PRIMARY KEY, product_id TEXT, stock INTEGER, updated_at DATETIME)").Error
	require.NoError(t, err)

	// Create test configuration with optimizations enabled
	config := config.SyncConfig{
		EnableBulkValidation:    true,
		EnableQueryOptimization: true,
		EnableCaching:           true,
		CacheTTL:                5 * time.Minute,
		MaxCacheEntries:         1000,
		CacheCleanupInterval:    1 * time.Minute,
		EnableBatchProcessing:   true,
		OptimalBatchSize:        50,
		EnablePerformanceLog:    true,
		MaxResultsPerQuery:      100,
	}

	t.Run("BulkValidateProductLicenses", func(t *testing.T) {
		testBulkValidateProductLicenses(t, db, config)
	})

	t.Run("BulkValidateTransactionLicenses", func(t *testing.T) {
		testBulkValidateTransactionLicenses(t, db, config)
	})

	t.Run("BulkValidateStockHistoryAccess", func(t *testing.T) {
		testBulkValidateStockHistoryAccess(t, db, config)
	})

	t.Run("CacheManagerFunctionality", func(t *testing.T) {
		testCacheManagerFunctionality(t, db, config)
	})

	t.Run("OptimizedQueryPerformance", func(t *testing.T) {
		testOptimizedQueryPerformance(t, db, config)
	})

	t.Run("BatchProcessingEfficiency", func(t *testing.T) {
		testBatchProcessingEfficiency(t, db, config)
	})
}

func testBulkValidateProductLicenses(t *testing.T, db *gorm.DB, config config.SyncConfig) {
	ctx := context.Background()
	
	// Create test data
	licenseID := uuid.New()
	shopID1 := uuid.New()
	shopID2 := uuid.New()
	invalidShopID := uuid.New()

	// Create shops
	shop1 := entities.Shop{ID: shopID1, LicenseID: licenseID, Name: "Shop 1"}
	shop2 := entities.Shop{ID: shopID2, LicenseID: licenseID, Name: "Shop 2"}
	db.Create(&shop1)
	db.Create(&shop2)

	// Create products
	products := []entities.Product{
		{ID: uuid.New(), ShopID: shopID1, Name: "Product 1"},
		{ID: uuid.New(), ShopID: shopID2, Name: "Product 2"},
		{ID: uuid.New(), ShopID: invalidShopID, Name: "Invalid Product"},
	}

	// Test bulk validation
	optimizer := NewSyncPerformanceOptimizer(db, SyncPerformanceConfig{
		EnableBulkValidation: true,
		EnableQueryLogging:   true,
	})

	startTime := time.Now()
	validation, err := optimizer.BulkValidateProductLicenses(ctx, products, licenseID)
	duration := time.Since(startTime)

	require.NoError(t, err)
	assert.True(t, validation[products[0].ID], "Product 1 should be valid")
	assert.True(t, validation[products[1].ID], "Product 2 should be valid")
	assert.False(t, validation[products[2].ID], "Invalid product should be false")

	t.Logf("Bulk validation of %d products completed in %v", len(products), duration)
}

func testBulkValidateTransactionLicenses(t *testing.T, db *gorm.DB, config config.SyncConfig) {
	ctx := context.Background()
	
	// Create test data
	licenseID := uuid.New()
	shopID1 := uuid.New()
	shopID2 := uuid.New()
	invalidShopID := uuid.New()

	// Create shops
	shop1 := entities.Shop{ID: shopID1, LicenseID: licenseID, Name: "Shop 1"}
	shop2 := entities.Shop{ID: shopID2, LicenseID: licenseID, Name: "Shop 2"}
	db.Create(&shop1)
	db.Create(&shop2)

	// Create transactions
	transactions := []entities.Transaction{
		{ID: uuid.New(), ShopID: shopID1, TotalPrice: 100.0},
		{ID: uuid.New(), ShopID: shopID2, TotalPrice: 200.0},
		{ID: uuid.New(), ShopID: invalidShopID, TotalPrice: 300.0},
	}

	// Test bulk validation
	optimizer := NewSyncPerformanceOptimizer(db, SyncPerformanceConfig{
		EnableBulkValidation: true,
		EnableQueryLogging:   true,
	})

	startTime := time.Now()
	validation, err := optimizer.BulkValidateTransactionLicenses(ctx, transactions, licenseID)
	duration := time.Since(startTime)

	require.NoError(t, err)
	assert.True(t, validation[transactions[0].ID], "Transaction 1 should be valid")
	assert.True(t, validation[transactions[1].ID], "Transaction 2 should be valid")
	assert.False(t, validation[transactions[2].ID], "Invalid transaction should be false")

	t.Logf("Bulk validation of %d transactions completed in %v", len(transactions), duration)
}

func testBulkValidateStockHistoryAccess(t *testing.T, db *gorm.DB, config config.SyncConfig) {
	ctx := context.Background()
	
	// Create test data
	licenseID := uuid.New()
	shopID1 := uuid.New()
	shopID2 := uuid.New()
	productID1 := uuid.New()
	productID2 := uuid.New()
	invalidProductID := uuid.New()

	// Create shops
	shop1 := entities.Shop{ID: shopID1, LicenseID: licenseID, Name: "Shop 1"}
	shop2 := entities.Shop{ID: shopID2, LicenseID: licenseID, Name: "Shop 2"}
	db.Create(&shop1)
	db.Create(&shop2)

	// Create products
	product1 := entities.Product{ID: productID1, ShopID: shopID1, Name: "Product 1"}
	product2 := entities.Product{ID: productID2, ShopID: shopID2, Name: "Product 2"}
	db.Create(&product1)
	db.Create(&product2)

	// Create stock histories
	stockHistories := []entities.StockHistory{
		{ID: uuid.New(), ProductID: productID1, Stock: 100},
		{ID: uuid.New(), ProductID: productID2, Stock: 200},
		{ID: uuid.New(), ProductID: invalidProductID, Stock: 300},
	}

	// Test bulk validation
	optimizer := NewSyncPerformanceOptimizer(db, SyncPerformanceConfig{
		EnableBulkValidation: true,
		EnableQueryLogging:   true,
	})

	accessibleShops := []uuid.UUID{shopID1, shopID2}
	startTime := time.Now()
	validation, err := optimizer.BulkValidateStockHistoryAccess(ctx, stockHistories, accessibleShops)
	duration := time.Since(startTime)

	require.NoError(t, err)
	assert.True(t, validation[stockHistories[0].ID], "Stock history 1 should be accessible")
	assert.True(t, validation[stockHistories[1].ID], "Stock history 2 should be accessible")
	assert.False(t, validation[stockHistories[2].ID], "Invalid stock history should not be accessible")

	t.Logf("Bulk validation of %d stock histories completed in %v", len(stockHistories), duration)
}

func testCacheManagerFunctionality(t *testing.T, db *gorm.DB, config config.SyncConfig) {
	ctx := context.Background()
	
	// Create test data
	licenseID := uuid.New()
	shopID1 := uuid.New()
	shopID2 := uuid.New()

	// Create shops
	shop1 := entities.Shop{ID: shopID1, LicenseID: licenseID, Name: "Shop 1"}
	shop2 := entities.Shop{ID: shopID2, LicenseID: licenseID, Name: "Shop 2"}
	db.Create(&shop1)
	db.Create(&shop2)

	// Create cache manager
	cacheConfig := SyncCacheConfig{
		EnableCaching:           true,
		ShopLicenseCacheTTL:     5 * time.Minute,
		ProductShopCacheTTL:     5 * time.Minute,
		UserShopsCacheTTL:       5 * time.Minute,
		MaxCacheEntries:         100,
		CacheCleanupInterval:    1 * time.Minute,
		EnableCacheStatistics:   true,
	}
	
	cacheManager := NewSyncCacheManager(db, cacheConfig)

	// Test shop license mapping cache
	t.Run("ShopLicenseMapping", func(t *testing.T) {
		// First call - cache miss
		startTime := time.Now()
		mapping1, err := cacheManager.GetShopLicenseMapping(ctx, licenseID)
		firstCallDuration := time.Since(startTime)
		require.NoError(t, err)
		assert.Len(t, mapping1, 2)

		// Second call - cache hit (should be faster)
		startTime = time.Now()
		mapping2, err := cacheManager.GetShopLicenseMapping(ctx, licenseID)
		secondCallDuration := time.Since(startTime)
		require.NoError(t, err)
		assert.Equal(t, mapping1, mapping2)

		t.Logf("Cache performance: First call %v, Second call %v", firstCallDuration, secondCallDuration)
		// Second call should be significantly faster (cache hit)
		assert.True(t, secondCallDuration < firstCallDuration/2, "Cache hit should be much faster")
	})

	// Test cache statistics
	stats := cacheManager.GetCacheStats()
	assert.True(t, stats.Hits >= 1, "Should have at least one cache hit")
	assert.True(t, stats.Misses >= 1, "Should have at least one cache miss")
	t.Logf("Cache stats: Hits=%d, Misses=%d, Entries=%d", stats.Hits, stats.Misses, stats.Entries)
}

func testOptimizedQueryPerformance(t *testing.T, db *gorm.DB, config config.SyncConfig) {
	ctx := context.Background()
	
	// Create test data
	licenseID := uuid.New()
	shopID := uuid.New()

	// Create shop
	shop := entities.Shop{ID: shopID, LicenseID: licenseID, Name: "Test Shop"}
	db.Create(&shop)

	// Create multiple products for performance testing
	products := make([]entities.Product, 50)
	for i := 0; i < 50; i++ {
		products[i] = entities.Product{
			ID:        uuid.New(),
			ShopID:    shopID,
			Name:      "Test Product",
			UpdatedAt: time.Now().Add(time.Duration(i) * time.Second),
		}
	}
	db.CreateInBatches(products, 10)

	// Test bulk existence check
	optimizer := NewSyncPerformanceOptimizer(db, SyncPerformanceConfig{
		EnableBulkValidation: true,
		BatchSize:            20,
		EnableQueryLogging:   true,
	})

	productIDs := make([]uuid.UUID, len(products))
	for i, product := range products {
		productIDs[i] = product.ID
	}

	startTime := time.Now()
	existence, err := optimizer.BulkFindExistingEntities(ctx, db, "products", productIDs)
	duration := time.Since(startTime)

	require.NoError(t, err)
	assert.Len(t, existence, len(products))
	
	// All products should exist
	for _, productID := range productIDs {
		assert.True(t, existence[productID], "Product should exist")
	}

	t.Logf("Bulk existence check for %d products completed in %v", len(products), duration)
}

func testBatchProcessingEfficiency(t *testing.T, db *gorm.DB, config config.SyncConfig) {
	ctx := context.Background()
	
	optimizer := NewSyncPerformanceOptimizer(db, SyncPerformanceConfig{
		BatchSize:          10,
		EnableQueryLogging: true,
	})

	// Test batch processing with different batch sizes
	totalItems := 47 // Prime number to test edge cases
	
	t.Run("StandardBatchSize", func(t *testing.T) {
		processedItems := 0
		batches := 0
		
		startTime := time.Now()
		err := optimizer.BatchProcessEntities(ctx, totalItems, 10, func(startIdx, endIdx int) error {
			batches++
			processedItems += (endIdx - startIdx)
			t.Logf("Processing batch %d: items %d-%d (%d items)", batches, startIdx, endIdx-1, endIdx-startIdx)
			return nil
		})
		duration := time.Since(startTime)
		
		require.NoError(t, err)
		assert.Equal(t, totalItems, processedItems)
		assert.Equal(t, 5, batches) // 47 items in batches of 10: 10+10+10+10+7 = 5 batches
		
		t.Logf("Batch processing: %d items in %d batches, completed in %v", totalItems, batches, duration)
	})

	t.Run("SingleBatch", func(t *testing.T) {
		processedItems := 0
		batches := 0
		
		err := optimizer.BatchProcessEntities(ctx, totalItems, totalItems+10, func(startIdx, endIdx int) error {
			batches++
			processedItems += (endIdx - startIdx)
			return nil
		})
		
		require.NoError(t, err)
		assert.Equal(t, totalItems, processedItems)
		assert.Equal(t, 1, batches) // Should be processed as single batch
	})
}

func BenchmarkSyncPerformanceOptimizations(b *testing.B) {
	// Setup test database
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(b, err)

	// Migrate the schema
	err = db.AutoMigrate(&entities.Shop{}, &entities.Product{})
	require.NoError(b, err)

	// Create test data
	licenseID := uuid.New()
	shopID := uuid.New()
	shop := entities.Shop{ID: shopID, LicenseID: licenseID, Name: "Benchmark Shop"}
	db.Create(&shop)

	// Create products for benchmarking
	products := make([]entities.Product, 1000)
	for i := 0; i < 1000; i++ {
		products[i] = entities.Product{
			ID:     uuid.New(),
			ShopID: shopID,
			Name:   "Benchmark Product",
		}
	}
	db.CreateInBatches(products, 100)

	optimizer := NewSyncPerformanceOptimizer(db, SyncPerformanceConfig{
		EnableBulkValidation: true,
		EnableQueryLogging:   false, // Disable logging for benchmarks
	})

	b.Run("BulkValidation", func(b *testing.B) {
		ctx := context.Background()
		b.ResetTimer()
		
		for i := 0; i < b.N; i++ {
			_, err := optimizer.BulkValidateProductLicenses(ctx, products[:100], licenseID)
			require.NoError(b, err)
		}
	})

	b.Run("IndividualValidation", func(b *testing.B) {
		b.ResetTimer()
		
		for i := 0; i < b.N; i++ {
			// Simulate individual validation queries
			for j := 0; j < 100; j++ {
				var count int64
				db.Model(&entities.Shop{}).Where("id = ? AND license_id = ?", products[j].ShopID, licenseID).Count(&count)
			}
		}
	})
}

func TestPerformanceOptimizationFallbacks(t *testing.T) {
	// Test that performance optimizations gracefully fall back to original implementation
	// when errors occur or optimizations are disabled
	
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	t.Run("DisabledOptimizations", func(t *testing.T) {
		// Test with optimizations disabled
		optimizer := NewSyncPerformanceOptimizer(db, SyncPerformanceConfig{
			EnableBulkValidation: false,
			EnableQueryLogging:   false,
		})

		products := []entities.Product{
			{ID: uuid.New(), ShopID: uuid.New(), Name: "Test Product"},
		}

		// Should handle gracefully even with no data in database
		validation, err := optimizer.BulkValidateProductLicenses(context.Background(), products, uuid.New())
		require.NoError(t, err)
		assert.Len(t, validation, 1)
		assert.False(t, validation[products[0].ID]) // Should be false since no data exists
	})

	t.Run("EmptyInputHandling", func(t *testing.T) {
		optimizer := NewSyncPerformanceOptimizer(db, SyncPerformanceConfig{
			EnableBulkValidation: true,
		})

		// Test empty input handling
		validation, err := optimizer.BulkValidateProductLicenses(context.Background(), []entities.Product{}, uuid.New())
		require.NoError(t, err)
		assert.Empty(t, validation)
	})
}