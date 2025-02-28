package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

// PostgresTokenRepository implements TokenRepository with PostgreSQL
type PostgresTokenRepository struct {
	db *sql.DB
}

// NewPostgresTokenRepository creates a new instance
func NewPostgresTokenRepository(db *sql.DB) *PostgresTokenRepository {
	return &PostgresTokenRepository{db: db}
}

// InsertTokenBlacklist adds a token to the blacklist
func (r *PostgresTokenRepository) InsertTokenBlacklist(ctx context.Context, tokenID, userID string, expiresAt time.Time, reason string) error {
	query := `
		INSERT INTO token_blacklist (token_id, user_id, expires_at, reason, blacklisted_at)
		VALUES ($1, $2, $3, $4, $5)
	`
	_, err := r.db.ExecContext(ctx, query, tokenID, userID, expiresAt, reason, time.Now())
	if err != nil {
		return fmt.Errorf("failed to insert token blacklist: %w", err)
	}
	return nil
}

// IsTokenBlacklisted checks if a token is in the blacklist
func (r *PostgresTokenRepository) IsTokenBlacklisted(ctx context.Context, tokenID string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM token_blacklist WHERE token_id = $1 AND expires_at > NOW())`
	var exists bool
	err := r.db.QueryRowContext(ctx, query, tokenID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check token blacklist: %w", err)
	}
	return exists, nil
}
