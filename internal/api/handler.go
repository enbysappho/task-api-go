// Package api exposes the task store as a JSON REST API.
package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"task-api-go/internal/store"
)

type Handler struct{ s *store.Store }

// NewRouter wires all routes (Go 1.22 method + path patterns).
func NewRouter(s *store.Store) http.Handler {
	h := &Handler{s: s}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})
	mux.HandleFunc("GET /tasks", h.list)
	mux.HandleFunc("POST /tasks", h.create)
	mux.HandleFunc("GET /tasks/{id}", h.get)
	mux.HandleFunc("PUT /tasks/{id}", h.update)
	mux.HandleFunc("DELETE /tasks/{id}", h.delete)
	return mux
}

type taskInput struct {
	Title  string `json:"title"`
	Status string `json:"status"`
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}

func writeErr(w http.ResponseWriter, err error) {
	code := http.StatusBadRequest
	if errors.Is(err, store.ErrNotFound) {
		code = http.StatusNotFound
	}
	writeJSON(w, code, map[string]string{"error": err.Error()})
}

func parseID(r *http.Request) (int, error) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || id < 1 {
		return 0, errors.New("id must be a positive integer")
	}
	return id, nil
}

func decode(r *http.Request) (taskInput, error) {
	var in taskInput
	dec := json.NewDecoder(http.MaxBytesReader(nil, r.Body, 1<<20))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&in); err != nil {
		return in, errors.New("invalid JSON body")
	}
	return in, nil
}

func (h *Handler) list(w http.ResponseWriter, r *http.Request) {
	status := r.URL.Query().Get("status")
	if status != "" && status != store.StatusTodo && status != store.StatusDoing && status != store.StatusDone {
		writeErr(w, store.ErrBadStatus)
		return
	}
	writeJSON(w, http.StatusOK, h.s.List(status))
}

func (h *Handler) create(w http.ResponseWriter, r *http.Request) {
	in, err := decode(r)
	if err != nil {
		writeErr(w, err)
		return
	}
	t, err := h.s.Create(in.Title, in.Status)
	if err != nil {
		writeErr(w, err)
		return
	}
	w.Header().Set("Location", "/tasks/"+strconv.Itoa(t.ID))
	writeJSON(w, http.StatusCreated, t)
}

func (h *Handler) get(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		writeErr(w, err)
		return
	}
	t, err := h.s.Get(id)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, t)
}

func (h *Handler) update(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		writeErr(w, err)
		return
	}
	in, err := decode(r)
	if err != nil {
		writeErr(w, err)
		return
	}
	t, err := h.s.Update(id, in.Title, in.Status)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, t)
}

func (h *Handler) delete(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		writeErr(w, err)
		return
	}
	if err := h.s.Delete(id); err != nil {
		writeErr(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
