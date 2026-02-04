package config

import (
	"github.com/spf13/viper"
)

type Config struct {
	Server    ServerConfig    `mapstructure:"server"`
	Database  DatabaseConfig  `mapstructure:"database"`
	Access    AccessConfig    `mapstructure:"access"`
	Forward   ForwardConfig   `mapstructure:"forward"`
	Logging   LoggingConfig   `mapstructure:"logging"`
	Heartbeat HeartbeatConfig `mapstructure:"heartbeat"`
}

type ServerConfig struct {
	Host string `mapstructure:"host"`
	Port int    `mapstructure:"port"`
	Mode string `mapstructure:"mode"`
}

type DatabaseConfig struct {
	Path string `mapstructure:"path"`
}

type AccessConfig struct {
	MaxConnections int `mapstructure:"max_connections"`
}

type ForwardConfig struct {
	MaxQueueSize int `mapstructure:"max_queue_size"`
	WorkerCount  int `mapstructure:"worker_count"`
}

type LoggingConfig struct {
	Level string `mapstructure:"level"`
	File  string `mapstructure:"file"`
}

type HeartbeatConfig struct {
	CheckInterval  int `mapstructure:"check_interval"`  // In seconds
	DefaultTimeout int `mapstructure:"default_timeout"` // In seconds
}

func Load() (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./config")

	viper.SetDefault("server.host", "0.0.0.0")
	viper.SetDefault("server.port", 8080)
	viper.SetDefault("server.mode", "debug")
	viper.SetDefault("database.path", "./data/platform.db")
	viper.SetDefault("access.max_connections", 10000)
	viper.SetDefault("forward.max_queue_size", 10000)
	viper.SetDefault("forward.worker_count", 4)
	viper.SetDefault("logging.level", "info")
	viper.SetDefault("logging.file", "./logs/platform.log")
	viper.SetDefault("heartbeat.check_interval", 60)
	viper.SetDefault("heartbeat.default_timeout", 300)

	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
