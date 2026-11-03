package main

import (
	"context"
	"flag"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/ranonbezerra/seezle-calculator/backend/internal/httpapi"
)

func main() {
	checkHealth := flag.Bool("healthcheck", false, "check /healthz of the local server, exit 0 if healthy and 1 otherwise")
	flag.Parse()

	cfg, err := loadConfig(os.Getenv)
	if err != nil {
		fatal("load config", err)
	}
	if *checkHealth {
		if err := healthcheck(healthcheckURL(cfg.Addr), healthcheckTimeout); err != nil {
			fatal("healthcheck", err)
		}
		return
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	srv := newServer(cfg, httpapi.NewHandler(slog.Default()))

	ln, err := net.Listen("tcp", cfg.Addr)
	if err != nil {
		fatal("listen", err)
	}
	slog.Info("server listening", "addr", ln.Addr().String())

	if err := run(ctx, srv, ln, cfg.ShutdownTimeout); err != nil {
		fatal("server stopped", err)
	}
	slog.Info("server stopped")
}

func fatal(msg string, err error) {
	slog.Error(msg, "err", err)
	os.Exit(1)
}
