package handlers_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/aircloud/template-backend-go/internal/handlers"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newWidgetRouter mounts the widgets handler under chi so URL params resolve.
func newWidgetRouter(t *testing.T) http.Handler {
	t.Helper()
	h := handlers.NewWidgets(nil)
	r := chi.NewRouter()
	r.Route("/api/v1/widgets", func(wr chi.Router) {
		wr.Get("/", h.List)
		wr.Post("/", h.Create)
		wr.Get("/{id}", h.Get)
		wr.Put("/{id}", h.Update)
		wr.Delete("/{id}", h.Delete)
	})
	return r
}

func do(t *testing.T, srv http.Handler, method, path, body string) *httptest.ResponseRecorder {
	t.Helper()
	var r *http.Request
	if body == "" {
		r = httptest.NewRequestWithContext(t.Context(), method, path, nil)
	} else {
		r = httptest.NewRequestWithContext(t.Context(), method, path, bytes.NewBufferString(body))
	}
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, r)
	return rec
}

func TestWidgets_CreateAndGet(t *testing.T) {
	srv := newWidgetRouter(t)

	rec := do(t, srv, http.MethodPost, "/api/v1/widgets",
		`{"name":"Bolt","color":"#ff8800","quantity":42}`)
	require.Equal(t, http.StatusCreated, rec.Code)
	assert.NotEmpty(t, rec.Header().Get("Location"))

	var created handlers.Widget
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &created))
	assert.Equal(t, "Bolt", created.Name)
	assert.NotEmpty(t, created.ID)

	rec = do(t, srv, http.MethodGet, "/api/v1/widgets/"+created.ID, "")
	require.Equal(t, http.StatusOK, rec.Code)
}

func TestWidgets_CreateValidation(t *testing.T) {
	srv := newWidgetRouter(t)

	tests := map[string]struct {
		body   string
		status int
	}{
		"missing name":   {`{"color":"#ffffff","quantity":1}`, http.StatusUnprocessableEntity},
		"bad color":      {`{"name":"x","color":"notacolor","quantity":1}`, http.StatusUnprocessableEntity},
		"negative qty":   {`{"name":"x","color":"#fff","quantity":-1}`, http.StatusUnprocessableEntity},
		"malformed json": {`{"name":`, http.StatusBadRequest},
		"unknown field":  {`{"name":"x","color":"#fff","bogus":true}`, http.StatusBadRequest},
		"empty body":     {``, http.StatusBadRequest},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			rec := do(t, srv, http.MethodPost, "/api/v1/widgets", tc.body)
			assert.Equal(t, tc.status, rec.Code)
		})
	}
}

func TestWidgets_ListEmpty(t *testing.T) {
	srv := newWidgetRouter(t)
	rec := do(t, srv, http.MethodGet, "/api/v1/widgets", "")
	require.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), `"count":0`)
}

func TestWidgets_UpdateLifecycle(t *testing.T) {
	srv := newWidgetRouter(t)

	rec := do(t, srv, http.MethodPost, "/api/v1/widgets",
		`{"name":"Bolt","color":"#ffffff","quantity":1}`)
	require.Equal(t, http.StatusCreated, rec.Code)
	var created handlers.Widget
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &created))

	rec = do(t, srv, http.MethodPut, "/api/v1/widgets/"+created.ID,
		`{"name":"Nut","color":"#000000","quantity":7}`)
	require.Equal(t, http.StatusOK, rec.Code)
	var updated handlers.Widget
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &updated))
	assert.Equal(t, "Nut", updated.Name)
	assert.Equal(t, 7, updated.Quantity)
}

func TestWidgets_Delete(t *testing.T) {
	srv := newWidgetRouter(t)

	rec := do(t, srv, http.MethodPost, "/api/v1/widgets",
		`{"name":"Bolt","color":"#ffffff","quantity":1}`)
	var created handlers.Widget
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &created))

	rec = do(t, srv, http.MethodDelete, "/api/v1/widgets/"+created.ID, "")
	require.Equal(t, http.StatusNoContent, rec.Code)

	rec = do(t, srv, http.MethodGet, "/api/v1/widgets/"+created.ID, "")
	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestWidgets_NotFound(t *testing.T) {
	srv := newWidgetRouter(t)

	for _, method := range []string{http.MethodGet, http.MethodDelete} {
		rec := do(t, srv, method, "/api/v1/widgets/missing", "")
		assert.Equal(t, http.StatusNotFound, rec.Code)
	}

	rec := do(t, srv, http.MethodPut, "/api/v1/widgets/missing",
		`{"name":"x","color":"#fff","quantity":1}`)
	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestWidgets_BodyTooLarge(t *testing.T) {
	srv := newWidgetRouter(t)
	huge := `{"name":"` + strings.Repeat("a", 2<<20) + `","color":"#fff"}`
	rec := do(t, srv, http.MethodPost, "/api/v1/widgets", huge)
	assert.Equal(t, http.StatusRequestEntityTooLarge, rec.Code)
}
