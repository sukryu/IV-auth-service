package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

// Config holds all configuration values for the service.
type Config struct {
	Environment string `mapstructure:"environment"`
	Server      struct {
		Port int `mapstructure:"port"`
	} `mapstructure:"server"`
	Database struct {
		Host     string `mapstructure:"host"`
		Port     int    `mapstructure:"port"`
		User     string `mapstructure:"user"`
		Password string `mapstructure:"password"`
		Name     string `mapstructure:"name"`
	} `mapstructure:"database"`
	Redis struct {
		Addr string `mapstructure:"addr"`
	} `mapstructure:"redis"`
	Kafka struct {
		Broker string `mapstructure:"broker"`
	} `mapstructure:"kafka"`
	JWT struct {
		PrivateKeyPath string `mapstructure:"private_key_path"`
		PublicKeyPath  string `mapstructure:"public_key_path"`
	} `mapstructure:"jwt"`
}

// LoadConfig loads configuration from environment variables and config file.
func LoadConfig() (*Config, error) {
	v := viper.New()

	// 환경 변수 설정
	v.SetEnvPrefix("IV_AUTH") // 예: IV_AUTH_DB_HOST
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// 기본 설정 파일
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath(".")
	v.AddConfigPath("./config/")

	// 기본값 설정
	v.SetDefault("environment", "development")
	v.SetDefault("server.port", 50051)
	v.SetDefault("database.host", "localhost")
	v.SetDefault("database.port", 5432)
	v.SetDefault("database.user", "auth_user")
	v.SetDefault("database.password", "auth_password")
	v.SetDefault("database.name", "auth_db")
	v.SetDefault("redis.addr", "localhost:6379")
	v.SetDefault("kafka.broker", "localhost:9092")
	v.SetDefault("jwt.private_key_path", "./certs/private.pem")
	v.SetDefault("jwt.public_key_path", "./certs/public.pem")

	// 설정 파일 읽기 (없으면 기본값 사용)
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
	}

	// 환경 변수로 오버라이드
	if err := v.BindEnv("database.password", "DB_PASS"); err != nil {
		return nil, fmt.Errorf("failed to bind env: %w", err)
	}

	// 구조체로 언마샬
	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// 필수 값 검증
	if cfg.Database.Password == "" {
		return nil, fmt.Errorf("database password is required")
	}

	return &cfg, nil
}

// GetDSN returns the PostgreSQL connection string.
func (c *Config) GetDSN() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable",
		c.Database.User, c.Database.Password, c.Database.Host, c.Database.Port, c.Database.Name)
}
