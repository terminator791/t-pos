package handlers

import (
	"context"
	"log"
	"strings"
	"time"

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
	userDomainRepo  repositories.UserDomainRepository
	roleRepo        repositories.RoleRepository
	licenseRepo     repositories.LicenseRepository
	shopRepo        repositories.ShopRepository
	jwtService      *auth.JWTService
	passwordService *auth.PasswordService
	enforcerService *casbin.EnforcerService
}

// LoginRequest represents login request payload
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Pin      string `json:"pin" binding:"required"`
	Domain   string `json:"domain,omitempty"` // Optional tenant/shop domain
}

// RegisterRequest represents registration request payload
type RegisterRequest struct {
	Username     string `json:"username" binding:"required,min=3"`
	SerialNumber string `json:"serial_number" binding:"required"`
	// Pin          string `json:"pin" binding:"required,min=6,max=6"`
	// Name         string `json:"name" binding:"required"`
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
	userDomainRepo repositories.UserDomainRepository,
	roleRepo repositories.RoleRepository,
	licenseRepo repositories.LicenseRepository,
	shopRepo repositories.ShopRepository,
	jwtService *auth.JWTService,
	passwordService *auth.PasswordService,
	enforcerService *casbin.EnforcerService,
) *AuthHandler {
	return &AuthHandler{
		userRepo:        userRepo,
		userDomainRepo:  userDomainRepo,
		roleRepo:        roleRepo,
		licenseRepo:     licenseRepo,
		shopRepo:        shopRepo,
		jwtService:      jwtService,
		passwordService: passwordService,
		enforcerService: enforcerService,
	}
}

// Login handles user login with new single role system
func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorBadRequest(c, "Invalid request data", err.Error())
		return
	}

	// Find user by email
	user, err := h.userRepo.GetByUsername(context.Background(), req.Username)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			response.ErrorUnauthorized(c, "Invalid credentials", nil)
		} else {
			response.ErrorInternalServer(c, "Database error", err.Error())
		}
		return
	}

	// Verify pin
	if user.Pin == nil {
		response.ErrorUnauthorized(c, "PIN not set", nil)
		return
	}
	if err := h.passwordService.VerifyPin(*user.Pin, req.Pin); err != nil {
		response.ErrorUnauthorized(c, "Invalid credentials", nil)
		return
	}

	// Get user's role
	if user.RoleID == nil {
		response.ErrorUnauthorized(c, "User has no assigned role", nil)
		return
	}

	role, err := h.roleRepo.GetByID(context.Background(), *user.RoleID)
	if err != nil {
		response.ErrorInternalServer(c, "Failed to get user role", err.Error())
		return
	}

	// Set domain for token
	domain := req.Domain
	if domain == "" {
		// Get user's accessible domains
		userDomains, err := h.userDomainRepo.GetByUserID(context.Background(), user.ID)
		if err != nil || len(userDomains) == 0 {
			// Fallback based on role
			switch role.Name {
			case "super_admin":
				domain = "*"
			case "admin":
				domain = "shop*"
			default:
				domain = "shop*" // Default fallback
			}
		} else {
			// Collect all domains into a slice of strings
			var domains []string
			for _, ud := range userDomains {
				domains = append(domains, ud.Domain)
			}
			// Join all domains into a single comma-separated string
			domain = strings.Join(domains, ",")
		}
	}

	// Verify user has access to requested domain
	if req.Domain != "" {
		hasAccess := false
		userDomains, err := h.userDomainRepo.GetByUserID(context.Background(), user.ID)
		if err == nil {
			for _, ud := range userDomains {
				if ud.Domain == req.Domain || ud.Domain == "*" || ud.Domain == "shop*" {
					hasAccess = true
					break
				}
			}
		}
		if !hasAccess {
			response.ErrorUnauthorized(c, "Access denied to requested domain", nil)
			return
		}
	}

	// Generate JWT token
	username := ""
	if user.Username != nil {
		username = *user.Username
	}
	
	email := ""
	if user.Email != nil {
		email = *user.Email
	}
	
	token, err := h.jwtService.GenerateToken(user.ID, email, username, user.Name, domain)
	if err != nil {
		response.ErrorInternalServer(c, "Failed to generate token", err.Error())
		return
	}

	// Calculate expires at
	expiresAt := time.Now().Add(h.jwtService.GetExpiryTime()).Unix()

	// Remove password from response
	user.Password = ""

	response.SuccessOK(c, "Login successful", LoginResponse{
		Token:     token,
		User:      user,
		Roles:     []string{role.Name}, // Single role
		Domain:    domain,
		ExpiresAt: expiresAt,
	})
}

