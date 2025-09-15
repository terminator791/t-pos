package services

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/terminator791/t-pos/internal/domain/entities"
	"gorm.io/gorm"
)

// SyncPerformanceOptimizer handles performance optimizations for sync operations
type SyncPerformanceOptimizer struct {
	db *gorm.DB
	config SyncPerformanceConfig
}

// SyncPerformanceConfig contains configuration for performance optimizations
type SyncPerformanceConfig struct {
	EnableBulkValidation bool          `json:"enable_bulk_validation"`
	CacheShopLicenseMap  bool          `json:"cache_shop_license_map"`
	CacheTTL             time.Duration `json:"cache_ttl"`
	BatchSize            int           `json:"batch_size"`
	EnableQueryLogging   bool          `json:"enable_query_logging"`
}

// ShopLicenseMap represents a cached mapping of shop IDs to license IDs
type ShopLicenseMap map[uuid.UUID]uuid.UUID

// EntityValidationCache contains pre-validated entity sets for bulk operations
type EntityValidationCache struct {
	ValidShopIDs     map[uuid.UUID]bool
	ValidProductIDs  map[uuid.UUID]uuid.UUID // ProductID -> ShopID mapping
	ValidUserShops   map[uuid.UUID][]uuid.UUID // UserID -> accessible shop IDs
	CacheTimestamp   time.Time
}

// NewSyncPerformanceOptimizer creates a new performance optimizer
func NewSyncPerformanceOptimizer(db *gorm.DB, config SyncPerformanceConfig) *SyncPerformanceOptimizer {
	return &SyncPerformanceOptimizer{
		db:     db,
		config: config,
	}
}

// BulkValidateProductLicenses replaces individual validateProductLicense calls
func (o *SyncPerformanceOptimizer) BulkValidateProductLicenses(ctx context.Context, products []entities.Product, licenseID uuid.UUID) (map[uuid.UUID]bool, error) {
	if len(products) == 0 {
		return make(map[uuid.UUID]bool), nil
	}

	startTime := time.Now()
	defer func() {
		if o.config.EnableQueryLogging {
			log.Printf("BulkValidateProductLicenses: processed %d products in %v", len(products), time.Since(startTime))
		}
	}()

	// Extract unique shop IDs from products
	shopIDs := make([]uuid.UUID, 0, len(products))
	productToShop := make(map[uuid.UUID]uuid.UUID)
	shopSet := make(map[uuid.UUID]bool)

	for _, product := range products {
		if !shopSet[product.ShopID] {
			shopIDs = append(shopIDs, product.ShopID)
			shopSet[product.ShopID] = true
		}
		productToShop[product.ID] = product.ShopID
	}

	// Single query to validate all shops belong to license
	var validShops []entities.Shop
	err := o.db.WithContext(ctx).
		Select("id").
		Where("id IN (?) AND license_id = ?", shopIDs, licenseID).
		Find(&validShops).Error

	if err != nil {
		return nil, fmt.Errorf("failed to bulk validate shop licenses: %w", err)
	}

	// Create validation map
	validShopSet := make(map[uuid.UUID]bool)
	for _, shop := range validShops {
		validShopSet[shop.ID] = true
	}

	// Map products to validation results
	productValidation := make(map[uuid.UUID]bool)
	for _, product := range products {
		productValidation[product.ID] = validShopSet[product.ShopID]
	}

	return productValidation, nil
}

