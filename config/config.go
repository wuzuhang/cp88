package config

import (
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/spf13/viper"
)

type ServerConfig struct {
	Port int    `mapstructure:"port"`
	Mode string `mapstructure:"mode"`
}

type DatabaseConfig struct {
	Enabled     bool   `mapstructure:"enabled"`
	Driver      string `mapstructure:"driver"`
	DSN         string `mapstructure:"dsn"`
	AutoMigrate bool   `mapstructure:"auto_migrate"`
}

type Config struct {
	Fucai3DSourceURL string         `mapstructure:"fucai3d_source_url"`
	SSQSourceURL     string         `mapstructure:"source_url"`
	Server           ServerConfig   `mapstructure:"server"`
	Database         DatabaseConfig `mapstructure:"database"`
}

var AppConfig *Config

func InitConfig() error {
	v := viper.New()
	path := os.Getenv("SSQ_CONFIG")
	if path == "" {
		path = "config.yaml"
	}
	v.SetConfigFile(path)
	v.SetEnvPrefix("SSQ")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()
	v.SetDefault("server.port", 8080)
	v.SetDefault("server.mode", "debug")
	v.SetDefault("database.enabled", false)
	v.SetDefault("database.driver", "mysql")
	v.SetDefault("database.dsn", "")
	v.SetDefault("database.auto_migrate", false)
	v.SetDefault("source_url", "")
	v.SetDefault("fucai3d_source_url", "http://data.17500.cn/3d_asc.txt")
	if err := v.ReadInConfig(); err != nil {
		return fmt.Errorf("读取配置: %w", err)
	}
	cfg := &Config{}
	if err := v.Unmarshal(cfg); err != nil {
		return fmt.Errorf("解析配置: %w", err)
	}
	if cfg.Server.Port < 1 || cfg.Server.Port > 65535 {
		return fmt.Errorf("server.port 必须在 1–65535 之间")
	}
	switch cfg.Server.Mode {
	case "debug", "release", "test":
	default:
		return fmt.Errorf("server.mode 必须是 debug、release 或 test")
	}
	source, err := url.Parse(cfg.SSQSourceURL)
	if err != nil || source.Hostname() == "" || (source.Scheme != "http" && source.Scheme != "https") {
		return fmt.Errorf("source_url 必须是有效的 HTTP 或 HTTPS 地址")
	}
	threeDSource, err := url.Parse(cfg.Fucai3DSourceURL)
	if err != nil || threeDSource.Hostname() == "" || (threeDSource.Scheme != "http" && threeDSource.Scheme != "https") {
		return fmt.Errorf("fucai3d_source_url 必须是有效的 HTTP 或 HTTPS 地址")
	}
	AppConfig = cfg
	return nil
}
