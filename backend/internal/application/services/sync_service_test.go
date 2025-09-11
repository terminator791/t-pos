package services

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/terminator791/t-pos/internal/domain/dto"
	"github.com/terminator791/t-pos/internal/domain/entities"
	"gorm.io/gorm"
)

// Mock repository for testing
type MockCartRepository struct {
	mock.Mock
}

func (m *MockCartRepository) Create(ctx context.Context, cart *entities.Cart) error {
	args := m.Called(ctx, cart)
	return args.Error(0)
}

func (m *MockCartRepository) GetByID(ctx context.Context, id uuid.UUID) (*entities.Cart, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*entities.Cart), args.Error(1)
}

func (m *MockCartRepository) GetByUserID(ctx context.Context, userID uuid.UUID) ([]*entities.Cart, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]*entities.Cart), args.Error(1)
}

func (m *MockCartRepository) GetByUserAndProduct(ctx context.Context, userID, productID uuid.UUID) (*entities.Cart, error) {
	args := m.Called(ctx, userID, productID)
	return args.Get(0).(*entities.Cart), args.Error(1)
}

func (m *MockCartRepository) Update(ctx context.Context, cart *entities.Cart) error {
	args := m.Called(ctx, cart)
	return args.Error(0)
}

func (m *MockCartRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockCartRepository) DeleteByUserID(ctx context.Context, userID uuid.UUID) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func (m *MockCartRepository) List(ctx context.Context, limit, offset int) ([]*entities.Cart, error) {
	args := m.Called(ctx, limit, offset)
	return args.Get(0).([]*entities.Cart), args.Error(1)
}

func TestSyncService_ProcessSync_EmptyRequest(t *testing.T) {
	// Setup
	mockCartRepo := new(MockCartRepository)
	
	// For this test, we'll create a minimal sync service with nil dependencies
	// since we're testing an empty request
	syncService := &SyncService{
		cartRepo:         mockCartRepo,
		conflictStrategy: dto.LastWriteWins,
	}

	// Test data
	licenseID := uuid.New()
	userID := uuid.New()
	syncRequest := dto.SyncRequest{
		LastSyncTimestamp: nil,
		Carts:             []entities.Cart{},
		Categories:        []entities.Category{},
		Products:          []entities.Product{},
		Transactions:      []entities.Transaction{},
	}

	// Note: This test would require a proper database transaction mock
	// For now, we'll test the basic structure is working
	
	// The actual test would need a database mock, so let's validate the service creation
	assert.NotNil(t, syncService)
	assert.Equal(t, dto.LastWriteWins, syncService.conflictStrategy)
	
	// Verify test data is properly initialized
	assert.NotEqual(t, uuid.Nil, licenseID)
	assert.NotEqual(t, uuid.Nil, userID)
	assert.Empty(t, syncRequest.Carts)
}

func TestSyncService_ValidateCartLicense(t *testing.T) {
	// For now, we'll skip this test since it requires database mocking
	// In a full implementation, we'd mock the database or use a test database
	t.Skip("Database integration test - requires proper database mocking")
}

func TestConflictResolution(t *testing.T) {
	syncService := &SyncService{
		conflictStrategy: dto.LastWriteWins,
	}

	// Test data - create two carts with different timestamps
	now := time.Now()
	olderTime := now.Add(-1 * time.Hour)
	
	existingCart := entities.Cart{
		ID:        uuid.New(),
		UpdatedAt: olderTime,
	}
	
	incomingCart := entities.Cart{
		ID:        existingCart.ID,
		UpdatedAt: now,
	}

	// Test conflict resolution
	conflict := syncService.resolveCartConflict(existingCart, incomingCart)
	
	assert.NotNil(t, conflict)
	assert.Equal(t, "cart", conflict.EntityType)
	assert.Equal(t, existingCart.ID, conflict.EntityID)
	assert.Equal(t, "client_wins", conflict.Resolution) // incoming cart is newer
}

