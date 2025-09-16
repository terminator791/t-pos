package services

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/terminator791/t-pos/config"
	"github.com/terminator791/t-pos/internal/domain/dto"
	"github.com/terminator791/t-pos/internal/domain/entities"
	"github.com/terminator791/t-pos/internal/domain/repositories"
	"github.com/terminator791/t-pos/internal/domain/validators"
	"gorm.io/gorm"
)

// Sync configuration constants
const (
	// Maximum number of entities to process in a single batch
	DefaultBatchSize = 100

	// Maximum number of entities per sync operation to prevent memory issues
	MaxEntitiesPerSync = 1000

	// Default transaction timeout
	DefaultTransactionTimeout = 30 * time.Second

	// Maximum transaction timeout for large syncs
	MaxTransactionTimeout = 5 * time.Minute
)

// SyncService handles data synchronization between mobile and server
type SyncService struct {
	db                     *gorm.DB
	cartRepo               repositories.CartRepository
	categoryRepo           repositories.CategoryRepository
	expenseRepo            repositories.ExpenseRepository
	historyRepo            repositories.HistoryRepository
	paymentRepo            repositories.PaymentRepository
	productRepo            repositories.ProductRepository
	receiptRepo            repositories.ReceiptRepository
	shopRepo               repositories.ShopRepository
	stockHistoryRepo       repositories.StockHistoryRepository
	transactionRepo        repositories.TransactionRepository
	transactionProductRepo repositories.TransactionProductRepository
	userRepo               repositories.UserRepository
	conflictStrategy       dto.ConflictResolutionStrategy
	config                 config.SyncConfig

	// Security enhancements
	validator   *validators.SyncEntityValidator
	lockManager *SyncLockManager

	// Performance optimizations (Session 3)
	optimizer    *SyncPerformanceOptimizer
	cacheManager *SyncCacheManager
}

// NewSyncService creates a new sync service instance
func NewSyncService(
	db *gorm.DB,
	cartRepo repositories.CartRepository,
	categoryRepo repositories.CategoryRepository,
	expenseRepo repositories.ExpenseRepository,
	historyRepo repositories.HistoryRepository,
	paymentRepo repositories.PaymentRepository,
	productRepo repositories.ProductRepository,
	receiptRepo repositories.ReceiptRepository,
	shopRepo repositories.ShopRepository,
	stockHistoryRepo repositories.StockHistoryRepository,
	transactionRepo repositories.TransactionRepository,
	transactionProductRepo repositories.TransactionProductRepository,
	userRepo repositories.UserRepository,
	syncConfig config.SyncConfig,
) *SyncService {
	// Initialize security components
	validator := validators.NewSyncEntityValidator(db)
	lockManager := NewSyncLockManager(SyncLockConfig{
		DefaultLockTimeout: time.Duration(syncConfig.TransactionTimeout) * time.Second,
		CleanupInterval:    10 * time.Second,
		MaxLockHoldTime:    5 * time.Minute,
	})

	// Initialize performance optimization components
	optimizer := NewSyncPerformanceOptimizer(db, SyncPerformanceConfig{
		EnableBulkValidation: syncConfig.EnableBulkValidation,
		CacheShopLicenseMap:  syncConfig.EnableCaching,
		CacheTTL:             syncConfig.CacheTTL,
		BatchSize:            syncConfig.OptimalBatchSize,
		EnableQueryLogging:   syncConfig.EnablePerformanceLog,
	})

	cacheManager := NewSyncCacheManager(db, SyncCacheConfig{
		EnableCaching:         syncConfig.EnableCaching,
		ShopLicenseCacheTTL:   syncConfig.CacheTTL,
		ProductShopCacheTTL:   syncConfig.CacheTTL,
		UserShopsCacheTTL:     syncConfig.CacheTTL,
		MaxCacheEntries:       syncConfig.MaxCacheEntries,
		CacheCleanupInterval:  syncConfig.CacheCleanupInterval,
		EnableCacheStatistics: syncConfig.EnablePerformanceLog,
	})

	return &SyncService{
		db:                     db,
		cartRepo:               cartRepo,
		categoryRepo:           categoryRepo,
		expenseRepo:            expenseRepo,
		historyRepo:            historyRepo,
		paymentRepo:            paymentRepo,
		productRepo:            productRepo,
		receiptRepo:            receiptRepo,
		shopRepo:               shopRepo,
		stockHistoryRepo:       stockHistoryRepo,
		transactionRepo:        transactionRepo,
		transactionProductRepo: transactionProductRepo,
		userRepo:               userRepo,
		conflictStrategy:       dto.LastWriteWins,
		config:                 syncConfig,
		validator:              validator,
		lockManager:            lockManager,
		optimizer:              optimizer,
		cacheManager:           cacheManager,
	}
}

// ProcessSyncWithSecurityEnhancements handles secure synchronization with distributed locking and comprehensive validation
func (s *SyncService) ProcessSyncWithSecurityEnhancements(ctx context.Context, req dto.SyncRequest, syncContext dto.SyncContext) (*dto.SyncResponse, error) {
	// Acquire distributed lock to prevent concurrent sync operations for the same user/license
	lockCtx, err := s.lockManager.NewSyncLockContext(ctx, syncContext.UserID, syncContext.LicenseID)
	if err != nil {
		return nil, fmt.Errorf("failed to acquire sync lock: %w", err)
	}

	// Execute sync within the locked context
	var response *dto.SyncResponse
	err = lockCtx.Execute(func(lockedCtx context.Context) error {
		var syncErr error
		response, syncErr = s.ProcessSyncWithRoleAccess(lockedCtx, req, syncContext)
		return syncErr
	})

	if err != nil {
		return nil, fmt.Errorf("secure sync processing failed: %w", err)
	}

	return response, nil
}

// Shutdown gracefully shuts down the sync service
func (s *SyncService) Shutdown() {
	if s.lockManager != nil {
		s.lockManager.Shutdown()
	}
}

// ProcessSyncWithRoleAccess handles the complete synchronization process with role-based access control
func (s *SyncService) ProcessSyncWithRoleAccess(ctx context.Context, req dto.SyncRequest, syncContext dto.SyncContext) (*dto.SyncResponse, error) {
	startTime := time.Now()

	// Validate sync request size to prevent memory issues
	if err := s.validateSyncRequest(req); err != nil {
		return nil, fmt.Errorf("sync request validation failed: %w", err)
	}

	// Log role-based sync operation
	log.Printf("Starting role-based sync for user %s (role: %s), license %s, shops: %v",
		syncContext.UserID, syncContext.UserRole, syncContext.LicenseID, syncContext.AccessibleShopIDs)

	response := &dto.SyncResponse{
		SyncTimestamp: time.Now(),
		Conflicts:     make([]dto.ConflictInfo, 0),
		Errors:        make([]dto.SyncError, 0),
		Stats: dto.SyncStats{
			ProcessedEntities: make(map[string]int),
			CreatedEntities:   make(map[string]int),
			UpdatedEntities:   make(map[string]int),
		},
	}

	// Create context with timeout for database operations
	ctxWithTimeout, cancel := context.WithTimeout(ctx, s.config.TransactionTimeout)
	defer cancel()

	// CRITICAL FIX: Use separate transactions for push and pull to prevent transaction abort errors
	// This prevents a failed operation in one phase from affecting the other

	// Phase 1: Push - Process incoming changes from mobile with role-based filtering
	if err := s.processPushWithTransaction(ctxWithTimeout, req, syncContext, response); err != nil {
		log.Printf("Push phase failed for user %s: %v", syncContext.UserID, err)
		// Don't return error immediately - allow pull phase to proceed
		s.addDetailedError(response, "sync", uuid.Nil, "push_failed", err.Error(),
			map[string]interface{}{"phase": "push", "user_id": syncContext.UserID})
	}

	// Phase 2: Pull - Get server changes since last sync with role-based filtering
	if err := s.processPullWithTransaction(ctxWithTimeout, req.LastSyncTimestamp, syncContext, response); err != nil {
		log.Printf("Pull phase failed for user %s: %v", syncContext.UserID, err)
		s.addDetailedError(response, "sync", uuid.Nil, "pull_failed", err.Error(),
			map[string]interface{}{"phase": "pull", "user_id": syncContext.UserID})
	}

	// Clean up response - remove empty arrays and fix conflict data
	s.cleanupResponse(response)

	// Calculate final stats
	response.Stats.ProcessingTimeMs = time.Since(startTime).Milliseconds()
	response.Stats.ConflictCount = len(response.Conflicts)
	response.Stats.ErrorCount = len(response.Errors)

	log.Printf("Role-based sync completed for user %s (role: %s): %d conflicts, %d errors, %dms, %d entities processed",
		syncContext.UserID, syncContext.UserRole, response.Stats.ConflictCount, response.Stats.ErrorCount,
		response.Stats.ProcessingTimeMs, s.getTotalProcessedEntities(response.Stats))

	return response, nil
}

// processPushWithTransaction handles push operations in an isolated transaction
func (s *SyncService) processPushWithTransaction(ctx context.Context, req dto.SyncRequest, syncContext dto.SyncContext, response *dto.SyncResponse) error {
	// CRITICAL FIX: Use a single transaction with proper error handling to prevent SQL abort errors
	// Instead of nested transactions, use savepoints for error isolation

	tx := s.db.Begin(&sql.TxOptions{
		Isolation: sql.LevelReadCommitted,
	})
	if tx.Error != nil {
		return fmt.Errorf("failed to start push transaction: %w", tx.Error)
	}

	// Ensure transaction is properly cleaned up
	defer func() {
		if r := recover(); r != nil {
			if rollbackErr := tx.Rollback().Error; rollbackErr != nil {
				log.Printf("Failed to rollback transaction after panic: %v", rollbackErr)
			}
			log.Printf("Push transaction rolled back due to panic: %v", r)
			panic(r)
		}
	}()

	// Process push changes with enhanced error handling but without nested transactions
	if err := s.pushChangesWithRoleAccessSafe(ctx, tx, req, syncContext, response); err != nil {
		if rollbackErr := tx.Rollback().Error; rollbackErr != nil {
			log.Printf("Failed to rollback transaction: %v", rollbackErr)
		}
		return fmt.Errorf("failed to push changes: %w", err)
	}

	// Commit push transaction
	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("failed to commit push transaction: %w", err)
	}

	log.Printf("Push phase completed successfully for user %s", syncContext.UserID)
	return nil
}

// processPullWithTransaction handles pull operations in an isolated transaction
func (s *SyncService) processPullWithTransaction(ctx context.Context, lastSync *time.Time, syncContext dto.SyncContext, response *dto.SyncResponse) error {
	// Start a dedicated transaction for pull operations
	tx := s.db.Begin(&sql.TxOptions{
		Isolation: sql.LevelReadCommitted,
		ReadOnly:  true, // Pull operations are read-only
	})
	if tx.Error != nil {
		return fmt.Errorf("failed to start pull transaction: %w", tx.Error)
	}

	// Ensure transaction is properly cleaned up
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			log.Printf("Pull transaction rolled back due to panic: %v", r)
			panic(r)
		}
	}()

	// Process pull changes with enhanced error handling
	if err := s.pullChangesWithRoleAccess(ctx, tx, lastSync, syncContext, response); err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to pull changes: %w", err)
	}

	// Commit pull transaction
	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("failed to commit pull transaction: %w", err)
	}

	log.Printf("Pull phase completed successfully for user %s", syncContext.UserID)
	return nil
}

// ProcessSync handles the complete synchronization process (legacy method for backward compatibility)
func (s *SyncService) ProcessSync(ctx context.Context, req dto.SyncRequest, licenseID uuid.UUID, userID uuid.UUID) (*dto.SyncResponse, error) {
	// Create a default sync context for backward compatibility
	syncContext := dto.SyncContext{
		UserID:            userID,
		UserRole:          "legacy", // Indicates legacy sync without role-based access
		LicenseID:         licenseID,
		AccessibleShopIDs: nil,  // No filtering for legacy sync
		HasGlobalAccess:   true, // Legacy behavior - global access
	}

	return s.ProcessSyncWithRoleAccess(ctx, req, syncContext)
}

// validateSyncRequest validates the sync request to prevent memory and performance issues
func (s *SyncService) validateSyncRequest(req dto.SyncRequest) error {
	totalEntities := len(req.Carts) + len(req.Categories) + len(req.Products) +
		len(req.Transactions) + len(req.Payments) + len(req.Expenses) +
		len(req.Receipts) + len(req.Histories) + len(req.Shops) +
		len(req.StockHistories) + len(req.TransactionProducts) + len(req.Users)

	// Check total entity count limit
	if totalEntities > s.config.MaxEntitiesPerSync {
		return fmt.Errorf("sync request too large: %d entities exceeds maximum of %d",
			totalEntities, s.config.MaxEntitiesPerSync)
	}

	// CRITICAL FIX: Enhanced memory validation with actual memory estimation
	estimatedMemoryMB := s.calculateMemoryUsage(req)
	if estimatedMemoryMB > float64(s.config.MaxMemoryUsageMB) {
		return fmt.Errorf("sync request memory usage too large: estimated %.2f MB exceeds maximum of %d MB",
			estimatedMemoryMB, s.config.MaxMemoryUsageMB)
	}

	// Validate individual entity type limits to prevent single type from dominating
	entityCounts := map[string]int{
		"carts":                len(req.Carts),
		"categories":           len(req.Categories),
		"products":             len(req.Products),
		"transactions":         len(req.Transactions),
		"payments":             len(req.Payments),
		"expenses":             len(req.Expenses),
		"receipts":             len(req.Receipts),
		"histories":            len(req.Histories),
		"shops":                len(req.Shops),
		"stock_histories":      len(req.StockHistories),
		"transaction_products": len(req.TransactionProducts),
		"users":                len(req.Users),
	}

	maxEntitiesPerType := s.config.MaxEntitiesPerSync / 2 // No single entity type should exceed half the limit
	for entityType, count := range entityCounts {
		if count > maxEntitiesPerType {
			return fmt.Errorf("%s count too large: %d exceeds maximum of %d",
				entityType, count, maxEntitiesPerType)
		}
	}

	// Log validation success for monitoring
	if s.config.EnablePerformanceLog {
		log.Printf("Sync request validation passed: %d entities, estimated %.2f MB memory",
			totalEntities, estimatedMemoryMB)
	}

	return nil
}

// calculateMemoryUsage estimates memory usage for a sync request
func (s *SyncService) calculateMemoryUsage(req dto.SyncRequest) float64 {
	// Entity-specific size estimates in KB (based on typical JSON serialized size)
	entitySizes := map[string]float64{
		"carts":                0.5, // Simple cart items
		"categories":           0.3, // Small category objects
		"products":             2.0, // Products can have images/descriptions
		"transactions":         1.5, // Transaction records
		"payments":             0.8, // Payment records
		"expenses":             1.0, // Expense records
		"receipts":             3.0, // Receipts can be large with text content
		"histories":            1.2, // History records
		"shops":                1.0, // Shop information
		"stock_histories":      0.7, // Stock movement records
		"transaction_products": 0.6, // Product-transaction relationships
		"users":                1.0, // User information
	}

	totalSizeKB := 0.0
	totalSizeKB += float64(len(req.Carts)) * entitySizes["carts"]
	totalSizeKB += float64(len(req.Categories)) * entitySizes["categories"]
	totalSizeKB += float64(len(req.Products)) * entitySizes["products"]
	totalSizeKB += float64(len(req.Transactions)) * entitySizes["transactions"]
	totalSizeKB += float64(len(req.Payments)) * entitySizes["payments"]
	totalSizeKB += float64(len(req.Expenses)) * entitySizes["expenses"]
	totalSizeKB += float64(len(req.Receipts)) * entitySizes["receipts"]
	totalSizeKB += float64(len(req.Histories)) * entitySizes["histories"]
	totalSizeKB += float64(len(req.Shops)) * entitySizes["shops"]
	totalSizeKB += float64(len(req.StockHistories)) * entitySizes["stock_histories"]
	totalSizeKB += float64(len(req.TransactionProducts)) * entitySizes["transaction_products"]
	totalSizeKB += float64(len(req.Users)) * entitySizes["users"]

	// Convert to MB and add overhead for processing (JSON parsing, GORM operations, etc.)
	totalSizeMB := totalSizeKB / 1024.0
	processingOverheadMultiplier := 3.0 // Assume 3x overhead for processing

	return totalSizeMB * processingOverheadMultiplier
}

// getTotalProcessedEntities calculates total entities processed across all types
func (s *SyncService) getTotalProcessedEntities(stats dto.SyncStats) int {
	total := 0
	for _, count := range stats.ProcessedEntities {
		total += count
	}
	return total
}

// pushChangesWithRoleAccessSafe processes incoming changes from mobile client with role-based filtering
// This version avoids nested transactions to prevent SQL transaction abort errors
func (s *SyncService) pushChangesWithRoleAccessSafe(ctx context.Context, tx *gorm.DB, req dto.SyncRequest, syncContext dto.SyncContext, response *dto.SyncResponse) error {
	// CRITICAL FIX: Filter and validate entities based on role access BEFORE processing
	// This prevents validation errors from aborting the entire transaction
	filteredReq, filterStats := s.filterAndValidateSyncRequest(req, syncContext, response)

	// Log filtering results with detailed stats
	log.Printf("Role-based filtering for user %s (role: %s): %s",
		syncContext.UserID, syncContext.UserRole, filterStats)

	// Add filter warnings to response instead of failing for inaccessible data
	for _, warning := range s.generateFilterWarnings(req, filteredReq, syncContext) {
		response.Errors = append(response.Errors, warning)
	}

	// CRITICAL FIX: Process entities in correct dependency order to avoid foreign key violations
	// This ensures that when validating entities, their referenced entities exist in the database

	// Level 1: Independent entities (no foreign key dependencies)
	if syncContext.UserRole != "cashier" {
		// 1. Users (no dependencies)
		if err := s.pushUsersSafe(ctx, tx, filteredReq.Users, syncContext.LicenseID, response); err != nil {
			return fmt.Errorf("failed to push users: %w", err)
		}

		// 2. Shops (references Users via UserID)
		if err := s.pushShopsSafe(ctx, tx, filteredReq.Shops, syncContext.LicenseID, response); err != nil {
			return fmt.Errorf("failed to push shops: %w", err)
		}
	} else {
		// Log filtered operations for cashiers
		if len(req.Shops) > 0 {
			log.Printf("Filtered shop sync for cashier user %s: %d shops filtered out", syncContext.UserID, len(req.Shops))
			s.addDetailedError(response, "shops", uuid.Nil, "filtered", "Cashiers cannot sync shop data",
				map[string]interface{}{"role": "cashier", "filtered_count": len(req.Shops)})
		}
		if len(req.Users) > 0 {
			log.Printf("Filtered user sync for cashier user %s: %d users filtered out", syncContext.UserID, len(req.Users))
			s.addDetailedError(response, "users", uuid.Nil, "filtered", "Cashiers cannot sync user data",
				map[string]interface{}{"role": "cashier", "filtered_count": len(req.Users)})
		}
	}

	// Level 2: Entities that depend only on Shops and/or Users
	// 3. Categories (references Shops via ShopID)
	if err := s.pushCategoriesSafe(ctx, tx, filteredReq.Categories, syncContext.LicenseID, response); err != nil {
		return fmt.Errorf("failed to push categories: %w", err)
	}

	// 4. Expenses (references Shops via ShopID)
	if err := s.pushExpensesSafe(ctx, tx, filteredReq.Expenses, syncContext.LicenseID, response); err != nil {
		return fmt.Errorf("failed to push expenses: %w", err)
	}

	// 5. Payments (references Shops via ShopID)
	if err := s.pushPaymentsSafe(ctx, tx, filteredReq.Payments, syncContext.LicenseID, response); err != nil {
		return fmt.Errorf("failed to push payments: %w", err)
	}

	// 6. Transactions (references Shops via ShopID and Users via CashierID)
	if err := s.pushTransactionsSafe(ctx, tx, filteredReq.Transactions, syncContext.LicenseID, response); err != nil {
		return fmt.Errorf("failed to push transactions: %w", err)
	}

	// Level 3: Entities that depend on Categories and/or Shops
	// 7. Products (references Shops via ShopID and optionally Categories via CatID)
	if err := s.pushProductsSafe(ctx, tx, filteredReq.Products, syncContext.LicenseID, response); err != nil {
		return fmt.Errorf("failed to push products: %w", err)
	}

	// Level 4: Entities that depend on Products and/or other Level 2/3 entities
	// 8. Carts (references Shops via ShopID, Products via ProductID, Users via UserID)
	if err := s.pushCartsSafe(ctx, tx, filteredReq.Carts, syncContext.LicenseID, response); err != nil {
		return fmt.Errorf("failed to push carts: %w", err)
	}

	// 9. Transaction Products (references Transactions via TransactionID)
	if err := s.pushTransactionProductsSafe(ctx, tx, filteredReq.TransactionProducts, filteredReq.Transactions, syncContext, response); err != nil {
		return fmt.Errorf("failed to push transaction products: %w", err)
	}

	// 10. Stock Histories (references Products via ProductID)
	if err := s.pushStockHistoriesSafe(ctx, tx, filteredReq.StockHistories, filteredReq.Products, syncContext, response); err != nil {
		return fmt.Errorf("failed to push stock histories: %w", err)
	}

	// Level 5: Entities that depend on Level 2 entities
	// 11. Receipts (references Shops via ShopID and Payments via PaymentsID)
	if err := s.pushReceiptsSafe(ctx, tx, filteredReq.Receipts, syncContext.LicenseID, response); err != nil {
		return fmt.Errorf("failed to push receipts: %w", err)
	}

	// 12. Histories (references Shops via ShopID and Transactions via TransactionID)
	if err := s.pushHistoriesSafe(ctx, tx, filteredReq.Histories, syncContext.LicenseID, response); err != nil {
		return fmt.Errorf("failed to push histories: %w", err)
	}

	return nil
}

