package main

import (
	"context"
	"net"
	"net/http"
	"time"
)

// newServer returns an http.Server for handler with the timeouts from cfg.
func newServer(cfg config, handler http.Handler) *http.Server {
	return &http.Server{
		Addr:              cfg.Addr,
		Handler:           handler,
		ReadHeaderTimeout: cfg.ReadHeaderTimeout,
		ReadTimeout:       cfg.ReadTimeout,
		WriteTimeout:      cfg.WriteTimeout,
		IdleTimeout:       cfg.IdleTimeout,
	}
}

// run serves on ln until ctx is done, then shuts down gracefully, waiting up
// to shutdownTimeout for in-flight requests. It returns early if serving fails.
func run(ctx context.Context, srv *http.Server, ln net.Listener, shutdownTimeout time.Duration) error {
	serveErr := make(chan error, 1)
	go func() { serveErr <- srv.Serve(ln) }()

	select {
	case err := <-serveErr:
		return err
	case <-ctx.Done():
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()
	return srv.Shutdown(shutdownCtx)
}
