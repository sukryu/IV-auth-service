package config

import (
	"fmt"
	"log"
	"os"

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
	// Set the path to the configuration files directory.
	v.AddConfigPath("configs")
	// Set the file name based on the environment (e.g., dev.yaml or prod.yaml).
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
