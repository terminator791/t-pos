package utils

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// HashPassword hashes a password using bcrypt
func HashPassword(password string) (string, error) {
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedBytes), nil
}

// CheckPassword checks if the provided password matches the hash
func CheckPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// GenerateRandomString generates a random string of specified length
func GenerateRandomString(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes)[:length], nil
}

// GenerateOrderNumber generates a unique order number
func GenerateOrderNumber() string {
	now := time.Now()
	return fmt.Sprintf("ORD-%d%02d%02d-%d",
		now.Year(), now.Month(), now.Day(),
		now.Unix()%10000)
}

// GenerateTransactionID generates a unique transaction ID
func GenerateTransactionID() string {
	now := time.Now()
	return fmt.Sprintf("TXN-%d%02d%02d-%d",
		now.Year(), now.Month(), now.Day(),
		now.Unix()%10000)
}

// FormatCurrency formats a float64 as currency string
func FormatCurrency(amount float64) string {
	return fmt.Sprintf("$%.2f", amount)
}

// ParseCurrency parses a currency string to float64
func ParseCurrency(currency string) (float64, error) {
	// Remove currency symbol and parse
	cleanStr := strings.ReplaceAll(currency, "$", "")
	cleanStr = strings.ReplaceAll(cleanStr, ",", "")
	return strconv.ParseFloat(cleanStr, 64)
}

// RoundTo rounds a float64 to specified decimal places
func RoundTo(value float64, decimals int) float64 {
	multiplier := math.Pow(10, float64(decimals))
	return math.Round(value*multiplier) / multiplier
}

// CalculatePercentage calculates percentage of a value
func CalculatePercentage(value, percentage float64) float64 {
	return (value * percentage) / 100
}

// CalculateTax calculates tax amount
func CalculateTax(amount, taxRate float64) float64 {
	return RoundTo(CalculatePercentage(amount, taxRate), 2)
}

// CalculateDiscount calculates discount amount
func CalculateDiscount(amount, discountRate float64) float64 {
	return RoundTo(CalculatePercentage(amount, discountRate), 2)
}

// GetPaginationOffset calculates offset for pagination
func GetPaginationOffset(page, limit int) int {
	if page <= 0 {
		page = 1
	}
	return (page - 1) * limit
}

// ValidateEmail performs basic email validation
func ValidateEmail(email string) bool {
	return strings.Contains(email, "@") && strings.Contains(email, ".")
}

// ValidatePhone performs basic phone validation
func ValidatePhone(phone string) bool {
	// Remove non-numeric characters
	cleanPhone := strings.ReplaceAll(phone, " ", "")
	cleanPhone = strings.ReplaceAll(cleanPhone, "-", "")
	cleanPhone = strings.ReplaceAll(cleanPhone, "(", "")
	cleanPhone = strings.ReplaceAll(cleanPhone, ")", "")
	cleanPhone = strings.ReplaceAll(cleanPhone, "+", "")
	
	// Check if it's a valid length (10-15 digits)
	return len(cleanPhone) >= 10 && len(cleanPhone) <= 15
}

// SanitizeString removes harmful characters from string
func SanitizeString(input string) string {
	// Basic sanitization - remove potential SQL injection characters
	harmful := []string{"'", "\"", ";", "--", "/*", "*/", "xp_", "sp_"}
	result := input
	for _, char := range harmful {
		result = strings.ReplaceAll(result, char, "")
	}
	return strings.TrimSpace(result)
}

// IsValidOrderStatus checks if order status is valid
func IsValidOrderStatus(status string) bool {
	validStatuses := []string{"pending", "confirmed", "processing", "completed", "cancelled", "refunded"}
	for _, validStatus := range validStatuses {
		if status == validStatus {
			return true
		}
	}
	return false
}

// IsValidPaymentMethod checks if payment method is valid
func IsValidPaymentMethod(method string) bool {
	validMethods := []string{"cash", "card", "debit_card", "credit_card", "digital_wallet", "check", "bank_transfer"}
	for _, validMethod := range validMethods {
		if method == validMethod {
			return true
		}
	}
	return false
}

// FormatDateRange formats date range for queries
func FormatDateRange(startDate, endDate string) (time.Time, time.Time, error) {
	start, err := time.Parse("2006-01-02", startDate)
	if err != nil {
		return time.Time{}, time.Time{}, err
	}
	
	end, err := time.Parse("2006-01-02", endDate)
	if err != nil {
		return time.Time{}, time.Time{}, err
	}
	
	// Set end date to end of day
	end = end.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
	
	return start, end, nil
}