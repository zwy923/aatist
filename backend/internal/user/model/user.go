package model

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// GuidedProfileQuestionsJSON stores optional private Q&A (JSONB)
type GuidedProfileQuestionsJSON map[string]interface{}

func (g GuidedProfileQuestionsJSON) Value() (driver.Value, error) {
	if g == nil || len(g) == 0 {
		return []byte("{}"), nil
	}
	return json.Marshal(g)
}

func (g *GuidedProfileQuestionsJSON) Scan(value interface{}) error {
	if value == nil {
		*g = nil
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("invalid type for guided_profile_questions: %T", value)
	}
	if len(bytes) == 0 {
		*g = nil
		return nil
	}
	var m map[string]interface{}
	if err := json.Unmarshal(bytes, &m); err != nil {
		return err
	}
	*g = m
	return nil
}

// Role represents user role
type Role string

const (
	RoleStudent   Role = "student"
	RoleAlumni    Role = "alumni"
	RoleOrgPerson Role = "org_person"
	RoleOrgTeam   Role = "org_team"
)

// IsValid checks if role is valid
func (r Role) IsValid() bool {
	return r == RoleStudent || r == RoleAlumni || r == RoleOrgPerson || r == RoleOrgTeam
}

// IsStudentRole checks if role is student or alumni
func (r Role) IsStudentRole() bool {
	return r == RoleStudent || r == RoleAlumni
}

// IsOrgRole checks if role is organization related
func (r Role) IsOrgRole() bool {
	return r == RoleOrgPerson || r == RoleOrgTeam
}

// String returns string representation of role
func (r Role) String() string {
	return string(r)
}

// ProfileVisibility represents profile visibility setting
type ProfileVisibility string

const (
	VisibilityPublic    ProfileVisibility = "public"
	VisibilityAaltoOnly ProfileVisibility = "aalto_only"
	VisibilityPrivate   ProfileVisibility = "private"
)

// IsValid checks if visibility is valid
func (v ProfileVisibility) IsValid() bool {
	return v == VisibilityPublic || v == VisibilityAaltoOnly || v == VisibilityPrivate
}

// String returns string representation of visibility
func (v ProfileVisibility) String() string {
	return string(v)
}

// PortfolioVisibility represents portfolio visibility setting
type PortfolioVisibility string

const (
	PortfolioVisibilityPublic    PortfolioVisibility = "public"
	PortfolioVisibilityAaltoOnly PortfolioVisibility = "aalto_only"
	PortfolioVisibilityPrivate   PortfolioVisibility = "private"
)

// IsValid checks if portfolio visibility is valid
func (v PortfolioVisibility) IsValid() bool {
	return v == PortfolioVisibilityPublic || v == PortfolioVisibilityAaltoOnly || v == PortfolioVisibilityPrivate
}

// String returns string representation of portfolio visibility
func (v PortfolioVisibility) String() string {
	return string(v)
}

// CanView checks if a viewer can view this profile
// viewerEmail can be nil for anonymous users
func (v ProfileVisibility) CanView(viewerEmail *string) bool {
	switch v {
	case VisibilityPublic:
		return true
	case VisibilityAaltoOnly:
		// Only Aalto users can view
		if viewerEmail == nil {
			return false
		}
		return strings.HasSuffix(strings.ToLower(*viewerEmail), "@aalto.fi")
	case VisibilityPrivate:
		// Only the owner can view (checked separately)
		return false
	default:
		return false
	}
}

// CanView checks if a viewer can view this portfolio
// viewerEmail can be nil for anonymous users
func (v PortfolioVisibility) CanView(viewerEmail *string) bool {
	switch v {
	case PortfolioVisibilityPublic:
		return true
	case PortfolioVisibilityAaltoOnly:
		// Only Aalto users can view
		if viewerEmail == nil {
			return false
		}
		return strings.HasSuffix(strings.ToLower(*viewerEmail), "@aalto.fi")
	case PortfolioVisibilityPrivate:
		// Only the owner can view (checked separately)
		return false
	default:
		return false
	}
}

