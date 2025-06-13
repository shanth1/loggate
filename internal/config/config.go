package config

import (
	"log"

	"github.com/shanth1/gotools/conf"
)

type Server struct {
	ListenAddress  string `mapstructure:"listen_address"`
	MetricsAddress string `mapstructure:"metrics_address"`
}

type Storage struct {
	Type      string   `mapstructure:"type"`
	Enabled   bool     `mapstructure:"enabled"`
	DSN       string   `mapstructure:"dsn,omitempty"`
	Addresses []string `mapstructure:"addresses,omitempty"`
	URL       string   `mapstructure:"url,omitempty"`
	Token     string   `mapstructure:"token,omitempty"`
	Org       string   `mapstructure:"org,omitempty"`
	Bucket    string   `mapstructure:"bucket,omitempty"`
}

type Config struct {
	Server   Server    `mapstructure:"server"`
	Storages []Storage `mapstructure:"storages"`
}

func MustGetConfig() *Config {
	cfg := &Config{}
	if err := conf.Load(conf.GetConfigPath(), cfg); err != nil {
		log.Fatalf("load config: %v", err)
	}

	return cfg
}
