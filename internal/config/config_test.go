package config_test

import (
	"testing"
	"time"

	"github.com/aircloud/template-backend-go/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoad_Defaults(t *testing.T) {
	// No environment variables set: defaults must produce a valid config.
	cfg, err := config.Load()
	require.NoError(t, err)

	assert.Equal(t, "0.0.0.0", cfg.Host)
	assert.Equal(t, 8080, cfg.Port)
	assert.Equal(t, "template-backend-go", cfg.ServiceName)
	assert.Equal(t, "development", cfg.Environment)
	assert.Equal(t, "info", cfg.LogLevel)
	assert.Equal(t, 5*time.Second, cfg.ReadHeaderTimeout)
	assert.Equal(t, "0.0.0.0:8080", cfg.Addr())
}

func TestLoad_OverridesAndValidation(t *testing.T) {
	tests := map[string]struct {
		env       map[string]string
		wantErr   bool
		assertCfg func(t *testing.T, c config.Config)
	}{
		"valid overrides": {
			env: map[string]string{
				"HOST":        "127.0.0.1",
				"PORT":        "9090",
				"LOG_LEVEL":   "debug",
				"ENVIRONMENT": "production",
			},
			assertCfg: func(t *testing.T, c config.Config) {
				t.Helper()
				assert.Equal(t, "127.0.0.1:9090", c.Addr())
				assert.Equal(t, "debug", c.LogLevel)
			},
		},
		"invalid port too high": {
			env:     map[string]string{"PORT": "70000"},
			wantErr: true,
		},
		"invalid log level": {
			env:     map[string]string{"LOG_LEVEL": "trace"},
			wantErr: true,
		},
		"invalid environment": {
			env:     map[string]string{"ENVIRONMENT": "prod"},
			wantErr: true,
		},
		"sample ratio out of range": {
			env:     map[string]string{"OTEL_TRACES_SAMPLER_ARG": "1.5"},
			wantErr: true,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			for k, v := range tc.env {
				t.Setenv(k, v)
			}

			cfg, err := config.Load()
			if tc.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			if tc.assertCfg != nil {
				tc.assertCfg(t, cfg)
			}
		})
	}
}
