// Package handlers contains the HTTP request handlers for the service:
// liveness/readiness probes and the v1 widgets CRUD API.
package handlers

import (
	"encoding/json"
	"net/http"
)

// ErrorResponse is the canonical error envelope returned by the API.
type ErrorResponse struct {
	Error   string            `json:"error"`
	Details map[string]string `json:"details,omitempty"`
}

// writeJSON serializes v as JSON with the given status code.
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if v == nil {
		return
	}
	// Encoding to an already-written header cannot change the status; on error
	// the connection is simply closed by the server.
	_ = json.NewEncoder(w).Encode(v)
}

// writeError writes a structured error envelope.
func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, ErrorResponse{Error: msg})
}

// writeValidationError writes a 422 with field-level details.
func writeValidationError(w http.ResponseWriter, details map[string]string) {
	writeJSON(w, http.StatusUnprocessableEntity, ErrorResponse{
		Error:   "validation failed",
		Details: details,
	})
}

// NotFound is the router-level 404 handler.
func NotFound(w http.ResponseWriter, _ *http.Request) {
	writeError(w, http.StatusNotFound, "resource not found")
}

// MethodNotAllowed is the router-level 405 handler.
func MethodNotAllowed(w http.ResponseWriter, _ *http.Request) {
	writeError(w, http.StatusMethodNotAllowed, "method not allowed")
}
