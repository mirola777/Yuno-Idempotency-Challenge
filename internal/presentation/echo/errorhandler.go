package echo

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/mirola777/Yuno-Idempotency-Challenge/internal/domain"
)

func CustomHTTPErrorHandler(err error, c echo.Context) {
	if c.Response().Committed {
		return
	}

	if appErr, ok := err.(*domain.AppError); ok {
		_ = c.JSON(appErr.HTTPCode, map[string]interface{}{
			"code":     appErr.Code,
			"messages": appErr.Messages,
		})
		return
	}

	if echoErr, ok := err.(*echo.HTTPError); ok {
		_ = c.JSON(echoErr.Code, map[string]interface{}{
			"code":    "HTTP_ERROR",
			"messages": []string{http.StatusText(echoErr.Code)},
		})
		return
	}

	_ = c.JSON(http.StatusInternalServerError, map[string]interface{}{
		"code":     "INTERNAL_SERVER_ERROR",
		"messages": []string{"an unexpected error occurred"},
	})
}
