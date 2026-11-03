package main

import (
	"testing"
	"time"
)

func TestLoadConfig(t *testing.T) {
	tests := []struct {
		name     string
		port     string
		wantAddr string
		wantErr  bool
	}{
		{name: "default port", port: "", wantAddr: ":8080"},
		{name: "custom port", port: "3000", wantAddr: ":3000"},
		{name: "lowest port", port: "1", wantAddr: ":1"},
		{name: "highest port", port: "65535", wantAddr: ":65535"},
		{name: "not a number", port: "abc", wantErr: true},
		{name: "zero", port: "0", wantErr: true},
		{name: "too large", port: "65536", wantErr: true},
		{name: "negative", port: "-1", wantErr: true},
		{name: "host included", port: "localhost:80", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			getenv := func(key string) string {
				if key == "PORT" {
					return tt.port
				}
				return ""
			}

			cfg, err := loadConfig(getenv)

			if tt.wantErr {
				if err == nil {
					t.Fatalf("loadConfig() error = nil, want an error (config %+v)", cfg)
				}
				return
			}
			if err != nil {
				t.Fatalf("loadConfig() error = %v", err)
			}
			if cfg.Addr != tt.wantAddr {
				t.Errorf("Addr = %q, want %q", cfg.Addr, tt.wantAddr)
			}
		})
	}
}

func TestLoadConfigDefaults(t *testing.T) {
	cfg, err := loadConfig(func(string) string { return "" })
	if err != nil {
		t.Fatalf("loadConfig() error = %v", err)
	}

	want := config{
		Addr:              ":8080",
		ReadHeaderTimeout: 2 * time.Second,
		ReadTimeout:       5 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
		ShutdownTimeout:   10 * time.Second,
	}
	if cfg != want {
		t.Errorf("loadConfig() = %+v, want %+v", cfg, want)
	}
}