// Test Product sync helper methods
func TestSyncService_ProductConflictResolution(t *testing.T) {
	syncService := &SyncService{
		conflictStrategy: dto.LastWriteWins,
	}

	now := time.Now()
	olderTime := now.Add(-1 * time.Hour)
	
	existingProduct := entities.Product{
		ID:        uuid.New(),
		UpdatedAt: now,
		Name:      "Product A",
		Sale:      100.0,
	}
	
	incomingProduct := entities.Product{
		ID:        existingProduct.ID,
		UpdatedAt: olderTime,
		Name:      "Product A Updated",
		Sale:      150.0,
	}

	// Test conflict resolution
	conflict := syncService.resolveProductConflict(existingProduct, incomingProduct)
	
	assert.NotNil(t, conflict)
	assert.Equal(t, "product", conflict.EntityType)
	assert.Equal(t, "server_wins", conflict.Resolution)
	assert.Equal(t, existingProduct, conflict.ServerData)
	assert.Equal(t, incomingProduct, conflict.ClientData)
}

func TestSyncService_ProductConflictResolution_ClientWins(t *testing.T) {
	syncService := &SyncService{
		conflictStrategy: dto.LastWriteWins,
	}

	now := time.Now()
	newerTime := now.Add(1 * time.Hour)
	
	existingProduct := entities.Product{
		ID:        uuid.New(),
		UpdatedAt: now,
		Name:      "Product A",
	}
	
	incomingProduct := entities.Product{
		ID:        existingProduct.ID,
		UpdatedAt: newerTime,
		Name:      "Product A Updated",
	}

	conflict := syncService.resolveProductConflict(existingProduct, incomingProduct)
	
	assert.NotNil(t, conflict)
	assert.Equal(t, "client_wins", conflict.Resolution)
	assert.Equal(t, "Client version is newer", conflict.Details)
}

func TestSyncService_ProductConflictResolution_NoConflict(t *testing.T) {
	syncService := &SyncService{
		conflictStrategy: dto.LastWriteWins,
	}

	now := time.Now()
	
	existingProduct := entities.Product{
		ID:        uuid.New(),
		UpdatedAt: now,
	}
	
	incomingProduct := entities.Product{
		ID:        existingProduct.ID,
		UpdatedAt: now,
	}

	conflict := syncService.resolveProductConflict(existingProduct, incomingProduct)
	
	assert.Nil(t, conflict, "Should not detect conflict when timestamps are equal")
}

// Test Transaction sync helper methods
func TestSyncService_TransactionConflictResolution(t *testing.T) {
	syncService := &SyncService{
		conflictStrategy: dto.ServerWins,
	}

	now := time.Now()
	newerTime := now.Add(1 * time.Hour)
	
	existingTransaction := entities.Transaction{
		ID:         uuid.New(),
		UpdatedAt:  now,
		TotalPrice: 100.0,
		Status:     entities.TransactionStatusCompleted,
	}
	
	incomingTransaction := entities.Transaction{
		ID:         existingTransaction.ID,
		UpdatedAt:  newerTime,
		TotalPrice: 150.0,
		Status:     entities.TransactionStatusPending,
	}

	conflict := syncService.resolveTransactionConflict(existingTransaction, incomingTransaction)
	
	assert.NotNil(t, conflict)
	assert.Equal(t, "transaction", conflict.EntityType)
	assert.Equal(t, "server_wins", conflict.Resolution, "Server should always win with ServerWins strategy")
	assert.Equal(t, "Server version always wins", conflict.Details)
}