// filterAndValidateSyncRequest filters and validates sync request entities based on role access
func (s *SyncService) filterAndValidateSyncRequest(req dto.SyncRequest, syncContext dto.SyncContext, response *dto.SyncResponse) (dto.SyncRequest, string) {
	if syncContext.HasGlobalAccess {
		return req, "global access - no filtering applied"
	}

	// Create a map for fast shop access checking
	accessibleShops := make(map[uuid.UUID]bool)
	for _, shopID := range syncContext.AccessibleShopIDs {
		accessibleShops[shopID] = true
	}

	// Debug logging for accessibility mapping
	log.Printf("DEBUG: filterAndValidateSyncRequest - User %s (role: %s) accessible shops map: %v",
		syncContext.UserID, syncContext.UserRole, accessibleShops)

	filteredReq := dto.SyncRequest{
		LastSyncTimestamp: req.LastSyncTimestamp,
	}

	stats := make(map[string]string)

	// Filter entities by accessible shops
	originalCount := len(req.Carts)
	for _, cart := range req.Carts {
		if accessibleShops[cart.ShopID] {
			filteredReq.Carts = append(filteredReq.Carts, cart)
		}
	}
	if originalCount > 0 {
		stats["carts"] = fmt.Sprintf("%d→%d", originalCount, len(filteredReq.Carts))
	}

	originalCount = len(req.Categories)
	for _, category := range req.Categories {
		if accessibleShops[category.ShopID] {
			filteredReq.Categories = append(filteredReq.Categories, category)
		}
	}
	if originalCount > 0 {
		stats["categories"] = fmt.Sprintf("%d→%d", originalCount, len(filteredReq.Categories))
	}

	originalCount = len(req.Products)
	for _, product := range req.Products {
		if accessibleShops[product.ShopID] {
			filteredReq.Products = append(filteredReq.Products, product)
		}
	}
	if originalCount > 0 {
		stats["products"] = fmt.Sprintf("%d→%d", originalCount, len(filteredReq.Products))
	}

	originalCount = len(req.Transactions)
	for _, transaction := range req.Transactions {
		if accessibleShops[transaction.ShopID] {
			filteredReq.Transactions = append(filteredReq.Transactions, transaction)
		}
	}
	if originalCount > 0 {
		stats["transactions"] = fmt.Sprintf("%d→%d", originalCount, len(filteredReq.Transactions))
	}

	originalCount = len(req.Expenses)
	for _, expense := range req.Expenses {
		if accessibleShops[expense.ShopID] {
			filteredReq.Expenses = append(filteredReq.Expenses, expense)
		}
	}
	if originalCount > 0 {
		stats["expenses"] = fmt.Sprintf("%d→%d", originalCount, len(filteredReq.Expenses))
	}

	originalCount = len(req.Payments)
	for _, payment := range req.Payments {
		if accessibleShops[payment.ShopID] {
			filteredReq.Payments = append(filteredReq.Payments, payment)
		}
	}
	if originalCount > 0 {
		stats["payments"] = fmt.Sprintf("%d→%d", originalCount, len(filteredReq.Payments))
	}

	originalCount = len(req.Receipts)
	for _, receipt := range req.Receipts {
		if accessibleShops[receipt.ShopID] {
			filteredReq.Receipts = append(filteredReq.Receipts, receipt)
		}
	}
	if originalCount > 0 {
		stats["receipts"] = fmt.Sprintf("%d→%d", originalCount, len(filteredReq.Receipts))
	}

	originalCount = len(req.Histories)
	for _, history := range req.Histories {
		if accessibleShops[history.ShopID] {
			filteredReq.Histories = append(filteredReq.Histories, history)
		}
	}
	if originalCount > 0 {
		stats["histories"] = fmt.Sprintf("%d→%d", originalCount, len(filteredReq.Histories))
	}

	// Filter shops based on role access
	if syncContext.UserRole != "cashier" {
		if syncContext.UserRole == "owner_business" {
			originalCount = len(req.Shops)
			log.Printf("DEBUG: Owner business filtering - original shops: %d, license_id: %s", originalCount, syncContext.LicenseID)
			for i, shop := range req.Shops {
				log.Printf("DEBUG: Checking shop %d: ID=%s, LicenseID=%s, Name=%s", i, shop.ID, shop.LicenseID, shop.Name)
				if shop.LicenseID == syncContext.LicenseID {
					filteredReq.Shops = append(filteredReq.Shops, shop)
					log.Printf("DEBUG: Shop %s ACCEPTED for owner_business (license match)", shop.ID)
				} else {
					log.Printf("ERROR: Shop %s REJECTED for owner_business (license mismatch: %s != %s)", shop.ID, shop.LicenseID, syncContext.LicenseID)
					// Add detailed error for license mismatch
					s.addDetailedError(response, "shops", shop.ID, "license_mismatch",
						fmt.Sprintf("Shop license %s does not match user license %s", shop.LicenseID, syncContext.LicenseID),
						map[string]interface{}{
							"shop_license_id": shop.LicenseID,
							"user_license_id": syncContext.LicenseID,
							"shop_name":       shop.Name,
							"user_role":       syncContext.UserRole,
						})
					continue
				}
			}
			log.Printf("DEBUG: Owner business filtering result - filtered shops: %d", len(filteredReq.Shops))
			if originalCount > 0 {
				stats["shops"] = fmt.Sprintf("%d→%d", originalCount, len(filteredReq.Shops))
			}
		} else {
			filteredReq.Shops = req.Shops // Global access users
		}
		filteredReq.Users = req.Users
	}

	// CRITICAL FIX: Filter stock histories and transaction products based on shop access
	// For entities with references, check both sync request data AND database data
	// Use the main database connection for filtering (not transaction) since we're checking existing state
	filteredReq.StockHistories = s.filterStockHistoriesByShopAccessWithSyncData(req.StockHistories, req.Products, accessibleShops, syncContext)
	if len(req.StockHistories) > 0 {
		stats["stock_histories"] = fmt.Sprintf("%d→%d", len(req.StockHistories), len(filteredReq.StockHistories))
	}

	filteredReq.TransactionProducts = s.filterTransactionProductsByShopAccessWithSyncData(req.TransactionProducts, req.Transactions, accessibleShops, syncContext)
	if len(req.TransactionProducts) > 0 {
		stats["transaction_products"] = fmt.Sprintf("%d→%d", len(req.TransactionProducts), len(filteredReq.TransactionProducts))
	}

	// Generate stats summary
	var statsParts []string
	for entityType, stat := range stats {
		statsParts = append(statsParts, fmt.Sprintf("%s %s", entityType, stat))
	}

	if len(statsParts) == 0 {
		return filteredReq, "no entities to filter"
	}

	return filteredReq, strings.Join(statsParts, ", ")
}

// generateFilterWarnings creates warning messages for filtered entities with enhanced detail
func (s *SyncService) generateFilterWarnings(original, filtered dto.SyncRequest, syncContext dto.SyncContext) []dto.SyncError {
	var warnings []dto.SyncError

	// Check for filtered carts
	if len(original.Carts) > len(filtered.Carts) {
		filteredCount := len(original.Carts) - len(filtered.Carts)
		warnings = append(warnings, dto.SyncError{
			EntityType: "carts",
			EntityID:   uuid.Nil,
			ErrorCode:  "access_filtered",
			Message:    fmt.Sprintf("Filtered %d cart(s) from inaccessible shops", filteredCount),
			Details:    fmt.Sprintf("User %s (role: %s) - accessible shops: %v", syncContext.UserID, syncContext.UserRole, syncContext.AccessibleShopIDs),
		})
	}

	// Check for filtered categories
	if len(original.Categories) > len(filtered.Categories) {
		filteredCount := len(original.Categories) - len(filtered.Categories)
		warnings = append(warnings, dto.SyncError{
			EntityType: "categories",
			EntityID:   uuid.Nil,
			ErrorCode:  "access_filtered",
			Message:    fmt.Sprintf("Filtered %d categor(y/ies) from inaccessible shops", filteredCount),
			Details:    fmt.Sprintf("User %s (role: %s) - accessible shops: %v", syncContext.UserID, syncContext.UserRole, syncContext.AccessibleShopIDs),
		})
	}

	// Check for filtered products
	if len(original.Products) > len(filtered.Products) {
		filteredCount := len(original.Products) - len(filtered.Products)
		warnings = append(warnings, dto.SyncError{
			EntityType: "products",
			EntityID:   uuid.Nil,
			ErrorCode:  "access_filtered",
			Message:    fmt.Sprintf("Filtered %d product(s) from inaccessible shops", filteredCount),
			Details:    fmt.Sprintf("User %s (role: %s) - accessible shops: %v", syncContext.UserID, syncContext.UserRole, syncContext.AccessibleShopIDs),
		})
	}

	// Check for filtered stock histories with enhanced detail
	if len(original.StockHistories) > len(filtered.StockHistories) {
		filteredCount := len(original.StockHistories) - len(filtered.StockHistories)

		// Create a map of products in the sync request for quick lookup
		syncProductMap := make(map[uuid.UUID]uuid.UUID) // product_id -> shop_id
		for _, product := range original.Products {
			syncProductMap[product.ID] = product.ShopID
		}

		// Count missing vs inaccessible entities for better error messaging
		var missingCount, inaccessibleCount int
		for _, stockHistory := range original.StockHistories {
			// Check if this stock history was filtered (not in filtered list)
			wasFiltered := true
			for _, filteredSH := range filtered.StockHistories {
				if filteredSH.ID == stockHistory.ID {
					wasFiltered = false
					break
				}
			}

			if !wasFiltered {
				continue // This wasn't filtered, skip it
			}

			var productShopID uuid.UUID
			var productFound bool

			// First check if product is in the sync request data
			if shopID, exists := syncProductMap[stockHistory.ProductID]; exists {
				productShopID = shopID
				productFound = true
			} else {
				// If not in sync data, check the database with enhanced error handling
				var product entities.Product
				err := s.db.Select("shop_id").
					Where("id = ?", stockHistory.ProductID).
					First(&product).Error

				if err != nil {
					if err == gorm.ErrRecordNotFound {
						log.Printf("DEBUG: generateFilterWarnings - Stock history %s references non-existent product %s", stockHistory.ID, stockHistory.ProductID)
					} else {
						log.Printf("ERROR: generateFilterWarnings - Database error checking product %s for stock history %s: %v", stockHistory.ProductID, stockHistory.ID, err)
					}
					missingCount++
					continue
				}
				productShopID = product.ShopID
				productFound = true
				log.Printf("DEBUG: generateFilterWarnings - Stock history %s references product %s in shop %s", stockHistory.ID, stockHistory.ProductID, productShopID)
			}

			if productFound {
				// CRITICAL FIX: Check if shop is accessible using direct shop ID comparison instead of map lookup
				accessible := false
				for _, shopID := range syncContext.AccessibleShopIDs {
					if shopID == productShopID {
						accessible = true
						break
					}
				}
				log.Printf("DEBUG: generateFilterWarnings - Stock history %s: product %s in shop %s, accessible: %v (user accessible shops: %v)",
					stockHistory.ID, stockHistory.ProductID, productShopID, accessible, syncContext.AccessibleShopIDs)
				if !accessible {
					inaccessibleCount++
				}
			}
		}

		var message string
		if missingCount > 0 && inaccessibleCount > 0 {
			message = fmt.Sprintf("Filtered %d stock histor(y/ies): %d reference missing products, %d from inaccessible shops",
				filteredCount, missingCount, inaccessibleCount)
		} else if missingCount > 0 {
			message = fmt.Sprintf("Filtered %d stock histor(y/ies) - reference missing products", filteredCount)
		} else {
			message = fmt.Sprintf("Filtered %d stock histor(y/ies) from inaccessible shops", filteredCount)
		}

		warnings = append(warnings, dto.SyncError{
			EntityType: "stock_histories",
			EntityID:   uuid.Nil,
			ErrorCode:  "access_filtered",
			Message:    message,
			Details: fmt.Sprintf("User %s (role: %s) - accessible shops: %v, missing products: %d, inaccessible shops: %d",
				syncContext.UserID, syncContext.UserRole, syncContext.AccessibleShopIDs, missingCount, inaccessibleCount),
		})
	}

	// Check for filtered transaction products with enhanced detail
	if len(original.TransactionProducts) > len(filtered.TransactionProducts) {
		filteredCount := len(original.TransactionProducts) - len(filtered.TransactionProducts)

		// Create a map of transactions in the sync request for quick lookup
		syncTransactionMap := make(map[uuid.UUID]uuid.UUID) // transaction_id -> shop_id
		for _, transaction := range original.Transactions {
			syncTransactionMap[transaction.ID] = transaction.ShopID
		}

		// Count missing vs inaccessible entities for better error messaging
		var missingCount, inaccessibleCount int
		for _, transactionProduct := range original.TransactionProducts {
			// Check if this transaction product was filtered (not in filtered list)
			wasFiltered := true
			for _, filteredTP := range filtered.TransactionProducts {
				if filteredTP.ID == transactionProduct.ID {
					wasFiltered = false
					break
				}
			}

			if !wasFiltered {
				continue // This wasn't filtered, skip it
			}

			var transactionShopID uuid.UUID
			var transactionFound bool

			// First check if transaction is in the sync request data
			if shopID, exists := syncTransactionMap[transactionProduct.TransactionID]; exists {
				transactionShopID = shopID
				transactionFound = true
			} else {
				// If not in sync data, check the database with enhanced error handling
				var transaction entities.Transaction
				err := s.db.Select("shop_id").
					Where("id = ?", transactionProduct.TransactionID).
					First(&transaction).Error

				if err != nil {
					if err == gorm.ErrRecordNotFound {
						log.Printf("DEBUG: generateFilterWarnings - Transaction product %s references non-existent transaction %s", transactionProduct.ID, transactionProduct.TransactionID)
					} else {
						log.Printf("ERROR: generateFilterWarnings - Database error checking transaction %s for transaction product %s: %v", transactionProduct.TransactionID, transactionProduct.ID, err)
					}
					missingCount++
					continue
				}
				transactionShopID = transaction.ShopID
				transactionFound = true
				log.Printf("DEBUG: generateFilterWarnings - Transaction product %s references transaction %s in shop %s", transactionProduct.ID, transactionProduct.TransactionID, transactionShopID)
			}

			if transactionFound {
				// CRITICAL FIX: Check if shop is accessible using direct shop ID comparison instead of map lookup
				accessible := false
				for _, shopID := range syncContext.AccessibleShopIDs {
					if shopID == transactionShopID {
						accessible = true
						break
					}
				}
				log.Printf("DEBUG: generateFilterWarnings - Transaction product %s: transaction %s in shop %s, accessible: %v (user accessible shops: %v)",
					transactionProduct.ID, transactionProduct.TransactionID, transactionShopID, accessible, syncContext.AccessibleShopIDs)
				if !accessible {
					inaccessibleCount++
				}
			}
		}

		var message string
		if missingCount > 0 && inaccessibleCount > 0 {
			message = fmt.Sprintf("Filtered %d transaction product(s): %d reference missing transactions, %d from inaccessible shops",
				filteredCount, missingCount, inaccessibleCount)
		} else if missingCount > 0 {
			message = fmt.Sprintf("Filtered %d transaction product(s) - reference missing transactions", filteredCount)
		} else {
			message = fmt.Sprintf("Filtered %d transaction product(s) from inaccessible shops", filteredCount)
		}

		warnings = append(warnings, dto.SyncError{
			EntityType: "transaction_products",
			EntityID:   uuid.Nil,
			ErrorCode:  "access_filtered",
			Message:    message,
			Details: fmt.Sprintf("User %s (role: %s) - accessible shops: %v, missing transactions: %d, inaccessible shops: %d",
				syncContext.UserID, syncContext.UserRole, syncContext.AccessibleShopIDs, missingCount, inaccessibleCount),
		})
	}

	return warnings
}

// pullChangesWithRoleAccess retrieves server changes since last sync with role-based filtering
func (s *SyncService) pullChangesWithRoleAccess(ctx context.Context, tx *gorm.DB, lastSync *time.Time, syncContext dto.SyncContext, response *dto.SyncResponse) error {
	// If no last sync timestamp, this is initial sync - return recent data
	if lastSync == nil {
		// For initial sync, return data from last 30 days
		thirtyDaysAgo := time.Now().AddDate(0, 0, -30)
		lastSync = &thirtyDaysAgo
	}

	// Pull changes for each entity type with role-based filtering
	if err := s.pullCartsWithRoleAccess(ctx, tx, *lastSync, syncContext, response); err != nil {
		return fmt.Errorf("failed to pull carts: %w", err)
	}

	if err := s.pullCategoriesWithRoleAccess(ctx, tx, *lastSync, syncContext, response); err != nil {
		return fmt.Errorf("failed to pull categories: %w", err)
	}

	if err := s.pullProductsWithRoleAccess(ctx, tx, *lastSync, syncContext, response); err != nil {
		return fmt.Errorf("failed to pull products: %w", err)
	}

	if err := s.pullTransactionsWithRoleAccess(ctx, tx, *lastSync, syncContext, response); err != nil {
		return fmt.Errorf("failed to pull transactions: %w", err)
	}

	if err := s.pullExpensesWithRoleAccess(ctx, tx, *lastSync, syncContext, response); err != nil {
		return fmt.Errorf("failed to pull expenses: %w", err)
	}

	if err := s.pullPaymentsWithRoleAccess(ctx, tx, *lastSync, syncContext, response); err != nil {
		return fmt.Errorf("failed to pull payments: %w", err)
	}

	if err := s.pullReceiptsWithRoleAccess(ctx, tx, *lastSync, syncContext, response); err != nil {
		return fmt.Errorf("failed to pull receipts: %w", err)
	}

	if err := s.pullHistoriesWithRoleAccess(ctx, tx, *lastSync, syncContext, response); err != nil {
		return fmt.Errorf("failed to pull histories: %w", err)
	}

	// Only sync shop and user data for non-cashier roles
	if syncContext.UserRole != "cashier" {
		if err := s.pullShopsWithRoleAccess(ctx, tx, *lastSync, syncContext, response); err != nil {
			return fmt.Errorf("failed to pull shops: %w", err)
		}

		if err := s.pullUsersWithRoleAccess(ctx, tx, *lastSync, syncContext, response); err != nil {
			return fmt.Errorf("failed to pull users: %w", err)
		}
	}

	if err := s.pullStockHistoriesWithRoleAccess(ctx, tx, *lastSync, syncContext, response); err != nil {
		return fmt.Errorf("failed to pull stock histories: %w", err)
	}

	if err := s.pullTransactionProductsWithRoleAccess(ctx, tx, *lastSync, syncContext, response); err != nil {
		return fmt.Errorf("failed to pull transaction products: %w", err)
	}

	return nil
}

// filterSyncRequestByRole filters the sync request entities based on user role and accessible shop IDs
func (s *SyncService) filterSyncRequestByRole(req dto.SyncRequest, syncContext dto.SyncContext) dto.SyncRequest {
	// If user has global access, return unfiltered request
	if syncContext.HasGlobalAccess {
		return req
	}

	// Create a map for fast shop access checking
	accessibleShops := make(map[uuid.UUID]bool)
	for _, shopID := range syncContext.AccessibleShopIDs {
		accessibleShops[shopID] = true
	}

	filteredReq := dto.SyncRequest{
		LastSyncTimestamp: req.LastSyncTimestamp,
	}

	// Filter entities by accessible shops
	for _, cart := range req.Carts {
		if accessibleShops[cart.ShopID] {
			filteredReq.Carts = append(filteredReq.Carts, cart)
		}
	}

	for _, category := range req.Categories {
		if accessibleShops[category.ShopID] {
			filteredReq.Categories = append(filteredReq.Categories, category)
		}
	}

	for _, product := range req.Products {
		if accessibleShops[product.ShopID] {
			filteredReq.Products = append(filteredReq.Products, product)
		}
	}

	for _, transaction := range req.Transactions {
		if accessibleShops[transaction.ShopID] {
			filteredReq.Transactions = append(filteredReq.Transactions, transaction)
		}
	}

	for _, expense := range req.Expenses {
		if accessibleShops[expense.ShopID] {
			filteredReq.Expenses = append(filteredReq.Expenses, expense)
		}
	}

	for _, payment := range req.Payments {
		if accessibleShops[payment.ShopID] {
			filteredReq.Payments = append(filteredReq.Payments, payment)
		}
	}

	for _, receipt := range req.Receipts {
		if accessibleShops[receipt.ShopID] {
			filteredReq.Receipts = append(filteredReq.Receipts, receipt)
		}
	}

	for _, history := range req.Histories {
		if accessibleShops[history.ShopID] {
			filteredReq.Histories = append(filteredReq.Histories, history)
		}
	}

	// Filter shops based on role access
	if syncContext.UserRole != "cashier" {
		// For owner_business, only include shops from their license
		// For super_admin/admin, include all shops (global access handled above)
		if syncContext.UserRole == "owner_business" {
			for _, shop := range req.Shops {
				if shop.LicenseID == syncContext.LicenseID {
					filteredReq.Shops = append(filteredReq.Shops, shop)
				}
			}
		} else {
			filteredReq.Shops = req.Shops // Global access users (already handled above)
		}
		filteredReq.Users = req.Users
	}

	// CRITICAL FIX: Filter stock histories based on product shop access
	filteredReq.StockHistories = s.filterStockHistoriesByShopAccess(req.StockHistories, accessibleShops, syncContext)

	// CRITICAL FIX: Filter transaction products based on transaction/product shop access
	filteredReq.TransactionProducts = s.filterTransactionProductsByShopAccess(req.TransactionProducts, accessibleShops, syncContext)

	return filteredReq
}

// pushCartsSafe handles cart synchronization with enhanced error handling and no nested transactions
func (s *SyncService) pushCartsSafe(ctx context.Context, tx *gorm.DB, carts []entities.Cart, licenseID uuid.UUID, response *dto.SyncResponse) error {
	totalCarts := len(carts)
	if totalCarts == 0 {
		return nil
	}

	log.Printf("Processing %d carts in batches of %d", totalCarts, s.config.BatchSize)

	// CRITICAL FIX: Enhanced error handling to prevent transaction abort
	successCount := 0
	errorCount := 0

	for i := 0; i < totalCarts; i += s.config.BatchSize {
		end := i + s.config.BatchSize
		if end > totalCarts {
			end = totalCarts
		}

		batch := carts[i:end]
		log.Printf("Processing cart batch %d-%d of %d", i+1, end, totalCarts)

		// Process each cart in the batch with individual error handling
		for batchIndex, cart := range batch {
			if err := s.processSingleCartWithErrorIsolation(ctx, tx, cart, licenseID, response); err != nil {
				log.Printf("Error processing cart %s (batch %d, index %d): %v", cart.ID, i/s.config.BatchSize+1, batchIndex, err)
				errorCount++

				// CRITICAL FIX: Use policy-based error handling instead of hardcoded continue
				if handleErr := s.handleEntityError(err, "carts", cart.ID, response,
					map[string]interface{}{
						"cart_shop_id": cart.ShopID,
						"license_id":   licenseID,
						"batch_index":  i + batchIndex,
						"batch_number": i/s.config.BatchSize + 1,
					}); handleErr != nil {
					// Error policy requires aborting
					return fmt.Errorf("cart processing failed with abort policy: %w", handleErr)
				}
				continue
			}
			successCount++
		}

		// Check context for cancellation between batches
		select {
		case <-ctx.Done():
			return fmt.Errorf("sync operation cancelled: %w", ctx.Err())
		default:
		}
	}

	log.Printf("Cart processing completed: %d successful, %d errors out of %d total",
		successCount, errorCount, totalCarts)

	// Return success even if some individual carts failed
	// The errors are recorded in the response for client handling
	return nil
}

// processSingleCartWithErrorIsolation processes a single cart entity with isolated error handling
func (s *SyncService) processSingleCartWithErrorIsolation(ctx context.Context, tx *gorm.DB, cart entities.Cart, licenseID uuid.UUID, response *dto.SyncResponse) error {
	// Validate cart belongs to license
	if !s.validateCartLicense(ctx, cart, licenseID) {
		return fmt.Errorf("cart does not belong to license %s", licenseID)
	}

	// CRITICAL FIX: Use savepoint-based processing for better error isolation
	return s.processEntityWithSavepoint(ctx, tx, cart.ID, func() error {
		// Check if cart exists with error handling
		existingCart, err := s.findCartByIDSafe(ctx, tx, cart.ID)
		if err != nil {
			return fmt.Errorf("failed to find cart: %w", err)
		}

		if existingCart == nil {
			// Create new cart
			if err := s.createCartSafe(ctx, tx, cart); err != nil {
				return fmt.Errorf("failed to create cart: %w", err)
			}
			s.incrementStat(response.Stats.CreatedEntities, "carts")
		} else {
			// Handle potential conflict
			if conflict := s.resolveCartConflict(*existingCart, cart); conflict != nil {
				response.Conflicts = append(response.Conflicts, *conflict)
				// Use server version in case of conflict (for LastWriteWins strategy)
				if existingCart.UpdatedAt.After(cart.UpdatedAt) {
					s.incrementStat(response.Stats.ProcessedEntities, "carts")
					return nil // Skip update, server version is newer
				}
			}

			// Update existing cart
			if err := s.updateCartSafe(ctx, tx, cart); err != nil {
				return fmt.Errorf("failed to update cart: %w", err)
			}
			s.incrementStat(response.Stats.UpdatedEntities, "carts")
		}

		s.incrementStat(response.Stats.ProcessedEntities, "carts")
		return nil
	})
}

// Safe database operations with enhanced error handling
func (s *SyncService) findCartByIDSafe(ctx context.Context, tx *gorm.DB, id uuid.UUID) (*entities.Cart, error) {
	var cart entities.Cart

	// Use a short timeout for individual operations
	ctxWithTimeout, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	err := tx.WithContext(ctxWithTimeout).Where("id = ?", id).First(&cart).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		// Log the specific database error for debugging
		log.Printf("Database error finding cart %s: %v", id, err)
		return nil, err
	}
	return &cart, nil
}

func (s *SyncService) createCartSafe(ctx context.Context, tx *gorm.DB, cart entities.Cart) error {
	ctxWithTimeout, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	err := tx.WithContext(ctxWithTimeout).Create(&cart).Error
	if err != nil {
		log.Printf("Database error creating cart %s: %v", cart.ID, err)
	}
	return err
}

func (s *SyncService) updateCartSafe(ctx context.Context, tx *gorm.DB, cart entities.Cart) error {
	ctxWithTimeout, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	err := tx.WithContext(ctxWithTimeout).Save(&cart).Error
	if err != nil {
		log.Printf("Database error updating cart %s: %v", cart.ID, err)
	}
	return err
}

