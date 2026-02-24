package errors

import "strings"

var messages = map[string]map[string]string{
	"en": {
		"IDEMPOTENCY_KEY_MISSING":   "X-Idempotency-Key header is required",
		"IDEMPOTENCY_KEY_TOO_LONG":  "X-Idempotency-Key must be at most 64 characters",
		"IDEMPOTENCY_KEY_CONFLICT":  "idempotency key already used with different request payload",
		"PAYMENT_PROCESSING":        "a payment with this idempotency key is currently being processed",
		"PAYMENT_NOT_FOUND":         "payment not found",
		"IDEMPOTENCY_KEY_NOT_FOUND": "idempotency key not found",
		"INVALID_PAYMENT_REQUEST":   "invalid payment request",
		"INVALID_CURRENCY":          "currency is not supported; valid currencies: IDR, THB, VND, PHP",
		"INTERNAL_ERROR":            "an internal error occurred",
	},
	"es": {
		"IDEMPOTENCY_KEY_MISSING":   "el encabezado X-Idempotency-Key es obligatorio",
		"IDEMPOTENCY_KEY_TOO_LONG":  "X-Idempotency-Key debe tener como maximo 64 caracteres",
		"IDEMPOTENCY_KEY_CONFLICT":  "la clave de idempotencia ya fue utilizada con un payload diferente",
		"PAYMENT_PROCESSING":        "un pago con esta clave de idempotencia esta siendo procesado actualmente",
		"PAYMENT_NOT_FOUND":         "pago no encontrado",
		"IDEMPOTENCY_KEY_NOT_FOUND": "clave de idempotencia no encontrada",
		"INVALID_PAYMENT_REQUEST":   "solicitud de pago invalida",
		"INVALID_CURRENCY":          "moneda no soportada; monedas validas: IDR, THB, VND, PHP",
		"INTERNAL_ERROR":            "ocurrio un error interno",
	},
}

func GetMessage(code string, lang string) string {
	base := strings.SplitN(lang, "-", 2)[0]
	base = strings.TrimSpace(strings.ToLower(base))

	if langMessages, ok := messages[base]; ok {
		if msg, ok := langMessages[code]; ok {
			return msg
		}
	}

	if base != "en" {
		if enMessages, ok := messages["en"]; ok {
			if msg, ok := enMessages[code]; ok {
				return msg
			}
		}
	}

	return code
}

func Localize(err *AppError, lang string) *AppError {
	return &AppError{
		Code:     err.Code,
		Message:  GetMessage(err.Code, lang),
		HTTPCode: err.HTTPCode,
	}
}
