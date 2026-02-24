package errors

import (
	"fmt"
	"strings"
)

type Messages map[string]string

type AppError struct {
	Code     string   `json:"code"`
	Message  string   `json:"message"`
	HTTPCode int      `json:"-"`
	messages Messages `json:"-"`
}

func (e *AppError) Error() string {
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func (e *AppError) Localize(lang string) *AppError {
	base := strings.SplitN(lang, "-", 2)[0]
	base = strings.TrimSpace(strings.ToLower(base))

	if msg, ok := e.messages[base]; ok {
		return &AppError{
			Code:     e.Code,
			Message:  msg,
			HTTPCode: e.HTTPCode,
			messages: e.messages,
		}
	}

	return e
}

func newAppError(code string, httpCode int, msgs Messages) *AppError {
	return &AppError{
		Code:     code,
		Message:  msgs["en"],
		HTTPCode: httpCode,
		messages: msgs,
	}
}