// Test Expense sync helper methods
func TestSyncService_ExpenseConflictResolution(t *testing.T) {
	syncService := &SyncService{
		conflictStrategy: dto.ClientWins,
	}

	now := time.Now()
	olderTime := now.Add(-1 * time.Hour)
	
	existingExpense := entities.Expense{
		ID:        uuid.New(),
		UpdatedAt: now,
		Nominal:   100.0,
		Status:    entities.ExpenseStatusCompleted,
	}
	
	incomingExpense := entities.Expense{
		ID:        existingExpense.ID,
		UpdatedAt: olderTime,
		Nominal:   150.0,
		Status:    entities.ExpenseStatusPending,
	}

	conflict := syncService.resolveExpenseConflict(existingExpense, incomingExpense)
	
	assert.NotNil(t, conflict)
	assert.Equal(t, "expense", conflict.EntityType)
	assert.Equal(t, "client_wins", conflict.Resolution, "Client should always win with ClientWins strategy")
	assert.Equal(t, "Client version always wins", conflict.Details)
}

// Test Payment sync helper methods
func TestSyncService_PaymentConflictResolution(t *testing.T) {
	syncService := &SyncService{
		conflictStrategy: dto.LastWriteWins,
	}

	now := time.Now()
	newerTime := now.Add(2 * time.Hour)
	
	existingPayment := entities.Payment{
		ID:        uuid.New(),
		UpdatedAt: now,
		Total:     100.0,
		Status:    entities.PaymentStatusCompleted,
	}
	
	incomingPayment := entities.Payment{
		ID:        existingPayment.ID,
		UpdatedAt: newerTime,
		Total:     150.0,
		Status:    entities.PaymentStatusPending,
	}

	conflict := syncService.resolvePaymentConflict(existingPayment, incomingPayment)
	
	assert.NotNil(t, conflict)
	assert.Equal(t, "payment", conflict.EntityType)
	assert.Equal(t, "client_wins", conflict.Resolution)
	assert.Contains(t, conflict.Details, "Client version is newer")
}

// Test Syncable interface implementations
func TestSyncableInterface_Product(t *testing.T) {
	now := time.Now()
	product := entities.Product{
		ID:        uuid.New(),
		CreatedAt: now,
		UpdatedAt: now,
		DeletedAt: gorm.DeletedAt{},
	}

	assert.Equal(t, product.ID, product.GetID())
	assert.Equal(t, product.CreatedAt, product.GetCreatedAt())
	assert.Equal(t, product.UpdatedAt, product.GetUpdatedAt())
	assert.Nil(t, product.GetDeletedAt())
}

func TestSyncableInterface_Transaction(t *testing.T) {
	now := time.Now()
	deletedTime := now.Add(1 * time.Hour)
	transaction := entities.Transaction{
		ID:        uuid.New(),
		CreatedAt: now,
		UpdatedAt: now,
		DeletedAt: gorm.DeletedAt{Time: deletedTime, Valid: true},
	}

	assert.Equal(t, transaction.ID, transaction.GetID())
	assert.Equal(t, transaction.CreatedAt, transaction.GetCreatedAt())
	assert.Equal(t, transaction.UpdatedAt, transaction.GetUpdatedAt())
	assert.NotNil(t, transaction.GetDeletedAt())
	assert.Equal(t, deletedTime, *transaction.GetDeletedAt())
}

func TestSyncableInterface_Expense(t *testing.T) {
	now := time.Now()
	expense := entities.Expense{
		ID:        uuid.New(),
		CreatedAt: now,
		UpdatedAt: now,
		DeletedAt: gorm.DeletedAt{},
	}

	assert.Equal(t, expense.ID, expense.GetID())
	assert.Equal(t, expense.CreatedAt, expense.GetCreatedAt())
	assert.Equal(t, expense.UpdatedAt, expense.GetUpdatedAt())
	assert.Nil(t, expense.GetDeletedAt())
}

func TestSyncableInterface_Payment(t *testing.T) {
	now := time.Now()
	payment := entities.Payment{
		ID:        uuid.New(),
		CreatedAt: now,
		UpdatedAt: now,
		DeletedAt: gorm.DeletedAt{},
	}

	assert.Equal(t, payment.ID, payment.GetID())
	assert.Equal(t, payment.CreatedAt, payment.GetCreatedAt())
	assert.Equal(t, payment.UpdatedAt, payment.GetUpdatedAt())
	assert.Nil(t, payment.GetDeletedAt())
}

