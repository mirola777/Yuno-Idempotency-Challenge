package errors

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetMessageReturnsEnglishMessage(t *testing.T) {
	msg := GetMessage("IDEMPOTENCY_KEY_MISSING", "en")

	assert.Equal(t, "X-Idempotency-Key header is required", msg)
}

func TestGetMessageReturnsSpanishMessage(t *testing.T) {
	msg := GetMessage("IDEMPOTENCY_KEY_MISSING", "es")

	assert.Equal(t, "el encabezado X-Idempotency-Key es obligatorio", msg)
}

func TestGetMessageFallsBackToEnglishForUnknownLanguage(t *testing.T) {
	msg := GetMessage("IDEMPOTENCY_KEY_MISSING", "fr")

	assert.Equal(t, "X-Idempotency-Key header is required", msg)
}

func TestGetMessageExtractsBaseLanguageFromLocale(t *testing.T) {
	msg := GetMessage("PAYMENT_NOT_FOUND", "es-CO")

	assert.Equal(t, "pago no encontrado", msg)
}

func TestGetMessageReturnsCodeForUnknownCode(t *testing.T) {
	msg := GetMessage("UNKNOWN_ERROR_CODE", "en")

	assert.Equal(t, "UNKNOWN_ERROR_CODE", msg)
}

func TestLocalizeReturnsCopyWithLocalizedMessage(t *testing.T) {
	original := New("PAYMENT_NOT_FOUND", http.StatusNotFound, messages["en"]["PAYMENT_NOT_FOUND"])

	localized := Localize(original, "es")

	assert.Equal(t, "PAYMENT_NOT_FOUND", localized.Code)
	assert.Equal(t, http.StatusNotFound, localized.HTTPCode)
	assert.Equal(t, "pago no encontrado", localized.Message)
	assert.Equal(t, "payment not found", original.Message)
}
