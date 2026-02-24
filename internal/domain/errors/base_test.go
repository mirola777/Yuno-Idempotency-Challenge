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

func TestNewCreatesCorrectError(t *testing.T) {
	appErr := New("TEST_CODE", http.StatusNotFound, "test message")

	assert.Equal(t, "TEST_CODE", appErr.Code)
	assert.Equal(t, "test message", appErr.Message)
	assert.Equal(t, http.StatusNotFound, appErr.HTTPCode)
}

func TestErrorReturnsFormattedString(t *testing.T) {
	appErr := New("TEST_CODE", http.StatusBadRequest, "test message")

	assert.Equal(t, "TEST_CODE: test message", appErr.Error())
}
