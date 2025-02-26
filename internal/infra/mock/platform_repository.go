package mock

import (
	"context"
	"fmt"
	"sync"

	"github.com/sukryu/IV-auth-services/internal/domain"
)

// PlatformRepository는 테스트용 Mock 플랫폼 계정 저장소를 구현합니다.
// domain/repo.PlatformRepository 인터페이스를 만족하며, 메모리 기반 데이터를 관리합니다.
type PlatformRepository struct {
	accounts map[string]*domain.PlatformAccount
	mu       sync.RWMutex
}

// NewPlatformRepository는 새로운 Mock PlatformRepository 인스턴스를 생성하여 반환합니다.
func NewPlatformRepository() *PlatformRepository {
	return &PlatformRepository{
		accounts: make(map[string]*domain.PlatformAccount),
	}
}

// FindByID는 플랫폼 계정 ID로 계정을 조회합니다.
func (r *PlatformRepository) FindByID(ctx context.Context, id string) (*domain.PlatformAccount, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	account, exists := r.accounts[id]
	if !exists {
		return nil, nil // 계정 없음
	}
	return account, nil
}

// Insert는 새로운 플랫폼 계정을 메모리에 추가합니다.
func (r *PlatformRepository) Insert(ctx context.Context, account *domain.PlatformAccount) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.accounts[account.ID]; exists {
		return fmt.Errorf("platform account already exists")
	}
	r.accounts[account.ID] = account
	return nil
}

// Update는 기존 플랫폼 계정을 갱신합니다.
func (r *PlatformRepository) Update(ctx context.Context, account *domain.PlatformAccount) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.accounts[account.ID]; !exists {
		return domain.ErrPlatformAccountNotFound
	}
	r.accounts[account.ID] = account
	return nil
}

// Delete는 플랫폼 계정 ID로 계정을 삭제합니다.
func (r *PlatformRepository) Delete(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.accounts[id]; !exists {
		return domain.ErrPlatformAccountNotFound
	}
	delete(r.accounts, id)
	return nil
}