// processSingleCart processes a single cart entity with enhanced error handling and comprehensive validation
func (s *SyncService) processSingleCart(ctx context.Context, tx *gorm.DB, cart entities.Cart, licenseID uuid.UUID, response *dto.SyncResponse) error {
	// SECURITY ENHANCEMENT: Comprehensive entity validation
	var existingCart *entities.Cart
	var err error

	// First, try to find existing cart for update validation
	existingCart, _ = s.findCartByID(ctx, tx, cart.ID)

	// Determine operation type
	operation := "create"
	if existingCart != nil {
		operation = "update"
	}

	// Validate entity with comprehensive security checks
	if err := s.validator.ValidateEntity(ctx, cart, operation, existingCart); err != nil {
		s.addDetailedValidationError(response, "carts", cart.ID, "validation_failed", err,
			map[string]interface{}{"operation": operation, "validation_type": "comprehensive"})
		return nil // Continue processing other entities based on error policy
	}

	// Validate cart belongs to license (existing validation)
	if !s.validateCartLicense(ctx, cart, licenseID) {
		s.addDetailedError(response, "carts", cart.ID, "unauthorized", "Cart does not belong to license",
			map[string]interface{}{"cart_shop_id": cart.ShopID, "license_id": licenseID})
		return nil // Continue processing other entities
	}

	// Re-find cart with retry for actual processing
	operation_find := func() error {
		existingCart, err = s.findCartByID(ctx, tx, cart.ID)
		return err
	}

	if retryErr := s.retryOperation(ctx, operation_find, s.config.MaxRetries, s.config.BaseRetryDelay, fmt.Sprintf("find_cart_%s", cart.ID)); retryErr != nil {
		s.addDetailedError(response, "carts", cart.ID, "database_error", retryErr.Error(),
			map[string]interface{}{"operation": "find", "retry_attempts": s.config.MaxRetries})
		return nil // Continue processing other entities
	}

	if existingCart == nil {
		// Create new cart with retry
		createOperation := func() error {
			return s.createCart(ctx, tx, cart)
		}

		if err := s.retryOperation(ctx, createOperation, s.config.MaxRetries, s.config.BaseRetryDelay, fmt.Sprintf("create_cart_%s", cart.ID)); err != nil {
			s.addDetailedError(response, "carts", cart.ID, "create_failed", err.Error(),
				map[string]interface{}{"operation": "create", "retry_attempts": s.config.MaxRetries})
			return nil // Continue processing other entities
		}
		s.incrementStat(response.Stats.CreatedEntities, "carts")
	} else {
		// Handle potential conflict
		if conflict := s.resolveCartConflict(*existingCart, cart); conflict != nil {
			response.Conflicts = append(response.Conflicts, *conflict)
			// Use server version in case of conflict (for LastWriteWins strategy)
			if existingCart.UpdatedAt.After(cart.UpdatedAt) {
				s.incrementStat(response.Stats.ProcessedEntities, "carts")
				return nil // Skip update, server version is newer
			}
		}

		// Update existing cart with retry
		updateOperation := func() error {
			return s.updateCart(ctx, tx, cart)
		}

		if err := s.retryOperation(ctx, updateOperation, s.config.MaxRetries, s.config.BaseRetryDelay, fmt.Sprintf("update_cart_%s", cart.ID)); err != nil {
			s.addDetailedError(response, "carts", cart.ID, "update_failed", err.Error(),
				map[string]interface{}{"operation": "update", "retry_attempts": s.config.MaxRetries})
			return nil // Continue processing other entities
		}
		s.incrementStat(response.Stats.UpdatedEntities, "carts")
	}

	s.incrementStat(response.Stats.ProcessedEntities, "carts")
	return nil
}

// pullCarts retrieves server-side cart changes with optimized queries
func (s *SyncService) pullCarts(ctx context.Context, tx *gorm.DB, lastSync time.Time, licenseID uuid.UUID, response *dto.SyncResponse) error {
	// Use optimized query with new composite indexes
	// The idx_carts_shop_updated index will be used for efficient filtering
	var carts []entities.Cart

	query := tx.WithContext(ctx).
		Select("carts.*").
		Table("carts").
		Joins("INNER JOIN shops ON carts.shop_id = shops.id").
		Where("shops.license_id = ? AND carts.updated_at > ?", licenseID, lastSync).
		Order("carts.updated_at ASC") // Order by updated_at for consistent pagination

	// Add pagination to prevent memory issues with large datasets
	maxResults := s.config.MaxResultsPerQuery
	err := query.Limit(maxResults).Find(&carts).Error

	if err != nil {
		return fmt.Errorf("failed to query carts: %w", err)
	}

	// Log query performance for monitoring
	log.Printf("Retrieved %d carts for license %s since %v", len(carts), licenseID, lastSync)

	// If we hit the limit, log a warning about potential incomplete sync
	if len(carts) == maxResults {
		log.Printf("WARNING: Cart sync hit result limit (%d), some data may be missing. Consider using smaller sync intervals.", maxResults)
		s.addError(response, "carts", uuid.Nil, "result_limit_reached",
			fmt.Sprintf("Retrieved maximum %d carts. Some data may be missing due to result size limits.", maxResults))
	}

	// Map entities to DTOs to exclude relationship fields
	response.Carts = dto.MapCartsToSyncDTOs(carts)
	return nil
}

// pullCartsWithRoleAccess retrieves server-side cart changes with role-based filtering
func (s *SyncService) pullCartsWithRoleAccess(ctx context.Context, tx *gorm.DB, lastSync time.Time, syncContext dto.SyncContext, response *dto.SyncResponse) error {
	var carts []entities.Cart
	query := tx.WithContext(ctx).
		Select("carts.*").
		Table("carts").
		Joins("INNER JOIN shops ON carts.shop_id = shops.id").
		Where("shops.license_id = ? AND carts.updated_at > ?", syncContext.LicenseID, lastSync)

	// Apply shop filtering for non-global access users
	if !syncContext.HasGlobalAccess && len(syncContext.AccessibleShopIDs) > 0 {
		query = query.Where("carts.shop_id IN ?", syncContext.AccessibleShopIDs)
	}

	err := query.Order("carts.updated_at ASC").
		Limit(s.config.MaxResultsPerQuery).
		Find(&carts).Error

	if err != nil {
		return fmt.Errorf("failed to query carts: %w", err)
	}

	log.Printf("Retrieved %d carts for user %s (role: %s) since %v", len(carts), syncContext.UserID, syncContext.UserRole, lastSync)
	response.Carts = dto.MapCartsToSyncDTOs(carts)
	return nil
}

// pullCategoriesWithRoleAccess retrieves server-side category changes with role-based filtering
func (s *SyncService) pullCategoriesWithRoleAccess(ctx context.Context, tx *gorm.DB, lastSync time.Time, syncContext dto.SyncContext, response *dto.SyncResponse) error {
	var categories []entities.Category
	query := tx.WithContext(ctx).
		Select("categories.*").
		Table("categories").
		Joins("JOIN shops ON categories.shop_id = shops.id").
		Where("shops.license_id = ? AND categories.updated_at > ?", syncContext.LicenseID, lastSync)

	// Apply shop filtering for non-global access users
	if !syncContext.HasGlobalAccess && len(syncContext.AccessibleShopIDs) > 0 {
		query = query.Where("categories.shop_id IN ?", syncContext.AccessibleShopIDs)
	}

	err := query.Find(&categories).Error
	if err != nil {
		return fmt.Errorf("failed to query categories: %w", err)
	}

	log.Printf("Retrieved %d categories for user %s (role: %s) since %v", len(categories), syncContext.UserID, syncContext.UserRole, lastSync)
	response.Categories = dto.MapCategoriesToSyncDTOs(categories)
	return nil
}

// pullProductsWithRoleAccess retrieves server-side product changes with role-based filtering
func (s *SyncService) pullProductsWithRoleAccess(ctx context.Context, tx *gorm.DB, lastSync time.Time, syncContext dto.SyncContext, response *dto.SyncResponse) error {
	var products []entities.Product
	query := tx.WithContext(ctx).
		Select("products.*").
		Table("products").
		Joins("JOIN shops ON products.shop_id = shops.id").
		Where("shops.license_id = ? AND products.updated_at > ?", syncContext.LicenseID, lastSync)

	// Apply shop filtering for non-global access users
	if !syncContext.HasGlobalAccess && len(syncContext.AccessibleShopIDs) > 0 {
		query = query.Where("products.shop_id IN ?", syncContext.AccessibleShopIDs)
	}

	err := query.Find(&products).Error
	if err != nil {
		return fmt.Errorf("failed to query products: %w", err)
	}

	log.Printf("Retrieved %d products for user %s (role: %s) since %v", len(products), syncContext.UserID, syncContext.UserRole, lastSync)
	response.Products = dto.MapProductsToSyncDTOs(products)
	return nil
}

// pullTransactionsWithRoleAccess retrieves server-side transaction changes with role-based filtering
func (s *SyncService) pullTransactionsWithRoleAccess(ctx context.Context, tx *gorm.DB, lastSync time.Time, syncContext dto.SyncContext, response *dto.SyncResponse) error {
	var transactions []entities.Transaction
	query := tx.WithContext(ctx).
		Select("transactions.*").
		Table("transactions").
		Joins("JOIN shops ON transactions.shop_id = shops.id").
		Where("shops.license_id = ? AND transactions.updated_at > ?", syncContext.LicenseID, lastSync)

	// Apply shop filtering for non-global access users
	if !syncContext.HasGlobalAccess && len(syncContext.AccessibleShopIDs) > 0 {
		query = query.Where("transactions.shop_id IN ?", syncContext.AccessibleShopIDs)
	}

	err := query.Find(&transactions).Error
	if err != nil {
		return fmt.Errorf("failed to query transactions: %w", err)
	}

	log.Printf("Retrieved %d transactions for user %s (role: %s) since %v", len(transactions), syncContext.UserID, syncContext.UserRole, lastSync)
	response.Transactions = dto.MapTransactionsToSyncDTOs(transactions)
	return nil
}

// pullExpensesWithRoleAccess retrieves server-side expense changes with role-based filtering
func (s *SyncService) pullExpensesWithRoleAccess(ctx context.Context, tx *gorm.DB, lastSync time.Time, syncContext dto.SyncContext, response *dto.SyncResponse) error {
	var expenses []entities.Expense
	query := tx.WithContext(ctx).
		Select("expenses.*").
		Table("expenses").
		Joins("JOIN shops ON expenses.shop_id = shops.id").
		Where("shops.license_id = ? AND expenses.updated_at > ?", syncContext.LicenseID, lastSync)

	// Apply shop filtering for non-global access users
	if !syncContext.HasGlobalAccess && len(syncContext.AccessibleShopIDs) > 0 {
		query = query.Where("expenses.shop_id IN ?", syncContext.AccessibleShopIDs)
	}

	err := query.Find(&expenses).Error
	if err != nil {
		return fmt.Errorf("failed to query expenses: %w", err)
	}

	log.Printf("Retrieved %d expenses for user %s (role: %s) since %v", len(expenses), syncContext.UserID, syncContext.UserRole, lastSync)
	response.Expenses = dto.MapExpensesToSyncDTOs(expenses)
	return nil
}

// pullPaymentsWithRoleAccess retrieves server-side payment changes with role-based filtering
func (s *SyncService) pullPaymentsWithRoleAccess(ctx context.Context, tx *gorm.DB, lastSync time.Time, syncContext dto.SyncContext, response *dto.SyncResponse) error {
	var payments []entities.Payment
	query := tx.WithContext(ctx).
		Select("payments.*").
		Table("payments").
		Joins("JOIN shops ON payments.shop_id = shops.id").
		Where("shops.license_id = ? AND payments.updated_at > ?", syncContext.LicenseID, lastSync)

	// Apply shop filtering for non-global access users
	if !syncContext.HasGlobalAccess && len(syncContext.AccessibleShopIDs) > 0 {
		query = query.Where("payments.shop_id IN ?", syncContext.AccessibleShopIDs)
	}

	err := query.Find(&payments).Error
	if err != nil {
		return fmt.Errorf("failed to query payments: %w", err)
	}

	log.Printf("Retrieved %d payments for user %s (role: %s) since %v", len(payments), syncContext.UserID, syncContext.UserRole, lastSync)
	response.Payments = dto.MapPaymentsToSyncDTOs(payments)
	return nil
}

// pullReceiptsWithRoleAccess retrieves server-side receipt changes with role-based filtering
func (s *SyncService) pullReceiptsWithRoleAccess(ctx context.Context, tx *gorm.DB, lastSync time.Time, syncContext dto.SyncContext, response *dto.SyncResponse) error {
	var receipts []entities.Receipt
	query := tx.WithContext(ctx).
		Select("receipts.*").
		Table("receipts").
		Joins("JOIN shops ON receipts.shop_id = shops.id").
		Where("shops.license_id = ? AND receipts.updated_at > ?", syncContext.LicenseID, lastSync)

	// Apply shop filtering for non-global access users
	if !syncContext.HasGlobalAccess && len(syncContext.AccessibleShopIDs) > 0 {
		query = query.Where("receipts.shop_id IN ?", syncContext.AccessibleShopIDs)
	}

	err := query.Find(&receipts).Error
	if err != nil {
		return fmt.Errorf("failed to query receipts: %w", err)
	}

	log.Printf("Retrieved %d receipts for user %s (role: %s) since %v", len(receipts), syncContext.UserID, syncContext.UserRole, lastSync)
	response.Receipts = dto.MapReceiptsToSyncDTOs(receipts)
	return nil
}

// pullHistoriesWithRoleAccess retrieves server-side history changes with role-based filtering
func (s *SyncService) pullHistoriesWithRoleAccess(ctx context.Context, tx *gorm.DB, lastSync time.Time, syncContext dto.SyncContext, response *dto.SyncResponse) error {
	var histories []entities.History
	query := tx.WithContext(ctx).
		Select("histories.*").
		Table("histories").
		Joins("JOIN shops ON histories.shop_id = shops.id").
		Where("shops.license_id = ? AND histories.updated_at > ?", syncContext.LicenseID, lastSync)

	// Apply shop filtering for non-global access users
	if !syncContext.HasGlobalAccess && len(syncContext.AccessibleShopIDs) > 0 {
		query = query.Where("histories.shop_id IN ?", syncContext.AccessibleShopIDs)
	}

	err := query.Find(&histories).Error
	if err != nil {
		return fmt.Errorf("failed to query histories: %w", err)
	}

	log.Printf("Retrieved %d histories for user %s (role: %s) since %v", len(histories), syncContext.UserID, syncContext.UserRole, lastSync)
	response.Histories = dto.MapHistoriesToSyncDTOs(histories)
	return nil
}

// pullShopsWithRoleAccess retrieves server-side shop changes with role-based filtering
func (s *SyncService) pullShopsWithRoleAccess(ctx context.Context, tx *gorm.DB, lastSync time.Time, syncContext dto.SyncContext, response *dto.SyncResponse) error {
	var shops []entities.Shop
	query := tx.WithContext(ctx).
		Select("shops.*").
		Where("license_id = ? AND updated_at > ?", syncContext.LicenseID, lastSync)

	// Apply shop filtering for non-global access users
	if !syncContext.HasGlobalAccess && len(syncContext.AccessibleShopIDs) > 0 {
		query = query.Where("id IN ?", syncContext.AccessibleShopIDs)
	}

	err := query.Find(&shops).Error
	if err != nil {
		return fmt.Errorf("failed to query shops: %w", err)
	}

	log.Printf("Retrieved %d shops for user %s (role: %s) since %v", len(shops), syncContext.UserID, syncContext.UserRole, lastSync)
	response.Shops = dto.MapShopsToSyncDTOs(shops)
	return nil
}

// pullUsersWithRoleAccess retrieves server-side user changes with role-based filtering
func (s *SyncService) pullUsersWithRoleAccess(ctx context.Context, tx *gorm.DB, lastSync time.Time, syncContext dto.SyncContext, response *dto.SyncResponse) error {
	var users []entities.User
	query := tx.WithContext(ctx).
		Select("users.*").
		Where("license_id = ? AND updated_at > ?", syncContext.LicenseID, lastSync)

	// For owner_business, only return users from their license
	// For super_admin/admin, return all users
	if !syncContext.HasGlobalAccess {
		// Additional filtering could be applied here if needed
	}

	err := query.Find(&users).Error
	if err != nil {
		return fmt.Errorf("failed to query users: %w", err)
	}

	log.Printf("Retrieved %d users for user %s (role: %s) since %v", len(users), syncContext.UserID, syncContext.UserRole, lastSync)
	response.Users = dto.MapUsersToSyncDTOs(users)
	return nil
}

// pullStockHistoriesWithRoleAccess retrieves server-side stock history changes with role-based filtering
func (s *SyncService) pullStockHistoriesWithRoleAccess(ctx context.Context, tx *gorm.DB, lastSync time.Time, syncContext dto.SyncContext, response *dto.SyncResponse) error {
	var stockHistories []entities.StockHistory
	query := tx.WithContext(ctx).
		Select("stock_histories.*").
		Table("stock_histories").
		Joins("JOIN products ON stock_histories.product_id = products.id").
		Joins("JOIN shops ON products.shop_id = shops.id").
		Where("shops.license_id = ? AND stock_histories.updated_at > ?", syncContext.LicenseID, lastSync)

	// Apply shop filtering for non-global access users
	if !syncContext.HasGlobalAccess && len(syncContext.AccessibleShopIDs) > 0 {
		query = query.Where("products.shop_id IN ?", syncContext.AccessibleShopIDs)
	}

	err := query.Find(&stockHistories).Error
	if err != nil {
		return fmt.Errorf("failed to query stock histories: %w", err)
	}

	log.Printf("Retrieved %d stock histories for user %s (role: %s) since %v", len(stockHistories), syncContext.UserID, syncContext.UserRole, lastSync)
	response.StockHistories = dto.MapStockHistoriesToSyncDTOs(stockHistories)
	return nil
}

// pullTransactionProductsWithRoleAccess retrieves server-side transaction product changes with role-based filtering
func (s *SyncService) pullTransactionProductsWithRoleAccess(ctx context.Context, tx *gorm.DB, lastSync time.Time, syncContext dto.SyncContext, response *dto.SyncResponse) error {
	var transactionProducts []entities.TransactionProduct
	query := tx.WithContext(ctx).
		Select("transaction_products.*").
		Table("transaction_products").
		Joins("JOIN transactions ON transaction_products.transaction_id = transactions.id").
		Joins("JOIN shops ON transactions.shop_id = shops.id").
		Where("shops.license_id = ? AND transaction_products.updated_at > ?", syncContext.LicenseID, lastSync)

	// Apply shop filtering for non-global access users
	if !syncContext.HasGlobalAccess && len(syncContext.AccessibleShopIDs) > 0 {
		query = query.Where("transactions.shop_id IN ?", syncContext.AccessibleShopIDs)
	}

	err := query.Find(&transactionProducts).Error
	if err != nil {
		return fmt.Errorf("failed to query transaction products: %w", err)
	}

	log.Printf("Retrieved %d transaction products for user %s (role: %s) since %v", len(transactionProducts), syncContext.UserID, syncContext.UserRole, lastSync)
	response.TransactionProducts = dto.MapTransactionProductsToSyncDTOs(transactionProducts)
	return nil
}

func (s *SyncService) findCartByID(ctx context.Context, tx *gorm.DB, id uuid.UUID) (*entities.Cart, error) {
	var cart entities.Cart
	err := tx.WithContext(ctx).Where("id = ?", id).First(&cart).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &cart, nil
}

func (s *SyncService) createCart(ctx context.Context, tx *gorm.DB, cart entities.Cart) error {
	return tx.WithContext(ctx).Create(&cart).Error
}

func (s *SyncService) updateCart(ctx context.Context, tx *gorm.DB, cart entities.Cart) error {
	return tx.WithContext(ctx).Save(&cart).Error
}

func (s *SyncService) resolveCartConflict(existing, incoming entities.Cart) *dto.ConflictInfo {
	if existing.UpdatedAt.Equal(incoming.UpdatedAt) {
		return nil // No conflict
	}

	conflict := &dto.ConflictInfo{
		EntityType:   "cart",
		EntityID:     existing.ID,
		ConflictType: "timestamp_mismatch",
		ServerData:   existing,
		ClientData:   incoming,
	}

	// Apply conflict resolution strategy
	switch s.conflictStrategy {
	case dto.LastWriteWins:
		if existing.UpdatedAt.After(incoming.UpdatedAt) {
			conflict.Resolution = "server_wins"
			conflict.Details = "Server version is newer"
		} else {
			conflict.Resolution = "client_wins"
			conflict.Details = "Client version is newer"
		}
	case dto.ServerWins:
		conflict.Resolution = "server_wins"
		conflict.Details = "Server version always wins"
	case dto.ClientWins:
		conflict.Resolution = "client_wins"
		conflict.Details = "Client version always wins"
	}

	return conflict
}

// pushChanges processes incoming changes from mobile client (legacy method for backward compatibility)
func (s *SyncService) pushChanges(ctx context.Context, tx *gorm.DB, req dto.SyncRequest, licenseID uuid.UUID, response *dto.SyncResponse) error {
	// Create a default sync context for backward compatibility
	syncContext := dto.SyncContext{
		UserID:            uuid.Nil, // Unknown user for legacy
		UserRole:          "legacy",
		LicenseID:         licenseID,
		AccessibleShopIDs: nil,  // No filtering for legacy sync
		HasGlobalAccess:   true, // Legacy behavior - global access
	}

	return s.pushChangesWithRoleAccessSafe(ctx, tx, req, syncContext, response)
}

// pullChanges retrieves server changes since last sync (legacy method for backward compatibility)
func (s *SyncService) pullChanges(ctx context.Context, tx *gorm.DB, lastSync *time.Time, licenseID uuid.UUID, response *dto.SyncResponse) error {
	// Create a default sync context for backward compatibility
	syncContext := dto.SyncContext{
		UserID:            uuid.Nil, // Unknown user for legacy
		UserRole:          "legacy",
		LicenseID:         licenseID,
		AccessibleShopIDs: nil,  // No filtering for legacy sync
		HasGlobalAccess:   true, // Legacy behavior - global access
	}

	return s.pullChangesWithRoleAccess(ctx, tx, lastSync, syncContext, response)
}

// Helper methods for cart operations
func (s *SyncService) validateCartLicense(ctx context.Context, cart entities.Cart, licenseID uuid.UUID) bool {
	// Validate that the cart's shop belongs to the license
	var count int64
	s.db.Model(&entities.Shop{}).Where("id = ? AND license_id = ?", cart.ShopID, licenseID).Count(&count)
	return count > 0
}

// Helper methods for category operations
func (s *SyncService) validateCategoryLicense(ctx context.Context, tx *gorm.DB, category entities.Category, licenseID uuid.UUID) bool {
	// Validate that the category's shop belongs to the license
	// Use the transaction context to see shops created in the same transaction
	var count int64
	tx.WithContext(ctx).Model(&entities.Shop{}).Where("id = ? AND license_id = ?", category.ShopID, licenseID).Count(&count)
	return count > 0
}

func (s *SyncService) findCategoryByID(ctx context.Context, tx *gorm.DB, id uuid.UUID) (*entities.Category, error) {
	var category entities.Category
	err := tx.WithContext(ctx).Select("id", "shop_id", "name", "created_at", "updated_at").Where("id = ?", id).First(&category).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &category, nil
}

func (s *SyncService) createCategory(ctx context.Context, tx *gorm.DB, category entities.Category) error {
	return tx.WithContext(ctx).Create(&category).Error
}

func (s *SyncService) updateCategory(ctx context.Context, tx *gorm.DB, category entities.Category) error {
	return tx.WithContext(ctx).Save(&category).Error
}

func (s *SyncService) resolveCategoryConflict(existing, incoming entities.Category) *dto.ConflictInfo {
	if existing.UpdatedAt.Equal(incoming.UpdatedAt) {
		return nil // No conflict
	}

	conflict := &dto.ConflictInfo{
		EntityType:   "category",
		EntityID:     existing.ID,
		ConflictType: "timestamp_mismatch",
		ServerData:   existing,
		ClientData:   incoming,
	}

	// Apply conflict resolution strategy
	switch s.conflictStrategy {
	case dto.LastWriteWins:
		if existing.UpdatedAt.After(incoming.UpdatedAt) {
			conflict.Resolution = "server_wins"
			conflict.Details = "Server version is newer"
		} else {
			conflict.Resolution = "client_wins"
			conflict.Details = "Client version is newer"
		}
	case dto.ServerWins:
		conflict.Resolution = "server_wins"
		conflict.Details = "Server version always wins"
	case dto.ClientWins:
		conflict.Resolution = "client_wins"
		conflict.Details = "Client version always wins"
	}

	return conflict
}

// Helper methods for other entities (products, transactions)
// Implementation follows the same pattern as carts and categories

func (s *SyncService) pushCategories(ctx context.Context, tx *gorm.DB, categories []entities.Category, licenseID uuid.UUID, response *dto.SyncResponse) error {
	for _, category := range categories {
		// Validate category belongs to license
		if !s.validateCategoryLicense(ctx, tx, category, licenseID) {
			s.addError(response, "categories", category.ID, "unauthorized", "Category does not belong to license")
			continue
		}

		// Check if category exists
		existingCategory, err := s.findCategoryByID(ctx, tx, category.ID)
		if err != nil && err != gorm.ErrRecordNotFound {
			s.addError(response, "categories", category.ID, "database_error", err.Error())
			continue
		}

		if existingCategory == nil {
			// Create new category
			if err := s.createCategory(ctx, tx, category); err != nil {
				s.addError(response, "categories", category.ID, "create_failed", err.Error())
				continue
			}
			s.incrementStat(response.Stats.CreatedEntities, "categories")
		} else {
			// Handle potential conflict
			if conflict := s.resolveCategoryConflict(*existingCategory, category); conflict != nil {
				response.Conflicts = append(response.Conflicts, *conflict)
				// Use server version in case of conflict (for LastWriteWins strategy)
				if existingCategory.UpdatedAt.After(category.UpdatedAt) {
					continue // Skip update, server version is newer
				}
			}

			// Update existing category
			if err := s.updateCategory(ctx, tx, category); err != nil {
				s.addError(response, "categories", category.ID, "update_failed", err.Error())
				continue
			}
			s.incrementStat(response.Stats.UpdatedEntities, "categories")
		}

		s.incrementStat(response.Stats.ProcessedEntities, "categories")
	}

	return nil
}

