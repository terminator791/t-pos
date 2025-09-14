package handlers

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/terminator791/t-pos/internal/application/services"
	"github.com/terminator791/t-pos/internal/domain/dto"
	"github.com/terminator791/t-pos/internal/domain/entities"
	"github.com/terminator791/t-pos/internal/domain/repositories"
	"github.com/terminator791/t-pos/internal/domain/validators"
	"github.com/terminator791/t-pos/internal/infrastructure/auth"
	"github.com/terminator791/t-pos/pkg/response"
)

// SyncHandler handles synchronization related HTTP requests
type SyncHandler struct {
	syncService     *services.SyncService
	userRepo        repositories.UserRepository
	shopRepo        repositories.ShopRepository
	roleRepo        repositories.RoleRepository
	productRepo     repositories.ProductRepository
	transactionRepo repositories.TransactionRepository
}

// NewSyncHandler creates a new sync handler
func NewSyncHandler(syncService *services.SyncService, userRepo repositories.UserRepository, shopRepo repositories.ShopRepository, roleRepo repositories.RoleRepository, productRepo repositories.ProductRepository, transactionRepo repositories.TransactionRepository) *SyncHandler {
	return &SyncHandler{
		syncService:     syncService,
		userRepo:        userRepo,
		shopRepo:        shopRepo,
		roleRepo:        roleRepo,
		productRepo:     productRepo,
		transactionRepo: transactionRepo,
	}
}

// ProcessSync handles the main synchronization endpoint
// POST /api/v1/sync
func (h *SyncHandler) ProcessSync(c *gin.Context) {
	// Get user info from JWT token
	userID, exists := auth.GetUserIDFromContext(c)
	if !exists {
		response.ErrorUnauthorized(c, "User not authenticated", nil)
		return
	}

	// Get user entity to extract license ID and role information
	user, err := h.userRepo.GetByID(c.Request.Context(), userID)
	if err != nil {
		response.ErrorNotFound(c, "User not found", nil)
		return
	}

	// Validate user has license ID
	if user.LicenseID == nil {
		response.ErrorUnauthorized(c, "User is not associated with a license", nil)
		return
	}

	licenseID := *user.LicenseID

	// Get user role for sync authorization
	var userRole string
	var accessibleShopIDs []uuid.UUID
	
	if user.RoleID != nil {
		role, err := h.roleRepo.GetByID(c.Request.Context(), *user.RoleID)
		if err != nil {
			response.ErrorInternalServer(c, "Failed to get user role", err.Error())
			return
		}
		userRole = role.Name
	}

	// Determine sync access based on role
	switch userRole {
	case "super_admin", "admin":
		// Global access - no shop filtering needed
		log.Printf("Admin user %s performing sync with global access", userID.String())
		
	case "owner_business":
		// Owner can sync all shops under their license
		shops, err := h.getShopsByLicenseID(c.Request.Context(), licenseID)
		if err != nil {
			response.ErrorInternalServer(c, "Failed to get shops for license", err.Error())
			return
		}
		for _, shop := range shops {
			accessibleShopIDs = append(accessibleShopIDs, shop.ID)
		}
		log.Printf("DEBUG: Owner business user %s syncing %d shops under license %s: %v", userID.String(), len(accessibleShopIDs), licenseID.String(), accessibleShopIDs)
		
	case "cashier":
		// Cashier can only sync their assigned shop
		if user.ShopID == nil {
			response.ErrorUnauthorized(c, "Cashier is not assigned to a shop", nil)
			return
		}
		accessibleShopIDs = append(accessibleShopIDs, *user.ShopID)
		log.Printf("DEBUG: Cashier user %s syncing single shop %s: [%s]", userID.String(), user.ShopID.String(), *user.ShopID)
		
	default:
		response.ErrorUnauthorized(c, fmt.Sprintf("Role '%s' is not authorized for sync operations", userRole), nil)
		return
	}

	// Parse sync request
	var syncRequest dto.SyncRequest
	if err := c.ShouldBindJSON(&syncRequest); err != nil {
		response.ErrorBadRequest(c, "Invalid sync request format", err.Error())
		return
	}

	// CRITICAL FIX: Handle shop_id requirements based on role
	if err := h.handleShopIDRequirements(&syncRequest, userRole, user); err != nil {
		response.ErrorBadRequest(c, "Shop ID validation failed", err.Error())
		return
	}

	// CRITICAL FIX: Validate enum fields and cashier IDs in sync request
	if err := h.validateSyncRequestEnumsAndCashiers(&syncRequest); err != nil {
		response.ErrorBadRequest(c, "Sync request validation failed", err.Error())
		return
	}

	// Validate sync request based on user role and access with enhanced domain validation
	if err := h.validateSyncRequestWithRoleAccess(&syncRequest, userRole, accessibleShopIDs, user); err != nil {
		response.ErrorBadRequest(c, "Invalid sync request data", err.Error())
		return
	}

	// Log sync request with role context
	log.Printf("Processing sync request for user %s (role: %s), license %s, accessible shops: %v", 
		userID.String(), userRole, licenseID.String(), accessibleShopIDs)

	// Create sync context with role-based access control
	syncContext := dto.SyncContext{
		UserID:            userID,
		UserRole:          userRole,
		LicenseID:         licenseID,
		AccessibleShopIDs: accessibleShopIDs,
		HasGlobalAccess:   userRole == "super_admin" || userRole == "admin",
	}

	// Process synchronization with role-based filtering
	syncResponse, err := h.syncService.ProcessSyncWithRoleAccess(c.Request.Context(), syncRequest, syncContext)
	if err != nil {
		log.Printf("Sync failed for user %s (role: %s): %v", userID.String(), userRole, err)
		response.ErrorInternalServer(c, "Sync processing failed", err.Error())
		return
	}

	// Log sync results with performance metrics
	log.Printf("Sync completed for user %s (role: %s): %d conflicts, %d errors, %dms, shops accessed: %v",
		userID.String(), userRole, syncResponse.Stats.ConflictCount, syncResponse.Stats.ErrorCount, 
		syncResponse.Stats.ProcessingTimeMs, accessibleShopIDs)

	// Return successful response
	response.SuccessOK(c, "Sync completed successfully", syncResponse)
}

