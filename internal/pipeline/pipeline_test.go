package pipeline_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/logslice/logslice/internal/config"
	"github.com/logslice/logslice/internal/pipeline"
)

func sinkServer(t *testing.T, received *[]map[string]any) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var entry map[string]any
		if err := json.NewDecoder(r.Body).Decode(&entry); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		*received = append(*received, entry)
		w.WriteHeader(http.StatusOK)
	}))
}

func TestNew_ValidPipeline(t *testing.T) {
	received := []map[string]any{}
	server := sinkServer(t, &received)
	defer server.Close()

	cfg := &config.Config{
		Sinks: []config.SinkConfig{
			{Name: "test-sink", URL: server.URL},
		},
		Filter: config.FilterConfig{MinLevel: "info"},
	}

	p, err := pipeline.New(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p == nil {
		t.Fatal("expected non-nil pipeline")
	}
}

func TestNew_InvalidSink(t *testing.T) {
	cfg := &config.Config{
		Sinks: []config.SinkConfig{
			{Name: "", URL: "http://localhost"},
		},
		Filter: config.FilterConfig{MinLevel: "info"},
	}

	_, err := pipeline.New(cfg)
	if err == nil {
		t.Fatal("expected error for missing sink name")
	}
}

func TestPipeline_Process(t *testing.T) {
	received := []map[string]any{}
	server := sinkServer(t, &received)
	defer server.Close()

	cfg := &config.Config{
		Sinks: []config.SinkConfig{
			{Name: "test-sink", URL: server.URL},
		},
		Filter: config.FilterConfig{MinLevel: "info"},
	}

	p, err := pipeline.New(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	body := `{"message":"hello","level":"info"}`
	req := httptest.NewRequest(http.MethodPost, "/logs", strings.NewReader(body))
	rec := httptest.NewRecorder()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	_ = ctx

	p.ServeHTTP(rec, req)

	if rec.Code != http.StatusAccepted {
		t.Errorf("expected 202, got %d", rec.Code)
	}
}
