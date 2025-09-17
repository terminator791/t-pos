package dto

import (
	"github.com/terminator791/t-pos/internal/domain/entities"
)

// MapCartToSyncDTO converts a Cart entity to SyncCartDTO
func MapCartToSyncDTO(cart entities.Cart) SyncCartDTO {
	return SyncCartDTO{
		ID:        cart.ID,
		ShopID:    cart.ShopID,
		ProductID: cart.ProductID,
		UserID:    cart.UserID,
		Quantity:  cart.Quantity,
		CreatedAt: cart.CreatedAt,
		UpdatedAt: cart.UpdatedAt,
	}
}

// MapCartsToSyncDTOs converts a slice of Cart entities to SyncCartDTOs
func MapCartsToSyncDTOs(carts []entities.Cart) []SyncCartDTO {
	result := make([]SyncCartDTO, len(carts))
	for i, cart := range carts {
		result[i] = MapCartToSyncDTO(cart)
	}
	return result
}

// MapCategoryToSyncDTO converts a Category entity to SyncCategoryDTO
func MapCategoryToSyncDTO(category entities.Category) SyncCategoryDTO {
	return SyncCategoryDTO{
		ID:        category.ID,
		ShopID:    category.ShopID,
		Name:      category.Name,
		CreatedAt: category.CreatedAt,
		UpdatedAt: category.UpdatedAt,
	}
}

// MapCategoriesToSyncDTOs converts a slice of Category entities to SyncCategoryDTOs
func MapCategoriesToSyncDTOs(categories []entities.Category) []SyncCategoryDTO {
	result := make([]SyncCategoryDTO, len(categories))
	for i, category := range categories {
		result[i] = MapCategoryToSyncDTO(category)
	}
	return result
}

// MapProductToSyncDTO converts a Product entity to SyncProductDTO
func MapProductToSyncDTO(product entities.Product) SyncProductDTO {
	return SyncProductDTO{
		ID:          product.ID,
		ShopID:      product.ShopID,
		CatID:       product.CatID,
		Photo:       product.Photo,
		Name:        product.Name,
		Barcode:     product.Barcode,
		Unit:        product.Unit,
		PPN:         product.PPN,
		Sale:        product.Sale,
		Buy:         product.Buy,
		Profit:      product.Profit,
		Stock:       product.Stock,
		IsSchedule:  product.IsSchedule,
		Schedule:    product.Schedule,
		Qty:         product.Qty,
		IsHaveStock: product.IsHaveStock,
		CreatedAt:   product.CreatedAt,
		UpdatedAt:   product.UpdatedAt,
	}
}

// MapProductsToSyncDTOs converts a slice of Product entities to SyncProductDTOs
func MapProductsToSyncDTOs(products []entities.Product) []SyncProductDTO {
	result := make([]SyncProductDTO, len(products))
	for i, product := range products {
		result[i] = MapProductToSyncDTO(product)
	}
	return result
}

// MapExpenseToSyncDTO converts an Expense entity to SyncExpenseDTO
func MapExpenseToSyncDTO(expense entities.Expense) SyncExpenseDTO {
	return SyncExpenseDTO{
		ID:        expense.ID,
		ShopID:    expense.ShopID,
		Nominal:   expense.Nominal,
		Status:    string(expense.Status),
		Date:      expense.Date,
		Label:     expense.Label,
		Desc:      expense.Desc,
		CreatedAt: expense.CreatedAt,
		UpdatedAt: expense.UpdatedAt,
	}
}

// MapExpensesToSyncDTOs converts a slice of Expense entities to SyncExpenseDTOs
func MapExpensesToSyncDTOs(expenses []entities.Expense) []SyncExpenseDTO {
	result := make([]SyncExpenseDTO, len(expenses))
	for i, expense := range expenses {
		result[i] = MapExpenseToSyncDTO(expense)
	}
	return result
}

