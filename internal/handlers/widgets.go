package handlers

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"sort"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

// maxBodyBytes caps request bodies to mitigate memory-exhaustion attacks.
const maxBodyBytes = 1 << 20 // 1 MiB

// Widget is the API resource. In a real service this would be persisted; here
// it lives in an in-memory store to keep the template dependency-free.
type Widget struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	Color       string    `json:"color"`
	Quantity    int       `json:"quantity"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// WidgetInput is the request payload for create/update operations.
type WidgetInput struct {
	Name        string `json:"name" validate:"required,min=1,max=120"`
	Description string `json:"description" validate:"max=1024"`
	Color       string `json:"color" validate:"required,hexcolor"`
	Quantity    int    `json:"quantity" validate:"gte=0,lte=1000000"`
}

// Widgets implements an in-memory CRUD API for Widget resources.
type Widgets struct {
	logger   *slog.Logger
	validate *validator.Validate

	mu    sync.RWMutex
	store map[string]Widget
}

// NewWidgets constructs a Widgets handler with an empty store.
func NewWidgets(logger *slog.Logger) *Widgets {
	if logger == nil {
		logger = slog.Default()
	}
	return &Widgets{
		logger:   logger,
		validate: validator.New(validator.WithRequiredStructEnabled()),
		store:    make(map[string]Widget),
	}
}

// List handles GET /api/v1/widgets and returns all widgets, newest first.
func (h *Widgets) List(w http.ResponseWriter, _ *http.Request) {
	h.mu.RLock()
	items := make([]Widget, 0, len(h.store))
	for _, item := range h.store {
		items = append(items, item)
	}
	h.mu.RUnlock()

	sort.Slice(items, func(i, j int) bool {
		return items[i].CreatedAt.After(items[j].CreatedAt)
	})

	writeJSON(w, http.StatusOK, map[string]any{
		"items": items,
		"count": len(items),
	})
}

// Get handles GET /api/v1/widgets/{id}.
func (h *Widgets) Get(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	h.mu.RLock()
	widget, ok := h.store[id]
	h.mu.RUnlock()

	if !ok {
		writeError(w, http.StatusNotFound, "widget not found")
		return
	}
	writeJSON(w, http.StatusOK, widget)
}

// Create handles POST /api/v1/widgets.
func (h *Widgets) Create(w http.ResponseWriter, r *http.Request) {
	input, ok := h.decodeAndValidate(w, r)
	if !ok {
		return
	}

	now := time.Now().UTC()
	widget := Widget{
		ID:          uuid.NewString(),
		Name:        input.Name,
		Description: input.Description,
		Color:       input.Color,
		Quantity:    input.Quantity,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	h.mu.Lock()
	h.store[widget.ID] = widget
	h.mu.Unlock()

	h.logger.InfoContext(r.Context(), "widget created", slog.String("widget_id", widget.ID))
	w.Header().Set("Location", "/api/v1/widgets/"+widget.ID)
	writeJSON(w, http.StatusCreated, widget)
}

// Update handles PUT /api/v1/widgets/{id} (full replacement).
func (h *Widgets) Update(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	input, ok := h.decodeAndValidate(w, r)
	if !ok {
		return
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	existing, found := h.store[id]
	if !found {
		writeError(w, http.StatusNotFound, "widget not found")
		return
	}

	existing.Name = input.Name
	existing.Description = input.Description
	existing.Color = input.Color
	existing.Quantity = input.Quantity
	existing.UpdatedAt = time.Now().UTC()
	h.store[id] = existing

	h.logger.InfoContext(r.Context(), "widget updated", slog.String("widget_id", id))
	writeJSON(w, http.StatusOK, existing)
}

// Delete handles DELETE /api/v1/widgets/{id}.
func (h *Widgets) Delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	h.mu.Lock()
	_, found := h.store[id]
	if found {
		delete(h.store, id)
	}
	h.mu.Unlock()

	if !found {
		writeError(w, http.StatusNotFound, "widget not found")
		return
	}

	h.logger.InfoContext(r.Context(), "widget deleted", slog.String("widget_id", id))
	w.WriteHeader(http.StatusNoContent)
}

// decodeAndValidate parses and validates a WidgetInput, writing an error
// response and returning ok=false if anything is wrong.
func (h *Widgets) decodeAndValidate(w http.ResponseWriter, r *http.Request) (WidgetInput, bool) {
	var input WidgetInput

	r.Body = http.MaxBytesReader(w, r.Body, maxBodyBytes)
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	if err := dec.Decode(&input); err != nil {
		var maxErr *http.MaxBytesError
		switch {
		case errors.As(err, &maxErr):
			writeError(w, http.StatusRequestEntityTooLarge, "request body too large")
		case errors.Is(err, io.EOF):
			writeError(w, http.StatusBadRequest, "request body is empty")
		default:
			writeError(w, http.StatusBadRequest, "malformed JSON request body")
		}
		return WidgetInput{}, false
	}

	if err := h.validate.Struct(input); err != nil {
		details := make(map[string]string)
		var verrs validator.ValidationErrors
		if errors.As(err, &verrs) {
			for _, fe := range verrs {
				details[fe.Field()] = fe.Tag()
			}
		}
		writeValidationError(w, details)
		return WidgetInput{}, false
	}

	return input, true
}