func (s *SyncService) pullCategories(ctx context.Context, tx *gorm.DB, lastSync time.Time, licenseID uuid.UUID, response *dto.SyncResponse) error {
	// Get all categories for the license that were updated after lastSync
	// Select only the category fields without relations to avoid unnecessary data
	var categories []entities.Category
	err := tx.WithContext(ctx).
		Select("categories.*").
		Table("categories").
		Joins("JOIN shops ON categories.shop_id = shops.id").
		Where("shops.license_id = ? AND categories.updated_at > ?", licenseID, lastSync).
		Find(&categories).Error

	if err != nil {
		return fmt.Errorf("failed to query categories: %w", err)
	}

	// Map entities to DTOs to exclude relationship fields
	response.Categories = dto.MapCategoriesToSyncDTOs(categories)
	return nil
}

func (s *SyncService) pushProducts(ctx context.Context, tx *gorm.DB, products []entities.Product, licenseID uuid.UUID, response *dto.SyncResponse) error {
	// Performance optimization: Use bulk validation instead of individual queries
	if s.config.EnableBulkValidation && len(products) > 1 {
		return s.pushProductsOptimized(ctx, tx, products, licenseID, response)
	}

	// Fallback to original implementation for single products or when optimization is disabled
	for _, product := range products {
		// Validate product belongs to license
		if !s.validateProductLicense(ctx, product, licenseID) {
			s.addError(response, "products", product.ID, "unauthorized", "Product does not belong to license")
			continue
		}

		// Check if product exists
		existingProduct, err := s.findProductByID(ctx, tx, product.ID)
		if err != nil && err != gorm.ErrRecordNotFound {
			s.addError(response, "products", product.ID, "database_error", err.Error())
			continue
		}

		if existingProduct == nil {
			// Create new product
			if err := s.createProduct(ctx, tx, product); err != nil {
				s.addError(response, "products", product.ID, "create_failed", err.Error())
				continue
			}
			s.incrementStat(response.Stats.CreatedEntities, "products")
		} else {
			// Handle potential conflict
			if conflict := s.resolveProductConflict(*existingProduct, product); conflict != nil {
				response.Conflicts = append(response.Conflicts, *conflict)
				// Use server version in case of conflict (for LastWriteWins strategy)
				if existingProduct.UpdatedAt.After(product.UpdatedAt) {
					continue // Skip update, server version is newer
				}
			}

			// Update existing product
			if err := s.updateProduct(ctx, tx, product); err != nil {
				s.addError(response, "products", product.ID, "update_failed", err.Error())
				continue
			}
			s.incrementStat(response.Stats.UpdatedEntities, "products")
		}

		s.incrementStat(response.Stats.ProcessedEntities, "products")
	}

	return nil
}

// pushProductsOptimized uses bulk validation and optimized processing for better performance
func (s *SyncService) pushProductsOptimized(ctx context.Context, tx *gorm.DB, products []entities.Product, licenseID uuid.UUID, response *dto.SyncResponse) error {
	startTime := time.Now()

	// Step 1: Bulk validate all products belong to the license
	productValidation, err := s.optimizer.BulkValidateProductLicenses(ctx, products, licenseID)
	if err != nil {
		log.Printf("Bulk validation failed, falling back to individual validation: %v", err)
		return s.pushProducts(ctx, tx, products, licenseID, response) // Fallback
	}

	// Step 2: Bulk check which products already exist
	productIDs := make([]uuid.UUID, len(products))
	for i, product := range products {
		productIDs[i] = product.ID
	}

	existingProducts, err := s.optimizer.BulkFindExistingEntities(ctx, tx, "products", productIDs)
	if err != nil {
		log.Printf("Bulk existence check failed, falling back to individual processing: %v", err)
		return s.pushProducts(ctx, tx, products, licenseID, response) // Fallback
	}

	// Step 3: Process products in optimized batches
	return s.optimizer.BatchProcessEntities(ctx, len(products), s.config.OptimalBatchSize, func(startIdx, endIdx int) error {
		batch := products[startIdx:endIdx]

		for _, product := range batch {
			// Check bulk validation result
			if !productValidation[product.ID] {
				s.addError(response, "products", product.ID, "unauthorized", "Product does not belong to license")
				continue
			}

			productExists := existingProducts[product.ID]

			if !productExists {
				// Create new product
				if err := s.createProduct(ctx, tx, product); err != nil {
					s.addError(response, "products", product.ID, "create_failed", err.Error())
					continue
				}
				s.incrementStat(response.Stats.CreatedEntities, "products")
			} else {
				// Get existing product for conflict resolution
				existingProduct, err := s.findProductByID(ctx, tx, product.ID)
				if err != nil {
					s.addError(response, "products", product.ID, "database_error", err.Error())
					continue
				}

				// Handle potential conflict
				if conflict := s.resolveProductConflict(*existingProduct, product); conflict != nil {
					response.Conflicts = append(response.Conflicts, *conflict)
					// Use server version in case of conflict (for LastWriteWins strategy)
					if existingProduct.UpdatedAt.After(product.UpdatedAt) {
						continue // Skip update, server version is newer
					}
				}

				// Update existing product
				if err := s.updateProduct(ctx, tx, product); err != nil {
					s.addError(response, "products", product.ID, "update_failed", err.Error())
					continue
				}
				s.incrementStat(response.Stats.UpdatedEntities, "products")
			}

			s.incrementStat(response.Stats.ProcessedEntities, "products")
		}

		return nil
	})

	if s.config.EnablePerformanceLog {
		log.Printf("Optimized product processing: %d products processed in %v", len(products), time.Since(startTime))
	}

	return nil
}

func (s *SyncService) pullProducts(ctx context.Context, tx *gorm.DB, lastSync time.Time, licenseID uuid.UUID, response *dto.SyncResponse) error {
	// Performance optimization: Use optimized query with proper indexing
	if s.config.EnableQueryOptimization {
		return s.pullProductsOptimized(ctx, tx, lastSync, licenseID, response)
	}

	// Fallback to original implementation
	var products []entities.Product
	err := tx.WithContext(ctx).
		Select("products.*").
		Table("products").
		Joins("JOIN shops ON products.shop_id = shops.id").
		Where("shops.license_id = ? AND products.updated_at > ?", licenseID, lastSync).
		Find(&products).Error

	if err != nil {
		return fmt.Errorf("failed to query products: %w", err)
	}

	// Map entities to DTOs to exclude relationship fields
	response.Products = dto.MapProductsToSyncDTOs(products)
	return nil
}

// pullProductsOptimized uses index-optimized queries for better performance
func (s *SyncService) pullProductsOptimized(ctx context.Context, tx *gorm.DB, lastSync time.Time, licenseID uuid.UUID, response *dto.SyncResponse) error {
	startTime := time.Now()

	// First get all shop IDs for this license (uses idx_shops_license_id index)
	var shopIDs []uuid.UUID
	err := tx.WithContext(ctx).
		Model(&entities.Shop{}).
		Select("id").
		Where("license_id = ?", licenseID).
		Pluck("id", &shopIDs).Error

	if err != nil {
		return fmt.Errorf("failed to query shops for license: %w", err)
	}

	if len(shopIDs) == 0 {
		response.Products = []dto.SyncProductDTO{}
		return nil
	}

	// Query products using shop_id + updated_at index (idx_products_shop_updated)
	var products []entities.Product
	query := tx.WithContext(ctx).
		Select("products.*").
		Where("products.shop_id IN (?) AND products.updated_at > ?", shopIDs, lastSync).
		Order("products.updated_at ASC")

	// Add index hint if enabled
	if s.config.EnableIndexHints {
		query = s.optimizer.OptimizeQueryWithIndexHints(query, "products", "sync_pull")
	}

	// Limit result set to prevent memory issues
	if s.config.MaxResultsPerQuery > 0 {
		query = query.Limit(s.config.MaxResultsPerQuery)
	}

	err = query.Find(&products).Error
	if err != nil {
		return fmt.Errorf("failed to query products: %w", err)
	}

	// Map entities to DTOs to exclude relationship fields
	response.Products = dto.MapProductsToSyncDTOs(products)

	if s.config.EnablePerformanceLog {
		log.Printf("Optimized product pull: %d products from %d shops in %v",
			len(products), len(shopIDs), time.Since(startTime))
	}

	return nil
}

// Helper methods for product operations
func (s *SyncService) validateProductLicense(ctx context.Context, product entities.Product, licenseID uuid.UUID) bool {
	// Validate that the product's shop belongs to the license
	var count int64
	s.db.Model(&entities.Shop{}).Where("id = ? AND license_id = ?", product.ShopID, licenseID).Count(&count)
	return count > 0
}

func (s *SyncService) findProductByID(ctx context.Context, tx *gorm.DB, id uuid.UUID) (*entities.Product, error) {
	var product entities.Product
	err := tx.WithContext(ctx).Where("id = ?", id).First(&product).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &product, nil
}

func (s *SyncService) createProduct(ctx context.Context, tx *gorm.DB, product entities.Product) error {
	return tx.WithContext(ctx).Create(&product).Error
}

func (s *SyncService) updateProduct(ctx context.Context, tx *gorm.DB, product entities.Product) error {
	return tx.WithContext(ctx).Save(&product).Error
}

func (s *SyncService) resolveProductConflict(existing, incoming entities.Product) *dto.ConflictInfo {
	if existing.UpdatedAt.Equal(incoming.UpdatedAt) {
		return nil // No conflict
	}

	conflict := &dto.ConflictInfo{
		EntityType:   "product",
		EntityID:     existing.ID,
		ConflictType: "timestamp_mismatch",
		ServerData:   existing,
		ClientData:   incoming,
	}

	// Apply conflict resolution strategy
	switch s.conflictStrategy {
	case dto.LastWriteWins:
		if existing.UpdatedAt.After(incoming.UpdatedAt) {
			conflict.Resolution = "server_wins"
			conflict.Details = "Server version is newer"
		} else {
			conflict.Resolution = "client_wins"
			conflict.Details = "Client version is newer"
		}
	case dto.ServerWins:
		conflict.Resolution = "server_wins"
		conflict.Details = "Server version always wins"
	case dto.ClientWins:
		conflict.Resolution = "client_wins"
		conflict.Details = "Client version always wins"
	}

	return conflict
}

func (s *SyncService) pushTransactions(ctx context.Context, tx *gorm.DB, transactions []entities.Transaction, licenseID uuid.UUID, response *dto.SyncResponse) error {
	// Performance optimization: Use bulk validation instead of individual queries
	if s.config.EnableBulkValidation && len(transactions) > 1 {
		return s.pushTransactionsOptimized(ctx, tx, transactions, licenseID, response)
	}

	// Fallback to original implementation for single transactions or when optimization is disabled
	for _, transaction := range transactions {
		// Validate transaction belongs to license
		if !s.validateTransactionLicense(ctx, transaction, licenseID) {
			s.addError(response, "transactions", transaction.ID, "unauthorized", "Transaction does not belong to license")
			continue
		}

		// Check if transaction exists
		existingTransaction, err := s.findTransactionByID(ctx, tx, transaction.ID)
		if err != nil && err != gorm.ErrRecordNotFound {
			s.addError(response, "transactions", transaction.ID, "database_error", err.Error())
			continue
		}

		if existingTransaction == nil {
			// Create new transaction
			if err := s.createTransaction(ctx, tx, transaction); err != nil {
				s.addError(response, "transactions", transaction.ID, "create_failed", err.Error())
				continue
			}
			s.incrementStat(response.Stats.CreatedEntities, "transactions")
		} else {
			// Handle potential conflict
			if conflict := s.resolveTransactionConflict(*existingTransaction, transaction); conflict != nil {
				response.Conflicts = append(response.Conflicts, *conflict)
				// Use server version in case of conflict (for LastWriteWins strategy)
				if existingTransaction.UpdatedAt.After(transaction.UpdatedAt) {
					continue // Skip update, server version is newer
				}
			}

			// Update existing transaction
			if err := s.updateTransaction(ctx, tx, transaction); err != nil {
				s.addError(response, "transactions", transaction.ID, "update_failed", err.Error())
				continue
			}
			s.incrementStat(response.Stats.UpdatedEntities, "transactions")
		}

		s.incrementStat(response.Stats.ProcessedEntities, "transactions")
	}

	return nil
}

// pushTransactionsOptimized uses bulk validation and optimized processing for better performance
func (s *SyncService) pushTransactionsOptimized(ctx context.Context, tx *gorm.DB, transactions []entities.Transaction, licenseID uuid.UUID, response *dto.SyncResponse) error {
	startTime := time.Now()

	// Step 1: Bulk validate all transactions belong to the license
	transactionValidation, err := s.optimizer.BulkValidateTransactionLicenses(ctx, transactions, licenseID)
	if err != nil {
		log.Printf("Bulk validation failed, falling back to individual validation: %v", err)
		return s.pushTransactions(ctx, tx, transactions, licenseID, response) // Fallback
	}

	// Step 2: Bulk check which transactions already exist
	transactionIDs := make([]uuid.UUID, len(transactions))
	for i, transaction := range transactions {
		transactionIDs[i] = transaction.ID
	}

	existingTransactions, err := s.optimizer.BulkFindExistingEntities(ctx, tx, "transactions", transactionIDs)
	if err != nil {
		log.Printf("Bulk existence check failed, falling back to individual processing: %v", err)
		return s.pushTransactions(ctx, tx, transactions, licenseID, response) // Fallback
	}

	// Step 3: Process transactions in optimized batches
	return s.optimizer.BatchProcessEntities(ctx, len(transactions), s.config.OptimalBatchSize, func(startIdx, endIdx int) error {
		batch := transactions[startIdx:endIdx]

		for _, transaction := range batch {
			// Check bulk validation result
			if !transactionValidation[transaction.ID] {
				s.addError(response, "transactions", transaction.ID, "unauthorized", "Transaction does not belong to license")
				continue
			}

			transactionExists := existingTransactions[transaction.ID]

			if !transactionExists {
				// Create new transaction
				if err := s.createTransaction(ctx, tx, transaction); err != nil {
					s.addError(response, "transactions", transaction.ID, "create_failed", err.Error())
					continue
				}
				s.incrementStat(response.Stats.CreatedEntities, "transactions")
			} else {
				// Get existing transaction for conflict resolution
				existingTransaction, err := s.findTransactionByID(ctx, tx, transaction.ID)
				if err != nil {
					s.addError(response, "transactions", transaction.ID, "database_error", err.Error())
					continue
				}

				// Handle potential conflict
				if conflict := s.resolveTransactionConflict(*existingTransaction, transaction); conflict != nil {
					response.Conflicts = append(response.Conflicts, *conflict)
					// Use server version in case of conflict (for LastWriteWins strategy)
					if existingTransaction.UpdatedAt.After(transaction.UpdatedAt) {
						continue // Skip update, server version is newer
					}
				}

				// Update existing transaction
				if err := s.updateTransaction(ctx, tx, transaction); err != nil {
					s.addError(response, "transactions", transaction.ID, "update_failed", err.Error())
					continue
				}
				s.incrementStat(response.Stats.UpdatedEntities, "transactions")
			}

			s.incrementStat(response.Stats.ProcessedEntities, "transactions")
		}

		return nil
	})

	if s.config.EnablePerformanceLog {
		log.Printf("Optimized transaction processing: %d transactions processed in %v", len(transactions), time.Since(startTime))
	}

	return nil
}

func (s *SyncService) pullTransactions(ctx context.Context, tx *gorm.DB, lastSync time.Time, licenseID uuid.UUID, response *dto.SyncResponse) error {
	// Performance optimization: Use optimized query with proper indexing
	if s.config.EnableQueryOptimization {
		return s.pullTransactionsOptimized(ctx, tx, lastSync, licenseID, response)
	}

	// Fallback to original implementation
	var transactions []entities.Transaction
	err := tx.WithContext(ctx).
		Select("transactions.*").
		Table("transactions").
		Joins("JOIN shops ON transactions.shop_id = shops.id").
		Where("shops.license_id = ? AND transactions.updated_at > ?", licenseID, lastSync).
		Find(&transactions).Error

	if err != nil {
		return fmt.Errorf("failed to query transactions: %w", err)
	}

	// Map entities to DTOs to exclude relationship fields
	response.Transactions = dto.MapTransactionsToSyncDTOs(transactions)
	return nil
}

// pullTransactionsOptimized uses index-optimized queries for better performance
func (s *SyncService) pullTransactionsOptimized(ctx context.Context, tx *gorm.DB, lastSync time.Time, licenseID uuid.UUID, response *dto.SyncResponse) error {
	startTime := time.Now()

	// Use cached shop IDs if available
	var shopIDs []uuid.UUID
	var err error

	if s.config.EnableCaching {
		// Try to get shop IDs from cache
		shopLicenseMap, cacheErr := s.cacheManager.GetShopLicenseMapping(ctx, licenseID)
		if cacheErr == nil {
			shopIDs = make([]uuid.UUID, 0, len(shopLicenseMap))
			for shopID := range shopLicenseMap {
				shopIDs = append(shopIDs, shopID)
			}
		} else {
			log.Printf("Cache lookup failed, using direct query: %v", cacheErr)
		}
	}

	// Fallback to direct query if cache is disabled or failed
	if len(shopIDs) == 0 {
		err = tx.WithContext(ctx).
			Model(&entities.Shop{}).
			Select("id").
			Where("license_id = ?", licenseID).
			Pluck("id", &shopIDs).Error

		if err != nil {
			return fmt.Errorf("failed to query shops for license: %w", err)
		}
	}

	if len(shopIDs) == 0 {
		response.Transactions = []dto.SyncTransactionDTO{}
		return nil
	}

	// Query transactions using shop_id + updated_at index (idx_transactions_shop_updated)
	var transactions []entities.Transaction
	query := tx.WithContext(ctx).
		Select("transactions.*").
		Where("transactions.shop_id IN (?) AND transactions.updated_at > ?", shopIDs, lastSync).
		Order("transactions.updated_at ASC")

	// Add index hint if enabled
	if s.config.EnableIndexHints {
		query = s.optimizer.OptimizeQueryWithIndexHints(query, "transactions", "sync_pull")
	}

	// Limit result set to prevent memory issues
	if s.config.MaxResultsPerQuery > 0 {
		query = query.Limit(s.config.MaxResultsPerQuery)
	}

	err = query.Find(&transactions).Error
	if err != nil {
		return fmt.Errorf("failed to query transactions: %w", err)
	}

	// Map entities to DTOs to exclude relationship fields
	response.Transactions = dto.MapTransactionsToSyncDTOs(transactions)

	if s.config.EnablePerformanceLog {
		log.Printf("Optimized transaction pull: %d transactions from %d shops in %v",
			len(transactions), len(shopIDs), time.Since(startTime))
	}

	return nil
}

// Helper methods for transaction operations
func (s *SyncService) validateTransactionLicense(ctx context.Context, transaction entities.Transaction, licenseID uuid.UUID) bool {
	// Validate that the transaction's shop belongs to the license
	var count int64
	s.db.Model(&entities.Shop{}).Where("id = ? AND license_id = ?", transaction.ShopID, licenseID).Count(&count)
	return count > 0
}

func (s *SyncService) findTransactionByID(ctx context.Context, tx *gorm.DB, id uuid.UUID) (*entities.Transaction, error) {
	var transaction entities.Transaction
	err := tx.WithContext(ctx).Where("id = ?", id).First(&transaction).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &transaction, nil
}

func (s *SyncService) createTransaction(ctx context.Context, tx *gorm.DB, transaction entities.Transaction) error {
	return tx.WithContext(ctx).Create(&transaction).Error
}

func (s *SyncService) updateTransaction(ctx context.Context, tx *gorm.DB, transaction entities.Transaction) error {
	return tx.WithContext(ctx).Save(&transaction).Error
}

func (s *SyncService) resolveTransactionConflict(existing, incoming entities.Transaction) *dto.ConflictInfo {
	if existing.UpdatedAt.Equal(incoming.UpdatedAt) {
		return nil // No conflict
	}

	conflict := &dto.ConflictInfo{
		EntityType:   "transaction",
		EntityID:     existing.ID,
		ConflictType: "timestamp_mismatch",
		ServerData:   existing,
		ClientData:   incoming,
	}

	// Apply conflict resolution strategy
	switch s.conflictStrategy {
	case dto.LastWriteWins:
		if existing.UpdatedAt.After(incoming.UpdatedAt) {
			conflict.Resolution = "server_wins"
			conflict.Details = "Server version is newer"
		} else {
			conflict.Resolution = "client_wins"
			conflict.Details = "Client version is newer"
		}
	case dto.ServerWins:
		conflict.Resolution = "server_wins"
		conflict.Details = "Server version always wins"
	case dto.ClientWins:
		conflict.Resolution = "client_wins"
		conflict.Details = "Client version always wins"
	}

	return conflict
}

// pushExpenses handles expense synchronization
func (s *SyncService) pushExpenses(ctx context.Context, tx *gorm.DB, expenses []entities.Expense, licenseID uuid.UUID, response *dto.SyncResponse) error {
	for _, expense := range expenses {
		// Validate expense belongs to license
		if !s.validateExpenseLicense(ctx, expense, licenseID) {
			s.addError(response, "expenses", expense.ID, "unauthorized", "Expense does not belong to license")
			continue
		}

		// Check if expense exists
		existingExpense, err := s.findExpenseByID(ctx, tx, expense.ID)
		if err != nil && err != gorm.ErrRecordNotFound {
			s.addError(response, "expenses", expense.ID, "database_error", err.Error())
			continue
		}

		if existingExpense == nil {
			// Create new expense
			if err := s.createExpense(ctx, tx, expense); err != nil {
				s.addError(response, "expenses", expense.ID, "create_failed", err.Error())
				continue
			}
			s.incrementStat(response.Stats.CreatedEntities, "expenses")
		} else {
			// Handle potential conflict
			if conflict := s.resolveExpenseConflict(*existingExpense, expense); conflict != nil {
				response.Conflicts = append(response.Conflicts, *conflict)
				// Use server version in case of conflict (for LastWriteWins strategy)
				if existingExpense.UpdatedAt.After(expense.UpdatedAt) {
					continue // Skip update, server version is newer
				}
			}

			// Update existing expense
			if err := s.updateExpense(ctx, tx, expense); err != nil {
				s.addError(response, "expenses", expense.ID, "update_failed", err.Error())
				continue
			}
			s.incrementStat(response.Stats.UpdatedEntities, "expenses")
		}

		s.incrementStat(response.Stats.ProcessedEntities, "expenses")
	}

	return nil
}

func (s *SyncService) pullExpenses(ctx context.Context, tx *gorm.DB, lastSync time.Time, licenseID uuid.UUID, response *dto.SyncResponse) error {
	// Get all expenses for the license that were updated after lastSync
	// Select only the expense fields without relations to avoid unnecessary data
	var expenses []entities.Expense
	err := tx.WithContext(ctx).
		Select("expenses.*").
		Table("expenses").
		Joins("JOIN shops ON expenses.shop_id = shops.id").
		Where("shops.license_id = ? AND expenses.updated_at > ?", licenseID, lastSync).
		Find(&expenses).Error

	if err != nil {
		return fmt.Errorf("failed to query expenses: %w", err)
	}

	// Map entities to DTOs to exclude relationship fields
	response.Expenses = dto.MapExpensesToSyncDTOs(expenses)
	return nil
}

// Helper methods for expense operations
func (s *SyncService) validateExpenseLicense(ctx context.Context, expense entities.Expense, licenseID uuid.UUID) bool {
	// Validate that the expense's shop belongs to the license
	var count int64
	s.db.Model(&entities.Shop{}).Where("id = ? AND license_id = ?", expense.ShopID, licenseID).Count(&count)
	return count > 0
}

func (s *SyncService) findExpenseByID(ctx context.Context, tx *gorm.DB, id uuid.UUID) (*entities.Expense, error) {
	var expense entities.Expense
	err := tx.WithContext(ctx).Where("id = ?", id).First(&expense).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &expense, nil
}

func (s *SyncService) createExpense(ctx context.Context, tx *gorm.DB, expense entities.Expense) error {
	return tx.WithContext(ctx).Create(&expense).Error
}

