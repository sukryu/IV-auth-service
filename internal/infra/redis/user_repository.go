package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/sukryu/IV-auth-services/internal/domain"
	"github.com/sukryu/IV-auth-services/internal/domain/repo"
)

// UserRepository는 Redis 캐싱을 사용한 사용자 데이터 접근을 구현합니다.
// domain/repo.UserRepository 인터페이스를 만족하며, 캐시와 DB를 조합하여 실시간 조회를 최적화합니다.
type UserRepository struct {
	client       *Client             // Redis 클라이언트
	fallbackRepo repo.UserRepository // DB 폴백 저장소 (PostgreSQL 등)
	cacheTTL     time.Duration       // 캐시 만료 시간 (기본 5분)
}

// NewUserRepository는 새로운 UserRepository 인스턴스를 생성하여 반환합니다.
// Redis 클라이언트와 폴백 저장소를 주입받아 초기화하며, 캐시 TTL을 설정합니다.
func NewUserRepository(client *Client, fallbackRepo repo.UserRepository, cacheTTL time.Duration) *UserRepository {
	if cacheTTL == 0 {
		cacheTTL = 5 * time.Minute // 기본 TTL 5분
	}
	return &UserRepository{
		client:       client,
		fallbackRepo: fallbackRepo,
		cacheTTL:     cacheTTL,
	}
}

// FindByUsername은 사용자명을 기준으로 사용자 엔티티를 조회합니다.
// Redis 캐시에서 먼저 확인하고, 미스 시 폴백 저장소에서 가져와 캐싱합니다.
func (r *UserRepository) FindByUsername(ctx context.Context, username string) (*domain.User, error) {
	key := fmt.Sprintf("user:username:%s", username)

	// Redis에서 조회
	cached, err := r.client.Get(ctx, key)
	if err != nil {
		// Redis 오류 시 폴백 저장소로 바로 이동
		return r.fallbackRepo.FindByUsername(ctx, username)
	}
	if cached != "" {
		var user domain.User
		if err := json.Unmarshal([]byte(cached), &user); err != nil {
			log.Printf("Failed to unmarshal cached user %s: %v", username, err)
		} else {
			return &user, nil // 캐시 히트
		}
	}

	// 캐시 미스 시 폴백 저장소 조회
	user, err := r.fallbackRepo.FindByUsername(ctx, username)
	if err != nil {
		return nil, fmt.Errorf("failed to find user by username in fallback: %w", err)
	}
	if user != nil {
		// 캐싱
		data, err := json.Marshal(user)
		if err == nil {
			if err := r.client.Set(ctx, key, string(data), r.cacheTTL); err != nil {
				log.Printf("Failed to cache user %s: %v", username, err)
			}
		}
	}
	return user, nil
}

// FindByID는 사용자 ID를 기준으로 사용자 엔티티를 조회합니다.
// Redis 캐시에서 먼저 확인하고, 미스 시 폴백 저장소에서 가져와 캐싱합니다.
func (r *UserRepository) FindByID(ctx context.Context, id string) (*domain.User, error) {
	key := fmt.Sprintf("user:id:%s", id)

	// Redis에서 조회
	cached, err := r.client.Get(ctx, key)
	if err != nil {
		return r.fallbackRepo.FindByID(ctx, id)
	}
	if cached != "" {
		var user domain.User
		if err := json.Unmarshal([]byte(cached), &user); err != nil {
			log.Printf("Failed to unmarshal cached user %s: %v", id, err)
		} else {
			return &user, nil // 캐시 히트
		}
	}

	// 캐시 미스 시 폴백 저장소 조회
	user, err := r.fallbackRepo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to find user by id in fallback: %w", err)
	}
	if user != nil {
		// 캐싱
		data, err := json.Marshal(user)
		if err == nil {
			if err := r.client.Set(ctx, key, string(data), r.cacheTTL); err != nil {
				log.Printf("Failed to cache user %s: %v", id, err)
			}
		}
	}
	return user, nil
}

// ExistsByUsername은 주어진 사용자명이 이미 존재하는지 확인합니다.
// 캐시 사용 없이 폴백 저장소만 조회하며, 캐싱 레이어에서는 불필요.
func (r *UserRepository) ExistsByUsername(ctx context.Context, username string) (bool, error) {
	return r.fallbackRepo.ExistsByUsername(ctx, username)
}

// Insert는 새로운 사용자 엔티티를 데이터베이스에 추가합니다.
// 삽입 후 캐시 무효화는 생략 (업데이트 빈도 낮음).
func (r *UserRepository) Insert(ctx context.Context, user *domain.User) error {
	if err := r.fallbackRepo.Insert(ctx, user); err != nil {
		return fmt.Errorf("failed to insert user in fallback: %w", err)
	}
	return nil
}

// Update는 기존 사용자 엔티티를 갱신합니다.
// 업데이트 후 관련 캐시를 무효화합니다.
func (r *UserRepository) Update(ctx context.Context, user *domain.User) error {
	if err := r.fallbackRepo.Update(ctx, user); err != nil {
		return fmt.Errorf("failed to update user in fallback: %w", err)
	}

	// 캐시 무효화
	keys := []string{
		fmt.Sprintf("user:id:%s", user.ID),
		fmt.Sprintf("user:username:%s", user.Username),
	}
	for _, key := range keys {
		if err := r.client.Del(ctx, key); err != nil {
			log.Printf("Failed to invalidate cache for %s: %v", key, err)
		}
	}
	return nil
}

// Delete는 사용자 ID로 사용자 엔티티를 삭제합니다.
// 삭제 후 관련 캐시를 무효화합니다.
func (r *UserRepository) Delete(ctx context.Context, id string) error {
	user, err := r.fallbackRepo.FindByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to find user for delete in fallback: %w", err)
	}
	if user == nil {
		return domain.ErrUserNotFound
	}

	if err := r.fallbackRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete user in fallback: %w", err)
	}

	// 캐시 무효화
	keys := []string{
		fmt.Sprintf("user:id:%s", id),
		fmt.Sprintf("user:username:%s", user.Username),
	}
	for _, key := range keys {
		if err := r.client.Del(ctx, key); err != nil {
			log.Printf("Failed to invalidate cache for %s: %v", key, err)
		}
	}
	return nil
}
