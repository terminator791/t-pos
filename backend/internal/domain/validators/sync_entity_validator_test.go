package validators_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/terminator791/t-pos/internal/domain/entities"
	"github.com/terminator791/t-pos/internal/domain/validators"
)

func TestSyncEntityValidatorBasicValidation(t *testing.T) {
	// Create a validator without database (for basic validation tests)
	validator := &validators.SyncEntityValidator{}
	ctx := context.Background()

	t.Run("ValidProductBusinessRules", func(t *testing.T) {
		product := entities.Product{
			ID:     uuid.New(),
			ShopID: uuid.New(),
			Name:   "Valid Product",
			Sale:   100.0,
			Buy:    80.0,
			Stock:  5,
		}

		err := validator.ValidateBusinessRules(ctx, product)
		assert.NoError(t, err)
	})

	t.Run("InvalidProductPricing", func(t *testing.T) {
		product := entities.Product{
			ID:     uuid.New(),
			ShopID: uuid.New(),
			Name:   "Invalid Product",
			Sale:   50.0, // Sale price less than buy price
			Buy:    100.0,
			Stock:  5,
		}

		err := validator.ValidateBusinessRules(ctx, product)
		assert.Error(t, err)

		validationErrors, ok := err.(validators.ValidationErrors)
		assert.True(t, ok)

		// Should have validation error for invalid profit margin
		found := false
		for _, ve := range validationErrors {
			if ve.Code == "INVALID_PROFIT_MARGIN" {
				found = true
				break
			}
		}
		assert.True(t, found, "Expected INVALID_PROFIT_MARGIN error")
	})

	t.Run("InvalidProductStock", func(t *testing.T) {
		product := entities.Product{
			ID:     uuid.New(),
			ShopID: uuid.New(),
			Name:   "Invalid Stock Product",
			Sale:   100.0,
			Buy:    80.0,
			Stock:  -5, // Negative stock
		}

		err := validator.ValidateBusinessRules(ctx, product)
		assert.Error(t, err)

		validationErrors, ok := err.(validators.ValidationErrors)
		assert.True(t, ok)

		// Should have validation error for invalid stock
		found := false
		for _, ve := range validationErrors {
			if ve.Code == "INVALID_STOCK" {
				found = true
				break
			}
		}
		assert.True(t, found, "Expected INVALID_STOCK error")
	})

	t.Run("ValidCartBusinessRules", func(t *testing.T) {
		cart := entities.Cart{
			ID:        uuid.New(),
			ShopID:    uuid.New(),
			ProductID: uuid.New(),
			UserID:    uuid.New(),
			Quantity:  2,
		}

		err := validator.ValidateBusinessRules(ctx, cart)
		assert.NoError(t, err)
	})

	t.Run("InvalidCartQuantity", func(t *testing.T) {
		cart := entities.Cart{
			ID:        uuid.New(),
			ShopID:    uuid.New(),
			ProductID: uuid.New(),
			UserID:    uuid.New(),
			Quantity:  0, // Invalid quantity
		}

		err := validator.ValidateBusinessRules(ctx, cart)
		assert.Error(t, err)

		validationErrors, ok := err.(validators.ValidationErrors)
		assert.True(t, ok)

		// Should have validation error for invalid quantity
		found := false
		for _, ve := range validationErrors {
			if ve.Code == "INVALID_QUANTITY" {
				found = true
				break
			}
		}
		assert.True(t, found, "Expected INVALID_QUANTITY error")
	})

	t.Run("ValidTransactionBusinessRules", func(t *testing.T) {
		transaction := entities.Transaction{
			ID:                 uuid.New(),
			ShopID:             uuid.New(),
			CashierID:          uuid.New(),
			Status:             entities.TransactionStatusCompleted,
			TotalPrice:         100.0,
			Discount:           10.0,
			DiscountPercentage: 5.0,
		}

		err := validator.ValidateBusinessRules(ctx, transaction)
		assert.NoError(t, err)
	})

	t.Run("InvalidTransactionDiscount", func(t *testing.T) {
		transaction := entities.Transaction{
			ID:                 uuid.New(),
			ShopID:             uuid.New(),
			CashierID:          uuid.New(),
			Status:             entities.TransactionStatusCompleted,
			TotalPrice:         100.0,
			Discount:           -10.0, // Negative discount
			DiscountPercentage: 150.0, // Invalid percentage
		}

		err := validator.ValidateBusinessRules(ctx, transaction)
		assert.Error(t, err)

		validationErrors, ok := err.(validators.ValidationErrors)
		assert.True(t, ok)

		// Should have validation errors for discount
		assert.True(t, len(validationErrors) >= 2, "Expected at least 2 validation errors")
	})

	t.Run("ValidPaymentBusinessRules", func(t *testing.T) {
		payment := entities.Payment{
			ID:            uuid.New(),
			ShopID:        uuid.New(),
			TransactionID: uuid.New(),
			Status:        entities.PaymentStatusCompleted,
			Total:         100.0,
		}

		err := validator.ValidateBusinessRules(ctx, payment)
		assert.NoError(t, err)
	})

	t.Run("InvalidPaymentAmount", func(t *testing.T) {
		payment := entities.Payment{
			ID:            uuid.New(),
			ShopID:        uuid.New(),
			TransactionID: uuid.New(),
			Status:        entities.PaymentStatusCompleted,
			Total:         -100.0, // Invalid amount
		}

		err := validator.ValidateBusinessRules(ctx, payment)
		assert.Error(t, err)

		validationErrors, ok := err.(validators.ValidationErrors)
		assert.True(t, ok)

		// Should have validation error for invalid amount
		found := false
		for _, ve := range validationErrors {
			if ve.Code == "INVALID_AMOUNT" {
				found = true
				break
			}
		}
		assert.True(t, found, "Expected INVALID_AMOUNT error")
	})

	t.Run("BarcodeValidation", func(t *testing.T) {
		validator := &validators.SyncEntityValidator{}

		// Valid barcode
		validBarcode := "1234567890123"
		product := entities.Product{
			ID:      uuid.New(),
			ShopID:  uuid.New(),
			Name:    "Barcode Product",
			Barcode: &validBarcode,
			Sale:    100.0,
			Buy:     80.0,
			Stock:   5,
		}

		err := validator.ValidateBusinessRules(ctx, product)
		assert.NoError(t, err)

		// Invalid barcode
		invalidBarcode := "invalid-barcode-@#$"
		product.Barcode = &invalidBarcode

		err = validator.ValidateBusinessRules(ctx, product)
		assert.Error(t, err)

		validationErrors, ok := err.(validators.ValidationErrors)
		assert.True(t, ok)

		// Should have validation error for invalid barcode
		found := false
		for _, ve := range validationErrors {
			if ve.Code == "INVALID_BARCODE" {
				found = true
				break
			}
		}
		assert.True(t, found, "Expected INVALID_BARCODE error")
	})
}
