package domain

import (
	"errors"
	"time"
)

// Event defines the common interface for all domain events.
type Event interface {
	EventName() string
}

// UserCreated represents an event when a new user is created.
type UserCreated struct {
	userID    string
	timestamp time.Time
}

// NewUserCreated creates a new UserCreated event.
func NewUserCreated(userID string, timestamp time.Time) (*UserCreated, error) {
	if userID == "" {
		return nil, errors.New("user id must not be empty")
	}
	if timestamp.IsZero() {
		return nil, errors.New("timestamp must not be zero")
	}
	return &UserCreated{
		userID:    userID,
		timestamp: timestamp,
	}, nil
}

// EventName returns the name of the UserCreated event.
func (e *UserCreated) EventName() string {
	return "UserCreated"
}

// UserID returns the ID of the created user.
func (e *UserCreated) UserID() string {
	return e.userID
}

// Timestamp returns the time when the event occurred.
func (e *UserCreated) Timestamp() time.Time {
	return e.timestamp
}

// LoginFailed represents an event when a login attempt fails.
type LoginFailed struct {
	userID    string
	timestamp time.Time
}

// NewLoginFailed creates a new LoginFailed event.
func NewLoginFailed(userID string, timestamp time.Time) (*LoginFailed, error) {
	if userID == "" {
		return nil, errors.New("user id must not be empty")
	}
	if timestamp.IsZero() {
		return nil, errors.New("timestamp must not be zero")
	}
	return &LoginFailed{
		userID:    userID,
		timestamp: timestamp,
	}, nil
}

// EventName returns the name of the LoginFailed event.
func (e *LoginFailed) EventName() string {
	return "LoginFailed"
}

// UserID returns the ID of the user who failed to log in.
func (e *LoginFailed) UserID() string {
	return e.userID
}

// Timestamp returns the time when the event occurred.
func (e *LoginFailed) Timestamp() time.Time {
	return e.timestamp
}

// PlatformLinked represents an event when a platform account is linked (deprecated, use PlatformConnected).
type PlatformLinked struct {
	userID     string
	platformID string
	timestamp  time.Time
}

// NewPlatformLinked creates a new PlatformLinked event (deprecated).
func NewPlatformLinked(userID, platformID string, timestamp time.Time) (*PlatformLinked, error) {
	if userID == "" {
		return nil, errors.New("user id must not be empty")
	}
	if platformID == "" {
		return nil, errors.New("platform id must not be empty")
	}
	if timestamp.IsZero() {
		return nil, errors.New("timestamp must not be zero")
	}
	return &PlatformLinked{
		userID:     userID,
		platformID: platformID,
		timestamp:  timestamp,
	}, nil
}

// EventName returns the name of the PlatformLinked event.
func (e *PlatformLinked) EventName() string {
	return "PlatformLinked"
}

// UserID returns the ID of the user.
func (e *PlatformLinked) UserID() string {
	return e.userID
}

// PlatformID returns the ID of the linked platform account.
func (e *PlatformLinked) PlatformID() string {
	return e.platformID
}

// Timestamp returns the time when the event occurred.
func (e *PlatformLinked) Timestamp() time.Time {
	return e.timestamp
}

// UserUpdated represents an event when a user's information is updated.
type UserUpdated struct {
	userID    string
	timestamp time.Time
}

// NewUserUpdated creates a new UserUpdated event.
func NewUserUpdated(userID string, timestamp time.Time) (*UserUpdated, error) {
	if userID == "" {
		return nil, errors.New("user id must not be empty")
	}
	if timestamp.IsZero() {
		return nil, errors.New("timestamp must not be zero")
	}
	return &UserUpdated{
		userID:    userID,
		timestamp: timestamp,
	}, nil
}

// EventName returns the name of the UserUpdated event.
func (e *UserUpdated) EventName() string {
	return "UserUpdated"
}

// UserID returns the ID of the updated user.
func (e *UserUpdated) UserID() string {
	return e.userID
}

// Timestamp returns the time when the event occurred.
func (e *UserUpdated) Timestamp() time.Time {
	return e.timestamp
}

// UserStatusChanged represents an event when a user's status changes.
type UserStatusChanged struct {
	userID    string
	newStatus UserStatus
	timestamp time.Time
}

