package errors

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestErrIdempotencyKeyMissing(t *testing.T) {
	err := ErrIdempotencyKeyMissing()

	assert.Equal(t, "IDEMPOTENCY_KEY_MISSING", err.Code)
	assert.Equal(t, http.StatusBadRequest, err.HTTPCode)
	assert.Equal(t, "X-Idempotency-Key header is required", err.Message)
}

func TestErrIdempotencyKeyTooLong(t *testing.T) {
	err := ErrIdempotencyKeyTooLong()

	assert.Equal(t, "IDEMPOTENCY_KEY_TOO_LONG", err.Code)
	assert.Equal(t, http.StatusBadRequest, err.HTTPCode)
	assert.Equal(t, "X-Idempotency-Key must be at most 64 characters", err.Message)
}

func TestErrIdempotencyKeyConflict(t *testing.T) {
	err := ErrIdempotencyKeyConflict()

	assert.Equal(t, "IDEMPOTENCY_KEY_CONFLICT", err.Code)
	assert.Equal(t, http.StatusConflict, err.HTTPCode)
	assert.Equal(t, "idempotency key already used with different request payload", err.Message)
}

func TestErrPaymentProcessing(t *testing.T) {
	err := ErrPaymentProcessing()

	assert.Equal(t, "PAYMENT_PROCESSING", err.Code)
	assert.Equal(t, http.StatusConflict, err.HTTPCode)
	assert.Equal(t, "a payment with this idempotency key is currently being processed", err.Message)
}

func TestErrPaymentNotFound(t *testing.T) {
	err := ErrPaymentNotFound()

	assert.Equal(t, "PAYMENT_NOT_FOUND", err.Code)
	assert.Equal(t, http.StatusNotFound, err.HTTPCode)
	assert.Equal(t, "payment not found", err.Message)
}

func TestErrIdempotencyKeyNotFound(t *testing.T) {
	err := ErrIdempotencyKeyNotFound()

	assert.Equal(t, "IDEMPOTENCY_KEY_NOT_FOUND", err.Code)
	assert.Equal(t, http.StatusNotFound, err.HTTPCode)
	assert.Equal(t, "idempotency key not found", err.Message)
}

func TestErrInvalidPaymentRequestIncludesDetail(t *testing.T) {
	err := ErrInvalidPaymentRequest("amount must be positive")

	assert.Equal(t, "INVALID_PAYMENT_REQUEST", err.Code)
	assert.Equal(t, http.StatusBadRequest, err.HTTPCode)
	assert.Equal(t, "invalid payment request: amount must be positive", err.Message)
}

func TestErrInvalidCurrencyIncludesCurrencyName(t *testing.T) {
	err := ErrInvalidCurrency("USD")

	assert.Equal(t, "INVALID_CURRENCY", err.Code)
	assert.Equal(t, http.StatusBadRequest, err.HTTPCode)
	assert.Equal(t, "currency is not supported; valid currencies: IDR, THB, VND, PHP: USD", err.Message)
}

func TestErrInternal(t *testing.T) {
	err := ErrInternal()

	assert.Equal(t, "INTERNAL_ERROR", err.Code)
	assert.Equal(t, http.StatusInternalServerError, err.HTTPCode)
	assert.Equal(t, "an internal error occurred", err.Message)
}
