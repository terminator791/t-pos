package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// SuccessResponse represents a successful API response
type SuccessResponse struct {
	Status  string      `json:"status"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

// ErrorResponse represents an error API response
type ErrorResponse struct {
	Status  string      `json:"status"`
	Message string      `json:"message"`
	Errors  interface{} `json:"errors"`
}

// Success sends a successful JSON response
func Success(c *gin.Context, statusCode int, message string, data interface{}) {
	c.JSON(statusCode, SuccessResponse{
		Status:  "success",
		Message: message,
		Data:    data,
	})
}

// Error sends an error JSON response
func Error(c *gin.Context, statusCode int, message string, errors interface{}) {
	c.JSON(statusCode, ErrorResponse{
		Status:  "failed",
		Message: message,
		Errors:  errors,
	})
}

// SuccessOK sends a 200 OK successful response
func SuccessOK(c *gin.Context, message string, data interface{}) {
	Success(c, http.StatusOK, message, data)
}

// SuccessCreated sends a 201 Created successful response
func SuccessCreated(c *gin.Context, message string, data interface{}) {
	Success(c, http.StatusCreated, message, data)
}

// ErrorBadRequest sends a 400 Bad Request error response
func ErrorBadRequest(c *gin.Context, message string, errors interface{}) {
	Error(c, http.StatusBadRequest, message, errors)
}

// ErrorUnauthorized sends a 401 Unauthorized error response
func ErrorUnauthorized(c *gin.Context, message string, errors interface{}) {
	Error(c, http.StatusUnauthorized, message, errors)
}

// ErrorForbidden sends a 403 Forbidden error response
func ErrorForbidden(c *gin.Context, message string, errors interface{}) {
	Error(c, http.StatusForbidden, message, errors)
}

// ErrorNotFound sends a 404 Not Found error response
func ErrorNotFound(c *gin.Context, message string, errors interface{}) {
	Error(c, http.StatusNotFound, message, errors)
}

// ErrorInternalServer sends a 500 Internal Server Error response
func ErrorInternalServer(c *gin.Context, message string, errors interface{}) {
	Error(c, http.StatusInternalServerError, message, errors)
}