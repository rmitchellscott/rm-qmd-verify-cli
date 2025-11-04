package config

import (
	"testing"
)

func TestLoad(t *testing.T) {
	tests := []struct {
		name       string
		envValue   string
		setEnv     bool
		wantHost   string
	}{
		{
			name:     "no env var - uses default",
			setEnv:   false,
			wantHost: DefaultHost,
		},
		{
			name:     "custom env var",
			envValue: "https://custom.example.com",
			setEnv:   true,
			wantHost: "https://custom.example.com",
		},
		{
			name:     "env var with trailing slash",
			envValue: "https://example.com/",
			setEnv:   true,
			wantHost: "https://example.com",
		},
		{
			name:     "env var with multiple trailing slashes",
			envValue: "https://example.com///",
			setEnv:   true,
			wantHost: "https://example.com//",
		},
		{
			name:     "localhost",
			envValue: "http://localhost:8080",
			setEnv:   true,
			wantHost: "http://localhost:8080",
		},
		{
			name:     "localhost with trailing slash",
			envValue: "http://localhost:8080/",
			setEnv:   true,
			wantHost: "http://localhost:8080",
		},
		{
			name:     "empty env var - uses default",
			envValue: "",
			setEnv:   true,
			wantHost: DefaultHost,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setEnv {
				t.Setenv(EnvVarHost, tt.envValue)
			}

			cfg := Load()

			if cfg.ServerHost != tt.wantHost {
				t.Errorf("Load() ServerHost = %v, want %v", cfg.ServerHost, tt.wantHost)
			}
		})
	}
}

func TestConfig_APIEndpoint(t *testing.T) {
	tests := []struct {
		name       string
		serverHost string
		path       string
		want       string
	}{
		{
			name:       "path with leading slash",
			serverHost: "https://example.com",
			path:       "/api/test",
			want:       "https://example.com/api/test",
		},
		{
			name:       "path without leading slash",
			serverHost: "https://example.com",
			path:       "api/test",
			want:       "https://example.comapi/test",
		},
		{
			name:       "empty path",
			serverHost: "https://example.com",
			path:       "",
			want:       "https://example.com",
		},
		{
			name:       "root path",
			serverHost: "https://example.com",
			path:       "/",
			want:       "https://example.com/",
		},
		{
			name:       "complex path",
			serverHost: "https://example.com",
			path:       "/api/v1/resource/123",
			want:       "https://example.com/api/v1/resource/123",
		},
		{
			name:       "server with port",
			serverHost: "http://localhost:8080",
			path:       "/api/test",
			want:       "http://localhost:8080/api/test",
		},
		{
			name:       "server with path already",
			serverHost: "https://example.com/base",
			path:       "/api/test",
			want:       "https://example.com/base/api/test",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{
				ServerHost: tt.serverHost,
			}

			got := cfg.APIEndpoint(tt.path)

			if got != tt.want {
				t.Errorf("APIEndpoint() = %v, want %v", got, tt.want)
			}
		})
	}
}
