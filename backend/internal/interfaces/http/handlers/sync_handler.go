package handlers

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/terminator791/t-pos/internal/application/services"
	"github.com/terminator791/t-pos/internal/domain/dto"
	"github.com/terminator791/t-pos/internal/domain/repositories"
	"github.com/terminator791/t-pos/internal/infrastructure/auth"
	"github.com/terminator791/t-pos/pkg/response"
)

// SyncHandler handles synchronization related HTTP requests
type SyncHandler struct {
	syncService *services.SyncService
	userRepo    repositories.UserRepository
}

// NewSyncHandler creates a new sync handler
func NewSyncHandler(syncService *services.SyncService, userRepo repositories.UserRepository) *SyncHandler {
	return &SyncHandler{
		syncService: syncService,
		userRepo:    userRepo,
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

	// Get user entity to extract license ID
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

	// Parse sync request
	var syncRequest dto.SyncRequest
	if err := c.ShouldBindJSON(&syncRequest); err != nil {
		response.ErrorBadRequest(c, "Invalid sync request format", err.Error())
		return
	}

	// Validate sync request
	if err := h.validateSyncRequest(&syncRequest); err != nil {
		response.ErrorBadRequest(c, "Invalid sync request data", err.Error())
		return
	}

	// Log sync request
	log.Printf("Processing sync request for user %s, license %s", userID.String(), licenseID.String())

	// Process synchronization
	syncResponse, err := h.syncService.ProcessSync(c.Request.Context(), syncRequest, licenseID, userID)
	if err != nil {
		log.Printf("Sync failed for user %s: %v", userID.String(), err)
		response.ErrorInternalServer(c, "Sync processing failed", err.Error())
		return
	}

	// Log sync results
	log.Printf("Sync completed for user %s: %d conflicts, %d errors, %dms",
		userID.String(), syncResponse.Stats.ConflictCount, syncResponse.Stats.ErrorCount, syncResponse.Stats.ProcessingTimeMs)

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
