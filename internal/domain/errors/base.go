package errors

import "fmt"

type AppError struct {
	Code     string `json:"code"`
	Message  string `json:"message"`
	HTTPCode int    `json:"-"`
}

func (e *AppError) Error() string {
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func New(code string, httpCode int, message string) *AppError {
	return &AppError{
		Code:     code,
		Message:  message,
		HTTPCode: httpCode,
	}
}
