package handlers

import (
	"net/http"
	"sync/atomic"
	"time"
)

// Health serves Kubernetes-style liveness and readiness probes.
//
// Liveness reports that the process is running; readiness additionally gates on
// the service having completed startup and not being in a draining state.
type Health struct {
	version   string
	startedAt time.Time
	ready     atomic.Bool
}

// HealthResponse is the JSON body returned by both probes.
type HealthResponse struct {
	Status    string `json:"status"`
	Version   string `json:"version"`
	UptimeSec int64  `json:"uptime_seconds"`
}

// NewHealth constructs a Health handler that is immediately marked ready.
// Call SetReady(false) before shutdown to fail readiness while draining.
func NewHealth(version string) *Health {
	h := &Health{version: version, startedAt: time.Now()}
	h.ready.Store(true)
	return h
}

// SetReady toggles the readiness state (e.g. set false during graceful drain).
func (h *Health) SetReady(ready bool) { h.ready.Store(ready) }

// Live handles GET /healthz and always returns 200 while the process runs.
func (h *Health) Live(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, HealthResponse{
		Status:    "ok",
		Version:   h.version,
		UptimeSec: int64(time.Since(h.startedAt).Seconds()),
	})
}

// Ready handles GET /readyz and returns 503 until the service is ready.
func (h *Health) Ready(w http.ResponseWriter, _ *http.Request) {
	if !h.ready.Load() {
		writeJSON(w, http.StatusServiceUnavailable, HealthResponse{
			Status:    "draining",
			Version:   h.version,
			UptimeSec: int64(time.Since(h.startedAt).Seconds()),
		})
		return
	}
	writeJSON(w, http.StatusOK, HealthResponse{
		Status:    "ready",
		Version:   h.version,
		UptimeSec: int64(time.Since(h.startedAt).Seconds()),
	})
}
