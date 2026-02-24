package handlers

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/mirola777/Yuno-Idempotency-Challenge/internal/domain"
)

type PaymentHandler struct {
	service domain.PaymentService
}

func NewPaymentHandler(service domain.PaymentService) *PaymentHandler {
	return &PaymentHandler{service: service}
}

func (h *PaymentHandler) CreatePayment(c echo.Context) error {
	idempotencyKey := c.Request().Header.Get("X-Idempotency-Key")

	var req domain.PaymentRequest
	if err := c.Bind(&req); err != nil {
		return domain.ErrInvalidPaymentRequest([]string{"invalid request body"})
	}

	payment, err := h.service.CreatePayment(c.Request().Context(), idempotencyKey, req)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusCreated, payment)
}

func (h *PaymentHandler) GetPayment(c echo.Context) error {
	id := c.Param("id")

	payment, err := h.service.GetPayment(c.Request().Context(), id)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, payment)
}

func (h *PaymentHandler) GetByIdempotencyKey(c echo.Context) error {
	key := c.Param("key")

	record, err := h.service.GetByIdempotencyKey(c.Request().Context(), key)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, record)
}
