package circuitbreaker

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func okBreaker(t *testing.T) *Breaker {
	t.Helper()
	return New(3, time.Minute)
}

func TestMiddleware_AllowsRequest(t *testing.T) {
	b := okBreaker(t)
	h := Middleware(b, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestMiddleware_BlocksWhenOpen(t *testing.T) {
	b := New(1, time.Minute)
	b.RecordFailure() // open the circuit

	h := Middleware(b, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d", rec.Code)
	}
}

func TestMiddleware_RecordsFailureOn5xx(t *testing.T) {
	b := New(3, time.Minute)
	h := Middleware(b, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/", nil))
	if b.failures != 1 {
		t.Fatalf("expected 1 failure recorded, got %d", b.failures)
	}
}

func TestMiddleware_RecordsSuccessOn2xx(t *testing.T) {
	b := New(3, time.Minute)
	b.RecordFailure()
	b.RecordFailure()
	h := Middleware(b, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusAccepted)
	}))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/", nil))
	if b.failures != 0 {
		t.Fatalf("expected failures reset to 0, got %d", b.failures)
	}
}

func TestMiddleware_ContentTypeOnBlock(t *testing.T) {
	b := New(1, time.Minute)
	b.RecordFailure()
	h := Middleware(b, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {}))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Fatalf("expected application/json, got %q", ct)
	}
}