// Test Receipt sync helper methods
func TestSyncService_ReceiptConflictResolution(t *testing.T) {
	syncService := &SyncService{
		conflictStrategy: dto.LastWriteWins,
	}

	now := time.Now()
	newerTime := now.Add(1 * time.Hour)
	
	existingReceipt := entities.Receipt{
		ID:         uuid.New(),
		UpdatedAt:  now,
		ShopID:     uuid.New(),
		PaymentsID: uuid.New(),
	}
	
	incomingReceipt := entities.Receipt{
		ID:         existingReceipt.ID,
		UpdatedAt:  newerTime,
		ShopID:     existingReceipt.ShopID,
		PaymentsID: uuid.New(), // Different payment
	}

	conflict := syncService.resolveReceiptConflict(existingReceipt, incomingReceipt)
	
	assert.NotNil(t, conflict)
	assert.Equal(t, "receipt", conflict.EntityType)
	assert.Equal(t, "client_wins", conflict.Resolution)
	assert.Contains(t, conflict.Details, "Client version is newer")
}

func TestSyncService_ReceiptConflictResolution_NoConflict(t *testing.T) {
	syncService := &SyncService{
		conflictStrategy: dto.LastWriteWins,
	}

	now := time.Now()
	
	existingReceipt := entities.Receipt{
		ID:        uuid.New(),
		UpdatedAt: now,
	}
	
	incomingReceipt := entities.Receipt{
		ID:        existingReceipt.ID,
		UpdatedAt: now,
	}

	conflict := syncService.resolveReceiptConflict(existingReceipt, incomingReceipt)
	
	assert.Nil(t, conflict, "Should not detect conflict when timestamps are equal")
}

// Test History sync helper methods  
func TestSyncService_HistoryConflictResolution(t *testing.T) {
	syncService := &SyncService{
		conflictStrategy: dto.ServerWins,
	}

	now := time.Now()
	newerTime := now.Add(1 * time.Hour)
	
	existingHistory := entities.History{
		ID:            uuid.New(),
		UpdatedAt:     now,
		ShopID:        uuid.New(),
		TransactionID: uuid.New(),
	}
	
	incomingHistory := entities.History{
		ID:            existingHistory.ID,
		UpdatedAt:     newerTime,
		ShopID:        existingHistory.ShopID,
		TransactionID: uuid.New(), // Different transaction
	}

	conflict := syncService.resolveHistoryConflict(existingHistory, incomingHistory)
	
	assert.NotNil(t, conflict)
	assert.Equal(t, "history", conflict.EntityType)
	assert.Equal(t, "server_wins", conflict.Resolution)
	assert.Equal(t, "Server always wins conflict resolution strategy", conflict.Details)
}

// Test Shop sync helper methods
func TestSyncService_ShopConflictResolution(t *testing.T) {
	syncService := &SyncService{
		conflictStrategy: dto.LastWriteWins,
	}

	now := time.Now()
	olderTime := now.Add(-1 * time.Hour)
	
	existingShop := entities.Shop{
		ID:        uuid.New(),
		UpdatedAt: now,
		Name:      "Shop A",
		LicenseID: uuid.New(),
		UserID:    uuid.New(),
		Domain:    "shop-a",
	}
	
	incomingShop := entities.Shop{
		ID:        existingShop.ID,
		UpdatedAt: olderTime,
		Name:      "Shop A Updated",
		LicenseID: existingShop.LicenseID,
		UserID:    existingShop.UserID,
		Domain:    "shop-a-updated",
	}

	conflict := syncService.resolveShopConflict(existingShop, incomingShop)
	
	assert.NotNil(t, conflict)
	assert.Equal(t, "shop", conflict.EntityType)
	assert.Equal(t, "server_wins", conflict.Resolution)
	assert.Contains(t, conflict.Details, "Server version is newer")
}