func (s *SyncService) updateExpense(ctx context.Context, tx *gorm.DB, expense entities.Expense) error {
	return tx.WithContext(ctx).Save(&expense).Error
}

func (s *SyncService) resolveExpenseConflict(existing, incoming entities.Expense) *dto.ConflictInfo {
	if existing.UpdatedAt.Equal(incoming.UpdatedAt) {
		return nil // No conflict
	}

	conflict := &dto.ConflictInfo{
		EntityType:   "expense",
		EntityID:     existing.ID,
		ConflictType: "timestamp_mismatch",
		ServerData:   existing,
		ClientData:   incoming,
	}

	// Apply conflict resolution strategy
	switch s.conflictStrategy {
	case dto.LastWriteWins:
		if existing.UpdatedAt.After(incoming.UpdatedAt) {
			conflict.Resolution = "server_wins"
			conflict.Details = "Server version is newer"
		} else {
			conflict.Resolution = "client_wins"
			conflict.Details = "Client version is newer"
		}
	case dto.ServerWins:
		conflict.Resolution = "server_wins"
		conflict.Details = "Server version always wins"
	case dto.ClientWins:
		conflict.Resolution = "client_wins"
		conflict.Details = "Client version always wins"
	}

	return conflict
}

// pushPayments handles payment synchronization
func (s *SyncService) pushPayments(ctx context.Context, tx *gorm.DB, payments []entities.Payment, licenseID uuid.UUID, response *dto.SyncResponse) error {
	for _, payment := range payments {
		// Validate payment belongs to license
		if !s.validatePaymentLicense(ctx, payment, licenseID) {
			s.addError(response, "payments", payment.ID, "unauthorized", "Payment does not belong to license")
			continue
		}

		// Check if payment exists
		existingPayment, err := s.findPaymentByID(ctx, tx, payment.ID)
		if err != nil && err != gorm.ErrRecordNotFound {
			s.addError(response, "payments", payment.ID, "database_error", err.Error())
			continue
		}

		if existingPayment == nil {
			// Create new payment
			if err := s.createPayment(ctx, tx, payment); err != nil {
				s.addError(response, "payments", payment.ID, "create_failed", err.Error())
				continue
			}
			s.incrementStat(response.Stats.CreatedEntities, "payments")
		} else {
			// Handle potential conflict
			if conflict := s.resolvePaymentConflict(*existingPayment, payment); conflict != nil {
				response.Conflicts = append(response.Conflicts, *conflict)
				// Use server version in case of conflict (for LastWriteWins strategy)
				if existingPayment.UpdatedAt.After(payment.UpdatedAt) {
					continue // Skip update, server version is newer
				}
			}

			// Update existing payment
			if err := s.updatePayment(ctx, tx, payment); err != nil {
				s.addError(response, "payments", payment.ID, "update_failed", err.Error())
				continue
			}
			s.incrementStat(response.Stats.UpdatedEntities, "payments")
		}

		s.incrementStat(response.Stats.ProcessedEntities, "payments")
	}

	return nil
}

func (s *SyncService) pullPayments(ctx context.Context, tx *gorm.DB, lastSync time.Time, licenseID uuid.UUID, response *dto.SyncResponse) error {
	// Get all payments for the license that were updated after lastSync
	// Select only the payment fields without relations to avoid unnecessary data
	var payments []entities.Payment
	err := tx.WithContext(ctx).
		Select("payments.*").
		Table("payments").
		Joins("JOIN shops ON payments.shop_id = shops.id").
		Where("shops.license_id = ? AND payments.updated_at > ?", licenseID, lastSync).
		Find(&payments).Error

	if err != nil {
		return fmt.Errorf("failed to query payments: %w", err)
	}

	// Map entities to DTOs to exclude relationship fields
	response.Payments = dto.MapPaymentsToSyncDTOs(payments)
	return nil
}

// Helper methods for payment operations
func (s *SyncService) validatePaymentLicense(ctx context.Context, payment entities.Payment, licenseID uuid.UUID) bool {
	// Validate that the payment's shop belongs to the license
	var count int64
	s.db.Model(&entities.Shop{}).Where("id = ? AND license_id = ?", payment.ShopID, licenseID).Count(&count)
	return count > 0
}

func (s *SyncService) findPaymentByID(ctx context.Context, tx *gorm.DB, id uuid.UUID) (*entities.Payment, error) {
	var payment entities.Payment
	err := tx.WithContext(ctx).Where("id = ?", id).First(&payment).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &payment, nil
}

func (s *SyncService) createPayment(ctx context.Context, tx *gorm.DB, payment entities.Payment) error {
	return tx.WithContext(ctx).Create(&payment).Error
}

func (s *SyncService) updatePayment(ctx context.Context, tx *gorm.DB, payment entities.Payment) error {
	return tx.WithContext(ctx).Save(&payment).Error
}

func (s *SyncService) resolvePaymentConflict(existing, incoming entities.Payment) *dto.ConflictInfo {
	if existing.UpdatedAt.Equal(incoming.UpdatedAt) {
		return nil // No conflict
	}

	conflict := &dto.ConflictInfo{
		EntityType:   "payment",
		EntityID:     existing.ID,
		ConflictType: "timestamp_mismatch",
		ServerData:   existing,
		ClientData:   incoming,
	}

	// Apply conflict resolution strategy
	switch s.conflictStrategy {
	case dto.LastWriteWins:
		if existing.UpdatedAt.After(incoming.UpdatedAt) {
			conflict.Resolution = "server_wins"
			conflict.Details = "Server version is newer"
		} else {
			conflict.Resolution = "client_wins"
			conflict.Details = "Client version is newer"
		}
	case dto.ServerWins:
		conflict.Resolution = "server_wins"
		conflict.Details = "Server version always wins"
	case dto.ClientWins:
		conflict.Resolution = "client_wins"
		conflict.Details = "Client version always wins"
	}

	return conflict
}

// Utility methods
func (s *SyncService) addError(response *dto.SyncResponse, entityType string, entityID uuid.UUID, errorCode, message string) {
	response.Errors = append(response.Errors, dto.SyncError{
		EntityType: entityType,
		EntityID:   entityID,
		ErrorCode:  errorCode,
		Message:    message,
	})
}

func (s *SyncService) incrementStat(stats map[string]int, entityType string) {
	if stats == nil {
		stats = make(map[string]int)
	}
	stats[entityType]++
}

// Receipt sync implementation
func (s *SyncService) pushReceipts(ctx context.Context, tx *gorm.DB, receipts []entities.Receipt, licenseID uuid.UUID, response *dto.SyncResponse) error {
	for _, receipt := range receipts {
		// Validate receipt belongs to license
		if !s.validateReceiptLicense(ctx, receipt, licenseID) {
			s.addError(response, "receipts", receipt.ID, "unauthorized", "Receipt does not belong to license")
			continue
		}

		// Check if receipt already exists
		existing, err := s.findReceiptByID(ctx, tx, receipt.ID)
		if err != nil {
			s.addError(response, "receipts", receipt.ID, "database_error", fmt.Sprintf("Failed to find receipt: %v", err))
			continue
		}

		s.incrementStat(response.Stats.ProcessedEntities, "receipts")

		if existing == nil {
			// Create new receipt
			if err := s.createReceipt(ctx, tx, receipt); err != nil {
				s.addError(response, "receipts", receipt.ID, "create_failed", fmt.Sprintf("Failed to create receipt: %v", err))
				continue
			}
			s.incrementStat(response.Stats.CreatedEntities, "receipts")
		} else {
			// Check for conflicts and resolve
			if conflict := s.resolveReceiptConflict(*existing, receipt); conflict != nil {
				response.Conflicts = append(response.Conflicts, *conflict)
			}

			// Update receipt
			if err := s.updateReceipt(ctx, tx, receipt); err != nil {
				s.addError(response, "receipts", receipt.ID, "update_failed", fmt.Sprintf("Failed to update receipt: %v", err))
				continue
			}
			s.incrementStat(response.Stats.UpdatedEntities, "receipts")
		}
	}
	return nil
}

func (s *SyncService) pullReceipts(ctx context.Context, tx *gorm.DB, lastSync time.Time, licenseID uuid.UUID, response *dto.SyncResponse) error {
	// Select only the receipt fields without relations to avoid unnecessary data
	var receipts []entities.Receipt
	err := tx.WithContext(ctx).
		Select("receipts.*").
		Table("receipts").
		Joins("JOIN shops ON receipts.shop_id = shops.id").
		Where("shops.license_id = ? AND receipts.updated_at > ?", licenseID, lastSync).
		Find(&receipts).Error
	if err != nil {
		return fmt.Errorf("failed to fetch receipts: %w", err)
	}

	// Map entities to DTOs to exclude relationship fields
	response.Receipts = dto.MapReceiptsToSyncDTOs(receipts)
	return nil
}

func (s *SyncService) validateReceiptLicense(ctx context.Context, receipt entities.Receipt, licenseID uuid.UUID) bool {
	var count int64
	err := s.db.Model(&entities.Shop{}).
		Where("id = ? AND license_id = ?", receipt.ShopID, licenseID).
		Count(&count).Error
	return err == nil && count > 0
}

func (s *SyncService) findReceiptByID(ctx context.Context, tx *gorm.DB, id uuid.UUID) (*entities.Receipt, error) {
	var receipt entities.Receipt
	err := tx.First(&receipt, "id = ?", id).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &receipt, nil
}

func (s *SyncService) createReceipt(ctx context.Context, tx *gorm.DB, receipt entities.Receipt) error {
	return tx.Create(&receipt).Error
}

func (s *SyncService) updateReceipt(ctx context.Context, tx *gorm.DB, receipt entities.Receipt) error {
	return tx.Save(&receipt).Error
}

func (s *SyncService) resolveReceiptConflict(existing, incoming entities.Receipt) *dto.ConflictInfo {
	// Check if there's actually a conflict (different updated_at times)
	if existing.UpdatedAt.Equal(incoming.UpdatedAt) {
		return nil // No conflict
	}

	conflict := &dto.ConflictInfo{
		EntityType:   "receipt",
		EntityID:     existing.ID,
		ConflictType: "timestamp_mismatch",
		ServerData:   existing,
		ClientData:   incoming,
	}

	// Apply conflict resolution strategy
	switch s.conflictStrategy {
	case dto.LastWriteWins:
		if incoming.UpdatedAt.After(existing.UpdatedAt) {
			conflict.Resolution = "client_wins"
			conflict.Details = fmt.Sprintf("Client version is newer (%s vs %s)", incoming.UpdatedAt, existing.UpdatedAt)
		} else {
			conflict.Resolution = "server_wins"
			conflict.Details = fmt.Sprintf("Server version is newer (%s vs %s)", existing.UpdatedAt, incoming.UpdatedAt)
		}
	case dto.ServerWins:
		conflict.Resolution = "server_wins"
		conflict.Details = "Server always wins conflict resolution strategy"
	case dto.ClientWins:
		conflict.Resolution = "client_wins"
		conflict.Details = "Client always wins conflict resolution strategy"
	}

	return conflict
}

// History sync implementation
func (s *SyncService) pushHistories(ctx context.Context, tx *gorm.DB, histories []entities.History, licenseID uuid.UUID, response *dto.SyncResponse) error {
	for _, history := range histories {
		// Validate history belongs to license
		if !s.validateHistoryLicense(ctx, history, licenseID) {
			s.addError(response, "histories", history.ID, "unauthorized", "History does not belong to license")
			continue
		}

		// Check if history already exists
		existing, err := s.findHistoryByID(ctx, tx, history.ID)
		if err != nil {
			s.addError(response, "histories", history.ID, "database_error", fmt.Sprintf("Failed to find history: %v", err))
			continue
		}

		s.incrementStat(response.Stats.ProcessedEntities, "histories")

		if existing == nil {
			// Create new history
			if err := s.createHistory(ctx, tx, history); err != nil {
				s.addError(response, "histories", history.ID, "create_failed", fmt.Sprintf("Failed to create history: %v", err))
				continue
			}
			s.incrementStat(response.Stats.CreatedEntities, "histories")
		} else {
			// Check for conflicts and resolve
			if conflict := s.resolveHistoryConflict(*existing, history); conflict != nil {
				response.Conflicts = append(response.Conflicts, *conflict)
			}

			// Update history
			if err := s.updateHistory(ctx, tx, history); err != nil {
				s.addError(response, "histories", history.ID, "update_failed", fmt.Sprintf("Failed to update history: %v", err))
				continue
			}
			s.incrementStat(response.Stats.UpdatedEntities, "histories")
		}
	}
	return nil
}

func (s *SyncService) pullHistories(ctx context.Context, tx *gorm.DB, lastSync time.Time, licenseID uuid.UUID, response *dto.SyncResponse) error {
	// Select only the history fields without relations to avoid unnecessary data
	var histories []entities.History
	err := tx.WithContext(ctx).
		Select("histories.*").
		Table("histories").
		Joins("JOIN shops ON histories.shop_id = shops.id").
		Where("shops.license_id = ? AND histories.updated_at > ?", licenseID, lastSync).
		Find(&histories).Error
	if err != nil {
		return fmt.Errorf("failed to fetch histories: %w", err)
	}

	// Map entities to DTOs to exclude relationship fields
	response.Histories = dto.MapHistoriesToSyncDTOs(histories)
	return nil
}

func (s *SyncService) validateHistoryLicense(ctx context.Context, history entities.History, licenseID uuid.UUID) bool {
	var count int64
	err := s.db.Model(&entities.Shop{}).
		Where("id = ? AND license_id = ?", history.ShopID, licenseID).
		Count(&count).Error
	return err == nil && count > 0
}

func (s *SyncService) findHistoryByID(ctx context.Context, tx *gorm.DB, id uuid.UUID) (*entities.History, error) {
	var history entities.History
	err := tx.First(&history, "id = ?", id).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &history, nil
}

func (s *SyncService) createHistory(ctx context.Context, tx *gorm.DB, history entities.History) error {
	return tx.Create(&history).Error
}

func (s *SyncService) updateHistory(ctx context.Context, tx *gorm.DB, history entities.History) error {
	return tx.Save(&history).Error
}

func (s *SyncService) resolveHistoryConflict(existing, incoming entities.History) *dto.ConflictInfo {
	// Check if there's actually a conflict (different updated_at times)
	if existing.UpdatedAt.Equal(incoming.UpdatedAt) {
		return nil // No conflict
	}

	conflict := &dto.ConflictInfo{
		EntityType:   "history",
		EntityID:     existing.ID,
		ConflictType: "timestamp_mismatch",
		ServerData:   existing,
		ClientData:   incoming,
	}

	// Apply conflict resolution strategy
	switch s.conflictStrategy {
	case dto.LastWriteWins:
		if incoming.UpdatedAt.After(existing.UpdatedAt) {
			conflict.Resolution = "client_wins"
			conflict.Details = fmt.Sprintf("Client version is newer (%s vs %s)", incoming.UpdatedAt, existing.UpdatedAt)
		} else {
			conflict.Resolution = "server_wins"
			conflict.Details = fmt.Sprintf("Server version is newer (%s vs %s)", existing.UpdatedAt, incoming.UpdatedAt)
		}
	case dto.ServerWins:
		conflict.Resolution = "server_wins"
		conflict.Details = "Server always wins conflict resolution strategy"
	case dto.ClientWins:
		conflict.Resolution = "client_wins"
		conflict.Details = "Client always wins conflict resolution strategy"
	}

	return conflict
}

// Shop sync implementation
func (s *SyncService) pushShops(ctx context.Context, tx *gorm.DB, shops []entities.Shop, licenseID uuid.UUID, response *dto.SyncResponse) error {
	for _, shop := range shops {
		// Validate shop belongs to license
		if !s.validateShopLicense(ctx, shop, licenseID) {
			s.addError(response, "shops", shop.ID, "unauthorized", "Shop does not belong to license")
			continue
		}

		// Check if shop already exists
		existing, err := s.findShopByID(ctx, tx, shop.ID)
		if err != nil {
			s.addError(response, "shops", shop.ID, "database_error", fmt.Sprintf("Failed to find shop: %v", err))
			continue
		}

		s.incrementStat(response.Stats.ProcessedEntities, "shops")

		if existing == nil {
			// Create new shop
			if err := s.createShop(ctx, tx, shop); err != nil {
				s.addError(response, "shops", shop.ID, "create_failed", fmt.Sprintf("Failed to create shop: %v", err))
				continue
			}
			s.incrementStat(response.Stats.CreatedEntities, "shops")
		} else {
			// Check for conflicts and resolve
			if conflict := s.resolveShopConflict(*existing, shop); conflict != nil {
				response.Conflicts = append(response.Conflicts, *conflict)
			}

			// Update shop
			if err := s.updateShop(ctx, tx, shop); err != nil {
				s.addError(response, "shops", shop.ID, "update_failed", fmt.Sprintf("Failed to update shop: %v", err))
				continue
			}
			s.incrementStat(response.Stats.UpdatedEntities, "shops")
		}
	}
	return nil
}

func (s *SyncService) pullShops(ctx context.Context, tx *gorm.DB, lastSync time.Time, licenseID uuid.UUID, response *dto.SyncResponse) error {
	// Select only the shop fields without relations to avoid unnecessary data
	var shops []entities.Shop
	err := tx.WithContext(ctx).
		Select("shops.*").
		Where("license_id = ? AND updated_at > ?", licenseID, lastSync).
		Find(&shops).Error
	if err != nil {
		return fmt.Errorf("failed to fetch shops: %w", err)
	}

	// Map entities to DTOs to exclude relationship fields
	response.Shops = dto.MapShopsToSyncDTOs(shops)
	return nil
}

func (s *SyncService) validateShopLicense(ctx context.Context, shop entities.Shop, licenseID uuid.UUID) bool {
	return shop.LicenseID == licenseID
}

func (s *SyncService) findShopByID(ctx context.Context, tx *gorm.DB, id uuid.UUID) (*entities.Shop, error) {
	var shop entities.Shop
	err := tx.First(&shop, "id = ?", id).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &shop, nil
}

func (s *SyncService) createShop(ctx context.Context, tx *gorm.DB, shop entities.Shop) error {
	return tx.Create(&shop).Error
}

func (s *SyncService) updateShop(ctx context.Context, tx *gorm.DB, shop entities.Shop) error {
	return tx.Save(&shop).Error
}

func (s *SyncService) resolveShopConflict(existing, incoming entities.Shop) *dto.ConflictInfo {
	// Check if there's actually a conflict (different updated_at times)
	if existing.UpdatedAt.Equal(incoming.UpdatedAt) {
		return nil // No conflict
	}

	conflict := &dto.ConflictInfo{
		EntityType:   "shop",
		EntityID:     existing.ID,
		ConflictType: "timestamp_mismatch",
		ServerData:   existing,
		ClientData:   incoming,
	}

	// Apply conflict resolution strategy
	switch s.conflictStrategy {
	case dto.LastWriteWins:
		if incoming.UpdatedAt.After(existing.UpdatedAt) {
			conflict.Resolution = "client_wins"
			conflict.Details = fmt.Sprintf("Client version is newer (%s vs %s)", incoming.UpdatedAt, existing.UpdatedAt)
		} else {
			conflict.Resolution = "server_wins"
			conflict.Details = fmt.Sprintf("Server version is newer (%s vs %s)", existing.UpdatedAt, incoming.UpdatedAt)
		}
	case dto.ServerWins:
		conflict.Resolution = "server_wins"
		conflict.Details = "Server always wins conflict resolution strategy"
	case dto.ClientWins:
		conflict.Resolution = "client_wins"
		conflict.Details = "Client always wins conflict resolution strategy"
	}

	return conflict
}

// StockHistory sync implementation
func (s *SyncService) pushStockHistories(ctx context.Context, tx *gorm.DB, stockHistories []entities.StockHistory, syncContext dto.SyncContext, response *dto.SyncResponse) error {
	for _, stockHistory := range stockHistories {
		// Enhanced validation: check both license and shop access
		if !s.validateStockHistoryAccess(ctx, stockHistory, syncContext) {
			s.addError(response, "stock_histories", stockHistory.ID, "unauthorized", "Stock history access denied for user role and shop")
			continue
		}

		// Check if stock history already exists
		existing, err := s.findStockHistoryByID(ctx, tx, stockHistory.ID)
		if err != nil {
			s.addError(response, "stock_histories", stockHistory.ID, "database_error", fmt.Sprintf("Failed to find stock history: %v", err))
			continue
		}

		s.incrementStat(response.Stats.ProcessedEntities, "stock_histories")

		if existing == nil {
			// Create new stock history
			if err := s.createStockHistory(ctx, tx, stockHistory); err != nil {
				s.addError(response, "stock_histories", stockHistory.ID, "create_failed", fmt.Sprintf("Failed to create stock history: %v", err))
				continue
			}
			s.incrementStat(response.Stats.CreatedEntities, "stock_histories")
		} else {
			// Check for conflicts and resolve
			if conflict := s.resolveStockHistoryConflict(*existing, stockHistory); conflict != nil {
				response.Conflicts = append(response.Conflicts, *conflict)
			}

			// Update stock history
			if err := s.updateStockHistory(ctx, tx, stockHistory); err != nil {
				s.addError(response, "stock_histories", stockHistory.ID, "update_failed", fmt.Sprintf("Failed to update stock history: %v", err))
				continue
			}
			s.incrementStat(response.Stats.UpdatedEntities, "stock_histories")
		}
	}
	return nil
}

func (s *SyncService) pullStockHistories(ctx context.Context, tx *gorm.DB, lastSync time.Time, licenseID uuid.UUID, response *dto.SyncResponse) error {
	// Select only the stock_history fields without relations to avoid unnecessary data
	var stockHistories []entities.StockHistory
	err := tx.WithContext(ctx).
		Select("stock_histories.*").
		Table("stock_histories").
		Joins("JOIN products ON stock_histories.product_id = products.id").
		Joins("JOIN shops ON products.shop_id = shops.id").
		Where("shops.license_id = ? AND stock_histories.updated_at > ?", licenseID, lastSync).
		Find(&stockHistories).Error
	if err != nil {
		return fmt.Errorf("failed to fetch stock histories: %w", err)
	}

	// Map entities to DTOs to exclude relationship fields
	response.StockHistories = dto.MapStockHistoriesToSyncDTOs(stockHistories)
	return nil
}

func (s *SyncService) validateStockHistoryLicense(ctx context.Context, stockHistory entities.StockHistory, licenseID uuid.UUID) bool {
	var count int64
	err := s.db.Model(&entities.Product{}).
		Joins("JOIN shops ON products.shop_id = shops.id").
		Where("products.id = ? AND shops.license_id = ?", stockHistory.ProductID, licenseID).
		Count(&count).Error
	return err == nil && count > 0
}

