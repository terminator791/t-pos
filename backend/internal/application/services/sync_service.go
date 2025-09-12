package services

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/terminator791/t-pos/internal/domain/dto"
	"github.com/terminator791/t-pos/internal/domain/entities"
	"github.com/terminator791/t-pos/internal/domain/repositories"
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
	db                      *gorm.DB
	cartRepo                repositories.CartRepository
	categoryRepo            repositories.CategoryRepository
	expenseRepo             repositories.ExpenseRepository
	historyRepo             repositories.HistoryRepository
	paymentRepo             repositories.PaymentRepository
	productRepo             repositories.ProductRepository
	receiptRepo             repositories.ReceiptRepository
	shopRepo                repositories.ShopRepository
	stockHistoryRepo        repositories.StockHistoryRepository
	transactionRepo         repositories.TransactionRepository
	transactionProductRepo  repositories.TransactionProductRepository
	userRepo                repositories.UserRepository
	conflictStrategy        dto.ConflictResolutionStrategy
	batchSize               int
	maxEntitiesPerSync      int
	transactionTimeout      time.Duration
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
) *SyncService {
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
		batchSize:              DefaultBatchSize,
		maxEntitiesPerSync:     MaxEntitiesPerSync,
		transactionTimeout:     DefaultTransactionTimeout,
	}
}

// SetBatchSize allows configuration of batch size for processing
func (s *SyncService) SetBatchSize(size int) {
	if size > 0 && size <= 500 {
		s.batchSize = size
	}
}

// SetMaxEntitiesPerSync sets the maximum entities allowed per sync operation
func (s *SyncService) SetMaxEntitiesPerSync(max int) {
	if max > 0 && max <= 10000 {
		s.maxEntitiesPerSync = max
	}
}

// SetTransactionTimeout sets the database transaction timeout
func (s *SyncService) SetTransactionTimeout(timeout time.Duration) {
	if timeout > 0 && timeout <= MaxTransactionTimeout {
		s.transactionTimeout = timeout
	}
}

// ProcessSync handles the complete synchronization process
func (s *SyncService) ProcessSync(ctx context.Context, req dto.SyncRequest, licenseID uuid.UUID, userID uuid.UUID) (*dto.SyncResponse, error) {
	startTime := time.Now()
	
	// Validate sync request size to prevent memory issues
	if err := s.validateSyncRequest(req); err != nil {
		return nil, fmt.Errorf("sync request validation failed: %w", err)
	}
	
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
	ctxWithTimeout, cancel := context.WithTimeout(ctx, s.transactionTimeout)
	defer cancel()

	// Start a database transaction with proper timeout and isolation level
	tx := s.db.Begin(&sql.TxOptions{
		Isolation: sql.LevelReadCommitted,
	})
	if tx.Error != nil {
		return nil, fmt.Errorf("failed to start transaction: %w", tx.Error)
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			log.Printf("Sync transaction rolled back due to panic: %v", r)
			panic(r)
		}
	}()

	// Phase 1: Push - Process incoming changes from mobile
	if err := s.pushChanges(ctxWithTimeout, tx, req, licenseID, response); err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to push changes: %w", err)
	}

	// Phase 2: Pull - Get server changes since last sync
	if err := s.pullChanges(ctxWithTimeout, tx, req.LastSyncTimestamp, licenseID, response); err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to pull changes: %w", err)
	}

	// Commit the transaction
	if err := tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	// Clean up response - remove empty arrays and fix conflict data
	s.cleanupResponse(response)

	// Calculate final stats
	response.Stats.ProcessingTimeMs = time.Since(startTime).Milliseconds()
	response.Stats.ConflictCount = len(response.Conflicts)
	response.Stats.ErrorCount = len(response.Errors)

	log.Printf("Sync completed for license %s: %d conflicts, %d errors, %dms, %d entities processed",
		licenseID.String(), response.Stats.ConflictCount, response.Stats.ErrorCount, 
		response.Stats.ProcessingTimeMs, s.getTotalProcessedEntities(response.Stats))

	return response, nil
}

