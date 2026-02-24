package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/mirola777/Yuno-Idempotency-Challenge/internal/application/use_cases"
	"github.com/mirola777/Yuno-Idempotency-Challenge/internal/domain"
	apperrors "github.com/mirola777/Yuno-Idempotency-Challenge/internal/domain/errors"
	"github.com/stretchr/testify/assert"
)

func samplePayment() *domain.Payment {
	return &domain.Payment{
		ID:         "pay-123",
		Amount:     100.0,
		Currency:   domain.CurrencyIDR,
		CustomerID: "cust-1",
		RideID:     "ride-1",
		Status:     domain.PaymentStatusSucceeded,
		CardLast4:  "1111",
		CreatedAt:  time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
	}
}

func TestCreatePayment_InvalidJSON_Returns400(t *testing.T) {
	container := &use_cases.Container{}
	h := NewPaymentHandler(container)

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/payments", strings.NewReader("not-json"))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req.Header.Set("X-Idempotency-Key", "key-xyz")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.CreatePayment(c)

	appErr, ok := err.(*apperrors.AppError)
	assert.True(t, ok)
	assert.Equal(t, http.StatusBadRequest, appErr.HTTPCode)
}

func TestGetPayment_ValidID_Returns200(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/payments/pay-123", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("pay-123")

	payment := samplePayment()
	handler := func(c echo.Context) error {
		return c.JSON(http.StatusOK, payment)
	}

	err := handler(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var resp domain.Payment
	assert.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.Equal(t, "pay-123", resp.ID)
}

func TestGetByIdempotencyKey_Response(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/idempotency/key-abc", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("key")
	c.SetParamValues("key-abc")

	record := &domain.IdempotencyRecord{
		Key:                "key-abc",
		RequestFingerprint: "abc123",
		PaymentID:          "pay-123",
		Status:             domain.IdempotencyStatusCompleted,
		CreatedAt:          time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
		ExpiresAt:          time.Date(2025, 1, 2, 0, 0, 0, 0, time.UTC),
	}

	handler := func(c echo.Context) error {
		return c.JSON(http.StatusOK, record)
	}

	err := handler(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var resp domain.IdempotencyRecord
	assert.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.Equal(t, "key-abc", resp.Key)
}

func TestHealthCheck_Returns200(t *testing.T) {
	h := NewHealthHandler()

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.Check(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
}
