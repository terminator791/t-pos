package handlers

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/terminator791/t-pos/internal/domain/entities"
	"github.com/terminator791/t-pos/internal/domain/repositories"
	"github.com/terminator791/t-pos/internal/infrastructure/auth"
	"github.com/terminator791/t-pos/internal/infrastructure/casbin"
	"github.com/terminator791/t-pos/pkg/response"
	"gorm.io/gorm"
)

// AuthHandler handles authentication related requests
type AuthHandler struct {
	userRepo        repositories.UserRepository
	userRoleRepo    repositories.UserRoleRepository
	roleRepo        repositories.RoleRepository
	jwtService      *auth.JWTService
	passwordService *auth.PasswordService
	enforcerService *casbin.EnforcerService
}

// LoginRequest represents login request payload
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
	Domain   string `json:"domain,omitempty"` // Optional tenant/shop domain
}

// RegisterRequest represents registration request payload
type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Username string `json:"username" binding:"required,min=3"`
	Name     string `json:"name" binding:"required,min=2"`
	Password string `json:"password" binding:"required,min=6"`
	Pin      *int   `json:"pin,omitempty"`
	Domain   string `json:"domain,omitempty"` // Optional tenant/shop domain
}

// LoginResponse represents login response
type LoginResponse struct {
	Token     string                 `json:"token"`
	User      *entities.User         `json:"user"`
	Roles     []string               `json:"roles"`
	Domain    string                 `json:"domain"`
	ExpiresAt int64                  `json:"expires_at"`
}

// NewAuthHandler creates a new authentication handler
func NewAuthHandler(
	userRepo repositories.UserRepository,
	userRoleRepo repositories.UserRoleRepository,
	roleRepo repositories.RoleRepository,
	jwtService *auth.JWTService,
	passwordService *auth.PasswordService,
	enforcerService *casbin.EnforcerService,
) *AuthHandler {
	return &AuthHandler{
		userRepo:        userRepo,
		userRoleRepo:    userRoleRepo,
		roleRepo:        roleRepo,
		jwtService:      jwtService,
		passwordService: passwordService,
		enforcerService: enforcerService,
	}
}

// Login handles user login
func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorBadRequest(c, "Invalid request data", err.Error())
		return
	}

	// Find user by email
	user, err := h.userRepo.GetByEmail(context.Background(), req.Email)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			response.ErrorUnauthorized(c, "Invalid credentials", nil)
		} else {
			response.ErrorInternalServer(c, "Database error", err.Error())
		}
		return
	}

	// Verify password
	if err := h.passwordService.VerifyPassword(user.Password, req.Password); err != nil {
		response.ErrorUnauthorized(c, "Invalid credentials", nil)
		return
	}

	// Set domain
	domain := req.Domain
	if domain == "" {
		domain = "*" // Default domain
	}

	// Get user roles for the domain
	userRoles, err := h.userRoleRepo.GetByUserAndDomain(context.Background(), user.ID, domain)
	if err != nil {
		response.ErrorInternalServer(c, "Failed to get user roles", err.Error())
		return
	}

	var roles []string
	for _, userRole := range userRoles {
		if userRole.Role != nil {
			roles = append(roles, userRole.Role.Name)
		}
	}

	// Generate JWT token
	username := ""
	if user.Username != nil {
		username = *user.Username
	}
	
	token, err := h.jwtService.GenerateToken(user.ID, *user.Email, username, user.Name, domain)
	if err != nil {
		response.ErrorInternalServer(c, "Failed to generate token", err.Error())
		return
	}

	// Remove password from response
	user.Password = ""

	response.SuccessOK(c, "Login successful", LoginResponse{
		Token:     token,
		User:      user,
		Roles:     roles,
		Domain:    domain,
		ExpiresAt: 0, // You might want to calculate this from JWT claims
	})
}