// filterStockHistoriesByShopAccessWithSyncData filters stock histories checking both sync request data and database
func (s *SyncService) filterStockHistoriesByShopAccessWithSyncData(stockHistories []entities.StockHistory, syncProducts []entities.Product, accessibleShops map[uuid.UUID]bool, syncContext dto.SyncContext) []entities.StockHistory {
	if syncContext.HasGlobalAccess {
		return stockHistories
	}

	// Performance optimization: Use bulk validation when enabled and there are many stock histories
	if s.config.EnableBulkValidation && len(stockHistories) > 10 {
		return s.filterStockHistoriesByShopAccessOptimized(stockHistories, syncProducts, syncContext.AccessibleShopIDs, syncContext)
	}

	// Fallback to original implementation for smaller sets or when optimization is disabled
	var filtered []entities.StockHistory

	// Create a map of products in the sync request for quick lookup
	syncProductMap := make(map[uuid.UUID]uuid.UUID) // product_id -> shop_id
	for _, product := range syncProducts {
		syncProductMap[product.ID] = product.ShopID
	}

	log.Printf("DEBUG: filterStockHistoriesByShopAccessWithSyncData - User %s (role: %s), accessible shops: %v, processing %d stock histories, %d products in sync",
		syncContext.UserID, syncContext.UserRole, syncContext.AccessibleShopIDs, len(stockHistories), len(syncProducts))

	for i, stockHistory := range stockHistories {
		log.Printf("DEBUG: Processing stock history %d/%d: ID=%s, ProductID=%s",
			i+1, len(stockHistories), stockHistory.ID, stockHistory.ProductID)

		var productShopID uuid.UUID
		var productFound bool

		// First check if product is in the sync request data
		if shopID, exists := syncProductMap[stockHistory.ProductID]; exists {
			productShopID = shopID
			productFound = true
			log.Printf("DEBUG: Stock history %s - Product %s found in sync request data, belongs to shop %s", stockHistory.ID, stockHistory.ProductID, productShopID)
		} else {
			// If not in sync data, check the database with enhanced error handling
			log.Printf("DEBUG: Stock history %s - Product %s not in sync request, checking database...", stockHistory.ID, stockHistory.ProductID)

			var product entities.Product
			err := s.db.Select("shop_id").
				Where("id = ?", stockHistory.ProductID).
				First(&product).Error

			if err != nil {
				if err == gorm.ErrRecordNotFound {
					log.Printf("ERROR: Stock history %s references non-existent product %s (not found in database)", stockHistory.ID, stockHistory.ProductID)
				} else {
					log.Printf("ERROR: Stock history %s - database error when checking product %s: %v", stockHistory.ID, stockHistory.ProductID, err)
				}
				continue
			}
			productShopID = product.ShopID
			productFound = true
			log.Printf("DEBUG: Stock history %s - Product %s found in database, belongs to shop %s", stockHistory.ID, stockHistory.ProductID, productShopID)
		}

		if !productFound {
			continue
		}

		// CRITICAL FIX: Check if shop is accessible using direct shop ID comparison instead of map lookup
		// The accessibleShops map might have issues, so let's use direct slice comparison
		isAccessible := false
		for _, accessibleShopID := range syncContext.AccessibleShopIDs {
			if accessibleShopID == productShopID {
				isAccessible = true
				break
			}
		}

		log.Printf("DEBUG: Stock history %s - Shop %s accessible check: %v (user accessible shops: %v)",
			stockHistory.ID, productShopID, isAccessible, syncContext.AccessibleShopIDs)

		if !isAccessible {
			log.Printf("WARNING: Stock history %s filtered - product %s in inaccessible shop %s (user %s has access to: %v)",
				stockHistory.ID, stockHistory.ProductID, productShopID, syncContext.UserID, syncContext.AccessibleShopIDs)
			continue
		}

		log.Printf("DEBUG: Stock history %s PASSED filtering (product %s in accessible shop %s)",
			stockHistory.ID, stockHistory.ProductID, productShopID)
		filtered = append(filtered, stockHistory)
	}

	log.Printf("SUMMARY: Filtered stock histories for user %s (role: %s): %d → %d",
		syncContext.UserID, syncContext.UserRole, len(stockHistories), len(filtered))
	return filtered
}

// filterStockHistoriesByShopAccessOptimized uses bulk database queries for better performance
func (s *SyncService) filterStockHistoriesByShopAccessOptimized(stockHistories []entities.StockHistory, syncProducts []entities.Product, accessibleShopIDs []uuid.UUID, syncContext dto.SyncContext) []entities.StockHistory {
	startTime := time.Now()

	// Create a map of products in the sync request for quick lookup
	syncProductMap := make(map[uuid.UUID]uuid.UUID) // product_id -> shop_id
	for _, product := range syncProducts {
		syncProductMap[product.ID] = product.ShopID
	}

	// Create accessible shop set for fast lookup
	accessibleShopSet := make(map[uuid.UUID]bool)
	for _, shopID := range accessibleShopIDs {
		accessibleShopSet[shopID] = true
	}

	// Get product IDs that are not in sync data and need database lookup
	missingProductIDs := make([]uuid.UUID, 0)
	for _, stockHistory := range stockHistories {
		if _, exists := syncProductMap[stockHistory.ProductID]; !exists {
			missingProductIDs = append(missingProductIDs, stockHistory.ProductID)
		}
	}

	// Bulk fetch missing products from database
	var missingProductToShop map[uuid.UUID]uuid.UUID
	var err error

	if len(missingProductIDs) > 0 {
		if s.config.EnableCaching {
			// Use cache manager for potentially cached results
			missingProductToShop, err = s.cacheManager.GetProductShopMapping(context.Background(), missingProductIDs)
		} else {
			// Use direct database query for missing products
			var products []entities.Product
			err = s.db.WithContext(context.Background()).
				Select("id, shop_id").
				Where("id IN (?)", missingProductIDs).
				Find(&products).Error

			if err == nil {
				missingProductToShop = make(map[uuid.UUID]uuid.UUID)
				for _, product := range products {
					missingProductToShop[product.ID] = product.ShopID
				}
			}
		}

		if err != nil {
			log.Printf("Bulk product fetch failed, falling back to individual lookup: %v", err)
			// Fallback to original method
			return s.filterStockHistoriesByShopAccessWithSyncData(stockHistories, syncProducts, accessibleShopSet, syncContext)
		}
	} else {
		missingProductToShop = make(map[uuid.UUID]uuid.UUID)
	}

	// Filter stock histories using combined data
	var filtered []entities.StockHistory

	for _, stockHistory := range stockHistories {
		var productShopID uuid.UUID
		var productFound bool

		// Check sync data first
		if shopID, exists := syncProductMap[stockHistory.ProductID]; exists {
			productShopID = shopID
			productFound = true
		} else if shopID, exists := missingProductToShop[stockHistory.ProductID]; exists {
			// Check bulk-fetched data
			productShopID = shopID
			productFound = true
		}

		if !productFound {
			continue
		}

		// Check accessibility
		if accessibleShopSet[productShopID] {
			filtered = append(filtered, stockHistory)
		}
	}

	if s.config.EnablePerformanceLog {
		log.Printf("Optimized stock history filtering: %d → %d stock histories processed in %v (sync products: %d, db lookups: %d)",
			len(stockHistories), len(filtered), time.Since(startTime), len(syncProducts), len(missingProductIDs))
	}

	return filtered
}

// filterTransactionProductsByShopAccessWithSyncData filters transaction products checking both sync request data and database
func (s *SyncService) filterTransactionProductsByShopAccessWithSyncData(transactionProducts []entities.TransactionProduct, syncTransactions []entities.Transaction, accessibleShops map[uuid.UUID]bool, syncContext dto.SyncContext) []entities.TransactionProduct {
	if syncContext.HasGlobalAccess {
		return transactionProducts
	}

	var filtered []entities.TransactionProduct

	// Create a map of transactions in the sync request for quick lookup
	syncTransactionMap := make(map[uuid.UUID]uuid.UUID) // transaction_id -> shop_id
	for _, transaction := range syncTransactions {
		syncTransactionMap[transaction.ID] = transaction.ShopID
	}

	log.Printf("DEBUG: filterTransactionProductsByShopAccessWithSyncData - User %s (role: %s), accessible shops: %v, processing %d transaction products, %d transactions in sync",
		syncContext.UserID, syncContext.UserRole, syncContext.AccessibleShopIDs, len(transactionProducts), len(syncTransactions))

	for i, transactionProduct := range transactionProducts {
		log.Printf("DEBUG: Processing transaction product %d/%d: ID=%s, TransactionID=%s",
			i+1, len(transactionProducts), transactionProduct.ID, transactionProduct.TransactionID)

		var transactionShopID uuid.UUID
		var transactionFound bool

		// First check if transaction is in the sync request data
		if shopID, exists := syncTransactionMap[transactionProduct.TransactionID]; exists {
			transactionShopID = shopID
			transactionFound = true
			log.Printf("DEBUG: Transaction product %s - Transaction %s found in sync request data, belongs to shop %s", transactionProduct.ID, transactionProduct.TransactionID, transactionShopID)
		} else {
			// If not in sync data, check the database with enhanced error handling
			log.Printf("DEBUG: Transaction product %s - Transaction %s not in sync request, checking database...", transactionProduct.ID, transactionProduct.TransactionID)

			var transaction entities.Transaction
			err := s.db.Select("shop_id").
				Where("id = ?", transactionProduct.TransactionID).
				First(&transaction).Error

			if err != nil {
				if err == gorm.ErrRecordNotFound {
					log.Printf("ERROR: Transaction product %s references non-existent transaction %s (not found in database)", transactionProduct.ID, transactionProduct.TransactionID)
				} else {
					log.Printf("ERROR: Transaction product %s - database error when checking transaction %s: %v", transactionProduct.ID, transactionProduct.TransactionID, err)
				}
				continue
			}
			transactionShopID = transaction.ShopID
			transactionFound = true
			log.Printf("DEBUG: Transaction product %s - Transaction %s found in database, belongs to shop %s", transactionProduct.ID, transactionProduct.TransactionID, transactionShopID)
		}

		if !transactionFound {
			continue
		}

		// CRITICAL FIX: Check if shop is accessible using direct shop ID comparison instead of map lookup
		// The accessibleShops map might have issues, so let's use direct slice comparison
		isAccessible := false
		for _, accessibleShopID := range syncContext.AccessibleShopIDs {
			if accessibleShopID == transactionShopID {
				isAccessible = true
				break
			}
		}

		log.Printf("DEBUG: Transaction product %s - Shop %s accessible check: %v (user accessible shops: %v)",
			transactionProduct.ID, transactionShopID, isAccessible, syncContext.AccessibleShopIDs)

		if !isAccessible {
			log.Printf("WARNING: Transaction product %s filtered - transaction %s in inaccessible shop %s (user %s has access to: %v)",
				transactionProduct.ID, transactionProduct.TransactionID, transactionShopID, syncContext.UserID, syncContext.AccessibleShopIDs)
			continue
		}

		log.Printf("DEBUG: Transaction product %s PASSED filtering (transaction %s in accessible shop %s)",
			transactionProduct.ID, transactionProduct.TransactionID, transactionShopID)
		filtered = append(filtered, transactionProduct)
	}

	log.Printf("SUMMARY: Filtered transaction products for user %s (role: %s): %d → %d",
		syncContext.UserID, syncContext.UserRole, len(transactionProducts), len(filtered))
	return filtered
}

// FilterResult holds filtering statistics for generating accurate warnings
type FilterResult struct {
	Filtered          []entities.StockHistory
	MissingProducts   []uuid.UUID
	InaccessibleShops []uuid.UUID
}

// filterStockHistoriesByShopAccess filters stock histories based on accessible shops
func (s *SyncService) filterStockHistoriesByShopAccess(stockHistories []entities.StockHistory, accessibleShops map[uuid.UUID]bool, syncContext dto.SyncContext) []entities.StockHistory {
	if syncContext.HasGlobalAccess {
		return stockHistories
	}

	var filtered []entities.StockHistory

	// Enhanced debug logging
	log.Printf("DEBUG: filterStockHistoriesByShopAccess - User %s (role: %s), accessible shops: %v, processing %d stock histories",
		syncContext.UserID, syncContext.UserRole, syncContext.AccessibleShopIDs, len(stockHistories))

	for i, stockHistory := range stockHistories {
		// Enhanced debug logging for each stock history
		log.Printf("DEBUG: Processing stock history %d/%d: ID=%s, ProductID=%s",
			i+1, len(stockHistories), stockHistory.ID, stockHistory.ProductID)

		// Query to check if the product exists and belongs to an accessible shop
		var product entities.Product
		err := s.db.Select("shop_id").
			Where("id = ?", stockHistory.ProductID).
			First(&product).Error

		if err != nil {
			// Product doesn't exist - enhanced logging
			log.Printf("ERROR: Stock history %s references non-existent product %s (error: %v)", stockHistory.ID, stockHistory.ProductID, err)
			continue
		}

		productShopID := product.ShopID

		log.Printf("DEBUG: Product %s found in shop %s, checking if shop is accessible", stockHistory.ProductID, productShopID)

		// Enhanced debug logging for shop access check
		isAccessible := accessibleShops[productShopID]
		log.Printf("DEBUG: Shop %s accessible check: %v (accessible shops map: %v)", productShopID, isAccessible, accessibleShops)

		if !isAccessible {
			// Product exists but shop is not accessible
			log.Printf("WARNING: Stock history %s references product %s in inaccessible shop %s (user has access to: %v)",
				stockHistory.ID, stockHistory.ProductID, productShopID, syncContext.AccessibleShopIDs)
			continue
		}

		// Product exists and shop is accessible
		log.Printf("DEBUG: Stock history %s PASSED filtering (product %s in accessible shop %s)",
			stockHistory.ID, stockHistory.ProductID, productShopID)
		filtered = append(filtered, stockHistory)
	}

	log.Printf("SUMMARY: Filtered stock histories for user %s (role: %s): %d → %d",
		syncContext.UserID, syncContext.UserRole, len(stockHistories), len(filtered))
	return filtered
}

// filterTransactionProductsByShopAccess filters transaction products based on accessible shops
func (s *SyncService) filterTransactionProductsByShopAccess(transactionProducts []entities.TransactionProduct, accessibleShops map[uuid.UUID]bool, syncContext dto.SyncContext) []entities.TransactionProduct {
	if syncContext.HasGlobalAccess {
		return transactionProducts
	}

	var filtered []entities.TransactionProduct

	// Enhanced debug logging
	log.Printf("DEBUG: filterTransactionProductsByShopAccess - User %s (role: %s), accessible shops: %v, processing %d transaction products",
		syncContext.UserID, syncContext.UserRole, syncContext.AccessibleShopIDs, len(transactionProducts))

	for i, transactionProduct := range transactionProducts {
		// Enhanced debug logging for each transaction product
		log.Printf("DEBUG: Processing transaction product %d/%d: ID=%s, TransactionID=%s",
			i+1, len(transactionProducts), transactionProduct.ID, transactionProduct.TransactionID)

		// Query to check if the transaction exists and belongs to an accessible shop
		var transaction entities.Transaction
		err := s.db.Select("shop_id").
			Where("id = ?", transactionProduct.TransactionID).
			First(&transaction).Error

		if err != nil {
			// Transaction doesn't exist - enhanced logging
			log.Printf("ERROR: Transaction product %s references non-existent transaction %s (error: %v)", transactionProduct.ID, transactionProduct.TransactionID, err)
			continue
		}

		transactionShopID := transaction.ShopID

		log.Printf("DEBUG: Transaction %s found in shop %s, checking if shop is accessible", transactionProduct.TransactionID, transactionShopID)

		// Enhanced debug logging for shop access check
		isAccessible := accessibleShops[transactionShopID]
		log.Printf("DEBUG: Shop %s accessible check: %v (accessible shops map: %v)", transactionShopID, isAccessible, accessibleShops)

		if !isAccessible {
			// Transaction exists but shop is not accessible
			log.Printf("WARNING: Transaction product %s references transaction %s in inaccessible shop %s (user has access to: %v)",
				transactionProduct.ID, transactionProduct.TransactionID, transactionShopID, syncContext.AccessibleShopIDs)
			continue
		}

		// Transaction exists and shop is accessible
		log.Printf("DEBUG: Transaction product %s PASSED filtering (transaction %s in accessible shop %s)",
			transactionProduct.ID, transactionProduct.TransactionID, transactionShopID)
		filtered = append(filtered, transactionProduct)
	}

	log.Printf("SUMMARY: Filtered transaction products for user %s (role: %s): %d → %d",
		syncContext.UserID, syncContext.UserRole, len(transactionProducts), len(filtered))
	return filtered
}

// validateStockHistoryAccess validates stock history access based on user role and accessible shops
func (s *SyncService) validateStockHistoryAccess(ctx context.Context, stockHistory entities.StockHistory, syncContext dto.SyncContext) bool {
	return s.validateStockHistoryAccessWithSyncData(ctx, stockHistory, syncContext, nil)
}

// validateStockHistoryAccessWithSyncData validates stock history access with sync request product data
func (s *SyncService) validateStockHistoryAccessWithSyncData(ctx context.Context, stockHistory entities.StockHistory, syncContext dto.SyncContext, syncProducts []entities.Product) bool {
	// Global access users (super_admin/admin) can access all stock histories
	if syncContext.HasGlobalAccess {
		log.Printf("DEBUG: validateStockHistoryAccess - User %s has global access, allowing stock history %s", syncContext.UserID, stockHistory.ID)
		return true
	}

	var productShopID uuid.UUID
	var productFound bool

	// First check if product is in the sync request data
	if syncProducts != nil {
		for _, product := range syncProducts {
			if product.ID == stockHistory.ProductID {
				productShopID = product.ShopID
				productFound = true
				log.Printf("DEBUG: validateStockHistoryAccess - Product %s found in sync request data, belongs to shop %s",
					stockHistory.ProductID, productShopID)
				break
			}
		}
	}

	// If not found in sync data, check the database
	if !productFound {
		var product entities.Product
		err := s.db.Select("shop_id").
			Where("id = ?", stockHistory.ProductID).
			First(&product).Error

		if err != nil {
			log.Printf("ERROR: validateStockHistoryAccess - Failed to find product shop for stock history %s (product ID: %s): %v",
				stockHistory.ID, stockHistory.ProductID, err)
			return false
		}

		productShopID = product.ShopID
		log.Printf("DEBUG: validateStockHistoryAccess - Product %s found in database, belongs to shop %s",
			stockHistory.ProductID, productShopID)
	}

	log.Printf("DEBUG: validateStockHistoryAccess - Stock history %s belongs to product %s in shop %s",
		stockHistory.ID, stockHistory.ProductID, productShopID)

	// Check if the shop is accessible to the user with enhanced debugging
	for _, accessibleShopID := range syncContext.AccessibleShopIDs {
		log.Printf("DEBUG: validateStockHistoryAccess - Checking if shop %s matches accessible shop %s", productShopID, accessibleShopID)
		if accessibleShopID == productShopID {
			log.Printf("DEBUG: validateStockHistoryAccess - ALLOWED: Stock history %s in accessible shop %s for user %s (role: %s)",
				stockHistory.ID, productShopID, syncContext.UserID, syncContext.UserRole)
			return true
		}
	}

	log.Printf("WARNING: validateStockHistoryAccess - DENIED: Stock history %s denied: product shop %s not accessible to user %s (role: %s, accessible shops: %v)",
		stockHistory.ID, productShopID, syncContext.UserID, syncContext.UserRole, syncContext.AccessibleShopIDs)
	return false

	// Check if the shop is accessible to the user with enhanced debugging
	for _, accessibleShopID := range syncContext.AccessibleShopIDs {
		log.Printf("DEBUG: validateStockHistoryAccess - Checking if shop %s matches accessible shop %s", productShopID, accessibleShopID)
		if accessibleShopID == productShopID {
			log.Printf("DEBUG: validateStockHistoryAccess - ALLOWED: Stock history %s in accessible shop %s for user %s (role: %s)",
				stockHistory.ID, productShopID, syncContext.UserID, syncContext.UserRole)
			return true
		}
	}

	log.Printf("WARNING: validateStockHistoryAccess - DENIED: Stock history %s denied: product shop %s not accessible to user %s (role: %s, accessible shops: %v)",
		stockHistory.ID, productShopID, syncContext.UserID, syncContext.UserRole, syncContext.AccessibleShopIDs)
	return false
}

// validateTransactionProductAccess validates transaction product access based on user role and accessible shops
func (s *SyncService) validateTransactionProductAccess(ctx context.Context, transactionProduct entities.TransactionProduct, syncContext dto.SyncContext) bool {
	return s.validateTransactionProductAccessWithSyncData(ctx, transactionProduct, syncContext, nil)
}

// validateTransactionProductAccessWithSyncData validates transaction product access with sync request transaction data
func (s *SyncService) validateTransactionProductAccessWithSyncData(ctx context.Context, transactionProduct entities.TransactionProduct, syncContext dto.SyncContext, syncTransactions []entities.Transaction) bool {
	// Global access users (super_admin/admin) can access all transaction products
	if syncContext.HasGlobalAccess {
		log.Printf("DEBUG: validateTransactionProductAccess - User %s has global access, allowing transaction product %s", syncContext.UserID, transactionProduct.ID)
		return true
	}

	var transactionShopID uuid.UUID
	var transactionFound bool

	// First check if transaction is in the sync request data
	if syncTransactions != nil {
		for _, transaction := range syncTransactions {
			if transaction.ID == transactionProduct.TransactionID {
				transactionShopID = transaction.ShopID
				transactionFound = true
				log.Printf("DEBUG: validateTransactionProductAccess - Transaction %s found in sync request data, belongs to shop %s",
					transactionProduct.TransactionID, transactionShopID)
				break
			}
		}
	}

	// If not found in sync data, check the database
	if !transactionFound {
		var transaction entities.Transaction
		err := s.db.Select("shop_id").
			Where("id = ?", transactionProduct.TransactionID).
			First(&transaction).Error

		if err != nil {
			log.Printf("ERROR: validateTransactionProductAccess - Failed to find transaction shop for transaction product %s (transaction ID: %s): %v",
				transactionProduct.ID, transactionProduct.TransactionID, err)
			return false
		}

		transactionShopID = transaction.ShopID
		log.Printf("DEBUG: validateTransactionProductAccess - Transaction %s found in database, belongs to shop %s",
			transactionProduct.TransactionID, transactionShopID)
	}

	log.Printf("DEBUG: validateTransactionProductAccess - Transaction product %s belongs to transaction %s in shop %s",
		transactionProduct.ID, transactionProduct.TransactionID, transactionShopID)

	// Check if the shop is accessible to the user with enhanced debugging
	for _, accessibleShopID := range syncContext.AccessibleShopIDs {
		log.Printf("DEBUG: validateTransactionProductAccess - Checking if shop %s matches accessible shop %s", transactionShopID, accessibleShopID)
		if accessibleShopID == transactionShopID {
			log.Printf("DEBUG: validateTransactionProductAccess - ALLOWED: Transaction product %s in accessible shop %s for user %s (role: %s)",
				transactionProduct.ID, transactionShopID, syncContext.UserID, syncContext.UserRole)
			return true
		}
	}

	log.Printf("WARNING: validateTransactionProductAccess - DENIED: Transaction product %s denied: transaction shop %s not accessible to user %s (role: %s, accessible shops: %v)",
		transactionProduct.ID, transactionShopID, syncContext.UserID, syncContext.UserRole, syncContext.AccessibleShopIDs)
	return false
}

func (s *SyncService) findStockHistoryByID(ctx context.Context, tx *gorm.DB, id uuid.UUID) (*entities.StockHistory, error) {
	var stockHistory entities.StockHistory
	err := tx.First(&stockHistory, "id = ?", id).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &stockHistory, nil
}

func (s *SyncService) createStockHistory(ctx context.Context, tx *gorm.DB, stockHistory entities.StockHistory) error {
	return tx.Create(&stockHistory).Error
}

func (s *SyncService) updateStockHistory(ctx context.Context, tx *gorm.DB, stockHistory entities.StockHistory) error {
	return tx.Save(&stockHistory).Error
}

func (s *SyncService) resolveStockHistoryConflict(existing, incoming entities.StockHistory) *dto.ConflictInfo {
	// Check if there's actually a conflict (different updated_at times)
	if existing.UpdatedAt.Equal(incoming.UpdatedAt) {
		return nil // No conflict
	}

	conflict := &dto.ConflictInfo{
		EntityType:   "stock_history",
		EntityID:     existing.ID,
		ConflictType: "timestamp_mismatch",
		ServerData:   existing,
		ClientData:   incoming,
	}

	// Apply conflict resolution strategy
	switch s.conflictStrategy {
	case dto.LastWriteWins:
		if incoming.UpdatedAt.After(existing.UpdatedAt) {
			conflict.Resolution = "client_wins"
			conflict.Details = fmt.Sprintf("Client version is newer (%s vs %s)", incoming.UpdatedAt, existing.UpdatedAt)
		} else {
			conflict.Resolution = "server_wins"
			conflict.Details = fmt.Sprintf("Server version is newer (%s vs %s)", existing.UpdatedAt, incoming.UpdatedAt)
		}
	case dto.ServerWins:
		conflict.Resolution = "server_wins"
		conflict.Details = "Server always wins conflict resolution strategy"
	case dto.ClientWins:
		conflict.Resolution = "client_wins"
		conflict.Details = "Client always wins conflict resolution strategy"
	}

	return conflict
}

// TransactionProduct sync implementation
func (s *SyncService) pushTransactionProducts(ctx context.Context, tx *gorm.DB, transactionProducts []entities.TransactionProduct, syncContext dto.SyncContext, response *dto.SyncResponse) error {
	for _, transactionProduct := range transactionProducts {
		// Enhanced validation: check both license and shop access
		if !s.validateTransactionProductAccess(ctx, transactionProduct, syncContext) {
			s.addError(response, "transaction_products", transactionProduct.ID, "unauthorized", "Transaction product access denied for user role and shop")
			continue
		}

		// Check if transaction product already exists
		existing, err := s.findTransactionProductByID(ctx, tx, transactionProduct.ID)
		if err != nil {
			s.addError(response, "transaction_products", transactionProduct.ID, "database_error", fmt.Sprintf("Failed to find transaction product: %v", err))
			continue
		}

		s.incrementStat(response.Stats.ProcessedEntities, "transaction_products")

		if existing == nil {
			// Create new transaction product
			if err := s.createTransactionProduct(ctx, tx, transactionProduct); err != nil {
				s.addError(response, "transaction_products", transactionProduct.ID, "create_failed", fmt.Sprintf("Failed to create transaction product: %v", err))
				continue
			}
			s.incrementStat(response.Stats.CreatedEntities, "transaction_products")
		} else {
			// Check for conflicts and resolve
			if conflict := s.resolveTransactionProductConflict(*existing, transactionProduct); conflict != nil {
				response.Conflicts = append(response.Conflicts, *conflict)
			}

			// Update transaction product
			if err := s.updateTransactionProduct(ctx, tx, transactionProduct); err != nil {
				s.addError(response, "transaction_products", transactionProduct.ID, "update_failed", fmt.Sprintf("Failed to update transaction product: %v", err))
				continue
			}
			s.incrementStat(response.Stats.UpdatedEntities, "transaction_products")
		}
	}
	return nil
}

