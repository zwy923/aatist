package handler

import (
	"strings"

	"github.com/aatist/backend/internal/user/model"
	"github.com/aatist/backend/internal/user/service"
)

// RegisterProfile represents profile data during registration
type RegisterProfile struct {
	// Student/Alumni fields
	StudentID string `json:"studentId,omitempty"`
	School    string `json:"school,omitempty"`
	Faculty   string `json:"faculty,omitempty"`
	Major     string `json:"major,omitempty"`
	// Organization fields
	OrganizationName       string `json:"organizationName,omitempty"`
	OrganizationBio        string `json:"organizationBio,omitempty"`
	ContactTitle           string `json:"contactTitle,omitempty"`
	IsAffiliatedWithSchool bool   `json:"isAffiliatedWithSchool,omitempty"`
	OrgSize                *int   `json:"orgSize,omitempty"`
}

type RegisterRequest struct {
	Email    string           `json:"email" binding:"required,email"`
	Password string           `json:"password" binding:"required,min=8"`
	Name     string           `json:"name" binding:"required,min=1,max=100"`
	Role     string           `json:"role"`
	Profile  *RegisterProfile `json:"profile"`
}

// LoginRequest represents login request
type LoginRequest struct {
	Email     string `json:"email" binding:"required,email"`
	Password  string `json:"password" binding:"required"`
	LoginType string `json:"login_type,omitempty"` // "client" or "student" - validates role/email
}

// RefreshTokenRequest represents refresh token request
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// UserResponse represents user information in response
type UserResponse struct {
	// Common fields
	ID                    int64   `json:"id"`
	Email                 string  `json:"email"`
	Name                  string  `json:"name"`
	AvatarURL             *string `json:"avatar_url,omitempty"`
	Role                  string  `json:"role"`
	Bio                   *string `json:"bio,omitempty"`
	Website               *string `json:"website,omitempty"`
	LinkedIn              *string `json:"linkedin,omitempty"`
	Behance               *string `json:"behance,omitempty"`
	Languages             *string `json:"languages,omitempty"`
	ProfessionalInterests *string `json:"professional_interests,omitempty"`
	ProfileVisibility     string  `json:"profile_visibility"`
	IsVerifiedEmail       bool    `json:"is_verified_email"`
	RoleVerified          bool    `json:"role_verified"` // True if email is from verified school domain
	OAuthProvider         *string `json:"oauth_provider,omitempty"`
	LastLoginAt           *string `json:"last_login_at,omitempty"`
	CreatedAt             string  `json:"created_at"`
	// Student/Alumni fields
	StudentID           *string       `json:"student_id,omitempty"`
	School              *string       `json:"school,omitempty"`
	Faculty             *string       `json:"faculty,omitempty"`
	Major               *string       `json:"major,omitempty"`
	Skills              model.Skills  `json:"skills,omitempty"`
	Courses             model.Courses `json:"courses,omitempty"`
	PortfolioVisibility string        `json:"portfolio_visibility,omitempty"`
	// Organization fields
	OrganizationName       *string `json:"organization_name,omitempty"`
	OrganizationBio        *string `json:"organization_bio,omitempty"`
	ContactTitle           *string `json:"contact_title,omitempty"`
	IsAffiliatedWithSchool bool    `json:"is_affiliated_with_school,omitempty"`
	OrgSize                *int    `json:"org_size,omitempty"`
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

// ChangePasswordRequest represents change password request
type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" binding:"required"`
	NewPassword     string `json:"new_password" binding:"required,min=8"`
}

// CheckExistsResponse represents the response for check username/email existence
type CheckExistsResponse struct {
	Exists bool `json:"exists"`
}

// ForgotPasswordRequest represents forgot password request
type ForgotPasswordRequest struct {
	Email string `json:"email" binding:"required,email"`
}

// ResetPasswordRequest represents reset password request
type ResetPasswordRequest struct {
	Token       string `json:"token" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=8"`
}

// SkillInput represents a skill with level in request.
type SkillInput struct {
	Name  string `json:"name" binding:"required"`
	Level string `json:"level" binding:"required,oneof=expert advanced intermediate"`
}

// UpdateProfileRequest represents PATCH /users/me payload.
type UpdateProfileRequest struct {
	// Common fields
	Name                   *string                 `json:"name" binding:"omitempty,max=100"`
	AvatarURL              *string                 `json:"avatar_url"`
	Bio                    *string                 `json:"bio" binding:"omitempty,max=2000"`
	Website                *string                 `json:"website" binding:"omitempty,max=500"`
	LinkedIn               *string                 `json:"linkedin" binding:"omitempty,max=500"`
	Behance                *string                 `json:"behance" binding:"omitempty,max=500"`
	Languages              *string                 `json:"languages" binding:"omitempty,max=500"`
	ProfessionalInterests  *string                 `json:"professional_interests" binding:"omitempty,max=1000"`
	GuidedProfileQuestions *map[string]interface{} `json:"guided_profile_questions"`
	ProfileVisibility      *string                 `json:"profile_visibility" binding:"omitempty,oneof=public aalto_only private"`
	// Student/Alumni fields
	StudentID           *string       `json:"student_id" binding:"omitempty,max=64"`
	School              *string       `json:"school" binding:"omitempty,max=255"`
	Faculty             *string       `json:"faculty" binding:"omitempty,max=255"`
	Major               *string       `json:"major" binding:"omitempty,max=255"`
	Skills              *[]SkillInput `json:"skills"`
	PortfolioVisibility *string       `json:"portfolio_visibility" binding:"omitempty,oneof=public aalto_only private"`
	// Organization fields
	OrganizationName       *string `json:"organization_name" binding:"omitempty,max=255"`
	OrganizationBio        *string `json:"organization_bio" binding:"omitempty,max=2000"`
	ContactTitle           *string `json:"contact_title" binding:"omitempty,max=100"`
	IsAffiliatedWithSchool *bool   `json:"is_affiliated_with_school"`
	OrgSize                *int    `json:"org_size" binding:"omitempty,min=1"`
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
		Name:                   r.Name,
		AvatarURL:              r.AvatarURL,
		StudentID:              r.StudentID,
		School:                 r.School,
		Faculty:                r.Faculty,
		Major:                  r.Major,
		Bio:                    r.Bio,
		Website:                r.Website,
		LinkedIn:               r.LinkedIn,
		Behance:                r.Behance,
		Languages:              r.Languages,
		ProfessionalInterests:  r.ProfessionalInterests,
		GuidedProfileQuestions: r.GuidedProfileQuestions,
		OrganizationName:       r.OrganizationName,
		OrganizationBio:        r.OrganizationBio,
		ContactTitle:           r.ContactTitle,
		IsAffiliatedWithSchool: r.IsAffiliatedWithSchool,
		OrgSize:                r.OrgSize,
	}

	if r.ProfileVisibility != nil {
		vis := model.ProfileVisibility(*r.ProfileVisibility)
		if vis.IsValid() {
			input.ProfileVisibility = &vis
		}
	}

	if r.PortfolioVisibility != nil {
		portVis := model.PortfolioVisibility(*r.PortfolioVisibility)
		if portVis.IsValid() {
			input.PortfolioVisibility = &portVis
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
