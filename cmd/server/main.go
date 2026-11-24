// Command server is the process entrypoint: it loads configuration, builds the
// JSON logger, sets up OpenTelemetry, serves HTTP with conservative timeouts and
// drains gracefully on SIGINT/SIGTERM. It also implements -healthcheck (used by
// the container HEALTHCHECK) and -version.
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/aircloud/template-backend-go/internal/config"
	"github.com/aircloud/template-backend-go/internal/handlers"
	"github.com/aircloud/template-backend-go/internal/httpserver"
	"github.com/aircloud/template-backend-go/internal/observability"
)

// Build metadata, stamped via -ldflags "-X main.version=... ".
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

// run parses flags and dispatches to the requested mode, returning the exit code.
func run(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("server", flag.ContinueOnError)
	fs.SetOutput(stderr)
	healthcheck := fs.Bool("healthcheck", false, "probe the local /healthz endpoint and exit")
	showVersion := fs.Bool("version", false, "print build information and exit")
	if err := fs.Parse(args); err != nil {
		return 2
	}

	if *showVersion {
		_, _ = fmt.Fprintf(stdout, "version=%s commit=%s date=%s\n", version, commit, date)
		return 0
	}

	cfg, err := config.Load()
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "configuration error: %v\n", err)
		return 1
	}

	if *healthcheck {
		if err := probe(cfg); err != nil {
			_, _ = fmt.Fprintf(stderr, "healthcheck failed: %v\n", err)
			return 1
		}
		return 0
	}

	logger := newLogger(cfg.LogLevel, stdout).With(
		slog.String("service", cfg.ServiceName),
		slog.String("version", version),
		slog.String("environment", cfg.Environment),
	)
	slog.SetDefault(logger)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	if err := serve(ctx, cfg, logger); err != nil {
		logger.Error("server stopped with error", slog.String("error", err.Error()))
		return 1
	}
	return 0
}

// serve runs the HTTP server until ctx is cancelled, then drains it.
func serve(ctx context.Context, cfg config.Config, logger *slog.Logger) error {
	shutdownTracing, err := observability.SetupTracing(ctx, observability.Config{
		Enabled:     cfg.OTelEnabled,
		Endpoint:    cfg.OTelExporterEndpoint,
		ServiceName: cfg.ServiceName,
		Environment: cfg.Environment,
		Version:     version,
		SampleRatio: cfg.OTelSampleRatio,
	})
	if err != nil {
		return fmt.Errorf("setup tracing: %w", err)
	}

	health := handlers.NewHealth(version)
	srv := &http.Server{
		Addr: cfg.Addr(),
		Handler: httpserver.New(httpserver.Options{
			Logger:      logger,
			Version:     version,
			ServiceName: cfg.ServiceName,
			OTelEnabled: cfg.OTelEnabled,
			Health:      health,
		}),
		ReadHeaderTimeout: cfg.ReadHeaderTimeout,
		ReadTimeout:       cfg.ReadTimeout,
		WriteTimeout:      cfg.WriteTimeout,
		IdleTimeout:       cfg.IdleTimeout,
		MaxHeaderBytes:    cfg.MaxHeaderBytes,
		ErrorLog:          slog.NewLogLogger(logger.Handler(), slog.LevelError),
		BaseContext:       func(net.Listener) context.Context { return ctx },
	}

	serverErr := make(chan error, 1)
	go func() {
		logger.Info("http server listening", slog.String("addr", srv.Addr))
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErr <- err
		}
		close(serverErr)
	}()

	select {
	case err := <-serverErr:
		return errors.Join(err, shutdownTracing(context.WithoutCancel(ctx)))
	case <-ctx.Done():
	}

	logger.Info("shutdown signal received, draining")
	health.SetReady(false)

	shutdownCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), cfg.ShutdownTimeout)
	defer cancel()

	return errors.Join(srv.Shutdown(shutdownCtx), shutdownTracing(shutdownCtx), <-serverErr)
}

// probe performs a GET against the local liveness endpoint.
func probe(cfg config.Config) error {
	host := cfg.Host
	if host == "" || host == "0.0.0.0" || host == "::" {
		host = "127.0.0.1"
	}
	url := "http://" + net.JoinHostPort(host, strconv.Itoa(cfg.Port)) + "/healthz"

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, http.NoBody)
	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("request %s: %w", url, err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status %d from %s", resp.StatusCode, url)
	}
	return nil
}

// newLogger builds a JSON slog logger at the configured level.
func newLogger(level string, w io.Writer) *slog.Logger {
	var lvl slog.Level
	if err := lvl.UnmarshalText([]byte(level)); err != nil {
		lvl = slog.LevelInfo
	}
	return slog.New(slog.NewJSONHandler(w, &slog.HandlerOptions{Level: lvl}))
}
