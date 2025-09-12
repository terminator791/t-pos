package casbin

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/terminator791/t-pos/internal/domain/entities"
	"github.com/terminator791/t-pos/internal/domain/repositories"
	"github.com/terminator791/t-pos/internal/infrastructure/auth"
	"github.com/terminator791/t-pos/pkg/response"
)

// AuthzMiddleware handles authorization using Casbin
type AuthzMiddleware struct {
	enforcerService *EnforcerService
	shopRepo        repositories.ShopRepository
	userRepo        repositories.UserRepository
	roleRepo        repositories.RoleRepository
}

// NewAuthzMiddleware creates a new authorization middleware
func NewAuthzMiddleware(enforcerService *EnforcerService, shopRepo repositories.ShopRepository, userRepo repositories.UserRepository, roleRepo repositories.RoleRepository) *AuthzMiddleware {
	return &AuthzMiddleware{
		enforcerService: enforcerService,
		shopRepo:        shopRepo,
		userRepo:        userRepo,
		roleRepo:        roleRepo,
	}
}

// RequirePermission middleware checks if user has permission for the requested resource
func (m *AuthzMiddleware) RequirePermission() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get user ID from context (set by auth middleware)
		userID, exists := auth.GetUserIDFromContext(c)
		if !exists {
			response.ErrorUnauthorized(c, "User not authenticated", nil)
			c.Abort()
			return
		}

		// Get user from context to check role
		userInterface, userExists := c.Get("user")
		if !userExists {
			response.ErrorUnauthorized(c, "User context not found", nil)
			c.Abort()
			return
		}
		
		user, ok := userInterface.(*entities.User)
		if !ok {
			response.ErrorInternalServer(c, "Invalid user context", nil)
			c.Abort()
			return
		}

		// Get domain from context
		domain, _ := auth.GetUserDomainFromContext(c)
		
		// Handle domain assignment based on role
		if domain == "" {
			// For tenant-specific users (owner_business, cashier), domain is required
			// Only allow wildcard access for super_admin and admin roles
			if user.RoleID != nil {
				// Get the role to check if wildcard is allowed
				// This is a safety check - ideally all users should have proper domains
				role, err := m.getUserRole(c.Request.Context(), *user.RoleID)
				if err != nil {
					response.ErrorInternalServer(c, "Failed to get user role", err.Error())
					c.Abort()
					return
				}
				
				// Only super_admin and admin can have wildcard access
				if role.Name == "super_admin" || role.Name == "admin" {
					domain = "*"
					log.Printf("Warning: User %s (%s role) has no domain, using wildcard", userID.String(), role.Name)
				} else {
					// For owner_business and cashier, domain is mandatory
					response.ErrorForbidden(c, "No domain assigned to user - multi-tenant isolation required", map[string]interface{}{
						"user":   userID.String(),
						"role":   role.Name,
						"object": c.FullPath(),
						"action": strings.ToUpper(c.Request.Method),
					})
					c.Abort()
					return
				}
			} else {
				response.ErrorForbidden(c, "No role or domain assigned to user", map[string]interface{}{
					"user":   userID.String(),
					"object": c.FullPath(),
					"action": strings.ToUpper(c.Request.Method),
				})
				c.Abort()
				return
			}
		}

		// Get request details
		object := c.FullPath() // Use full path for consistent pattern matching
		action := strings.ToUpper(c.Request.Method)

		// Check permission using Casbin
		allowed, err := m.enforcerService.Enforce(userID.String(), domain, object, action)
		if err != nil {
			response.ErrorInternalServer(c, "Authorization check failed", err.Error())
			c.Abort()
			return
		}

		if !allowed {
			response.ErrorForbidden(c, "Insufficient permissions", map[string]interface{}{
				"user":   userID.String(),
				"domain": domain,
				"object": object,
				"action": action,
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireRole middleware checks if user has a specific role
func (m *AuthzMiddleware) RequireRole(role string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get user ID from context
		userID, exists := auth.GetUserIDFromContext(c)
		if !exists {
			response.ErrorUnauthorized(c, "User not authenticated", nil)
			c.Abort()
			return
		}

		// Get domain from context
		domain, _ := auth.GetUserDomainFromContext(c)
		if domain == "" {
			domain = "*" // Default domain
		}

		// Get user roles from Casbin
		roles := m.enforcerService.GetRolesForUser(userID.String(), domain)

		// Check if user has the required role
		hasRole := false
		for _, userRole := range roles {
			if userRole == role {
				hasRole = true
				break
			}
		}

		if !hasRole {
			response.ErrorForbidden(c, fmt.Sprintf("Role '%s' required", role), map[string]interface{}{
				"user_roles":    roles,
				"required_role": role,
				"domain":        domain,
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireAnyRole middleware checks if user has any of the specified roles
func (m *AuthzMiddleware) RequireAnyRole(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get user ID from context
		userID, exists := auth.GetUserIDFromContext(c)
		if !exists {
			response.ErrorUnauthorized(c, "User not authenticated", nil)
			c.Abort()
			return
		}

		// Get domain from context
		domain, _ := auth.GetUserDomainFromContext(c)
		if domain == "" {
			domain = "*" // Default domain
		}

		// Get user roles from Casbin
		userRoles := m.enforcerService.GetRolesForUser(userID.String(), domain)

		// Check if user has any of the required roles
		hasRole := false
		for _, userRole := range userRoles {
			for _, requiredRole := range roles {
				if userRole == requiredRole {
					hasRole = true
					break
				}
			}
			if hasRole {
				break
			}
		}

		if !hasRole {
			response.ErrorForbidden(c, "Insufficient role permissions", map[string]interface{}{
				"user_roles":     userRoles,
				"required_roles": roles,
				"domain":         domain,
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// ValidateShopAccess validates if the current user can access the specified shop
func (m *AuthzMiddleware) ValidateShopAccess(c *gin.Context, shopID uuid.UUID) error {
	// Get user from context
	userInterface, userExists := c.Get("user")
	if !userExists {
		return fmt.Errorf("user context not found")
	}
	
	user, ok := userInterface.(*entities.User)
	if !ok {
		return fmt.Errorf("invalid user context")
	}

	// Get user domain to determine role and access level
	domain, _ := auth.GetUserDomainFromContext(c)
	
	// Get shop details
	shop, err := m.shopRepo.GetByID(c.Request.Context(), shopID)
	if err != nil {
		return fmt.Errorf("shop not found: %w", err)
	}

	// Check access based on domain and user context
	switch {
	case domain == "*":
		// Global access (super_admin, admin)
		return nil
	case user.LicenseID != nil && shop.LicenseID == *user.LicenseID:
		// Owner business accessing shop under their license
		return nil
	case user.ShopID != nil && *user.ShopID == shopID:
		// Cashier accessing their assigned shop
		return nil
	default:
		return fmt.Errorf("user cannot access shop %s (user domain: %s, user license: %v, user shop: %v, shop license: %s)", 
			shopID, domain, user.LicenseID, user.ShopID, shop.LicenseID)
	}
}

// RequireShopAccess middleware validates shop access for endpoints with :shopId parameter
func (m *AuthzMiddleware) RequireShopAccess() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extract shop ID from URL parameter
		shopIDStr := c.Param("shopId")
		if shopIDStr == "" {
			// No shop ID parameter, continue with normal flow
			c.Next()
			return
		}

		shopID, err := uuid.Parse(shopIDStr)
		if err != nil {
			response.ErrorBadRequest(c, "Invalid shop ID format", err.Error())
			c.Abort()
			return
		}

		// Validate shop access
		if err := m.ValidateShopAccess(c, shopID); err != nil {
			response.ErrorForbidden(c, "Cannot access this shop", map[string]interface{}{
				"shop_id": shopID,
				"error":   err.Error(),
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// ValidateLicenseAccess validates if the current user can access the specified license
func (m *AuthzMiddleware) ValidateLicenseAccess(c *gin.Context, licenseID uuid.UUID) error {
	// Get user from context
	userInterface, userExists := c.Get("user")
	if !userExists {
		return fmt.Errorf("user context not found")
	}
	
	user, ok := userInterface.(*entities.User)
	if !ok {
		return fmt.Errorf("invalid user context")
	}

	// Get user domain to determine access level
	domain, _ := auth.GetUserDomainFromContext(c)
	
	// Check access based on domain and user context
	switch {
	case domain == "*":
		// Global access (super_admin, admin)
		return nil
	case user.LicenseID != nil && *user.LicenseID == licenseID:
		// User accessing their own license
		return nil
	default:
		return fmt.Errorf("user cannot access license %s (user domain: %s, user license: %v)", 
			licenseID, domain, user.LicenseID)
	}
}

// RequireLicenseAccess middleware validates license access for endpoints with :licenseId parameter  
func (m *AuthzMiddleware) RequireLicenseAccess() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extract license ID from URL parameter
		licenseIDStr := c.Param("licenseId")
		if licenseIDStr == "" {
			// No license ID parameter, continue with normal flow
			c.Next()
			return
		}

		licenseID, err := uuid.Parse(licenseIDStr)
		if err != nil {
			response.ErrorBadRequest(c, "Invalid license ID format", err.Error())
			c.Abort()
			return
		}

		// Validate license access
		if err := m.ValidateLicenseAccess(c, licenseID); err != nil {
			response.ErrorForbidden(c, "Cannot access this license", map[string]interface{}{
				"license_id": licenseID,
				"error":      err.Error(),
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireResourceAccess middleware validates access to individual resources by checking their shop ownership
func (m *AuthzMiddleware) RequireResourceAccess(resourceType string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extract resource ID from URL parameter
		resourceIDStr := c.Param("id")
		if resourceIDStr == "" {
			// No resource ID parameter, continue with normal flow
			c.Next()
			return
		}

		resourceID, err := uuid.Parse(resourceIDStr)
		if err != nil {
			response.ErrorBadRequest(c, "Invalid resource ID format", err.Error())
			c.Abort()
			return
		}

		// Validate resource access
		if err := m.ValidateResourceAccess(c, resourceID, resourceType); err != nil {
			response.ErrorForbidden(c, "Cannot access this resource", map[string]interface{}{
				"resource_type": resourceType,
				"resource_id":   resourceID,
				"error":         err.Error(),
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// ValidateResourceAccess validates if the current user can access a resource by ID
// This is a basic implementation that can be extended for specific resource types
func (m *AuthzMiddleware) ValidateResourceAccess(c *gin.Context, resourceID uuid.UUID, resourceType string) error {
	// Get user from context
	userInterface, userExists := c.Get("user")
	if !userExists {
		return fmt.Errorf("user context not found")
	}
	
	user, ok := userInterface.(*entities.User)
	if !ok {
		return fmt.Errorf("invalid user context")
	}

	// Get user domain to determine access level
	domain, _ := auth.GetUserDomainFromContext(c)
	
	switch {
	case domain == "*":
		// Global access (super_admin, admin)
		return nil
	default:
		// For tenant-specific users, implement resource-specific validation
		switch resourceType {
		case "transaction":
			// For transactions, we need to validate the transaction belongs to an accessible shop
			return m.validateTransactionAccess(c.Request.Context(), user, domain, resourceID)
		case "shop":
			// For shops, validate direct shop access
			return m.ValidateShopAccess(c, resourceID)
		default:
			// For other resource types, log and allow (can be extended as needed)
			log.Printf("Resource access validation for %s:%s by user %s (domain: %s) - allowing access (validation not implemented)", 
				resourceType, resourceID, user.ID, domain)
			return nil
		}
	}
}

// validateTransactionAccess validates if a user can access a specific transaction
func (m *AuthzMiddleware) validateTransactionAccess(ctx context.Context, user *entities.User, domain string, transactionID uuid.UUID) error {
	// For this implementation, we need to check the transaction's shop
	// Since we don't have the transaction repository in middleware yet, 
	// we'll implement a simpler approach by checking user permissions
	
	// If user has shop assignment (cashier), they can only access transactions from their shop
	if user.ShopID != nil {
		// For cashiers, we would need to query the transaction to check its shop_id
		// For now, we'll allow access and log it for implementation
		log.Printf("Transaction access validation needed for cashier %s accessing transaction %s", user.ID, transactionID)
		// TODO: Implement transaction shop validation when transaction repository is available
		return nil
	}
	
	// For owner_business, they can access transactions from shops under their license
	if user.LicenseID != nil && domain != "*" {
		// Owner business users can access transactions from their license shops
		log.Printf("Transaction access validation for owner business %s accessing transaction %s", user.ID, transactionID)
		// TODO: Implement license-based transaction validation
		return nil
	}
	
	// For other cases, allow access
	return nil
}

// getUserRole gets the role for a user by role ID
func (m *AuthzMiddleware) getUserRole(ctx context.Context, roleID uuid.UUID) (*entities.Role, error) {
	return m.roleRepo.GetByID(ctx, roleID)
}
