package postgres

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sukryu/IV-auth-services/internal/config"
	"github.com/sukryu/IV-auth-services/pkg/logger"
	"go.uber.org/zap"
)

// DB encapsulates the PostgreSQL connection pool with logging.
type DB struct {
	Pool   *pgxpool.Pool
	logger *logger.Logger
}

// NewDB creates a new database connection pool using the provided config and logger.
func NewDB(cfg *config.Config, log *logger.Logger) (*DB, error) {
	// DSN 생성
	dsn := cfg.GetDSN()
	log.Debug("Initializing database connection", zap.String("dsn", maskDSN(dsn)))

	// 연결 풀 설정
	poolConfig, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		log.Error("Failed to parse DSN", zap.Error(err))
		return nil, fmt.Errorf("failed to parse DSN: %w", err)
	}
	// 대규모 트래픽의 경우 100+
	poolConfig.MaxConns = 50 // 넉넉한 최대 연결 수 설정 (운영 환경에 따라 조정 가능)

	// 연결 풀 생성
	pool, err := pgxpool.NewWithConfig(context.Background(), poolConfig)
	if err != nil {
		log.Error("Failed to create database pool", zap.Error(err))
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// 연결 테스트
	if err := pool.Ping(context.Background()); err != nil {
		log.Error("Database ping failed", zap.Error(err))
		pool.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// 기본 쿼리로 연결 검증
	_, err = pool.Exec(context.Background(), "SELECT 1")
	if err != nil {
		log.Error("Test query failed", zap.Error(err))
		pool.Close()
		return nil, fmt.Errorf("failed to execute test query: %w", err)
	}

	log.Info("Database connection established successfully")
	return &DB{
		Pool:   pool,
		logger: log.With(zap.String("component", "postgres")),
	}, nil
}

// Close shuts down the database connection pool.
func (db *DB) Close() {
	db.logger.Info("Closing database connection pool")
	db.Pool.Close()
}

// maskDSN masks sensitive information in the DSN (username, password).
func maskDSN(dsn string) string {
	parsedURL, err := url.Parse(dsn)
	if err != nil {
		return "****" // 파싱 실패 시 전체 마스킹
	}

	// 사용자명과 패스워드 마스킹
	if parsedURL.User != nil {
		username := parsedURL.User.Username()
		if username != "" {
			parsedURL.User = url.UserPassword("****", "****")
		} else {
			parsedURL.User = url.User("****")
		}
	}

	// 쿼리 파라미터 중 민감 정보 마스킹 (예: sslmode는 유지)
	query := parsedURL.Query()
	for key := range query {
		if strings.Contains(strings.ToLower(key), "pass") || strings.Contains(strings.ToLower(key), "key") {
			query.Set(key, "****")
		}
	}
	parsedURL.RawQuery = query.Encode()

	return parsedURL.String()
}
