package redis

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/sukryu/IV-auth-services/internal/config"
)

// Client는 Redis 연결을 관리하는 구조체입니다.
// go-redis 라이브러리를 사용하며, 설정을 통해 초기화됩니다.
type Client struct {
	client *redis.Client
}

// NewClient는 새로운 Redis 클라이언트를 생성하여 반환합니다.
// config.Config에서 Redis 설정을 받아 연결을 설정합니다.
func NewClient(cfg *config.Config) *Client {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Addr, // 예: "localhost:6379"
		Password: "",             // 비밀번호 없음 (환경에 따라 설정)
		DB:       cfg.Redis.DB,   // 기본 DB 번호 (예: 0)
	})

	// 연결 테스트
	ctx := context.Background()
	if err := client.Ping(ctx).Err(); err != nil {
		log.Fatalf("Failed to connect to Redis at %s: %v", cfg.Redis.Addr, err)
	}

	log.Printf("Connected to Redis at %s", cfg.Redis.Addr)
	return &Client{client: client}
}

// Close는 Redis 클라이언트를 안전하게 종료합니다.
// 리소스 정리를 위해 호출됩니다.
func (c *Client) Close() error {
	return c.client.Close()
}

// Set은 키-값 쌍을 Redis에 저장합니다.
// 만료 시간(TTL)을 설정할 수 있으며, 실패 시 에러를 반환합니다.
func (c *Client) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	if err := c.client.Set(ctx, key, value, expiration).Err(); err != nil {
		return fmt.Errorf("failed to set key %s in redis: %w", key, err)
	}
	return nil
}

// Get은 Redis에서 키에 해당하는 값을 조회합니다.
// 키가 없으면 빈 문자열과 nil 에러를 반환하며, 조회 실패 시 에러를 반환합니다.
func (c *Client) Get(ctx context.Context, key string) (string, error) {
	val, err := c.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return "", nil // 키 없음
	}
	if err != nil {
		return "", fmt.Errorf("failed to get key %s from redis: %w", key, err)
	}
	return val, nil
}

// Exists는 Redis에 키가 존재하는지 확인합니다.
// 존재 여부를 bool로 반환하며, 실패 시 에러를 반환합니다.
func (c *Client) Exists(ctx context.Context, key string) (bool, error) {
	count, err := c.client.Exists(ctx, key).Result()
	if err != nil {
		return false, fmt.Errorf("failed to check existence of key %s in redis: %w", key, err)
	}
	return count > 0, nil
}

// Del은 Redis에서 키를 삭제합니다.
// 성공 시 nil을 반환하며, 실패 시 에러를 반환합니다.
func (c *Client) Del(ctx context.Context, key string) error {
	if err := c.client.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("failed to delete key %s from redis: %w", key, err)
	}
	return nil
}
