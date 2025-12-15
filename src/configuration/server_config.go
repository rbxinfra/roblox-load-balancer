package configuration

import "time"

// ServerConfig is a configuration for a server.
type ServerConfig struct {
	// Interval is the health check interval.
	Interval time.Duration `json:"interval" yaml:"interval" toml:"interval"`

	// Rise is the number of successful checks a server
	// needs in order to be considered healthy.
	Rise int `json:"rise" yaml:"rise" toml:"rise"`

	// Fall is the number of failing checks a server
	// needs in order to be considered failing.
	Fall int `json:"fall" yaml:"fall" toml:"fall"`
}

// ServersConfig is the servers configuration
type ServersConfig struct {
	// Default is the default server config.
	Default *ServerConfig `json:"default" yaml:"default" toml:"default"`

	// PerServer is the per server config.
	PerServer *ServerConfig `json:"perServer" yaml:"per_server" toml:"per_server"`
}
