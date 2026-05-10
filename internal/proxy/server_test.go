package proxy_test

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/yourorg/logslice/internal/filter"
	"github.com/yourorg/logslice/internal/proxy"
	"github.com/yourorg/logslice/internal/sink"
)

// newTestHandler is a helper that creates a proxy.Handler backed by a test sink
// pointed at the given upstream URL and a filter for the specified log level.
func newTestHandler(t *testing.T, level, upstreamURL string) *proxy.Handler {
	t.Helper()
	f, err := filter.New(level)
	if err != nil {
		t.Fatalf("filter.New(%q): %v", level, err)
	}
	s, err := sink.New("test", upstreamURL)
	if err != nil {
		t.Fatalf("sink.New: %v", err)
	}
	return proxy.NewHandler(f, sink.NewFanout([]*sink.Sink{s}))
}

func TestServer_StartShutdown(t *testing.T) {
	sinkServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer sinkServer.Close()

	h := newTestHandler(t, "info", sinkServer.URL)
	srv := proxy.NewServer("127.0.0.1:0", h)

	errCh := make(chan error, 1)
	go func() { errCh <- srv.Start() }()

	time.Sleep(50 * time.Millisecond)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		t.Fatalf("Shutdown: %v", err)
	}

	select {
	case err := <-errCh:
		if err != nil {
			t.Fatalf("Start returned error: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("server did not stop in time")
	}
}

func TestServer_RouteNotFound(t *testing.T) {
	sinkServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer sinkServer.Close()

	h := newTestHandler(t, "info", sinkServer.URL)

	srv := proxy.NewServer("127.0.0.1:19876", h)
	go srv.Start() //nolint:errcheck
	time.Sleep(50 * time.Millisecond)
	defer srv.Shutdown(context.Background()) //nolint:errcheck

	resp, err := http.Post("http://127.0.0.1:19876/logs",
		"application/json",
		bytes.NewBufferString(`{"level":"info","msg":"ok"}`))
	if err != nil {
		t.Fatalf("POST /logs: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusAccepted {
		t.Errorf("expected 202, got %d", resp.StatusCode)
	}
}
