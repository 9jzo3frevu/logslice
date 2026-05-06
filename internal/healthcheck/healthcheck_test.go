package healthcheck_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/logslice/logslice/internal/healthcheck"
)

func newChecker(ready bool) *healthcheck.Checker {
	c := healthcheck.New("v0.1.0")
	c.SetReady(ready)
	return c
}

func TestLiveness_AlwaysOK(t *testing.T) {
	c := newChecker(false) // not ready, but liveness should still pass
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/healthz/live", nil)
	c.LivenessHandler(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	var s healthcheck.Status
	if err := json.NewDecoder(rr.Body).Decode(&s); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if !s.OK {
		t.Error("expected ok=true")
	}
	if s.Version != "v0.1.0" {
		t.Errorf("expected version v0.1.0, got %s", s.Version)
	}
}

func TestReadiness_NotReady(t *testing.T) {
	c := newChecker(false)
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/healthz/ready", nil)
	c.ReadinessHandler(rr, req)
	if rr.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d", rr.Code)
	}
	var s healthcheck.Status
	if err := json.NewDecoder(rr.Body).Decode(&s); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if s.OK {
		t.Error("expected ok=false")
	}
}

func TestReadiness_Ready(t *testing.T) {
	c := newChecker(true)
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/healthz/ready", nil)
	c.ReadinessHandler(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}

func TestLiveness_MethodNotAllowed(t *testing.T) {
	c := newChecker(true)
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/healthz/live", nil)
	c.LivenessHandler(rr, req)
	if rr.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", rr.Code)
	}
}

func TestReadiness_MethodNotAllowed(t *testing.T) {
	c := newChecker(true)
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/healthz/ready", nil)
	c.ReadinessHandler(rr, req)
	if rr.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", rr.Code)
	}
}

func TestContentType(t *testing.T) {
	c := newChecker(true)
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/healthz/live", nil)
	c.LivenessHandler(rr, req)
	ct := rr.Header().Get("Content-Type")
	if ct != "application/json" {
		t.Errorf("expected application/json, got %s", ct)
	}
}