// GetSyncInfo provides information about sync capabilities and status
// GET /api/v1/sync/info
func (h *SyncHandler) GetSyncInfo(c *gin.Context) {
	userID, exists := auth.GetUserIDFromContext(c)
	if !exists {
		response.ErrorUnauthorized(c, "User not authenticated", nil)
		return
	}

	// Get user entity to extract license ID
	user, err := h.userRepo.GetByID(c.Request.Context(), userID)
	if err != nil {
		response.ErrorNotFound(c, "User not found", nil)
		return
	}

	// Get domain from JWT claims
	domain, _ := auth.GetUserDomainFromContext(c)

	syncInfo := gin.H{
		"sync_version":             "1.0.0",
		"supported_entities":       getSupportedEntities(),
		"conflict_resolution":      "last_write_wins",
		"max_entities_per_request": 1000,
		"user_id":                  userID.String(),
		"license_id":               user.LicenseID,
		"domain":                   domain,
	}

	response.SuccessOK(c, "Sync info retrieved successfully", syncInfo)
}

// handleShopIDRequirements handles shop_id requirements based on user role
func (h *SyncHandler) handleShopIDRequirements(req *dto.SyncRequest, userRole string, user *entities.User) error {
	switch userRole {
	case "cashier":
		// CRITICAL: For cashiers, shop_id should NOT be provided in request body
		// Get shop_id from authenticated user and inject it into all entities
		if user.ShopID == nil {
			return fmt.Errorf("cashier user is not assigned to a shop")
		}
		
		cashierShopID := *user.ShopID
		log.Printf("Cashier sync: Automatically applying shop_id %s to all entities", cashierShopID)
		
		// FIRST: Validate that no entities in request body have different shop_ids (before injection)
		if err := h.validateCashierEntitiesShopID(req, cashierShopID); err != nil {
			return fmt.Errorf("cashier domain validation failed: %w", err)
		}
		
		// THEN: Inject shop_id into all entities that have a shop_id field
		h.injectShopIDIntoEntities(req, cashierShopID)
		
	case "owner_business", "admin", "super_admin":
		// For non-cashier roles, shop_id must be provided in request body
		// This allows them to specify which shop they're working with
		if err := h.validateNonCashierShopIDRequirements(req, userRole, user); err != nil {
			return fmt.Errorf("shop_id validation failed for %s: %w", userRole, err)
		}
		
	default:
		return fmt.Errorf("unknown user role: %s", userRole)
	}
	
	return nil
}

