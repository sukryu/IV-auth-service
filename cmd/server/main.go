package main

import (
	"context"
	"fmt"
	"log"

	"github.com/sukryu/IV-auth-services/configs"
	"github.com/sukryu/IV-auth-services/internal/repository"
)

func main() {
	cfg := configs.GlobalConfig
	fmt.Printf("Loaded config: DB Host=%s, Redis Addr=%s\n", cfg.DB.Host, cfg.Redis.Addr)

	// PostgreSQL 연결 테스트
	ctx := context.Background()
	pg, err := repository.NewPostgres(ctx)
	if err != nil {
		log.Fatalf("Failed to connect to PostgreSQL: %v", err)
	}
	defer pg.Close()

	if err := pg.Ping(ctx); err != nil {
		log.Fatalf("Failed to ping PostgreSQL: %v", err)
	}
	log.Println("Successfully connected to PostgreSQL!")

	log.Println("Authentication Service starting...")
}
