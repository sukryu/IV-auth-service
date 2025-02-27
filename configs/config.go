package configs

import (
	"fmt"
	"os"

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
		Addr string
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

	// Set config file
	v.SetConfigFile(configPath)
	v.SetConfigType("yaml")

	// Read config file
	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Unmarshal into struct
	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &cfg, nil
}

// Main entry point variable (to be set via ldflags)
var configPath string

func init() {
	// Default config path if not set via ldflags
	if configPath == "" {
		configPath = "configs/config.yaml"
	}

	// Load config at startup
	cfg, err := LoadConfig(configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Config load failed: %v\n", err)
		os.Exit(1)
	}
	// Global config variable (to be used throughout the app)
	GlobalConfig = cfg
}

// GlobalConfig is thesingleton instance of Config
var GlobalConfig *Config