// Register handles user registration
func (h *AuthHandler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorBadRequest(c, "Invalid request data", err.Error())
		return
	}

	// Check if user already exists
	existingUser, err := h.userRepo.GetByEmail(context.Background(), req.Email)
	if err == nil && existingUser != nil {
		response.ErrorBadRequest(c, "User with this email already exists", nil)
		return
	}

	// Hash password
	hashedPassword, err := h.passwordService.HashPassword(req.Password)
	if err != nil {
		response.ErrorInternalServer(c, "Failed to hash password", err.Error())
		return
	}

	// Create user
	user := &entities.User{
		Email:    &req.Email,
		Username: &req.Username,
		Name:     req.Name,
		Password: hashedPassword,
		Pin:      req.Pin,
	}

	if err := h.userRepo.Create(context.Background(), user); err != nil {
		response.ErrorInternalServer(c, "Failed to create user", err.Error())
		return
	}

	// Assign default role (if exists)
	defaultRole, err := h.roleRepo.GetByName(context.Background(), "user")
	if err == nil && defaultRole != nil {
		domain := req.Domain
		if domain == "" {
			domain = "*"
		}

		userRole := &entities.UserRole{
			UserID: user.ID,
			RoleID: defaultRole.ID,
			Domain: domain,
		}

		if err := h.userRoleRepo.Create(context.Background(), userRole); err == nil {
			// Add role to Casbin
			h.enforcerService.AddRoleForUser(user.ID.String(), defaultRole.Name, domain)
		}
	}

	// Remove password from response
	user.Password = ""

	response.SuccessCreated(c, "User registered successfully", user)
}

// Logout handles user logout (for client-side token invalidation)
func (h *AuthHandler) Logout(c *gin.Context) {
	response.SuccessOK(c, "Logout successful", gin.H{
		"message": "Please remove the token from client storage",
	})
}

// RefreshToken handles token refresh
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	// Get current claims from context (set by auth middleware)
	claimsInterface, exists := c.Get("claims")
	if !exists {
		response.ErrorUnauthorized(c, "No valid token found", nil)
		return
	}

	claims, ok := claimsInterface.(*auth.Claims)
	if !ok {
		response.ErrorUnauthorized(c, "Invalid token claims", nil)
		return
	}

	// Generate new token
	newToken, err := h.jwtService.RefreshToken(claims)
	if err != nil {
		response.ErrorInternalServer(c, "Failed to refresh token", err.Error())
		return
	}

	response.SuccessOK(c, "Token refreshed successfully", gin.H{
		"token": newToken,
	})
}

// Profile returns the current user's profile
func (h *AuthHandler) Profile(c *gin.Context) {
	userID, exists := auth.GetUserIDFromContext(c)
	if !exists {
		response.ErrorUnauthorized(c, "User not authenticated", nil)
		return
	}

	user, err := h.userRepo.GetByID(context.Background(), userID)
	if err != nil {
		response.ErrorNotFound(c, "User not found", nil)
		return
	}

	// Get user roles
	userRoles, err := h.userRoleRepo.GetByUserID(context.Background(), userID)
	if err != nil {
		response.ErrorInternalServer(c, "Failed to get user roles", err.Error())
		return
	}

	var roles []string
	for _, userRole := range userRoles {
		if userRole.Role != nil {
			roles = append(roles, userRole.Role.Name)
		}
	}

	// Remove password from response
	user.Password = ""

	response.SuccessOK(c, "Profile retrieved successfully", gin.H{
		"user":  user,
		"roles": roles,
	})
}

// GetPermissions returns the current user's permissions
func (h *AuthHandler) GetPermissions(c *gin.Context) {
	userID, exists := auth.GetUserIDFromContext(c)
	if !exists {
		response.ErrorUnauthorized(c, "User not authenticated", nil)
		return
	}

	domain, _ := auth.GetUserDomainFromContext(c)
	if domain == "" {
		domain = "*"
	}

	// Get user roles from Casbin
	roles := h.enforcerService.GetRolesForUser(userID.String(), domain)

	// Get all policies to determine permissions
	allPolicies := h.enforcerService.GetAllPolicies()
	var permissions []map[string]string

	for _, policy := range allPolicies {
		if len(policy) >= 4 {
			// Check if this policy applies to any of the user's roles
			for _, userRole := range roles {
				if policy[0] == userRole && policy[1] == domain {
					permissions = append(permissions, map[string]string{
						"role":   policy[0],
						"domain": policy[1],
						"object": policy[2],
						"action": policy[3],
					})
				}
			}
		}
	}

	response.SuccessOK(c, "Permissions retrieved successfully", gin.H{
		"roles":       roles,
		"permissions": permissions,
		"domain":      domain,
	})
}