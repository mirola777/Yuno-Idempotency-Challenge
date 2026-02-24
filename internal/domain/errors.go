package domain

import (
	"fmt"
	"net/http"
	"strings"
)

type AppError struct {
	Code     string   `json:"code"`
	Messages []string `json:"messages"`
	HTTPCode int      `json:"-"`
}

func (e *AppError) Error() string {
	return fmt.Sprintf("%s: %s", e.Code, strings.Join(e.Messages, "; "))
}

func ErrIdempotencyKeyMissing() *AppError {
	return &AppError{
		Code:     "IDEMPOTENCY_KEY_MISSING",
		Messages: []string{"X-Idempotency-Key header is required"},
		HTTPCode: http.StatusBadRequest,
	}
}

func ErrIdempotencyKeyTooLong() *AppError {
	return &AppError{
		Code:     "IDEMPOTENCY_KEY_TOO_LONG",
		Messages: []string{"X-Idempotency-Key must be at most 64 characters"},
		HTTPCode: http.StatusBadRequest,
	}
}

func ErrIdempotencyKeyConflict(key string) *AppError {
	return &AppError{
		Code:     "IDEMPOTENCY_KEY_CONFLICT",
		Messages: []string{fmt.Sprintf("idempotency key '%s' already used with different request payload", key)},
		HTTPCode: http.StatusConflict,
	}
}

func ErrPaymentProcessing() *AppError {
	return &AppError{
		Code:     "PAYMENT_PROCESSING",
		Messages: []string{"a payment with this idempotency key is currently being processed"},
		HTTPCode: http.StatusConflict,
	}
}

func ErrPaymentNotFound(id string) *AppError {
	return &AppError{
		Code:     "PAYMENT_NOT_FOUND",
		Messages: []string{fmt.Sprintf("payment '%s' not found", id)},
		HTTPCode: http.StatusNotFound,
	}
}

func ErrIdempotencyKeyNotFound(key string) *AppError {
	return &AppError{
		Code:     "IDEMPOTENCY_KEY_NOT_FOUND",
		Messages: []string{fmt.Sprintf("idempotency key '%s' not found", key)},
		HTTPCode: http.StatusNotFound,
	}
}

func ErrInvalidPaymentRequest(reasons []string) *AppError {
	return &AppError{
		Code:     "INVALID_PAYMENT_REQUEST",
		Messages: reasons,
		HTTPCode: http.StatusBadRequest,
	}
}

func ErrInvalidCurrency(currency string) *AppError {
	return &AppError{
		Code:     "INVALID_CURRENCY",
		Messages: []string{fmt.Sprintf("currency '%s' is not supported; valid currencies: IDR, THB, VND, PHP", currency)},
		HTTPCode: http.StatusBadRequest,
	}
}

func ErrInternal(msg string) *AppError {
	return &AppError{
		Code:     "INTERNAL_ERROR",
		Messages: []string{msg},
		HTTPCode: http.StatusInternalServerError,
	}
}
