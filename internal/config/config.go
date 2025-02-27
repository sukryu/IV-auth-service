package config

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

// Config holds the application configuration.
type Config struct {
	Env    string `mapstructure:"env"`
	Server struct {
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
		DB   int    `mapstructure:"db"`
	} `mapstructure:"redis"`
	Kafka struct {
		Broker      string `mapstructure:"broker"`
		TopicPrefix string `mapstructure:"topic_prefix"`
	} `mapstructure:"kafka"`
	JWT struct {
		PrivateKeyPath     string `mapstructure:"private_key_path"`
		PublicKeyPath      string `mapstructure:"public_key_path"`
		AccessTokenExpiry  string `mapstructure:"access_token_expiry"`
		RefreshTokenExpiry string `mapstructure:"refresh_token_expiry"`
	} `mapstructure:"jwt"`
	Logging struct {
		Level  string `mapstructure:"level"`
		Format string `mapstructure:"format"`
	} `mapstructure:"logging"`
}

// LoadConfig loads the configuration based on the provided environment ("dev" or "prod").
func LoadConfig(env string) (*Config, error) {
	v := viper.New()
	v.SetConfigType("yaml")

	// 환경 변수로 경로 지정 (선택)
	if configPath := os.Getenv("CONFIG_PATH"); configPath != "" {
		v.AddConfigPath(configPath)
	} else {
		// 기본 동적 계산
		dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
		if err != nil {
			return nil, fmt.Errorf("failed to get absolute directory: %w", err)
		}
		rootDir := dir
		for !fileExists(filepath.Join(rootDir, "configs")) && rootDir != "/" {
			rootDir = filepath.Dir(rootDir)
		}
		if rootDir == "/" {
			return nil, fmt.Errorf("failed to find project root directory containing 'configs'")
		}
		v.AddConfigPath(filepath.Join(rootDir, "configs"))
	}
	v.SetConfigName(env)

	// Read the configuration file.
	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := v.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &config, nil
}

// fileExists는 지정된 경로에 파일 또는 디렉토리가 존재하는지 확인합니다.
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

// GetEnvFromArgs retrieves the environment ("dev" or "prod") from command-line arguments.
// Default to "dev" if no valid argument is provided.
func GetEnvFromArgs() string {
	if len(os.Args) < 2 {
		log.Println("No environment argument provided; defaulting to 'dev'")
		return "dev"
	}
	env := os.Args[1]
	if env != "dev" && env != "prod" {
		log.Printf("Unknown environment '%s'; defaulting to 'dev'", env)
		return "dev"
	}
	return env
}
