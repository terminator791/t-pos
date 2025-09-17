package services

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/terminator791/t-pos/internal/domain/dto"
	"github.com/terminator791/t-pos/internal/domain/entities"
)

func TestPayloadValidation(t *testing.T) {
	// Create SyncService instance (without database for these tests)
	syncService := &SyncService{}

	// Test IDs
	shopID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440114")
	expenseID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440004")
	transactionID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440003")
	paymentID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440005")
	historyID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440006")
	receiptID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440007")

	// Create sync request with new shop and entities that reference it
	address := "Jl. Test Sync No. 123, Jakarta"
	label := "Office Supplies"

	syncRequest := dto.SyncRequest{
		Shops: []entities.Shop{
			{
				ID:        shopID,
				Name:      "Test Shop Sync",
				Domain:    "shop-550e8400-e29b-41d4-a716-446655440114",
				Address:   &address,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
		},
		Transactions: []entities.Transaction{
			{
				ID:         transactionID,
				ShopID:     shopID,
				TotalPrice: 100000.0,
				Status:     "pending",
				CreatedAt:  time.Now(),
				UpdatedAt:  time.Now(),
			},
		},
		Expenses: []entities.Expense{
			{
				ID:        expenseID,
				ShopID:    shopID,
				Nominal:   50000.0,
				Status:    "pending",
				Label:     &label,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
		},
		Payments: []entities.Payment{
			{
				ID:            paymentID,
				ShopID:        shopID,
				TransactionID: transactionID,
				Status:        "pending",
				Total:         100000.0,
				CreatedAt:     time.Now(),
				UpdatedAt:     time.Now(),
			},
		},
		Histories: []entities.History{
			{
				ID:            historyID,
				ShopID:        shopID,
				TransactionID: transactionID,
				CreatedAt:     time.Now(),
				UpdatedAt:     time.Now(),
			},
		},
		Receipts: []entities.Receipt{
			{
				ID:         receiptID,
				ShopID:     shopID,
				PaymentsID: paymentID,
				CreatedAt:  time.Now(),
				UpdatedAt:  time.Now(),
			},
		},
	}

	// Test 1: Check if shop exists in payload (should return true)
	t.Run("CheckShopInPayload", func(t *testing.T) {
		exists := syncService.checkShopInPayload(shopID, syncRequest)
		if !exists {
			t.Errorf("Expected shop to exist in payload, but it was not found")
		}
	})

	// Test 2: Check if transaction exists in payload (should return true)
	t.Run("CheckTransactionInPayload", func(t *testing.T) {
		exists := syncService.checkTransactionInPayload(transactionID, syncRequest)
		if !exists {
			t.Errorf("Expected transaction to exist in payload, but it was not found")
		}
	})

	// Test 3: Check if payment exists in payload (should return true)
	t.Run("CheckPaymentInPayload", func(t *testing.T) {
		exists := syncService.checkPaymentInPayload(paymentID, syncRequest)
		if !exists {
			t.Errorf("Expected payment to exist in payload, but it was not found")
		}
	})

	// Test 4: Check if non-existent shop exists in payload (should return false)
	t.Run("CheckNonExistentShopInPayload", func(t *testing.T) {
		nonExistentShopID := uuid.New()
		exists := syncService.checkShopInPayload(nonExistentShopID, syncRequest)
		if exists {
			t.Errorf("Expected non-existent shop to not exist in payload, but it was found")
		}
	})

	// Test 5: Check if non-existent transaction exists in payload (should return false)
	t.Run("CheckNonExistentTransactionInPayload", func(t *testing.T) {
		nonExistentTransactionID := uuid.New()
		exists := syncService.checkTransactionInPayload(nonExistentTransactionID, syncRequest)
		if exists {
			t.Errorf("Expected non-existent transaction to not exist in payload, but it was found")
		}
	})

	// Test 6: Check if non-existent payment exists in payload (should return false)
	t.Run("CheckNonExistentPaymentInPayload", func(t *testing.T) {
		nonExistentPaymentID := uuid.New()
		exists := syncService.checkPaymentInPayload(nonExistentPaymentID, syncRequest)
		if exists {
			t.Errorf("Expected non-existent payment to not exist in payload, but it was found")
		}
	})
}