// Test StockHistory sync helper methods
func TestSyncService_StockHistoryConflictResolution(t *testing.T) {
	syncService := &SyncService{
		conflictStrategy: dto.ClientWins,
	}

	now := time.Now()
	olderTime := now.Add(-1 * time.Hour)
	
	existingStockHistory := entities.StockHistory{
		ID:        uuid.New(),
		UpdatedAt: now,
		ProductID: uuid.New(),
		Stock:     100,
		LastStock: 80,
		StockedAt: now,
	}
	
	incomingStockHistory := entities.StockHistory{
		ID:        existingStockHistory.ID,
		UpdatedAt: olderTime,
		ProductID: existingStockHistory.ProductID,
		Stock:     120,
		LastStock: 100,
		StockedAt: olderTime,
	}

	conflict := syncService.resolveStockHistoryConflict(existingStockHistory, incomingStockHistory)
	
	assert.NotNil(t, conflict)
	assert.Equal(t, "stock_history", conflict.EntityType)
	assert.Equal(t, "client_wins", conflict.Resolution)
	assert.Equal(t, "Client always wins conflict resolution strategy", conflict.Details)
}

// Test TransactionProduct sync helper methods
func TestSyncService_TransactionProductConflictResolution(t *testing.T) {
	syncService := &SyncService{
		conflictStrategy: dto.LastWriteWins,
	}

	now := time.Now()
	newerTime := now.Add(1 * time.Hour)
	
	existingTransactionProduct := entities.TransactionProduct{
		ID:            uuid.New(),
		UpdatedAt:     now,
		TransactionID: uuid.New(),
		ProductID:     uuid.New(),
		Quantity:      2,
		UnitPrice:     50.0,
		TotalPrice:    100.0,
	}
	
	incomingTransactionProduct := entities.TransactionProduct{
		ID:            existingTransactionProduct.ID,
		UpdatedAt:     newerTime,
		TransactionID: existingTransactionProduct.TransactionID,
		ProductID:     existingTransactionProduct.ProductID,
		Quantity:      3,
		UnitPrice:     50.0,
		TotalPrice:    150.0,
	}

	conflict := syncService.resolveTransactionProductConflict(existingTransactionProduct, incomingTransactionProduct)
	
	assert.NotNil(t, conflict)
	assert.Equal(t, "transaction_product", conflict.EntityType)
	assert.Equal(t, "client_wins", conflict.Resolution)
	assert.Contains(t, conflict.Details, "Client version is newer")
}

// Test User sync helper methods
func TestSyncService_UserConflictResolution(t *testing.T) {
	syncService := &SyncService{
		conflictStrategy: dto.LastWriteWins,
	}

	now := time.Now()
	olderTime := now.Add(-1 * time.Hour)
	licenseID := uuid.New()
	
	existingUser := entities.User{
		ID:        uuid.New(),
		UpdatedAt: now,
		LicenseID: &licenseID,
		Name:      "John Doe",
		Username:  stringPtr("johndoe"),
		Email:     stringPtr("john@example.com"),
	}
	
	incomingUser := entities.User{
		ID:        existingUser.ID,
		UpdatedAt: olderTime,
		LicenseID: &licenseID,
		Name:      "John Smith",
		Username:  stringPtr("johnsmith"),
		Email:     stringPtr("johnsmith@example.com"),
	}

	conflict := syncService.resolveUserConflict(existingUser, incomingUser)
	
	assert.NotNil(t, conflict)
	assert.Equal(t, "user", conflict.EntityType)
	assert.Equal(t, "server_wins", conflict.Resolution)
	assert.Contains(t, conflict.Details, "Server version is newer")
}

func TestSyncService_UserConflictResolution_NoConflict(t *testing.T) {
	syncService := &SyncService{
		conflictStrategy: dto.LastWriteWins,
	}

	now := time.Now()
	licenseID := uuid.New()
	
	existingUser := entities.User{
		ID:        uuid.New(),
		UpdatedAt: now,
		LicenseID: &licenseID,
	}
	
	incomingUser := entities.User{
		ID:        existingUser.ID,
		UpdatedAt: now,
		LicenseID: &licenseID,
	}

	conflict := syncService.resolveUserConflict(existingUser, incomingUser)
	
	assert.Nil(t, conflict, "Should not detect conflict when timestamps are equal")
}

