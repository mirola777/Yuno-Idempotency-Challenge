package middleware

import (
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

func TraceID(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		traceID := c.Request().Header.Get("X-Trace-Id")
		if traceID == "" {
			traceID = uuid.New().String()
		}
		c.Response().Header().Set("X-Trace-Id", traceID)
		c.Set("trace_id", traceID)
		return next(c)
	}
}

func RequestLogger(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		start := time.Now()
		err := next(c)
		duration := time.Since(start)
		log.Printf("[%s] %s %s %d %s",
			c.Get("trace_id"),
			c.Request().Method,
			c.Request().URL.Path,
			c.Response().Status,
			duration,
		)
		return err
	}
}

func Recovery(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("PANIC recovered: %v", r)
				_ = c.JSON(http.StatusInternalServerError, map[string]interface{}{
					"code":     "INTERNAL_SERVER_ERROR",
					"messages": []string{"an unexpected error occurred"},
				})
			}
		}()
		return next(c)
	}
}
