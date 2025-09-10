package repositories

import (
	"time"

	"github.com/google/uuid"
	"github.com/terminator791/t-pos/internal/domain/entities"
	"github.com/terminator791/t-pos/internal/domain/sync"
	"gorm.io/gorm"
)

// SyncRepository handles batch operations and delta queries for synchronization
type SyncRepository struct {
	db *gorm.DB
}

// NewSyncRepository creates a new sync repository
func NewSyncRepository(db *gorm.DB) *SyncRepository {
	return &SyncRepository{db: db}
}

// GetChangedDataSince retrieves all data changed since the given timestamp for a license
func (sr *SyncRepository) GetChangedDataSince(licenseID uuid.UUID, since *time.Time, shopID *uuid.UUID) (*sync.SyncData, error) {
	data := &sync.SyncData{}
	var err error

	// Build base query with license filtering
	baseQuery := sr.db.Where("deleted_at IS NULL")
	
	// Add timestamp filter if provided
	if since != nil {
		baseQuery = baseQuery.Where("updated_at > ?", since)
	}

	// Get shops (with license filtering)
	var shops []entities.Shop
	shopQuery := baseQuery.Where("license_id = ?", licenseID)
	if shopID != nil {
		shopQuery = shopQuery.Where("id = ?", *shopID)
	}
	if err = shopQuery.Find(&shops).Error; err != nil {
		return nil, err
	}
	data.Shops = shops

	// Get shop IDs for filtering other entities
	shopIDs := make([]uuid.UUID, len(shops))
	for i, shop := range shops {
		shopIDs[i] = shop.ID
	}

	if len(shopIDs) == 0 {
		// No shops found, return empty data
		return data, nil
	}

	// Get categories
	categoryQuery := baseQuery.Where("shop_id IN ?", shopIDs)
	if err = categoryQuery.Find(&data.Categories).Error; err != nil {
		return nil, err
	}

	// Get products
	productQuery := baseQuery.Where("shop_id IN ?", shopIDs)
	if err = productQuery.Find(&data.Products).Error; err != nil {
		return nil, err
	}

	// Get carts
	cartQuery := baseQuery.Where("shop_id IN ?", shopIDs)
	if err = cartQuery.Find(&data.Carts).Error; err != nil {
		return nil, err
	}

	// Get transactions
	transactionQuery := baseQuery.Where("shop_id IN ?", shopIDs)
	if err = transactionQuery.Find(&data.Transactions).Error; err != nil {
		return nil, err
	}

	// Get expenses
	expenseQuery := baseQuery.Where("shop_id IN ?", shopIDs)
	if err = expenseQuery.Find(&data.Expenses).Error; err != nil {
		return nil, err
	}

	// Get payments
	paymentQuery := baseQuery.Where("shop_id IN ?", shopIDs)
	if err = paymentQuery.Find(&data.Payments).Error; err != nil {
		return nil, err
	}

	// Get receipts
	receiptQuery := baseQuery.Where("shop_id IN ?", shopIDs)
	if err = receiptQuery.Find(&data.Receipts).Error; err != nil {
		return nil, err
	}

	// Get histories
	historyQuery := baseQuery.Where("shop_id IN ?", shopIDs)
	if err = historyQuery.Find(&data.Histories).Error; err != nil {
		return nil, err
	}

	// Get transaction products (join with transactions to filter by shop)
	if err = baseQuery.
		Joins("JOIN transactions t ON transaction_products.transaction_id = t.id").
		Where("t.shop_id IN ? AND t.deleted_at IS NULL", shopIDs).
		Find(&data.TransactionProducts).Error; err != nil {
		return nil, err
	}

	// Get stock histories (join with products to filter by shop)
	if err = baseQuery.
		Joins("JOIN products p ON stock_histories.product_id = p.id").
		Where("p.shop_id IN ? AND p.deleted_at IS NULL", shopIDs).
		Find(&data.StockHistories).Error; err != nil {
		return nil, err
	}

	// Get users (filter by license)
	userQuery := baseQuery.Where("license_id = ?", licenseID)
	if shopID != nil {
		// For cashiers, also include users assigned to the specific shop
		userQuery = userQuery.Where("shop_id IS NULL OR shop_id = ?", *shopID)
	}
	if err = userQuery.Find(&data.Users).Error; err != nil {
		return nil, err
	}

	return data, nil
}

