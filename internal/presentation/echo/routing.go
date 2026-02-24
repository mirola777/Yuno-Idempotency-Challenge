package echo

import (
	echofw "github.com/labstack/echo/v4"
	"github.com/mirola777/Yuno-Idempotency-Challenge/internal/domain"
	"github.com/mirola777/Yuno-Idempotency-Challenge/internal/presentation/echo/handlers"
	"github.com/mirola777/Yuno-Idempotency-Challenge/internal/presentation/echo/middleware"
)

func ConfigureRoutes(e *echofw.Echo, paymentService domain.PaymentService) {
	e.Use(middleware.Recovery)
	e.Use(middleware.TraceID)
	e.Use(middleware.RequestLogger)

	healthHandler := handlers.NewHealthHandler()
	e.GET("/health", healthHandler.Check)

	paymentHandler := handlers.NewPaymentHandler(paymentService)
	v1 := e.Group("/v1")
	v1.POST("/payments", paymentHandler.CreatePayment)
	v1.GET("/payments/:id", paymentHandler.GetPayment)
	v1.GET("/idempotency/:key", paymentHandler.GetByIdempotencyKey)
}