// validateSyncRequest validates the sync request to prevent memory and performance issues
func (s *SyncService) validateSyncRequest(req dto.SyncRequest) error {
	totalEntities := len(req.Carts) + len(req.Categories) + len(req.Products) + 
		len(req.Transactions) + len(req.Payments) + len(req.Expenses) +
		len(req.Receipts) + len(req.Histories) + len(req.Shops) + 
		len(req.StockHistories) + len(req.TransactionProducts) + len(req.Users)
	
	if totalEntities > s.maxEntitiesPerSync {
		return fmt.Errorf("sync request too large: %d entities exceeds maximum of %d", 
			totalEntities, s.maxEntitiesPerSync)
	}
	
	// Validate individual entity type limits
	entityLimits := map[string]int{
		"carts": len(req.Carts),
		"categories": len(req.Categories),
		"products": len(req.Products),
		"transactions": len(req.Transactions),
		"payments": len(req.Payments),
		"expenses": len(req.Expenses),
		"receipts": len(req.Receipts),
		"histories": len(req.Histories),
		"shops": len(req.Shops),
		"stock_histories": len(req.StockHistories),
		"transaction_products": len(req.TransactionProducts),
		"users": len(req.Users),
	}
	
	for entityType, count := range entityLimits {
		if count > s.maxEntitiesPerSync/2 { // No single entity type should exceed half the limit
			return fmt.Errorf("%s count too large: %d exceeds maximum of %d", 
				entityType, count, s.maxEntitiesPerSync/2)
		}
	}
	
	return nil
}

// getTotalProcessedEntities calculates total entities processed across all types
func (s *SyncService) getTotalProcessedEntities(stats dto.SyncStats) int {
	total := 0
	for _, count := range stats.ProcessedEntities {
		total += count
	}
	return total
}

// pushChanges processes incoming changes from mobile client
func (s *SyncService) pushChanges(ctx context.Context, tx *gorm.DB, req dto.SyncRequest, licenseID uuid.UUID, response *dto.SyncResponse) error {
	// Process each entity type
	if err := s.pushCarts(ctx, tx, req.Carts, licenseID, response); err != nil {
		return fmt.Errorf("failed to push carts: %w", err)
	}

	if err := s.pushCategories(ctx, tx, req.Categories, licenseID, response); err != nil {
		return fmt.Errorf("failed to push categories: %w", err)
	}

	if err := s.pushProducts(ctx, tx, req.Products, licenseID, response); err != nil {
		return fmt.Errorf("failed to push products: %w", err)
	}

	if err := s.pushTransactions(ctx, tx, req.Transactions, licenseID, response); err != nil {
		return fmt.Errorf("failed to push transactions: %w", err)
	}

	if err := s.pushExpenses(ctx, tx, req.Expenses, licenseID, response); err != nil {
		return fmt.Errorf("failed to push expenses: %w", err)
	}

	if err := s.pushPayments(ctx, tx, req.Payments, licenseID, response); err != nil {
		return fmt.Errorf("failed to push payments: %w", err)
	}

	if err := s.pushReceipts(ctx, tx, req.Receipts, licenseID, response); err != nil {
		return fmt.Errorf("failed to push receipts: %w", err)
	}

	if err := s.pushHistories(ctx, tx, req.Histories, licenseID, response); err != nil {
		return fmt.Errorf("failed to push histories: %w", err)
	}

	if err := s.pushShops(ctx, tx, req.Shops, licenseID, response); err != nil {
		return fmt.Errorf("failed to push shops: %w", err)
	}

	if err := s.pushStockHistories(ctx, tx, req.StockHistories, licenseID, response); err != nil {
		return fmt.Errorf("failed to push stock histories: %w", err)
	}

	if err := s.pushTransactionProducts(ctx, tx, req.TransactionProducts, licenseID, response); err != nil {
		return fmt.Errorf("failed to push transaction products: %w", err)
	}

	if err := s.pushUsers(ctx, tx, req.Users, licenseID, response); err != nil {
		return fmt.Errorf("failed to push users: %w", err)
	}

	return nil
}