// injectShopIDIntoEntities injects the cashier's shop_id into all entities
func (h *SyncHandler) injectShopIDIntoEntities(req *dto.SyncRequest, shopID uuid.UUID) {
	// Inject shop_id into all entities that have ShopID field
	for i := range req.Carts {
		req.Carts[i].ShopID = shopID
	}
	for i := range req.Categories {
		req.Categories[i].ShopID = shopID
	}
	for i := range req.Products {
		req.Products[i].ShopID = shopID
	}
	for i := range req.Transactions {
		req.Transactions[i].ShopID = shopID
	}
	for i := range req.Expenses {
		req.Expenses[i].ShopID = shopID
	}
	for i := range req.Payments {
		req.Payments[i].ShopID = shopID
	}
	for i := range req.Receipts {
		req.Receipts[i].ShopID = shopID
	}
	for i := range req.Histories {
		req.Histories[i].ShopID = shopID
	}
}

// validateCashierEntitiesShopID validates that all entities belong to the cashier's shop
func (h *SyncHandler) validateCashierEntitiesShopID(req *dto.SyncRequest, expectedShopID uuid.UUID) error {
	// Check if any entities had different shop_ids (would indicate client error)
	// For cashiers, shop_id can be nil (will be auto-injected) or must match the cashier's shop
	var invalidEntities []string
	
	for i, cart := range req.Carts {
		if cart.ShopID != uuid.Nil && cart.ShopID != expectedShopID {
			invalidEntities = append(invalidEntities, fmt.Sprintf("cart[%d]: %s", i, cart.ShopID))
		}
	}
	for i, category := range req.Categories {
		if category.ShopID != uuid.Nil && category.ShopID != expectedShopID {
			invalidEntities = append(invalidEntities, fmt.Sprintf("category[%d]: %s", i, category.ShopID))
		}
	}
	for i, product := range req.Products {
		if product.ShopID != uuid.Nil && product.ShopID != expectedShopID {
			invalidEntities = append(invalidEntities, fmt.Sprintf("product[%d]: %s", i, product.ShopID))
		}
	}
	for i, transaction := range req.Transactions {
		if transaction.ShopID != uuid.Nil && transaction.ShopID != expectedShopID {
			invalidEntities = append(invalidEntities, fmt.Sprintf("transaction[%d]: %s", i, transaction.ShopID))
		}
	}
	for i, expense := range req.Expenses {
		if expense.ShopID != uuid.Nil && expense.ShopID != expectedShopID {
			invalidEntities = append(invalidEntities, fmt.Sprintf("expense[%d]: %s", i, expense.ShopID))
		}
	}
	for i, payment := range req.Payments {
		if payment.ShopID != uuid.Nil && payment.ShopID != expectedShopID {
			invalidEntities = append(invalidEntities, fmt.Sprintf("payment[%d]: %s", i, payment.ShopID))
		}
	}
	for i, receipt := range req.Receipts {
		if receipt.ShopID != uuid.Nil && receipt.ShopID != expectedShopID {
			invalidEntities = append(invalidEntities, fmt.Sprintf("receipt[%d]: %s", i, receipt.ShopID))
		}
	}
	for i, history := range req.Histories {
		if history.ShopID != uuid.Nil && history.ShopID != expectedShopID {
			invalidEntities = append(invalidEntities, fmt.Sprintf("history[%d]: %s", i, history.ShopID))
		}
	}
	
	if len(invalidEntities) > 0 {
		return fmt.Errorf("entities with invalid shop_id detected (cashiers can only sync their assigned shop %s): %v", expectedShopID, invalidEntities)
	}
	
	return nil
}

