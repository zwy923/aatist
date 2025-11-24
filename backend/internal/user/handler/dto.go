package handler

// RegisterRequest represents registration request
type RegisterProfile struct {
	StudentID        string `json:"studentId"`
	School           string `json:"school"`
	Faculty          string `json:"faculty"`
	OrganizationName string `json:"organizationName"`
	ContactTitle     string `json:"contactTitle"`
}

type RegisterRequest struct {
	Email    string           `json:"email" binding:"required,email"`
	Password string           `json:"password" binding:"required,min=10"`
	Name     string           `json:"name" binding:"required,min=1,max=100"`
	Role     string           `json:"role"`
	Profile  *RegisterProfile `json:"profile"`
}

// LoginRequest represents login request
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// RefreshTokenRequest represents refresh token request
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// UserResponse represents user information in response
type UserResponse struct {
	ID              int64   `json:"id"`
	Email           string  `json:"email"`
	Name            string  `json:"name"`
	Role            string  `json:"role"`
	StudentID       *string `json:"student_id,omitempty"`
	School          *string `json:"school,omitempty"`
	Faculty         *string `json:"faculty,omitempty"`
	IsVerifiedEmail bool    `json:"is_verified_email"`
	OAuthProvider   *string `json:"oauth_provider,omitempty"`
	LastLoginAt     *string `json:"last_login_at,omitempty"`
	CreatedAt       string  `json:"created_at"`
}

// AuthResponse represents authentication response
type AuthResponse struct {
	User         UserResponse `json:"user"`
	AccessToken  string       `json:"access_token"`
	RefreshToken string       `json:"refresh_token"`
}

// ErrorResponse represents error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
}
