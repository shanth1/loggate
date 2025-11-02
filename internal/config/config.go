package config

import (
	"log"

	"github.com/shanth1/gotools/conf"
)

func MustGetConfig() *Config {
	cfg := &Config{}
	if err := conf.Load(conf.GetConfigPath(), cfg); err != nil {
		log.Fatalf("load config: %v", err)
	}

	return cfg
}

type Config struct {
	Server              *Server             `mapstructure:"server"`
	Performance         *Performance        `mapstructure:"performance"`
	Storages            map[string]*Storage `mapstructure:"storages"`
	RoutingRules        []*RoutingRule      `mapstructure:"routing_rules"`
	DefaultDestinations []string            `mapstructure:"default_destinations"`
}

type Server struct {
	LogAddress  string `mapstructure:"log_address"`
	InfoAddress string `mapstructure:"info_address"`
}

type Performance struct {
	BufferSize     int `mapstructure:"buffer_size"`
	BatchSize      int `mapstructure:"batch_size"`
	BatchTimeoutMs int `mapstructure:"batch_timeout_ms"`
}

type Storage struct {
	Type    string `mapstructure:"type"`
	Enabled bool   `mapstructure:"enabled"`
	DSN     string `mapstructure:"dsn,omitempty"`
}

type RoutingRule struct {
	MatchCondition *MatchCondition `mapstructure:"match_condition"`
	Destinations   []string        `mapstructure:"destinations"`
}

type MatchCondition struct {
	Service string `mapstructure:"service"`
	Level   string `mapstructure:"level"`
}
