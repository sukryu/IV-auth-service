package domain

import (
	"time"
)

// Event defines the base structure for domain events
type Event interface {
	Name() string
	OccurredAt() time.Time
}

// UserCreated event indicates a new user has been created
type UserCreated struct {
	userID     string
	username   string
	email      string
	occurredAt time.Time
}

// NewUserCreated creates a new UserCreated event
func NewUserCreated(userID, username, email string) *UserCreated {
	return &UserCreated{
		userID:     userID,
		username:   username,
		email:      email,
		occurredAt: time.Now(),
	}
}

func (e *UserCreated) Name() string          { return "UserCreated" }
func (e *UserCreated) OccurredAt() time.Time { return e.occurredAt }
func (e *UserCreated) UserID() string        { return e.userID }
func (e *UserCreated) Username() string      { return e.username }
func (e *UserCreated) Email() string         { return e.email }

// LoginSucceeded event indicates a successful user login
type LoginSucceeded struct {
	userID     string
	occurredAt time.Time
}

// NewLoginSucceeded creates a new LoginSucceeded event
func NewLoginSucceeded(userID string) *LoginSucceeded {
	return &LoginSucceeded{
		userID:     userID,
		occurredAt: time.Now(),
	}
}

func (e *LoginSucceeded) Name() string          { return "LoginSucceeded" }
func (e *LoginSucceeded) OccurredAt() time.Time { return e.occurredAt }
func (e *LoginSucceeded) UserID() string        { return e.userID }