// NewUserStatusChanged creates a new UserStatusChanged event.
func NewUserStatusChanged(userID string, newStatus UserStatus, timestamp time.Time) (*UserStatusChanged, error) {
	if userID == "" {
		return nil, errors.New("user id must not be empty")
	}
	if newStatus == "" {
		return nil, errors.New("new status must not be empty")
	}
	if timestamp.IsZero() {
		return nil, errors.New("timestamp must not be zero")
	}
	return &UserStatusChanged{
		userID:    userID,
		newStatus: newStatus,
		timestamp: timestamp,
	}, nil
}

// EventName returns the name of the UserStatusChanged event.
func (e *UserStatusChanged) EventName() string {
	return "UserStatusChanged"
}

// UserID returns the ID of the user whose status changed.
func (e *UserStatusChanged) UserID() string {
	return e.userID
}

// NewStatus returns the new status of the user.
func (e *UserStatusChanged) NewStatus() UserStatus {
	return e.newStatus
}

// Timestamp returns the time when the event occurred.
func (e *UserStatusChanged) Timestamp() time.Time {
	return e.timestamp
}

// LoginSucceeded represents an event when a login attempt succeeds.
type LoginSucceeded struct {
	userID    string
	timestamp time.Time
}

// NewLoginSucceeded creates a new LoginSucceeded event.
func NewLoginSucceeded(userID string, timestamp time.Time) (*LoginSucceeded, error) {
	if userID == "" {
		return nil, errors.New("user id must not be empty")
	}
	if timestamp.IsZero() {
		return nil, errors.New("timestamp must not be zero")
	}
	return &LoginSucceeded{
		userID:    userID,
		timestamp: timestamp,
	}, nil
}

// EventName returns the name of the LoginSucceeded event.
func (e *LoginSucceeded) EventName() string {
	return "LoginSucceeded"
}

// UserID returns the ID of the user who logged in successfully.
func (e *LoginSucceeded) UserID() string {
	return e.userID
}

// Timestamp returns the time when the event occurred.
func (e *LoginSucceeded) Timestamp() time.Time {
	return e.timestamp
}

// PlatformConnected represents an event when a platform account is connected.
type PlatformConnected struct {
	userID     string
	platformID string
	timestamp  time.Time
}

// NewPlatformConnected creates a new PlatformConnected event.
func NewPlatformConnected(userID, platformID string, timestamp time.Time) (*PlatformConnected, error) {
	if userID == "" {
		return nil, errors.New("user id must not be empty")
	}
	if platformID == "" {
		return nil, errors.New("platform id must not be empty")
	}
	if timestamp.IsZero() {
		return nil, errors.New("timestamp must not be zero")
	}
	return &PlatformConnected{
		userID:     userID,
		platformID: platformID,
		timestamp:  timestamp,
	}, nil
}

// EventName returns the name of the PlatformConnected event.
func (e *PlatformConnected) EventName() string {
	return "PlatformConnected"
}

// UserID returns the ID of the user.
func (e *PlatformConnected) UserID() string {
	return e.userID
}

// PlatformID returns the ID of the connected platform account.
func (e *PlatformConnected) PlatformID() string {
	return e.platformID
}

// Timestamp returns the time when the event occurred.
func (e *PlatformConnected) Timestamp() time.Time {
	return e.timestamp
}

// PlatformDisconnected represents an event when a platform account is disconnected.
type PlatformDisconnected struct {
	userID     string
	platformID string
	timestamp  time.Time
}

// NewPlatformDisconnected creates a new PlatformDisconnected event.
func NewPlatformDisconnected(userID, platformID string, timestamp time.Time) (*PlatformDisconnected, error) {
	if userID == "" {
		return nil, errors.New("user id must not be empty")
	}
	if platformID == "" {
		return nil, errors.New("platform id must not be empty")
	}
	if timestamp.IsZero() {
		return nil, errors.New("timestamp must not be zero")
	}
	return &PlatformDisconnected{
		userID:     userID,
		platformID: platformID,
		timestamp:  timestamp,
	}, nil
}

// EventName returns the name of the PlatformDisconnected event.
func (e *PlatformDisconnected) EventName() string {
	return "PlatformDisconnected"
}

// UserID returns the ID of the user.
func (e *PlatformDisconnected) UserID() string {
	return e.userID
}

// PlatformID returns the ID of the disconnected platform account.
func (e *PlatformDisconnected) PlatformID() string {
	return e.platformID
}

// Timestamp returns the time when the event occurred.
func (e *PlatformDisconnected) Timestamp() time.Time {
	return e.timestamp
}
