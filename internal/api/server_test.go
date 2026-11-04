package api

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/ali/convex-hull-advanced-algorithms/internal/algorithms"
)

func TestRunCompareAll(t *testing.T) {
	s := NewServer(algorithms.All())
	req := httptest.NewRequest(http.MethodPost, "/api/run", strings.NewReader(`{"algorithm":"all","points":[{"x":1,"y":2}]}`))
	w := httptest.NewRecorder()
	s.ServeHTTP(w, req)
	if w.Code != 200 {
		t.Fatalf("status %d: %s", w.Code, w.Body.String())
	}
	for _, a := range algorithms.All() {
		if !strings.Contains(w.Body.String(), `"algorithm":"`+a.Name()+`"`) {
			t.Fatalf("missing %s in %s", a.Name(), w.Body.String())
		}
	}
	if strings.Contains(w.Body.String(), "not_implemented") {
		t.Fatalf("unexpected stub result: %s", w.Body.String())
	}
}
func TestRunRejectsInvalidPoints(t *testing.T) {
	s := NewServer(algorithms.All())
	req := httptest.NewRequest(http.MethodPost, "/api/run", strings.NewReader(`{"algorithm":"all","points":[{"x":1e999,"y":2}]}`))
	w := httptest.NewRecorder()
	s.ServeHTTP(w, req)
	if w.Code != 400 {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}
func TestDatasetEndpoint(t *testing.T) {
	s := NewServer(algorithms.All())
	req := httptest.NewRequest(http.MethodPost, "/api/dataset", strings.NewReader(`{"kind":"uniform","count":5,"seed":7}`))
	w := httptest.NewRecorder()
	s.ServeHTTP(w, req)
	if w.Code != 200 || !strings.Contains(w.Body.String(), `"count":5`) {
		t.Fatalf("status %d: %s", w.Code, w.Body.String())
	}
}
