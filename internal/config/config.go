// Package config loads and validates runtime configuration from the
// environment. All knobs are 12-factor friendly and have safe defaults so the
// service can boot locally without any environment variables set.
package config

import (
	"fmt"
	"net"
	"strconv"
	"time"

	"github.com/caarlos0/env/v11"
	"github.com/go-playground/validator/v10"
)

// Config is the fully-resolved application configuration.
type Config struct {
	// Networking.
	Host string `env:"HOST" envDefault:"0.0.0.0" validate:"required"`
	Port int    `env:"PORT" envDefault:"8080" validate:"required,min=1,max=65535"`

	// Service identity.
	ServiceName string `env:"SERVICE_NAME" envDefault:"template-backend-go" validate:"required"`
	Environment string `env:"ENVIRONMENT" envDefault:"development" validate:"required,oneof=development staging production"`

	// Logging.
	LogLevel string `env:"LOG_LEVEL" envDefault:"info" validate:"required,oneof=debug info warn error"`

	// HTTP server timeouts.
	ReadHeaderTimeout time.Duration `env:"READ_HEADER_TIMEOUT" envDefault:"5s" validate:"required"`
	ReadTimeout       time.Duration `env:"READ_TIMEOUT" envDefault:"15s" validate:"required"`
	WriteTimeout      time.Duration `env:"WRITE_TIMEOUT" envDefault:"15s" validate:"required"`
	IdleTimeout       time.Duration `env:"IDLE_TIMEOUT" envDefault:"60s" validate:"required"`
	ShutdownTimeout   time.Duration `env:"SHUTDOWN_TIMEOUT" envDefault:"20s" validate:"required"`
	MaxHeaderBytes    int           `env:"MAX_HEADER_BYTES" envDefault:"1048576" validate:"required,min=4096"`

	// OpenTelemetry.
	OTelEnabled          bool    `env:"OTEL_ENABLED" envDefault:"false"`
	OTelExporterEndpoint string  `env:"OTEL_EXPORTER_OTLP_ENDPOINT" envDefault:"localhost:4318"`
	OTelSampleRatio      float64 `env:"OTEL_TRACES_SAMPLER_ARG" envDefault:"1.0" validate:"min=0,max=1"`
}

// Addr returns the host:port listen address.
func (c Config) Addr() string {
	return net.JoinHostPort(c.Host, strconv.Itoa(c.Port))
}

// Load parses configuration from the process environment and validates it.
func Load() (Config, error) {
	var cfg Config
	if err := env.Parse(&cfg); err != nil {
		return Config{}, fmt.Errorf("parse environment: %w", err)
	}

	validate := validator.New(validator.WithRequiredStructEnabled())
	if err := validate.Struct(cfg); err != nil {
		return Config{}, fmt.Errorf("validate configuration: %w", err)
	}

	return cfg, nil
}
