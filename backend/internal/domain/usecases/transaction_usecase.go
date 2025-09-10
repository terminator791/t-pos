package usecases

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/terminator791/t-pos/internal/domain/entities"
	"github.com/terminator791/t-pos/internal/domain/repositories"
)

// TransactionUseCase handles transaction-related business logic
type TransactionUseCase struct {
	transactionRepo repositories.TransactionRepository
	productRepo     repositories.ProductRepository
	shopRepo        repositories.ShopRepository
	userRepo        repositories.UserRepository
	paymentRepo     repositories.PaymentRepository
	expenseRepo     repositories.ExpenseRepository
}

// NewTransactionUseCase creates a new TransactionUseCase
func NewTransactionUseCase(
	transactionRepo repositories.TransactionRepository,
	productRepo repositories.ProductRepository,
	shopRepo repositories.ShopRepository,
	userRepo repositories.UserRepository,
	paymentRepo repositories.PaymentRepository,
	expenseRepo repositories.ExpenseRepository,
) *TransactionUseCase {
	return &TransactionUseCase{
		transactionRepo: transactionRepo,
		productRepo:     productRepo,
		shopRepo:        shopRepo,
		userRepo:        userRepo,
		paymentRepo:     paymentRepo,
		expenseRepo:     expenseRepo,
	}
}

// CreateTransactionRequest represents the request to create a transaction
type CreateTransactionRequest struct {
	ShopID       uuid.UUID              `json:"shop_id"`
	CashierID    uuid.UUID              `json:"cashier_id"`
	CustomerName string                 `json:"customer_name"`
	Items        []CreateTransactionItem `json:"items"`
	Discount     float64                `json:"discount"`
}

// CreateTransactionItem represents an item in the transaction
type CreateTransactionItem struct {
	ProductID uuid.UUID `json:"product_id"`
	Quantity  int       `json:"quantity"`
}

// CreateTransactionResponse represents the response after creating a transaction
type CreateTransactionResponse struct {
	Transaction *entities.Transaction `json:"transaction"`
	Payment     *entities.Payment     `json:"payment"`
	Expense     *entities.Expense     `json:"expense"`
}

// PayTransactionResponse represents the response after paying a transaction
type PayTransactionResponse struct {
	Transaction *entities.Transaction `json:"transaction"`
	Payment     *entities.Payment     `json:"payment"`
	Change      float64              `json:"change"`
	Success     bool                 `json:"success"`
}

// CreateTransaction creates a new transaction with all pending records
func (uc *TransactionUseCase) CreateTransaction(ctx context.Context, req *CreateTransactionRequest) (*CreateTransactionResponse, error) {
	// Validate request
	if req.CustomerName == "" {
		return nil, errors.New("customer name is required")
	}
	if len(req.Items) == 0 {
		return nil, errors.New("at least one item is required")
	}

	// Validate shop exists
	_, err := uc.shopRepo.GetByID(ctx, req.ShopID)
	if err != nil {
		return nil, errors.New("invalid shop ID")
	}

	// Validate cashier exists
	_, err = uc.userRepo.GetByID(ctx, req.CashierID)
	if err != nil {
		return nil, errors.New("invalid cashier ID")
	}

	// Calculate total and validate products
	total, transactionProducts, err := uc.calculateTotalAndValidateProducts(ctx, req.Items)
	if err != nil {
		return nil, err
	}

	// Apply discount
	total = total - req.Discount

	if total <= 0 {
		return nil, errors.New("transaction total must be greater than 0")
	}

	// Create transaction (pending status)
	transaction := &entities.Transaction{
		ShopID:              req.ShopID,
		CashierID:           req.CashierID,
		CustomerName:        &req.CustomerName,
		Status:              entities.TransactionStatusPending,
		TotalPrice:          total,
		Discount:            req.Discount,
		TransactionProducts: transactionProducts,
	}

	err = uc.transactionRepo.Create(ctx, transaction)
	if err != nil {
		return nil, err
	}

	// Create payment record (pending status)
	payment := &entities.Payment{
		ShopID:        req.ShopID,
		TransactionID: transaction.ID,
		Status:        entities.PaymentStatusPending,
		Total:         total,
	}

	err = uc.paymentRepo.Create(ctx, payment)
	if err != nil {
		return nil, err
	}

	// Create expense record (pending status)
	expense := &entities.Expense{
		ShopID:  req.ShopID,
		Nominal: total,
		Status:  entities.ExpenseStatusPending,
		Date:    time.Now(),
		Label:   stringPtr("Transaction Expense"),
		Desc:    stringPtr("Expense for transaction " + transaction.ID.String()),
	}

	err = uc.expenseRepo.Create(ctx, expense)
	if err != nil {
		return nil, err
	}

	return &CreateTransactionResponse{
		Transaction: transaction,
		Payment:     payment,
		Expense:     expense,
	}, nil
}

