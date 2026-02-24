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

func TestErrIdempotencyKeyMissing_Localize(t *testing.T) {
	err := ErrIdempotencyKeyMissing().Localize("es")

	assert.Equal(t, "el encabezado X-Idempotency-Key es obligatorio", err.Message)
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

func TestErrIdempotencyKeyConflict_Localize(t *testing.T) {
	err := ErrIdempotencyKeyConflict().Localize("es")

	assert.Equal(t, "la clave de idempotencia ya fue utilizada con un payload diferente", err.Message)
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

func TestErrPaymentNotFound_Localize(t *testing.T) {
	err := ErrPaymentNotFound().Localize("es")

	assert.Equal(t, "pago no encontrado", err.Message)
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

func TestErrInvalidPaymentRequest_Localize(t *testing.T) {
	err := ErrInvalidPaymentRequest("amount must be positive").Localize("es")

	assert.Contains(t, err.Message, "solicitud de pago invalida")
}

func TestErrInvalidCurrencyIncludesCurrencyName(t *testing.T) {
	err := ErrInvalidCurrency("USD")

	assert.Equal(t, "INVALID_CURRENCY", err.Code)
	assert.Equal(t, http.StatusBadRequest, err.HTTPCode)
	assert.Contains(t, err.Message, "USD")
}

func TestErrInternal(t *testing.T) {
	err := ErrInternal()

	assert.Equal(t, "INTERNAL_ERROR", err.Code)
	assert.Equal(t, http.StatusInternalServerError, err.HTTPCode)
	assert.Equal(t, "an internal error occurred", err.Message)
}

func TestErrInternal_Localize(t *testing.T) {
	err := ErrInternal().Localize("es")

	assert.Equal(t, "ocurrio un error interno", err.Message)
}

func TestAllErrors_DefaultToEnglish(t *testing.T) {
	errors := []*AppError{
		ErrIdempotencyKeyMissing(),
		ErrIdempotencyKeyTooLong(),
		ErrIdempotencyKeyConflict(),
		ErrPaymentProcessing(),
		ErrPaymentNotFound(),
		ErrIdempotencyKeyNotFound(),
		ErrInvalidPaymentRequest("test"),
		ErrInvalidCurrency("USD"),
		ErrInternal(),
	}

	for _, err := range errors {
		assert.NotEmpty(t, err.Message, "error %s should have a default English message", err.Code)
		localized := err.Localize("en")
		assert.Equal(t, err.Message, localized.Message)
	}
}
