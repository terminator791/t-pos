package sync

import (
	"time"

	"github.com/google/uuid"
	"github.com/terminator791/t-pos/internal/domain/entities"
)

// ConflictResolver handles conflict resolution logic for sync operations
type ConflictResolver struct {
	// Add configuration options if needed in the future
	DefaultStrategy ConflictStrategy
}

// ConflictStrategy defines how conflicts should be resolved
type ConflictStrategy string

const (
	// StrategyLastWriteWins uses timestamp to determine winner
	StrategyLastWriteWins ConflictStrategy = "last_write_wins"
	// StrategyServerWins always prefers server version
	StrategyServerWins ConflictStrategy = "server_wins"
	// StrategyClientWins always prefers client version  
	StrategyClientWins ConflictStrategy = "client_wins"
)

// EntityConflict represents a conflict between server and client data
type EntityConflict struct {
	EntityType    string      `json:"entity_type"`
	EntityID      uuid.UUID   `json:"entity_id"`
	ServerVersion interface{} `json:"server_version"`
	ClientVersion interface{} `json:"client_version"`
	ServerTime    time.Time   `json:"server_time"`
	ClientTime    time.Time   `json:"client_time"`
}

// ConflictResult represents the outcome of conflict resolution
type ConflictResult struct {
	Resolution ConflictResolution `json:"resolution"`
	Winner     interface{}        `json:"winner"`
	ShouldSync bool              `json:"should_sync"` // Whether to proceed with sync
}

// NewConflictResolver creates a new conflict resolver with default LWW strategy
func NewConflictResolver() *ConflictResolver {
	return &ConflictResolver{
		DefaultStrategy: StrategyLastWriteWins,
	}
}

// ResolveCartConflict resolves conflicts for Cart entities
func (cr *ConflictResolver) ResolveCartConflict(serverCart *entities.Cart, clientCart entities.Cart) ConflictResult {
	if serverCart == nil {
		// No conflict - new record
		return ConflictResult{
			Resolution: ConflictResolution{
				EntityType:    "cart",
				EntityID:      clientCart.ID,
				ConflictType:  "no_conflict",
				Resolution:    "client_wins",
				ClientVersion: clientCart.UpdatedAt,
				ResolvedAt:    time.Now().UTC(),
			},
			Winner:     clientCart,
			ShouldSync: true,
		}
	}

	return cr.resolveLWW("cart", serverCart.ID, serverCart.UpdatedAt, clientCart.UpdatedAt, serverCart, clientCart)
}

// ResolveCategoryConflict resolves conflicts for Category entities
func (cr *ConflictResolver) ResolveCategoryConflict(serverCategory *entities.Category, clientCategory entities.Category) ConflictResult {
	if serverCategory == nil {
		return ConflictResult{
			Resolution: ConflictResolution{
				EntityType:    "category",
				EntityID:      clientCategory.ID,
				ConflictType:  "no_conflict",
				Resolution:    "client_wins",
				ClientVersion: clientCategory.UpdatedAt,
				ResolvedAt:    time.Now().UTC(),
			},
			Winner:     clientCategory,
			ShouldSync: true,
		}
	}

	return cr.resolveLWW("category", serverCategory.ID, serverCategory.UpdatedAt, clientCategory.UpdatedAt, serverCategory, clientCategory)
}

// ResolveProductConflict resolves conflicts for Product entities
func (cr *ConflictResolver) ResolveProductConflict(serverProduct *entities.Product, clientProduct entities.Product) ConflictResult {
	if serverProduct == nil {
		return ConflictResult{
			Resolution: ConflictResolution{
				EntityType:    "product",
				EntityID:      clientProduct.ID,
				ConflictType:  "no_conflict",
				Resolution:    "client_wins",
				ClientVersion: clientProduct.UpdatedAt,
				ResolvedAt:    time.Now().UTC(),
			},
			Winner:     clientProduct,
			ShouldSync: true,
		}
	}

	// Special handling for products with stock changes
	result := cr.resolveLWW("product", serverProduct.ID, serverProduct.UpdatedAt, clientProduct.UpdatedAt, serverProduct, clientProduct)
	
	// If server wins but client has stock changes, we might need special handling
	if result.Resolution.Resolution == "server_wins" && serverProduct.Stock != clientProduct.Stock {
		// Log the stock discrepancy for manual review
		// In a more advanced system, we could create a stock adjustment record
		result.Resolution.ConflictType = "stock_conflict"
	}

	return result
}

