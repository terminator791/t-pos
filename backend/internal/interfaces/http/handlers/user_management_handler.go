package handlers

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/terminator791/t-pos/internal/application/services"
	"github.com/terminator791/t-pos/pkg/response"
)

// UserManagementHandler handles user management-related HTTP requests
type UserManagementHandler struct {
	userService *services.UserManagementService
}

// NewUserManagementHandler creates a new user management handler
func NewUserManagementHandler(userService *services.UserManagementService) *UserManagementHandler {
	return &UserManagementHandler{
		userService: userService,
	}
}

// GetUser handles GET /api/v1/users/:id
func (h *UserManagementHandler) GetUser(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.ErrorBadRequest(c, "Invalid user ID", err.Error())
		return
	}

	user, err := h.userService.GetUser(c.Request.Context(), id)
	if err != nil {
		response.ErrorNotFound(c, "User not found", err.Error())
		return
	}

	response.SuccessOK(c, "User retrieved successfully", user)
}

// GetAllUsers handles GET /api/v1/users
func (h *UserManagementHandler) GetAllUsers(c *gin.Context) {
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

	users, err := h.userService.GetAllUsers(c.Request.Context(), limit, offset)
	if err != nil {
		response.ErrorInternalServer(c, "Failed to retrieve users", err.Error())
		return
	}

	data := gin.H{
		"users":  users,
		"count":  len(users),
		"limit":  limit,
		"offset": offset,
	}
	response.SuccessOK(c, "Users retrieved successfully", data)
}

// CreateUser handles POST /api/v1/users
func (h *UserManagementHandler) CreateUser(c *gin.Context) {
	var req services.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorBadRequest(c, "Invalid request body", err.Error())
		return
	}

	user, err := h.userService.CreateUser(c.Request.Context(), req)
	if err != nil {
		response.ErrorBadRequest(c, "Failed to create user", err.Error())
		return
	}

	response.SuccessCreated(c, "User created successfully", user)
}

// UpdateUserPassword handles PUT /api/v1/users/:id
func (h *UserManagementHandler) UpdateUserPassword(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.ErrorBadRequest(c, "Invalid user ID", err.Error())
		return
	}

	var req services.UpdateUserPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorBadRequest(c, "Invalid request body", err.Error())
		return
	}

	err = h.userService.UpdateUserPassword(c.Request.Context(), id, req)
	if err != nil {
		response.ErrorBadRequest(c, "Failed to update user password", err.Error())
		return
	}

	response.SuccessOK(c, "User password updated successfully", nil)
}

// DeleteUser handles DELETE /api/v1/users/:id
func (h *UserManagementHandler) DeleteUser(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.ErrorBadRequest(c, "Invalid user ID", err.Error())
		return
	}

	err = h.userService.DeleteUser(c.Request.Context(), id)
	if err != nil {
		response.ErrorBadRequest(c, "Failed to delete user", err.Error())
		return
	}

	response.SuccessOK(c, "User deleted successfully", nil)
}
