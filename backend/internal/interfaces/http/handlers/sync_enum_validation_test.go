package handlers

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/terminator791/t-pos/internal/domain/dto"
	"github.com/terminator791/t-pos/internal/domain/entities"
)

// Mock repository for testing cashier validation
type MockUserRepo struct {
	mock.Mock
}

func (m *MockUserRepo) GetByID(ctx context.Context, id uuid.UUID) (*entities.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.User), args.Error(1)
}

func (m *MockUserRepo) Create(ctx context.Context, user *entities.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepo) CreatePin(ctx context.Context, id uuid.UUID, pin string) error {
	args := m.Called(ctx, id, pin)
	return args.Error(0)
}

func (m *MockUserRepo) UpdatePin(ctx context.Context, id uuid.UUID, pin string) error {
	args := m.Called(ctx, id, pin)
	return args.Error(0)
}

func (m *MockUserRepo) DeletePin(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockUserRepo) Update(ctx context.Context, user *entities.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepo) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockUserRepo) GetByLicenseID(ctx context.Context, licenseID uuid.UUID) ([]*entities.User, error) {
	args := m.Called(ctx, licenseID)
	return args.Get(0).([]*entities.User), args.Error(1)
}

func (m *MockUserRepo) GetByEmail(ctx context.Context, email string) (*entities.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.User), args.Error(1)
}

func (m *MockUserRepo) GetByUsername(ctx context.Context, username string) (*entities.User, error) {
	args := m.Called(ctx, username)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.User), args.Error(1)
}

func (m *MockUserRepo) List(ctx context.Context, limit, offset int) ([]*entities.User, error) {
	args := m.Called(ctx, limit, offset)
	return args.Get(0).([]*entities.User), args.Error(1)
}

type MockRoleRepo struct {
	mock.Mock
}

func (m *MockRoleRepo) GetByID(ctx context.Context, id uuid.UUID) (*entities.Role, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.Role), args.Error(1)
}

func (m *MockRoleRepo) Create(ctx context.Context, role *entities.Role) error {
	args := m.Called(ctx, role)
	return args.Error(0)
}

func (m *MockRoleRepo) Update(ctx context.Context, role *entities.Role) error {
	args := m.Called(ctx, role)
	return args.Error(0)
}

func (m *MockRoleRepo) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockRoleRepo) GetByName(ctx context.Context, name string) (*entities.Role, error) {
	args := m.Called(ctx, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.Role), args.Error(1)
}

func (m *MockRoleRepo) GetAll(ctx context.Context) ([]*entities.Role, error) {
	args := m.Called(ctx)
	return args.Get(0).([]*entities.Role), args.Error(1)
}

func (m *MockRoleRepo) List(ctx context.Context, limit, offset int) ([]*entities.Role, error) {
	args := m.Called(ctx, limit, offset)
	return args.Get(0).([]*entities.Role), args.Error(1)
}

func (m *MockRoleRepo) GetActiveRoles(ctx context.Context) ([]*entities.Role, error) {
	args := m.Called(ctx)
	return args.Get(0).([]*entities.Role), args.Error(1)
}

func TestSyncHandler_ValidateEnumsAndCashiers_ValidRequest(t *testing.T) {
	// Create mock repositories
	mockUserRepo := &MockUserRepo{}
	mockRoleRepo := &MockRoleRepo{}

	// Create sync handler
	handler := &SyncHandler{
		userRepo: mockUserRepo,
		roleRepo: mockRoleRepo,
	}

	// Create test data
	cashierID := uuid.New()
	roleID := uuid.New()

	// Mock valid cashier user
	cashierUser := &entities.User{
		ID:     cashierID,
		Name:   "Test Cashier",
		RoleID: &roleID,
	}

	cashierRole := &entities.Role{
		ID:   roleID,
		Name: "cashier",
	}

	// Set up mock expectations
	mockUserRepo.On("GetByID", mock.Anything, cashierID).Return(cashierUser, nil)
	mockRoleRepo.On("GetByID", mock.Anything, roleID).Return(cashierRole, nil)

	// Create sync request with valid enums and cashier
	req := &dto.SyncRequest{
		Expenses: []entities.Expense{
			{Status: entities.ExpenseStatusPending},
			{Status: entities.ExpenseStatusCompleted},
		},
		Payments: []entities.Payment{
			{Status: entities.PaymentStatusPending},
			{Status: entities.PaymentStatusCompleted},
		},
		Transactions: []entities.Transaction{
			{
				CashierID: cashierID,
				Status:    entities.TransactionStatusPending,
			},
			{
				CashierID: cashierID,
				Status:    entities.TransactionStatusCompleted,
			},
		},
		Shops: []entities.Shop{
			{
				ID:     uuid.New(),
				Domain: "", // Should be auto-initialized
			},
		},
	}

	// Test validation
	err := handler.validateSyncRequestEnumsAndCashiers(req)

	// Assertions
	assert.NoError(t, err)
	// Check that shop domain was auto-initialized
	assert.NotEmpty(t, req.Shops[0].Domain)
	assert.Contains(t, req.Shops[0].Domain, "shop-")

	// Verify all mock expectations were met
	mockUserRepo.AssertExpectations(t)
	mockRoleRepo.AssertExpectations(t)
}

