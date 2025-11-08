package config

import (
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Config struct {
	GRPCPort    string   `mapstructure:"grpc_port"`
	HTTPPort    string   `mapstructure:"http_port"`
	DBURL       string   `mapstructure:"db_url"`
	KafkaBrokers []string `mapstructure:"kafka_brokers"`
	KafkaTopic  string   `mapstructure:"kafka_topic"`
	LogLevel    string   `mapstructure:"log_level"`
	Logger      *zap.Logger
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