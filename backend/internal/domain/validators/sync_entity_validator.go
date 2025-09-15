package validators

import (
	"context"
	"fmt"
	"regexp"
	"time"

	"github.com/google/uuid"
	"github.com/terminator791/t-pos/internal/domain/entities"
	"gorm.io/gorm"
)

// EntityValidator defines the interface for validating entities during sync operations
type EntityValidator interface {
	ValidateForCreate(ctx context.Context, entity interface{}) error
	ValidateForUpdate(ctx context.Context, existing, incoming interface{}) error
	ValidateBusinessRules(ctx context.Context, entity interface{}) error
	ValidateForeignKeys(ctx context.Context, db *gorm.DB, entity interface{}) error
}

// SyncEntityValidator provides comprehensive validation for sync operations
type SyncEntityValidator struct {
	db *gorm.DB
}

// NewSyncEntityValidator creates a new sync entity validator
func NewSyncEntityValidator(db *gorm.DB) *SyncEntityValidator {
	return &SyncEntityValidator{db: db}
}

// Validation error types
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
	Code    string `json:"code"`
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("validation error on field '%s': %s", e.Field, e.Message)
}

// Multi-validation error
type ValidationErrors []ValidationError

func (e ValidationErrors) Error() string {
	if len(e) == 0 {
		return "no validation errors"
	}
	return fmt.Sprintf("validation failed with %d errors: %s", len(e), e[0].Message)
}

// ValidateEntity is the main entry point for entity validation
func (v *SyncEntityValidator) ValidateEntity(ctx context.Context, entity interface{}, operation string, existing interface{}) error {
	var errors ValidationErrors

	// 1. Basic field validation
	if err := v.validateBasicFields(entity); err != nil {
		if ve, ok := err.(ValidationErrors); ok {
			errors = append(errors, ve...)
		} else {
			errors = append(errors, ValidationError{
				Field:   "general",
				Message: err.Error(),
				Code:    "BASIC_VALIDATION_FAILED",
			})
		}
	}

	// 2. Business rules validation
	if err := v.ValidateBusinessRules(ctx, entity); err != nil {
		if ve, ok := err.(ValidationErrors); ok {
			errors = append(errors, ve...)
		} else {
			errors = append(errors, ValidationError{
				Field:   "business_rules",
				Message: err.Error(),
				Code:    "BUSINESS_RULE_VALIDATION_FAILED",
			})
		}
	}

	// 3. Foreign key validation (only for create/update operations)
	if operation == "create" || operation == "update" {
		if err := v.ValidateForeignKeys(ctx, v.db, entity); err != nil {
			if ve, ok := err.(ValidationErrors); ok {
				errors = append(errors, ve...)
			} else {
				errors = append(errors, ValidationError{
					Field:   "foreign_keys",
					Message: err.Error(),
					Code:    "FOREIGN_KEY_VALIDATION_FAILED",
				})
			}
		}
	}

	// 4. Operation-specific validation
	switch operation {
	case "create":
		if err := v.ValidateForCreate(ctx, entity); err != nil {
			if ve, ok := err.(ValidationErrors); ok {
				errors = append(errors, ve...)
			} else {
				errors = append(errors, ValidationError{
					Field:   "create_operation",
					Message: err.Error(),
					Code:    "CREATE_VALIDATION_FAILED",
				})
			}
		}
	case "update":
		if err := v.ValidateForUpdate(ctx, existing, entity); err != nil {
			if ve, ok := err.(ValidationErrors); ok {
				errors = append(errors, ve...)
			} else {
				errors = append(errors, ValidationError{
					Field:   "update_operation",
					Message: err.Error(),
					Code:    "UPDATE_VALIDATION_FAILED",
				})
			}
		}
	}

	if len(errors) > 0 {
		return errors
	}

	return nil
}

