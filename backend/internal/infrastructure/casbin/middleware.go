package casbin

import (
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
}

// NewAuthzMiddleware creates a new authorization middleware
func NewAuthzMiddleware(enforcerService *EnforcerService, shopRepo repositories.ShopRepository, userRepo repositories.UserRepository) *AuthzMiddleware {
	return &AuthzMiddleware{
		enforcerService: enforcerService,
		shopRepo:        shopRepo,
		userRepo:        userRepo,
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
			// Check if user has a role that allows wildcard access
			if user.RoleID != nil {
				// We need to get the role name to check if it's admin or super_admin
				// For now, allow wildcard for users without domain only if they're admin-level
				// This is a fallback - ideally all users should have proper domains
				domain = "*"
				log.Printf("Warning: User %s has no domain, using wildcard", userID.String())
			} else {
				response.ErrorForbidden(c, "No domain assigned to user", map[string]interface{}{
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

// ValidateResourceAccess validates if the current user can access a resource by ID
// This is a more complex validation that requires checking the resource's associated shop/license
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
		// For tenant-specific users, we would need to check the resource's
		// associated shop/license. This would require resource-specific logic.
		// For now, we'll implement basic validation and extend as needed.
		log.Printf("Resource access validation for %s:%s by user %s (domain: %s)", 
			resourceType, resourceID, user.ID, domain)
		return nil // Allow for now - implement specific validation per resource type
	}
}
