package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/terminator791/t-pos/internal/application/services"
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
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid license ID",
			"message": err.Error(),
		})
		return
	}

	license, err := h.licenseService.GetLicense(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "License not found",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":    license,
		"message": "License retrieved successfully",
	})
}

// GetAllLicenses handles GET /api/v1/licenses
func (h *LicenseHandler) GetAllLicenses(c *gin.Context) {
	// Parse query parameters
	limitStr := c.DefaultQuery("limit", "10")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid limit parameter",
			"message": err.Error(),
		})
		return
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid offset parameter",
			"message": err.Error(),
		})
		return
	}

	licenses, err := h.licenseService.GetAllLicenses(c.Request.Context(), limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to retrieve licenses",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":    licenses,
		"count":   len(licenses),
		"limit":   limit,
		"offset":  offset,
		"message": "Licenses retrieved successfully",
	})
}

// CreateLicense handles POST /api/v1/licenses
func (h *LicenseHandler) CreateLicense(c *gin.Context) {
	var req services.CreateLicenseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"message": err.Error(),
		})
		return
	}

	// Get user ID from context (set by auth middleware)
	userIDInterface, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "User not authenticated",
			"message": "User ID not found in context",
		})
		return
	}

	userID, ok := userIDInterface.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Invalid user ID format",
			"message": "User ID in context is not a valid UUID",
		})
		return
	}

	license, err := h.licenseService.CreateLicense(c.Request.Context(), req, userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Failed to create license",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"data":    license,
		"message": "License created successfully",
	})
}

// DeleteLicense handles DELETE /api/v1/licenses/:id
func (h *LicenseHandler) DeleteLicense(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid license ID",
			"message": err.Error(),
		})
		return
	}

	err = h.licenseService.DeleteLicense(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Failed to delete license",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "License deleted successfully",
	})
}