// pullChanges retrieves server changes since last sync
func (s *SyncService) pullChanges(ctx context.Context, tx *gorm.DB, lastSync *time.Time, licenseID uuid.UUID, response *dto.SyncResponse) error {
	// If no last sync timestamp, this is initial sync - return recent data
	if lastSync == nil {
		// For initial sync, return data from last 30 days
		thirtyDaysAgo := time.Now().AddDate(0, 0, -30)
		lastSync = &thirtyDaysAgo
	}

	// Pull changes for each entity type
	if err := s.pullCarts(ctx, tx, *lastSync, licenseID, response); err != nil {
		return fmt.Errorf("failed to pull carts: %w", err)
	}

	if err := s.pullCategories(ctx, tx, *lastSync, licenseID, response); err != nil {
		return fmt.Errorf("failed to pull categories: %w", err)
	}

	if err := s.pullProducts(ctx, tx, *lastSync, licenseID, response); err != nil {
		return fmt.Errorf("failed to pull products: %w", err)
	}

	if err := s.pullTransactions(ctx, tx, *lastSync, licenseID, response); err != nil {
		return fmt.Errorf("failed to pull transactions: %w", err)
	}

	if err := s.pullExpenses(ctx, tx, *lastSync, licenseID, response); err != nil {
		return fmt.Errorf("failed to pull expenses: %w", err)
	}

	if err := s.pullPayments(ctx, tx, *lastSync, licenseID, response); err != nil {
		return fmt.Errorf("failed to pull payments: %w", err)
	}

	if err := s.pullReceipts(ctx, tx, *lastSync, licenseID, response); err != nil {
		return fmt.Errorf("failed to pull receipts: %w", err)
	}

	if err := s.pullHistories(ctx, tx, *lastSync, licenseID, response); err != nil {
		return fmt.Errorf("failed to pull histories: %w", err)
	}

	if err := s.pullShops(ctx, tx, *lastSync, licenseID, response); err != nil {
		return fmt.Errorf("failed to pull shops: %w", err)
	}

	if err := s.pullStockHistories(ctx, tx, *lastSync, licenseID, response); err != nil {
		return fmt.Errorf("failed to pull stock histories: %w", err)
	}

	if err := s.pullTransactionProducts(ctx, tx, *lastSync, licenseID, response); err != nil {
		return fmt.Errorf("failed to pull transaction products: %w", err)
	}

	if err := s.pullUsers(ctx, tx, *lastSync, licenseID, response); err != nil {
		return fmt.Errorf("failed to pull users: %w", err)
	}

	return nil
}

// pushCarts handles cart synchronization
// pushCarts handles cart synchronization with batch processing
func (s *SyncService) pushCarts(ctx context.Context, tx *gorm.DB, carts []entities.Cart, licenseID uuid.UUID, response *dto.SyncResponse) error {
	totalCarts := len(carts)
	if totalCarts == 0 {
		return nil
	}
	
	log.Printf("Processing %d carts in batches of %d", totalCarts, s.batchSize)
	
	for i := 0; i < totalCarts; i += s.batchSize {
		end := i + s.batchSize
		if end > totalCarts {
			end = totalCarts
		}
		
		batch := carts[i:end]
		log.Printf("Processing cart batch %d-%d of %d", i+1, end, totalCarts)
		
		// Process each cart in the batch
		for _, cart := range batch {
			if err := s.processSingleCart(ctx, tx, cart, licenseID, response); err != nil {
				log.Printf("Error processing cart %s: %v", cart.ID, err)
				// Continue with next entity instead of failing entire batch
				continue
			}
		}
		
		// Check context for cancellation between batches
		select {
		case <-ctx.Done():
			return fmt.Errorf("sync operation cancelled: %w", ctx.Err())
		default:
		}
	}
	
	return nil
}