// validateNonCashierShopIDRequirements validates shop_id requirements for non-cashier roles
func (h *SyncHandler) validateNonCashierShopIDRequirements(req *dto.SyncRequest, userRole string, user *entities.User) error {
	// Collect all shop_ids mentioned in the request
	shopIDsInRequest := make(map[uuid.UUID]bool)
	
	for _, cart := range req.Carts {
		shopIDsInRequest[cart.ShopID] = true
	}
	for _, category := range req.Categories {
		shopIDsInRequest[category.ShopID] = true
	}
	for _, product := range req.Products {
		shopIDsInRequest[product.ShopID] = true
	}
	for _, transaction := range req.Transactions {
		shopIDsInRequest[transaction.ShopID] = true
	}
	for _, expense := range req.Expenses {
		shopIDsInRequest[expense.ShopID] = true
	}
	for _, payment := range req.Payments {
		shopIDsInRequest[payment.ShopID] = true
	}
	for _, receipt := range req.Receipts {
		shopIDsInRequest[receipt.ShopID] = true
	}
	for _, history := range req.Histories {
		shopIDsInRequest[history.ShopID] = true
	}
	for _, shop := range req.Shops {
		shopIDsInRequest[shop.ID] = true
	}
	
	// For owner_business, validate all shop_ids belong to their license
	if userRole == "owner_business" {
		if user.LicenseID == nil {
			return fmt.Errorf("owner_business user has no license assignment")
		}
		
		// Validate each shop_id belongs to the user's license
		for shopID := range shopIDsInRequest {
			shop, err := h.shopRepo.GetByID(context.Background(), shopID)
			if err != nil {
				return fmt.Errorf("invalid shop_id %s: shop not found", shopID)
			}
			
			if shop.LicenseID != *user.LicenseID {
				return fmt.Errorf("shop_id %s belongs to license %s, but user belongs to license %s (domain mismatch)", 
					shopID, shop.LicenseID, *user.LicenseID)
			}
		}
		
		// Validate products and other referenced entities belong to correct domain
		if err := h.validateProductDomainAccess(req, *user.LicenseID); err != nil {
			return fmt.Errorf("product domain validation failed: %w", err)
		}
	}
	
	// For admin/super_admin, they can access any shop but we still validate they exist
	if userRole == "admin" || userRole == "super_admin" {
		for shopID := range shopIDsInRequest {
			_, err := h.shopRepo.GetByID(context.Background(), shopID)
			if err != nil {
				return fmt.Errorf("invalid shop_id %s: shop not found", shopID)
			}
		}
	}
	
	return nil
}

