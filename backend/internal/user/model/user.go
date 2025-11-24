package model

import (
	"strings"
	"time"
)

// Role represents user role
type Role string

const (
	RoleStudent Role = "student"
	RoleCompany Role = "company"
	RoleAdmin   Role = "admin"
)

// IsValid checks if role is valid
func (r Role) IsValid() bool {
	return r == RoleStudent || r == RoleCompany || r == RoleAdmin
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

// User represents a user in the system
type User struct {
	ID                 int64                   `db:"id" json:"id"`
	Email              string                  `db:"email" json:"email"`
	PasswordHash       string                  `db:"password_hash" json:"-"`
	Name               string                  `db:"name" json:"name"`
	Nickname           *string                 `db:"nickname" json:"nickname,omitempty"`
	AvatarURL          *string                 `db:"avatar_url" json:"avatar_url,omitempty"`
	Role               Role                    `db:"role" json:"role"`
	StudentID          *string                 `db:"student_id" json:"student_id,omitempty"`
	School             *string                 `db:"school" json:"school,omitempty"`
	Faculty            *string                 `db:"faculty" json:"faculty,omitempty"`
	Major              *string                 `db:"major" json:"major,omitempty"`
	WeeklyHours        *int                    `db:"weekly_hours" json:"weekly_hours,omitempty"`
	EmotionalStatus    *string                 `db:"emotional_status" json:"emotional_status,omitempty"`
	WeeklyAvailability WeeklyAvailabilityArray `db:"weekly_availability" json:"weekly_availability,omitempty"`
	Skills             Skills                  `db:"skills" json:"skills,omitempty"`
	Bio                *string                 `db:"bio" json:"bio,omitempty"`
	ProfileVisibility  ProfileVisibility       `db:"profile_visibility" json:"profile_visibility"`
	IsVerifiedEmail    bool                    `db:"is_verified_email" json:"is_verified_email"`
	OAuthProvider      *string                 `db:"oauth_provider" json:"oauth_provider,omitempty"`
	OAuthSubject       *string                 `db:"oauth_subject" json:"oauth_subject,omitempty"`
	LastLoginAt        *time.Time              `db:"last_login_at" json:"last_login_at,omitempty"`
	FailedAttempts     int                     `db:"failed_attempts" json:"-"`
	LockedUntil        *time.Time              `db:"locked_until" json:"-"`
	CreatedAt          time.Time               `db:"created_at" json:"created_at"`
	UpdatedAt          time.Time               `db:"updated_at" json:"updated_at"`
}

// IsLocked checks if the account is currently locked
func (u *User) IsLocked() bool {
	if u.LockedUntil == nil {
		return false
	}
	return u.LockedUntil.After(time.Now())
}
