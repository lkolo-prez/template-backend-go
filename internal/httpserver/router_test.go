package httpserver_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/aircloud/template-backend-go/internal/httpserver"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestServer(t *testing.T) http.Handler {
	t.Helper()
	return httpserver.New(httpserver.Options{Version: "test", ServiceName: "test"})
}

func TestRouter_HealthEndpoints(t *testing.T) {
	srv := newTestServer(t)

	for _, path := range []string{"/healthz", "/readyz"} {
		req := httptest.NewRequest(http.MethodGet, path, nil)
		rec := httptest.NewRecorder()
		srv.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code, "path %s", path)
		assert.Contains(t, rec.Body.String(), `"status"`)
	}
}

func TestRouter_SecurityHeaders(t *testing.T) {
	srv := newTestServer(t)
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, req)

	assert.Equal(t, "nosniff", rec.Header().Get("X-Content-Type-Options"))
	assert.Equal(t, "DENY", rec.Header().Get("X-Frame-Options"))
	assert.NotEmpty(t, rec.Header().Get("Content-Security-Policy"))
}

func TestRouter_RequestIDGenerated(t *testing.T) {
	srv := newTestServer(t)
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, req)

	require.NotEmpty(t, rec.Header().Get("X-Request-Id"))
}

func TestRouter_RequestIDPropagated(t *testing.T) {
	srv := newTestServer(t)
	const id = "11111111-1111-1111-1111-111111111111"
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	req.Header.Set("X-Request-Id", id)
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, req)

	assert.Equal(t, id, rec.Header().Get("X-Request-Id"))
}

func TestRouter_NotFound(t *testing.T) {
	srv := newTestServer(t)
	req := httptest.NewRequest(http.MethodGet, "/does-not-exist", nil)
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)
}
