package usecases

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/terminator791/t-pos/internal/domain/entities"
	"github.com/terminator791/t-pos/internal/domain/repositories"
	"gorm.io/gorm"
)

// TransactionUseCase handles transaction-related business logic
type TransactionUseCase struct {
	db              *gorm.DB
	transactionRepo repositories.TransactionRepository
	productRepo     repositories.ProductRepository
	shopRepo        repositories.ShopRepository
	userRepo        repositories.UserRepository
	paymentRepo     repositories.PaymentRepository
	expenseRepo     repositories.ExpenseRepository
	historyRepo     repositories.HistoryRepository
	receiptRepo     repositories.ReceiptRepository
}

// NewTransactionUseCase creates a new TransactionUseCase
func NewTransactionUseCase(
	db *gorm.DB,
	transactionRepo repositories.TransactionRepository,
	productRepo repositories.ProductRepository,
	shopRepo repositories.ShopRepository,
	userRepo repositories.UserRepository,
	paymentRepo repositories.PaymentRepository,
	expenseRepo repositories.ExpenseRepository,
	historyRepo repositories.HistoryRepository,
	receiptRepo repositories.ReceiptRepository,
) *TransactionUseCase {
	return &TransactionUseCase{
		db:              db,
		transactionRepo: transactionRepo,
		productRepo:     productRepo,
		shopRepo:        shopRepo,
		userRepo:        userRepo,
		paymentRepo:     paymentRepo,
		expenseRepo:     expenseRepo,
		historyRepo:     historyRepo,
		receiptRepo:     receiptRepo,
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

	// Start database transaction
	tx := uc.db.Begin()
	if tx.Error != nil {
		return nil, tx.Error
	}

	// Ensure rollback on error
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Validate shop exists
	_, err := uc.shopRepo.GetByID(ctx, req.ShopID)
	if err != nil {
		tx.Rollback()
		return nil, errors.New("invalid shop ID")
	}

	// Validate cashier exists
	_, err = uc.userRepo.GetByID(ctx, req.CashierID)
	if err != nil {
		tx.Rollback()
		return nil, errors.New("invalid cashier ID")
	}

	// Calculate total and validate products
	total, transactionProducts, err := uc.calculateTotalAndValidateProducts(ctx, req.Items)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	// Apply discount
	total = total - req.Discount

	if total <= 0 {
		tx.Rollback()
		return nil, errors.New("transaction total must be greater than 0")
	}

	// Create transaction (pending status) with row lock
	transaction := &entities.Transaction{
		ShopID:              req.ShopID,
		CashierID:           req.CashierID,
		CustomerName:        &req.CustomerName,
		Status:              entities.TransactionStatusPending,
		TotalPrice:          total,
		Discount:            req.Discount,
		TransactionProducts: transactionProducts,
	}

	err = tx.WithContext(ctx).Create(transaction).Error
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	// Create payment record (pending status)
	payment := &entities.Payment{
		ShopID:        req.ShopID,
		TransactionID: transaction.ID,
		Status:        entities.PaymentStatusPending,
		Total:         total,
	}

	err = tx.WithContext(ctx).Create(payment).Error
	if err != nil {
		tx.Rollback()
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

	err = tx.WithContext(ctx).Create(expense).Error
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
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
	// Start database transaction
	tx := uc.db.Begin()
	if tx.Error != nil {
		return nil, tx.Error
	}

	// Ensure rollback on error
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Get transaction with row lock to prevent race conditions
	var transaction entities.Transaction
	err := tx.WithContext(ctx).Set("gorm:query_option", "FOR UPDATE").Where("id = ?", transactionID).First(&transaction).Error
	if err != nil {
		tx.Rollback()
		return nil, errors.New("transaction not found")
	}

	if transaction.Status != entities.TransactionStatusPending {
		tx.Rollback()
		return nil, errors.New("transaction is not in pending status")
	}

	// Check if amount is sufficient
	if amount < transaction.TotalPrice {
		tx.Rollback()
		return &PayTransactionResponse{
			Transaction: &transaction,
			Success:     false,
		}, errors.New("insufficient payment amount")
	}

	// Calculate change
	change := amount - transaction.TotalPrice

	// Update transaction
	transaction.Status = entities.TransactionStatusCompleted
	transaction.Amount = int64(amount)
	transaction.Change = &change

	err = tx.WithContext(ctx).Save(&transaction).Error
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	// Get and update payment status with row lock
	var payment entities.Payment
	err = tx.WithContext(ctx).Set("gorm:query_option", "FOR UPDATE").Where("transaction_id = ?", transactionID).First(&payment).Error
	if err != nil {
		tx.Rollback()
		return nil, errors.New("payment record not found")
	}

	payment.Status = entities.PaymentStatusCompleted
	err = tx.WithContext(ctx).Save(&payment).Error
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	// Update related expense status
	var expense entities.Expense
	err = tx.WithContext(ctx).Set("gorm:query_option", "FOR UPDATE").Where("shop_id = ? AND status = ?", transaction.ShopID, entities.ExpenseStatusPending).First(&expense).Error
	if err == nil {
		expense.Status = entities.ExpenseStatusCompleted
		err = tx.WithContext(ctx).Save(&expense).Error
		if err != nil {
			tx.Rollback()
			return nil, err
		}
	}

	// Create history record
	history := &entities.History{
		ShopID:        transaction.ShopID,
		TransactionID: transaction.ID,
	}

	err = tx.WithContext(ctx).Create(history).Error
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	// Create receipt record
	receipt := &entities.Receipt{
		ShopID:     transaction.ShopID,
		PaymentsID: payment.ID,
	}

	err = tx.WithContext(ctx).Create(receipt).Error
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	return &PayTransactionResponse{
		Transaction: &transaction,
		Payment:     &payment,
		Change:      change,
		Success:     true,
	}, nil
}

// CancelTransaction cancels a pending transaction
func (uc *TransactionUseCase) CancelTransaction(ctx context.Context, transactionID uuid.UUID) error {
	// Start database transaction
	tx := uc.db.Begin()
	if tx.Error != nil {
		return tx.Error
	}

	// Ensure rollback on error
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Get transaction with row lock to prevent race conditions
	var transaction entities.Transaction
	err := tx.WithContext(ctx).Set("gorm:query_option", "FOR UPDATE").Where("id = ?", transactionID).First(&transaction).Error
	if err != nil {
		tx.Rollback()
		return errors.New("transaction not found")
	}

	if transaction.Status != entities.TransactionStatusPending {
		tx.Rollback()
		return errors.New("only pending transactions can be cancelled")
	}

	// Update transaction status
	transaction.Status = entities.TransactionStatusCancelled
	err = tx.WithContext(ctx).Save(&transaction).Error
	if err != nil {
		tx.Rollback()
		return err
	}

	// Update payment status with row lock
	var payment entities.Payment
	err = tx.WithContext(ctx).Set("gorm:query_option", "FOR UPDATE").Where("transaction_id = ?", transactionID).First(&payment).Error
	if err == nil {
		payment.Status = entities.PaymentStatusCancelled
		err = tx.WithContext(ctx).Save(&payment).Error
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	// Update related expense status with row lock
	var expense entities.Expense
	err = tx.WithContext(ctx).Set("gorm:query_option", "FOR UPDATE").Where("shop_id = ? AND status = ?", transaction.ShopID, entities.ExpenseStatusPending).First(&expense).Error
	if err == nil {
		expense.Status = entities.ExpenseStatusCancelled
		err = tx.WithContext(ctx).Save(&expense).Error
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return err
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
