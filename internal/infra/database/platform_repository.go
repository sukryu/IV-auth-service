package database

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/sukryu/IV-auth-services/internal/domain"
)

// PlatformRepository는 PostgreSQL을 사용한 플랫폼 계정 데이터 접근을 구현합니다.
// domain/repo.PlatformRepository 인터페이스를 만족하며, CRUD 작업을 제공합니다.
type PlatformRepository struct {
	db *sql.DB
}

// NewPlatformRepository는 새로운 PlatformRepository 인스턴스를 생성하여 반환합니다.
// DB 연결 객체를 주입받아 초기화합니다.
func NewPlatformRepository(db *sql.DB) *PlatformRepository {
	return &PlatformRepository{db: db}
}

// FindByID는 플랫폼 계정 ID로 계정을 조회합니다.
// 계정이 없으면 nil과 nil 에러를 반환하며, 조회 실패 시 에러를 반환합니다.
func (r *PlatformRepository) FindByID(ctx context.Context, id string) (*domain.PlatformAccount, error) {
	query := `
		SELECT id, user_id, platform, platform_user_id, platform_username, access_token, refresh_token, token_expires_at, created_at, updated_at
		FROM platform_accounts
		WHERE id = $1
		LIMIT 1
	`
	row := r.db.QueryRowContext(ctx, query, id)

	account := &domain.PlatformAccount{}
	err := row.Scan(
		&account.ID,
		&account.UserID,
		&account.Platform,
		&account.PlatformUserID,
		&account.PlatformUsername,
		&account.AccessToken,
		&account.RefreshToken,
		&account.TokenExpiresAt,
		&account.CreatedAt,
		&account.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil // 계정 없음
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find platform account by id: %w", err)
	}
	return account, nil
}

// Insert는 새로운 플랫폼 계정을 데이터베이스에 추가합니다.
// 성공 시 nil을 반환하며, 삽입 실패(중복 키 등) 시 에러를 반환합니다.
func (r *PlatformRepository) Insert(ctx context.Context, account *domain.PlatformAccount) error {
	query := `
		INSERT INTO platform_accounts (id, user_id, platform, platform_user_id, platform_username, access_token, refresh_token, token_expires_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`
	_, err := r.db.ExecContext(ctx, query,
		account.ID,
		account.UserID,
		account.Platform,
		account.PlatformUserID,
		account.PlatformUsername,
		account.AccessToken,
		account.RefreshToken,
		account.TokenExpiresAt,
		account.CreatedAt,
		account.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to insert platform account: %w", err)
	}
	return nil
}

// Update는 기존 플랫폼 계정을 갱신합니다.
// 토큰 갱신 등을 처리하며, 실패 시 에러를 반환합니다.
func (r *PlatformRepository) Update(ctx context.Context, account *domain.PlatformAccount) error {
	query := `
		UPDATE platform_accounts
		SET user_id = $2, platform = $3, platform_user_id = $4, platform_username = $5,
			access_token = $6, refresh_token = $7, token_expires_at = $8, created_at = $9, updated_at = $10
		WHERE id = $1
	`
	_, err := r.db.ExecContext(ctx, query,
		account.ID,
		account.UserID,
		account.Platform,
		account.PlatformUserID,
		account.PlatformUsername,
		account.AccessToken,
		account.RefreshToken,
		account.TokenExpiresAt,
		account.CreatedAt,
		account.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to update platform account: %w", err)
	}
	return nil
}

// Delete는 플랫폼 계정 ID로 계정을 삭제합니다.
// 물리적 삭제로 구현하며, 실패 시 에러를 반환합니다.
func (r *PlatformRepository) Delete(ctx context.Context, id string) error {
	query := `
		DELETE FROM platform_accounts
		WHERE id = $1
	`
	_, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete platform account: %w", err)
	}
	return nil
}
