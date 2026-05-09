package replay

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func newTestHandler(cap int) *Handler {
	s := New(cap)
	return NewHandler(s, map[string]Sender{})
}

func TestHandler_MethodNotAllowed(t *testing.T) {
	h := newTestHandler(10)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodDelete, "/replay", nil))
	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", rec.Code)
	}
}

func TestHandler_Get_EmptyStore(t *testing.T) {
	h := newTestHandler(10)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/replay", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var out []entryView
	if err := json.NewDecoder(rec.Body).Decode(&out); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if len(out) != 0 {
		t.Fatalf("expected empty list, got %d entries", len(out))
	}
}

func TestHandler_Get_ListsEntries(t *testing.T) {
	s := New(10)
	s.Add("s1", []byte(`{"msg":"x"}`), 1)
	h := NewHandler(s, map[string]Sender{})
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/replay", nil))
	var out []entryView
	_ = json.NewDecoder(rec.Body).Decode(&out)
	if len(out) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(out))
	}
	if out[0].Sink != "s1" {
		t.Fatalf("unexpected sink %q", out[0].Sink)
	}
}

func TestHandler_Post_Replays(t *testing.T) {
	s := New(10)
	s.Add("good", []byte(`{}`), 1)
	h := NewHandler(s, map[string]Sender{"good": &mockSender{}})
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/replay", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var result map[string]int
	_ = json.NewDecoder(rec.Body).Decode(&result)
	if result["remaining"] != 0 {
		t.Fatalf("expected 0 remaining, got %d", result["remaining"])
	}
}

func TestHandler_ContentType(t *testing.T) {
	h := newTestHandler(10)
	for _, method := range []string{http.MethodGet, http.MethodPost} {
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, httptest.NewRequest(method, "/replay", nil))
		ct := rec.Header().Get("Content-Type")
		if ct != "application/json" {
			t.Fatalf("%s: expected application/json, got %q", method, ct)
		}
	}
}

// Ensure mockSender satisfies Sender (compile-time check in test package).
var _ Sender = (*mockSender)(nil)

type neverSender struct{}

func (n *neverSender) Send(_ context.Context, _ []byte) error { return nil }
