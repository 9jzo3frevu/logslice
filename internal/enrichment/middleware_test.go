package enrichment

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func captureHandler(t *testing.T) (http.Handler, func() map[string]any) {
	t.Helper()
	var captured map[string]any
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		json.Unmarshal(body, &captured)
		w.WriteHeader(http.StatusOK)
	})
	return h, func() map[string]any { return captured }
}

func TestMiddleware_EnrichesEntry(t *testing.T) {
	e, _ := New(map[string]string{"env": "prod", "app": "logslice"})
	h, getCapture := captureHandler(t)
	mw := Middleware(e)(h)

	body := `{"message":"hello","level":"info"}`
	req := httptest.NewRequest(http.MethodPost, "/logs", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	mw.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	cap := getCapture()
	if cap["env"] != "prod" {
		t.Errorf("expected env=prod, got %v", cap["env"])
	}
	if cap["app"] != "logslice" {
		t.Errorf("expected app=logslice, got %v", cap["app"])
	}
	if cap["message"] != "hello" {
		t.Error("original message field should be preserved")
	}
}

func TestMiddleware_PassesThroughNonPost(t *testing.T) {
	e, _ := New(map[string]string{"env": "prod"})
	h, _ := captureHandler(t)
	mw := Middleware(e)(h)

	req := httptest.NewRequest(http.MethodGet, "/logs", nil)
	rr := httptest.NewRecorder()
	mw.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}

func TestMiddleware_PassesThroughNonJSON(t *testing.T) {
	e, _ := New(map[string]string{"env": "prod"})
	h, _ := captureHandler(t)
	mw := Middleware(e)(h)

	req := httptest.NewRequest(http.MethodPost, "/logs", bytes.NewBufferString("plain text"))
	req.Header.Set("Content-Type", "text/plain")
	rr := httptest.NewRecorder()
	mw.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}

func TestMiddleware_InvalidJSONPassesThrough(t *testing.T) {
	e, _ := New(map[string]string{"env": "prod"})
	h, _ := captureHandler(t)
	mw := Middleware(e)(h)

	req := httptest.NewRequest(http.MethodPost, "/logs", bytes.NewBufferString("not-json"))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	mw.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected passthrough with 200, got %d", rr.Code)
	}
}
