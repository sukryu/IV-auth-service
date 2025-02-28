package main

import (
	"github.com/sukryu/IV-auth-services/internal/config"
	"github.com/sukryu/IV-auth-services/pkg/logger"
	"go.uber.org/zap"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		panic("Failed to load config: " + err.Error())
	}

	log, err := logger.NewLogger(cfg.Environment)
	if err != nil {
		panic("Failed to initialize logger: " + err.Error())
	}
	defer log.Sync()

	log.Info("Service starting",
		zap.String("env", cfg.Environment),
		zap.Int("port", cfg.Server.Port),
		zap.String("db_dsn", cfg.GetDSN()))
}
