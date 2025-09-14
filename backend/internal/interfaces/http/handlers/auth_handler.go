package handlers

import (
	"context"
	"log"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/terminator791/t-pos/internal/domain/entities"
	"github.com/terminator791/t-pos/internal/domain/repositories"
	"github.com/terminator791/t-pos/internal/infrastructure/auth"
	"github.com/terminator791/t-pos/internal/infrastructure/casbin"
	"github.com/terminator791/t-pos/internal/infrastructure/seeders"
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
	authSeeder      *seeders.AuthSeeder
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
	Domain       string `json:"domain,omitempty"` // Optional tenant/shop domain
	// Pin          string `json:"pin" binding:"required,min=6,max=6"`
	// Name         string `json:"name" binding:"required"`
}

// RegisterCashierRequest represents cashier registration request payload
type RegisterCashierRequest struct {
	Username string    `json:"username" binding:"required,min=3"`
	ShopID   uuid.UUID `json:"shop_id" binding:"required"`
	Name     string    `json:"name"`
}

type CreatePinRequest struct {
	Pin string `json:"pin" binding:"required,min=6,max=6"`
}

type UpdatePinRequest struct {
	OldPin        string `json:"old_pin" binding:"required,min=6,max=6"`
	NewPin        string `json:"new_pin" binding:"required,min=6,max=6"`
	ConfirmNewPin string `json:"confirm_new_pin" binding:"required,min=6,max=6"`
}

// LoginResponse represents login response
type LoginResponse struct {
	Token     string         `json:"token"`
	User      *entities.User `json:"user"`
	Roles     []string       `json:"roles"`
	Domain    string         `json:"domain"`
	ExpiresAt int64          `json:"expires_at"`
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
	authSeeder *seeders.AuthSeeder,
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
		authSeeder:      authSeeder,
	}
}

