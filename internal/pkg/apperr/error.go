package apperr

import "fmt"

// AppError represents an application error with code and message
type AppError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Detail  string `json:"detail,omitempty"`
	Cause   error  `json:"-"`
}

func (e *AppError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("[%d] %s: %v", e.Code, e.Message, e.Cause)
	}
	return fmt.Sprintf("[%d] %s", e.Code, e.Message)
}

func (e *AppError) Unwrap() error {
	return e.Cause
}

// New creates a new AppError
func New(code int, message string) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
	}
}

// Wrap wraps an error with AppError
func Wrap(code int, message string, cause error) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
		Cause:   cause,
	}
}

// WithDetail adds detail to AppError
func (e *AppError) WithDetail(detail string) *AppError {
	e.Detail = detail
	return e
}

// BadRequest creates a bad request error
func BadRequest(message string) *AppError {
	return New(CodeInvalidParam, message)
}

// NotFound creates a not found error
func NotFound(resource, field, value string) *AppError {
	return New(CodeNotFound, fmt.Sprintf("%s not found: %s=%s", resource, field, value))
}