// BatchUpsertCarts performs batch upsert operations for carts
func (sr *SyncRepository) BatchUpsertCarts(carts []entities.Cart) error {
	if len(carts) == 0 {
		return nil
	}

	// Use ON CONFLICT DO UPDATE for PostgreSQL
	return sr.db.Transaction(func(tx *gorm.DB) error {
		for _, cart := range carts {
			if err := tx.Save(&cart).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// BatchUpsertCategories performs batch upsert operations for categories
func (sr *SyncRepository) BatchUpsertCategories(categories []entities.Category) error {
	if len(categories) == 0 {
		return nil
	}

	return sr.db.Transaction(func(tx *gorm.DB) error {
		for _, category := range categories {
			if err := tx.Save(&category).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// BatchUpsertProducts performs batch upsert operations for products
func (sr *SyncRepository) BatchUpsertProducts(products []entities.Product) error {
	if len(products) == 0 {
		return nil
	}

	return sr.db.Transaction(func(tx *gorm.DB) error {
		for _, product := range products {
			if err := tx.Save(&product).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// BatchUpsertTransactions performs batch upsert operations for transactions
func (sr *SyncRepository) BatchUpsertTransactions(transactions []entities.Transaction) error {
	if len(transactions) == 0 {
		return nil
	}

	return sr.db.Transaction(func(tx *gorm.DB) error {
		for _, transaction := range transactions {
			if err := tx.Save(&transaction).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// BatchUpsertExpenses performs batch upsert operations for expenses
func (sr *SyncRepository) BatchUpsertExpenses(expenses []entities.Expense) error {
	if len(expenses) == 0 {
		return nil
	}

	return sr.db.Transaction(func(tx *gorm.DB) error {
		for _, expense := range expenses {
			if err := tx.Save(&expense).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// BatchUpsertPayments performs batch upsert operations for payments
func (sr *SyncRepository) BatchUpsertPayments(payments []entities.Payment) error {
	if len(payments) == 0 {
		return nil
	}

	return sr.db.Transaction(func(tx *gorm.DB) error {
		for _, payment := range payments {
			if err := tx.Save(&payment).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// BatchUpsertReceipts performs batch upsert operations for receipts
func (sr *SyncRepository) BatchUpsertReceipts(receipts []entities.Receipt) error {
	if len(receipts) == 0 {
		return nil
	}

	return sr.db.Transaction(func(tx *gorm.DB) error {
		for _, receipt := range receipts {
			if err := tx.Save(&receipt).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// BatchUpsertShops performs batch upsert operations for shops
func (sr *SyncRepository) BatchUpsertShops(shops []entities.Shop) error {
	if len(shops) == 0 {
		return nil
	}

	return sr.db.Transaction(func(tx *gorm.DB) error {
		for _, shop := range shops {
			if err := tx.Save(&shop).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// BatchUpsertHistories performs batch upsert operations for histories
func (sr *SyncRepository) BatchUpsertHistories(histories []entities.History) error {
	if len(histories) == 0 {
		return nil
	}

	return sr.db.Transaction(func(tx *gorm.DB) error {
		for _, history := range histories {
			if err := tx.Save(&history).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// BatchUpsertStockHistories performs batch upsert operations for stock histories
func (sr *SyncRepository) BatchUpsertStockHistories(stockHistories []entities.StockHistory) error {
	if len(stockHistories) == 0 {
		return nil
	}

	return sr.db.Transaction(func(tx *gorm.DB) error {
		for _, stockHistory := range stockHistories {
			if err := tx.Save(&stockHistory).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// BatchUpsertTransactionProducts performs batch upsert operations for transaction products
func (sr *SyncRepository) BatchUpsertTransactionProducts(transactionProducts []entities.TransactionProduct) error {
	if len(transactionProducts) == 0 {
		return nil
	}

	return sr.db.Transaction(func(tx *gorm.DB) error {
		for _, transactionProduct := range transactionProducts {
			if err := tx.Save(&transactionProduct).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// BatchUpsertUsers performs batch upsert operations for users
func (sr *SyncRepository) BatchUpsertUsers(users []entities.User) error {
	if len(users) == 0 {
		return nil
	}

	return sr.db.Transaction(func(tx *gorm.DB) error {
		for _, user := range users {
			if err := tx.Save(&user).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// GetExistingEntitiesByIDs retrieves existing entities by their IDs for conflict detection
func (sr *SyncRepository) GetExistingEntitiesByIDs(entityType string, ids []uuid.UUID) (map[uuid.UUID]interface{}, error) {
	result := make(map[uuid.UUID]interface{})
	
	if len(ids) == 0 {
		return result, nil
	}

	switch entityType {
	case "cart":
		var entities []entities.Cart
		if err := sr.db.Where("id IN ? AND deleted_at IS NULL", ids).Find(&entities).Error; err != nil {
			return nil, err
		}
		for _, entity := range entities {
			result[entity.ID] = entity
		}

	case "category":
		var entities []entities.Category
		if err := sr.db.Where("id IN ? AND deleted_at IS NULL", ids).Find(&entities).Error; err != nil {
			return nil, err
		}
		for _, entity := range entities {
			result[entity.ID] = entity
		}

	case "product":
		var entities []entities.Product
		if err := sr.db.Where("id IN ? AND deleted_at IS NULL", ids).Find(&entities).Error; err != nil {
			return nil, err
		}
		for _, entity := range entities {
			result[entity.ID] = entity
		}

	case "transaction":
		var entities []entities.Transaction
		if err := sr.db.Where("id IN ? AND deleted_at IS NULL", ids).Find(&entities).Error; err != nil {
			return nil, err
		}
		for _, entity := range entities {
			result[entity.ID] = entity
		}

	case "expense":
		var entities []entities.Expense
		if err := sr.db.Where("id IN ? AND deleted_at IS NULL", ids).Find(&entities).Error; err != nil {
			return nil, err
		}
		for _, entity := range entities {
			result[entity.ID] = entity
		}

	case "payment":
		var entities []entities.Payment
		if err := sr.db.Where("id IN ? AND deleted_at IS NULL", ids).Find(&entities).Error; err != nil {
			return nil, err
		}
		for _, entity := range entities {
			result[entity.ID] = entity
		}

	case "receipt":
		var entities []entities.Receipt
		if err := sr.db.Where("id IN ? AND deleted_at IS NULL", ids).Find(&entities).Error; err != nil {
			return nil, err
		}
		for _, entity := range entities {
			result[entity.ID] = entity
		}

	case "shop":
		var entities []entities.Shop
		if err := sr.db.Where("id IN ? AND deleted_at IS NULL", ids).Find(&entities).Error; err != nil {
			return nil, err
		}
		for _, entity := range entities {
			result[entity.ID] = entity
		}

	case "history":
		var entities []entities.History
		if err := sr.db.Where("id IN ? AND deleted_at IS NULL", ids).Find(&entities).Error; err != nil {
			return nil, err
		}
		for _, entity := range entities {
			result[entity.ID] = entity
		}

	case "stock_history":
		var entities []entities.StockHistory
		if err := sr.db.Where("id IN ? AND deleted_at IS NULL", ids).Find(&entities).Error; err != nil {
			return nil, err
		}
		for _, entity := range entities {
			result[entity.ID] = entity
		}

	case "transaction_product":
		var entities []entities.TransactionProduct
		if err := sr.db.Where("id IN ? AND deleted_at IS NULL", ids).Find(&entities).Error; err != nil {
			return nil, err
		}
		for _, entity := range entities {
			result[entity.ID] = entity
		}

	case "user":
		var entities []entities.User
		if err := sr.db.Where("id IN ? AND deleted_at IS NULL", ids).Find(&entities).Error; err != nil {
			return nil, err
		}
		for _, entity := range entities {
			result[entity.ID] = entity
		}
	}

	return result, nil
}

// ValidateShopAccess validates that all entities belong to shops accessible by the user
func (sr *SyncRepository) ValidateShopAccess(licenseID uuid.UUID, shopID *uuid.UUID, data *sync.SyncData) error {
	// Get accessible shop IDs
	var accessibleShopIDs []uuid.UUID
	query := sr.db.Model(&entities.Shop{}).Where("license_id = ? AND deleted_at IS NULL", licenseID)
	
	if shopID != nil {
		// Cashier - only their assigned shop
		query = query.Where("id = ?", *shopID)
	}
	
	if err := query.Pluck("id", &accessibleShopIDs).Error; err != nil {
		return err
	}

	accessibleShopMap := make(map[uuid.UUID]bool)
	for _, id := range accessibleShopIDs {
		accessibleShopMap[id] = true
	}

	// Validate each entity type
	for _, cart := range data.Carts {
		if !accessibleShopMap[cart.ShopID] {
			return gorm.ErrRecordNotFound
		}
	}

	for _, category := range data.Categories {
		if !accessibleShopMap[category.ShopID] {
			return gorm.ErrRecordNotFound
		}
	}

	for _, product := range data.Products {
		if !accessibleShopMap[product.ShopID] {
			return gorm.ErrRecordNotFound
		}
	}

	for _, transaction := range data.Transactions {
		if !accessibleShopMap[transaction.ShopID] {
			return gorm.ErrRecordNotFound
		}
	}

	for _, expense := range data.Expenses {
		if !accessibleShopMap[expense.ShopID] {
			return gorm.ErrRecordNotFound
		}
	}

	for _, payment := range data.Payments {
		if !accessibleShopMap[payment.ShopID] {
			return gorm.ErrRecordNotFound
		}
	}

	for _, receipt := range data.Receipts {
		if !accessibleShopMap[receipt.ShopID] {
			return gorm.ErrRecordNotFound
		}
	}

	for _, shop := range data.Shops {
		if shop.LicenseID != licenseID {
			return gorm.ErrRecordNotFound
		}
		if shopID != nil && shop.ID != *shopID {
			return gorm.ErrRecordNotFound
		}
	}

	for _, history := range data.Histories {
		if !accessibleShopMap[history.ShopID] {
			return gorm.ErrRecordNotFound
		}
	}

	// For users, validate they belong to the same license
	for _, user := range data.Users {
		if user.LicenseID == nil || *user.LicenseID != licenseID {
			return gorm.ErrRecordNotFound
		}
	}

	return nil
}