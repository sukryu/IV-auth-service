package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/sukryu/IV-auth-services/internal/domain"
)

// UserRepository는 PostgreSQL을 사용한 사용자 데이터 접근을 구현합니다.
// domain/repo.UserRepository 인터페이스를 만족하며, CRUD 작업을 제공합니다.
type UserRepository struct {
	db *sql.DB
}

// NewUserRepository는 새로운 UserRepository 인스턴스를 생성하여 반환합니다.
// DB 연결 객체를 주입받아 초기화합니다.
func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

// FindByUsername은 사용자명을 기준으로 사용자 엔티티를 조회합니다.
// 사용자가 없으면 nil과 nil 에러를 반환하며, 쿼리 실패 시 에러를 반환합니다.
func (r *UserRepository) FindByUsername(ctx context.Context, username string) (*domain.User, error) {
	query := `
		SELECT id, username, email, password_hash, status, subscription_tier, created_at, updated_at, last_login_at
		FROM users
		WHERE username = $1
		LIMIT 1
	`
	row := r.db.QueryRowContext(ctx, query, username)

	user := &domain.User{}
	var lastLoginAt sql.NullTime
	err := row.Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.PasswordHash,
		&user.Status,
		&user.SubscriptionTier,
		&user.CreatedAt,
		&user.UpdatedAt,
		&lastLoginAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil // 사용자 없음
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find user by username: %w", err)
	}

	if lastLoginAt.Valid {
		user.LastLoginAt = lastLoginAt.Time
	}
	// 역할(Role)은 별도 테이블(user_roles)에서 조회 (아직 미구현)
	// user.Roles = fetchRoles(ctx, r.db, user.ID)
	return user, nil
}

// FindByID는 사용자 ID를 기준으로 사용자 엔티티를 조회합니다.
// 사용자가 없으면 nil과 nil 에러를 반환하며, 쿼리 실패 시 에러를 반환합니다.
func (r *UserRepository) FindByID(ctx context.Context, id string) (*domain.User, error) {
	query := `
		SELECT id, username, email, password_hash, status, subscription_tier, created_at, updated_at, last_login_at
		FROM users
		WHERE id = $1
		LIMIT 1
	`
	row := r.db.QueryRowContext(ctx, query, id)

	user := &domain.User{}
	var lastLoginAt sql.NullTime
	err := row.Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.PasswordHash,
		&user.Status,
		&user.SubscriptionTier,
		&user.CreatedAt,
		&user.UpdatedAt,
		&lastLoginAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil // 사용자 없음
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find user by id: %w", err)
	}

	if lastLoginAt.Valid {
		user.LastLoginAt = lastLoginAt.Time
	}
	// 역할(Role)은 별도 조회 필요 (미구현)
	return user, nil
}

// ExistsByUsername은 주어진 사용자명이 이미 존재하는지 확인합니다.
// 존재 여부를 bool로 반환하며, 쿼리 실패 시 에러를 반환합니다.
func (r *UserRepository) ExistsByUsername(ctx context.Context, username string) (bool, error) {
	query := `
		SELECT EXISTS (
			SELECT 1
			FROM users
			WHERE username = $1
		)
	`
	var exists bool
	err := r.db.QueryRowContext(ctx, query, username).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check username existence: %w", err)
	}
	return exists, nil
}

// Insert는 새로운 사용자 엔티티를 데이터베이스에 추가합니다.
// 성공 시 nil을 반환하며, 삽입 실패(중복 키 등) 시 에러를 반환합니다.
func (r *UserRepository) Insert(ctx context.Context, user *domain.User) error {
	query := `
		INSERT INTO users (id, username, email, password_hash, status, subscription_tier, created_at, updated_at, last_login_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`
	_, err := r.db.ExecContext(ctx, query,
		user.ID,
		user.Username,
		user.Email,
		user.PasswordHash,
		user.Status,
		user.SubscriptionTier,
		user.CreatedAt,
		user.UpdatedAt,
		nullTime(user.LastLoginAt),
	)
	if err != nil {
		return fmt.Errorf("failed to insert user: %w", err)
	}
	return nil
}

// Update는 기존 사용자 엔티티를 갱신합니다.
// 상태, 비밀번호, 마지막 로그인 시간 등을 업데이트하며, 실패 시 에러를 반환합니다.
func (r *UserRepository) Update(ctx context.Context, user *domain.User) error {
	query := `
		UPDATE users
		SET username = $2, email = $3, password_hash = $4, status = $5, subscription_tier = $6,
			created_at = $7, updated_at = $8, last_login_at = $9
		WHERE id = $1
	`
	_, err := r.db.ExecContext(ctx, query,
		user.ID,
		user.Username,
		user.Email,
		user.PasswordHash,
		user.Status,
		user.SubscriptionTier,
		user.CreatedAt,
		user.UpdatedAt,
		nullTime(user.LastLoginAt),
	)
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}
	return nil
}

// Delete는 사용자 ID로 사용자 엔티티를 삭제합니다.
// 논리적 삭제(상태를 DELETED로 변경)로 구현하며, 실패 시 에러를 반환합니다.
func (r *UserRepository) Delete(ctx context.Context, id string) error {
	query := `
		UPDATE users
		SET status = $2, updated_at = $3
		WHERE id = $1
	`
	_, err := r.db.ExecContext(ctx, query,
		id,
		domain.UserStatusDeleted,
		time.Now(),
	)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}
	return nil
}

// nullTime은 time.Time을 sql.NullTime으로 변환합니다.
// 값이 없는 경우(제로 값) null로 처리합니다.
func nullTime(t time.Time) sql.NullTime {
	if t.IsZero() {
		return sql.NullTime{Valid: false}
	}
	return sql.NullTime{Time: t, Valid: true}
}
