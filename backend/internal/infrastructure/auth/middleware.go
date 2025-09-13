package auth

import (
	"context"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/terminator791/t-pos/internal/domain/entities"
	"github.com/terminator791/t-pos/internal/domain/repositories"
	"github.com/terminator791/t-pos/pkg/response"
)

// AuthMiddleware handles JWT authentication
type AuthMiddleware struct {
	jwtService *JWTService
	userRepo   repositories.UserRepository
	roleRepo   repositories.RoleRepository
	shopRepo   repositories.ShopRepository
}

// NewAuthMiddleware creates a new authentication middleware
func NewAuthMiddleware(jwtService *JWTService, userRepo repositories.UserRepository, roleRepo repositories.RoleRepository, shopRepo repositories.ShopRepository) *AuthMiddleware {
	return &AuthMiddleware{
		jwtService: jwtService,
		userRepo:   userRepo,
		roleRepo:   roleRepo,
		shopRepo:   shopRepo,
	}
}

// RequireAuth middleware validates JWT token and sets user context
func (m *AuthMiddleware) RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.ErrorUnauthorized(c, "Authorization header required", nil)
			c.Abort()
			return
		}

		tokenString := m.jwtService.ExtractTokenFromHeader(authHeader)
		if tokenString == "" {
			response.ErrorUnauthorized(c, "Invalid authorization header format", nil)
			c.Abort()
			return
		}

		claims, err := m.jwtService.ValidateToken(tokenString)
		if err != nil {
			response.ErrorUnauthorized(c, "Invalid or expired token", err.Error())
			c.Abort()
			return
		}

		// Verify user still exists in database
		user, err := m.userRepo.GetByID(context.Background(), claims.UserID)
		if err != nil {
			response.ErrorUnauthorized(c, "User not found", nil)
			c.Abort()
			return
		}

		// Validate and set domain from database
		domain, err := m.validateUserDomain(user)
		if err != nil {
			response.ErrorUnauthorized(c, "Invalid domain access", err.Error())
			c.Abort()
			return
		}

		// Set user context
		c.Set("user_id", claims.UserID)
		c.Set("user_email", claims.Email)
		c.Set("user_username", claims.Username)
		c.Set("user_name", claims.Name)
		c.Set("user_domain", domain)  // Use validated domain from database
		c.Set("user_shop_id", claims.ShopID)
		c.Set("user", user)
		c.Set("claims", claims)

		c.Next()
	}
}

// OptionalAuth middleware validates JWT token if present but doesn't require it
func (m *AuthMiddleware) OptionalAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.Next()
			return
		}

		tokenString := m.jwtService.ExtractTokenFromHeader(authHeader)
		if tokenString == "" {
			c.Next()
			return
		}

		claims, err := m.jwtService.ValidateToken(tokenString)
		if err != nil {
			c.Next()
			return
		}

		// Verify user still exists in database
		user, err := m.userRepo.GetByID(context.Background(), claims.UserID)
		if err != nil {
			c.Next()
			return
		}

		// Validate domain from database (optional auth, so don't fail on domain errors)
		domain, err := m.validateUserDomain(user)
		if err != nil {
			// For optional auth, continue without setting domain if validation fails
			c.Next()
			return
		}

		// Set user context
		c.Set("user_id", claims.UserID)
		c.Set("user_email", claims.Email)
		c.Set("user_username", claims.Username)
		c.Set("user_name", claims.Name)
		c.Set("user_domain", domain)  // Use validated domain from database
		c.Set("user_shop_id", claims.ShopID)
		c.Set("user", user)
		c.Set("claims", claims)

		c.Next()
	}
}

// GetUserFromContext retrieves user ID from gin context
func GetUserIDFromContext(c *gin.Context) (uuid.UUID, bool) {
	if userID, exists := c.Get("user_id"); exists {
		if id, ok := userID.(uuid.UUID); ok {
			return id, true
		}
	}
	return uuid.Nil, false
}

// GetUserDomainFromContext retrieves user domain from gin context
func GetUserDomainFromContext(c *gin.Context) (string, bool) {
	if domain, exists := c.Get("user_domain"); exists {
		if d, ok := domain.(string); ok {
			return d, true
		}
	}
	return "", false
}

// GetUserShopIDFromContext retrieves user shop ID from gin context
func GetUserShopIDFromContext(c *gin.Context) (*uuid.UUID, bool) {
	if shopID, exists := c.Get("user_shop_id"); exists {
		if id, ok := shopID.(*uuid.UUID); ok {
			return id, true
		}
	}
	return nil, false
}

// GetUserFromContext retrieves user entity from gin context
func GetUserFromContext(c *gin.Context) (*entities.User, bool) {
	if user, exists := c.Get("user"); exists {
		if u, ok := user.(*entities.User); ok {
			return u, true
		}
	}
	return nil, false
}

// validateUserDomain validates and determines the appropriate domain for a user from database
func (m *AuthMiddleware) validateUserDomain(user *entities.User) (string, error) {
	ctx := context.Background()
	
	// Get user role information
	if user.RoleID == nil {
		return "", fmt.Errorf("user has no role assigned")
	}
	
	role, err := m.roleRepo.GetByID(ctx, *user.RoleID)
	if err != nil {
		return "", fmt.Errorf("failed to get user role: %w", err)
	}
	
	// Determine domain based on role and user assignments
	switch role.Name {
	case "super_admin", "admin":
		// Global access users
		return "*", nil
		
	case "owner_business":
		// Owner business users must have a license
		if user.LicenseID == nil {
			return "", fmt.Errorf("owner_business user missing license assignment")
		}
		
		// Validate license exists and create domain from license serial
		// For consistency with existing implementation, use license serial format
		return fmt.Sprintf("%s", user.LicenseID.String()), nil
		
	case "cashier":
		// Cashier users must have a shop assignment
		if user.ShopID == nil {
			return "", fmt.Errorf("cashier user missing shop assignment")
		}
		
		// Validate shop exists and user has access
		shop, err := m.shopRepo.GetByID(ctx, *user.ShopID)
		if err != nil {
			return "", fmt.Errorf("assigned shop not found: %w", err)
		}
		
		// Ensure shop is accessible (cross-validation with license if needed)
		if user.LicenseID != nil && shop.LicenseID != *user.LicenseID {
			return "", fmt.Errorf("shop license mismatch: user license %s, shop license %s", 
				user.LicenseID.String(), shop.LicenseID.String())
		}
		
		// Create shop-specific domain
		return fmt.Sprintf("shop-%s", user.ShopID.String()), nil
		
	default:
		return "", fmt.Errorf("unknown role: %s", role.Name)
	}
}
