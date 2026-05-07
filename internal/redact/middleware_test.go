package redact_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/yourorg/logslice/internal/redact"
)

func captureHandler(t *testing.T) (http.Handler, *string) {
	t.Helper()
	var body string
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		body = string(b)
		w.WriteHeader(http.StatusOK)
	})
	return h, &body
}

func TestMiddleware_RedactsSensitiveFields(t *testing.T) {
	r, _ := redact.New([]string{"password"})
	next, body := captureHandler(t)
	mw := redact.Middleware(r)(next)

	payload := `{"message":"login","password":"secret"}`
	req := httptest.NewRequest(http.MethodPost, "/logs", strings.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	mw.ServeHTTP(w, req)

	if strings.Contains(*body, "secret") {
		t.Errorf("expected password to be redacted, body: %s", *body)
	}
	if !strings.Contains(*body, "[REDACTED]") {
		t.Errorf("expected redaction mask in body, got: %s", *body)
	}
}

func TestMiddleware_PassesThroughNonPost(t *testing.T) {
	r, _ := redact.New([]string{"password"})
	next, body := captureHandler(t)
	mw := redact.Middleware(r)(next)

	req := httptest.NewRequest(http.MethodGet, "/logs", nil)
	w := httptest.NewRecorder()
	mw.ServeHTTP(w, req)

	_ = body // not checked — GET has no body
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestMiddleware_PassesThroughNonJSON(t *testing.T) {
	r, _ := redact.New([]string{"password"})
	next, body := captureHandler(t)
	mw := redact.Middleware(r)(next)

	req := httptest.NewRequest(http.MethodPost, "/logs", strings.NewReader("plain text"))
	w := httptest.NewRecorder()
	mw.ServeHTTP(w, req)

	if *body != "plain text" {
		t.Errorf("expected body unchanged, got: %s", *body)
	}
}

func TestMiddleware_PreservesUnrelatedFields(t *testing.T) {
	r, _ := redact.New([]string{"token"})
	next, body := captureHandler(t)
	mw := redact.Middleware(r)(next)

	payload := `{"level":"info","message":"ok","token":"abc"}`
	req := httptest.NewRequest(http.MethodPost, "/logs", strings.NewReader(payload))
	w := httptest.NewRecorder()
	mw.ServeHTTP(w, req)

	if !strings.Contains(*body, "\"level\"") {
		t.Errorf("expected level field preserved, body: %s", *body)
	}
	if !strings.Contains(*body, "\"message\"") {
		t.Errorf("expected message field preserved, body: %s", *body)
	}
}
