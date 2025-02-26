package database

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/sukryu/IV-auth-services/internal/domain"
)

// TokenRepository는 PostgreSQL을 사용한 토큰 블랙리스트 관리 접근을 구현합니다.
// domain/repo.TokenRepository 인터페이스를 만족하며, 블랙리스트 CRUD 작업을 제공합니다.
type TokenRepository struct {
	db *sql.DB
}

// NewTokenRepository는 새로운 TokenRepository 인스턴스를 생성하여 반환합니다.
// DB 연결 객체를 주입받아 초기화합니다.
func NewTokenRepository(db *sql.DB) *TokenRepository {
	return &TokenRepository{db: db}
}

// IsBlacklisted는 주어진 토큰이 블랙리스트에 존재하는지 확인합니다.
// 존재 여부를 bool로 반환하며, 쿼리 실패 시 에러를 반환합니다.
func (r *TokenRepository) IsBlacklisted(ctx context.Context, tokenID string) (bool, error) {
	query := `
		SELECT EXISTS (
			SELECT 1
			FROM token_blacklist
			WHERE token_id = $1 AND expires_at > NOW()
		)
	`
	var exists bool
	err := r.db.QueryRowContext(ctx, query, tokenID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check token blacklist: %w", err)
	}
	return exists, nil
}

// AddToBlacklist는 토큰을 블랙리스트에 추가합니다.
// 성공 시 nil을 반환하며, 삽입 실패 시 에러를 반환합니다.
func (r *TokenRepository) AddToBlacklist(ctx context.Context, entry *domain.TokenBlacklist) error {
	query := `
		INSERT INTO token_blacklist (token_id, user_id, expires_at, reason, blacklisted_at)
		VALUES ($1, $2, $3, $4, $5)
	`
	_, err := r.db.ExecContext(ctx, query,
		entry.TokenID,
		entry.UserID,
		entry.ExpiresAt,
		entry.Reason,
		entry.BlacklistedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to add token to blacklist: %w", err)
	}
	return nil
}

// RemoveFromBlacklist는 블랙리스트에서 특정 토큰을 제거합니다.
// 성공 시 nil을 반환하며, 제거 실패 시 에러를 반환합니다 (미존재 시 무시).
func (r *TokenRepository) RemoveFromBlacklist(ctx context.Context, tokenID string) error {
	query := `
		DELETE FROM token_blacklist
		WHERE token_id = $1
	`
	_, err := r.db.ExecContext(ctx, query, tokenID)
	if err != nil {
		return fmt.Errorf("failed to remove token from blacklist: %w", err)
	}
	return nil
}

// CleanupExpired는 만료된 블랙리스트 항목을 정리합니다.
// 주기적 실행(예: cron)을 위해 설계되었으며, 실패 시 에러를 반환합니다.
func (r *TokenRepository) CleanupExpired(ctx context.Context) error {
	query := `
		DELETE FROM token_blacklist
		WHERE expires_at <= NOW()
	`
	_, err := r.db.ExecContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to cleanup expired blacklist entries: %w", err)
	}
	return nil
}
