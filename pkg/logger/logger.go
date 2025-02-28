package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Logger is a wrapper around zap.Logger for custom configuration.
type Logger struct {
	zap *zap.Logger
}

// NewLogger initializes a production-ready logger based on the environment.
func NewLogger(env string) (*Logger, error) {
	var config zap.Config
	if env == "development" {
		// 개발 환경: 콘솔 출력, Debug 레벨
		config = zap.Config{
			Level:            zap.NewAtomicLevelAt(zap.DebugLevel),
			Development:      true,
			Encoding:         "console",
			EncoderConfig:    zap.NewDevelopmentEncoderConfig(),
			OutputPaths:      []string{"stdout"},
			ErrorOutputPaths: []string{"stderr"},
		}
	} else {
		// 프로덕션/스테이징: JSON 출력, Info 레벨, 파일 로깅 추가
		config = zap.Config{
			Level:            zap.NewAtomicLevelAt(zap.InfoLevel),
			Development:      false,
			Encoding:         "json",
			EncoderConfig:    zap.NewProductionEncoderConfig(),
			OutputPaths:      []string{"stdout", "/var/log/iv-auth-service.log"},
			ErrorOutputPaths: []string{"stderr"},
		}
	}

	// 커스텀 시간 포맷
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	// 로거 생성
	zapLogger, err := config.Build(zap.AddCaller(), zap.AddStacktrace(zap.ErrorLevel))
	if err != nil {
		return nil, err
	}

	return &Logger{zap: zapLogger}, nil
}

// Sync flushes any buffered log entries.
func (l *Logger) Sync() error {
	return l.zap.Sync()
}

// Debug logs a debug-level message.
func (l *Logger) Debug(msg string, fields ...zap.Field) {
	l.zap.Debug(msg, fields...)
}

// Info logs an info-level message.
func (l *Logger) Info(msg string, fields ...zap.Field) {
	l.zap.Info(msg, fields...)
}

// Warn logs a warning-level message.
func (l *Logger) Warn(msg string, fields ...zap.Field) {
	l.zap.Warn(msg, fields...)
}

// Error logs an error-level message.
func (l *Logger) Error(msg string, fields ...zap.Field) {
	l.zap.Error(msg, fields...)
}

// Fatal logs a fatal-level message and exits the program.
func (l *Logger) Fatal(msg string, fields ...zap.Field) {
	l.zap.Fatal(msg, fields...)
}

// With creates a child logger with additional fields.
func (l *Logger) With(fields ...zap.Field) *Logger {
	return &Logger{zap: l.zap.With(fields...)}
}
