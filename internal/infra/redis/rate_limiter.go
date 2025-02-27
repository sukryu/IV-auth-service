package redis

import (
	"context"
	"fmt"
	"time"
)

// RateLimiter는 Redis를 활용한 속도 제한 기능을 제공합니다.
type RateLimiter struct {
	client    *Client
	limit     int           // 허용 요청 수
	ttl       time.Duration // 제한 기간
	keyPrefix string        // 키 접두사
}

// NewRateLimiter는 새로운 RateLimiter 인스턴스를 생성합니다.
func NewRateLimiter(client *Client, limit int, ttl time.Duration) *RateLimiter {
	return &RateLimiter{
		client:    client,
		limit:     limit,
		ttl:       ttl,
		keyPrefix: "rate-limit:",
	}
}

// Allow는 주어진 키에 대한 요청이 허용되는지 확인합니다.
// 요청이 초과되면 false와 에러를 반환합니다.
func (r *RateLimiter) Allow(ctx context.Context, key string) (bool, error) {
	rateKey := fmt.Sprintf("%s%s", r.keyPrefix, key)

	// Redis 트랜잭션으로 카운트 증가 및 TTL 설정
	pipe := r.client.client.TxPipeline()
	incr := pipe.Incr(ctx, rateKey)
	pipe.Expire(ctx, rateKey, r.ttl)
	_, err := pipe.Exec(ctx)
	if err != nil {
		return false, fmt.Errorf("failed to execute rate limit transaction: %w", err)
	}

	count := incr.Val()
	if count > int64(r.limit) {
		return false, fmt.Errorf("rate limit exceeded: %d requests in %s", r.limit, r.ttl)
	}
	return true, nil
}
