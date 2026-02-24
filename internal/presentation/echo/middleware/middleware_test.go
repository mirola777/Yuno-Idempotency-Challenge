package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func TestTraceID_UsesProvidedHeader(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Trace-Id", "my-trace-123")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := TraceID(func(c echo.Context) error {
		return c.NoContent(http.StatusOK)
	})

	err := handler(c)

	assert.NoError(t, err)
	assert.Equal(t, "my-trace-123", rec.Header().Get("X-Trace-Id"))
	assert.Equal(t, "my-trace-123", c.Get("trace_id"))
}

func TestTraceID_GeneratesWhenMissing(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := TraceID(func(c echo.Context) error {
		return c.NoContent(http.StatusOK)
	})

	err := handler(c)

	assert.NoError(t, err)
	traceID := rec.Header().Get("X-Trace-Id")
	assert.NotEmpty(t, traceID)
	assert.Len(t, traceID, 36)
}

func TestTraceID_AddsToResponseHeader(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Trace-Id", "resp-trace-456")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := TraceID(func(c echo.Context) error {
		return c.NoContent(http.StatusOK)
	})

	err := handler(c)

	assert.NoError(t, err)
	assert.Equal(t, "resp-trace-456", rec.Header().Get("X-Trace-Id"))
}

func TestRequestLogger_CallsNext(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	called := false
	handler := RequestLogger(func(c echo.Context) error {
		called = true
		return c.NoContent(http.StatusOK)
	})

	err := handler(c)

	assert.NoError(t, err)
	assert.True(t, called)
}

func TestRecovery_CatchesPanicReturns500(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/panic", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := Recovery(func(c echo.Context) error {
		panic("something went wrong")
	})

	assert.NotPanics(t, func() {
		_ = handler(c)
	})
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}
