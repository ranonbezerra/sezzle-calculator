package main

import (
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/ranonbezerra/seezle-calculator/backend/internal/httpapi"
)

func TestHealthcheckURL(t *testing.T) {
	if got, want := healthcheckURL(":8080"), "http://localhost:8080/healthz"; got != want {
		t.Errorf("healthcheckURL() = %q, want %q", got, want)
	}
}

func TestHealthcheck(t *testing.T) {
	tests := []struct {
		name    string
		handler http.HandlerFunc
		wantErr bool
	}{
		{
			name:    "healthy",
			handler: func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) },
		},
		{
			name:    "unhealthy status",
			handler: func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusServiceUnavailable) },
			wantErr: true,
		},
		{
			name:    "not found",
			handler: http.NotFound,
			wantErr: true,
		},
		{
			name: "slower than timeout",
			handler: func(w http.ResponseWriter, r *http.Request) {
				select {
				case <-r.Context().Done():
				case <-time.After(time.Second):
				}
				w.WriteHeader(http.StatusOK)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := httptest.NewServer(tt.handler)
			defer srv.Close()

			err := healthcheck(srv.URL, 50*time.Millisecond)

			if (err != nil) != tt.wantErr {
				t.Errorf("healthcheck() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestHealthcheckServerDown(t *testing.T) {
	srv := httptest.NewServer(http.NotFoundHandler())
	srv.Close()

	if err := healthcheck(srv.URL, time.Second); err == nil {
		t.Error("healthcheck() error = nil, want a connection error")
	}
}

func TestHealthcheckAgainstAPI(t *testing.T) {
	srv := httptest.NewServer(httpapi.NewHandler(slog.New(slog.DiscardHandler)))
	defer srv.Close()
	u, err := url.Parse(srv.URL)
	if err != nil {
		t.Fatalf("parse server URL: %v", err)
	}

	if err := healthcheck(healthcheckURL(":"+u.Port()), time.Second); err != nil {
		t.Errorf("healthcheck() error = %v, want nil", err)
	}
}
