// Package observability wires OpenTelemetry tracing for the service. It exports
// spans over OTLP/HTTP when enabled and otherwise installs a no-op provider so
// callers never have to nil-check the tracer.
package observability

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
)

// Config controls tracer setup.
type Config struct {
	Enabled     bool
	Endpoint    string
	ServiceName string
	Environment string
	Version     string
	SampleRatio float64
}

// ShutdownFunc flushes and releases tracing resources.
type ShutdownFunc func(context.Context) error

// SetupTracing configures the global tracer provider and text-map propagators.
//
// When cfg.Enabled is false a no-op shutdown is returned and the default no-op
// provider remains installed, so instrumented code keeps working with zero
// overhead. The returned ShutdownFunc must be called on exit.
func SetupTracing(ctx context.Context, cfg Config) (ShutdownFunc, error) {
	noop := func(context.Context) error { return nil }

	// Always install W3C trace-context + baggage propagation so inbound
	// headers are honoured even when local exporting is disabled.
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	if !cfg.Enabled {
		return noop, nil
	}

	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceName(cfg.ServiceName),
			semconv.ServiceVersion(cfg.Version),
			semconv.DeploymentEnvironment(cfg.Environment),
		),
		resource.WithFromEnv(),
		resource.WithTelemetrySDK(),
	)
	if err != nil {
		return noop, fmt.Errorf("build otel resource: %w", err)
	}

	exporter, err := otlptracehttp.New(ctx,
		otlptracehttp.WithEndpoint(cfg.Endpoint),
		otlptracehttp.WithInsecure(),
	)
	if err != nil {
		return noop, fmt.Errorf("create otlp exporter: %w", err)
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sdktrace.ParentBased(sdktrace.TraceIDRatioBased(cfg.SampleRatio))),
		sdktrace.WithBatcher(exporter,
			sdktrace.WithBatchTimeout(5*time.Second),
		),
	)
	otel.SetTracerProvider(tp)

	return func(shutdownCtx context.Context) error {
		return errors.Join(tp.ForceFlush(shutdownCtx), tp.Shutdown(shutdownCtx))
	}, nil
}