func (s *SyncService) pullTransactionProducts(ctx context.Context, tx *gorm.DB, lastSync time.Time, licenseID uuid.UUID, response *dto.SyncResponse) error {
	// Select only the transaction_product fields without relations to avoid unnecessary data
	var transactionProducts []entities.TransactionProduct
	err := tx.WithContext(ctx).
		Select("transaction_products.*").
		Table("transaction_products").
		Joins("JOIN transactions ON transaction_products.transaction_id = transactions.id").
		Joins("JOIN shops ON transactions.shop_id = shops.id").
		Where("shops.license_id = ? AND transaction_products.updated_at > ?", licenseID, lastSync).
		Find(&transactionProducts).Error
	if err != nil {
		return fmt.Errorf("failed to fetch transaction products: %w", err)
	}

	// Map entities to DTOs to exclude relationship fields
	response.TransactionProducts = dto.MapTransactionProductsToSyncDTOs(transactionProducts)
	return nil
}

func (s *SyncService) validateTransactionProductLicense(ctx context.Context, transactionProduct entities.TransactionProduct, licenseID uuid.UUID) bool {
	var count int64
	err := s.db.Model(&entities.Transaction{}).
		Joins("JOIN shops ON transactions.shop_id = shops.id").
		Where("transactions.id = ? AND shops.license_id = ?", transactionProduct.TransactionID, licenseID).
		Count(&count).Error
	return err == nil && count > 0
}

func (s *SyncService) findTransactionProductByID(ctx context.Context, tx *gorm.DB, id uuid.UUID) (*entities.TransactionProduct, error) {
	var transactionProduct entities.TransactionProduct
	err := tx.First(&transactionProduct, "id = ?", id).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &transactionProduct, nil
}

func (s *SyncService) createTransactionProduct(ctx context.Context, tx *gorm.DB, transactionProduct entities.TransactionProduct) error {
	return tx.Create(&transactionProduct).Error
}

func (s *SyncService) updateTransactionProduct(ctx context.Context, tx *gorm.DB, transactionProduct entities.TransactionProduct) error {
	return tx.Save(&transactionProduct).Error
}

func (s *SyncService) resolveTransactionProductConflict(existing, incoming entities.TransactionProduct) *dto.ConflictInfo {
	// Check if there's actually a conflict (different updated_at times)
	if existing.UpdatedAt.Equal(incoming.UpdatedAt) {
		return nil // No conflict
	}

	conflict := &dto.ConflictInfo{
		EntityType:   "transaction_product",
		EntityID:     existing.ID,
		ConflictType: "timestamp_mismatch",
		ServerData:   existing,
		ClientData:   incoming,
	}

	// Apply conflict resolution strategy
	switch s.conflictStrategy {
	case dto.LastWriteWins:
		if incoming.UpdatedAt.After(existing.UpdatedAt) {
			conflict.Resolution = "client_wins"
			conflict.Details = fmt.Sprintf("Client version is newer (%s vs %s)", incoming.UpdatedAt, existing.UpdatedAt)
		} else {
			conflict.Resolution = "server_wins"
			conflict.Details = fmt.Sprintf("Server version is newer (%s vs %s)", existing.UpdatedAt, incoming.UpdatedAt)
		}
	case dto.ServerWins:
		conflict.Resolution = "server_wins"
		conflict.Details = "Server always wins conflict resolution strategy"
	case dto.ClientWins:
		conflict.Resolution = "client_wins"
		conflict.Details = "Client always wins conflict resolution strategy"
	}

	return conflict
}

// User sync implementation
func (s *SyncService) pushUsers(ctx context.Context, tx *gorm.DB, users []entities.User, licenseID uuid.UUID, response *dto.SyncResponse) error {
	for _, user := range users {
		// Validate user belongs to license
		if !s.validateUserLicense(ctx, user, licenseID) {
			s.addError(response, "users", user.ID, "unauthorized", "User does not belong to license")
			continue
		}

		// Check if user already exists
		existing, err := s.findUserByID(ctx, tx, user.ID)
		if err != nil {
			s.addError(response, "users", user.ID, "database_error", fmt.Sprintf("Failed to find user: %v", err))
			continue
		}

		s.incrementStat(response.Stats.ProcessedEntities, "users")

		if existing == nil {
			// Create new user
			if err := s.createUser(ctx, tx, user); err != nil {
				s.addError(response, "users", user.ID, "create_failed", fmt.Sprintf("Failed to create user: %v", err))
				continue
			}
			s.incrementStat(response.Stats.CreatedEntities, "users")
		} else {
			// Check for conflicts and resolve
			if conflict := s.resolveUserConflict(*existing, user); conflict != nil {
				response.Conflicts = append(response.Conflicts, *conflict)
			}

			// Update user (be careful with sensitive fields like password)
			if err := s.updateUser(ctx, tx, user); err != nil {
				s.addError(response, "users", user.ID, "update_failed", fmt.Sprintf("Failed to update user: %v", err))
				continue
			}
			s.incrementStat(response.Stats.UpdatedEntities, "users")
		}
	}
	return nil
}

func (s *SyncService) pullUsers(ctx context.Context, tx *gorm.DB, lastSync time.Time, licenseID uuid.UUID, response *dto.SyncResponse) error {
	// Select only the user fields without relations to avoid unnecessary data
	var users []entities.User
	err := tx.WithContext(ctx).
		Select("users.*").
		Where("license_id = ? AND updated_at > ?", licenseID, lastSync).
		Find(&users).Error
	if err != nil {
		return fmt.Errorf("failed to fetch users: %w", err)
	}

	// Map entities to DTOs to exclude relationship fields
	response.Users = dto.MapUsersToSyncDTOs(users)
	return nil
}

func (s *SyncService) validateUserLicense(ctx context.Context, user entities.User, licenseID uuid.UUID) bool {
	return user.LicenseID != nil && *user.LicenseID == licenseID
}

