package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/terminator791/t-pos/internal/application/services"
)

// Test the response format of license handlers
func TestLicenseHandlerResponseFormat(t *testing.T) {
	// Set Gin to test mode
	gin.SetMode(gin.TestMode)

	// Create a mock service (we're mainly testing response format)
	mockService := &services.LicenseService{}
	handler := NewLicenseHandler(mockService)

	// Test GetLicense with invalid UUID
	t.Run("GetLicense with invalid UUID returns proper error format", func(t *testing.T) {
		router := gin.New()
		router.GET("/licenses/:id", handler.GetLicense)

		req := httptest.NewRequest("GET", "/licenses/invalid-uuid", nil)
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
		assert.Equal(t, "Invalid license ID", response["message"])
		assert.Contains(t, response, "errors")
	})

	// Test CreateLicense with invalid JSON
	t.Run("CreateLicense with invalid JSON returns proper error format", func(t *testing.T) {
		router := gin.New()
		router.POST("/licenses", handler.CreateLicense)

		// Send invalid JSON
		invalidJSON := bytes.NewBufferString(`{"invalid_json":}`)
		req := httptest.NewRequest("POST", "/licenses", invalidJSON)
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

	// Test CreateLicense with missing required fields
	t.Run("CreateLicense with missing fields returns proper error format", func(t *testing.T) {
		router := gin.New()
		router.POST("/licenses", handler.CreateLicense)

		// Send empty JSON object
		emptyJSON := bytes.NewBufferString(`{}`)
		req := httptest.NewRequest("POST", "/licenses", emptyJSON)
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

	// Test pagination parameter validation
	t.Run("GetAllLicenses with invalid limit returns proper error format", func(t *testing.T) {
		router := gin.New()
		router.GET("/licenses", handler.GetAllLicenses)

		req := httptest.NewRequest("GET", "/licenses?limit=invalid", nil)
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
		assert.Equal(t, "Invalid limit parameter", response["message"])
		assert.Contains(t, response, "errors")
	})
}

// Test the response format of customer handlers
func TestCustomerHandlerResponseFormat(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockService := &services.CustomerService{}
	handler := NewCustomerHandler(mockService)

	// Test GetCustomer with invalid UUID
	t.Run("GetCustomer with invalid UUID returns proper error format", func(t *testing.T) {
		router := gin.New()
		router.GET("/customers/:id", handler.GetCustomer)

		req := httptest.NewRequest("GET", "/customers/invalid-uuid", nil)
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)

		assert.Equal(t, http.StatusBadRequest, resp.Code)

		var response map[string]interface{}
		err := json.Unmarshal(resp.Body.Bytes(), &response)
		assert.NoError(t, err)

		assert.Equal(t, "failed", response["status"])
		assert.Equal(t, "Invalid customer ID", response["message"])
		assert.Contains(t, response, "errors")
	})
}

// Test the response format of user management handlers
func TestUserManagementHandlerResponseFormat(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockService := &services.UserManagementService{}
	handler := NewUserManagementHandler(mockService)

	// Test GetUser with invalid UUID
	t.Run("GetUser with invalid UUID returns proper error format", func(t *testing.T) {
		router := gin.New()
		router.GET("/users/:id", handler.GetUser)

		req := httptest.NewRequest("GET", "/users/invalid-uuid", nil)
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)

		assert.Equal(t, http.StatusBadRequest, resp.Code)

		var response map[string]interface{}
		err := json.Unmarshal(resp.Body.Bytes(), &response)
		assert.NoError(t, err)

		assert.Equal(t, "failed", response["status"])
		assert.Equal(t, "Invalid user ID", response["message"])
		assert.Contains(t, response, "errors")
	})

	// Test UpdateUserPassword with invalid UUID
	t.Run("UpdateUserPassword with invalid UUID returns proper error format", func(t *testing.T) {
		router := gin.New()
		router.PUT("/users/:id", handler.UpdateUserPassword)

		validJSON := bytes.NewBufferString(`{"password":"newpass"}`)
		req := httptest.NewRequest("PUT", "/users/invalid-uuid", validJSON)
		req.Header.Set("Content-Type", "application/json")
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)

		assert.Equal(t, http.StatusBadRequest, resp.Code)

		var response map[string]interface{}
		err := json.Unmarshal(resp.Body.Bytes(), &response)
		assert.NoError(t, err)

		assert.Equal(t, "failed", response["status"])
		assert.Equal(t, "Invalid user ID", response["message"])
		assert.Contains(t, response, "errors")
	})
}