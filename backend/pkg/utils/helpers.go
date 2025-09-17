package utils

import (
	"fmt"
	"time"
)

// GenerateOrderNumber generates a unique order number
func GenerateOrderNumber() string {
	now := time.Now()
	return fmt.Sprintf("ORD-%d%02d%02d-%d",
		now.Year(), now.Month(), now.Day(), now.Unix()%10000)
}

// GenerateSKU generates a SKU based on category and timestamp
func GenerateSKU(categoryCode string) string {
	now := time.Now()
	return fmt.Sprintf("%s-%d%02d%02d-%d",
		categoryCode, now.Year(), now.Month(), now.Day(), now.Unix()%10000)
}

// CalculatePercentage calculates percentage of a value
func CalculatePercentage(value, percentage float64) float64 {
	return (value * percentage) / 100
}

// FormatCurrency formats a float64 value as currency
func FormatCurrency(amount float64) string {
	return fmt.Sprintf("$%.2f", amount)
}

// GenerateSerialNumber generates a serial number with a given prefix
func GenerateSerialNumber(prefix string) string {
	// 10 characters random alphanumeric string
	now := time.Now()
	return fmt.Sprintf("%s-%d%02d%02d-%d",
		prefix, now.Year(), now.Month(), now.Day(), now.Unix()%10000)
}
