package database

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/lib/pq"
	"github.com/sukryu/IV-auth-services/internal/config"
)

// NewPostgresDB는 PostgreSQL 데이터베이스에 연결하고 DB 객체를 반환합니다.
func NewPostgresDB(cfg *config.Config) (*sql.DB, error) {
	// 연결 문자열 구성
	connStr := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		cfg.Database.Host, cfg.Database.Port, cfg.Database.User,
		cfg.Database.Password, cfg.Database.Name,
	)

	// 데이터베이스 연결
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("PostgreSQL 연결 실패: %w", err)
	}

	// 연결 테스트
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("PostgreSQL 핑 테스트 실패: %w", err)
	}

	// 커넥션 풀 설정
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	log.Printf("PostgreSQL 데이터베이스(%s:%d/%s)에 성공적으로 연결되었습니다.",
		cfg.Database.Host, cfg.Database.Port, cfg.Database.Name)

	return db, nil
}
