package main

import (
	"log/slog"
	"net/http"
	"os"
	"time"
)

const defaultPort = "8080"

func main() {
	addr := ":" + port()
	srv := &http.Server{
		Addr:              addr,
		Handler:           newMux(),
		ReadHeaderTimeout: 5 * time.Second,
	}

	slog.Info("server listening", "addr", addr)
	err := srv.ListenAndServe()
	slog.Error("server stopped", "err", err)
	os.Exit(1)
}

func port() string {
	if p := os.Getenv("PORT"); p != "" {
		return p
	}
	return defaultPort
}

func newMux() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", healthz)
	return mux
}

func healthz(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write([]byte(`{"status":"ok"}`))
}
