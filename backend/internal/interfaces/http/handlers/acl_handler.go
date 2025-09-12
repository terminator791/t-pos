package handlers

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/terminator791/t-pos/internal/domain/repositories"
	"github.com/terminator791/t-pos/internal/infrastructure/casbin"
	"github.com/terminator791/t-pos/pkg/response"
)

// ACLHandler handles ACL management operations
type ACLHandler struct {
	enforcerService *casbin.EnforcerService
	roleRepo        repositories.RoleRepository
	policyRepo      repositories.PolicyRepository
}

// NewACLHandler creates a new ACL handler
func NewACLHandler(
	enforcerService *casbin.EnforcerService,
	roleRepo repositories.RoleRepository,
	policyRepo repositories.PolicyRepository,
) *ACLHandler {
	return &ACLHandler{
		enforcerService: enforcerService,
		roleRepo:        roleRepo,
		policyRepo:      policyRepo,
	}
}

// GetAllPolicies returns all Casbin policies
func (h *ACLHandler) GetAllPolicies(c *gin.Context) {
	policies := h.enforcerService.GetAllPolicies()

	response.SuccessOK(c, "Policies retrieved successfully", gin.H{
		"policies": policies,
		"count":    len(policies),
	})
}

// GetAllRoles returns all Casbin role assignments
func (h *ACLHandler) GetAllRoles(c *gin.Context) {
	roles := h.enforcerService.GetAllRoles()

	response.SuccessOK(c, "Role assignments retrieved successfully", gin.H{
		"role_assignments": roles,
		"count":            len(roles),
	})
}

// GetUserRoles returns roles for a specific user
func (h *ACLHandler) GetUserRoles(c *gin.Context) {
	userID := c.Param("userId")
	domain := c.Query("domain")

	if domain == "" {
		domain = "*"
	}

	roles := h.enforcerService.GetRolesForUser(userID, domain)

	response.SuccessOK(c, "User roles retrieved successfully", gin.H{
		"user_id": userID,
		"domain":  domain,
		"roles":   roles,
	})
}

// GetRoleUsers returns users for a specific role
func (h *ACLHandler) GetRoleUsers(c *gin.Context) {
	role := c.Param("role")
	domain := c.Query("domain")

	if domain == "" {
		domain = "*"
	}

	users := h.enforcerService.GetUsersForRole(role, domain)

	response.SuccessOK(c, "Role users retrieved successfully", gin.H{
		"role":   role,
		"domain": domain,
		"users":  users,
	})
}

// AddRoleForUser adds a role to a user
func (h *ACLHandler) AddRoleForUser(c *gin.Context) {
	var req struct {
		UserID string `json:"user_id" binding:"required"`
		Role   string `json:"role" binding:"required"`
		Domain string `json:"domain"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorBadRequest(c, "Invalid request data", err.Error())
		return
	}

	if req.Domain == "" {
		req.Domain = "*"
	}

	success, err := h.enforcerService.AddRoleForUser(req.UserID, req.Role, req.Domain)
	if err != nil {
		response.ErrorInternalServer(c, "Failed to add role", err.Error())
		return
	}

	if !success {
		response.ErrorBadRequest(c, "Role assignment already exists", nil)
		return
	}

	response.SuccessCreated(c, "Role added successfully", gin.H{
		"user_id": req.UserID,
		"role":    req.Role,
		"domain":  req.Domain,
	})
}

// RemoveRoleForUser removes a role from a user
func (h *ACLHandler) RemoveRoleForUser(c *gin.Context) {
	var req struct {
		UserID string `json:"user_id" binding:"required"`
		Role   string `json:"role" binding:"required"`
		Domain string `json:"domain"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorBadRequest(c, "Invalid request data", err.Error())
		return
	}

	if req.Domain == "" {
		req.Domain = "*"
	}

	success, err := h.enforcerService.RemoveRoleForUser(req.UserID, req.Role, req.Domain)
	if err != nil {
		response.ErrorInternalServer(c, "Failed to remove role", err.Error())
		return
	}

	if !success {
		response.ErrorNotFound(c, "Role assignment not found", nil)
		return
	}

	response.SuccessOK(c, "Role removed successfully", gin.H{
		"user_id": req.UserID,
		"role":    req.Role,
		"domain":  req.Domain,
	})
}