// PayTransaction processes payment for a transaction
func (uc *TransactionUseCase) PayTransaction(ctx context.Context, transactionID uuid.UUID, amount float64) (*PayTransactionResponse, error) {
	// Get transaction
	transaction, err := uc.transactionRepo.GetByID(ctx, transactionID)
	if err != nil {
		return nil, errors.New("transaction not found")
	}

	if transaction.Status != entities.TransactionStatusPending {
		return nil, errors.New("transaction is not in pending status")
	}

	// Check if amount is sufficient
	if amount < transaction.TotalPrice {
		return &PayTransactionResponse{
			Transaction: transaction,
			Success:     false,
		}, errors.New("insufficient payment amount")
	}

	// Calculate change
	change := amount - transaction.TotalPrice

	// Update transaction
	transaction.Status = entities.TransactionStatusCompleted
	transaction.Amount = int64(amount)
	transaction.Change = &change

	err = uc.transactionRepo.Update(ctx, transaction)
	if err != nil {
		return nil, err
	}

	// Update payment status
	payments, err := uc.paymentRepo.GetByTransactionID(ctx, transactionID)
	if err != nil || len(payments) == 0 {
		return nil, errors.New("payment record not found")
	}

	payment := payments[0]
	payment.Status = entities.PaymentStatusCompleted

	err = uc.paymentRepo.Update(ctx, payment)
	if err != nil {
		return nil, err
	}

	return &PayTransactionResponse{
		Transaction: transaction,
		Payment:     payment,
		Change:      change,
		Success:     true,
	}, nil
}

// CancelTransaction cancels a pending transaction
func (uc *TransactionUseCase) CancelTransaction(ctx context.Context, transactionID uuid.UUID) error {
	// Get transaction
	transaction, err := uc.transactionRepo.GetByID(ctx, transactionID)
	if err != nil {
		return errors.New("transaction not found")
	}

	if transaction.Status != entities.TransactionStatusPending {
		return errors.New("only pending transactions can be cancelled")
	}

	// Update transaction status
	transaction.Status = entities.TransactionStatusCancelled

	err = uc.transactionRepo.Update(ctx, transaction)
	if err != nil {
		return err
	}

	// Update payment status
	payments, err := uc.paymentRepo.GetByTransactionID(ctx, transactionID)
	if err == nil && len(payments) > 0 {
		payment := payments[0]
		payment.Status = entities.PaymentStatusCancelled
		uc.paymentRepo.Update(ctx, payment)
	}

	// Update expense status
	expenses, err := uc.expenseRepo.GetByShopID(ctx, transaction.ShopID)
	if err == nil {
		for _, expense := range expenses {
			if expense.Status == entities.ExpenseStatusPending {
				expense.Status = entities.ExpenseStatusCancelled
				uc.expenseRepo.Update(ctx, expense)
				break
			}
		}
	}

	return nil
}

// GetTransaction retrieves a transaction by ID
func (uc *TransactionUseCase) GetTransaction(ctx context.Context, transactionID uuid.UUID) (*entities.Transaction, error) {
	return uc.transactionRepo.GetByID(ctx, transactionID)
}

// calculateTotalAndValidateProducts calculates the total and validates all products
func (uc *TransactionUseCase) calculateTotalAndValidateProducts(ctx context.Context, items []CreateTransactionItem) (float64, []entities.TransactionProduct, error) {
	var total float64
	var transactionProducts []entities.TransactionProduct

	for _, item := range items {
		// Validate product exists
		product, err := uc.productRepo.GetByID(ctx, item.ProductID)
		if err != nil {
			return 0, nil, errors.New("product not found: " + item.ProductID.String())
		}

		if item.Quantity <= 0 {
			return 0, nil, errors.New("quantity must be greater than 0")
		}

		// Calculate item total
		itemTotal := product.Sale * float64(item.Quantity)
		total += itemTotal

		// Create transaction product
		transactionProduct := entities.TransactionProduct{
			ProductID:  item.ProductID,
			Quantity:   item.Quantity,
			UnitPrice:  product.Sale,
			TotalPrice: itemTotal,
		}

		transactionProducts = append(transactionProducts, transactionProduct)
	}

	return total, transactionProducts, nil
}

// stringPtr returns a pointer to the given string
func stringPtr(s string) *string {
	return &s
}
