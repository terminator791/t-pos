package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/terminator791/t-pos/internal/application/services"
	"github.com/terminator791/t-pos/internal/infrastructure/repositories"
	"gorm.io/gorm"
)

func TestSyncHandler_Health(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	
	// Create sync handler with nil dependencies for this simple test
	syncHandler := &SyncHandler{}
	
	// Create test router
	router := gin.New()
	router.GET("/sync/health", syncHandler.Health)
	
	// Create test request
	req, _ := http.NewRequest("GET", "/sync/health", nil)
	w := httptest.NewRecorder()
	
	// Execute request
	router.ServeHTTP(w, req)
	
	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "healthy")
	assert.Contains(t, w.Body.String(), "sync")
}

func TestSyncHandler_ProcessSync_MissingAuth(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	
	// Create sync handler with nil dependencies
	syncHandler := &SyncHandler{}
	
	// Create test router
	router := gin.New()
	router.POST("/sync", syncHandler.ProcessSync)
	
	// Create test request without authentication
	req, _ := http.NewRequest("POST", "/sync", nil)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	
	// Execute request
	router.ServeHTTP(w, req)
	
	// Assertions - should fail due to missing authentication
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "not authenticated")
}

func TestSyncHandler_GetSyncInfo_MissingAuth(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	
	// Create sync handler with nil dependencies
	syncHandler := &SyncHandler{}
	
	// Create test router
	router := gin.New()
	router.GET("/sync/info", syncHandler.GetSyncInfo)
	
	// Create test request without authentication
	req, _ := http.NewRequest("GET", "/sync/info", nil)
	w := httptest.NewRecorder()
	
	// Execute request
	router.ServeHTTP(w, req)
	
	// Assertions - should fail due to missing authentication
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "not authenticated")
}

func TestNewSyncHandler(t *testing.T) {
	// Test that we can create a sync handler with proper dependencies
	// This tests the constructor without requiring actual database connections
	
	// Use nil for database-dependent components
	var db *gorm.DB = nil
	var syncService *services.SyncService = nil
	var userRepo = repositories.NewUserRepository(db)
	
	syncHandler := NewSyncHandler(syncService, userRepo)
	
	assert.NotNil(t, syncHandler)
	assert.Equal(t, syncService, syncHandler.syncService)
	assert.Equal(t, userRepo, syncHandler.userRepo)
}