// ValidateForCreate validates entity for creation
func (v *SyncEntityValidator) ValidateForCreate(ctx context.Context, entity interface{}) error {
	var errors ValidationErrors

	switch e := entity.(type) {
	case entities.Product:
		if e.ID == uuid.Nil {
			errors = append(errors, ValidationError{
				Field:   "id",
				Message: "product ID is required for sync operations",
				Code:    "REQUIRED_FIELD_MISSING",
			})
		}
		if e.ShopID == uuid.Nil {
			errors = append(errors, ValidationError{
				Field:   "shop_id",
				Message: "shop ID is required",
				Code:    "REQUIRED_FIELD_MISSING",
			})
		}
	case entities.Cart:
		if e.ID == uuid.Nil {
			errors = append(errors, ValidationError{
				Field:   "id",
				Message: "cart ID is required for sync operations",
				Code:    "REQUIRED_FIELD_MISSING",
			})
		}
		if e.ProductID == uuid.Nil {
			errors = append(errors, ValidationError{
				Field:   "product_id",
				Message: "product ID is required",
				Code:    "REQUIRED_FIELD_MISSING",
			})
		}
		if e.UserID == uuid.Nil {
			errors = append(errors, ValidationError{
				Field:   "user_id",
				Message: "user ID is required",
				Code:    "REQUIRED_FIELD_MISSING",
			})
		}
	case entities.Transaction:
		if e.ID == uuid.Nil {
			errors = append(errors, ValidationError{
				Field:   "id",
				Message: "transaction ID is required for sync operations",
				Code:    "REQUIRED_FIELD_MISSING",
			})
		}
		if e.ShopID == uuid.Nil {
			errors = append(errors, ValidationError{
				Field:   "shop_id",
				Message: "shop ID is required",
				Code:    "REQUIRED_FIELD_MISSING",
			})
		}
	}

	if len(errors) > 0 {
		return errors
	}
	return nil
}

// ValidateForUpdate validates entity for update operations
func (v *SyncEntityValidator) ValidateForUpdate(ctx context.Context, existing, incoming interface{}) error {
	var errors ValidationErrors

	// Ensure entities are of the same type
	if existing != nil {
		switch e := existing.(type) {
		case entities.Product:
			if inc, ok := incoming.(entities.Product); ok {
				if e.ID != inc.ID {
					errors = append(errors, ValidationError{
						Field:   "id",
						Message: "cannot change entity ID during update",
						Code:    "IMMUTABLE_FIELD_CHANGE",
					})
				}
				if e.ShopID != inc.ShopID {
					errors = append(errors, ValidationError{
						Field:   "shop_id",
						Message: "cannot change shop ID during update",
						Code:    "IMMUTABLE_FIELD_CHANGE",
					})
				}
			}
		case entities.Cart:
			if inc, ok := incoming.(entities.Cart); ok {
				if e.ID != inc.ID {
					errors = append(errors, ValidationError{
						Field:   "id",
						Message: "cannot change entity ID during update",
						Code:    "IMMUTABLE_FIELD_CHANGE",
					})
				}
			}
		case entities.Transaction:
			if inc, ok := incoming.(entities.Transaction); ok {
				if e.ID != inc.ID {
					errors = append(errors, ValidationError{
						Field:   "id",
						Message: "cannot change entity ID during update",
						Code:    "IMMUTABLE_FIELD_CHANGE",
					})
				}
				if e.ShopID != inc.ShopID {
					errors = append(errors, ValidationError{
						Field:   "shop_id",
						Message: "cannot change shop ID during update",
						Code:    "IMMUTABLE_FIELD_CHANGE",
					})
				}
			}
		}
	}

	if len(errors) > 0 {
		return errors
	}
	return nil
}

