package ratelimit

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func baseHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
}

func TestKeyedMiddleware_AllowsFirstRequest(t *testing.T) {
	h := KeyedMiddleware(10, 1, time.Minute, baseHandler())
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.RemoteAddr = "192.0.2.1:1234"
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}

func TestKeyedMiddleware_BlocksAfterBurst(t *testing.T) {
	h := KeyedMiddleware(0.001, 1, time.Minute, baseHandler())

	req1 := httptest.NewRequest(http.MethodPost, "/", nil)
	req1.RemoteAddr = "10.0.0.1:9000"
	rr1 := httptest.NewRecorder()
	h.ServeHTTP(rr1, req1)

	req2 := httptest.NewRequest(http.MethodPost, "/", nil)
	req2.RemoteAddr = "10.0.0.1:9001"
	rr2 := httptest.NewRecorder()
	h.ServeHTTP(rr2, req2)

	if rr2.Code != http.StatusTooManyRequests {
		t.Fatalf("expected 429, got %d", rr2.Code)
	}
}

func TestKeyedMiddleware_DifferentIPsIndependent(t *testing.T) {
	h := KeyedMiddleware(0.001, 1, time.Minute, baseHandler())

	for _, ip := range []string{"1.1.1.1:0", "2.2.2.2:0", "3.3.3.3:0"} {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.RemoteAddr = ip
		rr := httptest.NewRecorder()
		h.ServeHTTP(rr, req)
		if rr.Code != http.StatusOK {
			t.Fatalf("ip %s: expected 200, got %d", ip, rr.Code)
		}
	}
}

func TestKeyedMiddleware_XForwardedFor(t *testing.T) {
	h := KeyedMiddleware(0.001, 1, time.Minute, baseHandler())

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Forwarded-For", "203.0.113.5, 10.0.0.1")
	req.RemoteAddr = "10.0.0.1:80"
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 for XFF request, got %d", rr.Code)
	}
}
