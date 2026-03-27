package errs

import (
	"errors"
	"net/http"
)

// Error codes
const (
	CodeUserNotFound        = "USER_NOT_FOUND"
	CodeUserNotRegistered   = "USER_NOT_REGISTERED"
	CodeInvalidCredentials  = "INVALID_CREDENTIALS"
	CodeAccountLocked      = "ACCOUNT_LOCKED"
	CodeEmailExists        = "EMAIL_EXISTS"
	CodeInvalidToken       = "INVALID_TOKEN"
	CodeTokenExpired       = "TOKEN_EXPIRED"
	CodeUnauthorized       = "UNAUTHORIZED"
	CodeForbidden          = "FORBIDDEN"
	CodeRateLimitExceeded  = "RATE_LIMIT_EXCEEDED"
	CodeInvalidInput       = "INVALID_INPUT"
	CodeInternalError      = "INTERNAL_ERROR"
	CodeBadRequest         = "BAD_REQUEST"
	CodeNotFound           = "NOT_FOUND"
)

var (
	ErrInternalError = errors.New("internal error")
	ErrNotFound      = errors.New("not found")
)

// Common application errors
var (
	ErrUserNotFound       = errors.New("user not found")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrAccountLocked      = errors.New("account is locked")
	ErrEmailExists        = errors.New("email already exists")
	ErrInvalidToken       = errors.New("invalid token")
	ErrTokenExpired       = errors.New("token expired")
	ErrUnauthorized       = errors.New("unauthorized")
	ErrForbidden          = errors.New("forbidden")
	ErrRateLimitExceeded  = errors.New("rate limit exceeded")
	ErrInvalidInput       = errors.New("invalid input")
)

// AppError represents an application error with HTTP status code and error code
type AppError struct {
	Err        error
	StatusCode int
	Message    string
	Code       string
	Details    map[string]interface{}
}

// Error implements error interface
func (e *AppError) Error() string {
	if e.Message != "" {
		return e.Message
	}
	return e.Err.Error()
}

// Unwrap returns the underlying error
func (e *AppError) Unwrap() error {
	return e.Err
}

// NewAppError creates a new AppError
func NewAppError(err error, statusCode int, message string) *AppError {
	return &AppError{
		Err:        err,
		StatusCode: statusCode,
		Message:    message,
		Code:       CodeInternalError,
		Details:    make(map[string]interface{}),
	}
}

// WithCode sets the error code
func (e *AppError) WithCode(code string) *AppError {
	e.Code = code
	return e
}

// WithHTTPStatus sets the HTTP status code
func (e *AppError) WithHTTPStatus(statusCode int) *AppError {
	e.StatusCode = statusCode
	return e
}

// WithDetails adds details to the error
func (e *AppError) WithDetails(details map[string]interface{}) *AppError {
	if e.Details == nil {
		e.Details = make(map[string]interface{})
	}
	for k, v := range details {
		e.Details[k] = v
	}
	return e
}

// Is checks if an error matches a specific error code
func Is(err error, code string) bool {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr.Code == code
	}
	return false
}

// ToHTTPStatus maps common errors to HTTP status codes
func ToHTTPStatus(err error) int {
	if err == nil {
		return http.StatusOK
	}

	var appErr *AppError
	if errors.As(err, &appErr) {
		if appErr.StatusCode != 0 {
			return appErr.StatusCode
		}
	}

	switch err {
	case ErrUserNotFound:
		return http.StatusNotFound
	case ErrInvalidCredentials:
		return http.StatusUnauthorized
	case ErrAccountLocked:
		return http.StatusLocked
	case ErrEmailExists:
		return http.StatusConflict
	case ErrInvalidToken, ErrTokenExpired, ErrUnauthorized:
		return http.StatusUnauthorized
	case ErrForbidden:
		return http.StatusForbidden
	case ErrNotFound, ErrUserNotFound:
		return http.StatusNotFound
	case ErrRateLimitExceeded:
		return http.StatusTooManyRequests
	case ErrInvalidInput:
		return http.StatusBadRequest
	default:
		return http.StatusInternalServerError
	}
}

// GetErrorCode extracts error code from error
func GetErrorCode(err error) string {
	var appErr *AppError
	if errors.As(err, &appErr) {
		if appErr.Code != "" {
			return appErr.Code
		}
	}

	// Map common errors to codes
	switch err {
	case ErrUserNotFound:
		return CodeUserNotFound
	case ErrInvalidCredentials:
		return CodeInvalidCredentials
	case ErrAccountLocked:
		return CodeAccountLocked
	case ErrEmailExists:
		return CodeEmailExists
	case ErrInvalidToken:
		return CodeInvalidToken
	case ErrTokenExpired:
		return CodeTokenExpired
	case ErrUnauthorized:
		return CodeUnauthorized
	case ErrForbidden:
		return CodeForbidden
	case ErrRateLimitExceeded:
		return CodeRateLimitExceeded
	case ErrInvalidInput:
		return CodeInvalidInput
	case ErrNotFound:
		return CodeNotFound
	default:
		return CodeInternalError
	}
}
