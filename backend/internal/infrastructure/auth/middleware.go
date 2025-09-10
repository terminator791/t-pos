package auth

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/terminator791/t-pos/internal/domain/repositories"
	"github.com/terminator791/t-pos/pkg/response"
)

// AuthMiddleware handles JWT authentication
type AuthMiddleware struct {
	jwtService   *JWTService
	userRepo     repositories.UserRepository
}

// NewAuthMiddleware creates a new authentication middleware
func NewAuthMiddleware(jwtService *JWTService, userRepo repositories.UserRepository) *AuthMiddleware {
	return &AuthMiddleware{
		jwtService: jwtService,
		userRepo:   userRepo,
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

		// Set user context
		c.Set("user_id", claims.UserID)
		c.Set("user_email", claims.Email)
		c.Set("user_username", claims.Username)
		c.Set("user_name", claims.Name)
		c.Set("user_domain", claims.Domain)
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

		// Set user context
		c.Set("user_id", claims.UserID)
		c.Set("user_email", claims.Email)
		c.Set("user_username", claims.Username)
		c.Set("user_name", claims.Name)
		c.Set("user_domain", claims.Domain)
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