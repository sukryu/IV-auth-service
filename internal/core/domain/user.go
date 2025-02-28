package domain

import (
	"errors"
	"time"
)

// UserStatus defines the possible states of a user account.
type UserStatus string

const (
	UserStatusActive    UserStatus = "ACTIVE"
	UserStatusSuspended UserStatus = "SUSPENDED"
	UserStatusDeleted   UserStatus = "DELETED"
)

// User represents a user entity in the ImmersiVerse system.
type User struct {
	id               string
	username         string
	email            Email    // 값 객체 사용
	passwordHash     Password // 값 객체 사용
	roleIDs          []string
	status           UserStatus
	subscriptionTier string
	createdAt        time.Time
	updatedAt        time.Time
	lastLoginAt      *time.Time // nullable
}

// NewUser creates a new User instance with the given attributes.
func NewUser(id, username string, email Email, passwordHash Password, roleIDs []string, subscriptionTier string) (*User, error) {
	if id == "" {
		return nil, errors.New("user id must not be empty")
	}
	if username == "" {
		return nil, errors.New("username must not be empty")
	}
	if !email.IsValid() {
		return nil, errors.New("invalid email address")
	}

	now := time.Now()
	return &User{
		id:               id,
		username:         username,
		email:            email,
		passwordHash:     passwordHash,
		roleIDs:          roleIDs,
		status:           UserStatusActive,
		subscriptionTier: subscriptionTier,
		createdAt:        now,
		updatedAt:        now,
	}, nil
}

// ID returns the user's unique identifier.
func (u *User) ID() string {
	return u.id
}

// Username returns the user's username.
func (u *User) Username() string {
	return u.username
}

// Email returns the user's email address.
func (u *User) Email() Email {
	return u.email
}

// PasswordHash returns the hashed password.
func (u *User) PasswordHash() Password {
	return u.passwordHash
}

// RoleIDs returns the list of role identifiers assigned to the user.
func (u *User) RoleIDs() []string {
	return u.roleIDs
}

// Status returns the user's account status.
func (u *User) Status() UserStatus {
	return u.status
}

// SubscriptionTier returns the user's subscription tier.
func (u *User) SubscriptionTier() string {
	return u.subscriptionTier
}

// CreatedAt returns the time when the user was created.
func (u *User) CreatedAt() time.Time {
	return u.createdAt
}

// UpdatedAt returns the time when the user was last updated.
func (u *User) UpdatedAt() time.Time {
	return u.updatedAt
}

// LastLoginAt returns the time of the user's last login, or nil if not recorded.
func (u *User) LastLoginAt() *time.Time {
	return u.lastLoginAt
}

// SetStatus updates the user's status and marks the update time.
func (u *User) SetStatus(status UserStatus) error {
	switch status {
	case UserStatusActive, UserStatusSuspended, UserStatusDeleted:
		u.status = status
		u.updatedAt = time.Now()
		return nil
	default:
		return errors.New("invalid user status")
	}
}

// SetLastLoginAt updates the last login time.
func (u *User) SetLastLoginAt(t time.Time) {
	u.lastLoginAt = &t
	u.updatedAt = time.Now()
}
