package mock

import (
	"context"
	"sync"
	"time"

	"github.com/sukryu/IV-auth-services/internal/domain"
)

// TokenRepository는 테스트용 Mock 토큰 블랙리스트 저장소를 구현합니다.
// domain/repo.TokenRepository 인터페이스를 만족하며, 메모리 기반 데이터를 관리합니다.
type TokenRepository struct {
	blacklist map[string]*domain.TokenBlacklist
	mu        sync.RWMutex
}

// NewTokenRepository는 새로운 Mock TokenRepository 인스턴스를 생성하여 반환합니다.
func NewTokenRepository() *TokenRepository {
	return &TokenRepository{
		blacklist: make(map[string]*domain.TokenBlacklist),
	}
}

// IsBlacklisted는 주어진 토큰이 블랙리스트에 존재하는지 확인합니다.
func (r *TokenRepository) IsBlacklisted(ctx context.Context, tokenID string) (bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	entry, exists := r.blacklist[tokenID]
	if !exists {
		return false, nil
	}
	return time.Now().Before(entry.ExpiresAt), nil // 만료 여부 확인
}

// AddToBlacklist는 토큰을 블랙리스트에 추가합니다.
func (r *TokenRepository) AddToBlacklist(ctx context.Context, entry *domain.TokenBlacklist) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.blacklist[entry.TokenID] = entry
	return nil
}

// RemoveFromBlacklist는 블랙리스트에서 특정 토큰을 제거합니다.
func (r *TokenRepository) RemoveFromBlacklist(ctx context.Context, tokenID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.blacklist, tokenID)
	return nil
}

// CleanupExpired는 만료된 블랙리스트 항목을 정리합니다.
func (r *TokenRepository) CleanupExpired(ctx context.Context) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	for tokenID, entry := range r.blacklist {
		if now.After(entry.ExpiresAt) {
			delete(r.blacklist, tokenID)
		}
	}
	return nil
}