// ValidateBusinessRules validates business-specific rules
func (v *SyncEntityValidator) ValidateBusinessRules(ctx context.Context, entity interface{}) error {
	var errors ValidationErrors

	switch e := entity.(type) {
	case entities.Product:
		// Price validation
		if e.Sale < 0 {
			errors = append(errors, ValidationError{
				Field:   "sale",
				Message: "sale price cannot be negative",
				Code:    "INVALID_PRICE",
			})
		}
		if e.Buy < 0 {
			errors = append(errors, ValidationError{
				Field:   "buy",
				Message: "buy price cannot be negative",
				Code:    "INVALID_PRICE",
			})
		}
		if e.Sale < e.Buy {
			errors = append(errors, ValidationError{
				Field:   "sale",
				Message: "sale price should not be less than buy price",
				Code:    "INVALID_PROFIT_MARGIN",
			})
		}

		// Stock validation
		if e.Stock < 0 {
			errors = append(errors, ValidationError{
				Field:   "stock",
				Message: "stock cannot be negative",
				Code:    "INVALID_STOCK",
			})
		}

		// PPN (tax) validation
		if e.PPN != nil && (*e.PPN < 0 || *e.PPN > 100) {
			errors = append(errors, ValidationError{
				Field:   "ppn",
				Message: "tax percentage must be between 0 and 100",
				Code:    "INVALID_TAX_PERCENTAGE",
			})
		}

		// Name validation
		if len(e.Name) == 0 {
			errors = append(errors, ValidationError{
				Field:   "name",
				Message: "product name is required",
				Code:    "REQUIRED_FIELD_MISSING",
			})
		}
		if len(e.Name) > 255 {
			errors = append(errors, ValidationError{
				Field:   "name",
				Message: "product name cannot exceed 255 characters",
				Code:    "FIELD_TOO_LONG",
			})
		}

		// Barcode validation (if provided)
		if e.Barcode != nil {
			if err := v.validateBarcode(*e.Barcode); err != nil {
				errors = append(errors, ValidationError{
					Field:   "barcode",
					Message: err.Error(),
					Code:    "INVALID_BARCODE",
				})
			}
		}

	case entities.Cart:
		// Quantity validation
		if e.Quantity <= 0 {
			errors = append(errors, ValidationError{
				Field:   "quantity",
				Message: "cart quantity must be positive",
				Code:    "INVALID_QUANTITY",
			})
		}
		if e.Quantity > 10000 {
			errors = append(errors, ValidationError{
				Field:   "quantity",
				Message: "cart quantity cannot exceed 10000",
				Code:    "QUANTITY_TOO_LARGE",
			})
		}

	case entities.Transaction:
		// Status validation
		if err := ValidateTransactionStatus(e.Status); err != nil {
			errors = append(errors, ValidationError{
				Field:   "status",
				Message: err.Error(),
				Code:    "INVALID_STATUS",
			})
		}

		// Amount validation
		if e.TotalPrice < 0 {
			errors = append(errors, ValidationError{
				Field:   "total_price",
				Message: "total price cannot be negative",
				Code:    "INVALID_AMOUNT",
			})
		}

		// Discount validation
		if e.Discount < 0 {
			errors = append(errors, ValidationError{
				Field:   "discount",
				Message: "discount cannot be negative",
				Code:    "INVALID_DISCOUNT",
			})
		}

		if e.DiscountPercentage < 0 || e.DiscountPercentage > 100 {
			errors = append(errors, ValidationError{
				Field:   "discount_percentage",
				Message: "discount percentage must be between 0 and 100",
				Code:    "INVALID_DISCOUNT_PERCENTAGE",
			})
		}

	case entities.Payment:
		// Status validation
		if err := ValidatePaymentStatus(e.Status); err != nil {
			errors = append(errors, ValidationError{
				Field:   "status",
				Message: err.Error(),
				Code:    "INVALID_STATUS",
			})
		}

		// Amount validation
		if e.Total <= 0 {
			errors = append(errors, ValidationError{
				Field:   "total",
				Message: "payment total must be positive",
				Code:    "INVALID_AMOUNT",
			})
		}

	case entities.Expense:
		// Status validation
		if err := ValidateExpenseStatus(e.Status); err != nil {
			errors = append(errors, ValidationError{
				Field:   "status",
				Message: err.Error(),
				Code:    "INVALID_STATUS",
			})
		}

		// Amount validation
		if e.Nominal <= 0 {
			errors = append(errors, ValidationError{
				Field:   "nominal",
				Message: "expense nominal must be positive",
				Code:    "INVALID_AMOUNT",
			})
		}

		// Label validation (if provided)
		if e.Label != nil && len(*e.Label) == 0 {
			errors = append(errors, ValidationError{
				Field:   "label",
				Message: "expense label cannot be empty if provided",
				Code:    "INVALID_LABEL",
			})
		}

	case entities.StockHistory:
		// Stock validation
		if e.Stock < 0 || e.LastStock < 0 {
			errors = append(errors, ValidationError{
				Field:   "stock",
				Message: "stock values cannot be negative",
				Code:    "INVALID_STOCK",
			})
		}

		// Date validation
		if e.StockedAt.IsZero() {
			errors = append(errors, ValidationError{
				Field:   "stocked_at",
				Message: "stocked_at date is required",
				Code:    "REQUIRED_FIELD_MISSING",
			})
		}

		// Future date validation
		if e.StockedAt.After(time.Now().Add(24 * time.Hour)) {
			errors = append(errors, ValidationError{
				Field:   "stocked_at",
				Message: "stocked_at cannot be more than 24 hours in the future",
				Code:    "INVALID_DATE",
			})
		}
	}

	if len(errors) > 0 {
		return errors
	}
	return nil
}

