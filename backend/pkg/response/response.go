package response

import (
	"github.com/aatist/backend/pkg/errs"
)

// Response represents a unified API response
type Response struct {
	Success bool                   `json:"success"`
	Data    interface{}            `json:"data,omitempty"`
	Error   *ErrorResponse          `json:"error,omitempty"`
}

// ErrorResponse represents error information in response
type ErrorResponse struct {
	Code    string                 `json:"code"`
	Message string                 `json:"message"`
	Details map[string]interface{} `json:"details,omitempty"`
}

// Success creates a success response
func Success(data interface{}) *Response {
	return &Response{
		Success: true,
		Data:    data,
	}
}

// Error creates an error response
func Error(err error) *Response {
	code := errs.GetErrorCode(err)
	message := err.Error()

	var details map[string]interface{}
	if appErr, ok := err.(*errs.AppError); ok {
		if appErr.Message != "" {
			message = appErr.Message
		}
		if appErr.Details != nil {
			details = appErr.Details
		}
	}

	return &Response{
		Success: false,
		Error: &ErrorResponse{
			Code:    code,
			Message: message,
			Details: details,
		},
	}
}

// ErrorWithCode creates an error response with custom code
func ErrorWithCode(err error, code string, message string) *Response {
	return &Response{
		Success: false,
		Error: &ErrorResponse{
			Code:    code,
			Message: message,
		},
	}
}

