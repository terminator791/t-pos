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

	// Continue with other entity types...
	// For now, implementing core entities (carts, categories, products, transactions)

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

// Helper methods for other entities (categories, products, transactions)
// Implementation follows the same pattern as carts

func (s *SyncService) pushCategories(ctx context.Context, tx *gorm.DB, categories []entities.Category, licenseID uuid.UUID, response *dto.SyncResponse) error {
	// Similar implementation to pushCarts
	// TODO: Implement category sync logic
	return nil
}

func (s *SyncService) pullCategories(ctx context.Context, tx *gorm.DB, lastSync time.Time, licenseID uuid.UUID, response *dto.SyncResponse) error {
	// Similar implementation to pullCarts
	// TODO: Implement category pull logic
	return nil
}

func (s *SyncService) pushProducts(ctx context.Context, tx *gorm.DB, products []entities.Product, licenseID uuid.UUID, response *dto.SyncResponse) error {
	// Similar implementation to pushCarts
	// TODO: Implement product sync logic
	return nil
}

func (s *SyncService) pullProducts(ctx context.Context, tx *gorm.DB, lastSync time.Time, licenseID uuid.UUID, response *dto.SyncResponse) error {
	// Similar implementation to pullCarts
	// TODO: Implement product pull logic
	return nil
}

func (s *SyncService) pushTransactions(ctx context.Context, tx *gorm.DB, transactions []entities.Transaction, licenseID uuid.UUID, response *dto.SyncResponse) error {
	// Similar implementation to pushCarts
	// TODO: Implement transaction sync logic
	return nil
}

func (s *SyncService) pullTransactions(ctx context.Context, tx *gorm.DB, lastSync time.Time, licenseID uuid.UUID, response *dto.SyncResponse) error {
	// Similar implementation to pullCarts
	// TODO: Implement transaction pull logic
	return nil
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