package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sukryu/IV-auth-services/internal/core/domain"

	"go.uber.org/zap"
)

// tokenRepository implements domain.TokenRepository for PostgreSQL.
type tokenRepository struct {
	db     *pgxpool.Pool
	logger *zap.Logger
}

// NewTokenRepository creates a new tokenRepository instance.
func NewTokenRepository(db *pgxpool.Pool, logger *zap.Logger) domain.TokenRepository {
	return &tokenRepository{
		db:     db,
		logger: logger.With(zap.String("component", "token_repository")),
	}
}

// BlacklistToken adds a token to the blacklist in the database.
func (r *tokenRepository) BlacklistToken(ctx context.Context, tokenID, userID, reason string, expiresAt time.Time) error {
	if tokenID == "" {
		return errors.New("token id must not be empty")
	}

	query := `
        INSERT INTO token_blacklist (token_id, user_id, expires_at, reason, blacklisted_at)
        VALUES ($1, $2, $3, $4, $5)
        ON CONFLICT (token_id) DO NOTHING
    `
	_, err := r.db.Exec(ctx, query, tokenID, userID, expiresAt, reason, time.Now())
	if err != nil {
		r.logger.Error("Failed to blacklist token", zap.Error(err), zap.String("token_id", tokenID))
		return errors.New("failed to blacklist token: " + err.Error())
	}
	r.logger.Debug("Token blacklisted successfully", zap.String("token_id", tokenID))
	return nil
}

// IsBlacklisted checks if a token is in the blacklist.
func (r *tokenRepository) IsBlacklisted(ctx context.Context, tokenID string) (bool, error) {
	if tokenID == "" {
		return false, errors.New("token id must not be empty")
	}

	query := `
        SELECT EXISTS (
            SELECT 1 FROM token_blacklist
            WHERE token_id = $1 AND expires_at > $2
        )
    `
	var exists bool
	err := r.db.QueryRow(ctx, query, tokenID, time.Now()).Scan(&exists)
	if err != nil {
		r.logger.Error("Failed to check if token is blacklisted", zap.Error(err), zap.String("token_id", tokenID))
		return false, errors.New("failed to check blacklist: " + err.Error())
	}
	return exists, nil
}
