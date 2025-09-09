package usecases

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/terminator791/t-pos/internal/domain/entities"
	"github.com/terminator791/t-pos/internal/domain/repositories"
)

// CheckoutUseCase handles checkout process business logic
type CheckoutUseCase struct {
	transactionRepo repositories.TransactionRepository
	productRepo     repositories.ProductRepository
	shopRepo        repositories.ShopRepository
	userRepo        repositories.UserRepository
	paymentRepo     repositories.PaymentRepository
}

// NewCheckoutUseCase creates a new CheckoutUseCase
func NewCheckoutUseCase(
	transactionRepo repositories.TransactionRepository,
	productRepo repositories.ProductRepository,
	shopRepo repositories.ShopRepository,
	userRepo repositories.UserRepository,
	paymentRepo repositories.PaymentRepository,
) *CheckoutUseCase {
	return &CheckoutUseCase{
		transactionRepo: transactionRepo,
		productRepo:     productRepo,
		shopRepo:        shopRepo,
		userRepo:        userRepo,
		paymentRepo:     paymentRepo,
	}
}

// CheckoutRequest represents the request to create a checkout
type CheckoutRequest struct {
	ShopID         uuid.UUID                `json:"shop_id"`
	CashierID      uuid.UUID                `json:"cashier_id"`
	CustomerID     *uuid.UUID               `json:"customer_id"` // Optional customer
	Items          []CheckoutItemRequest    `json:"items"`
	PaymentMethod  string                   `json:"payment_method"` // cash, card, digital
	Discount       float64                  `json:"discount"`
	DiscountPercentage float64              `json:"discount_percentage"`
	AdditionalCost float64                  `json:"additional_cost"`
	AmountPaid     int64                    `json:"amount_paid"`
}

// CheckoutItemRequest represents an item in the checkout request
type CheckoutItemRequest struct {
	ProductID uuid.UUID `json:"product_id"`
	Quantity  int       `json:"quantity"`
}

// CheckoutResponse represents the response after successful checkout
type CheckoutResponse struct {
	Transaction *entities.Transaction `json:"transaction"`
	Payment     *entities.Payment     `json:"payment"`
	Change      float64              `json:"change"`
	Message     string               `json:"message"`
}

// ProcessCheckout processes the entire checkout flow
func (uc *CheckoutUseCase) ProcessCheckout(ctx context.Context, req *CheckoutRequest) (*CheckoutResponse, error) {
	// Validate request
	if err := uc.validateCheckoutRequest(ctx, req); err != nil {
		return nil, err
	}

	// Calculate totals and validate stock
	subtotal, transactionProducts, err := uc.calculateTotalsAndValidateStock(ctx, req.Items)
	if err != nil {
		return nil, err
	}

	// Create transaction
	transaction := &entities.Transaction{
		ShopID:             req.ShopID,
		CashierID:          req.CashierID,
		UserID:             req.CustomerID,
		Discount:           req.Discount,
		DiscountPercentage: req.DiscountPercentage,
		AdditionalCost:     req.AdditionalCost,
		Status:             entities.TransactionStatusPending,
		TotalPrice:         subtotal,
		Amount:             req.AmountPaid,
		TransactionProducts: transactionProducts,
	}

	// Apply discounts and additional costs
	transaction.CalculateTotal()

	// Calculate change
	totalPaid := float64(req.AmountPaid)
	change := totalPaid - transaction.TotalPrice
	if change < 0 {
		return nil, errors.New("insufficient payment amount")
	}
	transaction.Change = &change

	// Save transaction
	if err := uc.transactionRepo.Create(ctx, transaction); err != nil {
		return nil, fmt.Errorf("failed to create transaction: %w", err)
	}

	// Create payment record
	payment := &entities.Payment{
		ShopID:        req.ShopID,
		UserID:        req.CustomerID,
		TransactionID: transaction.ID,
		Status:        entities.PaymentStatusPending,
		Total:         transaction.TotalPrice,
	}

	if err := uc.paymentRepo.Create(ctx, payment); err != nil {
		return nil, fmt.Errorf("failed to create payment: %w", err)
	}

	// Update stock for sold items
	if err := uc.updateProductStock(ctx, req.Items); err != nil {
		return nil, fmt.Errorf("failed to update stock: %w", err)
	}

	return &CheckoutResponse{
		Transaction: transaction,
		Payment:     payment,
		Change:      change,
		Message:     "Checkout processed successfully",
	}, nil
}

// CompletePayment completes the payment and transaction
func (uc *CheckoutUseCase) CompletePayment(ctx context.Context, transactionID uuid.UUID) error {
	// Get transaction
	transaction, err := uc.transactionRepo.GetByID(ctx, transactionID)
	if err != nil {
		return fmt.Errorf("transaction not found: %w", err)
	}

	if transaction.Status == entities.TransactionStatusCompleted {
		return errors.New("transaction is already completed")
	}

	// Update transaction status
	transaction.Status = entities.TransactionStatusCompleted
	if err := uc.transactionRepo.Update(ctx, transaction); err != nil {
		return fmt.Errorf("failed to update transaction: %w", err)
	}

	// Update payment status
	for _, payment := range transaction.Payments {
		payment.Status = entities.PaymentStatusCompleted
		if err := uc.paymentRepo.Update(ctx, &payment); err != nil {
			return fmt.Errorf("failed to update payment: %w", err)
		}
	}

	return nil
}

