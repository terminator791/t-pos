package validators

import (
	"fmt"
	"github.com/terminator791/t-pos/internal/domain/entities"
)

// ValidateExpenseStatus validates if the expense status is valid
func ValidateExpenseStatus(status entities.ExpenseStatus) error {
	switch status {
	case entities.ExpenseStatusPending, entities.ExpenseStatusCompleted, entities.ExpenseStatusFailed, entities.ExpenseStatusCancelled:
		return nil
	default:
		return fmt.Errorf("invalid expense status: %s. Valid values are: pending, completed, failed, cancelled", status)
	}
}

// ValidatePaymentStatus validates if the payment status is valid
func ValidatePaymentStatus(status entities.PaymentStatus) error {
	switch status {
	case entities.PaymentStatusPending, entities.PaymentStatusCompleted, entities.PaymentStatusFailed, entities.PaymentStatusCancelled:
		return nil
	default:
		return fmt.Errorf("invalid payment status: %s. Valid values are: pending, completed, failed, cancelled", status)
	}
}

// ValidateTransactionStatus validates if the transaction status is valid
func ValidateTransactionStatus(status entities.TransactionStatus) error {
	switch status {
	case entities.TransactionStatusPending, entities.TransactionStatusCompleted, entities.TransactionStatusCancelled, entities.TransactionStatusFailed:
		return nil
	default:
		return fmt.Errorf("invalid transaction status: %s. Valid values are: pending, completed, cancelled, failed", status)
	}
}