func (s *SyncService) findUserByID(ctx context.Context, tx *gorm.DB, id uuid.UUID) (*entities.User, error) {
	var user entities.User
	err := tx.First(&user, "id = ?", id).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

func (s *SyncService) createUser(ctx context.Context, tx *gorm.DB, user entities.User) error {
	return tx.Create(&user).Error
}

func (s *SyncService) updateUser(ctx context.Context, tx *gorm.DB, user entities.User) error {
	return tx.Save(&user).Error
}

func (s *SyncService) resolveUserConflict(existing, incoming entities.User) *dto.ConflictInfo {
	// Check if there's actually a conflict (different updated_at times)
	if existing.UpdatedAt.Equal(incoming.UpdatedAt) {
		return nil // No conflict
	}

	conflict := &dto.ConflictInfo{
		EntityType:   "user",
		EntityID:     existing.ID,
		ConflictType: "timestamp_mismatch",
		ServerData:   existing,
		ClientData:   incoming,
	}

	// Apply conflict resolution strategy
	switch s.conflictStrategy {
	case dto.LastWriteWins:
		if incoming.UpdatedAt.After(existing.UpdatedAt) {
			conflict.Resolution = "client_wins"
			conflict.Details = fmt.Sprintf("Client version is newer (%s vs %s)", incoming.UpdatedAt, existing.UpdatedAt)
		} else {
			conflict.Resolution = "server_wins"
			conflict.Details = fmt.Sprintf("Server version is newer (%s vs %s)", existing.UpdatedAt, incoming.UpdatedAt)
		}
	case dto.ServerWins:
		conflict.Resolution = "server_wins"
		conflict.Details = "Server always wins conflict resolution strategy"
	case dto.ClientWins:
		conflict.Resolution = "client_wins"
		conflict.Details = "Client always wins conflict resolution strategy"
	}

	return conflict
}

// cleanupResponse removes empty arrays and cleans up conflict data to remove nested empty relations
func (s *SyncService) cleanupResponse(response *dto.SyncResponse) {
	// Clean up conflicts by removing nested empty relations
	for i, conflict := range response.Conflicts {
		response.Conflicts[i].ServerData = s.cleanupEntityData(conflict.ServerData)
		response.Conflicts[i].ClientData = s.cleanupEntityData(conflict.ClientData)
	}

	// Remove empty slices - set to nil so omitempty works
	if len(response.Carts) == 0 {
		response.Carts = nil
	}
	if len(response.Categories) == 0 {
		response.Categories = nil
	}
	if len(response.Expenses) == 0 {
		response.Expenses = nil
	}
	if len(response.Histories) == 0 {
		response.Histories = nil
	}
	if len(response.Payments) == 0 {
		response.Payments = nil
	}
	if len(response.Products) == 0 {
		response.Products = nil
	}
	if len(response.Receipts) == 0 {
		response.Receipts = nil
	}
	if len(response.Shops) == 0 {
		response.Shops = nil
	}
	if len(response.StockHistories) == 0 {
		response.StockHistories = nil
	}
	if len(response.TransactionProducts) == 0 {
		response.TransactionProducts = nil
	}
	if len(response.Transactions) == 0 {
		response.Transactions = nil
	}
	if len(response.Users) == 0 {
		response.Users = nil
	}
	if len(response.Conflicts) == 0 {
		response.Conflicts = nil
	}
	if len(response.Errors) == 0 {
		response.Errors = nil
	}
}

// cleanupEntityData removes nested empty relations from entity data
func (s *SyncService) cleanupEntityData(data interface{}) interface{} {
	if data == nil {
		return nil
	}

	// Convert to map for easier manipulation
	dataMap := make(map[string]interface{})

	// Handle different entity types and create clean data
	switch entity := data.(type) {
	case entities.Category:
		dataMap["id"] = entity.ID
		dataMap["shop_id"] = entity.ShopID
		dataMap["name"] = entity.Name
		dataMap["created_at"] = entity.CreatedAt
		dataMap["updated_at"] = entity.UpdatedAt
		return dataMap

	case entities.Product:
		dataMap["id"] = entity.ID
		dataMap["shop_id"] = entity.ShopID
		if entity.CatID != nil {
			dataMap["cat_id"] = *entity.CatID
		}
		dataMap["name"] = entity.Name
		if entity.Barcode != nil {
			dataMap["barcode"] = *entity.Barcode
		}
		dataMap["sale"] = entity.Sale
		dataMap["buy"] = entity.Buy
		dataMap["stock"] = entity.Stock
		dataMap["created_at"] = entity.CreatedAt
		dataMap["updated_at"] = entity.UpdatedAt
		return dataMap

	case entities.Cart:
		dataMap["id"] = entity.ID
		dataMap["shop_id"] = entity.ShopID
		dataMap["product_id"] = entity.ProductID
		dataMap["user_id"] = entity.UserID
		dataMap["quantity"] = entity.Quantity
		dataMap["created_at"] = entity.CreatedAt
		dataMap["updated_at"] = entity.UpdatedAt
		return dataMap

	case entities.Transaction:
		dataMap["id"] = entity.ID
		dataMap["shop_id"] = entity.ShopID
		dataMap["cashier_id"] = entity.CashierID
		if entity.CustomerName != nil {
			dataMap["customer_name"] = *entity.CustomerName
		}
		dataMap["discount"] = entity.Discount
		dataMap["discount_percentage"] = entity.DiscountPercentage
		dataMap["additional_cost"] = entity.AdditionalCost
		dataMap["status"] = entity.Status
		dataMap["total_price"] = entity.TotalPrice
		dataMap["created_at"] = entity.CreatedAt
		dataMap["updated_at"] = entity.UpdatedAt
		return dataMap

	case entities.Payment:
		dataMap["id"] = entity.ID
		dataMap["shop_id"] = entity.ShopID
		if entity.UserID != nil {
			dataMap["user_id"] = *entity.UserID
		}
		dataMap["transaction_id"] = entity.TransactionID
		dataMap["status"] = entity.Status
		dataMap["total"] = entity.Total
		dataMap["created_at"] = entity.CreatedAt
		dataMap["updated_at"] = entity.UpdatedAt
		return dataMap

	case entities.Expense:
		dataMap["id"] = entity.ID
		dataMap["shop_id"] = entity.ShopID
		dataMap["created_at"] = entity.CreatedAt
		dataMap["updated_at"] = entity.UpdatedAt
		return dataMap

	case entities.Receipt:
		dataMap["id"] = entity.ID
		dataMap["shop_id"] = entity.ShopID
		dataMap["created_at"] = entity.CreatedAt
		dataMap["updated_at"] = entity.UpdatedAt
		return dataMap

	case entities.History:
		dataMap["id"] = entity.ID
		dataMap["shop_id"] = entity.ShopID
		dataMap["created_at"] = entity.CreatedAt
		dataMap["updated_at"] = entity.UpdatedAt
		return dataMap

	case entities.StockHistory:
		dataMap["id"] = entity.ID
		dataMap["product_id"] = entity.ProductID
		dataMap["created_at"] = entity.CreatedAt
		dataMap["updated_at"] = entity.UpdatedAt
		return dataMap

	case entities.TransactionProduct:
		dataMap["id"] = entity.ID
		dataMap["transaction_id"] = entity.TransactionID
		dataMap["product_id"] = entity.ProductID
		dataMap["created_at"] = entity.CreatedAt
		dataMap["updated_at"] = entity.UpdatedAt
		return dataMap

	case entities.Shop:
		dataMap["id"] = entity.ID
		dataMap["license_id"] = entity.LicenseID
		dataMap["user_id"] = entity.UserID
		dataMap["name"] = entity.Name
		dataMap["domain"] = entity.Domain
		dataMap["created_at"] = entity.CreatedAt
		dataMap["updated_at"] = entity.UpdatedAt
		return dataMap

	case entities.User:
		dataMap["id"] = entity.ID
		if entity.LicenseID != nil {
			dataMap["license_id"] = *entity.LicenseID
		}
		dataMap["name"] = entity.Name
		dataMap["created_at"] = entity.CreatedAt
		dataMap["updated_at"] = entity.UpdatedAt
		return dataMap

	default:
		// For unknown types, try to extract basic fields if it's a map
		if entityMap, ok := data.(map[string]interface{}); ok {
			cleanMap := make(map[string]interface{})
			// Copy only non-nested fields
			for key, value := range entityMap {
				if key != "shop" && key != "product" && key != "user" && key != "category" &&
					key != "transaction" && key != "license" && key != "owner" && key != "cashier" &&
					key != "products" && key != "carts" && key != "transaction_products" &&
					key != "stock_histories" && key != "payments" && key != "receipts" && key != "histories" {
					cleanMap[key] = value
				}
			}
			return cleanMap
		}
		// Return as is for unknown types
		return data
	}
}

// Retry mechanism and enhanced error handling utilities

// retryOperation executes an operation with exponential backoff retry logic
func (s *SyncService) retryOperation(ctx context.Context, operation func() error, maxRetries int, baseDelay time.Duration, operationName string) error {
	var lastErr error

	for attempt := 0; attempt <= maxRetries; attempt++ {
		// Check context cancellation
		select {
		case <-ctx.Done():
			return fmt.Errorf("operation cancelled: %w", ctx.Err())
		default:
		}

		// Execute the operation
		err := operation()
		if err == nil {
			// Success
			if attempt > 0 {
				log.Printf("Operation %s succeeded after %d retries", operationName, attempt)
			}
			return nil
		}

		lastErr = err

		// Check if error is retryable
		if !s.isRetryableError(err) {
			log.Printf("Non-retryable error in operation %s: %v", operationName, err)
			return err
		}

		// Do not retry on the last attempt
		if attempt == maxRetries {
			break
		}

		// Calculate delay with exponential backoff
		delay := baseDelay * time.Duration(1<<uint(attempt))
		log.Printf("Operation %s failed (attempt %d/%d), retrying in %v: %v",
			operationName, attempt+1, maxRetries+1, delay, err)

		// Wait before retry
		select {
		case <-ctx.Done():
			return fmt.Errorf("operation cancelled during retry delay: %w", ctx.Err())
		case <-time.After(delay):
		}
	}

	return fmt.Errorf("operation %s failed after %d retries: %w", operationName, maxRetries+1, lastErr)
}

// isRetryableError determines if an error is worth retrying
func (s *SyncService) isRetryableError(err error) bool {
	if err == nil {
		return false
	}

	errStr := strings.ToLower(err.Error())

	// Database connection and timeout errors are retryable
	retryablePatterns := []string{
		"connection refused",
		"connection reset",
		"timeout",
		"deadlock",
		"database is locked",
		"too many connections",
		"server is shutting down",
		"context deadline exceeded",
	}

	for _, pattern := range retryablePatterns {
		if strings.Contains(errStr, pattern) {
			return true
		}
	}

	return false
}

// addDetailedError adds a detailed error with context information
func (s *SyncService) addDetailedError(response *dto.SyncResponse, entityType string, entityID uuid.UUID, errorCode string, message string, details map[string]interface{}) {
	errorDetails := ""
	if len(details) > 0 {
		errorDetails = fmt.Sprintf("Details: %+v", details)
	}

	syncError := dto.SyncError{
		EntityType: entityType,
		EntityID:   entityID,
		ErrorCode:  errorCode,
		Message:    message,
		Details:    errorDetails,
	}

	response.Errors = append(response.Errors, syncError)
	log.Printf("Sync error - Type: %s, ID: %s, Code: %s, Message: %s, Details: %s",
		entityType, entityID, errorCode, message, errorDetails)
}

// addDetailedValidationError adds comprehensive validation errors to the sync response
func (s *SyncService) addDetailedValidationError(response *dto.SyncResponse, entityType string, entityID uuid.UUID, errorCode string, validationErr error, details map[string]interface{}) {
	// Handle different types of validation errors
	if validationErrors, ok := validationErr.(validators.ValidationErrors); ok {
		// Multiple validation errors
		for i, ve := range validationErrors {
			errorDetails := fmt.Sprintf("Field: %s, Code: %s, Details: %+v", ve.Field, ve.Code, details)
			syncError := dto.SyncError{
				EntityType: entityType,
				EntityID:   entityID,
				ErrorCode:  fmt.Sprintf("%s_%d", errorCode, i+1),
				Message:    ve.Message,
				Details:    errorDetails,
			}
			response.Errors = append(response.Errors, syncError)

			log.Printf("Validation error - Type: %s, ID: %s, Field: %s, Code: %s, Message: %s",
				entityType, entityID, ve.Field, ve.Code, ve.Message)
		}
	} else {
		// Single validation error
		errorDetails := ""
		if len(details) > 0 {
			errorDetails = fmt.Sprintf("Details: %+v", details)
		}

		syncError := dto.SyncError{
			EntityType: entityType,
			EntityID:   entityID,
			ErrorCode:  errorCode,
			Message:    validationErr.Error(),
			Details:    errorDetails,
		}
		response.Errors = append(response.Errors, syncError)

		log.Printf("Validation error - Type: %s, ID: %s, Message: %s, Details: %s",
			entityType, entityID, validationErr.Error(), errorDetails)
	}
}

// CRITICAL FIX: Enhanced error handling with configurable policies
// handleEntityError processes errors according to the configured error policy
func (s *SyncService) handleEntityError(err error, entityType string, entityID uuid.UUID, response *dto.SyncResponse, details map[string]interface{}) error {
	// Parse the error policy from configuration
	errorPolicy := dto.ParseSyncErrorPolicy(s.config.ErrorPolicy)

	// Add the error to the response for tracking
	s.addDetailedError(response, entityType, entityID, "processing_error", err.Error(), details)

	// Check if we've exceeded the maximum allowed errors
	if len(response.Errors) > s.config.MaxEntityErrorsPerSync {
		log.Printf("Maximum entity errors exceeded (%d), aborting sync", s.config.MaxEntityErrorsPerSync)
		return fmt.Errorf("too many entity errors during sync: %d errors", len(response.Errors))
	}

	// Handle based on policy
	switch errorPolicy {
	case dto.ContinueOnError:
		// Log and continue - this is the current behavior
		log.Printf("Continuing sync despite error in %s %s: %v", entityType, entityID, err)
		return nil

	case dto.AbortOnError:
		// Stop processing immediately
		log.Printf("Aborting sync due to error in %s %s: %v", entityType, entityID, err)
		return fmt.Errorf("sync aborted due to error policy: %w", err)

	case dto.RetryOnError:
		// For now, just log and continue - retry logic would require more complex state management
		// This could be enhanced in future versions with proper retry queues
		log.Printf("Entity error (retry policy not yet implemented, continuing): %s %s: %v", entityType, entityID, err)
		return nil

	default:
		// Default to continue
		log.Printf("Unknown error policy, continuing: %s %s: %v", entityType, entityID, err)
		return nil
	}
}

// CRITICAL FIX: Savepoint-based transaction processing for better error isolation
// processEntityWithSavepoint processes an entity within a savepoint for isolated error handling
func (s *SyncService) processEntityWithSavepoint(ctx context.Context, tx *gorm.DB, entityID uuid.UUID, processFunc func() error) error {
	// Create savepoint name
	savepointName := fmt.Sprintf("sp_%s", entityID.String()[:8])

	// Create savepoint
	if err := tx.SavePoint(savepointName).Error; err != nil {
		log.Printf("Warning: Failed to create savepoint %s, proceeding without isolation: %v", savepointName, err)
		// If savepoint creation fails, proceed without it - this maintains backward compatibility
		return processFunc()
	}

	// Execute the processing function
	if err := processFunc(); err != nil {
		// Rollback to savepoint on error
		if rollbackErr := tx.RollbackTo(savepointName).Error; rollbackErr != nil {
			log.Printf("Warning: Failed to rollback to savepoint %s: %v", savepointName, rollbackErr)
		}
		return err
	}

	// Release savepoint on success (optional, but good practice)
	if err := tx.Exec(fmt.Sprintf("RELEASE SAVEPOINT %s", savepointName)).Error; err != nil {
		log.Printf("Warning: Failed to release savepoint %s: %v", savepointName, err)
		// This is not critical - the savepoint will be cleaned up when the transaction ends
	}

	return nil
}

// logPerformanceMetrics logs performance metrics for monitoring
func (s *SyncService) logPerformanceMetrics(entityType string, count int, duration time.Duration, operation string) {
	rate := float64(count) / duration.Seconds()
	log.Printf("Performance - %s %s: %d entities in %v (%.2f entities/sec)",
		operation, entityType, count, duration, rate)

	// Log warning if performance is below expected thresholds
	minRatePerSecond := 10.0 // Expected minimum processing rate
	if rate < minRatePerSecond && count > 10 {
		log.Printf("WARNING: Low performance detected for %s %s: %.2f entities/sec (expected > %.2f)",
			operation, entityType, rate, minRatePerSecond)
	}
}

// Safe versions of push methods that avoid nested transactions and provide enhanced error handling

// pushCategoriesSafe handles category synchronization safely
func (s *SyncService) pushCategoriesSafe(ctx context.Context, tx *gorm.DB, categories []entities.Category, licenseID uuid.UUID, response *dto.SyncResponse) error {
	for i, category := range categories {
		if err := s.processSingleCategorySafe(ctx, tx, category, licenseID, response); err != nil {
			log.Printf("Error processing category %s (index %d): %v", category.ID, i, err)
			s.addDetailedError(response, "categories", category.ID, "processing_failed", err.Error(),
				map[string]interface{}{"category_shop_id": category.ShopID, "license_id": licenseID, "index": i})
			continue
		}
	}
	return nil
}

// pushProductsSafe handles product synchronization safely
func (s *SyncService) pushProductsSafe(ctx context.Context, tx *gorm.DB, products []entities.Product, licenseID uuid.UUID, response *dto.SyncResponse) error {
	for i, product := range products {
		if err := s.processSingleProductSafe(ctx, tx, product, licenseID, response); err != nil {
			log.Printf("Error processing product %s (index %d): %v", product.ID, i, err)
			s.addDetailedError(response, "products", product.ID, "processing_failed", err.Error(),
				map[string]interface{}{"product_shop_id": product.ShopID, "license_id": licenseID, "index": i})
			continue
		}
	}
	return nil
}

// pushTransactionsSafe handles transaction synchronization safely
func (s *SyncService) pushTransactionsSafe(ctx context.Context, tx *gorm.DB, transactions []entities.Transaction, licenseID uuid.UUID, response *dto.SyncResponse) error {
	for i, transaction := range transactions {
		if err := s.processSingleTransactionSafe(ctx, tx, transaction, licenseID, response); err != nil {
			log.Printf("Error processing transaction %s (index %d): %v", transaction.ID, i, err)
			s.addDetailedError(response, "transactions", transaction.ID, "processing_failed", err.Error(),
				map[string]interface{}{"transaction_shop_id": transaction.ShopID, "license_id": licenseID, "index": i})
			continue
		}
	}
	return nil
}

// pushExpensesSafe handles expense synchronization safely
func (s *SyncService) pushExpensesSafe(ctx context.Context, tx *gorm.DB, expenses []entities.Expense, licenseID uuid.UUID, response *dto.SyncResponse) error {
	for i, expense := range expenses {
		if err := s.processSingleExpenseSafe(ctx, tx, expense, licenseID, response); err != nil {
			log.Printf("Error processing expense %s (index %d): %v", expense.ID, i, err)
			s.addDetailedError(response, "expenses", expense.ID, "processing_failed", err.Error(),
				map[string]interface{}{"expense_shop_id": expense.ShopID, "license_id": licenseID, "index": i})
			continue
		}
	}
	return nil
}

// pushPaymentsSafe handles payment synchronization safely
func (s *SyncService) pushPaymentsSafe(ctx context.Context, tx *gorm.DB, payments []entities.Payment, licenseID uuid.UUID, response *dto.SyncResponse) error {
	for i, payment := range payments {
		if err := s.processSinglePaymentSafe(ctx, tx, payment, licenseID, response); err != nil {
			log.Printf("Error processing payment %s (index %d): %v", payment.ID, i, err)
			s.addDetailedError(response, "payments", payment.ID, "processing_failed", err.Error(),
				map[string]interface{}{"payment_shop_id": payment.ShopID, "license_id": licenseID, "index": i})
			continue
		}
	}
	return nil
}

// pushReceiptsSafe handles receipt synchronization safely
func (s *SyncService) pushReceiptsSafe(ctx context.Context, tx *gorm.DB, receipts []entities.Receipt, licenseID uuid.UUID, response *dto.SyncResponse) error {
	for i, receipt := range receipts {
		if err := s.processSingleReceiptSafe(ctx, tx, receipt, licenseID, response); err != nil {
			log.Printf("Error processing receipt %s (index %d): %v", receipt.ID, i, err)
			s.addDetailedError(response, "receipts", receipt.ID, "processing_failed", err.Error(),
				map[string]interface{}{"receipt_shop_id": receipt.ShopID, "license_id": licenseID, "index": i})
			continue
		}
	}
	return nil
}

// pushHistoriesSafe handles history synchronization safely
func (s *SyncService) pushHistoriesSafe(ctx context.Context, tx *gorm.DB, histories []entities.History, licenseID uuid.UUID, response *dto.SyncResponse) error {
	for i, history := range histories {
		if err := s.processSingleHistorySafe(ctx, tx, history, licenseID, response); err != nil {
			log.Printf("Error processing history %s (index %d): %v", history.ID, i, err)
			s.addDetailedError(response, "histories", history.ID, "processing_failed", err.Error(),
				map[string]interface{}{"history_shop_id": history.ShopID, "license_id": licenseID, "index": i})
			continue
		}
	}
	return nil
}

// pushShopsSafe handles shop synchronization safely
func (s *SyncService) pushShopsSafe(ctx context.Context, tx *gorm.DB, shops []entities.Shop, licenseID uuid.UUID, response *dto.SyncResponse) error {
	log.Printf("DEBUG: pushShopsSafe - Processing %d shops for license %s", len(shops), licenseID)

	if len(shops) == 0 {
		log.Printf("DEBUG: pushShopsSafe - No shops to process")
		return nil
	}

	for i, shop := range shops {
		log.Printf("DEBUG: pushShopsSafe - Processing shop %d: ID=%s, Name=%s, LicenseID=%s", i, shop.ID, shop.Name, shop.LicenseID)
		if err := s.processSingleShopSafe(ctx, tx, shop, licenseID, response); err != nil {
			log.Printf("ERROR: Processing shop %s (index %d): %v", shop.ID, i, err)
			s.addDetailedError(response, "shops", shop.ID, "processing_failed", err.Error(),
				map[string]interface{}{"shop_license_id": shop.LicenseID, "license_id": licenseID, "index": i})
			continue
		}
		log.Printf("DEBUG: pushShopsSafe - Successfully processed shop %s", shop.ID)
	}

	log.Printf("DEBUG: pushShopsSafe - Completed processing %d shops", len(shops))
	return nil
}

// pushUsersSafe handles user synchronization safely with role validation
func (s *SyncService) pushUsersSafe(ctx context.Context, tx *gorm.DB, users []entities.User, licenseID uuid.UUID, response *dto.SyncResponse) error {
	for i, user := range users {
		if err := s.processSingleUserSafe(ctx, tx, user, licenseID, response); err != nil {
			log.Printf("Error processing user %s (index %d): %v", user.ID, i, err)
			s.addDetailedError(response, "users", user.ID, "processing_failed", err.Error(),
				map[string]interface{}{"user_license_id": user.LicenseID, "license_id": licenseID, "index": i})
			continue
		}
	}
	return nil
}

// pushStockHistoriesSafe handles stock history synchronization safely
func (s *SyncService) pushStockHistoriesSafe(ctx context.Context, tx *gorm.DB, stockHistories []entities.StockHistory, syncProducts []entities.Product, syncContext dto.SyncContext, response *dto.SyncResponse) error {
	for i, stockHistory := range stockHistories {
		if err := s.processSingleStockHistorySafe(ctx, tx, stockHistory, syncProducts, syncContext, response); err != nil {
			log.Printf("Error processing stock history %s (index %d): %v", stockHistory.ID, i, err)
			s.addDetailedError(response, "stock_histories", stockHistory.ID, "processing_failed", err.Error(),
				map[string]interface{}{"product_id": stockHistory.ProductID, "license_id": syncContext.LicenseID, "index": i})
			continue
		}
	}
	return nil
}

// pushTransactionProductsSafe handles transaction product synchronization safely
func (s *SyncService) pushTransactionProductsSafe(ctx context.Context, tx *gorm.DB, transactionProducts []entities.TransactionProduct, syncTransactions []entities.Transaction, syncContext dto.SyncContext, response *dto.SyncResponse) error {
	for i, transactionProduct := range transactionProducts {
		if err := s.processSingleTransactionProductSafe(ctx, tx, transactionProduct, syncTransactions, syncContext, response); err != nil {
			log.Printf("Error processing transaction product %s (index %d): %v", transactionProduct.ID, i, err)
			s.addDetailedError(response, "transaction_products", transactionProduct.ID, "processing_failed", err.Error(),
				map[string]interface{}{"transaction_id": transactionProduct.TransactionID, "license_id": syncContext.LicenseID, "index": i})
			continue
		}
	}
	return nil
}

// Safe processing methods for individual entities

// processSingleCategorySafe processes a single category entity safely
func (s *SyncService) processSingleCategorySafe(ctx context.Context, tx *gorm.DB, category entities.Category, licenseID uuid.UUID, response *dto.SyncResponse) error {
	if !s.validateCategoryLicense(ctx, tx, category, licenseID) {
		return fmt.Errorf("category does not belong to license %s", licenseID)
	}

	existing, err := s.findCategoryByID(ctx, tx, category.ID)
	if err != nil {
		return fmt.Errorf("failed to find category: %w", err)
	}

	if existing == nil {
		if err := s.createCategory(ctx, tx, category); err != nil {
			return fmt.Errorf("failed to create category: %w", err)
		}
		s.incrementStat(response.Stats.CreatedEntities, "categories")
	} else {
		if conflict := s.resolveCategoryConflict(*existing, category); conflict != nil {
			response.Conflicts = append(response.Conflicts, *conflict)
			if existing.UpdatedAt.After(category.UpdatedAt) {
				s.incrementStat(response.Stats.ProcessedEntities, "categories")
				return nil
			}
		}
		if err := s.updateCategory(ctx, tx, category); err != nil {
			return fmt.Errorf("failed to update category: %w", err)
		}
		s.incrementStat(response.Stats.UpdatedEntities, "categories")
	}

	s.incrementStat(response.Stats.ProcessedEntities, "categories")
	return nil
}

// processSingleProductSafe processes a single product entity safely
func (s *SyncService) processSingleProductSafe(ctx context.Context, tx *gorm.DB, product entities.Product, licenseID uuid.UUID, response *dto.SyncResponse) error {
	if !s.validateProductLicense(ctx, product, licenseID) {
		return fmt.Errorf("product does not belong to license %s", licenseID)
	}

	existing, err := s.findProductByID(ctx, tx, product.ID)
	if err != nil {
		return fmt.Errorf("failed to find product: %w", err)
	}

	if existing == nil {
		if err := s.createProduct(ctx, tx, product); err != nil {
			return fmt.Errorf("failed to create product: %w", err)
		}
		s.incrementStat(response.Stats.CreatedEntities, "products")
	} else {
		if conflict := s.resolveProductConflict(*existing, product); conflict != nil {
			response.Conflicts = append(response.Conflicts, *conflict)
			if existing.UpdatedAt.After(product.UpdatedAt) {
				s.incrementStat(response.Stats.ProcessedEntities, "products")
				return nil
			}
		}
		if err := s.updateProduct(ctx, tx, product); err != nil {
			return fmt.Errorf("failed to update product: %w", err)
		}
		s.incrementStat(response.Stats.UpdatedEntities, "products")
	}

	s.incrementStat(response.Stats.ProcessedEntities, "products")
	return nil
}

// processSingleTransactionSafe processes a single transaction entity safely
func (s *SyncService) processSingleTransactionSafe(ctx context.Context, tx *gorm.DB, transaction entities.Transaction, licenseID uuid.UUID, response *dto.SyncResponse) error {
	if !s.validateTransactionLicense(ctx, transaction, licenseID) {
		return fmt.Errorf("transaction does not belong to license %s", licenseID)
	}

	existing, err := s.findTransactionByID(ctx, tx, transaction.ID)
	if err != nil {
		return fmt.Errorf("failed to find transaction: %w", err)
	}

	if existing == nil {
		if err := s.createTransaction(ctx, tx, transaction); err != nil {
			return fmt.Errorf("failed to create transaction: %w", err)
		}
		s.incrementStat(response.Stats.CreatedEntities, "transactions")
	} else {
		if conflict := s.resolveTransactionConflict(*existing, transaction); conflict != nil {
			response.Conflicts = append(response.Conflicts, *conflict)
			if existing.UpdatedAt.After(transaction.UpdatedAt) {
				s.incrementStat(response.Stats.ProcessedEntities, "transactions")
				return nil
			}
		}
		if err := s.updateTransaction(ctx, tx, transaction); err != nil {
			return fmt.Errorf("failed to update transaction: %w", err)
		}
		s.incrementStat(response.Stats.UpdatedEntities, "transactions")
	}

	s.incrementStat(response.Stats.ProcessedEntities, "transactions")
	return nil
}

// processSingleExpenseSafe processes a single expense entity safely
func (s *SyncService) processSingleExpenseSafe(ctx context.Context, tx *gorm.DB, expense entities.Expense, licenseID uuid.UUID, response *dto.SyncResponse) error {
	if !s.validateExpenseLicense(ctx, expense, licenseID) {
		return fmt.Errorf("expense does not belong to license %s", licenseID)
	}

	existing, err := s.findExpenseByID(ctx, tx, expense.ID)
	if err != nil {
		return fmt.Errorf("failed to find expense: %w", err)
	}

	if existing == nil {
		if err := s.createExpense(ctx, tx, expense); err != nil {
			return fmt.Errorf("failed to create expense: %w", err)
		}
		s.incrementStat(response.Stats.CreatedEntities, "expenses")
	} else {
		if conflict := s.resolveExpenseConflict(*existing, expense); conflict != nil {
			response.Conflicts = append(response.Conflicts, *conflict)
			if existing.UpdatedAt.After(expense.UpdatedAt) {
				s.incrementStat(response.Stats.ProcessedEntities, "expenses")
				return nil
			}
		}
		if err := s.updateExpense(ctx, tx, expense); err != nil {
			return fmt.Errorf("failed to update expense: %w", err)
		}
		s.incrementStat(response.Stats.UpdatedEntities, "expenses")
	}

	s.incrementStat(response.Stats.ProcessedEntities, "expenses")
	return nil
}

// processSinglePaymentSafe processes a single payment entity safely
func (s *SyncService) processSinglePaymentSafe(ctx context.Context, tx *gorm.DB, payment entities.Payment, licenseID uuid.UUID, response *dto.SyncResponse) error {
	if !s.validatePaymentLicense(ctx, payment, licenseID) {
		return fmt.Errorf("payment does not belong to license %s", licenseID)
	}

	existing, err := s.findPaymentByID(ctx, tx, payment.ID)
	if err != nil {
		return fmt.Errorf("failed to find payment: %w", err)
	}

	if existing == nil {
		if err := s.createPayment(ctx, tx, payment); err != nil {
			return fmt.Errorf("failed to create payment: %w", err)
		}
		s.incrementStat(response.Stats.CreatedEntities, "payments")
	} else {
		if conflict := s.resolvePaymentConflict(*existing, payment); conflict != nil {
			response.Conflicts = append(response.Conflicts, *conflict)
			if existing.UpdatedAt.After(payment.UpdatedAt) {
				s.incrementStat(response.Stats.ProcessedEntities, "payments")
				return nil
			}
		}
		if err := s.updatePayment(ctx, tx, payment); err != nil {
			return fmt.Errorf("failed to update payment: %w", err)
		}
		s.incrementStat(response.Stats.UpdatedEntities, "payments")
	}

	s.incrementStat(response.Stats.ProcessedEntities, "payments")
	return nil
}

// processSingleReceiptSafe processes a single receipt entity safely
func (s *SyncService) processSingleReceiptSafe(ctx context.Context, tx *gorm.DB, receipt entities.Receipt, licenseID uuid.UUID, response *dto.SyncResponse) error {
	if !s.validateReceiptLicense(ctx, receipt, licenseID) {
		return fmt.Errorf("receipt does not belong to license %s", licenseID)
	}

	existing, err := s.findReceiptByID(ctx, tx, receipt.ID)
	if err != nil {
		return fmt.Errorf("failed to find receipt: %w", err)
	}

	if existing == nil {
		if err := s.createReceipt(ctx, tx, receipt); err != nil {
			return fmt.Errorf("failed to create receipt: %w", err)
		}
		s.incrementStat(response.Stats.CreatedEntities, "receipts")
	} else {
		if conflict := s.resolveReceiptConflict(*existing, receipt); conflict != nil {
			response.Conflicts = append(response.Conflicts, *conflict)
			if existing.UpdatedAt.After(receipt.UpdatedAt) {
				s.incrementStat(response.Stats.ProcessedEntities, "receipts")
				return nil
			}
		}
		if err := s.updateReceipt(ctx, tx, receipt); err != nil {
			return fmt.Errorf("failed to update receipt: %w", err)
		}
		s.incrementStat(response.Stats.UpdatedEntities, "receipts")
	}

	s.incrementStat(response.Stats.ProcessedEntities, "receipts")
	return nil
}

// processSingleHistorySafe processes a single history entity safely
func (s *SyncService) processSingleHistorySafe(ctx context.Context, tx *gorm.DB, history entities.History, licenseID uuid.UUID, response *dto.SyncResponse) error {
	if !s.validateHistoryLicense(ctx, history, licenseID) {
		return fmt.Errorf("history does not belong to license %s", licenseID)
	}

	existing, err := s.findHistoryByID(ctx, tx, history.ID)
	if err != nil {
		return fmt.Errorf("failed to find history: %w", err)
	}

	if existing == nil {
		if err := s.createHistory(ctx, tx, history); err != nil {
			return fmt.Errorf("failed to create history: %w", err)
		}
		s.incrementStat(response.Stats.CreatedEntities, "histories")
	} else {
		if conflict := s.resolveHistoryConflict(*existing, history); conflict != nil {
			response.Conflicts = append(response.Conflicts, *conflict)
			if existing.UpdatedAt.After(history.UpdatedAt) {
				s.incrementStat(response.Stats.ProcessedEntities, "histories")
				return nil
			}
		}
		if err := s.updateHistory(ctx, tx, history); err != nil {
			return fmt.Errorf("failed to update history: %w", err)
		}
		s.incrementStat(response.Stats.UpdatedEntities, "histories")
	}

	s.incrementStat(response.Stats.ProcessedEntities, "histories")
	return nil
}

// processSingleShopSafe processes a single shop entity safely
func (s *SyncService) processSingleShopSafe(ctx context.Context, tx *gorm.DB, shop entities.Shop, licenseID uuid.UUID, response *dto.SyncResponse) error {
	log.Printf("DEBUG: processSingleShopSafe - Validating shop %s license: shop.LicenseID=%s vs licenseID=%s", shop.ID, shop.LicenseID, licenseID)

	if !s.validateShopLicense(ctx, shop, licenseID) {
		log.Printf("ERROR: Shop %s license validation failed: shop.LicenseID=%s does not match licenseID=%s", shop.ID, shop.LicenseID, licenseID)
		return fmt.Errorf("shop does not belong to license %s", licenseID)
	}

	log.Printf("DEBUG: processSingleShopSafe - Shop %s passed license validation", shop.ID)

	existing, err := s.findShopByID(ctx, tx, shop.ID)
	if err != nil {
		log.Printf("ERROR: Failed to find shop %s: %v", shop.ID, err)
		return fmt.Errorf("failed to find shop: %w", err)
	}

	if existing == nil {
		log.Printf("DEBUG: processSingleShopSafe - Creating new shop %s", shop.ID)
		if err := s.createShop(ctx, tx, shop); err != nil {
			log.Printf("ERROR: Failed to create shop %s: %v", shop.ID, err)
			return fmt.Errorf("failed to create shop: %w", err)
		}
		log.Printf("DEBUG: processSingleShopSafe - Successfully created shop %s", shop.ID)
		s.incrementStat(response.Stats.CreatedEntities, "shops")
	} else {
		log.Printf("DEBUG: processSingleShopSafe - Updating existing shop %s", shop.ID)
		if conflict := s.resolveShopConflict(*existing, shop); conflict != nil {
			response.Conflicts = append(response.Conflicts, *conflict)
			if existing.UpdatedAt.After(shop.UpdatedAt) {
				log.Printf("DEBUG: processSingleShopSafe - Shop %s conflict resolved, skipping update (existing is newer)", shop.ID)
				s.incrementStat(response.Stats.ProcessedEntities, "shops")
				return nil
			}
		}
		if err := s.updateShop(ctx, tx, shop); err != nil {
			log.Printf("ERROR: Failed to update shop %s: %v", shop.ID, err)
			return fmt.Errorf("failed to update shop: %w", err)
		}
		log.Printf("DEBUG: processSingleShopSafe - Successfully updated shop %s", shop.ID)
		s.incrementStat(response.Stats.UpdatedEntities, "shops")
	}

	s.incrementStat(response.Stats.ProcessedEntities, "shops")
	log.Printf("DEBUG: processSingleShopSafe - Completed processing shop %s", shop.ID)
	return nil
}

// processSingleUserSafe processes a single user entity safely with role validation
func (s *SyncService) processSingleUserSafe(ctx context.Context, tx *gorm.DB, user entities.User, licenseID uuid.UUID, response *dto.SyncResponse) error {
	if !s.validateUserLicense(ctx, user, licenseID) {
		return fmt.Errorf("user does not belong to license %s", licenseID)
	}

	// CRITICAL FIX: Validate role exists before creating/updating user
	if user.RoleID != nil {
		var roleExists bool
		err := tx.Model(&entities.Role{}).Select("count(*) > 0").Where("id = ?", *user.RoleID).Find(&roleExists).Error
		if err != nil {
			return fmt.Errorf("failed to validate role: %w", err)
		}
		if !roleExists {
			return fmt.Errorf("role %s does not exist", *user.RoleID)
		}
	}

	existing, err := s.findUserByID(ctx, tx, user.ID)
	if err != nil {
		return fmt.Errorf("failed to find user: %w", err)
	}

	if existing == nil {
		if err := s.createUser(ctx, tx, user); err != nil {
			return fmt.Errorf("failed to create user: %w", err)
		}
		s.incrementStat(response.Stats.CreatedEntities, "users")
	} else {
		if conflict := s.resolveUserConflict(*existing, user); conflict != nil {
			response.Conflicts = append(response.Conflicts, *conflict)
			if existing.UpdatedAt.After(user.UpdatedAt) {
				s.incrementStat(response.Stats.ProcessedEntities, "users")
				return nil
			}
		}
		if err := s.updateUser(ctx, tx, user); err != nil {
			return fmt.Errorf("failed to update user: %w", err)
		}
		s.incrementStat(response.Stats.UpdatedEntities, "users")
	}

	s.incrementStat(response.Stats.ProcessedEntities, "users")
	return nil
}

// processSingleStockHistorySafe processes a single stock history entity safely
func (s *SyncService) processSingleStockHistorySafe(ctx context.Context, tx *gorm.DB, stockHistory entities.StockHistory, syncProducts []entities.Product, syncContext dto.SyncContext, response *dto.SyncResponse) error {
	if !s.validateStockHistoryAccessWithSyncData(ctx, stockHistory, syncContext, syncProducts) {
		return fmt.Errorf("stock history access denied for user role and shop")
	}

	existing, err := s.findStockHistoryByID(ctx, tx, stockHistory.ID)
	if err != nil {
		return fmt.Errorf("failed to find stock history: %w", err)
	}

	if existing == nil {
		if err := s.createStockHistory(ctx, tx, stockHistory); err != nil {
			return fmt.Errorf("failed to create stock history: %w", err)
		}
		s.incrementStat(response.Stats.CreatedEntities, "stock_histories")
	} else {
		if conflict := s.resolveStockHistoryConflict(*existing, stockHistory); conflict != nil {
			response.Conflicts = append(response.Conflicts, *conflict)
			if existing.UpdatedAt.After(stockHistory.UpdatedAt) {
				s.incrementStat(response.Stats.ProcessedEntities, "stock_histories")
				return nil
			}
		}
		if err := s.updateStockHistory(ctx, tx, stockHistory); err != nil {
			return fmt.Errorf("failed to update stock history: %w", err)
		}
		s.incrementStat(response.Stats.UpdatedEntities, "stock_histories")
	}

	s.incrementStat(response.Stats.ProcessedEntities, "stock_histories")
	return nil
}

// processSingleTransactionProductSafe processes a single transaction product entity safely
func (s *SyncService) processSingleTransactionProductSafe(ctx context.Context, tx *gorm.DB, transactionProduct entities.TransactionProduct, syncTransactions []entities.Transaction, syncContext dto.SyncContext, response *dto.SyncResponse) error {
	// Enhanced validation with detailed error logging and sync data support
	if !s.validateTransactionProductAccessWithSyncData(ctx, transactionProduct, syncContext, syncTransactions) {
		// Add detailed error information for better debugging
		var transaction entities.Transaction
		var transactionLookupError string
		err := s.db.Select("shop_id").
			Where("id = ?", transactionProduct.TransactionID).
			First(&transaction).Error

		if err != nil {
			transactionLookupError = err.Error()
		}

		errorDetails := map[string]interface{}{
			"transaction_id": transactionProduct.TransactionID,
			"license_id":     syncContext.LicenseID,
			"user_role":      syncContext.UserRole,
			"user_id":        syncContext.UserID,
		}

		if transactionLookupError != "" {
			errorDetails["transaction_lookup_error"] = transactionLookupError
		} else {
			errorDetails["transaction_shop_id"] = transaction.ShopID
			errorDetails["accessible_shops"] = syncContext.AccessibleShopIDs
		}

		return fmt.Errorf("transaction product access denied for user role and shop - details: %+v", errorDetails)
	}

	existing, err := s.findTransactionProductByID(ctx, tx, transactionProduct.ID)
	if err != nil {
		return fmt.Errorf("failed to find transaction product: %w", err)
	}

	if existing == nil {
		if err := s.createTransactionProduct(ctx, tx, transactionProduct); err != nil {
			return fmt.Errorf("failed to create transaction product: %w", err)
		}
		s.incrementStat(response.Stats.CreatedEntities, "transaction_products")
	} else {
		if conflict := s.resolveTransactionProductConflict(*existing, transactionProduct); conflict != nil {
			response.Conflicts = append(response.Conflicts, *conflict)
			if existing.UpdatedAt.After(transactionProduct.UpdatedAt) {
				s.incrementStat(response.Stats.ProcessedEntities, "transaction_products")
				return nil
			}
		}
		if err := s.updateTransactionProduct(ctx, tx, transactionProduct); err != nil {
			return fmt.Errorf("failed to update transaction product: %w", err)
		}
		s.incrementStat(response.Stats.UpdatedEntities, "transaction_products")
	}

	s.incrementStat(response.Stats.ProcessedEntities, "transaction_products")
	return nil
}
