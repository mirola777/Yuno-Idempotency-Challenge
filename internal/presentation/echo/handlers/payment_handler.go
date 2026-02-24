package handlers

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/mirola777/Yuno-Idempotency-Challenge/internal/application/use_cases"
	"github.com/mirola777/Yuno-Idempotency-Challenge/internal/domain"
	apperrors "github.com/mirola777/Yuno-Idempotency-Challenge/internal/domain/errors"
)

type PaymentHandler struct {
	createPayment       *use_cases.CreatePaymentUseCase
	getPayment          *use_cases.GetPaymentUseCase
	getByIdempotencyKey *use_cases.GetByIdempotencyKeyUseCase
}

func NewPaymentHandler(container *use_cases.Container) *PaymentHandler {
	return &PaymentHandler{
		createPayment:       container.CreatePayment,
		getPayment:          container.GetPayment,
		getByIdempotencyKey: container.GetByIdempotencyKey,
	}
}

func (h *PaymentHandler) CreatePayment(c echo.Context) error {
	idempotencyKey := c.Request().Header.Get("X-Idempotency-Key")

	var req domain.PaymentRequest
	if err := c.Bind(&req); err != nil {
		return apperrors.ErrInvalidPaymentRequest("invalid request body")
	}

	payment, err := h.createPayment.Execute(c.Request().Context(), idempotencyKey, req)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusCreated, payment)
}

func (h *PaymentHandler) GetPayment(c echo.Context) error {
	id := c.Param("id")

	payment, err := h.getPayment.Execute(c.Request().Context(), id)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, payment)
}

func (h *PaymentHandler) GetByIdempotencyKey(c echo.Context) error {
	key := c.Param("key")

	record, err := h.getByIdempotencyKey.Execute(c.Request().Context(), key)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, record)
}
