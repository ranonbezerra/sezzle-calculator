package main

import (
	"context"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/ranonbezerra/seezle-calculator/backend/internal/httpapi"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	cfg, err := loadConfig(os.Getenv)
	if err != nil {
		fatal("load config", err)
	}
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
