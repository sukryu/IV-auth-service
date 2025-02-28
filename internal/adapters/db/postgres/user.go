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

// userRepository implements domain.UserRepository for PostgreSQL.
type userRepository struct {
	db     *pgxpool.Pool
	logger *zap.Logger
}

// NewUserRepository creates a new userRepository instance.
func NewUserRepository(db *pgxpool.Pool, logger *zap.Logger) domain.UserRepository {
	return &userRepository{
		db:     db,
		logger: logger.With(zap.String("component", "user_repository")),
	}
}

// SaveUser saves a user to the database.
func (r *userRepository) SaveUser(ctx context.Context, user *domain.User) error {
	if user == nil {
		return errors.New("user must not be nil")
	}

	query := `
        INSERT INTO users (id, username, email, password_hash, status, subscription_tier, created_at, updated_at, last_login_at)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
        ON CONFLICT (id) DO UPDATE SET
            username = EXCLUDED.username,
            email = EXCLUDED.email,
            password_hash = EXCLUDED.password_hash,
            status = EXCLUDED.status,
            subscription_tier = EXCLUDED.subscription_tier,
            updated_at = EXCLUDED.updated_at,
            last_login_at = EXCLUDED.last_login_at
    `
	_, err := r.db.Exec(ctx, query,
		user.ID(),
		user.Username(),
		user.Email().String(),
		user.PasswordHash().Hash(),
		user.Status(),
		user.SubscriptionTier(),
		user.CreatedAt(),
		user.UpdatedAt(),
		user.LastLoginAt(),
	)
	if err != nil {
		r.logger.Error("Failed to save user", zap.Error(err), zap.String("user_id", user.ID()))
		return errors.New("failed to save user: " + err.Error())
	}
	r.logger.Debug("User saved successfully", zap.String("user_id", user.ID()))
	return nil
}

// FindByUsername retrieves a user by username from the database.
func (r *userRepository) FindByUsername(ctx context.Context, username string) (*domain.User, error) {
	if username == "" {
		return nil, errors.New("username must not be empty")
	}

	query := `
        SELECT id, username, email, password_hash, status, subscription_tier, created_at, updated_at, last_login_at
        FROM users
        WHERE username = $1
    `
	row := r.db.QueryRow(ctx, query, username)

	var (
		id               string
		emailStr         string
		passwordHash     []byte
		status           domain.UserStatus
		subscriptionTier string
		createdAt        time.Time
		updatedAt        time.Time
		lastLoginAt      sql.NullTime
	)
	err := row.Scan(&id, &username, &emailStr, &passwordHash, &status, &subscriptionTier, &createdAt, &updatedAt, &lastLoginAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil // 사용자 없음
		}
		r.logger.Error("Failed to find user by username", zap.Error(err), zap.String("username", username))
		return nil, errors.New("failed to find user: " + err.Error())
	}

	email, err := domain.NewEmail(emailStr)
	if err != nil {
		return nil, err
	}
	pwd, err := domain.NewPasswordFromHash(passwordHash, nil) // Salt 미구현 상태
	if err != nil {
		return nil, err
	}
	user, err := domain.NewUser(id, username, email, pwd, nil, subscriptionTier)
	if err != nil {
		return nil, err
	}
	user.SetStatus(status) // 상태 복원
	if lastLoginAt.Valid {
		user.SetLastLoginAt(lastLoginAt.Time)
	}
	return user, nil
}