// ValidateForeignKeys validates that referenced entities exist
func (v *SyncEntityValidator) ValidateForeignKeys(ctx context.Context, db *gorm.DB, entity interface{}) error {
	var errors ValidationErrors

	switch e := entity.(type) {
	case entities.Product:
		// Validate shop exists
		if err := v.validateShopExists(ctx, db, e.ShopID); err != nil {
			errors = append(errors, ValidationError{
				Field:   "shop_id",
				Message: fmt.Sprintf("shop does not exist: %v", err),
				Code:    "FOREIGN_KEY_NOT_FOUND",
			})
		}

		// Validate category exists (if provided)
		if e.CatID != nil {
			if err := v.validateCategoryExists(ctx, db, *e.CatID); err != nil {
				errors = append(errors, ValidationError{
					Field:   "cat_id",
					Message: fmt.Sprintf("category does not exist: %v", err),
					Code:    "FOREIGN_KEY_NOT_FOUND",
				})
			}
		}

	case entities.Cart:
		// Validate shop exists
		if err := v.validateShopExists(ctx, db, e.ShopID); err != nil {
			errors = append(errors, ValidationError{
				Field:   "shop_id",
				Message: fmt.Sprintf("shop does not exist: %v", err),
				Code:    "FOREIGN_KEY_NOT_FOUND",
			})
		}

		// Validate product exists
		if err := v.validateProductExists(ctx, db, e.ProductID); err != nil {
			errors = append(errors, ValidationError{
				Field:   "product_id",
				Message: fmt.Sprintf("product does not exist: %v", err),
				Code:    "FOREIGN_KEY_NOT_FOUND",
			})
		}

		// Validate user exists
		if err := v.validateUserExists(ctx, db, e.UserID); err != nil {
			errors = append(errors, ValidationError{
				Field:   "user_id",
				Message: fmt.Sprintf("user does not exist: %v", err),
				Code:    "FOREIGN_KEY_NOT_FOUND",
			})
		}

	case entities.Transaction:
		// Validate shop exists
		if err := v.validateShopExists(ctx, db, e.ShopID); err != nil {
			errors = append(errors, ValidationError{
				Field:   "shop_id",
				Message: fmt.Sprintf("shop does not exist: %v", err),
				Code:    "FOREIGN_KEY_NOT_FOUND",
			})
		}

		// Validate cashier exists
		if err := v.validateUserExists(ctx, db, e.CashierID); err != nil {
			errors = append(errors, ValidationError{
				Field:   "cashier_id",
				Message: fmt.Sprintf("cashier does not exist: %v", err),
				Code:    "FOREIGN_KEY_NOT_FOUND",
			})
		}

	case entities.Payment:
		// Validate transaction exists
		if err := v.validateTransactionExists(ctx, db, e.TransactionID); err != nil {
			errors = append(errors, ValidationError{
				Field:   "transaction_id",
				Message: fmt.Sprintf("transaction does not exist: %v", err),
				Code:    "FOREIGN_KEY_NOT_FOUND",
			})
		}

	case entities.TransactionProduct:
		// Validate transaction exists
		if err := v.validateTransactionExists(ctx, db, e.TransactionID); err != nil {
			errors = append(errors, ValidationError{
				Field:   "transaction_id",
				Message: fmt.Sprintf("transaction does not exist: %v", err),
				Code:    "FOREIGN_KEY_NOT_FOUND",
			})
		}

		// Validate product exists
		if err := v.validateProductExists(ctx, db, e.ProductID); err != nil {
			errors = append(errors, ValidationError{
				Field:   "product_id",
				Message: fmt.Sprintf("product does not exist: %v", err),
				Code:    "FOREIGN_KEY_NOT_FOUND",
			})
		}

	case entities.StockHistory:
		// Validate product exists
		if err := v.validateProductExists(ctx, db, e.ProductID); err != nil {
			errors = append(errors, ValidationError{
				Field:   "product_id",
				Message: fmt.Sprintf("product does not exist: %v", err),
				Code:    "FOREIGN_KEY_NOT_FOUND",
			})
		}

	case entities.Expense:
		// Validate shop exists
		if err := v.validateShopExists(ctx, db, e.ShopID); err != nil {
			errors = append(errors, ValidationError{
				Field:   "shop_id",
				Message: fmt.Sprintf("shop does not exist: %v", err),
				Code:    "FOREIGN_KEY_NOT_FOUND",
			})
		}
	}

	if len(errors) > 0 {
		return errors
	}
	return nil
}

