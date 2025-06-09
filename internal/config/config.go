package config

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
