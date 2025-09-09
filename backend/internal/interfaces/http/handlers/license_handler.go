package handlers

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/terminator791/t-pos/internal/application/services"
	"github.com/terminator791/t-pos/pkg/response"
)

// LicenseHandler handles license-related HTTP requests
type LicenseHandler struct {
	licenseService *services.LicenseService
}

// NewLicenseHandler creates a new license handler
func NewLicenseHandler(licenseService *services.LicenseService) *LicenseHandler {
	return &LicenseHandler{
		licenseService: licenseService,
	}
}

// GetLicense handles GET /api/v1/licenses/:id
func (h *LicenseHandler) GetLicense(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.ErrorBadRequest(c, "Invalid license ID", err.Error())
		return
	}

	license, err := h.licenseService.GetLicense(c.Request.Context(), id)
	if err != nil {
		response.ErrorNotFound(c, "License not found", err.Error())
		return
	}

	response.SuccessOK(c, "License retrieved successfully", license)
}

// GetAllLicenses handles GET /api/v1/licenses
func (h *LicenseHandler) GetAllLicenses(c *gin.Context) {
	// Parse query parameters
	limitStr := c.DefaultQuery("limit", "10")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		response.ErrorBadRequest(c, "Invalid limit parameter", err.Error())
		return
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil {
		response.ErrorBadRequest(c, "Invalid offset parameter", err.Error())
		return
	}

	licenses, err := h.licenseService.GetAllLicenses(c.Request.Context(), limit, offset)
	if err != nil {
		response.ErrorInternalServer(c, "Failed to retrieve licenses", err.Error())
		return
	}

	data := gin.H{
		"licenses": licenses,
		"count":    len(licenses),
		"limit":    limit,
		"offset":   offset,
	}
	response.SuccessOK(c, "Licenses retrieved successfully", data)
}

// CreateLicense handles POST /api/v1/licenses
func (h *LicenseHandler) CreateLicense(c *gin.Context) {
	var req services.CreateLicenseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorBadRequest(c, "Invalid request body", err.Error())
		return
	}

	// Get user ID from context (set by auth middleware)
	userIDInterface, exists := c.Get("user_id")
	if !exists {
		response.ErrorUnauthorized(c, "User not authenticated", "User ID not found in context")
		return
	}

	userID, ok := userIDInterface.(uuid.UUID)
	if !ok {
		response.ErrorInternalServer(c, "Invalid user ID format", "User ID in context is not a valid UUID")
		return
	}

	license, err := h.licenseService.CreateLicense(c.Request.Context(), req, userID)
	if err != nil {
		response.ErrorBadRequest(c, "Failed to create license", err.Error())
		return
	}

	response.SuccessCreated(c, "License created successfully", license)
}

// DeleteLicense handles DELETE /api/v1/licenses/:id
func (h *LicenseHandler) DeleteLicense(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.ErrorBadRequest(c, "Invalid license ID", err.Error())
		return
	}

	err = h.licenseService.DeleteLicense(c.Request.Context(), id)
	if err != nil {
		response.ErrorBadRequest(c, "Failed to delete license", err.Error())
		return
	}

	response.SuccessOK(c, "License deleted successfully", nil)
}