// AddPolicy adds a new policy
func (h *ACLHandler) AddPolicy(c *gin.Context) {
	var req struct {
		Role   string `json:"role" binding:"required"`
		Domain string `json:"domain" binding:"required"`
		Object string `json:"object" binding:"required"`
		Action string `json:"action" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorBadRequest(c, "Invalid request data", err.Error())
		return
	}

	success, err := h.enforcerService.AddPolicy(req.Role, req.Domain, req.Object, req.Action)
	if err != nil {
		response.ErrorInternalServer(c, "Failed to add policy", err.Error())
		return
	}

	if !success {
		response.ErrorBadRequest(c, "Policy already exists", nil)
		return
	}

	response.SuccessCreated(c, "Policy added successfully", gin.H{
		"role":   req.Role,
		"domain": req.Domain,
		"object": req.Object,
		"action": req.Action,
	})
}

// RemovePolicy removes a policy
func (h *ACLHandler) RemovePolicy(c *gin.Context) {
	var req struct {
		Role   string `json:"role" binding:"required"`
		Domain string `json:"domain" binding:"required"`
		Object string `json:"object" binding:"required"`
		Action string `json:"action" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorBadRequest(c, "Invalid request data", err.Error())
		return
	}

	success, err := h.enforcerService.RemovePolicy(req.Role, req.Domain, req.Object, req.Action)
	if err != nil {
		response.ErrorInternalServer(c, "Failed to remove policy", err.Error())
		return
	}

	if !success {
		response.ErrorNotFound(c, "Policy not found", nil)
		return
	}

	response.SuccessOK(c, "Policy removed successfully", gin.H{
		"role":   req.Role,
		"domain": req.Domain,
		"object": req.Object,
		"action": req.Action,
	})
}

// CheckPermission checks if a user has permission
func (h *ACLHandler) CheckPermission(c *gin.Context) {
	var req struct {
		UserID string `json:"user_id" binding:"required"`
		Domain string `json:"domain" binding:"required"`
		Object string `json:"object" binding:"required"`
		Action string `json:"action" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorBadRequest(c, "Invalid request data", err.Error())
		return
	}

	allowed, err := h.enforcerService.Enforce(req.UserID, req.Domain, req.Object, req.Action)
	if err != nil {
		response.ErrorInternalServer(c, "Failed to check permission", err.Error())
		return
	}

	response.SuccessOK(c, "Permission check completed", gin.H{
		"user_id": req.UserID,
		"domain":  req.Domain,
		"object":  req.Object,
		"action":  req.Action,
		"allowed": allowed,
	})
}

// ReloadPolicies reloads policies from database
func (h *ACLHandler) ReloadPolicies(c *gin.Context) {
	if err := h.enforcerService.LoadPolicy(); err != nil {
		response.ErrorInternalServer(c, "Failed to reload policies", err.Error())
		return
	}

	response.SuccessOK(c, "Policies reloaded successfully", nil)
}

// GetSystemRoles returns all system roles from database
func (h *ACLHandler) GetSystemRoles(c *gin.Context) {
	roles, err := h.roleRepo.GetAll(context.Background())
	if err != nil {
		response.ErrorInternalServer(c, "Failed to get system roles", err.Error())
		return
	}

	response.SuccessOK(c, "System roles retrieved successfully", gin.H{
		"roles": roles,
		"count": len(roles),
	})
}

// GetSystemPolicies returns all system policies from database
func (h *ACLHandler) GetSystemPolicies(c *gin.Context) {
	policies, err := h.policyRepo.GetAll(context.Background())
	if err != nil {
		response.ErrorInternalServer(c, "Failed to get system policies", err.Error())
		return
	}

	response.SuccessOK(c, "System policies retrieved successfully", gin.H{
		"policies": policies,
		"count":    len(policies),
	})
}
