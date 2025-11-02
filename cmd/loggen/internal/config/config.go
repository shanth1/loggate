package config

import (
	"time"

	"github.com/spf13/viper"
)

// TODO: changed to gotools

type Config struct {
	Target    string
	Load      loadConfig
	Templates []TemplateConfig
}

type loadConfig struct {
	Workers int
	RPS     int
	Jitter  time.Duration
}

type TemplateConfig struct {
	App      string
	Service  string
	Levels   map[string]float64
	Messages map[string][]string
	Fields   []fieldConfig
}

type fieldConfig struct {
	Key  string
	Type string
}

func Load(path string) (*Config, error) {
	viper.SetConfigFile(path)
	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