// MapPaymentToSyncDTO converts a Payment entity to SyncPaymentDTO
func MapPaymentToSyncDTO(payment entities.Payment) SyncPaymentDTO {
	return SyncPaymentDTO{
		ID:            payment.ID,
		ShopID:        payment.ShopID,
		UserID:        payment.UserID,
		TransactionID: payment.TransactionID,
		Status:        string(payment.Status),
		Total:         payment.Total,
		CreatedAt:     payment.CreatedAt,
		UpdatedAt:     payment.UpdatedAt,
	}
}

// MapPaymentsToSyncDTOs converts a slice of Payment entities to SyncPaymentDTOs
func MapPaymentsToSyncDTOs(payments []entities.Payment) []SyncPaymentDTO {
	result := make([]SyncPaymentDTO, len(payments))
	for i, payment := range payments {
		result[i] = MapPaymentToSyncDTO(payment)
	}
	return result
}

// MapTransactionToSyncDTO converts a Transaction entity to SyncTransactionDTO
func MapTransactionToSyncDTO(transaction entities.Transaction) SyncTransactionDTO {
	return SyncTransactionDTO{
		ID:                   transaction.ID,
		ShopID:               transaction.ShopID,
		CashierID:            transaction.CashierID,
		CustomerName:         transaction.CustomerName,
		Discount:             transaction.Discount,
		DiscountPercentage:   transaction.DiscountPercentage,
		AdditionalCost:       transaction.AdditionalCost,
		Status:               string(transaction.Status),
		TotalPrice:           transaction.TotalPrice,
		ProfitTransaction:    transaction.ProfitTransaction,
		CashierName:          transaction.CashierName,
		Change:               transaction.Change,
		Amount:               transaction.Amount,
		InitialPaymentStatus: transaction.InitialPaymentStatus,
		CreatedAt:            transaction.CreatedAt,
		UpdatedAt:            transaction.UpdatedAt,
	}
}

// MapTransactionsToSyncDTOs converts a slice of Transaction entities to SyncTransactionDTOs
func MapTransactionsToSyncDTOs(transactions []entities.Transaction) []SyncTransactionDTO {
	result := make([]SyncTransactionDTO, len(transactions))
	for i, transaction := range transactions {
		result[i] = MapTransactionToSyncDTO(transaction)
	}
	return result
}

// MapReceiptToSyncDTO converts a Receipt entity to SyncReceiptDTO
func MapReceiptToSyncDTO(receipt entities.Receipt) SyncReceiptDTO {
	return SyncReceiptDTO{
		ID:         receipt.ID,
		ShopID:     receipt.ShopID,
		PaymentsID: receipt.PaymentsID,
		CreatedAt:  receipt.CreatedAt,
		UpdatedAt:  receipt.UpdatedAt,
	}
}

// MapReceiptsToSyncDTOs converts a slice of Receipt entities to SyncReceiptDTOs
func MapReceiptsToSyncDTOs(receipts []entities.Receipt) []SyncReceiptDTO {
	result := make([]SyncReceiptDTO, len(receipts))
	for i, receipt := range receipts {
		result[i] = MapReceiptToSyncDTO(receipt)
	}
	return result
}

// MapHistoryToSyncDTO converts a History entity to SyncHistoryDTO
func MapHistoryToSyncDTO(history entities.History) SyncHistoryDTO {
	return SyncHistoryDTO{
		ID:            history.ID,
		ShopID:        history.ShopID,
		TransactionID: history.TransactionID,
		CreatedAt:     history.CreatedAt,
		UpdatedAt:     history.UpdatedAt,
	}
}

// MapHistoriesToSyncDTOs converts a slice of History entities to SyncHistoryDTOs
func MapHistoriesToSyncDTOs(histories []entities.History) []SyncHistoryDTO {
	result := make([]SyncHistoryDTO, len(histories))
	for i, history := range histories {
		result[i] = MapHistoryToSyncDTO(history)
	}
	return result
}

