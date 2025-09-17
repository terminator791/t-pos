package services

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/terminator791/t-pos/internal/domain/dto"
	"github.com/terminator791/t-pos/internal/domain/entities"
)

func TestSyncValidationWithPayload(t *testing.T) {
	// Create SyncService instance
	syncService := &SyncService{}

	// Test with the exact payload from the user's request
	shopID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440114")
	licenseID := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	userID := uuid.MustParse("cccccccc-cccc-cccc-cccc-cccccccccccc")
	cashierID := uuid.MustParse("ffffffff-ffff-ffff-ffff-ffffffffffff")
	categoryID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440027")
	productID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440001")
	cartID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440002")
	transactionID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440003")
	expenseID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440004")
	paymentID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440005")
	historyID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440006")
	receiptID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440007")
	transactionProductID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440008")

	// Create the exact sync request from the user
	address := "Jl. Test Sync No. 123, Jakarta"
	slogan := "toko kami hebat"
	label := "Office Supplies"
	desc := "Test expense description"
	customerName := "John Doe"
	cashierName := "Cashier Name"
	photo := "/public/photos/shops"
	productPhoto := "/public/photos/products"
	barcode := "1234567890123"
	unit := "pcs"
	ppn := 10.0
	profit := 10000.0
	qty := 1
	profitTransaction := 20000.0
	change := 0.0
	initialPaymentStatus := "pending"

	syncRequest := dto.SyncRequest{
		LastSyncTimestamp: nil,
		Shops: []entities.Shop{
			{
				ID:              shopID,
				LicenseID:       licenseID,
				UserID:          userID,
				Name:            "Test Shop Sync",
				Domain:          "shop-550e8400-e29b-41d4-a716-446655440114",
				Photo:           &photo,
				Address:         &address,
				Slogan:          &slogan,
				ProfitCalculate: 1000000,
				CreatedAt:       time.Date(2026, 1, 15, 12, 9, 0, 0, time.UTC),
				UpdatedAt:       time.Date(2026, 1, 15, 12, 10, 0, 0, time.UTC),
			},
		},
		Categories: []entities.Category{
			{
				ID:        categoryID,
				ShopID:    shopID,
				Name:      "Test Category Syncs",
				CreatedAt: time.Date(2026, 1, 15, 11, 0, 0, 0, time.UTC),
				UpdatedAt: time.Date(2026, 1, 15, 11, 30, 0, 0, time.UTC),
			},
		},
		Products: []entities.Product{
			{
				ID:          productID,
				ShopID:      shopID,
				CatID:       &categoryID,
				Photo:       &productPhoto,
				Name:        "Test Product Sync",
				Barcode:     &barcode,
				Unit:        &unit,
				PPN:         &ppn,
				Sale:        50000.0,
				Buy:         40000.0,
				Profit:      &profit,
				Stock:       100,
				IsSchedule:  false,
				Qty:         &qty,
				IsHaveStock: true,
				CreatedAt:   time.Date(2026, 1, 15, 12, 0, 0, 0, time.UTC),
				UpdatedAt:   time.Date(2026, 1, 15, 12, 5, 0, 0, time.UTC),
			},
		},
		Carts: []entities.Cart{
			{
				ID:        cartID,
				ShopID:    shopID,
				ProductID: productID,
				UserID:    userID,
				Quantity:  2,
				CreatedAt: time.Date(2026, 1, 15, 12, 15, 0, 0, time.UTC),
				UpdatedAt: time.Date(2026, 1, 15, 12, 20, 0, 0, time.UTC),
			},
		},
		Transactions: []entities.Transaction{
			{
				ID:                   transactionID,
				ShopID:               shopID,
				CashierID:            cashierID,
				CustomerName:         &customerName,
				Discount:             0.0,
				DiscountPercentage:   0.0,
				AdditionalCost:       0.0,
				Status:               "pending",
				TotalPrice:           100000.0,
				ProfitTransaction:    &profitTransaction,
				CashierName:          &cashierName,
				Change:               &change,
				Amount:               100000,
				InitialPaymentStatus: &initialPaymentStatus,
				CreatedAt:            time.Date(2026, 1, 15, 12, 25, 0, 0, time.UTC),
				UpdatedAt:            time.Date(2026, 1, 15, 12, 30, 0, 0, time.UTC),
			},
		},
		Expenses: []entities.Expense{
			{
				ID:        expenseID,
				ShopID:    shopID,
				Nominal:   50000.0,
				Status:    "pending",
				Date:      time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC),
				Label:     &label,
				Desc:      &desc,
				CreatedAt: time.Date(2026, 1, 15, 12, 35, 0, 0, time.UTC),
				UpdatedAt: time.Date(2026, 1, 15, 12, 40, 0, 0, time.UTC),
			},
		},
		Payments: []entities.Payment{
			{
				ID:            paymentID,
				ShopID:        shopID,
				UserID:        &userID,
				TransactionID: transactionID,
				Status:        "pending",
				Total:         100000.0,
				CreatedAt:     time.Date(2026, 1, 15, 12, 45, 0, 0, time.UTC),
				UpdatedAt:     time.Date(2026, 1, 15, 12, 50, 0, 0, time.UTC),
			},
		},
		Histories: []entities.History{
			{
				ID:            historyID,
				ShopID:        shopID,
				TransactionID: transactionID,
				CreatedAt:     time.Date(2026, 1, 15, 12, 55, 0, 0, time.UTC),
				UpdatedAt:     time.Date(2026, 1, 15, 13, 0, 0, 0, time.UTC),
			},
		},
		Receipts: []entities.Receipt{
			{
				ID:         receiptID,
				ShopID:     shopID,
				PaymentsID: paymentID,
				CreatedAt:  time.Date(2026, 1, 15, 13, 5, 0, 0, time.UTC),
				UpdatedAt:  time.Date(2026, 1, 15, 13, 10, 0, 0, time.UTC),
			},
		},
		TransactionProducts: []entities.TransactionProduct{
			{
				ID:            transactionProductID,
				TransactionID: transactionID,
				ProductID:     productID,
				Quantity:      2,
				UnitPrice:     50000.0,
				TotalPrice:    100000.0,
				CreatedAt:     time.Date(2026, 1, 15, 13, 15, 0, 0, time.UTC),
				UpdatedAt:     time.Date(2026, 1, 15, 13, 20, 0, 0, time.UTC),
			},
		},
	}

	// Test 1: All payload checks should work
	t.Run("ValidateAllPayloadReferences", func(t *testing.T) {
		// Test shop in payload
		if !syncService.checkShopInPayload(shopID, syncRequest) {
			t.Errorf("Expected shop %s to be found in payload", shopID)
		}

		// Test transaction in payload
		if !syncService.checkTransactionInPayload(transactionID, syncRequest) {
			t.Errorf("Expected transaction %s to be found in payload", transactionID)
		}

		// Test payment in payload
		if !syncService.checkPaymentInPayload(paymentID, syncRequest) {
			t.Errorf("Expected payment %s to be found in payload", paymentID)
		}

		// Test that non-existent IDs are not found
		nonExistentID := uuid.New()
		if syncService.checkShopInPayload(nonExistentID, syncRequest) {
			t.Errorf("Expected non-existent shop %s to NOT be found in payload", nonExistentID)
		}
	})

	// Test 2: Test the actual validation logic that was failing
	t.Run("ValidateExpenseShopReference", func(t *testing.T) {
		expense := syncRequest.Expenses[0]

		// This should pass because the shop exists in the payload
		// (simulating what validateExpenseReferencesWithPayload would do)
		if !syncService.checkShopInPayload(expense.ShopID, syncRequest) {
			t.Errorf("Expected expense shop reference %s to be found in payload", expense.ShopID)
		}
	})

	t.Run("ValidatePaymentReferences", func(t *testing.T) {
		payment := syncRequest.Payments[0]

		// Shop reference should be found in payload
		if !syncService.checkShopInPayload(payment.ShopID, syncRequest) {
			t.Errorf("Expected payment shop reference %s to be found in payload", payment.ShopID)
		}

		// Transaction reference should be found in payload
		if !syncService.checkTransactionInPayload(payment.TransactionID, syncRequest) {
			t.Errorf("Expected payment transaction reference %s to be found in payload", payment.TransactionID)
		}
	})

	t.Run("ValidateHistoryReferences", func(t *testing.T) {
		history := syncRequest.Histories[0]

		// Shop reference should be found in payload
		if !syncService.checkShopInPayload(history.ShopID, syncRequest) {
			t.Errorf("Expected history shop reference %s to be found in payload", history.ShopID)
		}

		// Transaction reference should be found in payload
		if !syncService.checkTransactionInPayload(history.TransactionID, syncRequest) {
			t.Errorf("Expected history transaction reference %s to be found in payload", history.TransactionID)
		}
	})

	t.Run("ValidateReceiptReferences", func(t *testing.T) {
		receipt := syncRequest.Receipts[0]

		// Shop reference should be found in payload
		if !syncService.checkShopInPayload(receipt.ShopID, syncRequest) {
			t.Errorf("Expected receipt shop reference %s to be found in payload", receipt.ShopID)
		}

		// Payment reference should be found in payload
		if !syncService.checkPaymentInPayload(receipt.PaymentsID, syncRequest) {
			t.Errorf("Expected receipt payment reference %s to be found in payload", receipt.PaymentsID)
		}
	})

	// Test 3: Verify the dependency chain works
	t.Run("ValidateDependencyChain", func(t *testing.T) {
		// Shop -> Transaction -> Payment -> Receipt (dependency chain)

		// 1. Shop should exist
		if !syncService.checkShopInPayload(shopID, syncRequest) {
			t.Errorf("Shop %s should exist in payload", shopID)
		}

		// 2. Transaction references the shop
		transaction := syncRequest.Transactions[0]
		if transaction.ShopID != shopID {
			t.Errorf("Transaction should reference shop %s", shopID)
		}

		// 3. Payment references both shop and transaction
		payment := syncRequest.Payments[0]
		if payment.ShopID != shopID {
			t.Errorf("Payment should reference shop %s", shopID)
		}
		if payment.TransactionID != transactionID {
			t.Errorf("Payment should reference transaction %s", transactionID)
		}

		// 4. Receipt references both shop and payment
		receipt := syncRequest.Receipts[0]
		if receipt.ShopID != shopID {
			t.Errorf("Receipt should reference shop %s", shopID)
		}
		if receipt.PaymentsID != paymentID {
			t.Errorf("Receipt should reference payment %s", paymentID)
		}

		t.Log("✅ All dependency references are correctly set up in the payload")
	})
}
