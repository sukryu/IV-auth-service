package postgres

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sukryu/IV-auth-services/internal/core/domain"
	"go.uber.org/zap"
)

// platformAccountRepository implements domain.PlatformAccountRepository for PostgreSQL.
type platformAccountRepository struct {
	db     *pgxpool.Pool
	logger *zap.Logger
}

// NewPlatformAccountRepository creates a new platformAccountRepository instance.
func NewPlatformAccountRepository(db *pgxpool.Pool, logger *zap.Logger) domain.PlatformAccountRepository {
	return &platformAccountRepository{
		db:     db,
		logger: logger.With(zap.String("component", "platform_account_repository")),
	}
}

// Save saves a platform account to the database.
func (r *platformAccountRepository) Save(ctx context.Context, account *domain.PlatformAccount) error {
	if account == nil {
		return errors.New("platform account must not be nil")
	}

	query := `
        INSERT INTO platform_accounts (id, user_id, platform, platform_user_id, platform_username, access_token, refresh_token, token_expires_at, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
        ON CONFLICT (id) DO UPDATE SET
            user_id = EXCLUDED.user_id,
            platform = EXCLUDED.platform,
            platform_user_id = EXCLUDED.platform_user_id,
            platform_username = EXCLUDED.platform_username,
            access_token = EXCLUDED.access_token,
            refresh_token = EXCLUDED.refresh_token,
            token_expires_at = EXCLUDED.token_expires_at,
            updated_at = EXCLUDED.updated_at
    `
	_, err := r.db.Exec(ctx, query,
		account.ID(),
		account.UserID(),
		account.Platform(),
		account.PlatformUserID(),
		account.PlatformUsername(),
		account.AccessToken(),
		account.RefreshToken(),
		account.TokenExpiresAt(),
		account.CreatedAt(),
		account.UpdatedAt(),
	)
	if err != nil {
		r.logger.Error("Failed to save platform account", zap.Error(err), zap.String("platform_id", account.ID()))
		return errors.New("failed to save platform account: " + err.Error())
	}
	r.logger.Debug("Platform account saved successfully", zap.String("platform_id", account.ID()))
	return nil
}

// FindByID retrieves a platform account by ID from the database.
func (r *platformAccountRepository) FindByID(ctx context.Context, id string) (*domain.PlatformAccount, error) {
	if id == "" {
		return nil, errors.New("platform account id must not be empty")
	}

	query := `
        SELECT id, user_id, platform, platform_user_id, platform_username, access_token, refresh_token, token_expires_at, created_at, updated_at
        FROM platform_accounts
        WHERE id = $1
    `
	row := r.db.QueryRow(ctx, query, id)

	var (
		userID           string
		platform         domain.PlatformType
		platformUserID   string
		platformUsername string
		accessToken      string
		refreshToken     string
		tokenExpiresAt   sql.NullTime
		createdAt        time.Time
		updatedAt        time.Time
	)
	err := row.Scan(&id, &userID, &platform, &platformUserID, &platformUsername, &accessToken, &refreshToken, &tokenExpiresAt, &createdAt, &updatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil // 계정 없음
		}
		r.logger.Error("Failed to find platform account by id", zap.Error(err), zap.String("platform_id", id))
		return nil, errors.New("failed to find platform account: " + err.Error())
	}

	var expiresAt *time.Time
	if tokenExpiresAt.Valid {
		expiresAt = &tokenExpiresAt.Time
	}
	return domain.NewPlatformAccount(id, userID, platform, platformUserID, platformUsername, accessToken, refreshToken, expiresAt)
}

// FindByUserID retrieves platform accounts by user ID from the database.
func (r *platformAccountRepository) FindByUserID(ctx context.Context, userID string) ([]*domain.PlatformAccount, error) {
	if userID == "" {
		return nil, errors.New("user id must not be empty")
	}

	query := `
        SELECT id, user_id, platform, platform_user_id, platform_username, access_token, refresh_token, token_expires_at, created_at, updated_at
        FROM platform_accounts
        WHERE user_id = $1
    `
	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		r.logger.Error("Failed to find platform accounts by user id", zap.Error(err), zap.String("user_id", userID))
		return nil, errors.New("failed to find platform accounts: " + err.Error())
	}
	defer rows.Close()

	var accounts []*domain.PlatformAccount
	for rows.Next() {
		var (
			id               string
			platform         domain.PlatformType
			platformUserID   string
			platformUsername string
			accessToken      string
			refreshToken     string
			tokenExpiresAt   sql.NullTime
			createdAt        time.Time
			updatedAt        time.Time
		)
		if err := rows.Scan(&id, &userID, &platform, &platformUserID, &platformUsername, &accessToken, &refreshToken, &tokenExpiresAt, &createdAt, &updatedAt); err != nil {
			r.logger.Error("Failed to scan platform account row", zap.Error(err))
			return nil, errors.New("failed to scan platform account: " + err.Error())
		}

		var expiresAt *time.Time
		if tokenExpiresAt.Valid {
			expiresAt = &tokenExpiresAt.Time
		}
		account, err := domain.NewPlatformAccount(id, userID, platform, platformUserID, platformUsername, accessToken, refreshToken, expiresAt)
		if err != nil {
			return nil, err
		}
		accounts = append(accounts, account)
	}

	if err := rows.Err(); err != nil {
		return nil, errors.New("error iterating platform account rows: " + err.Error())
	}
	return accounts, nil
}

// Delete removes a platform account from the database.
func (r *platformAccountRepository) Delete(ctx context.Context, id string) error {
	if id == "" {
		return errors.New("platform account id must not be empty")
	}

	query := `DELETE FROM platform_accounts WHERE id = $1`
	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		r.logger.Error("Failed to delete platform account", zap.Error(err), zap.String("platform_id", id))
		return errors.New("failed to delete platform account: " + err.Error())
	}
	if rowsAffected := result.RowsAffected(); rowsAffected == 0 {
		return errors.New("platform account not found")
	}
	r.logger.Debug("Platform account deleted successfully", zap.String("platform_id", id))
	return nil
}
