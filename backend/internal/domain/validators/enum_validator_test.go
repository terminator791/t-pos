package validators

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/terminator791/t-pos/internal/domain/entities"
)

func TestValidateExpenseStatus_ValidStatuses(t *testing.T) {
	validStatuses := []entities.ExpenseStatus{
		entities.ExpenseStatusPending,
		entities.ExpenseStatusCompleted,
		entities.ExpenseStatusFailed,
		entities.ExpenseStatusCancelled,
	}

	for _, status := range validStatuses {
		t.Run(string(status), func(t *testing.T) {
			err := ValidateExpenseStatus(status)
			assert.NoError(t, err)
		})
	}
}

func TestValidateExpenseStatus_InvalidStatus(t *testing.T) {
	err := ValidateExpenseStatus("invalid_status")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid expense status")
	assert.Contains(t, err.Error(), "pending, completed, failed, cancelled")
}

func TestValidatePaymentStatus_ValidStatuses(t *testing.T) {
	validStatuses := []entities.PaymentStatus{
		entities.PaymentStatusPending,
		entities.PaymentStatusCompleted,
		entities.PaymentStatusFailed,
		entities.PaymentStatusCancelled,
	}

	for _, status := range validStatuses {
		t.Run(string(status), func(t *testing.T) {
			err := ValidatePaymentStatus(status)
			assert.NoError(t, err)
		})
	}
}

func TestValidatePaymentStatus_InvalidStatus(t *testing.T) {
	err := ValidatePaymentStatus("invalid_status")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid payment status")
	assert.Contains(t, err.Error(), "pending, completed, failed, cancelled")
}

func TestValidateTransactionStatus_ValidStatuses(t *testing.T) {
	validStatuses := []entities.TransactionStatus{
		entities.TransactionStatusPending,
		entities.TransactionStatusCompleted,
		entities.TransactionStatusCancelled,
		entities.TransactionStatusFailed,
	}

	for _, status := range validStatuses {
		t.Run(string(status), func(t *testing.T) {
			err := ValidateTransactionStatus(status)
			assert.NoError(t, err)
		})
	}
}

func TestValidateTransactionStatus_InvalidStatus(t *testing.T) {
	err := ValidateTransactionStatus("invalid_status")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid transaction status")
	assert.Contains(t, err.Error(), "pending, completed, cancelled, failed")
}
