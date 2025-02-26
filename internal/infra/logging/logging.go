package logging

import (
	"fmt"
	"os"

	"github.com/sukryu/IV-auth-services/internal/config"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Logger는 zap 기반 구조화된 로깅을 제공하는 구조체입니다.
// 서비스 전반에서 로깅을 일관되게 처리하며, production-ready 수준의 설정을 지원합니다.
type Logger struct {
	logger *zap.Logger
}

// NewLogger는 새로운 Logger 인스턴스를 생성하여 반환합니다.
// config.Config에서 로깅 설정(레벨, 포맷)을 받아 초기화합니다.
func NewLogger(cfg *config.Config) *Logger {
	var level zapcore.Level
	switch cfg.Logging.Level {
	case "debug":
		level = zapcore.DebugLevel
	case "info":
		level = zapcore.InfoLevel
	case "warn":
		level = zapcore.WarnLevel
	case "error":
		level = zapcore.ErrorLevel
	default:
		level = zapcore.InfoLevel
	}

	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder

	var encoder zapcore.Encoder
	if cfg.Logging.Format == "json" {
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	} else {
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
	}

	core := zapcore.NewCore(
		encoder,
		zapcore.AddSync(os.Stdout),
		level,
	)

	logger := zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1))
	logger.Info("Logger initialized", zap.String("level", cfg.Logging.Level), zap.String("format", cfg.Logging.Format))
	return &Logger{logger: logger}
}

// Sync는 로거의 버퍼를 플러시하여 모든 로그를 출력합니다.
// graceful shutdown 시 호출됩니다.
func (l *Logger) Sync() error {
	return l.logger.Sync()
}

// Debug는 디버그 레벨 로그를 기록합니다.
// 개발 환경에서 상세 디버깅에 사용됩니다.
func (l *Logger) Debug(msg string, fields ...zap.Field) {
	l.logger.Debug(msg, fields...)
}

// Info는 정보 레벨 로그를 기록합니다.
// 일반적인 상태 정보를 출력합니다.
func (l *Logger) Info(msg string, fields ...zap.Field) {
	l.logger.Info(msg, fields...)
}

// Warn는 경고 레벨 로그를 기록합니다.
// 잠재적 문제를 알리는 데 사용됩니다.
func (l *Logger) Warn(msg string, fields ...zap.Field) {
	l.logger.Warn(msg, fields...)
}

// Error는 에러 레벨 로그를 기록합니다.
// 오류 상황을 기록하며, 문제 진단에 활용됩니다.
func (l *Logger) Error(msg string, fields ...zap.Field) {
	l.logger.Error(msg, fields...)
}

// Fatal는 치명적 에러 로그를 기록하고 프로세스를 종료합니다.
// 복구 불가능한 상황에서 호출됩니다.
func (l *Logger) Fatal(msg string, fields ...zap.Field) {
	l.logger.Fatal(msg, fields...)
}

// WithDebug는 디버그 레벨 로그를 기록하고 nil을 반환합니다.
// 성공적인 중간 상태를 디버깅 목적으로 로깅합니다.
func (l *Logger) WithDebug(msg string, fields ...zap.Field) {
	l.logger.Debug(msg, fields...)
}

// WithInfo는 정보 레벨 로그를 기록하고 nil을 반환합니다.
// 일반적인 상태 정보를 로깅하며, 호출자가 추가 작업을 이어갈 수 있습니다.
func (l *Logger) WithInfo(msg string, fields ...zap.Field) {
	l.logger.Info(msg, fields...)
}

// WithWarn는 경고 레벨 로그를 기록하고 nil을 반환합니다.
// 잠재적 문제를 알리며, 호출자가 상황에 따라 처리를 결정할 수 있습니다.
func (l *Logger) WithWarn(err error, msg string, fields ...zap.Field) error {
	if err == nil {
		return nil
	}
	fullFields := append(fields, zap.Error(err))
	l.logger.Warn(msg, fullFields...)
	return nil // 경고는 에러로 간주하지 않음
}

// WithError는 에러가 있을 경우 로깅하고 에러를 래핑하여 반환합니다.
// nil 에러는 무시하며, 호출자가 추가 컨텍스트를 제공할 수 있습니다.
func (l *Logger) WithError(err error, msg string, fields ...zap.Field) error {
	if err == nil {
		return nil
	}
	fullFields := append(fields, zap.Error(err))
	l.logger.Error(msg, fullFields...)
	return fmt.Errorf("%s: %w", msg, err)
}
