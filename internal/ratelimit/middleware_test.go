package ratelimit_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/logslice/logslice/internal/ratelimit"
)

func okHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
}

func TestMiddleware_AllowsWithinLimit(t *testing.T) {
	l := ratelimit.New(5)
	h := ratelimit.Middleware(l, okHandler())

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestMiddleware_BlocksWhenExceeded(t *testing.T) {
	l := ratelimit.New(1)
	h := ratelimit.Middleware(l, okHandler())

	req := httptest.NewRequest(http.MethodPost, "/", nil)

	// Exhaust the single token.
	rec1 := httptest.NewRecorder()
	h.ServeHTTP(rec1, req)
	if rec1.Code != http.StatusOK {
		t.Fatalf("expected first request 200, got %d", rec1.Code)
	}

	// Second request should be rate-limited.
	rec2 := httptest.NewRecorder()
	h.ServeHTTP(rec2, req)
	if rec2.Code != http.StatusTooManyRequests {
		t.Fatalf("expected 429, got %d", rec2.Code)
	}
}

func TestMiddleware_ResponseBody(t *testing.T) {
	l := ratelimit.New(0) // defaults to 1, exhaust immediately
	h := ratelimit.Middleware(l, okHandler())
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	h.ServeHTTP(httptest.NewRecorder(), req) // consume token

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusTooManyRequests {
		t.Fatalf("expected 429, got %d", rec.Code)
	}
	if body := rec.Body.String(); body == "" {
		t.Fatal("expected non-empty error body")
	}
}
