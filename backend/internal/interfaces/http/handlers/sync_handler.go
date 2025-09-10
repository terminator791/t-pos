package handlers

import (
	"context"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/terminator791/t-pos/internal/application/services"
	"github.com/terminator791/t-pos/internal/domain/repositories"
	"github.com/terminator791/t-pos/internal/domain/sync"
	"github.com/terminator791/t-pos/internal/infrastructure/auth"
	"github.com/terminator791/t-pos/pkg/response"
)

// SyncHandler handles HTTP requests for data synchronization
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

// PushSync handles POST /api/v1/sync/push
// @Summary Push data from mobile to server
// @Description Process incoming data from mobile client and resolve conflicts
// @Tags sync
// @Accept json
// @Produce json
// @Param request body sync.SyncRequest true "Sync data from mobile"
// @Success 200 {object} response.Response{data=sync.SyncResponse}
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/sync/push [post]
func (sh *SyncHandler) PushSync(c *gin.Context) {
	var request sync.SyncRequest

	// Parse request body
	if err := c.ShouldBindJSON(&request); err != nil {
		log.Printf("Failed to parse push sync request: %v", err)
		response.Error(c, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	// Get user info from JWT token
	userClaims, exists := c.Get("user")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "User not authenticated", "")
		return
	}

	claims, ok := userClaims.(*auth.Claims)
	if !ok {
		response.Error(c, http.StatusUnauthorized, "Invalid token claims", "")
		return
	}

	// Set metadata from JWT claims and fetch license_id from user
	request.Metadata.UserID = claims.UserID
	request.Metadata.ShopID = claims.ShopID

	// Get license_id from user record
	user, err := sh.userRepo.GetByID(context.Background(), claims.UserID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to get user information", err.Error())
		return
	}

	if user.LicenseID == nil {
		response.Error(c, http.StatusBadRequest, "User has no license assigned", "")
		return
	}

	request.Metadata.LicenseID = *user.LicenseID

	log.Printf("Processing push sync for user %s, license %s, %d total entities",
		claims.UserID, request.Metadata.LicenseID, request.Data.GetEntityCount())

	// Process the sync request
	syncResponse := sh.syncService.ProcessPushSync(request)

	// Return appropriate status based on sync result
	if syncResponse.Success {
		response.Success(c, http.StatusOK, "Push sync completed successfully", syncResponse)
	} else {
		// Still return 200 but with error details in the response
		response.Success(c, http.StatusOK, "Push sync completed with errors", syncResponse)
	}
}

// PullSync handles GET /api/v1/sync/pull
// @Summary Pull data from server to mobile
// @Description Retrieve server changes since the given timestamp
// @Tags sync
// @Produce json
// @Param since query string false "Timestamp to get changes since (RFC3339 format)"
// @Param limit query int false "Limit number of records (max 1000)" default(1000)
// @Param offset query int false "Offset for pagination" default(0)
// @Success 200 {object} response.Response{data=sync.SyncResponse}
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/sync/pull [get]
func (sh *SyncHandler) PullSync(c *gin.Context) {
	// Get user info from JWT token
	userClaims, exists := c.Get("user")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "User not authenticated", "")
		return
	}

	claims, ok := userClaims.(*auth.Claims)
	if !ok {
		response.Error(c, http.StatusUnauthorized, "Invalid token claims", "")
		return
	}

	// Parse query parameters
	var pullRequest sync.PullSyncRequest
	pullRequest.UserID = claims.UserID
	pullRequest.ShopID = claims.ShopID

	// Get license_id from user record
	user, err := sh.userRepo.GetByID(context.Background(), claims.UserID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to get user information", err.Error())
		return
	}

	if user.LicenseID == nil {
		response.Error(c, http.StatusBadRequest, "User has no license assigned", "")
		return
	}

	pullRequest.LicenseID = *user.LicenseID

	// Parse since timestamp
	sinceStr := c.Query("since")
	if sinceStr != "" {
		since, err := time.Parse(time.RFC3339, sinceStr)
		if err != nil {
			response.Error(c, http.StatusBadRequest, "Invalid since timestamp format (use RFC3339)", err.Error())
			return
		}
		pullRequest.SinceTimestamp = &since
	}

	// Parse limit and offset
	if limitStr := c.Query("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil && limit > 0 && limit <= 1000 {
			pullRequest.Limit = limit
		}
	}
	if pullRequest.Limit == 0 {
		pullRequest.Limit = 1000 // Default limit
	}

	if offsetStr := c.Query("offset"); offsetStr != "" {
		if offset, err := strconv.Atoi(offsetStr); err == nil && offset >= 0 {
			pullRequest.Offset = offset
		}
	}

	log.Printf("Processing pull sync for user %s, license %s, since %v",
		claims.UserID, pullRequest.LicenseID, pullRequest.SinceTimestamp)

	// Process the pull request
	syncResponse := sh.syncService.ProcessPullSync(pullRequest)

	if syncResponse.Success {
		response.Success(c, http.StatusOK, "Pull sync completed successfully", syncResponse)
	} else {
		response.Success(c, http.StatusOK, "Pull sync completed with errors", syncResponse)
	}
}

