package domain

import (
	"time"

	"github.com/google/uuid"
)

// User represents a user entity in the ImmersiVerse platform
type User struct {
	id               string
	username         string
	email            string
	passwordHash     string
	status           string
	subscriptionTier string
	createdAt        time.Time
	updatedAt        time.Time
	lastLoginAt      *time.Time
}

// NewUser creates a new User instance with initial values
func NewUser(username, email, passwordHash string) (*User, error) {
	id := uuid.New().String()
	now := time.Now()

	return &User{
		id:               id,
		username:         username,
		email:            email,
		passwordHash:     passwordHash,
		status:           "ACTIVE",
		subscriptionTier: "FREE",
		createdAt:        now,
		updatedAt:        now,
		lastLoginAt:      nil,
	}, nil
}

// Getters (Kubernetes 스타일: 메서드명은 속성명과 동일하게)
func (u *User) ID() string               { return u.id }
func (u *User) Username() string         { return u.username }
func (u *User) Email() string            { return u.email }
func (u *User) PasswordHash() string     { return u.passwordHash }
func (u *User) Status() string           { return u.status }
func (u *User) SubscriptionTier() string { return u.subscriptionTier }
func (u *User) CreatedAt() time.Time     { return u.createdAt }
func (u *User) UpdatedAt() time.Time     { return u.updatedAt }
func (u *User) LastLoginAt() *time.Time  { return u.lastLoginAt }

// Setters (상태 변경 메서드)
func (u *User) SetStatus(status string) {
	u.status = status
	u.updatedAt = time.Now()
}

func (u *User) SetSubscriptionTier(tier string) {
	u.subscriptionTier = tier
	u.updatedAt = time.Now()
}

func (u *User) RecordLogin() {
	now := time.Now()
	u.lastLoginAt = &now
	u.updatedAt = now
}
