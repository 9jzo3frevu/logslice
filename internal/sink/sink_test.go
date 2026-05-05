package sink

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNew_Valid(t *testing.T) {
	s, err := New("test", "http://localhost:9999")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s.Name != "test" {
		t.Errorf("expected name 'test', got %q", s.Name)
	}
}

func TestNew_MissingName(t *testing.T) {
	_, err := New("", "http://localhost:9999")
	if err == nil {
		t.Fatal("expected error for empty name")
	}
}

func TestNew_MissingURL(t *testing.T) {
	_, err := New("test", "")
	if err == nil {
		t.Fatal("expected error for empty url")
	}
}

func TestSend_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	s, _ := New("test", server.URL)
	if err := s.Send([]byte(`{"level":"info","msg":"hello"}`)); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestSend_NonSuccessStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	s, _ := New("test", server.URL)
	if err := s.Send([]byte(`{}`)); err == nil {
		t.Error("expected error for non-2xx status")
	}
}

func TestSend_Unreachable(t *testing.T) {
	s, _ := New("test", "http://127.0.0.1:19999")
	if err := s.Send([]byte(`{}`)); err == nil {
		t.Error("expected error for unreachable server")
	}
}
