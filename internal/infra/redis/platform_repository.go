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

// PlatformRepository는 Redis 캐싱을 사용한 플랫폼 계정 데이터 접근을 구현합니다.
// domain/repo.PlatformRepository 인터페이스를 만족하며, 캐시와 DB를 조합하여 실시간 조회를 최적화합니다.
type PlatformRepository struct {
	client       *Client                 // Redis 클라이언트
	fallbackRepo repo.PlatformRepository // DB 폴백 저장소 (PostgreSQL 등)
	cacheTTL     time.Duration           // 캐시 만료 시간 (기본 5분)
}

// NewPlatformRepository는 새로운 PlatformRepository 인스턴스를 생성하여 반환합니다.
// Redis 클라이언트와 폴백 저장소를 주입받아 초기화하며, 캐시 TTL을 설정합니다.
func NewPlatformRepository(client *Client, fallbackRepo repo.PlatformRepository, cacheTTL time.Duration) *PlatformRepository {
	if cacheTTL == 0 {
		cacheTTL = 5 * time.Minute // 기본 TTL 5분
	}
	return &PlatformRepository{
		client:       client,
		fallbackRepo: fallbackRepo,
		cacheTTL:     cacheTTL,
	}
}

// FindByID는 플랫폼 계정 ID로 계정을 조회합니다.
// Redis 캐시에서 먼저 확인하고, 미스 시 폴백 저장소에서 가져와 캐싱합니다.
func (r *PlatformRepository) FindByID(ctx context.Context, id string) (*domain.PlatformAccount, error) {
	key := fmt.Sprintf("platform:id:%s", id)

	// Redis에서 조회
	cached, err := r.client.Get(ctx, key)
	if err != nil {
		return r.fallbackRepo.FindByID(ctx, id)
	}
	if cached != "" {
		var account domain.PlatformAccount
		if err := json.Unmarshal([]byte(cached), &account); err != nil {
			log.Printf("Failed to unmarshal cached platform account %s: %v", id, err)
		} else {
			return &account, nil // 캐시 히트
		}
	}

	// 캐시 미스 시 폴백 저장소 조회
	account, err := r.fallbackRepo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to find platform account by id in fallback: %w", err)
	}
	if account != nil {
		// 캐싱
		data, err := json.Marshal(account)
		if err == nil {
			if err := r.client.Set(ctx, key, string(data), r.cacheTTL); err != nil {
				log.Printf("Failed to cache platform account %s: %v", id, err)
			}
		}
	}
	return account, nil
}

// Insert는 새로운 플랫폼 계정을 데이터베이스에 추가합니다.
// 삽입 후 캐시는 별도 무효화 없이 진행 (업데이트 빈도 낮음).
func (r *PlatformRepository) Insert(ctx context.Context, account *domain.PlatformAccount) error {
	if err := r.fallbackRepo.Insert(ctx, account); err != nil {
		return fmt.Errorf("failed to insert platform account in fallback: %w", err)
	}
	return nil
}

// Update는 기존 플랫폼 계정을 갱신합니다.
// 업데이트 후 캐시를 무효화합니다.
func (r *PlatformRepository) Update(ctx context.Context, account *domain.PlatformAccount) error {
	if err := r.fallbackRepo.Update(ctx, account); err != nil {
		return fmt.Errorf("failed to update platform account in fallback: %w", err)
	}

	// 캐시 무효화
	key := fmt.Sprintf("platform:id:%s", account.ID)
	if err := r.client.Del(ctx, key); err != nil {
		log.Printf("Failed to invalidate cache for %s: %v", key, err)
	}
	return nil
}

// Delete는 플랫폼 계정 ID로 계정을 삭제합니다.
// 삭제 후 캐시를 무효화합니다.
func (r *PlatformRepository) Delete(ctx context.Context, id string) error {
	if err := r.fallbackRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete platform account in fallback: %w", err)
	}

	// 캐시 무효화
	key := fmt.Sprintf("platform:id:%s", id)
	if err := r.client.Del(ctx, key); err != nil {
		log.Printf("Failed to invalidate cache for %s: %v", key, err)
	}
	return nil
}
