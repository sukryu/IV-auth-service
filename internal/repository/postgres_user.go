package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/sukryu/IV-auth-services/internal/domain"
)

// PostgresUserRepository implements UserRepository with PostgreSQL
type PostgresUserRepository struct {
	db *sql.DB
}

// NewPostgresUserRepository creates a new PostgresUserRepository
func NewPostgresUserRepository(db *sql.DB) *PostgresUserRepository {
	return &PostgresUserRepository{db: db}
}

// InsertUser adds a new user to the database
func (r *PostgresUserRepository) InsertUser(ctx context.Context, user *domain.User) error {
	query := `
		INSERT INTO users (id, username, email, password_hash, status, subscription_tier, created_at, updated_at, last_login_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`
	_, err := r.db.ExecContext(ctx, query,
		user.ID(), user.Username(), user.Email(), user.PasswordHash(),
		user.Status(), user.SubscriptionTier(), user.CreatedAt(), user.UpdatedAt(), user.LastLoginAt(),
	)
	if err != nil {
		return fmt.Errorf("failed to insert user: %w", err)
	}
	return nil
}

// GetUserByUsername retrieves a user by username
func (r *PostgresUserRepository) GetUserByUsername(ctx context.Context, username string) (*domain.User, error) {
	query := `
		SELECT id, username, email, password_hash, status, subscription_tier, created_at, updated_at, last_login_at
		FROM users WHERE username = $1
	`
	row := r.db.QueryRowContext(ctx, query, username)

	var id, usernameVal, email, passwordHash, status, subscriptionTier string
	var createdAt, updatedAt time.Time
	var lastLoginAt *time.Time

	err := row.Scan(&id, &usernameVal, &email, &passwordHash, &status, &subscriptionTier, &createdAt, &updatedAt, &lastLoginAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get user by username: %w", err)
	}

	user := domain.NewUserFromDB(id, usernameVal, email, passwordHash, status, subscriptionTier, createdAt, updatedAt, lastLoginAt)
	return user, nil
}

// UpdateUser updates an existing user
func (r *PostgresUserRepository) UpdateUser(ctx context.Context, user *domain.User) error {
	query := `
		UPDATE users SET 
			email = $1, password_hash = $2, status = $3, subscription_tier = $4,
			updated_at = $5, last_login_at = $6
		WHERE id = $7
	`
	_, err := r.db.ExecContext(ctx, query,
		user.Email(), user.PasswordHash(), user.Status(), user.SubscriptionTier(),
		user.UpdatedAt(), user.LastLoginAt(), user.ID(),
	)
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}
	return nil
}

// ExistsByUsername checks if a username already exists
func (r *PostgresUserRepository) ExistsByUsername(ctx context.Context, username string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM users WHERE username = $1)`
	var exists bool
	err := r.db.QueryRowContext(ctx, query, username).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check username existence: %w", err)
	}
	return exists, nil
}
