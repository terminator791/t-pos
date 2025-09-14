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
	"github.com/terminator791/t-pos/internal/infrastructure/auth"
	"github.com/terminator791/t-pos/pkg/response"
)

// SyncHandler handles synchronization related HTTP requests
type SyncHandler struct {
	syncService *services.SyncService
	userRepo    repositories.UserRepository
	shopRepo    repositories.ShopRepository
	roleRepo    repositories.RoleRepository
}

// NewSyncHandler creates a new sync handler
func NewSyncHandler(syncService *services.SyncService, userRepo repositories.UserRepository, shopRepo repositories.ShopRepository, roleRepo repositories.RoleRepository) *SyncHandler {
	return &SyncHandler{
		syncService: syncService,
		userRepo:    userRepo,
		shopRepo:    shopRepo,
		roleRepo:    roleRepo,
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
		log.Printf("Owner business user %s syncing %d shops under license %s", userID.String(), len(accessibleShopIDs), licenseID.String())
		
	case "cashier":
		// Cashier can only sync their assigned shop
		if user.ShopID == nil {
			response.ErrorUnauthorized(c, "Cashier is not assigned to a shop", nil)
			return
		}
		accessibleShopIDs = append(accessibleShopIDs, *user.ShopID)
		log.Printf("Cashier user %s syncing single shop %s", userID.String(), user.ShopID.String())
		
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

	// Validate sync request based on user role and access
	if err := h.validateSyncRequestWithRoleAccess(&syncRequest, userRole, accessibleShopIDs); err != nil {
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

// validateSyncRequestWithRoleAccess validates the incoming sync request with role-based access control
func (h *SyncHandler) validateSyncRequestWithRoleAccess(req *dto.SyncRequest, userRole string, accessibleShopIDs []uuid.UUID) error {
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
