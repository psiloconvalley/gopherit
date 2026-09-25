package main

import (
	"context"
	"errors"
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/psiloconvalley/gopherit/internal/ui"
)

// version is injected at link time:
// go build -ldflags="-X main.version=abc1234" ./cmd/server
var version = "1.0.0"

func main() {
	startTime := time.Now().UTC()

	// ── Structured Logging (JSON on stdout) ─────────────
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	// ── Environment Configuration ───────────────────────
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// ── Sub-Filesystems ─────────────────────────────────
	staticFS, err := fs.Sub(ui.FS, "static")
	if err != nil {
		logger.Error("failed to create static sub-filesystem", "error", err)
		os.Exit(1)
	}

	templatesFS, err := fs.Sub(ui.FS, "templates")
	if err != nil {
		logger.Error("failed to create templates sub-filesystem", "error", err)
		os.Exit(1)
	}

	// ── Pre-compile Template Cache ──────────────────────
	tmplCache, err := newTemplateCache(templatesFS)
	if err != nil {
		logger.Error("failed to compile template cache", "error", err)
		os.Exit(1)
	}
	logger.Info("template cache initialized successfully")

	// ── Router Setup (Go 1.22+ Exact-Match Syntax) ──────
	mux := http.NewServeMux()

	// Static assets file server
	fileServer := http.FileServer(http.FS(staticFS))
	mux.Handle("GET /static/", http.StripPrefix("/static/", fileServer))

	// Routes
	mux.HandleFunc("GET /{$}", homeHandler(tmplCache, logger, version))
	mux.HandleFunc("GET /healthz", healthHandler(version, startTime))
	mux.HandleFunc("GET /api/gopher", gopherHandler(logger))

	// 404 Catch-All
	mux.HandleFunc("GET /", notFoundHandler(tmplCache, logger, version))

	// ── Middleware Pipeline ─────────────────────────────
	// Pipeline executes in order: Security Headers -> Logging -> Recovery -> Mux
	var handler http.Handler = mux
	handler = recovery(logger)(handler)
	handler = requestLogger(logger)(handler)
	handler = securityHeaders(handler)

	// ── HTTP Server ─────────────────────────────────────
	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      handler,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// ── Graceful Shutdown Listener ──────────────────────
	ctx, stop := signal.NotifyContext(context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stop()

	go func() {
		logger.Info("server starting", "port", port, "version", version)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("server fatal error", "error", err)
			os.Exit(1)
		}
	}()

	// Block until a termination signal is received
	<-ctx.Done()
	logger.Info("shutdown signal received, closing active connections")

	// 10-second drain window
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Error("forced shutdown error", "error", err)
		os.Exit(1)
	}

	logger.Info("server exited cleanly")
}
