package redis

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/sukryu/IV-auth-services/internal/domain"
	"github.com/sukryu/IV-auth-services/internal/domain/repo"
)

// TokenRepository는 Redis를 사용한 토큰 블랙리스트 관리 접근을 구현합니다.
// domain/repo.TokenRepository 인터페이스를 만족하며, 캐싱 기반 빠른 조회를 제공합니다.
type TokenRepository struct {
	client       *Client              // Redis 클라이언트
	fallbackRepo repo.TokenRepository // DB 폴백 저장소 (PostgreSQL 등)
}

// NewTokenRepository는 새로운 TokenRepository 인스턴스를 생성하여 반환합니다.
// Redis 클라이언트와 폴백 저장소를 주입받아 초기화합니다.
func NewTokenRepository(client *Client, fallbackRepo repo.TokenRepository) *TokenRepository {
	return &TokenRepository{
		client:       client,
		fallbackRepo: fallbackRepo,
	}
}

// IsBlacklisted는 주어진 토큰이 블랙리스트에 존재하는지 확인합니다.
// Redis에서 먼저 확인하고, 캐시 미스 시 폴백 저장소를 조회합니다.
func (r *TokenRepository) IsBlacklisted(ctx context.Context, tokenID string) (bool, error) {
	// Redis에서 조회
	key := fmt.Sprintf("blacklist:%s", tokenID)
	exists, err := r.client.Exists(ctx, key)
	if err != nil {
		// Redis 오류 시 폴백 저장소로 조회
		return r.fallbackRepo.IsBlacklisted(ctx, tokenID)
	}
	if exists {
		return true, nil // 캐시 히트, 블랙리스트에 존재
	}

	// 캐시 미스 시 폴백 저장소 확인
	isBlacklisted, err := r.fallbackRepo.IsBlacklisted(ctx, tokenID)
	if err != nil {
		return false, fmt.Errorf("failed to check token blacklist in fallback: %w", err)
	}
	if isBlacklisted {
		// 폴백에서 블랙리스트 확인 시 Redis에 캐싱 (만료 시간은 DB 기준)
		// 실제 만료 시간은 DB에서 가져와야 함, 여기선 예시로 24시간
		if err := r.client.Set(ctx, key, "1", 24*time.Hour); err != nil {
			// 캐싱 실패는 로그만 남기고 무시 (최적화 문제일 뿐)
			log.Printf("Failed to cache blacklisted token %s: %v", tokenID, err)
		}
	}
	return isBlacklisted, nil
}

// AddToBlacklist는 토큰을 블랙리스트에 추가합니다.
// Redis와 폴백 저장소 모두에 저장하며, 실패 시 에러를 반환합니다.
func (r *TokenRepository) AddToBlacklist(ctx context.Context, entry *domain.TokenBlacklist) error {
	// 폴백 저장소에 먼저 저장
	if err := r.fallbackRepo.AddToBlacklist(ctx, entry); err != nil {
		return fmt.Errorf("failed to add token to fallback blacklist: %w", err)
	}

	// Redis에 캐싱 (TTL은 토큰 만료 시간까지)
	key := fmt.Sprintf("blacklist:%s", entry.TokenID)
	expiration := time.Until(entry.ExpiresAt)
	if err := r.client.Set(ctx, key, "1", expiration); err != nil {
		// 캐싱 실패는 로그만 남기고 무시 (DB에는 저장됨)
		log.Printf("Failed to cache blacklisted token %s: %v", entry.TokenID, err)
	}
	return nil
}

// RemoveFromBlacklist는 블랙리스트에서 특정 토큰을 제거합니다.
// Redis와 폴백 저장소 모두에서 삭제하며, 실패 시 에러를 반환합니다.
func (r *TokenRepository) RemoveFromBlacklist(ctx context.Context, tokenID string) error {
	// Redis에서 삭제
	key := fmt.Sprintf("blacklist:%s", tokenID)
	if err := r.client.Del(ctx, key); err != nil {
		log.Printf("Failed to remove token %s from redis blacklist: %v", tokenID, err)
		// Redis 오류는 무시하고 DB 작업 계속
	}

	// 폴백 저장소에서 삭제
	if err := r.fallbackRepo.RemoveFromBlacklist(ctx, tokenID); err != nil {
		return fmt.Errorf("failed to remove token from fallback blacklist: %w", err)
	}
	return nil
}

// CleanupExpired는 만료된 블랙리스트 항목을 정리합니다.
// Redis는 TTL로 자동 정리되므로 폴백 저장소만 처리하며, 실패 시 에러를 반환합니다.
func (r *TokenRepository) CleanupExpired(ctx context.Context) error {
	return r.fallbackRepo.CleanupExpired(ctx)
}