// Login handles user login with new single role system
func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorBadRequest(c, "Invalid request data", err.Error())
		return
	}

	// Find user by username
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

	// Get license to get serialNumber or handle users without license
	var domain string
	if user.LicenseID != nil {
		license, err := h.licenseRepo.GetByID(context.Background(), *user.LicenseID)
		if err != nil {
			response.ErrorInternalServer(c, "Failed to get license", err.Error())
			return
		}
		domain = license.SerialNumber
	} else {
		// For users without license cant login
		response.ErrorForbidden(c, "User has no license assigned", nil)
		return
	}

	// check in role is owner_business
	if role.Name != "owner_business" {
		response.ErrorForbidden(c, "This endpoint is for owner business only", nil)
		return
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

	// not set shopID (its always owner_business login)
	var shopID *uuid.UUID = nil

	token, err := h.jwtService.GenerateToken(user.ID, email, username, user.Name, domain, shopID)
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

	// Set domain for token
	domain := req.SerialNumber

	email := ""
	if user.Email != nil {
		email = *user.Email
	}

	username := ""
	if user.Username != nil {
		username = *user.Username
	}

	name := ""
	if user.Name != "" {
		name = user.Name
	}

	// New users typically don't have shop_id assigned yet during registration
	var shopID *uuid.UUID = nil

	token, err := h.jwtService.GenerateToken(user.ID, email, username, name, domain, shopID)
	if err != nil {
		response.ErrorInternalServer(c, "Failed to generate token", err.Error())
		return
	}

	// Calculate expires at
	expiresAt := time.Now().Add(h.jwtService.GetExpiryTime()).Unix()

	// Assign domain access based on role (use license.SerialNumber as domain)
	var domains []string

	if defaultRole.Name == "owner_business" {
		// Owner business gets access to their license domain (can access all shops under it)
		domains = []string{license.SerialNumber}
		
		// Assign domain-specific policies for owner_business
		if err := h.authSeeder.AssignPoliciesForRole("owner_business", license.SerialNumber); err != nil {
			log.Printf("Failed to assign policies for owner_business %s: %v", user.ID, err)
			// Don't fail registration, but log the error
		}
	} else {
		// response with error since default role must be owner_business
		response.ErrorForbidden(c, "Role must be owner_business", nil)
		return
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

	data := LoginResponse{
		Token:     token,
		User:      user,
		Roles:     []string{defaultRole.Name}, // Single role
		Domain:    domain,
		ExpiresAt: expiresAt,
	}

	response.SuccessCreated(c, "User registered successfully", data)
}

// RegisterCashier handles cashier registration by owner_business
func (h *AuthHandler) RegisterCashier(c *gin.Context) {
	var req RegisterCashierRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorBadRequest(c, "Invalid request data", err.Error())
		return
	}

	// Get the owner_business from token to verify they can create cashiers
	userID, exists := auth.GetUserIDFromContext(c)
	if !exists {
		response.ErrorUnauthorized(c, "User not authenticated", nil)
		return
	}

	// Get the creating user (owner_business)
	creatingUser, err := h.userRepo.GetByID(context.Background(), userID)
	if err != nil {
		response.ErrorInternalServer(c, "Failed to get creating user", err.Error())
		return
	}

	// Verify creating user is owner_business and has license
	if creatingUser.LicenseID == nil {
		response.ErrorBadRequest(c, "Only license holders can create cashiers", nil)
		return
	}

	// Get user's role to verify it's owner_business
	if creatingUser.RoleID == nil {
		response.ErrorUnauthorized(c, "User has no assigned role", nil)
		return
	}

	role, err := h.roleRepo.GetByID(context.Background(), *creatingUser.RoleID)
	if err != nil {
		response.ErrorInternalServer(c, "Failed to get user role", err.Error())
		return
	}

	if role.Name != "owner_business" {
		response.ErrorForbidden(c, "Only owner business can create cashiers", nil)
		return
	}

	// Verify the shop belongs to the same license as the creating user
	shop, err := h.shopRepo.GetByID(context.Background(), req.ShopID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			response.ErrorBadRequest(c, "Shop not found", nil)
		} else {
			response.ErrorInternalServer(c, "Database error", err.Error())
		}
		return
	}

	// Check if shop belongs to the same license
	if shop.LicenseID != *creatingUser.LicenseID {
		response.ErrorForbidden(c, "Can only create cashiers for shops under your license", nil)
		return
	}

	// Check if user already exists by username
	existingUser, err := h.userRepo.GetByUsername(context.Background(), req.Username)
	if err == nil && existingUser != nil {
		response.ErrorBadRequest(c, "User with this username already exists", nil)
		return
	}

	// Get cashier role
	cashierRole, err := h.roleRepo.GetByName(context.Background(), "cashier")
	if err != nil {
		response.ErrorInternalServer(c, "Cashier role not found", err.Error())
		return
	}

	// Create cashier user
	user := &entities.User{
		Username:  &req.Username,
		LicenseID: creatingUser.LicenseID, // Inherit license from creating user
		RoleID:    &cashierRole.ID,
		ShopID:    &req.ShopID, // Bind to specific shop
		Name:      req.Name,
	}

	if err := h.userRepo.Create(context.Background(), user); err != nil {
		response.ErrorInternalServer(c, "Failed to create cashier", err.Error())
		return
	}

	// Set domain for token (shop domain)
	domain := shop.Domain

	email := ""
	if user.Email != nil {
		email = *user.Email
	}

	username := ""
	if user.Username != nil {
		username = *user.Username
	}

	name := ""
	if user.Name != "" {
		name = user.Name
	}

	// Generate token for cashier
	token, err := h.jwtService.GenerateToken(user.ID, email, username, name, domain, &req.ShopID)
	if err != nil {
		response.ErrorInternalServer(c, "Failed to generate token", err.Error())
		return
	}

	// Calculate expires at
	expiresAt := time.Now().Add(h.jwtService.GetExpiryTime()).Unix()

	// Create user domain record for shop-specific access
	userDomain := &entities.UserDomain{
		UserID: user.ID,
		Domain: domain, // shop domain
	}

	if err := h.userDomainRepo.Create(context.Background(), userDomain); err != nil {
		log.Printf("Failed to create user domain %s for cashier %s: %v", domain, user.ID, err)
	}

	// Assign domain-specific policies for cashier
	if err := h.authSeeder.AssignPoliciesForRole("cashier", domain); err != nil {
		log.Printf("Failed to assign policies for cashier %s: %v", user.ID, err)
		// Don't fail registration, but log the error
	}

	// Add role to Casbin for this shop domain
	h.enforcerService.AddRoleForUser(user.ID.String(), "cashier", domain)

	// Remove password from response
	user.Password = ""

	data := LoginResponse{
		Token:     token,
		User:      user,
		Roles:     []string{"cashier"},
		Domain:    domain,
		ExpiresAt: expiresAt,
	}

	response.SuccessCreated(c, "Cashier registered successfully", data)
}

// LoginCashier handles cashier login
func (h *AuthHandler) LoginCashier(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorBadRequest(c, "Invalid request data", err.Error())
		return
	}

	// Find user by username
	user, err := h.userRepo.GetByUsername(context.Background(), req.Username)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			response.ErrorUnauthorized(c, "Invalid credentials", nil)
		} else {
			response.ErrorInternalServer(c, "Database error", err.Error())
		}
		return
	}

	// Verify user is a cashier
	if user.RoleID == nil {
		response.ErrorUnauthorized(c, "User has no assigned role", nil)
		return
	}

	role, err := h.roleRepo.GetByID(context.Background(), *user.RoleID)
	if err != nil {
		response.ErrorInternalServer(c, "Failed to get user role", err.Error())
		return
	}

	if role.Name != "cashier" {
		response.ErrorUnauthorized(c, "This endpoint is for cashiers only", nil)
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

	// Verify cashier is assigned to a shop
	if user.ShopID == nil {
		response.ErrorUnauthorized(c, "Cashier not assigned to any shop", nil)
		return
	}

	// Get assigned shop
	shop, err := h.shopRepo.GetByID(context.Background(), *user.ShopID)
	if err != nil {
		response.ErrorInternalServer(c, "Failed to get assigned shop", err.Error())
		return
	}

	// Set domain to shop domain
	domain := shop.Domain

	// Generate JWT token
	username := ""
	if user.Username != nil {
		username = *user.Username
	}

	email := ""
	if user.Email != nil {
		email = *user.Email
	}

	token, err := h.jwtService.GenerateToken(user.ID, email, username, user.Name, domain, user.ShopID)
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
		Roles:     []string{role.Name},
		Domain:    domain,
		ExpiresAt: expiresAt,
	})
}

