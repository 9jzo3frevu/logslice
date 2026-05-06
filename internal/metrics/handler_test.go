package metrics

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestMetricsHandler_Get(t *testing.T) {
	c := &Counters{}
	c.Received.Add(3)
	c.Forwarded.Add(2)
	c.Filtered.Add(1)
	c.Errors.Add(0)

	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	rec := httptest.NewRecorder()

	Handler(c)(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var snap Snapshot
	if err := json.NewDecoder(rec.Body).Decode(&snap); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if snap.Received != 3 {
		t.Errorf("expected Received=3, got %d", snap.Received)
	}
	if snap.Forwarded != 2 {
		t.Errorf("expected Forwarded=2, got %d", snap.Forwarded)
	}
	if snap.Filtered != 1 {
		t.Errorf("expected Filtered=1, got %d", snap.Filtered)
	}
}

func TestMetricsHandler_MethodNotAllowed(t *testing.T) {
	c := &Counters{}
	req := httptest.NewRequest(http.MethodPost, "/metrics", nil)
	rec := httptest.NewRecorder()

	Handler(c)(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", rec.Code)
	}
}

func TestMetricsHandler_ContentType(t *testing.T) {
	c := &Counters{}
	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	rec := httptest.NewRecorder()

	Handler(c)(rec, req)

	ct := rec.Header().Get("Content-Type")
	if ct != "application/json" {
		t.Errorf("expected Content-Type application/json, got %s", ct)
	}
}
