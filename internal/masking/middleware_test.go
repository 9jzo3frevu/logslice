package masking_test

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/yourorg/logslice/internal/masking"
)

func captureHandler(t *testing.T) (http.Handler, func() map[string]any) {
	t.Helper()
	var captured map[string]any
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(body, &captured)
		w.WriteHeader(http.StatusOK)
	})
	return h, func() map[string]any { return captured }
}

func TestMiddleware_MasksSensitiveField(t *testing.T) {
	m, _ := masking.New([]masking.Config{
		{Field: "token", Pattern: `[a-zA-Z0-9]{8,}`, Replacement: "[REDACTED]"},
	})
	handler, getEntry := captureHandler(t)
	mw := masking.Middleware(m)(handler)

	body := `{"message":"login","token":"supersecrettoken123","level":"info"}`
	req := httptest.NewRequest(http.MethodPost, "/logs", strings.NewReader(body))
	rec := httptest.NewRecorder()
	mw.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	entry := getEntry()
	if entry["token"] == "supersecrettoken123" {
		t.Error("token should have been masked")
	}
	if entry["message"] != "login" {
		t.Error("unrelated field should be preserved")
	}
}

func TestMiddleware_PassesThroughNonPost(t *testing.T) {
	m, _ := masking.New([]masking.Config{
		{Field: "token", Pattern: `.+`, Replacement: "***"},
	})
	handler, _ := captureHandler(t)
	mw := masking.Middleware(m)(handler)

	req := httptest.NewRequest(http.MethodGet, "/logs", nil)
	rec := httptest.NewRecorder()
	mw.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestMiddleware_PassesThroughNonJSON(t *testing.T) {
	m, _ := masking.New([]masking.Config{
		{Field: "token", Pattern: `.+`, Replacement: "***"},
	})
	handler, _ := captureHandler(t)
	mw := masking.Middleware(m)(handler)

	req := httptest.NewRequest(http.MethodPost, "/logs", strings.NewReader("not-json"))
	rec := httptest.NewRecorder()
	mw.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestMiddleware_PreservesUnrelatedFields(t *testing.T) {
	m, _ := masking.New([]masking.Config{
		{Field: "secret", Pattern: `.+`, Replacement: "***"},
	})
	handler, getEntry := captureHandler(t)
	mw := masking.Middleware(m)(handler)

	body := `{"message":"ok","level":"debug","secret":"abc"}`
	req := httptest.NewRequest(http.MethodPost, "/logs", strings.NewReader(body))
	rec := httptest.NewRecorder()
	mw.ServeHTTP(rec, req)

	entry := getEntry()
	if entry["level"] != "debug" {
		t.Errorf("expected level=debug, got %v", entry["level"])
	}
	if entry["secret"] != "***" {
		t.Errorf("expected secret=***, got %v", entry["secret"])
	}
}
