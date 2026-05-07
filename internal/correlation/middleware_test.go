package correlation_test

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/logslice/logslice/internal/correlation"
)

type fixedGen struct{ id string }

func (f *fixedGen) Generate() string { return f.id }

func captureHandler(t *testing.T) (http.Handler, func() map[string]interface{}) {
	t.Helper()
	var captured map[string]interface{}
	h := http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(body, &captured)
	})
	return h, func() map[string]interface{} { return captured }
}

func TestMiddleware_InjectsGeneratedID(t *testing.T) {
	inj := correlation.New(&fixedGen{id: "abc123"})
	next, snap := captureHandler(t)
	h := correlation.Middleware(inj, next)

	body := `{"message":"hello","level":"info"}`
	req := httptest.NewRequest(http.MethodPost, "/logs", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	entry := snap()
	if entry == nil {
		t.Fatal("handler did not receive a body")
	}
	if got := entry[correlation.FieldName]; got != "abc123" {
		t.Errorf("expected correlation_id=abc123, got %v", got)
	}
	if got := rec.Header().Get(correlation.HeaderName); got != "abc123" {
		t.Errorf("expected response header %s=abc123, got %s", correlation.HeaderName, got)
	}
}

func TestMiddleware_PropagatesExistingHeader(t *testing.T) {
	inj := correlation.New(&fixedGen{id: "generated"})
	next, snap := captureHandler(t)
	h := correlation.Middleware(inj, next)

	body := `{"message":"hi","level":"debug"}`
	req := httptest.NewRequest(http.MethodPost, "/logs", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set(correlation.HeaderName, "existing-id-999")
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	entry := snap()
	if entry[correlation.FieldName] != "existing-id-999" {
		t.Errorf("expected propagated ID, got %v", entry[correlation.FieldName])
	}
}

func TestMiddleware_PassesThroughNonPost(t *testing.T) {
	inj := correlation.New(nil)
	called := false
	next := http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) { called = true })
	h := correlation.Middleware(inj, next)

	req := httptest.NewRequest(http.MethodGet, "/logs", nil)
	h.ServeHTTP(httptest.NewRecorder(), req)

	if !called {
		t.Error("expected next handler to be called for GET request")
	}
}

func TestMiddleware_PassesThroughNonJSON(t *testing.T) {
	inj := correlation.New(&fixedGen{id: "x"})
	next, snap := captureHandler(t)
	h := correlation.Middleware(inj, next)

	req := httptest.NewRequest(http.MethodPost, "/logs", bytes.NewBufferString("plain text"))
	req.Header.Set("Content-Type", "text/plain")
	h.ServeHTTP(httptest.NewRecorder(), req)

	if snap() != nil {
		t.Error("expected no JSON parsing for non-JSON content type")
	}
}
