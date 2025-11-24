package model

// EmailVerificationRequest represents a request to send email verification
type EmailVerificationRequest struct {
	UserID int64  `json:"user_id"`
	Email  string `json:"email"`
	Name   string `json:"name"`
	Token  string `json:"token"`
}
