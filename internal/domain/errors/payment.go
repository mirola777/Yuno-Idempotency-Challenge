package errors

import (
	"fmt"
	"net/http"
)

func ErrIdempotencyKeyMissing() *AppError {
	return newAppError("IDEMPOTENCY_KEY_MISSING", http.StatusBadRequest, Messages{
		"en": "X-Idempotency-Key header is required",
		"es": "el encabezado X-Idempotency-Key es obligatorio",
	})
}

func ErrIdempotencyKeyTooLong() *AppError {
	return newAppError("IDEMPOTENCY_KEY_TOO_LONG", http.StatusBadRequest, Messages{
		"en": "X-Idempotency-Key must be at most 64 characters",
		"es": "X-Idempotency-Key debe tener como maximo 64 caracteres",
	})
}

func ErrIdempotencyKeyConflict() *AppError {
	return newAppError("IDEMPOTENCY_KEY_CONFLICT", http.StatusConflict, Messages{
		"en": "idempotency key already used with different request payload",
		"es": "la clave de idempotencia ya fue utilizada con un payload diferente",
	})
}

func ErrPaymentProcessing() *AppError {
	return newAppError("PAYMENT_PROCESSING", http.StatusConflict, Messages{
		"en": "a payment with this idempotency key is currently being processed",
		"es": "un pago con esta clave de idempotencia esta siendo procesado actualmente",
	})
}

func ErrPaymentNotFound() *AppError {
	return newAppError("PAYMENT_NOT_FOUND", http.StatusNotFound, Messages{
		"en": "payment not found",
		"es": "pago no encontrado",
	})
}

func ErrIdempotencyKeyNotFound() *AppError {
	return newAppError("IDEMPOTENCY_KEY_NOT_FOUND", http.StatusNotFound, Messages{
		"en": "idempotency key not found",
		"es": "clave de idempotencia no encontrada",
	})
}

func ErrInvalidPaymentRequest(detail string) *AppError {
	return newAppError("INVALID_PAYMENT_REQUEST", http.StatusBadRequest, Messages{
		"en": fmt.Sprintf("invalid payment request: %s", detail),
		"es": fmt.Sprintf("solicitud de pago invalida: %s", detail),
	})
}

func ErrInvalidCurrency(currency string) *AppError {
	return newAppError("INVALID_CURRENCY", http.StatusBadRequest, Messages{
		"en": fmt.Sprintf("currency is not supported; valid currencies: IDR, THB, VND, PHP: %s", currency),
		"es": fmt.Sprintf("moneda no soportada; monedas validas: IDR, THB, VND, PHP: %s", currency),
	})
}

func ErrInternal() *AppError {
	return newAppError("INTERNAL_ERROR", http.StatusInternalServerError, Messages{
		"en": "an internal error occurred",
		"es": "ocurrio un error interno",
	})
}