// FullSync handles POST /api/v1/sync/full
// @Summary Perform full two-way synchronization
// @Description Processes push data first, then returns pull data
// @Tags sync
// @Accept json
// @Produce json
// @Param request body sync.FullSyncRequest true "Full sync request with push data and pull parameters"
// @Success 200 {object} response.Response{data=sync.FullSyncResponse}
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/sync/full [post]
func (sh *SyncHandler) FullSync(c *gin.Context) {
	var request sync.FullSyncRequest

	// Parse request body
	if err := c.ShouldBindJSON(&request); err != nil {
		log.Printf("Failed to parse full sync request: %v", err)
		response.Error(c, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	// Get user info from JWT token
	userClaims, exists := c.Get("user")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "User not authenticated", "")
		return
	}

	claims, ok := userClaims.(*auth.Claims)
	if !ok {
		response.Error(c, http.StatusUnauthorized, "Invalid token claims", "")
		return
	}

	// Set metadata from JWT claims and fetch license_id from user
	request.PushData.Metadata.UserID = claims.UserID
	request.PushData.Metadata.ShopID = claims.ShopID

	// Get license_id from user record
	user, err := sh.userRepo.GetByID(context.Background(), claims.UserID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to get user information", err.Error())
		return
	}

	if user.LicenseID == nil {
		response.Error(c, http.StatusBadRequest, "User has no license assigned", "")
		return
	}

	request.PushData.Metadata.LicenseID = *user.LicenseID

	log.Printf("Processing full sync for user %s, license %s, %d push entities",
		claims.UserID, request.PushData.Metadata.LicenseID, request.PushData.Data.GetEntityCount())

	// Process the full sync request
	syncResponse := sh.syncService.ProcessFullSync(request)

	// Determine overall success
	overallSuccess := syncResponse.PushResult.Success && syncResponse.PullResult.Success
	message := "Full sync completed successfully"
	if !overallSuccess {
		message = "Full sync completed with some errors"
	}

	response.Success(c, http.StatusOK, message, syncResponse)
}

// GetSyncStatus handles GET /api/v1/sync/status
// @Summary Get synchronization status and metrics
// @Description Returns information about the last sync operations and current sync state
// @Tags sync
// @Produce json
// @Success 200 {object} response.Response{data=map[string]interface{}}
// @Failure 401 {object} response.Response
// @Router /api/v1/sync/status [get]
func (sh *SyncHandler) GetSyncStatus(c *gin.Context) {
	// Get user info from JWT token
	userClaims, exists := c.Get("user")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "User not authenticated", "")
		return
	}

	claims, ok := userClaims.(*auth.Claims)
	if !ok {
		response.Error(c, http.StatusUnauthorized, "Invalid token claims", "")
		return
	}

	// For now, return basic status info
	// In the future, this could include metrics from a cache or database
	status := map[string]interface{}{
		"user_id":           claims.UserID,
		"shop_id":           claims.ShopID,
		"server_timestamp":  time.Now().UTC(),
		"sync_available":    true,
		"last_sync":         nil, // TODO: Implement last sync tracking
		"pending_changes":   0,   // TODO: Implement change tracking
	}

	response.Success(c, http.StatusOK, "Sync status retrieved", status)
}

// ValidateSyncData handles POST /api/v1/sync/validate
// @Summary Validate sync data without processing
// @Description Validates sync data format and business rules without actually syncing
// @Tags sync
// @Accept json
// @Produce json
// @Param request body sync.SyncRequest true "Sync data to validate"
// @Success 200 {object} response.Response{data=[]sync.SyncError}
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Router /api/v1/sync/validate [post]
func (sh *SyncHandler) ValidateSyncData(c *gin.Context) {
	var request sync.SyncRequest

	// Parse request body
	if err := c.ShouldBindJSON(&request); err != nil {
		log.Printf("Failed to parse validation request: %v", err)
		response.Error(c, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	// Get user info from JWT token
	userClaims, exists := c.Get("user")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "User not authenticated", "")
		return
	}

	claims, ok := userClaims.(*auth.Claims)
	if !ok {
		response.Error(c, http.StatusUnauthorized, "Invalid token claims", "")
		return
	}

	// Set metadata from JWT claims and validate
	request.Metadata.UserID = claims.UserID
	request.Metadata.ShopID = claims.ShopID

	// Get license_id from user record for validation
	user, err := sh.userRepo.GetByID(context.Background(), claims.UserID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to get user information", err.Error())
		return
	}

	if user.LicenseID == nil {
		response.Error(c, http.StatusBadRequest, "User has no license assigned", "")
		return
	}

	request.Metadata.LicenseID = *user.LicenseID

	// Validate the request
	validationErrors := request.Validate()

	// TODO: Add more comprehensive validation
	// - Business rule validation
	// - Reference integrity checks
	// - Shop access validation

	if len(validationErrors) == 0 {
		response.Success(c, http.StatusOK, "Sync data is valid", validationErrors)
	} else {
		response.Success(c, http.StatusOK, "Sync data has validation errors", validationErrors)
	}
}