// CancelTransaction cancels a pending transaction
func (uc *CheckoutUseCase) CancelTransaction(ctx context.Context, transactionID uuid.UUID) error {
	transaction, err := uc.transactionRepo.GetByID(ctx, transactionID)
	if err != nil {
		return fmt.Errorf("transaction not found: %w", err)
	}

	if transaction.Status == entities.TransactionStatusCompleted {
		return errors.New("cannot cancel completed transaction")
	}

	// Restore stock for cancelled items
	if err := uc.restoreProductStock(ctx, transaction.TransactionProducts); err != nil {
		return fmt.Errorf("failed to restore stock: %w", err)
	}

	// Update transaction status
	transaction.Status = entities.TransactionStatusCancelled
	if err := uc.transactionRepo.Update(ctx, transaction); err != nil {
		return fmt.Errorf("failed to update transaction: %w", err)
	}

	// Update payment status
	for _, payment := range transaction.Payments {
		payment.Status = entities.PaymentStatusCancelled
		if err := uc.paymentRepo.Update(ctx, &payment); err != nil {
			return fmt.Errorf("failed to update payment: %w", err)
		}
	}

	return nil
}

// validateCheckoutRequest validates the checkout request
func (uc *CheckoutUseCase) validateCheckoutRequest(ctx context.Context, req *CheckoutRequest) error {
	if req.ShopID == uuid.Nil {
		return errors.New("shop ID is required")
	}

	if req.CashierID == uuid.Nil {
		return errors.New("cashier ID is required")
	}

	if len(req.Items) == 0 {
		return errors.New("at least one item is required")
	}

	// Validate shop exists
	_, err := uc.shopRepo.GetByID(ctx, req.ShopID)
	if err != nil {
		return errors.New("invalid shop ID")
	}

	// Validate cashier exists
	_, err = uc.userRepo.GetByID(ctx, req.CashierID)
	if err != nil {
		return errors.New("invalid cashier ID")
	}

	// Validate customer if provided
	if req.CustomerID != nil {
		_, err = uc.userRepo.GetByID(ctx, *req.CustomerID)
		if err != nil {
			return errors.New("invalid customer ID")
		}
	}

	return nil
}

// calculateTotalsAndValidateStock calculates subtotal and validates stock availability
func (uc *CheckoutUseCase) calculateTotalsAndValidateStock(ctx context.Context, items []CheckoutItemRequest) (float64, []entities.TransactionProduct, error) {
	var subtotal float64
	var transactionProducts []entities.TransactionProduct

	for _, item := range items {
		if item.Quantity <= 0 {
			return 0, nil, errors.New("item quantity must be greater than 0")
		}

		product, err := uc.productRepo.GetByID(ctx, item.ProductID)
		if err != nil {
			return 0, nil, fmt.Errorf("invalid product ID: %s", item.ProductID)
		}

		// Check stock availability if product has stock tracking
		if product.IsHaveStock && product.Stock < item.Quantity {
			return 0, nil, fmt.Errorf("insufficient stock for product %s. Available: %d, Requested: %d", 
				product.Name, product.Stock, item.Quantity)
		}

		transactionProduct := entities.TransactionProduct{
			ProductID:  item.ProductID,
			Quantity:   item.Quantity,
			UnitPrice:  product.Sale,
		}
		transactionProduct.CalculateTotalPrice()

		transactionProducts = append(transactionProducts, transactionProduct)
		subtotal += transactionProduct.TotalPrice
	}

	return subtotal, transactionProducts, nil
}

// updateProductStock updates product stock after successful checkout
func (uc *CheckoutUseCase) updateProductStock(ctx context.Context, items []CheckoutItemRequest) error {
	for _, item := range items {
		product, err := uc.productRepo.GetByID(ctx, item.ProductID)
		if err != nil {
			return err
		}

		// Only update stock if product has stock tracking
		if product.IsHaveStock {
			product.Stock -= item.Quantity
			if err := uc.productRepo.Update(ctx, product); err != nil {
				return err
			}
		}
	}
	return nil
}

// restoreProductStock restores product stock after transaction cancellation
func (uc *CheckoutUseCase) restoreProductStock(ctx context.Context, transactionProducts []entities.TransactionProduct) error {
	for _, tp := range transactionProducts {
		product, err := uc.productRepo.GetByID(ctx, tp.ProductID)
		if err != nil {
			return err
		}

		// Only restore stock if product has stock tracking
		if product.IsHaveStock {
			product.Stock += tp.Quantity
			if err := uc.productRepo.Update(ctx, product); err != nil {
				return err
			}
		}
	}
	return nil
}

// GetTransactionWithDetails gets transaction with all related data
func (uc *CheckoutUseCase) GetTransactionWithDetails(ctx context.Context, transactionID uuid.UUID) (*entities.Transaction, error) {
	return uc.transactionRepo.GetByID(ctx, transactionID)
}

// GetTodaysTransactions gets today's transactions for a shop
func (uc *CheckoutUseCase) GetTodaysTransactions(ctx context.Context, shopID uuid.UUID) ([]*entities.Transaction, error) {
	return uc.transactionRepo.GetTodaysTransactions(ctx, shopID)
}
