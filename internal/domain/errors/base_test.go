package errors

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAppErrorImplementsErrorInterface(t *testing.T) {
	var err error = &AppError{
		Code:     "TEST_CODE",
		Message:  "test message",
		HTTPCode: http.StatusBadRequest,
	}

	assert.NotNil(t, err)
	assert.Implements(t, (*error)(nil), &AppError{})
}

func TestNewAppErrorDefaultsToEnglish(t *testing.T) {
	appErr := newAppError("TEST_CODE", http.StatusNotFound, Messages{
		"en": "not found",
		"es": "no encontrado",
	})

	assert.Equal(t, "TEST_CODE", appErr.Code)
	assert.Equal(t, "not found", appErr.Message)
	assert.Equal(t, http.StatusNotFound, appErr.HTTPCode)
}

func TestErrorReturnsFormattedString(t *testing.T) {
	appErr := newAppError("TEST_CODE", http.StatusBadRequest, Messages{
		"en": "test message",
	})

	assert.Equal(t, "TEST_CODE: test message", appErr.Error())
}

func TestLocalize_ReturnsSpanish(t *testing.T) {
	appErr := newAppError("TEST_CODE", http.StatusNotFound, Messages{
		"en": "not found",
		"es": "no encontrado",
	})

	localized := appErr.Localize("es")
	assert.Equal(t, "no encontrado", localized.Message)
	assert.Equal(t, "TEST_CODE", localized.Code)
	assert.Equal(t, http.StatusNotFound, localized.HTTPCode)
}

func TestLocalize_FallsBackToEnglishForUnknownLanguage(t *testing.T) {
	appErr := newAppError("TEST_CODE", http.StatusNotFound, Messages{
		"en": "not found",
		"es": "no encontrado",
	})

	localized := appErr.Localize("fr")
	assert.Equal(t, "not found", localized.Message)
}

func TestLocalize_ExtractsBaseLanguageFromLocale(t *testing.T) {
	appErr := newAppError("TEST_CODE", http.StatusNotFound, Messages{
		"en": "not found",
		"es": "no encontrado",
	})

	localized := appErr.Localize("es-CO")
	assert.Equal(t, "no encontrado", localized.Message)
}

func TestLocalize_EnglishByDefault(t *testing.T) {
	appErr := newAppError("TEST_CODE", http.StatusNotFound, Messages{
		"en": "not found",
	})

	localized := appErr.Localize("en")
	assert.Equal(t, "not found", localized.Message)
}

func TestLocalize_PreservesOriginal(t *testing.T) {
	original := newAppError("TEST_CODE", http.StatusNotFound, Messages{
		"en": "not found",
		"es": "no encontrado",
	})

	localized := original.Localize("es")
	assert.Equal(t, "no encontrado", localized.Message)
	assert.Equal(t, "not found", original.Message)
}
