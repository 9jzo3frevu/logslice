package proxy_test

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/yourorg/logslice/internal/filter"
	"github.com/yourorg/logslice/internal/proxy"
	"github.com/yourorg/logslice/internal/sink"
)

func newTestHandler(t *testing.T, minLevel string, sinkURL string) *proxy.Handler {
	t.Helper()
	f, err := filter.New(minLevel)
	if err != nil {
		t.Fatalf("filter.New: %v", err)
	}
	s, err := sink.New("test", sinkURL)
	if err != nil {
		t.Fatalf("sink.New: %v", err)
	}
	fo := sink.NewFanout([]*sink.Sink{s})
	return proxy.NewHandler(f, fo)
}

func TestHandler_NonPost(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	h := newTestHandler(t, "info", server.URL)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/logs", nil)
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", rec.Code)
	}
}

func TestHandler_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	h := newTestHandler(t, "info", server.URL)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/logs", bytes.NewBufferString("not-json"))
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

func TestHandler_FilteredOut(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	h := newTestHandler(t, "error", server.URL)
	rec := httptest.NewRecorder()
	body := bytes.NewBufferString(`{"level":"debug","msg":"verbose"}`)
	req := httptest.NewRequest(http.MethodPost, "/logs", body)
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d", rec.Code)
	}
}

func TestHandler_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	h := newTestHandler(t, "info", server.URL)
	rec := httptest.NewRecorder()
	body := bytes.NewBufferString(`{"level":"info","msg":"hello"}`)
	req := httptest.NewRequest(http.MethodPost, "/logs", body)
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusAccepted {
		t.Errorf("expected 202, got %d", rec.Code)
	}
}
