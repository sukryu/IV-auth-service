package domain

import (
	"context"
	"time"
)

// UserRepository defines the interface for user data access.
type UserRepository interface {
	// SaveUser saves a user to the underlying storage.
	SaveUser(ctx context.Context, user *User) error
	// FindByUsername retrieves a user by username from the storage.
	FindByUsername(ctx context.Context, username string) (*User, error)
}

// PlatformAccountRepository defines the interface for platform account data access.
type PlatformAccountRepository interface {
	// Save saves a platform account to the underlying storage.
	Save(ctx context.Context, account *PlatformAccount) error
	// FindByID retrieves a platform account by its ID from the storage.
	FindByID(ctx context.Context, id string) (*PlatformAccount, error)
	// FindByUserID retrieves all platform accounts associated with a user from the storage.
	FindByUserID(ctx context.Context, userID string) ([]*PlatformAccount, error)
	// Delete removes a platform account from the storage.
	Delete(ctx context.Context, id string) error
}

// TokenRepository defines the interface for token blacklist management.
type TokenRepository interface {
	// BlacklistToken adds a token to the blacklist with the specified details.
	BlacklistToken(ctx context.Context, tokenID, userID, reason string, expiresAt time.Time) error
	// IsBlacklisted checks if a token is currently blacklisted.
	IsBlacklisted(ctx context.Context, tokenID string) (bool, error)
}

// AuditLogRepository defines the interface for audit log data access.
type AuditLogRepository interface {
	// LogAction records an audit log entry in the storage.
	LogAction(ctx context.Context, log *AuditLog) error
	// GetLogs retrieves audit logs, optionally filtered by user ID, with pagination.
	GetLogs(ctx context.Context, userID string, limit, offset int) ([]*AuditLog, error)
}

// EventPublisher defines the interface for publishing domain events.
type EventPublisher interface {
	// Publish sends an event to the underlying event system.
	Publish(event Event) error
	// Close shuts down the event publisher and releases resources.
	Close() error
}