// ResolveTransactionConflict resolves conflicts for Transaction entities
func (cr *ConflictResolver) ResolveTransactionConflict(serverTransaction *entities.Transaction, clientTransaction entities.Transaction) ConflictResult {
	if serverTransaction == nil {
		return ConflictResult{
			Resolution: ConflictResolution{
				EntityType:    "transaction",
				EntityID:      clientTransaction.ID,
				ConflictType:  "no_conflict",
				Resolution:    "client_wins",
				ClientVersion: clientTransaction.UpdatedAt,
				ResolvedAt:    time.Now().UTC(),
			},
			Winner:     clientTransaction,
			ShouldSync: true,
		}
	}

	// Transactions should rarely have conflicts since they're usually created offline
	// But if they do, use LWW with special consideration for status changes
	result := cr.resolveLWW("transaction", serverTransaction.ID, serverTransaction.UpdatedAt, clientTransaction.UpdatedAt, serverTransaction, clientTransaction)
	
	// Special handling for transaction status conflicts
	if serverTransaction.Status != clientTransaction.Status {
		result.Resolution.ConflictType = "status_conflict"
		
		// Business rule: completed transactions from server should generally win
		if serverTransaction.Status == entities.TransactionStatusCompleted {
			result.Resolution.Resolution = "server_wins"
			result.Winner = serverTransaction
		}
	}

	return result
}

// ResolveExpenseConflict resolves conflicts for Expense entities
func (cr *ConflictResolver) ResolveExpenseConflict(serverExpense *entities.Expense, clientExpense entities.Expense) ConflictResult {
	if serverExpense == nil {
		return ConflictResult{
			Resolution: ConflictResolution{
				EntityType:    "expense",
				EntityID:      clientExpense.ID,
				ConflictType:  "no_conflict",
				Resolution:    "client_wins",
				ClientVersion: clientExpense.UpdatedAt,
				ResolvedAt:    time.Now().UTC(),
			},
			Winner:     clientExpense,
			ShouldSync: true,
		}
	}

	return cr.resolveLWW("expense", serverExpense.ID, serverExpense.UpdatedAt, clientExpense.UpdatedAt, serverExpense, clientExpense)
}

// ResolvePaymentConflict resolves conflicts for Payment entities
func (cr *ConflictResolver) ResolvePaymentConflict(serverPayment *entities.Payment, clientPayment entities.Payment) ConflictResult {
	if serverPayment == nil {
		return ConflictResult{
			Resolution: ConflictResolution{
				EntityType:    "payment",
				EntityID:      clientPayment.ID,
				ConflictType:  "no_conflict",
				Resolution:    "client_wins",
				ClientVersion: clientPayment.UpdatedAt,
				ResolvedAt:    time.Now().UTC(),
			},
			Winner:     clientPayment,
			ShouldSync: true,
		}
	}

	return cr.resolveLWW("payment", serverPayment.ID, serverPayment.UpdatedAt, clientPayment.UpdatedAt, serverPayment, clientPayment)
}

// ResolveShopConflict resolves conflicts for Shop entities
func (cr *ConflictResolver) ResolveShopConflict(serverShop *entities.Shop, clientShop entities.Shop) ConflictResult {
	if serverShop == nil {
		return ConflictResult{
			Resolution: ConflictResolution{
				EntityType:    "shop",
				EntityID:      clientShop.ID,
				ConflictType:  "no_conflict",
				Resolution:    "client_wins",
				ClientVersion: clientShop.UpdatedAt,
				ResolvedAt:    time.Now().UTC(),
			},
			Winner:     clientShop,
			ShouldSync: true,
		}
	}

	return cr.resolveLWW("shop", serverShop.ID, serverShop.UpdatedAt, clientShop.UpdatedAt, serverShop, clientShop)
}

// ResolveUserConflict resolves conflicts for User entities
func (cr *ConflictResolver) ResolveUserConflict(serverUser *entities.User, clientUser entities.User) ConflictResult {
	if serverUser == nil {
		return ConflictResult{
			Resolution: ConflictResolution{
				EntityType:    "user",
				EntityID:      clientUser.ID,
				ConflictType:  "no_conflict",
				Resolution:    "client_wins",
				ClientVersion: clientUser.UpdatedAt,
				ResolvedAt:    time.Now().UTC(),
			},
			Winner:     clientUser,
			ShouldSync: true,
		}
	}

	// For users, server usually wins for security-related fields
	// But use LWW for basic profile information
	return cr.resolveLWW("user", serverUser.ID, serverUser.UpdatedAt, clientUser.UpdatedAt, serverUser, clientUser)
}

