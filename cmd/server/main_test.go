package main

import (
	"bytes"
	"net"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/aircloud/template-backend-go/internal/config"
)

func TestRunVersion(t *testing.T) {
	t.Parallel()

	var out, errOut bytes.Buffer
	code := run([]string{"-version"}, &out, &errOut)

	assert.Equal(t, 0, code)
	assert.Contains(t, out.String(), "version=dev")
}

func TestRunUnknownFlag(t *testing.T) {
	t.Parallel()

	var out, errOut bytes.Buffer
	assert.Equal(t, 2, run([]string{"-nope"}, &out, &errOut))
}

func TestProbe(t *testing.T) {
	t.Parallel()

	ok := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/healthz" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(ok.Close)

	host, portStr, err := net.SplitHostPort(ok.Listener.Addr().String())
	require.NoError(t, err)
	port, err := strconv.Atoi(portStr)
	require.NoError(t, err)

	require.NoError(t, probe(config.Config{Host: host, Port: port}))

	bad := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	t.Cleanup(bad.Close)

	_, badPortStr, err := net.SplitHostPort(bad.Listener.Addr().String())
	require.NoError(t, err)
	badPort, err := strconv.Atoi(badPortStr)
	require.NoError(t, err)

	assert.Error(t, probe(config.Config{Host: "0.0.0.0", Port: badPort}))
}

func TestNewLogger(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	newLogger("not-a-level", &buf).Info("hello")
	assert.Contains(t, buf.String(), `"msg":"hello"`)
}
