package audit

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestMemoryStore_AddAndSnapshot(t *testing.T) {
	store := NewMemoryStore(3)
	for i := 0; i < 5; i++ {
		store.Add(Event{Type: EventReceived, Timestamp: time.Now().UTC()})
	}
	snap := store.Snapshot()
	if len(snap) != 3 {
		t.Fatalf("expected 3 events (ring eviction), got %d", len(snap))
	}
}

func TestMemoryStore_DefaultCapacity(t *testing.T) {
	store := NewMemoryStore(0)
	if store.max != 100 {
		t.Errorf("expected default max 100, got %d", store.max)
	}
}

func TestHandler_Get_ReturnsJSON(t *testing.T) {
	store := NewMemoryStore(10)
	store.Add(Event{Type: EventForwarded, Sink: "s1", Message: "hello", Timestamp: time.Now().UTC()})
	store.Add(Event{Type: EventFiltered, Message: "filtered", Timestamp: time.Now().UTC()})

	req := httptest.NewRequest(http.MethodGet, "/audit", nil)
	rec := httptest.NewRecorder()
	store.Handler()(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("expected application/json, got %q", ct)
	}

	var events []Event
	if err := json.Unmarshal(rec.Body.Bytes(), &events); err != nil {
		t.Fatalf("response is not valid JSON array: %v\nbody: %s", err, rec.Body.String())
	}
	if len(events) != 2 {
		t.Errorf("expected 2 events, got %d", len(events))
	}
}

func TestHandler_MethodNotAllowed(t *testing.T) {
	store := NewMemoryStore(10)
	req := httptest.NewRequest(http.MethodPost, "/audit", nil)
	rec := httptest.NewRecorder()
	store.Handler()(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", rec.Code)
	}
}

func TestHandler_EmptyStore(t *testing.T) {
	store := NewMemoryStore(10)
	req := httptest.NewRequest(http.MethodGet, "/audit", nil)
	rec := httptest.NewRecorder()
	store.Handler()(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var events []Event
	_ = json.Unmarshal(rec.Body.Bytes(), &events)
	if len(events) != 0 {
		t.Errorf("expected empty array, got %d events", len(events))
	}
}
