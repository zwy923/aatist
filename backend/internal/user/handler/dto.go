package handler

import (
	"strings"

	"github.com/aalto-talent-network/backend/internal/user/model"
	"github.com/aalto-talent-network/backend/internal/user/service"
)

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
	ID                 int64                         `json:"id"`
	Email              string                        `json:"email"`
	Name               string                        `json:"name"`
	Nickname           *string                       `json:"nickname,omitempty"`
	AvatarURL          *string                       `json:"avatar_url,omitempty"`
	Role               string                        `json:"role"`
	StudentID          *string                       `json:"student_id,omitempty"`
	School             *string                       `json:"school,omitempty"`
	Faculty            *string                       `json:"faculty,omitempty"`
	Major              *string                       `json:"major,omitempty"`
	WeeklyHours        *int                          `json:"weekly_hours,omitempty"`
	EmotionalStatus    *string                       `json:"emotional_status,omitempty"`
	WeeklyAvailability model.WeeklyAvailabilityArray `json:"weekly_availability,omitempty"`
	Skills             model.Skills                  `json:"skills,omitempty"`
	Bio                *string                       `json:"bio,omitempty"`
	IsVerifiedEmail    bool                          `json:"is_verified_email"`
	OAuthProvider      *string                       `json:"oauth_provider,omitempty"`
	LastLoginAt        *string                       `json:"last_login_at,omitempty"`
	CreatedAt          string                        `json:"created_at"`
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

// SkillInput represents a skill with level in request.
type SkillInput struct {
	Name  string `json:"name" binding:"required"`
	Level string `json:"level" binding:"required,oneof=expert advanced intermediate"`
}

// UpdateProfileRequest represents PATCH /users/me payload.
type UpdateProfileRequest struct {
	Name               *string                     `json:"name" binding:"omitempty,max=100"`
	Nickname           *string                     `json:"nickname" binding:"omitempty,max=100"`
	AvatarURL          *string                     `json:"avatar_url"`
	StudentID          *string                     `json:"student_id" binding:"omitempty,max=64"`
	School             *string                     `json:"school" binding:"omitempty,max=255"`
	Faculty            *string                     `json:"faculty" binding:"omitempty,max=255"`
	Major              *string                     `json:"major" binding:"omitempty,max=255"`
	WeeklyHours        *int                        `json:"weekly_hours" binding:"omitempty,min=0,max=168"`
	EmotionalStatus    *string                     `json:"emotional_status" binding:"omitempty,max=50"`
	WeeklyAvailability *[]model.WeeklyAvailability `json:"weekly_availability"`
	Bio                *string                     `json:"bio" binding:"omitempty,max=2000"`
	ProfileVisibility  *string                     `json:"profile_visibility" binding:"omitempty,oneof=public aalto_only private"`
	Skills             *[]SkillInput               `json:"skills"`
}

// ProjectInput is the request representation of a project entry.
type ProjectInput struct {
	Title       string   `json:"title" binding:"omitempty,max=200"`
	ClientName  string   `json:"client_name" binding:"omitempty,max=200"`
	Description string   `json:"description" binding:"omitempty,max=2000"`
	Tags        []string `json:"tags"`
	Year        *int     `json:"year"`
}

func (r UpdateProfileRequest) Validate() error {
	// Projects are now managed via separate /portfolio endpoints
	return nil
}

func (r UpdateProfileRequest) ToServiceInput() service.UpdateProfileInput {
	input := service.UpdateProfileInput{
		Name:            r.Name,
		Nickname:        r.Nickname,
		AvatarURL:       r.AvatarURL,
		StudentID:       r.StudentID,
		School:          r.School,
		Faculty:         r.Faculty,
		Major:           r.Major,
		WeeklyHours:     r.WeeklyHours,
		EmotionalStatus: r.EmotionalStatus,
		Bio:             r.Bio,
	}

	if r.WeeklyAvailability != nil {
		wa := model.WeeklyAvailabilityArray(*r.WeeklyAvailability)
		input.WeeklyAvailability = &wa
	}

	if r.ProfileVisibility != nil {
		vis := model.ProfileVisibility(*r.ProfileVisibility)
		if vis.IsValid() {
			input.ProfileVisibility = &vis
		}
	}

	if r.Skills != nil {
		skills := make([]model.Skill, len(*r.Skills))
		for i, skill := range *r.Skills {
			skills[i] = model.Skill{
				Name:  strings.TrimSpace(skill.Name),
				Level: strings.ToLower(strings.TrimSpace(skill.Level)),
			}
		}
		input.Skills = &skills
	}

	return input
}
