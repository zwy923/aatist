package model

import (
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

// User represents a user in the system
type User struct {
	ID              int64       `db:"id" json:"id"`
	Email           string      `db:"email" json:"email"`
	PasswordHash    string      `db:"password_hash" json:"-"`
	Name            string      `db:"name" json:"name"`
	Nickname        *string     `db:"nickname" json:"nickname,omitempty"`
	AvatarURL       *string     `db:"avatar_url" json:"avatar_url,omitempty"`
	Role            Role        `db:"role" json:"role"`
	StudentID       *string     `db:"student_id" json:"student_id,omitempty"`
	School          *string     `db:"school" json:"school,omitempty"`
	Faculty         *string     `db:"faculty" json:"faculty,omitempty"`
	Major           *string     `db:"major" json:"major,omitempty"`
	Availability    *string     `db:"availability" json:"availability,omitempty"`
	Projects        Projects    `db:"projects" json:"projects,omitempty"`
	Skills          StringArray `db:"skills" json:"skills,omitempty"`
	Bio             *string     `db:"bio" json:"bio,omitempty"`
	IsVerifiedEmail bool        `db:"is_verified_email" json:"is_verified_email"`
	OAuthProvider   *string     `db:"oauth_provider" json:"oauth_provider,omitempty"`
	LastLoginAt     *time.Time  `db:"last_login_at" json:"last_login_at,omitempty"`
	FailedAttempts  int         `db:"failed_attempts" json:"-"`
	LockedUntil     *time.Time  `db:"locked_until" json:"-"`
	CreatedAt       time.Time   `db:"created_at" json:"created_at"`
	UpdatedAt       time.Time   `db:"updated_at" json:"updated_at"`
}

// IsLocked checks if the account is currently locked
func (u *User) IsLocked() bool {
	if u.LockedUntil == nil {
		return false
	}
	return u.LockedUntil.After(time.Now())
}
