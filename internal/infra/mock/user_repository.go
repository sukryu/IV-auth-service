package mock

import (
	"context"
	"sync"

	"github.com/sukryu/IV-auth-services/internal/domain"
)

// UserRepository는 테스트용 Mock 사용자 저장소를 구현합니다.
// domain/repo.UserRepository 인터페이스를 만족하며, 메모리 기반 데이터를 관리합니다.
type UserRepository struct {
	users map[string]*domain.User
	mu    sync.RWMutex
}

// NewUserRepository는 새로운 Mock UserRepository 인스턴스를 생성하여 반환합니다.
func NewUserRepository() *UserRepository {
	return &UserRepository{
		users: make(map[string]*domain.User),
	}
}

// FindByUsername은 사용자명을 기준으로 사용자 엔티티를 조회합니다.
func (r *UserRepository) FindByUsername(ctx context.Context, username string) (*domain.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, user := range r.users {
		if user.Username == username {
			return user, nil
		}
	}
	return nil, nil // 사용자 없음
}

// FindByID는 사용자 ID를 기준으로 사용자 엔티티를 조회합니다.
func (r *UserRepository) FindByID(ctx context.Context, id string) (*domain.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	user, exists := r.users[id]
	if !exists {
		return nil, nil // 사용자 없음
	}
	return user, nil
}

// ExistsByUsername은 주어진 사용자명이 이미 존재하는지 확인합니다.
func (r *UserRepository) ExistsByUsername(ctx context.Context, username string) (bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, user := range r.users {
		if user.Username == username {
			return true, nil
		}
	}
	return false, nil
}

// Insert는 새로운 사용자 엔티티를 메모리에 추가합니다.
func (r *UserRepository) Insert(ctx context.Context, user *domain.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.users[user.ID]; exists {
		return domain.ErrUserAlreadyExists
	}
	r.users[user.ID] = user
	return nil
}

// Update는 기존 사용자 엔티티를 갱신합니다.
func (r *UserRepository) Update(ctx context.Context, user *domain.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.users[user.ID]; !exists {
		return domain.ErrUserNotFound
	}
	r.users[user.ID] = user
	return nil
}

// Delete는 사용자 ID로 사용자 엔티티를 삭제합니다.
func (r *UserRepository) Delete(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if user, exists := r.users[id]; !exists {
		return domain.ErrUserNotFound
	} else {
		user.Status = domain.UserStatusDeleted
		r.users[id] = user
	}
	return nil
}