// Helper validation functions
func (v *SyncEntityValidator) validateShopExists(ctx context.Context, db *gorm.DB, shopID uuid.UUID) error {
	var count int64
	if err := db.WithContext(ctx).Model(&entities.Shop{}).Where("id = ?", shopID).Count(&count).Error; err != nil {
		return err
	}
	if count == 0 {
		return fmt.Errorf("shop with ID %s does not exist", shopID)
	}
	return nil
}

func (v *SyncEntityValidator) validateProductExists(ctx context.Context, db *gorm.DB, productID uuid.UUID) error {
	var count int64
	if err := db.WithContext(ctx).Model(&entities.Product{}).Where("id = ?", productID).Count(&count).Error; err != nil {
		return err
	}
	if count == 0 {
		return fmt.Errorf("product with ID %s does not exist", productID)
	}
	return nil
}

func (v *SyncEntityValidator) validateCategoryExists(ctx context.Context, db *gorm.DB, categoryID uuid.UUID) error {
	var count int64
	if err := db.WithContext(ctx).Model(&entities.Category{}).Where("id = ?", categoryID).Count(&count).Error; err != nil {
		return err
	}
	if count == 0 {
		return fmt.Errorf("category with ID %s does not exist", categoryID)
	}
	return nil
}

func (v *SyncEntityValidator) validateUserExists(ctx context.Context, db *gorm.DB, userID uuid.UUID) error {
	var count int64
	if err := db.WithContext(ctx).Model(&entities.User{}).Where("id = ?", userID).Count(&count).Error; err != nil {
		return err
	}
	if count == 0 {
		return fmt.Errorf("user with ID %s does not exist", userID)
	}
	return nil
}

func (v *SyncEntityValidator) validateTransactionExists(ctx context.Context, db *gorm.DB, transactionID uuid.UUID) error {
	var count int64
	if err := db.WithContext(ctx).Model(&entities.Transaction{}).Where("id = ?", transactionID).Count(&count).Error; err != nil {
		return err
	}
	if count == 0 {
		return fmt.Errorf("transaction with ID %s does not exist", transactionID)
	}
	return nil
}

func (v *SyncEntityValidator) validateBasicFields(entity interface{}) error {
	var errors ValidationErrors

	// Common ID validation for all entities
	switch entity.(type) {
	case entities.Product, entities.Cart, entities.Transaction, entities.Payment, entities.Expense, entities.Category, entities.Shop, entities.User, entities.TransactionProduct, entities.StockHistory, entities.Receipt, entities.History:
		// All entities should have valid UUIDs for sync operations
		// This is handled in specific validation methods
	default:
		errors = append(errors, ValidationError{
			Field:   "entity_type",
			Message: "unknown entity type for validation",
			Code:    "UNKNOWN_ENTITY_TYPE",
		})
	}

	if len(errors) > 0 {
		return errors
	}
	return nil
}

func (v *SyncEntityValidator) validateBarcode(barcode string) error {
	// Basic barcode validation - should be alphanumeric
	if len(barcode) == 0 {
		return fmt.Errorf("barcode cannot be empty")
	}
	
	// Check for valid barcode format (EAN-13, UPC-A, etc.)
	validBarcode := regexp.MustCompile(`^[0-9]{8,13}$|^[A-Za-z0-9]{8,20}$`)
	if !validBarcode.MatchString(barcode) {
		return fmt.Errorf("invalid barcode format")
	}
	
	return nil
}