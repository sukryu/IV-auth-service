package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/sukryu/IV-auth-services/internal/adapters/db/postgres"
	"github.com/sukryu/IV-auth-services/internal/config"
	"github.com/sukryu/IV-auth-services/pkg/logger"
	"go.uber.org/zap"
)

func main() {
	// Config 로드
	cfg, err := config.LoadConfig()
	if err != nil {
		panic("Failed to load config: " + err.Error()) // 초기화 실패 시 즉시 종료
	}

	// Logger 초기화
	log, err := logger.NewLogger(cfg.Environment)
	if err != nil {
		panic("Failed to initialize logger: " + err.Error())
	}
	defer func() {
		if err := log.Sync(); err != nil {
			// Sync 실패 시 stderr로 출력 (프로세스 종료 직전)
			_, _ = fmt.Fprintf(os.Stderr, "Failed to sync logger: %v\n", err)
		}
	}()

	// Database 연결
	db, err := postgres.NewDB(cfg, log)
	if err != nil {
		log.Fatal("Failed to initialize database connection", zap.Error(err))
	}
	defer db.Close()

	// 서비스 시작 로그
	log.Info("IV-auth-service started successfully",
		zap.String("environment", cfg.Environment),
		zap.Int("port", cfg.Server.Port))

	// graceful shutdown 대기
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	log.Info("Shutting down service")
}
