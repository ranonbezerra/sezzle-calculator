package main

import (
	"fmt"
	"strconv"
	"time"
)

const defaultPort = "8080"

// config holds the server settings. Only the port is configurable, through
// the PORT environment variable.
type config struct {
	Addr              string
	ReadHeaderTimeout time.Duration
	ReadTimeout       time.Duration
	WriteTimeout      time.Duration
	IdleTimeout       time.Duration
	ShutdownTimeout   time.Duration
}

// loadConfig builds the config from environment variables read via getenv.
func loadConfig(getenv func(string) string) (config, error) {
	port := getenv("PORT")
	if port == "" {
		port = defaultPort
	}
	if n, err := strconv.Atoi(port); err != nil || n < 1 || n > 65535 {
		return config{}, fmt.Errorf("invalid PORT %q: must be an integer between 1 and 65535", port)
	}

	return config{
		Addr:              ":" + port,
		ReadHeaderTimeout: 2 * time.Second,
		ReadTimeout:       5 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
		ShutdownTimeout:   10 * time.Second,
	}, nil
}
