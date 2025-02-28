package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/sukryu/IV-auth-services/internal/domain"
)

// PostgresPlatformAccountRepository implements PlatformAccountRepository with PostgreSQL
type PostgresPlatformAccountRepository struct {
	db *sql.DB
}

// NewPostgresPlatformAccountRepository creates a new instance
func NewPostgresPlatformAccountRepository(db *sql.DB) *PostgresPlatformAccountRepository {
	return &PostgresPlatformAccountRepository{db: db}
}

// InsertPlatformAccount adds a new platform account
func (r *PostgresPlatformAccountRepository) InsertPlatformAccount(ctx context.Context, account *domain.PlatformAccount) error {
	query := `
		INSERT INTO platform_accounts (id, user_id, platform, platform_user_id, platform_username, access_token, refresh_token, token_expires_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`
	_, err := r.db.ExecContext(ctx, query,
		account.ID(), account.UserID(), account.Platform(), account.PlatformUserID(), account.PlatformUsername(),
		account.AccessToken(), account.RefreshToken(), account.TokenExpiresAt(), account.CreatedAt(), account.UpdatedAt(),
	)
	if err != nil {
		return fmt.Errorf("failed to insert platform account: %w", err)
	}
	return nil
}

// GetPlatformAccountsByUserID retrieves all platform accounts for a user
func (r *PostgresPlatformAccountRepository) GetPlatformAccountsByUserID(ctx context.Context, userID string) ([]*domain.PlatformAccount, error) {
	query := `
		SELECT id, user_id, platform, platform_user_id, platform_username, access_token, refresh_token, token_expires_at, created_at, updated_at
		FROM platform_accounts WHERE user_id = $1
	`
	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get platform accounts: %w", err)
	}
	defer rows.Close()

	var accounts []*domain.PlatformAccount
	for rows.Next() {
		var id, userIDVal, platform, platformUserID, platformUsername, accessToken, refreshToken string
		var tokenExpiresAt *time.Time
		var createdAt, updatedAt time.Time

		err := rows.Scan(&id, &userIDVal, &platform, &platformUserID, &platformUsername,
			&accessToken, &refreshToken, &tokenExpiresAt, &createdAt, &updatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan platform account: %w", err)
		}

		acc := domain.NewPlatformAccountFromDB(id, userIDVal, platform, platformUserID, platformUsername, accessToken, refreshToken, tokenExpiresAt, createdAt, updatedAt)
		accounts = append(accounts, acc)
	}
	return accounts, nil
}

// UpdatePlatformAccount updates an existing platform account
func (r *PostgresPlatformAccountRepository) UpdatePlatformAccount(ctx context.Context, account *domain.PlatformAccount) error {
	query := `
		UPDATE platform_accounts SET 
			access_token = $1, refresh_token = $2, token_expires_at = $3, updated_at = $4
		WHERE id = $5
	`
	_, err := r.db.ExecContext(ctx, query,
		account.AccessToken(), account.RefreshToken(), account.TokenExpiresAt(), account.UpdatedAt(), account.ID(),
	)
	if err != nil {
		return fmt.Errorf("failed to update platform account: %w", err)
	}
	return nil
}