// BulkValidateTransactionLicenses replaces individual validateTransactionLicense calls
func (o *SyncPerformanceOptimizer) BulkValidateTransactionLicenses(ctx context.Context, transactions []entities.Transaction, licenseID uuid.UUID) (map[uuid.UUID]bool, error) {
	if len(transactions) == 0 {
		return make(map[uuid.UUID]bool), nil
	}

	startTime := time.Now()
	defer func() {
		if o.config.EnableQueryLogging {
			log.Printf("BulkValidateTransactionLicenses: processed %d transactions in %v", len(transactions), time.Since(startTime))
		}
	}()

	// Extract unique shop IDs from transactions
	shopIDs := make([]uuid.UUID, 0, len(transactions))
	transactionToShop := make(map[uuid.UUID]uuid.UUID)
	shopSet := make(map[uuid.UUID]bool)

	for _, transaction := range transactions {
		if !shopSet[transaction.ShopID] {
			shopIDs = append(shopIDs, transaction.ShopID)
			shopSet[transaction.ShopID] = true
		}
		transactionToShop[transaction.ID] = transaction.ShopID
	}

	// Single query to validate all shops belong to license
	var validShops []entities.Shop
	err := o.db.WithContext(ctx).
		Select("id").
		Where("id IN (?) AND license_id = ?", shopIDs, licenseID).
		Find(&validShops).Error

	if err != nil {
		return nil, fmt.Errorf("failed to bulk validate shop licenses for transactions: %w", err)
	}

	// Create validation map
	validShopSet := make(map[uuid.UUID]bool)
	for _, shop := range validShops {
		validShopSet[shop.ID] = true
	}

	// Map transactions to validation results
	transactionValidation := make(map[uuid.UUID]bool)
	for _, transaction := range transactions {
		transactionValidation[transaction.ID] = validShopSet[transaction.ShopID]
	}

	return transactionValidation, nil
}

// BulkValidateStockHistoryAccess replaces individual product queries in stock history validation
func (o *SyncPerformanceOptimizer) BulkValidateStockHistoryAccess(ctx context.Context, stockHistories []entities.StockHistory, accessibleShopIDs []uuid.UUID) (map[uuid.UUID]bool, error) {
	if len(stockHistories) == 0 {
		return make(map[uuid.UUID]bool), nil
	}

	startTime := time.Now()
	defer func() {
		if o.config.EnableQueryLogging {
			log.Printf("BulkValidateStockHistoryAccess: processed %d stock histories in %v", len(stockHistories), time.Since(startTime))
		}
	}()

	// Extract unique product IDs from stock histories
	productIDs := make([]uuid.UUID, 0, len(stockHistories))
	stockHistoryToProduct := make(map[uuid.UUID]uuid.UUID)
	productSet := make(map[uuid.UUID]bool)

	for _, stockHistory := range stockHistories {
		if !productSet[stockHistory.ProductID] {
			productIDs = append(productIDs, stockHistory.ProductID)
			productSet[stockHistory.ProductID] = true
		}
		stockHistoryToProduct[stockHistory.ID] = stockHistory.ProductID
	}

	// Create accessible shop set for fast lookup
	accessibleShopSet := make(map[uuid.UUID]bool)
	for _, shopID := range accessibleShopIDs {
		accessibleShopSet[shopID] = true
	}

	// Single query to get all product shop mappings
	var products []entities.Product
	err := o.db.WithContext(ctx).
		Select("id, shop_id").
		Where("id IN (?)", productIDs).
		Find(&products).Error

	if err != nil {
		return nil, fmt.Errorf("failed to bulk validate stock history product access: %w", err)
	}

	// Create product to shop mapping
	productToShop := make(map[uuid.UUID]uuid.UUID)
	for _, product := range products {
		productToShop[product.ID] = product.ShopID
	}

	// Map stock histories to validation results
	stockHistoryValidation := make(map[uuid.UUID]bool)
	for _, stockHistory := range stockHistories {
		productID := stockHistory.ProductID
		if shopID, exists := productToShop[productID]; exists {
			stockHistoryValidation[stockHistory.ID] = accessibleShopSet[shopID]
		} else {
			// Product doesn't exist
			stockHistoryValidation[stockHistory.ID] = false
		}
	}

	return stockHistoryValidation, nil
}

