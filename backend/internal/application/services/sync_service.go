package services

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/terminator791/t-pos/internal/domain/dto"
	"github.com/terminator791/t-pos/internal/domain/entities"
	"github.com/terminator791/t-pos/internal/domain/repositories"
	"gorm.io/gorm"
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
	}
}

// ProcessSync handles the complete synchronization process
func (s *SyncService) ProcessSync(ctx context.Context, req dto.SyncRequest, licenseID uuid.UUID, userID uuid.UUID) (*dto.SyncResponse, error) {
	startTime := time.Now()
	
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

	// Start a database transaction
	tx := s.db.Begin()
	if tx.Error != nil {
		return nil, fmt.Errorf("failed to start transaction: %w", tx.Error)
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()

	// Phase 1: Push - Process incoming changes from mobile
	if err := s.pushChanges(ctx, tx, req, licenseID, response); err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to push changes: %w", err)
	}

	// Phase 2: Pull - Get server changes since last sync
	if err := s.pullChanges(ctx, tx, req.LastSyncTimestamp, licenseID, response); err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to pull changes: %w", err)
	}

	// Commit the transaction
	if err := tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	// Calculate final stats
	response.Stats.ProcessingTimeMs = time.Since(startTime).Milliseconds()
	response.Stats.ConflictCount = len(response.Conflicts)
	response.Stats.ErrorCount = len(response.Errors)

	log.Printf("Sync completed for license %s: %d conflicts, %d errors, %dms",
		licenseID.String(), response.Stats.ConflictCount, response.Stats.ErrorCount, response.Stats.ProcessingTimeMs)

	return response, nil
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

	// Continue with other entity types...
	// For now, implementing core entities (carts, categories, products, transactions, expenses, payments)

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

	return nil
}

// pushCarts handles cart synchronization
func (s *SyncService) pushCarts(ctx context.Context, tx *gorm.DB, carts []entities.Cart, licenseID uuid.UUID, response *dto.SyncResponse) error {
	for _, cart := range carts {
		// Validate cart belongs to license
		if !s.validateCartLicense(ctx, cart, licenseID) {
			s.addError(response, "carts", cart.ID, "unauthorized", "Cart does not belong to license")
			continue
		}

		// Check if cart exists
		existingCart, err := s.findCartByID(ctx, tx, cart.ID)
		if err != nil && err != gorm.ErrRecordNotFound {
			s.addError(response, "carts", cart.ID, "database_error", err.Error())
			continue
		}

		if existingCart == nil {
			// Create new cart
			if err := s.createCart(ctx, tx, cart); err != nil {
				s.addError(response, "carts", cart.ID, "create_failed", err.Error())
				continue
			}
			s.incrementStat(response.Stats.CreatedEntities, "carts")
		} else {
			// Handle potential conflict
			if conflict := s.resolveCartConflict(*existingCart, cart); conflict != nil {
				response.Conflicts = append(response.Conflicts, *conflict)
				// Use server version in case of conflict (for LastWriteWins strategy)
				if existingCart.UpdatedAt.After(cart.UpdatedAt) {
					continue // Skip update, server version is newer
				}
			}

			// Update existing cart
			if err := s.updateCart(ctx, tx, cart); err != nil {
				s.addError(response, "carts", cart.ID, "update_failed", err.Error())
				continue
			}
			s.incrementStat(response.Stats.UpdatedEntities, "carts")
		}

		s.incrementStat(response.Stats.ProcessedEntities, "carts")
	}

	return nil
}

// pullCarts retrieves server-side cart changes
func (s *SyncService) pullCarts(ctx context.Context, tx *gorm.DB, lastSync time.Time, licenseID uuid.UUID, response *dto.SyncResponse) error {
	// Get all carts for the license that were updated after lastSync
	var carts []entities.Cart
	err := tx.WithContext(ctx).
		Joins("JOIN shops ON carts.shop_id = shops.id").
		Where("shops.license_id = ? AND carts.updated_at > ?", licenseID, lastSync).
		Find(&carts).Error

	if err != nil {
		return fmt.Errorf("failed to query carts: %w", err)
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
	err := tx.WithContext(ctx).Where("id = ?", id).First(&category).Error
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