// validateProductDomainAccess validates that referenced products belong to the correct domain
func (h *SyncHandler) validateProductDomainAccess(req *dto.SyncRequest, userLicenseID uuid.UUID) error {
	// Check stock histories reference products in correct domain
	for i, stockHistory := range req.StockHistories {
		// First check if product is in the sync request
		var productFound bool
		var productShopID uuid.UUID
		
		for _, product := range req.Products {
			if product.ID == stockHistory.ProductID {
				productShopID = product.ShopID
				productFound = true
				break
			}
		}
		
		// If not in sync request, check database
		if !productFound {
			product, err := h.productRepo.GetByID(context.Background(), stockHistory.ProductID)
			if err != nil {
				return fmt.Errorf("stock_history[%d] references invalid product %s", i, stockHistory.ProductID)
			}
			productShopID = product.ShopID
		}
		
		// Validate the product's shop belongs to user's license
		shop, err := h.shopRepo.GetByID(context.Background(), productShopID)
		if err != nil {
			return fmt.Errorf("stock_history[%d] references product %s in invalid shop %s", i, stockHistory.ProductID, productShopID)
		}
		
		if shop.LicenseID != userLicenseID {
			return fmt.Errorf("stock_history[%d] references product %s in shop %s (license %s), but user belongs to license %s", 
				i, stockHistory.ProductID, productShopID, shop.LicenseID, userLicenseID)
		}
	}
	
	// Check transaction products reference transactions in correct domain
	for i, transactionProduct := range req.TransactionProducts {
		// First check if transaction is in the sync request
		var transactionFound bool
		var transactionShopID uuid.UUID
		
		for _, transaction := range req.Transactions {
			if transaction.ID == transactionProduct.TransactionID {
				transactionShopID = transaction.ShopID
				transactionFound = true
				break
			}
		}
		
		// If not in sync request, check database
		if !transactionFound {
			transaction, err := h.transactionRepo.GetByID(context.Background(), transactionProduct.TransactionID)
			if err != nil {
				return fmt.Errorf("transaction_product[%d] references invalid transaction %s", i, transactionProduct.TransactionID)
			}
			transactionShopID = transaction.ShopID
		}
		
		// Validate the transaction's shop belongs to user's license
		shop, err := h.shopRepo.GetByID(context.Background(), transactionShopID)
		if err != nil {
			return fmt.Errorf("transaction_product[%d] references transaction %s in invalid shop %s", i, transactionProduct.TransactionID, transactionShopID)
		}
		
		if shop.LicenseID != userLicenseID {
			return fmt.Errorf("transaction_product[%d] references transaction %s in shop %s (license %s), but user belongs to license %s", 
				i, transactionProduct.TransactionID, transactionShopID, shop.LicenseID, userLicenseID)
		}
	}
	
	return nil
}

// validateSyncRequestWithRoleAccess validates the incoming sync request with role-based access control
func (h *SyncHandler) validateSyncRequestWithRoleAccess(req *dto.SyncRequest, userRole string, accessibleShopIDs []uuid.UUID, user *entities.User) error {
	if req == nil {
		return fmt.Errorf("sync request is required")
	}

	// Basic validation
	if err := h.validateSyncRequest(req); err != nil {
		return err
	}

	// CRITICAL FIX: Remove hard validation for role-based access
	// The sync service will now filter entities instead of failing validation
	// This allows partial sync success when some entities are filtered out
	
	switch userRole {
	case "cashier":
		// Log cashier sync attempt for monitoring but don't fail
		if len(accessibleShopIDs) == 0 {
			return fmt.Errorf("cashier has no accessible shops")
		}
		log.Printf("Cashier sync validation: user has access to %d shops, request contains %d entities total", 
			len(accessibleShopIDs), 
			len(req.Carts)+len(req.Categories)+len(req.Products)+len(req.Transactions)+
			len(req.Expenses)+len(req.Payments)+len(req.Receipts)+len(req.Histories)+
			len(req.StockHistories)+len(req.TransactionProducts))
		
		// Note: Entity access validation is now handled by filtering in sync service
		// This prevents hard errors and allows partial sync success
		
	case "owner_business":
		// Owner business validation passed - filtering handled in sync service
		log.Printf("Owner business sync validation passed for %d shops", len(accessibleShopIDs))
		
	case "super_admin", "admin":
		// Global access - no additional validation needed
		log.Printf("Admin sync validation passed with global access")
		
	default:
		return fmt.Errorf("unknown user role: %s", userRole)
	}

	return nil
}

