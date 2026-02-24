package errors

import (
	"fmt"
	"net/http"
)

func ErrIdempotencyKeyMissing() *AppError {
	return New("IDEMPOTENCY_KEY_MISSING", http.StatusBadRequest, messages["en"]["IDEMPOTENCY_KEY_MISSING"])
}

func ErrIdempotencyKeyTooLong() *AppError {
	return New("IDEMPOTENCY_KEY_TOO_LONG", http.StatusBadRequest, messages["en"]["IDEMPOTENCY_KEY_TOO_LONG"])
}

func ErrIdempotencyKeyConflict() *AppError {
	return New("IDEMPOTENCY_KEY_CONFLICT", http.StatusConflict, messages["en"]["IDEMPOTENCY_KEY_CONFLICT"])
}

func ErrPaymentProcessing() *AppError {
	return New("PAYMENT_PROCESSING", http.StatusConflict, messages["en"]["PAYMENT_PROCESSING"])
}

func ErrPaymentNotFound() *AppError {
	return New("PAYMENT_NOT_FOUND", http.StatusNotFound, messages["en"]["PAYMENT_NOT_FOUND"])
}

func ErrIdempotencyKeyNotFound() *AppError {
	return New("IDEMPOTENCY_KEY_NOT_FOUND", http.StatusNotFound, messages["en"]["IDEMPOTENCY_KEY_NOT_FOUND"])
}

func ErrInvalidPaymentRequest(detail string) *AppError {
	return New("INVALID_PAYMENT_REQUEST", http.StatusBadRequest, fmt.Sprintf("%s: %s", messages["en"]["INVALID_PAYMENT_REQUEST"], detail))
}

func ErrInvalidCurrency(currency string) *AppError {
	return New("INVALID_CURRENCY", http.StatusBadRequest, fmt.Sprintf("%s: %s", messages["en"]["INVALID_CURRENCY"], currency))
}

func ErrInternal() *AppError {
	return New("INTERNAL_ERROR", http.StatusInternalServerError, messages["en"]["INTERNAL_ERROR"])
}
