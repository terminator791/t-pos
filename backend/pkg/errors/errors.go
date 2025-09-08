package errors

import (
	"errors"
	"fmt"
)

// Common application errors
var (
	ErrNotFound          = errors.New("resource not found")
	ErrInvalidInput      = errors.New("invalid input")
	ErrUnauthorized      = errors.New("unauthorized")
	ErrForbidden         = errors.New("forbidden")
	ErrConflict          = errors.New("resource already exists")
	ErrInternalServer    = errors.New("internal server error")
	ErrBadRequest        = errors.New("bad request")
	ErrValidationFailed  = errors.New("validation failed")
	ErrDatabaseOperation = errors.New("database operation failed")
	ErrTokenInvalid      = errors.New("token is invalid")
	ErrTokenExpired      = errors.New("token has expired")
	ErrInsufficientStock = errors.New("insufficient stock")
	ErrOrderNotFound     = errors.New("order not found")
	ErrPaymentFailed     = errors.New("payment processing failed")
)

// AppError represents an application error with additional context
type AppError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
	Err     error  `json:"-"`
}

// Error implements the error interface
func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s", e.Message, e.Err.Error())
	}
	return e.Message
}

// Unwrap returns the underlying error
func (e *AppError) Unwrap() error {
	return e.Err
}

// NewAppError creates a new application error
func NewAppError(code, message string, err error) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
		Err:     err,
	}
}

// NewValidationError creates a validation error
func NewValidationError(message string) *AppError {
	return &AppError{
		Code:    "VALIDATION_ERROR",
		Message: message,
		Err:     ErrValidationFailed,
	}
}

// NewNotFoundError creates a not found error
func NewNotFoundError(resource string) *AppError {
	return &AppError{
		Code:    "NOT_FOUND",
		Message: fmt.Sprintf("%s not found", resource),
		Err:     ErrNotFound,
	}
}

// NewConflictError creates a conflict error
func NewConflictError(message string) *AppError {
	return &AppError{
		Code:    "CONFLICT",
		Message: message,
		Err:     ErrConflict,
	}
}

// NewUnauthorizedError creates an unauthorized error
func NewUnauthorizedError(message string) *AppError {
	return &AppError{
		Code:    "UNAUTHORIZED",
		Message: message,
		Err:     ErrUnauthorized,
	}
}

// NewInternalServerError creates an internal server error
func NewInternalServerError(message string, err error) *AppError {
	return &AppError{
		Code:    "INTERNAL_SERVER_ERROR",
		Message: message,
		Err:     err,
	}
}