// MapShopToSyncDTO converts a Shop entity to SyncShopDTO
func MapShopToSyncDTO(shop entities.Shop) SyncShopDTO {
	return SyncShopDTO{
		ID:              shop.ID,
		LicenseID:       shop.LicenseID,
		UserID:          shop.UserID,
		Name:            shop.Name,
		Domain:          shop.Domain,
		Photo:           shop.Photo,
		Address:         shop.Address,
		Slogan:          shop.Slogan,
		ProfitCalculate: shop.ProfitCalculate,
		CreatedAt:       shop.CreatedAt,
		UpdatedAt:       shop.UpdatedAt,
	}
}

// MapShopsToSyncDTOs converts a slice of Shop entities to SyncShopDTOs
func MapShopsToSyncDTOs(shops []entities.Shop) []SyncShopDTO {
	result := make([]SyncShopDTO, len(shops))
	for i, shop := range shops {
		result[i] = MapShopToSyncDTO(shop)
	}
	return result
}

// MapStockHistoryToSyncDTO converts a StockHistory entity to SyncStockHistoryDTO
func MapStockHistoryToSyncDTO(stockHistory entities.StockHistory) SyncStockHistoryDTO {
	return SyncStockHistoryDTO{
		ID:        stockHistory.ID,
		ProductID: stockHistory.ProductID,
		Stock:     stockHistory.Stock,
		LastStock: stockHistory.LastStock,
		StockedAt: stockHistory.StockedAt,
		CreatedAt: stockHistory.CreatedAt,
		UpdatedAt: stockHistory.UpdatedAt,
	}
}

// MapStockHistoriesToSyncDTOs converts a slice of StockHistory entities to SyncStockHistoryDTOs
func MapStockHistoriesToSyncDTOs(stockHistories []entities.StockHistory) []SyncStockHistoryDTO {
	result := make([]SyncStockHistoryDTO, len(stockHistories))
	for i, stockHistory := range stockHistories {
		result[i] = MapStockHistoryToSyncDTO(stockHistory)
	}
	return result
}

// MapTransactionProductToSyncDTO converts a TransactionProduct entity to SyncTransactionProductDTO
func MapTransactionProductToSyncDTO(transactionProduct entities.TransactionProduct) SyncTransactionProductDTO {
	return SyncTransactionProductDTO{
		ID:            transactionProduct.ID,
		TransactionID: transactionProduct.TransactionID,
		ProductID:     transactionProduct.ProductID,
		Quantity:      transactionProduct.Quantity,
		UnitPrice:     transactionProduct.UnitPrice,
		TotalPrice:    transactionProduct.TotalPrice,
		CreatedAt:     transactionProduct.CreatedAt,
		UpdatedAt:     transactionProduct.UpdatedAt,
	}
}

// MapTransactionProductsToSyncDTOs converts a slice of TransactionProduct entities to SyncTransactionProductDTOs
func MapTransactionProductsToSyncDTOs(transactionProducts []entities.TransactionProduct) []SyncTransactionProductDTO {
	result := make([]SyncTransactionProductDTO, len(transactionProducts))
	for i, transactionProduct := range transactionProducts {
		result[i] = MapTransactionProductToSyncDTO(transactionProduct)
	}
	return result
}

// MapUserToSyncDTO converts a User entity to SyncUserDTO
func MapUserToSyncDTO(user entities.User) SyncUserDTO {
	return SyncUserDTO{
		ID:              user.ID,
		LicenseID:       user.LicenseID,
		RoleID:          user.RoleID,
		ShopID:          user.ShopID,
		Email:           user.Email,
		EmailVerifiedAt: user.EmailVerifiedAt,
		Username:        user.Username,
		Name:            user.Name,
		InfoDevice:      user.InfoDevice,
		FCMToken:        user.FCMToken,
		CreatedAt:       user.CreatedAt,
		UpdatedAt:       user.UpdatedAt,
	}
}

// MapUsersToSyncDTOs converts a slice of User entities to SyncUserDTOs
func MapUsersToSyncDTOs(users []entities.User) []SyncUserDTO {
	result := make([]SyncUserDTO, len(users))
	for i, user := range users {
		result[i] = MapUserToSyncDTO(user)
	}
	return result
}
