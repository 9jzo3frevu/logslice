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

func TestServer_StartShutdown(t *testing.T) {
	sinkServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer sinkServer.Close()

	f, err := filter.New("info")
	if err != nil {
		t.Fatalf("filter.New: %v", err)
	}
	s, err := sink.New("test", sinkServer.URL)
	if err != nil {
		t.Fatalf("sink.New: %v", err)
	}
	fo := sink.NewFanout([]*sink.Sink{s})
	h := proxy.NewHandler(f, fo)

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

	f, _ := filter.New("info")
	s, _ := sink.New("t", sinkServer.URL)
	fo := sink.NewFanout([]*sink.Sink{s})
	h := proxy.NewHandler(f, fo)

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