// Register handles user registration with new single role system
func (h *AuthHandler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorBadRequest(c, "Invalid request data", err.Error())
		return
	}

	// Check if user already exists by username
	existingUser, err := h.userRepo.GetByUsername(context.Background(), req.Username)
	if err == nil && existingUser != nil {
		response.ErrorBadRequest(c, "User with this username already exists", nil)
		return
	}

	// Check if license exists
	license, err := h.licenseRepo.GetBySerialNumber(context.Background(), req.SerialNumber)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			response.ErrorBadRequest(c, "License not found", nil)
		} else {
			response.ErrorInternalServer(c, "Database error", err.Error())
		}
		return
	}

	// Get default role (owner_business) for new registrations
	defaultRole, err := h.roleRepo.GetByName(context.Background(), "owner_business")
	if err != nil {
		response.ErrorInternalServer(c, "Default role not found", err.Error())
		return
	}

	// hash pin
	// hashedPin, err := h.passwordService.HashPin(req.Pin)
	// if err != nil {
	// 	response.ErrorInternalServer(c, "Failed to hash pin", err.Error())
	// 	return
	// }

	// Create user with role
	user := &entities.User{
		Username:  &req.Username,
		LicenseID: &license.ID,
		RoleID:    &defaultRole.ID,
		// Pin:       &hashedPin,
	}

	if err := h.userRepo.Create(context.Background(), user); err != nil {
		response.ErrorInternalServer(c, "Failed to create user", err.Error())
		return
	}

	// Assign domain access based on role
	var domains []string
	
	switch defaultRole.Name {
	case "super_admin":
		// Super admin gets access to all domains
		domains = []string{"*"}
	case "admin":
		// Admin gets access to all shops (shop*)
		domains = []string{"shop*"}
	case "owner_business":
		// Owner business gets access to shops under their license
		shops, err := h.shopRepo.GetByLicenseID(context.Background(), license.ID)
		if err != nil {
			response.ErrorInternalServer(c, "Failed to get shops for license", err.Error())
			return
		}
		for _, shop := range shops {
			domains = append(domains, shop.Domain)
		}
		// If no shops exist yet, give access to all shops under license (they can create shops later)
		if len(domains) == 0 {
			domains = []string{"shop*"}
		}
	case "cashier":
		// Cashier needs to be assigned to specific shop - this would be done by owner_business later
		// For now, give no domain access (will be assigned by owner)
		domains = []string{}
	}

	// Create user domain records
	for _, domain := range domains {
		userDomain := &entities.UserDomain{
			UserID: user.ID,
			Domain: domain,
		}
		
		if err := h.userDomainRepo.Create(context.Background(), userDomain); err != nil {
			// Log error but don't fail registration
			log.Printf("Failed to create user domain %s for user %s: %v", domain, user.ID, err)
		}

		// Add role to Casbin for this domain
		h.enforcerService.AddRoleForUser(user.ID.String(), defaultRole.Name, domain)
	}

	// Remove password from response
	user.Password = ""

	response.SuccessCreated(c, "User registered successfully", user)
}

// Logout handles user logout (for client-side token invalidation)
func (h *AuthHandler) Logout(c *gin.Context) {
	// Extract token from header
	authHeader := c.GetHeader("Authorization")
	tokenString := h.jwtService.ExtractTokenFromHeader(authHeader)
	if tokenString != "" {
		// Blacklist the token
		h.jwtService.BlacklistToken(tokenString)
	}

	response.SuccessOK(c, "Logout successful", gin.H{
		"message": "Token has been invalidated",
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

// Profile returns the current user's profile with new single role system
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

	// Get user's role
	var roleName string
	if user.RoleID != nil {
		role, err := h.roleRepo.GetByID(context.Background(), *user.RoleID)
		if err == nil {
			roleName = role.Name
		}
	}

	// Get user's accessible domains
	userDomains, err := h.userDomainRepo.GetByUserID(context.Background(), userID)
	if err != nil {
		response.ErrorInternalServer(c, "Failed to get user domains", err.Error())
		return
	}

	var domains []string
	for _, ud := range userDomains {
		domains = append(domains, ud.Domain)
	}

	// Remove password from response
	user.Password = ""

	response.SuccessOK(c, "Profile retrieved successfully", gin.H{
		"user":    user,
		"role":    roleName, // Single role
		"domains": domains,  // Multiple domains
	})
}

// GetPermissions returns the current user's permissions with new single role system
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

	// Get user
	user, err := h.userRepo.GetByID(context.Background(), userID)
	if err != nil {
		response.ErrorInternalServer(c, "Failed to get user", err.Error())
		return
	}

	// Get user's role
	var roleName string
	if user.RoleID != nil {
		role, err := h.roleRepo.GetByID(context.Background(), *user.RoleID)
		if err == nil {
			roleName = role.Name
		}
	}

	// Get user roles from Casbin for the specific domain
	roles := h.enforcerService.GetRolesForUser(userID.String(), domain)

	// Get all policies to determine permissions
	allPolicies := h.enforcerService.GetAllPolicies()
	var permissions []map[string]string

	for _, policy := range allPolicies {
		if len(policy) >= 4 {
			// Check if this policy applies to the user's role and domain
			if policy[0] == roleName && (policy[1] == domain || policy[1] == "*" || policy[1] == "shop*") {
				permissions = append(permissions, map[string]string{
					"role":   policy[0],
					"domain": policy[1],
					"object": policy[2],
					"action": policy[3],
				})
			}
		}
	}

	// Get user's accessible domains
	userDomains, err := h.userDomainRepo.GetByUserID(context.Background(), userID)
	if err != nil {
		response.ErrorInternalServer(c, "Failed to get user domains", err.Error())
		return
	}

	var accessibleDomains []string
	for _, ud := range userDomains {
		accessibleDomains = append(accessibleDomains, ud.Domain)
	}

	response.SuccessOK(c, "Permissions retrieved successfully", gin.H{
		"role":                roleName,         // Single role
		"casbin_roles":        roles,            // Casbin roles for this domain
		"permissions":         permissions,      // Permissions for this domain
		"current_domain":      domain,           // Current domain context
		"accessible_domains":  accessibleDomains, // All domains user can access
	})
}