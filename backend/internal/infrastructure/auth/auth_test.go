package auth

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestJWTService(t *testing.T) {
	jwtService := NewJWTService("test-secret", "test-issuer", 24)
	
	// Test token generation
	userID := uuid.New()
	email := "test@example.com"
	username := "testuser"
	name := "Test User"
	domain := "test-domain"
	
	token, err := jwtService.GenerateToken(userID, email, username, name, domain)
	assert.NoError(t, err)
	assert.NotEmpty(t, token)
	
	// Test token validation
	claims, err := jwtService.ValidateToken(token)
	assert.NoError(t, err)
	assert.Equal(t, userID, claims.UserID)
	assert.Equal(t, email, claims.Email)
	assert.Equal(t, username, claims.Username)
	assert.Equal(t, name, claims.Name)
	assert.Equal(t, domain, claims.Domain)
	
	// Test invalid token
	_, err = jwtService.ValidateToken("invalid-token")
	assert.Error(t, err)
	
	// Test token refresh
	newToken, err := jwtService.RefreshToken(claims)
	assert.NoError(t, err)
	assert.NotEmpty(t, newToken)
	// Note: tokens may be same if generated at exact same time, just verify it's valid
	newClaims, err := jwtService.ValidateToken(newToken)
	assert.NoError(t, err)
	assert.Equal(t, userID, newClaims.UserID)
}

func TestPasswordService(t *testing.T) {
	passwordService := NewPasswordService()
	
	password := "testpassword123"
	
	// Test password hashing
	hashedPassword, err := passwordService.HashPassword(password)
	assert.NoError(t, err)
	assert.NotEmpty(t, hashedPassword)
	assert.NotEqual(t, password, hashedPassword)
	
	// Test password verification - correct password
	err = passwordService.VerifyPassword(hashedPassword, password)
	assert.NoError(t, err)
	
	// Test password verification - incorrect password
	err = passwordService.VerifyPassword(hashedPassword, "wrongpassword")
	assert.Error(t, err)
}