// validateEntitiesShopAccess validates that all entities in the sync request belong to accessible shops
func (h *SyncHandler) validateEntitiesShopAccess(req *dto.SyncRequest, accessibleShopIDs []uuid.UUID) error {
	shopSet := make(map[uuid.UUID]bool)
	for _, shopID := range accessibleShopIDs {
		shopSet[shopID] = true
	}

	// Validate carts
	for i, cart := range req.Carts {
		if !shopSet[cart.ShopID] {
			return fmt.Errorf("cart at index %d belongs to inaccessible shop %s", i, cart.ShopID)
		}
	}

	// Validate categories
	for i, category := range req.Categories {
		if !shopSet[category.ShopID] {
			return fmt.Errorf("category at index %d belongs to inaccessible shop %s", i, category.ShopID)
		}
	}

	// Validate products
	for i, product := range req.Products {
		if !shopSet[product.ShopID] {
			return fmt.Errorf("product at index %d belongs to inaccessible shop %s", i, product.ShopID)
		}
	}

	// Validate transactions
	for i, transaction := range req.Transactions {
		if !shopSet[transaction.ShopID] {
			return fmt.Errorf("transaction at index %d belongs to inaccessible shop %s", i, transaction.ShopID)
		}
	}

	// Validate expenses
	for i, expense := range req.Expenses {
		if !shopSet[expense.ShopID] {
			return fmt.Errorf("expense at index %d belongs to inaccessible shop %s", i, expense.ShopID)
		}
	}

	// Validate payments
	for i, payment := range req.Payments {
		if !shopSet[payment.ShopID] {
			return fmt.Errorf("payment at index %d belongs to inaccessible shop %s", i, payment.ShopID)
		}
	}

	// Validate receipts
	for i, receipt := range req.Receipts {
		if !shopSet[receipt.ShopID] {
			return fmt.Errorf("receipt at index %d belongs to inaccessible shop %s", i, receipt.ShopID)
		}
	}

	// Validate histories
	for i, history := range req.Histories {
		if !shopSet[history.ShopID] {
			return fmt.Errorf("history at index %d belongs to inaccessible shop %s", i, history.ShopID)
		}
	}

	// Validate shops - for cashiers, they shouldn't be syncing shop data at all
	if len(req.Shops) > 0 && len(accessibleShopIDs) == 1 {
		return fmt.Errorf("cashiers cannot sync shop configuration data")
	}

	return nil
}

// getShopsByLicenseID retrieves all shops for a given license ID
func (h *SyncHandler) getShopsByLicenseID(ctx context.Context, licenseID uuid.UUID) ([]entities.Shop, error) {
	shopPtrs, err := h.shopRepo.GetByLicenseID(ctx, licenseID)
	if err != nil {
		return nil, fmt.Errorf("failed to get shops for license %s: %w", licenseID, err)
	}
	
	// Convert []*entities.Shop to []entities.Shop
	shops := make([]entities.Shop, len(shopPtrs))
	for i, shopPtr := range shopPtrs {
		shops[i] = *shopPtr
	}
	
	return shops, nil
}

// validateSyncRequest validates the incoming sync request
func (h *SyncHandler) validateSyncRequest(req *dto.SyncRequest) error {
	if req == nil {
		return fmt.Errorf("sync request is required")
	}

	// Validate entity limits
	maxEntitiesPerType := 1000

	if len(req.Carts) > maxEntitiesPerType {
		return fmt.Errorf("too many carts in sync request (max %d)", maxEntitiesPerType)
	}

	if len(req.Categories) > maxEntitiesPerType {
		return fmt.Errorf("too many categories in sync request (max %d)", maxEntitiesPerType)
	}

	if len(req.Products) > maxEntitiesPerType {
		return fmt.Errorf("too many products in sync request (max %d)", maxEntitiesPerType)
	}

	if len(req.Transactions) > maxEntitiesPerType {
		return fmt.Errorf("too many transactions in sync request (max %d)", maxEntitiesPerType)
	}

	// Validate UUIDs are not nil for entities
	for i, cart := range req.Carts {
		if cart.ID == uuid.Nil {
			return fmt.Errorf("cart at index %d has invalid ID", i)
		}
	}

	for i, category := range req.Categories {
		if category.ID == uuid.Nil {
			return fmt.Errorf("category at index %d has invalid ID", i)
		}
	}

	for i, product := range req.Products {
		if product.ID == uuid.Nil {
			return fmt.Errorf("product at index %d has invalid ID", i)
		}
	}

	for i, transaction := range req.Transactions {
		if transaction.ID == uuid.Nil {
			return fmt.Errorf("transaction at index %d has invalid ID", i)
		}
	}

	return nil
}

