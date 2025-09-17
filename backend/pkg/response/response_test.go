package response

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestResponseHelpers(t *testing.T) {
	// Create a test router
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Test success response
	router.GET("/success", func(c *gin.Context) {
		SuccessOK(c, "Test success", map[string]string{"key": "value"})
	})

	// Test error response
	router.GET("/error", func(c *gin.Context) {
		ErrorBadRequest(c, "Test error", "validation failed")
	})

	// Test success endpoint
	req := httptest.NewRequest("GET", "/success", nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusOK, resp.Code)
	assert.Contains(t, resp.Body.String(), `"status":"success"`)
	assert.Contains(t, resp.Body.String(), `"message":"Test success"`)
	assert.Contains(t, resp.Body.String(), `"data":{"key":"value"}`)

	// Test error endpoint
	req = httptest.NewRequest("GET", "/error", nil)
	resp = httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusBadRequest, resp.Code)
	assert.Contains(t, resp.Body.String(), `"status":"failed"`)
	assert.Contains(t, resp.Body.String(), `"message":"Test error"`)
	assert.Contains(t, resp.Body.String(), `"errors":"validation failed"`)
}