// User represents a user in the system
type User struct {
	// Common fields (all roles)
	ID                int64             `db:"id" json:"id"`
	Email             string            `db:"email" json:"email"`
	PasswordHash      *string           `db:"password_hash" json:"-"` // OAuth users may not have password
	Name              string            `db:"name" json:"name"`
	AvatarURL         *string           `db:"avatar_url" json:"avatar_url,omitempty"`
	BannerURL         *string           `db:"banner_url" json:"banner_url,omitempty"`
	Role              Role              `db:"role" json:"role"`
	Bio               *string           `db:"bio" json:"bio,omitempty"`
	Website           *string           `db:"website" json:"website,omitempty"`
	LinkedIn          *string           `db:"linkedin" json:"linkedin,omitempty"`
	Behance           *string           `db:"behance" json:"behance,omitempty"`
	Languages         *string           `db:"languages" json:"languages,omitempty"`
	ProfessionalInterests *string        `db:"professional_interests" json:"professional_interests,omitempty"`
	GuidedProfileQuestions GuidedProfileQuestionsJSON `db:"guided_profile_questions" json:"guided_profile_questions,omitempty"` // Private
	ProfileVisibility ProfileVisibility `db:"profile_visibility" json:"profile_visibility"`
	IsVerifiedEmail   bool              `db:"is_verified_email" json:"is_verified_email"`
	RoleVerified      bool              `db:"role_verified" json:"role_verified"` // True if email is from verified school domain
	OAuthProvider     *string           `db:"oauth_provider" json:"oauth_provider,omitempty"`
	OAuthSubject      *string           `db:"oauth_subject" json:"oauth_subject,omitempty"`
	LastLoginAt       *time.Time        `db:"last_login_at" json:"last_login_at,omitempty"`
	FailedAttempts    int               `db:"failed_attempts" json:"-"`
	LockedUntil       *time.Time        `db:"locked_until" json:"-"`
	CreatedAt         time.Time         `db:"created_at" json:"created_at"`
	UpdatedAt         time.Time         `db:"updated_at" json:"updated_at"`
	// Student/Alumni specific fields
	StudentID           *string                 `db:"student_id" json:"student_id,omitempty"`
	School              *string                 `db:"school" json:"school,omitempty"`
	Faculty             *string                 `db:"faculty" json:"faculty,omitempty"`
	Major               *string                 `db:"major" json:"major,omitempty"`
	Skills              Skills                  `db:"skills" json:"skills,omitempty"`
	Courses             Courses                 `db:"courses" json:"courses,omitempty"`
	PortfolioVisibility PortfolioVisibility     `db:"portfolio_visibility" json:"portfolio_visibility"`
	// Organization specific fields (org_person, org_team)
	OrganizationName       *string `db:"organization_name" json:"organization_name,omitempty"`
	OrganizationBio        *string `db:"organization_bio" json:"organization_bio,omitempty"`
	ContactTitle           *string `db:"contact_title" json:"contact_title,omitempty"`
	IsAffiliatedWithSchool bool    `db:"is_affiliated_with_school" json:"is_affiliated_with_school"`
	OrgSize                *int    `db:"org_size" json:"org_size,omitempty"`
}

// IsLocked checks if the account is currently locked
func (u *User) IsLocked() bool {
	if u.LockedUntil == nil {
		return false
	}
	return u.LockedUntil.After(time.Now())
}

// Course represents a course taken by a user
type Course struct {
	Code string `json:"code"`
	Name string `json:"name"`
}

// Courses is a JSONB-backed slice of Course
type Courses []Course

func (c Courses) Value() (driver.Value, error) {
	if len(c) == 0 {
		return []byte("[]"), nil
	}
	b, err := json.Marshal(c)
	if err != nil {
		return nil, err
	}
	return string(b), nil
}

func (c *Courses) Scan(value interface{}) error {
	if value == nil {
		*c = nil
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("invalid type for Courses: %T", value)
	}

	if len(bytes) == 0 {
		*c = nil
		return nil
	}

	var temp []Course
	if err := json.Unmarshal(bytes, &temp); err != nil {
		return err
	}
	*c = temp
	return nil
}

// SkillMetadata represents a skill in the global skills table
type SkillMetadata struct {
	ID        int64     `db:"id" json:"id"`
	Name      string    `db:"name" json:"name"`
	Category  *string   `db:"category" json:"category,omitempty"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
}

// CourseMetadata represents a course in the global courses table
type CourseMetadata struct {
	ID        int64     `db:"id" json:"id"`
	Code      string    `db:"code" json:"code"`
	Name      string    `db:"name" json:"name"`
	School    *string   `db:"school" json:"school,omitempty"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
}

// TagMetadata represents a tag in the global tags table
type TagMetadata struct {
	ID        int64     `db:"id" json:"id"`
	Name      string    `db:"name" json:"name"`
	Type      string    `db:"type" json:"type"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
}
