package timeout_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/logslice/logslice/internal/timeout"
)

func slowHandler(delay time.Duration) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		select {
		case <-time.After(delay):
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("ok"))
		case <-r.Context().Done():
			// context cancelled — do nothing so the middleware can respond
		}
	})
}

func TestNew_DefaultTimeout(t *testing.T) {
	l := timeout.New(0)
	if l.Duration() != 5*time.Second {
		t.Fatalf("expected 5s default, got %v", l.Duration())
	}
}

func TestNew_CustomTimeout(t *testing.T) {
	l := timeout.New(10 * time.Second)
	if l.Duration() != 10*time.Second {
		t.Fatalf("expected 10s, got %v", l.Duration())
	}
}

func TestMiddleware_CompletesInTime(t *testing.T) {
	l := timeout.New(200 * time.Millisecond)
	handler := l.Middleware(slowHandler(10 * time.Millisecond))

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/ingest", nil)
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestMiddleware_ExceedsTimeout(t *testing.T) {
	l := timeout.New(20 * time.Millisecond)
	handler := l.Middleware(slowHandler(300 * time.Millisecond))

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/ingest", nil)
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d", rec.Code)
	}
	if body := rec.Body.String(); body != "request timeout\n" {
		t.Fatalf("unexpected body: %q", body)
	}
}

func TestMiddleware_ContentTypeOnTimeout(t *testing.T) {
	l := timeout.New(20 * time.Millisecond)
	handler := l.Middleware(slowHandler(300 * time.Millisecond))

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/ingest", nil)
	handler.ServeHTTP(rec, req)

	ct := rec.Header().Get("Content-Type")
	if ct != "text/plain; charset=utf-8" {
		t.Fatalf("expected plain text content-type, got %q", ct)
	}
}