func TestSyncHandler_ValidateEnumsAndCashiers_InvalidExpenseStatus(t *testing.T) {
	handler := &SyncHandler{}

	req := &dto.SyncRequest{
		Expenses: []entities.Expense{
			{Status: "invalid_status"},
		},
	}

	err := handler.validateSyncRequestEnumsAndCashiers(req)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid expense status")
}

func TestSyncHandler_ValidateEnumsAndCashiers_InvalidPaymentStatus(t *testing.T) {
	handler := &SyncHandler{}

	req := &dto.SyncRequest{
		Payments: []entities.Payment{
			{Status: "invalid_status"},
		},
	}

	err := handler.validateSyncRequestEnumsAndCashiers(req)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid payment status")
}

func TestSyncHandler_ValidateEnumsAndCashiers_InvalidTransactionStatus(t *testing.T) {
	handler := &SyncHandler{}

	req := &dto.SyncRequest{
		Transactions: []entities.Transaction{
			{
				CashierID: uuid.New(),
				Status:    "invalid_status",
			},
		},
	}

	err := handler.validateSyncRequestEnumsAndCashiers(req)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid transaction status")
}

func TestSyncHandler_ValidateEnumsAndCashiers_InvalidCashierID(t *testing.T) {
	mockUserRepo := &MockUserRepo{}
	handler := &SyncHandler{
		userRepo: mockUserRepo,
	}

	cashierID := uuid.New()

	// Mock user not found
	mockUserRepo.On("GetByID", mock.Anything, cashierID).Return(nil, assert.AnError)

	req := &dto.SyncRequest{
		Transactions: []entities.Transaction{
			{
				CashierID: cashierID,
				Status:    entities.TransactionStatusPending,
			},
		},
	}

	err := handler.validateSyncRequestEnumsAndCashiers(req)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cashier_id")
	assert.Contains(t, err.Error(), "not found")

	mockUserRepo.AssertExpectations(t)
}

func TestSyncHandler_ValidateEnumsAndCashiers_UserNotCashier(t *testing.T) {
	mockUserRepo := &MockUserRepo{}
	mockRoleRepo := &MockRoleRepo{}

	handler := &SyncHandler{
		userRepo: mockUserRepo,
		roleRepo: mockRoleRepo,
	}

	cashierID := uuid.New()
	roleID := uuid.New()

	// Mock user with non-cashier role
	user := &entities.User{
		ID:     cashierID,
		Name:   "Test User",
		RoleID: &roleID,
	}

	role := &entities.Role{
		ID:   roleID,
		Name: "owner_business", // Not a cashier
	}

	mockUserRepo.On("GetByID", mock.Anything, cashierID).Return(user, nil)
	mockRoleRepo.On("GetByID", mock.Anything, roleID).Return(role, nil)

	req := &dto.SyncRequest{
		Transactions: []entities.Transaction{
			{
				CashierID: cashierID,
				Status:    entities.TransactionStatusPending,
			},
		},
	}

	err := handler.validateSyncRequestEnumsAndCashiers(req)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "expected 'cashier'")

	mockUserRepo.AssertExpectations(t)
	mockRoleRepo.AssertExpectations(t)
}

func TestSyncHandler_ValidateShopDomain_AutoInitialize(t *testing.T) {
	handler := &SyncHandler{}

	// Test shop without domain
	shop := &entities.Shop{
		ID:     uuid.New(),
		Domain: "",
	}

	err := handler.validateAndInitializeShopDomain(shop)

	assert.NoError(t, err)
	assert.Equal(t, "shop-"+shop.ID.String(), shop.Domain)
}

func TestSyncHandler_ValidateShopDomain_InvalidFormat(t *testing.T) {
	handler := &SyncHandler{}

	shopID := uuid.New()
	shop := &entities.Shop{
		ID:     shopID,
		Domain: "invalid-domain",
	}

	err := handler.validateAndInitializeShopDomain(shop)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "does not match expected format")
}
