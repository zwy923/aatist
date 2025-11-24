package handler

import (
	"fmt"
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
	ID              int64             `json:"id"`
	Email           string            `json:"email"`
	Name            string            `json:"name"`
	Nickname        *string           `json:"nickname,omitempty"`
	AvatarURL       *string           `json:"avatar_url,omitempty"`
	Role            string            `json:"role"`
	StudentID       *string           `json:"student_id,omitempty"`
	School          *string           `json:"school,omitempty"`
	Faculty         *string           `json:"faculty,omitempty"`
	Major           *string           `json:"major,omitempty"`
	Availability    *string           `json:"availability,omitempty"`
	Projects        model.Projects    `json:"projects,omitempty"`
	Skills          model.StringArray `json:"skills,omitempty"`
	Bio             *string           `json:"bio,omitempty"`
	IsVerifiedEmail bool              `json:"is_verified_email"`
	OAuthProvider   *string           `json:"oauth_provider,omitempty"`
	LastLoginAt     *string           `json:"last_login_at,omitempty"`
	CreatedAt       string            `json:"created_at"`
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

// UpdateProfileRequest represents PATCH /users/me payload.
type UpdateProfileRequest struct {
	Name         *string         `json:"name" binding:"omitempty,max=100"`
	Nickname     *string         `json:"nickname" binding:"omitempty,max=100"`
	AvatarURL    *string         `json:"avatar_url"`
	StudentID    *string         `json:"student_id" binding:"omitempty,max=64"`
	School       *string         `json:"school" binding:"omitempty,max=255"`
	Faculty      *string         `json:"faculty" binding:"omitempty,max=255"`
	Major        *string         `json:"major" binding:"omitempty,max=255"`
	Availability *string         `json:"availability" binding:"omitempty,max=100"`
	Bio          *string         `json:"bio" binding:"omitempty,max=2000"`
	Projects     *[]ProjectInput `json:"projects"`
	Skills       *[]string       `json:"skills"`
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
	if r.Projects != nil {
		for i, project := range *r.Projects {
			if strings.TrimSpace(project.Title) == "" {
				return fmt.Errorf("projects[%d].title is required", i)
			}
			if strings.TrimSpace(project.Description) == "" {
				return fmt.Errorf("projects[%d].description is required", i)
			}
		}
	}
	return nil
}

func (r UpdateProfileRequest) ToServiceInput() service.UpdateProfileInput {
	input := service.UpdateProfileInput{
		Name:         r.Name,
		Nickname:     r.Nickname,
		AvatarURL:    r.AvatarURL,
		StudentID:    r.StudentID,
		School:       r.School,
		Faculty:      r.Faculty,
		Major:        r.Major,
		Availability: r.Availability,
		Bio:          r.Bio,
		Skills:       r.Skills,
	}

	if r.Projects != nil {
		projects := make([]model.Project, len(*r.Projects))
		for i, project := range *r.Projects {
			projects[i] = model.Project{
				Title:       strings.TrimSpace(project.Title),
				ClientName:  strings.TrimSpace(project.ClientName),
				Description: strings.TrimSpace(project.Description),
				Tags:        project.Tags,
				Year:        project.Year,
			}
		}
		input.Projects = &projects
	}

	return input
}
