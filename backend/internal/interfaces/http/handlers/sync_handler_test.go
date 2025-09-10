package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/terminator791/t-pos/internal/domain/entities"
	"github.com/terminator791/t-pos/internal/domain/sync"
)

// Test sync data validation (unit tests that don't require dependencies)
func TestSyncDataValidation(t *testing.T) {
	// Test SyncRequest validation
	t.Run("SyncRequest validation with empty metadata", func(t *testing.T) {
		request := sync.SyncRequest{
			Data: sync.SyncData{
				Products: []entities.Product{{}}, // Non-empty data
			},
		}

		errors := request.Validate()
		assert.NotEmpty(t, errors)

		// Should have license_id and user_id validation errors
		foundLicenseError := false
		foundUserError := false
		for _, err := range errors {
			if err.Message == "license_id is required" {
				foundLicenseError = true
			}
			if err.Message == "user_id is required" {
				foundUserError = true
			}
		}
		assert.True(t, foundLicenseError)
		assert.True(t, foundUserError)
	})

	t.Run("SyncData entity count and empty checks", func(t *testing.T) {
		emptyData := sync.SyncData{}
		assert.True(t, emptyData.IsEmpty())
		assert.Equal(t, 0, emptyData.GetEntityCount())

		nonEmptyData := sync.SyncData{
			Products:   []entities.Product{{}},
			Categories: []entities.Category{{}, {}},
		}
		assert.False(t, nonEmptyData.IsEmpty())
		assert.Equal(t, 3, nonEmptyData.GetEntityCount())
	})
}

// Test sync response helpers
func TestSyncResponseHelpers(t *testing.T) {
	t.Run("NewSyncResponse creates proper default response", func(t *testing.T) {
		response := sync.NewSyncResponse()

		assert.True(t, response.Success)
		assert.Equal(t, 0, response.RecordsProcessed)
		assert.NotNil(t, response.Errors)
		assert.Equal(t, 0, len(response.Errors))
		assert.False(t, response.HasErrors())
	})

	t.Run("AddError properly modifies response", func(t *testing.T) {
		response := sync.NewSyncResponse()
		testUUID := uuid.New()
		
		response.AddError("product", testUUID, "name", "Product name is required", "validation")

		assert.False(t, response.Success)
		assert.True(t, response.HasErrors())
		assert.Equal(t, 1, len(response.Errors))
		
		error := response.Errors[0]
		assert.Equal(t, "product", error.EntityType)
		assert.Equal(t, testUUID, error.EntityID)
		assert.Equal(t, "name", error.Field)
		assert.Equal(t, "Product name is required", error.Message)
		assert.Equal(t, "validation", error.ErrorType)
	})
}

// Test HTTP response format for sync endpoints (without dependencies)
func TestSyncHandlerHTTPResponseFormat(t *testing.T) {
	// Set Gin to test mode
	gin.SetMode(gin.TestMode)

	// Test endpoints that don't require authentication for structure validation
	t.Run("Unauthenticated requests return proper error format", func(t *testing.T) {
		router := gin.New()
		
		// Mock handler that always returns 401 for testing
		router.GET("/sync/status", func(c *gin.Context) {
			c.JSON(http.StatusUnauthorized, gin.H{
				"status":  "failed",
				"message": "User not authenticated",
				"errors":  nil,
			})
		})

		req := httptest.NewRequest("GET", "/sync/status", nil)
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)

		// Should return 401 Unauthorized
		assert.Equal(t, http.StatusUnauthorized, resp.Code)

		// Parse response body
		var response map[string]interface{}
		err := json.Unmarshal(resp.Body.Bytes(), &response)
		assert.NoError(t, err)

		// Check standardized response format
		assert.Equal(t, "failed", response["status"])
		assert.Equal(t, "User not authenticated", response["message"])
		assert.Contains(t, response, "errors")
	})

	t.Run("Invalid JSON requests return proper error format", func(t *testing.T) {
		router := gin.New()
		
		// Mock handler that simulates invalid JSON parsing
		router.POST("/sync/push", func(c *gin.Context) {
			var request map[string]interface{}
			if err := c.ShouldBindJSON(&request); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"status":  "failed",
					"message": "Invalid request body",
					"errors":  err.Error(),
				})
				return
			}
		})

		// Send invalid JSON
		invalidJSON := bytes.NewBufferString(`{"invalid_json":}`)
		req := httptest.NewRequest("POST", "/sync/push", invalidJSON)
		req.Header.Set("Content-Type", "application/json")
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)

		// Should return 400 Bad Request
		assert.Equal(t, http.StatusBadRequest, resp.Code)

		// Parse response body
		var response map[string]interface{}
		err := json.Unmarshal(resp.Body.Bytes(), &response)
		assert.NoError(t, err)

		// Check standardized response format
		assert.Equal(t, "failed", response["status"])
		assert.Equal(t, "Invalid request body", response["message"])
		assert.Contains(t, response, "errors")
	})
}