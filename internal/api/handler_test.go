package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"task-api-go/internal/store"
)

func do(t *testing.T, h http.Handler, method, path, body string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func TestHealth(t *testing.T) {
	rec := do(t, NewRouter(store.New()), "GET", "/health", "")
	if rec.Code != 200 {
		t.Fatalf("got %d", rec.Code)
	}
}

func TestCreateThenGet(t *testing.T) {
	h := NewRouter(store.New())
	rec := do(t, h, "POST", "/tasks", `{"title":"learn go"}`)
	if rec.Code != http.StatusCreated || rec.Header().Get("Location") != "/tasks/1" {
		t.Fatalf("create: %d %v", rec.Code, rec.Header())
	}
	rec = do(t, h, "GET", "/tasks/1", "")
	var got store.Task
	_ = json.Unmarshal(rec.Body.Bytes(), &got)
	if rec.Code != 200 || got.Title != "learn go" || got.Status != "todo" {
		t.Fatalf("get: %d %+v", rec.Code, got)
	}
}

func TestStatusCodes(t *testing.T) {
	h := NewRouter(store.New())
	do(t, h, "POST", "/tasks", `{"title":"a"}`)
	cases := []struct {
		name, method, path, body string
		want                     int
	}{
		{"empty title", "POST", "/tasks", `{"title":""}`, 400},
		{"bad json", "POST", "/tasks", `{oops`, 400},
		{"unknown field", "POST", "/tasks", `{"title":"x","extra":1}`, 400},
		{"bad status filter", "GET", "/tasks?status=nope", "", 400},
		{"missing task", "GET", "/tasks/42", "", 404},
		{"non numeric id", "GET", "/tasks/abc", "", 400},
		{"update ok", "PUT", "/tasks/1", `{"title":"b","status":"done"}`, 200},
		{"update missing", "PUT", "/tasks/42", `{"title":"b","status":"done"}`, 404},
		{"delete ok", "DELETE", "/tasks/1", "", 204},
		{"delete again", "DELETE", "/tasks/1", "", 404},
		{"wrong method", "PATCH", "/tasks/1", "", 405},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if rec := do(t, h, c.method, c.path, c.body); rec.Code != c.want {
				t.Fatalf("%s %s: got %d, want %d (%s)", c.method, c.path, rec.Code, c.want, rec.Body.String())
			}
		})
	}
}

func TestListFilter(t *testing.T) {
	h := NewRouter(store.New())
	do(t, h, "POST", "/tasks", `{"title":"a","status":"done"}`)
	do(t, h, "POST", "/tasks", `{"title":"b"}`)
	var got []store.Task
	_ = json.Unmarshal(do(t, h, "GET", "/tasks?status=done", "").Body.Bytes(), &got)
	if len(got) != 1 || got[0].Title != "a" {
		t.Fatalf("unexpected: %+v", got)
	}
}