// BulkFindExistingEntities performs bulk existence checks for any entity type
func (o *SyncPerformanceOptimizer) BulkFindExistingEntities(ctx context.Context, tx *gorm.DB, tableName string, entityIDs []uuid.UUID) (map[uuid.UUID]bool, error) {
	if len(entityIDs) == 0 {
		return make(map[uuid.UUID]bool), nil
	}

	startTime := time.Now()
	defer func() {
		if o.config.EnableQueryLogging {
			log.Printf("BulkFindExistingEntities[%s]: processed %d entities in %v", tableName, len(entityIDs), time.Since(startTime))
		}
	}()

	// Single query to check existence of all entities
	var existingIDs []uuid.UUID
	err := tx.WithContext(ctx).
		Table(tableName).
		Select("id").
		Where("id IN (?)", entityIDs).
		Pluck("id", &existingIDs).Error

	if err != nil {
		return nil, fmt.Errorf("failed to bulk check entity existence in %s: %w", tableName, err)
	}

	// Create existence map
	existenceMap := make(map[uuid.UUID]bool)
	for _, id := range entityIDs {
		existenceMap[id] = false // Default to false
	}
	for _, id := range existingIDs {
		existenceMap[id] = true
	}

	return existenceMap, nil
}

// BatchProcessEntities processes entities in optimized batches
func (o *SyncPerformanceOptimizer) BatchProcessEntities(ctx context.Context, entityCount int, batchSize int, processor func(startIdx, endIdx int) error) error {
	if batchSize <= 0 {
		batchSize = o.config.BatchSize
	}
	if batchSize <= 0 {
		batchSize = 100 // Default batch size
	}

	for i := 0; i < entityCount; i += batchSize {
		endIdx := i + batchSize
		if endIdx > entityCount {
			endIdx = entityCount
		}

		if err := processor(i, endIdx); err != nil {
			return fmt.Errorf("batch processing failed at indices %d-%d: %w", i, endIdx-1, err)
		}
	}

	return nil
}

// GetShopLicenseMapping creates a cached mapping of all shops to their licenses for a given license
func (o *SyncPerformanceOptimizer) GetShopLicenseMapping(ctx context.Context, licenseID uuid.UUID) (ShopLicenseMap, error) {
	startTime := time.Now()
	defer func() {
		if o.config.EnableQueryLogging {
			log.Printf("GetShopLicenseMapping: completed for license %s in %v", licenseID, time.Since(startTime))
		}
	}()

	var shops []entities.Shop
	err := o.db.WithContext(ctx).
		Select("id, license_id").
		Where("license_id = ?", licenseID).
		Find(&shops).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get shop license mapping: %w", err)
	}

	mapping := make(ShopLicenseMap)
	for _, shop := range shops {
		mapping[shop.ID] = shop.LicenseID
	}

	return mapping, nil
}

// OptimizeQueryWithIndexHints adds index hints to common sync queries for better performance
func (o *SyncPerformanceOptimizer) OptimizeQueryWithIndexHints(query *gorm.DB, tableName string, queryType string) *gorm.DB {
	if !o.config.EnableQueryLogging {
		return query
	}

	// Add appropriate index hints based on query type and table
	switch tableName {
	case "products":
		if queryType == "sync_pull" {
			// Hint for shop_id + updated_at index
			return query.Raw("SELECT /*+ USE INDEX (idx_products_shop_updated) */ products.* FROM products WHERE ?", query)
		}
	case "transactions":
		if queryType == "sync_pull" {
			// Hint for shop_id + updated_at index
			return query.Raw("SELECT /*+ USE INDEX (idx_transactions_shop_updated) */ transactions.* FROM transactions WHERE ?", query)
		}
	case "shops":
		if queryType == "license_validation" {
			// Hint for license_id index
			return query.Raw("SELECT /*+ USE INDEX (idx_shops_license) */ shops.id FROM shops WHERE ?", query)
		}
	}

	return query
}