func (h *AuthHandler) CreatePin(c *gin.Context) {
	var req CreatePinRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorBadRequest(c, "Invalid request data", err.Error())
		return
	}

	authHeader := c.GetHeader("Authorization")
	tokenString := h.jwtService.ExtractTokenFromHeader(authHeader)

	// get user from token
	userID, err := h.jwtService.GetUserIDFromToken(tokenString)
	if err != nil {
		response.ErrorUnauthorized(c, "Invalid token", err.Error())
		return
	}

	// if user.Pin exist return eror pin already exist
	existingUser, err := h.userRepo.GetByID(context.Background(), userID)
	if err != nil {
		response.ErrorInternalServer(c, "Failed to get user", err.Error())
		return
	}
	if existingUser.Pin != nil {
		response.ErrorBadRequest(c, "PIN already exists", nil)
		return
	}

	// Hash the PIN
	hashedPin, err := h.passwordService.HashPin(req.Pin)
	if err != nil {
		response.ErrorInternalServer(c, "Failed to hash PIN", err.Error())
		return
	}

	// Create PIN for user
	if err := h.userRepo.CreatePin(context.Background(), userID, hashedPin); err != nil {
		response.ErrorInternalServer(c, "Failed to create PIN", err.Error())
		return
	}

	response.SuccessCreated(c, "PIN created successfully", nil)
}

func (h *AuthHandler) UpdatePin(c *gin.Context) {
	var req UpdatePinRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorBadRequest(c, "Invalid request data", err.Error())
		return
	}

	authHeader := c.GetHeader("Authorization")
	tokenString := h.jwtService.ExtractTokenFromHeader(authHeader)

	// get user from token
	userID, err := h.jwtService.GetUserIDFromToken(tokenString)
	if err != nil {
		response.ErrorUnauthorized(c, "Invalid token", err.Error())
		return
	}

	// Get user to verify old PIN
	user, err := h.userRepo.GetByID(context.Background(), userID)
	if err != nil {
		response.ErrorInternalServer(c, "Failed to get user", err.Error())
		return
	}

	// Check if PIN exists
	if user.Pin == nil {
		response.ErrorBadRequest(c, "PIN not set", nil)
		return
	}

	// Verify old PIN
	if err := h.passwordService.VerifyPin(*user.Pin, req.OldPin); err != nil {
		response.ErrorBadRequest(c, "Old PIN is incorrect", nil)
		return
	}

	// Check if new PIN matches confirmation
	if req.NewPin != req.ConfirmNewPin {
		response.ErrorBadRequest(c, "New PIN and confirmation do not match", nil)
		return
	}

	// Check if new PIN is different from old PIN
	if req.NewPin == req.OldPin {
		response.ErrorBadRequest(c, "New PIN cannot be the same as old PIN", nil)
		return
	}

	// Hash the new PIN
	hashedNewPin, err := h.passwordService.HashPin(req.NewPin)
	if err != nil {
		response.ErrorInternalServer(c, "Failed to hash new PIN", err.Error())
		return
	}

	// Update PIN for user
	if err := h.userRepo.UpdatePin(context.Background(), userID, hashedNewPin); err != nil {
		response.ErrorInternalServer(c, "Failed to update PIN", err.Error())
		return
	}

	response.SuccessOK(c, "PIN updated successfully", nil)
}

func (h *AuthHandler) DeletePin(c *gin.Context) {
	authHeader := c.GetHeader("Authorization")
	tokenString := h.jwtService.ExtractTokenFromHeader(authHeader)

	// get user from token
	userID, err := h.jwtService.GetUserIDFromToken(tokenString)
	if err != nil {
		response.ErrorUnauthorized(c, "Invalid token", err.Error())
		return
	}

	// Delete PIN for user
	if err := h.userRepo.DeletePin(context.Background(), userID); err != nil {
		response.ErrorInternalServer(c, "Failed to delete PIN", err.Error())
		return
	}

	response.SuccessOK(c, "PIN deleted successfully", nil)
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

	// Get user's accessible domains (now based on license.SerialNumber)
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

	// Retrieve claims directly from context to get the domain
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

    domain := claims.Domain
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
		"role":               roleName,          // Single role
		"casbin_roles":       roles,             // Casbin roles for this domain
		"permissions":        permissions,       // Permissions for this domain
		"current_domain":     domain,            // Current domain context
		"accessible_domains": accessibleDomains, // All domains user can access
	})
}
