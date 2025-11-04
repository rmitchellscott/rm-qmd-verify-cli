package config

import (
	"os"
	"strings"
)

const (
	DefaultHost = "http://localhost:8080"
	EnvVarHost  = "QMDVERIFY_HOST"
)

type Config struct {
	ServerHost string
}

func Load() *Config {
	host := os.Getenv(EnvVarHost)
	if host == "" {
		host = DefaultHost
	}

	host = strings.TrimSuffix(host, "/")

	return &Config{
		ServerHost: host,
	}
}

func (c *Config) APIEndpoint(path string) string {
	return c.ServerHost + path
}