// processSingleCart processes a single cart entity with enhanced error handling
func (s *SyncService) processSingleCart(ctx context.Context, tx *gorm.DB, cart entities.Cart, licenseID uuid.UUID, response *dto.SyncResponse) error {
	// Validate cart belongs to license
	if !s.validateCartLicense(ctx, cart, licenseID) {
		s.addDetailedError(response, "carts", cart.ID, "unauthorized", "Cart does not belong to license", 
			map[string]interface{}{"cart_shop_id": cart.ShopID, "license_id": licenseID})
		return nil // Continue processing other entities
	}

	// Check if cart exists with retry
	var existingCart *entities.Cart
	var err error
	
	operation := func() error {
		existingCart, err = s.findCartByID(ctx, tx, cart.ID)
		return err
	}
	
	if retryErr := s.retryOperation(ctx, operation, 2, 50*time.Millisecond, fmt.Sprintf("find_cart_%s", cart.ID)); retryErr != nil {
		s.addDetailedError(response, "carts", cart.ID, "database_error", retryErr.Error(), 
			map[string]interface{}{"operation": "find", "retry_attempts": 2})
		return nil // Continue processing other entities
	}

	if existingCart == nil {
		// Create new cart with retry
		createOperation := func() error {
			return s.createCart(ctx, tx, cart)
		}
		
		if err := s.retryOperation(ctx, createOperation, 3, 100*time.Millisecond, fmt.Sprintf("create_cart_%s", cart.ID)); err != nil {
			s.addDetailedError(response, "carts", cart.ID, "create_failed", err.Error(), 
				map[string]interface{}{"operation": "create", "retry_attempts": 3})
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
		
		if err := s.retryOperation(ctx, updateOperation, 3, 100*time.Millisecond, fmt.Sprintf("update_cart_%s", cart.ID)); err != nil {
			s.addDetailedError(response, "carts", cart.ID, "update_failed", err.Error(), 
				map[string]interface{}{"operation": "update", "retry_attempts": 3})
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
	const maxResultsPerType = 1000
	err := query.Limit(maxResultsPerType).Find(&carts).Error
	
	if err != nil {
		return fmt.Errorf("failed to query carts: %w", err)
	}
	
	// Log query performance for monitoring
	log.Printf("Retrieved %d carts for license %s since %v", len(carts), licenseID, lastSync)
	
	// If we hit the limit, log a warning about potential incomplete sync
	if len(carts) == maxResultsPerType {
		log.Printf("WARNING: Cart sync hit result limit (%d), some data may be missing. Consider using smaller sync intervals.", maxResultsPerType)
		s.addError(response, "carts", uuid.Nil, "result_limit_reached", 
			fmt.Sprintf("Retrieved maximum %d carts. Some data may be missing due to result size limits.", maxResultsPerType))
	}

	response.Carts = carts
	return nil
}

// Helper methods for cart operations
func (s *SyncService) validateCartLicense(ctx context.Context, cart entities.Cart, licenseID uuid.UUID) bool {
	// Validate that the cart's shop belongs to the license
	var count int64
	s.db.Model(&entities.Shop{}).Where("id = ? AND license_id = ?", cart.ShopID, licenseID).Count(&count)
	return count > 0
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

// Helper methods for category operations
func (s *SyncService) validateCategoryLicense(ctx context.Context, category entities.Category, licenseID uuid.UUID) bool {
	// Validate that the category's shop belongs to the license
	var count int64
	s.db.Model(&entities.Shop{}).Where("id = ? AND license_id = ?", category.ShopID, licenseID).Count(&count)
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
		if !s.validateCategoryLicense(ctx, category, licenseID) {
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
	var categories []entities.Category
	err := tx.WithContext(ctx).
		Select("categories.id", "categories.shop_id", "categories.name", "categories.created_at", "categories.updated_at").
		Joins("JOIN shops ON categories.shop_id = shops.id").
		Where("shops.license_id = ? AND categories.updated_at > ?", licenseID, lastSync).
		Find(&categories).Error

	if err != nil {
		return fmt.Errorf("failed to query categories: %w", err)
	}

	response.Categories = categories
	return nil
}

func (s *SyncService) pushProducts(ctx context.Context, tx *gorm.DB, products []entities.Product, licenseID uuid.UUID, response *dto.SyncResponse) error {
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

func (s *SyncService) pullProducts(ctx context.Context, tx *gorm.DB, lastSync time.Time, licenseID uuid.UUID, response *dto.SyncResponse) error {
	// Get all products for the license that were updated after lastSync
	var products []entities.Product
	err := tx.WithContext(ctx).
		Joins("JOIN shops ON products.shop_id = shops.id").
		Where("shops.license_id = ? AND products.updated_at > ?", licenseID, lastSync).
		Find(&products).Error

	if err != nil {
		return fmt.Errorf("failed to query products: %w", err)
	}

	response.Products = products
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

func (s *SyncService) pullTransactions(ctx context.Context, tx *gorm.DB, lastSync time.Time, licenseID uuid.UUID, response *dto.SyncResponse) error {
	// Get all transactions for the license that were updated after lastSync
	var transactions []entities.Transaction
	err := tx.WithContext(ctx).
		Joins("JOIN shops ON transactions.shop_id = shops.id").
		Where("shops.license_id = ? AND transactions.updated_at > ?", licenseID, lastSync).
		Find(&transactions).Error

	if err != nil {
		return fmt.Errorf("failed to query transactions: %w", err)
	}

	response.Transactions = transactions
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
	var expenses []entities.Expense
	err := tx.WithContext(ctx).
		Joins("JOIN shops ON expenses.shop_id = shops.id").
		Where("shops.license_id = ? AND expenses.updated_at > ?", licenseID, lastSync).
		Find(&expenses).Error

	if err != nil {
		return fmt.Errorf("failed to query expenses: %w", err)
	}

	response.Expenses = expenses
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
	var payments []entities.Payment
	err := tx.WithContext(ctx).
		Joins("JOIN shops ON payments.shop_id = shops.id").
		Where("shops.license_id = ? AND payments.updated_at > ?", licenseID, lastSync).
		Find(&payments).Error

	if err != nil {
		return fmt.Errorf("failed to query payments: %w", err)
	}

	response.Payments = payments
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
	var receipts []entities.Receipt
	err := tx.Joins("JOIN shops ON receipts.shop_id = shops.id").
		Where("shops.license_id = ? AND receipts.updated_at > ?", licenseID, lastSync).
		Find(&receipts).Error
	if err != nil {
		return fmt.Errorf("failed to fetch receipts: %w", err)
	}

	response.Receipts = receipts
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
	var histories []entities.History
	err := tx.Joins("JOIN shops ON histories.shop_id = shops.id").
		Where("shops.license_id = ? AND histories.updated_at > ?", licenseID, lastSync).
		Find(&histories).Error
	if err != nil {
		return fmt.Errorf("failed to fetch histories: %w", err)
	}

	response.Histories = histories
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
	var shops []entities.Shop
	err := tx.Where("license_id = ? AND updated_at > ?", licenseID, lastSync).
		Find(&shops).Error
	if err != nil {
		return fmt.Errorf("failed to fetch shops: %w", err)
	}

	response.Shops = shops
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
func (s *SyncService) pushStockHistories(ctx context.Context, tx *gorm.DB, stockHistories []entities.StockHistory, licenseID uuid.UUID, response *dto.SyncResponse) error {
	for _, stockHistory := range stockHistories {
		// Validate stock history belongs to license
		if !s.validateStockHistoryLicense(ctx, stockHistory, licenseID) {
			s.addError(response, "stock_histories", stockHistory.ID, "unauthorized", "Stock history does not belong to license")
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
	var stockHistories []entities.StockHistory
	err := tx.Joins("JOIN products ON stock_histories.product_id = products.id").
		Joins("JOIN shops ON products.shop_id = shops.id").
		Where("shops.license_id = ? AND stock_histories.updated_at > ?", licenseID, lastSync).
		Find(&stockHistories).Error
	if err != nil {
		return fmt.Errorf("failed to fetch stock histories: %w", err)
	}

	response.StockHistories = stockHistories
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
func (s *SyncService) pushTransactionProducts(ctx context.Context, tx *gorm.DB, transactionProducts []entities.TransactionProduct, licenseID uuid.UUID, response *dto.SyncResponse) error {
	for _, transactionProduct := range transactionProducts {
		// Validate transaction product belongs to license
		if !s.validateTransactionProductLicense(ctx, transactionProduct, licenseID) {
			s.addError(response, "transaction_products", transactionProduct.ID, "unauthorized", "Transaction product does not belong to license")
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
	var transactionProducts []entities.TransactionProduct
	err := tx.Joins("JOIN transactions ON transaction_products.transaction_id = transactions.id").
		Joins("JOIN shops ON transactions.shop_id = shops.id").
		Where("shops.license_id = ? AND transaction_products.updated_at > ?", licenseID, lastSync).
		Find(&transactionProducts).Error
	if err != nil {
		return fmt.Errorf("failed to fetch transaction products: %w", err)
	}

	response.TransactionProducts = transactionProducts
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
	var users []entities.User
	err := tx.Where("license_id = ? AND updated_at > ?", licenseID, lastSync).
		Find(&users).Error
	if err != nil {
		return fmt.Errorf("failed to fetch users: %w", err)
	}

	response.Users = users
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
