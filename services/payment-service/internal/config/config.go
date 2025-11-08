package config

import (
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

type Config struct {
	GRPCPort string `mapstructure:"grpc_port"`
	HTTPPort string `mapstructure:"http_port"`
	DBURL    string `mapstructure:"db_url"`
	RedisURL string `mapstructure:"redis_url"`
	LogLevel string `mapstructure:"log_level"`
	Logger   *zap.Logger
}

func LoadConfig() (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./config")
	viper.AddConfigPath("../config")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	cfg := &Config{}
	if err := viper.Unmarshal(cfg); err != nil {
		return nil, err
	}

	cfg.Logger = zap.NewNop() // Simple logger for now
	return cfg, nil
}
