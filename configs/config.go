package configs

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/spf13/viper"
)

// Config holds all configuration values
type Config struct {
	DB struct {
		Host     string
		Port     int
		User     string
		Password string
		Name     string
		DSN      string
	}
	Redis struct {
		Addr     string
		Password string
	}
	Kafka struct {
		Broker string
	}
	JWT struct {
		PrivateKeyPath string
		PublicKeyPath  string
	}
}

// LoadConfig loads configuration from YAML file using Viper
func LoadConfig(configPath string) (*Config, error) {
	v := viper.New()
	v.SetConfigFile(configPath)
	v.SetConfigType("yaml")
	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}
	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}
	return &cfg, nil
}

// ConfigPath is the main entry point variable (to be set via ldflags or tests)
var ConfigPath string

func init() {
	if ConfigPath == "" {
		// 현재 파일(config.go)의 위치에서 프로젝트 루트(IV-auth-service) 찾기
		_, filename, _, ok := runtime.Caller(0)
		if !ok {
			fmt.Fprintf(os.Stderr, "Failed to determine project root\n")
			os.Exit(1)
		}
		// configs/config.go 기준으로 루트로 이동 (한 단계 상위)
		rootDir := filepath.Dir(filename) // IV-auth-service
		configPath := filepath.Join(rootDir, "config.yaml")

		// 경로가 IV-auth-service를 포함하는지 확인
		if !strings.Contains(rootDir, "IV-auth-service") {
			fmt.Fprintf(os.Stderr, "Project root does not contain IV-auth-service: %s\n", rootDir)
			os.Exit(1)
		}
		ConfigPath = configPath
	}

	cfg, err := LoadConfig(ConfigPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Config load failed: %v\n", err)
		os.Exit(1)
	}
	GlobalConfig = cfg
}

// GlobalConfig is the singleton instance of Config
var GlobalConfig *Config
