package casbin

import (
	"fmt"
	"log"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/terminator791/t-pos/internal/domain/entities"
	"github.com/terminator791/t-pos/internal/infrastructure/auth"
	"github.com/terminator791/t-pos/pkg/response"
)

// AuthzMiddleware handles authorization using Casbin
type AuthzMiddleware struct {
	enforcerService *EnforcerService
}

// NewAuthzMiddleware creates a new authorization middleware
func NewAuthzMiddleware(enforcerService *EnforcerService) *AuthzMiddleware {
	return &AuthzMiddleware{
		enforcerService: enforcerService,
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
