package usecases

import (
	"context"
	"errors"

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
}

// NewTransactionUseCase creates a new TransactionUseCase
func NewTransactionUseCase(
	transactionRepo repositories.TransactionRepository,
	productRepo repositories.ProductRepository,
	shopRepo repositories.ShopRepository,
	userRepo repositories.UserRepository,
	paymentRepo repositories.PaymentRepository,
) *TransactionUseCase {
	return &TransactionUseCase{
		transactionRepo: transactionRepo,
		productRepo:     productRepo,
		shopRepo:        shopRepo,
		userRepo:        userRepo,
		paymentRepo:     paymentRepo,
	}
}

// CreateTransaction creates a new transaction
func (uc *TransactionUseCase) CreateTransaction(ctx context.Context, transaction *entities.Transaction) error {
	if transaction.TotalPrice <= 0 {
		return errors.New("transaction total price must be greater than 0")
	}

	// Check if shop exists
	_, err := uc.shopRepo.GetByID(ctx, transaction.ShopID)
	if err != nil {
		return errors.New("invalid shop ID")
	}

	// Check if cashier exists
	_, err = uc.userRepo.GetByID(ctx, transaction.CashierID)
	if err != nil {
		return errors.New("invalid cashier ID")
	}

	// Check if customer exists (optional)
	if transaction.UserID != nil {
		_, err = uc.userRepo.GetByID(ctx, *transaction.UserID)
		if err != nil {
			return errors.New("invalid customer ID")
		}
	}

	// Set default status if not provided
	if transaction.Status == "" {
		transaction.Status = entities.TransactionStatusPending
	}

	return uc.transactionRepo.Create(ctx, transaction)
}

// GetTransaction retrieves a transaction by ID
func (uc *TransactionUseCase) GetTransaction(ctx context.Context, id uuid.UUID) (*entities.Transaction, error) {
	return uc.transactionRepo.GetByID(ctx, id)
}

// GetTransactionsByShop retrieves transactions by shop ID
func (uc *TransactionUseCase) GetTransactionsByShop(ctx context.Context, shopID uint) ([]*entities.Transaction, error) {
	return uc.transactionRepo.GetByShopID(ctx, shopID)
}

// GetTransactionsByCashier retrieves transactions by cashier ID
func (uc *TransactionUseCase) GetTransactionsByCashier(ctx context.Context, cashierID uuid.UUID) ([]*entities.Transaction, error) {
	return uc.transactionRepo.GetByCashierID(ctx, cashierID)
}

// GetTodaysTransactions retrieves today's transactions for a shop
func (uc *TransactionUseCase) GetTodaysTransactions(ctx context.Context, shopID uint) ([]*entities.Transaction, error) {
	return uc.transactionRepo.GetTodaysTransactions(ctx, shopID)
}

// UpdateTransactionStatus updates the status of a transaction
func (uc *TransactionUseCase) UpdateTransactionStatus(ctx context.Context, id uuid.UUID, status entities.TransactionStatus) error {
	transaction, err := uc.transactionRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if transaction == nil {
		return errors.New("transaction not found")
	}

	transaction.Status = status
	return uc.transactionRepo.Update(ctx, transaction)
}

// CompleteTransaction completes a transaction
func (uc *TransactionUseCase) CompleteTransaction(ctx context.Context, id uuid.UUID) error {
	transaction, err := uc.transactionRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if transaction == nil {
		return errors.New("transaction not found")
	}

	if transaction.Status == entities.TransactionStatusCompleted {
		return errors.New("transaction is already completed")
	}

	transaction.Status = entities.TransactionStatusCompleted
	return uc.transactionRepo.Update(ctx, transaction)
}

// CancelTransaction cancels a transaction
func (uc *TransactionUseCase) CancelTransaction(ctx context.Context, id uuid.UUID) error {
	transaction, err := uc.transactionRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if transaction == nil {
		return errors.New("transaction not found")
	}

	if transaction.Status == entities.TransactionStatusCompleted {
		return errors.New("cannot cancel completed transaction")
	}

	transaction.Status = entities.TransactionStatusCancelled
	return uc.transactionRepo.Update(ctx, transaction)
}

// UpdateTransaction updates an existing transaction
func (uc *TransactionUseCase) UpdateTransaction(ctx context.Context, transaction *entities.Transaction) error {
	if transaction.ID == uuid.Nil {
		return errors.New("transaction ID is required")
	}

	existing, err := uc.transactionRepo.GetByID(ctx, transaction.ID)
	if err != nil {
		return err
	}
	if existing == nil {
		return errors.New("transaction not found")
	}

	return uc.transactionRepo.Update(ctx, transaction)
}

// ListTransactions retrieves a list of transactions
func (uc *TransactionUseCase) ListTransactions(ctx context.Context, limit, offset int) ([]*entities.Transaction, error) {
	return uc.transactionRepo.List(ctx, limit, offset)
}