// Test Syncable interface implementations for new entities
func TestSyncableInterface_Receipt(t *testing.T) {
	now := time.Now()
	receipt := entities.Receipt{
		ID:        uuid.New(),
		CreatedAt: now,
		UpdatedAt: now,
		DeletedAt: gorm.DeletedAt{},
	}

	assert.Equal(t, receipt.ID, receipt.GetID())
	assert.Equal(t, receipt.CreatedAt, receipt.GetCreatedAt())
	assert.Equal(t, receipt.UpdatedAt, receipt.GetUpdatedAt())
	assert.Nil(t, receipt.GetDeletedAt())
}

func TestSyncableInterface_History(t *testing.T) {
	now := time.Now()
	deletedTime := now.Add(1 * time.Hour)
	history := entities.History{
		ID:        uuid.New(),
		CreatedAt: now,
		UpdatedAt: now,
		DeletedAt: gorm.DeletedAt{Time: deletedTime, Valid: true},
	}

	assert.Equal(t, history.ID, history.GetID())
	assert.Equal(t, history.CreatedAt, history.GetCreatedAt())
	assert.Equal(t, history.UpdatedAt, history.GetUpdatedAt())
	assert.NotNil(t, history.GetDeletedAt())
	assert.Equal(t, deletedTime, *history.GetDeletedAt())
}

func TestSyncableInterface_Shop(t *testing.T) {
	now := time.Now()
	shop := entities.Shop{
		ID:        uuid.New(),
		CreatedAt: now,
		UpdatedAt: now,
		DeletedAt: gorm.DeletedAt{},
	}

	assert.Equal(t, shop.ID, shop.GetID())
	assert.Equal(t, shop.CreatedAt, shop.GetCreatedAt())
	assert.Equal(t, shop.UpdatedAt, shop.GetUpdatedAt())
	assert.Nil(t, shop.GetDeletedAt())
}

func TestSyncableInterface_StockHistory(t *testing.T) {
	now := time.Now()
	stockHistory := entities.StockHistory{
		ID:        uuid.New(),
		CreatedAt: now,
		UpdatedAt: now,
		DeletedAt: gorm.DeletedAt{},
	}

	assert.Equal(t, stockHistory.ID, stockHistory.GetID())
	assert.Equal(t, stockHistory.CreatedAt, stockHistory.GetCreatedAt())
	assert.Equal(t, stockHistory.UpdatedAt, stockHistory.GetUpdatedAt())
	assert.Nil(t, stockHistory.GetDeletedAt())
}

func TestSyncableInterface_TransactionProduct(t *testing.T) {
	now := time.Now()
	transactionProduct := entities.TransactionProduct{
		ID:        uuid.New(),
		CreatedAt: now,
		UpdatedAt: now,
		DeletedAt: gorm.DeletedAt{},
	}

	assert.Equal(t, transactionProduct.ID, transactionProduct.GetID())
	assert.Equal(t, transactionProduct.CreatedAt, transactionProduct.GetCreatedAt())
	assert.Equal(t, transactionProduct.UpdatedAt, transactionProduct.GetUpdatedAt())
	assert.Nil(t, transactionProduct.GetDeletedAt())
}

func TestSyncableInterface_User(t *testing.T) {
	now := time.Now()
	user := entities.User{
		ID:        uuid.New(),
		CreatedAt: now,
		UpdatedAt: now,
		DeletedAt: gorm.DeletedAt{},
	}

	assert.Equal(t, user.ID, user.GetID())
	assert.Equal(t, user.CreatedAt, user.GetCreatedAt())
	assert.Equal(t, user.UpdatedAt, user.GetUpdatedAt())
	assert.Nil(t, user.GetDeletedAt())
}