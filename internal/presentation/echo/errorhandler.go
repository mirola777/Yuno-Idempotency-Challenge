package echo

import (
	"net/http"
	"strings"

	echofw "github.com/labstack/echo/v4"
	apperrors "github.com/mirola777/Yuno-Idempotency-Challenge/internal/domain/errors"
)

func CustomHTTPErrorHandler(err error, c echofw.Context) {
	if c.Response().Committed {
		return
	}

	lang := parseAcceptLanguage(c.Request().Header.Get("Accept-Language"))

	if appErr, ok := err.(*apperrors.AppError); ok {
		localized := apperrors.Localize(appErr, lang)
		_ = c.JSON(localized.HTTPCode, map[string]interface{}{
			"code":    localized.Code,
			"message": localized.Message,
		})
		return
	}

	if echoErr, ok := err.(*echofw.HTTPError); ok {
		_ = c.JSON(echoErr.Code, map[string]interface{}{
			"code":    "HTTP_ERROR",
			"message": http.StatusText(echoErr.Code),
		})
		return
	}

	msg := apperrors.GetMessage("INTERNAL_ERROR", lang)
	_ = c.JSON(http.StatusInternalServerError, map[string]interface{}{
		"code":    "INTERNAL_ERROR",
		"message": msg,
	})
}

func parseAcceptLanguage(header string) string {
	if header == "" {
		return "en"
	}
	parts := strings.Split(header, ",")
	if len(parts) == 0 {
		return "en"
	}
	lang := strings.TrimSpace(parts[0])
	lang = strings.Split(lang, ";")[0]
	return lang
}
