package main

import (
	"fmt"
	"net/http"
	"time"
)

const healthcheckTimeout = 2 * time.Second

// healthcheckURL returns the /healthz URL of a server listening locally on addr.
func healthcheckURL(addr string) string {
	return "http://localhost" + addr + "/healthz"
}

// healthcheck GETs url and returns an error unless it answers 200 OK within timeout.
func healthcheck(url string, timeout time.Duration) error {
	client := &http.Client{Timeout: timeout}
	resp, err := client.Get(url)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status %d", resp.StatusCode)
	}
	return nil
}
