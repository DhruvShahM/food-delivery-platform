package config

import (
	"fmt"
	"os"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Config struct {
	GRPCPort     string   `mapstructure:"grpc_port"`
	HTTPPort     string   `mapstructure:"http_port"`
	DBURL        string   `mapstructure:"db_url"`
	
	// Individual DB components for environment variable override
	DBHost       string   `mapstructure:"db_host" env:"DB_HOST"`
	DBPort       string   `mapstructure:"db_port" env:"DB_PORT"`
	DBUser       string   `mapstructure:"db_user" env:"DB_USER"`
	DBPassword   string   `mapstructure:"db_password" env:"DB_PASSWORD"`
	DBName       string   `mapstructure:"db_name" env:"DB_NAME"`
	
	JWTSecret    string   `mapstructure:"jwt_secret"`
	KafkaBrokers []string `mapstructure:"kafka_brokers"`
	LogLevel     string   `mapstructure:"log_level"`
	Logger       *zap.Logger
}

func LoadConfig() (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./config")
	viper.AddConfigPath("../config")
	
	// Enable environment variable reading
	viper.AutomaticEnv()
	
	// Set environment variable prefixes for better organization
	viper.SetEnvPrefix("FOOD_DELIVERY") // Optional: FOOD_DELIVERY_DB_HOST, etc.
	
	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	cfg := &Config{}
	if err := viper.Unmarshal(cfg); err != nil {
		return nil, err
	}

	// Construct DB URL from environment variables if available
	cfg.DBURL = buildDBURL(cfg)

	level := zapcore.InfoLevel
	if cfg.LogLevel == "debug" {
		level = zapcore.DebugLevel
	}
	lcfg := zap.NewProductionConfig()
	lcfg.Level.SetLevel(level)
	l, err := lcfg.Build()
	if err != nil {
		return nil, err
	}
	cfg.Logger = l
	return cfg, nil
}

// buildDBURL constructs database URL from individual components
// Priority: Environment variables > YAML config > defaults
func buildDBURL(cfg *Config) string {
	host := getEnvOrDefault("DB_HOST", cfg.DBHost, "localhost")
	port := getEnvOrDefault("DB_PORT", cfg.DBPort, "5432")
	user := getEnvOrDefault("DB_USER", cfg.DBUser, "root")
	password := getEnvOrDefault("DB_PASSWORD", cfg.DBPassword, "root")
	dbname := getEnvOrDefault("DB_NAME", cfg.DBName, "fooddb")

	// If full DB_URL is provided via environment, use it
	if fullURL := os.Getenv("DATABASE_URL"); fullURL != "" {
		return fullURL
	}

	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		user, password, host, port, dbname)
}

// getEnvOrDefault gets value from environment, then config, then default
func getEnvOrDefault(envKey, configValue, defaultValue string) string {
	if envValue := os.Getenv(envKey); envValue != "" {
		return envValue
	}
	if configValue != "" {
		return configValue
	}
	return defaultValue
}