// resolveLWW implements Last-Write-Wins conflict resolution
func (cr *ConflictResolver) resolveLWW(entityType string, entityID uuid.UUID, serverTime, clientTime time.Time, serverEntity, clientEntity interface{}) ConflictResult {
	resolution := ConflictResolution{
		EntityType:    entityType,
		EntityID:      entityID,
		ConflictType:  "update_conflict",
		ServerVersion: serverTime,
		ClientVersion: clientTime,
		ResolvedAt:    time.Now().UTC(),
	}

	// Compare timestamps (Last-Write-Wins)
	if clientTime.After(serverTime) {
		resolution.Resolution = "client_wins"
		return ConflictResult{
			Resolution: resolution,
			Winner:     clientEntity,
			ShouldSync: true,
		}
	} else if serverTime.After(clientTime) {
		resolution.Resolution = "server_wins"
		return ConflictResult{
			Resolution: resolution,
			Winner:     serverEntity,
			ShouldSync: false, // Don't sync since server is newer
		}
	} else {
		// Timestamps are equal - prefer server version for consistency
		resolution.Resolution = "server_wins"
		resolution.ConflictType = "timestamp_tie"
		return ConflictResult{
			Resolution: resolution,
			Winner:     serverEntity,
			ShouldSync: false,
		}
	}
}

// ValidateBusinessRules performs business rule validation before applying changes
func (cr *ConflictResolver) ValidateBusinessRules(entityType string, entity interface{}) []SyncError {
	var errors []SyncError

	switch entityType {
	case "product":
		if product, ok := entity.(entities.Product); ok {
			errors = append(errors, cr.validateProduct(product)...)
		}
	case "transaction":
		if transaction, ok := entity.(entities.Transaction); ok {
			errors = append(errors, cr.validateTransaction(transaction)...)
		}
	case "expense":
		if expense, ok := entity.(entities.Expense); ok {
			errors = append(errors, cr.validateExpense(expense)...)
		}
	// Add more entity validations as needed
	}

	return errors
}

// validateProduct performs business rule validation for products
func (cr *ConflictResolver) validateProduct(product entities.Product) []SyncError {
	var errors []SyncError

	if product.Sale <= 0 {
		errors = append(errors, SyncError{
			EntityType: "product",
			EntityID:   product.ID,
			Field:      "sale",
			Message:    "sale price must be greater than 0",
			ErrorType:  "validation",
		})
	}

	if product.Buy < 0 {
		errors = append(errors, SyncError{
			EntityType: "product",
			EntityID:   product.ID,
			Field:      "buy",
			Message:    "buy price cannot be negative",
			ErrorType:  "validation",
		})
	}

	if product.Stock < 0 && product.IsHaveStock {
		errors = append(errors, SyncError{
			EntityType: "product",
			EntityID:   product.ID,
			Field:      "stock",
			Message:    "stock cannot be negative for products with stock tracking",
			ErrorType:  "validation",
		})
	}

	return errors
}

// validateTransaction performs business rule validation for transactions
func (cr *ConflictResolver) validateTransaction(transaction entities.Transaction) []SyncError {
	var errors []SyncError

	if transaction.TotalPrice < 0 {
		errors = append(errors, SyncError{
			EntityType: "transaction",
			EntityID:   transaction.ID,
			Field:      "total_price",
			Message:    "total price cannot be negative",
			ErrorType:  "validation",
		})
	}

	if transaction.Discount < 0 {
		errors = append(errors, SyncError{
			EntityType: "transaction",
			EntityID:   transaction.ID,
			Field:      "discount",
			Message:    "discount cannot be negative",
			ErrorType:  "validation",
		})
	}

	if transaction.DiscountPercentage < 0 || transaction.DiscountPercentage > 100 {
		errors = append(errors, SyncError{
			EntityType: "transaction",
			EntityID:   transaction.ID,
			Field:      "discount_percentage",
			Message:    "discount percentage must be between 0 and 100",
			ErrorType:  "validation",
		})
	}

	return errors
}

// validateExpense performs business rule validation for expenses
func (cr *ConflictResolver) validateExpense(expense entities.Expense) []SyncError {
	var errors []SyncError

	if expense.Nominal <= 0 {
		errors = append(errors, SyncError{
			EntityType: "expense",
			EntityID:   expense.ID,
			Field:      "nominal",
			Message:    "expense nominal must be greater than 0",
			ErrorType:  "validation",
		})
	}

	return errors
}