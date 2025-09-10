package handlers

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/terminator791/t-pos/internal/domain/repositories"
	"github.com/terminator791/t-pos/pkg/response"
)

// RoleHandler handles role-related HTTP requests
type RoleHandler struct {
	roleRepo repositories.RoleRepository
}

// NewRoleHandler creates a new role handler
func NewRoleHandler(roleRepo repositories.RoleRepository) *RoleHandler {
	return &RoleHandler{
		roleRepo: roleRepo,
	}
}

// GetAllRoles handles GET /api/v1/roles
func (h *RoleHandler) GetAllRoles(c *gin.Context) {
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

	roles, err := h.roleRepo.List(c.Request.Context(), limit, offset)
	if err != nil {
		response.ErrorInternalServer(c, "Failed to retrieve roles", err.Error())
		return
	}

	data := gin.H{
		"roles":  roles,
		"count":  len(roles),
		"limit":  limit,
		"offset": offset,
	}
	response.SuccessOK(c, "Roles retrieved successfully", data)
}

// GetRole handles GET /api/v1/roles/:id
func (h *RoleHandler) GetRole(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.ErrorBadRequest(c, "Invalid role ID", err.Error())
		return
	}

	role, err := h.roleRepo.GetByID(c.Request.Context(), id)
	if err != nil {
		response.ErrorNotFound(c, "Role not found", err.Error())
		return
	}

	response.SuccessOK(c, "Role retrieved successfully", role)
}

// GetRoleByName handles GET /api/v1/roles/name/:name
func (h *RoleHandler) GetRoleByName(c *gin.Context) {
	name := c.Param("name")
	if name == "" {
		response.ErrorBadRequest(c, "Role name is required", nil)
		return
	}

	role, err := h.roleRepo.GetByName(c.Request.Context(), name)
	if err != nil {
		response.ErrorNotFound(c, "Role not found", err.Error())
		return
	}

	response.SuccessOK(c, "Role retrieved successfully", role)
}
