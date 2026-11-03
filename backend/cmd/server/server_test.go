package main

import (
	"context"
	"errors"
	"net"
	"net/http"
	"testing"
	"time"
)

func TestNewServer(t *testing.T) {
	cfg := config{
		Addr:              ":9999",
		ReadHeaderTimeout: 1 * time.Second,
		ReadTimeout:       2 * time.Second,
		WriteTimeout:      3 * time.Second,
		IdleTimeout:       4 * time.Second,
	}
	handler := http.NotFoundHandler()

	srv := newServer(cfg, handler)

	if srv.Addr != cfg.Addr ||
		srv.ReadHeaderTimeout != cfg.ReadHeaderTimeout ||
		srv.ReadTimeout != cfg.ReadTimeout ||
		srv.WriteTimeout != cfg.WriteTimeout ||
		srv.IdleTimeout != cfg.IdleTimeout {
		t.Errorf("newServer() = %+v, want settings from %+v", srv, cfg)
	}
	if srv.Handler == nil {
		t.Error("Handler is nil")
	}
}

func listen(t *testing.T) net.Listener {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	return ln
}

// waitUntilClosed blocks until addr refuses connections, which shows that
// shutdown has started.
func waitUntilClosed(t *testing.T, addr string) {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		conn, err := net.Dial("tcp", addr)
		if err != nil {
			return
		}
		_ = conn.Close()
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatal("listener still open after cancelling the context")
}

func TestRunServesUntilContextIsDone(t *testing.T) {
	ln := listen(t)
	srv := &http.Server{
		Handler:           http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusNoContent) }),
		ReadHeaderTimeout: time.Second,
	}
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- run(ctx, srv, ln, time.Second) }()

	resp, err := http.Get("http://" + ln.Addr().String())
	if err != nil {
		t.Fatalf("GET: %v", err)
	}
	_ = resp.Body.Close()
	if resp.StatusCode != http.StatusNoContent {
		t.Errorf("status = %d, want 204", resp.StatusCode)
	}

	cancel()
	if err := <-done; err != nil {
		t.Errorf("run() = %v, want nil after graceful shutdown", err)
	}
	if _, err := http.Get("http://" + ln.Addr().String()); err == nil {
		t.Error("server still accepts connections after shutdown")
	}
}

func TestRunWaitsForInFlightRequests(t *testing.T) {
	ln := listen(t)
	started := make(chan struct{})
	release := make(chan struct{})
	srv := &http.Server{
		Handler: http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			close(started)
			<-release
			w.WriteHeader(http.StatusOK)
		}),
		ReadHeaderTimeout: time.Second,
	}
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- run(ctx, srv, ln, 5*time.Second) }()

	respStatus := make(chan int, 1)
	go func() {
		resp, err := http.Get("http://" + ln.Addr().String())
		if err != nil {
			respStatus <- 0
			return
		}
		_ = resp.Body.Close()
		respStatus <- resp.StatusCode
	}()

	<-started
	cancel()
	waitUntilClosed(t, ln.Addr().String())
	close(release)

	if got := <-respStatus; got != http.StatusOK {
		t.Errorf("in-flight request status = %d, want 200", got)
	}
	if err := <-done; err != nil {
		t.Errorf("run() = %v, want nil", err)
	}
}

func TestRunShutdownTimeout(t *testing.T) {
	ln := listen(t)
	started := make(chan struct{})
	release := make(chan struct{})
	defer close(release)
	srv := &http.Server{
		Handler: http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
			close(started)
			<-release
		}),
		ReadHeaderTimeout: time.Second,
	}
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- run(ctx, srv, ln, 10*time.Millisecond) }()

	go func() {
		if resp, err := http.Get("http://" + ln.Addr().String()); err == nil {
			_ = resp.Body.Close()
		}
	}()

	<-started
	cancel()
	if err := <-done; !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("run() = %v, want context.DeadlineExceeded", err)
	}
}

func TestRunReturnsServeError(t *testing.T) {
	ln := listen(t)
	_ = ln.Close()
	srv := &http.Server{ReadHeaderTimeout: time.Second}

	err := run(context.Background(), srv, ln, time.Second)

	if err == nil || errors.Is(err, http.ErrServerClosed) {
		t.Errorf("run() = %v, want a listener error", err)
	}
}