// getSupportedEntities returns list of entities supported for sync
func getSupportedEntities() []string {
	return []string{
		"carts",
		"categories",
		"expenses",
		"histories",
		"payments",
		"products",
		"receipts",
		"shops",
		"stock_histories",
		"transaction_products",
		"transactions",
		"users",
	}
}

// Health check endpoint for sync service
// GET /api/v1/sync/health
func (h *SyncHandler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "healthy",
		"service": "sync",
		"version": "1.0.0",
	})
}

// validateSyncRequestEnumsAndCashiers validates enum fields and cashier IDs in the sync request
func (h *SyncHandler) validateSyncRequestEnumsAndCashiers(req *dto.SyncRequest) error {
	// Validate enum fields in expenses
	for i, expense := range req.Expenses {
		if err := validators.ValidateExpenseStatus(expense.Status); err != nil {
			return fmt.Errorf("expense[%d]: %w", i, err)
		}
	}

	// Validate enum fields in payments
	for i, payment := range req.Payments {
		if err := validators.ValidatePaymentStatus(payment.Status); err != nil {
			return fmt.Errorf("payment[%d]: %w", i, err)
		}
	}

	// Validate enum fields and cashier IDs in transactions
	for i, transaction := range req.Transactions {
		// Validate transaction status enum
		if err := validators.ValidateTransactionStatus(transaction.Status); err != nil {
			return fmt.Errorf("transaction[%d]: %w", i, err)
		}

		// Validate cashier_id exists and is a valid cashier user
		if err := h.validateCashierID(transaction.CashierID); err != nil {
			return fmt.Errorf("transaction[%d] cashier validation failed: %w", i, err)
		}
	}

	// Validate and auto-initialize shop domains
	for i := range req.Shops {
		if err := h.validateAndInitializeShopDomain(&req.Shops[i]); err != nil {
			return fmt.Errorf("shop[%d] domain validation failed: %w", i, err)
		}
	}

	return nil
}

// validateCashierID validates that the cashier_id exists and refers to a user with cashier role
func (h *SyncHandler) validateCashierID(cashierID uuid.UUID) error {
	// Get the user by ID
	user, err := h.userRepo.GetByID(context.Background(), cashierID)
	if err != nil {
		return fmt.Errorf("cashier_id %s not found", cashierID)
	}

	// Check if user has a role assigned
	if user.RoleID == nil {
		return fmt.Errorf("cashier_id %s has no role assigned", cashierID)
	}

	// Get the role to verify it's a cashier
	role, err := h.roleRepo.GetByID(context.Background(), *user.RoleID)
	if err != nil {
		return fmt.Errorf("failed to get role for cashier_id %s", cashierID)
	}

	// Verify the role is "cashier"
	if role.Name != "cashier" {
		return fmt.Errorf("cashier_id %s has role '%s', expected 'cashier'", cashierID, role.Name)
	}

	return nil
}

// validateAndInitializeShopDomain ensures shop domain is properly initialized
func (h *SyncHandler) validateAndInitializeShopDomain(shop *entities.Shop) error {
	// If shop ID is nil, generate one
	if shop.ID == uuid.Nil {
		shop.ID = uuid.New()
	}

	// Auto-initialize domain if it's empty (following the entity's BeforeCreate logic)
	if shop.Domain == "" {
		shop.Domain = "shop-" + shop.ID.String()
		log.Printf("Auto-initialized shop domain: %s for shop ID: %s", shop.Domain, shop.ID.String())
	}

	// Validate domain format for existing domains
	expectedDomain := "shop-" + shop.ID.String()
	if shop.Domain != expectedDomain {
		return fmt.Errorf("shop domain '%s' does not match expected format 'shop-%s'", shop.Domain, shop.ID.String())
	}

	return nil
}
