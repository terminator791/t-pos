package auth

import (
	"errors"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// JWTService handles JWT token operations
type JWTService struct {
	secret     []byte
	issuer     string
	expiryTime time.Duration
	blacklist  map[string]bool
	mutex      sync.RWMutex
}

// Claims represents the JWT claims structure
type Claims struct {
	UserID   uuid.UUID `json:"user_id"`
	Email    string    `json:"email"`
	Username string    `json:"username"`
	Name     string    `json:"name"`
	Domain   string    `json:"domain,omitempty"` // tenant/shop domain
	jwt.RegisteredClaims
}

// NewJWTService creates a new JWT service instance
func NewJWTService(secret string, issuer string, expiryHours int) *JWTService {
	return &JWTService{
		secret:     []byte(secret),
		issuer:     issuer,
		expiryTime: time.Duration(expiryHours) * time.Hour,
		blacklist:  make(map[string]bool),
	}
}

// GenerateToken generates a new JWT token for the given user
func (j *JWTService) GenerateToken(userID uuid.UUID, email, username, name, domain string) (string, error) {
	now := time.Now()
	claims := &Claims{
		UserID:   userID,
		Email:    email,
		Username: username,
		Name:     name,
		Domain:   domain,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    j.issuer,
			Subject:   userID.String(),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(j.expiryTime)),
			NotBefore: jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(j.secret)
}

// ValidateToken validates and parses a JWT token
func (j *JWTService) ValidateToken(tokenString string) (*Claims, error) {
	// Check if token is blacklisted
	j.mutex.RLock()
	if j.blacklist[tokenString] {
		j.mutex.RUnlock()
		return nil, errors.New("token has been revoked")
	}
	j.mutex.RUnlock()

	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// Validate the signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("invalid signing method")
		}
		return j.secret, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}

// RefreshToken generates a new token with extended expiry time
func (j *JWTService) RefreshToken(claims *Claims) (string, error) {
	// Create new claims with extended expiry
	now := time.Now()
	newClaims := &Claims{
		UserID:   claims.UserID,
		Email:    claims.Email,
		Username: claims.Username,
		Name:     claims.Name,
		Domain:   claims.Domain,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    j.issuer,
			Subject:   claims.UserID.String(),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(j.expiryTime)),
			NotBefore: jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, newClaims)
	return token.SignedString(j.secret)
}

// GetExpiryTime returns the token expiry duration
func (j *JWTService) GetExpiryTime() time.Duration {
	return j.expiryTime
}

// BlacklistToken adds a token to the blacklist
func (j *JWTService) BlacklistToken(tokenString string) {
	j.mutex.Lock()
	defer j.mutex.Unlock()
	j.blacklist[tokenString] = true
}

// ExtractTokenFromHeader extracts token from Authorization header
func (j *JWTService) ExtractTokenFromHeader(authHeader string) string {
	if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
		return authHeader[7:]
	}
	return ""
}