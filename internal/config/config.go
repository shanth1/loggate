package config

type Config struct {
	App                 string              `mapstructure:"app"`
	Service             string              `mapstructure:"service"`
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
