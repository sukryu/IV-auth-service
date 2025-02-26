package repo

import (
	"context"

	"github.com/sukryu/IV-auth-services/internal/domain"
)

// TokenRepository는 토큰 블랙리스트 관리에 대한 접근을 정의하는 인터페이스입니다.
// Redis 또는 PostgreSQL과 같은 저장소와의 상호작용을 추상화하며, 토큰 무효화 관련 작업을 제공합니다.
type TokenRepository interface {
	// IsBlacklisted는 주어진 토큰이 블랙리스트에 존재하는지 확인합니다.
	// 존재 여부를 bool로 반환하며, 쿼리 실패 시 에러를 반환합니다.
	IsBlacklisted(ctx context.Context, tokenID string) (bool, error)

	// AddToBlacklist는 토큰을 블랙리스트에 추가합니다.
	// 성공 시 nil을 반환하며, 삽입 실패 시(예: 중복 키) 에러를 반환합니다.
	AddToBlacklist(ctx context.Context, entry *domain.TokenBlacklist) error

	// RemoveFromBlacklist는 블랙리스트에서 특정 토큰을 제거합니다.
	// 성공 시 nil을 반환하며, 제거 실패 또는 토큰 미존재 시 에러를 반환하지 않고 무시합니다.
	RemoveFromBlacklist(ctx context.Context, tokenID string) error

	// CleanupExpired는 만료된 블랙리스트 항목을 정리합니다.
	// 주기적 실행(예: cron)을 위해 설계되었으며, 실패 시 에러를 반환합니다.
	CleanupExpired(ctx context.Context) error
}
