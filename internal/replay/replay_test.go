package replay

import (
	"context"
	"errors"
	"testing"
)

func TestNew_DefaultCapacity(t *testing.T) {
	s := New(0)
	if s.cap != 100 {
		t.Fatalf("expected cap 100, got %d", s.cap)
	}
}

func TestAdd_And_Len(t *testing.T) {
	s := New(10)
	s.Add("s1", []byte(`{"msg":"a"}`), 1)
	s.Add("s2", []byte(`{"msg":"b"}`), 2)
	if s.Len() != 2 {
		t.Fatalf("expected 2, got %d", s.Len())
	}
}

func TestAdd_Evicts_WhenFull(t *testing.T) {
	s := New(3)
	for i := 0; i < 4; i++ {
		s.Add("sink", []byte(`{}`), 1)
	}
	if s.Len() != 3 {
		t.Fatalf("expected 3, got %d", s.Len())
	}
}

func TestDrain_ClearsStore(t *testing.T) {
	s := New(10)
	s.Add("s1", []byte(`{}`), 1)
	out := s.Drain()
	if len(out) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(out))
	}
	if s.Len() != 0 {
		t.Fatal("store should be empty after drain")
	}
}

type mockSender struct{ err error }

func (m *mockSender) Send(_ context.Context, _ []byte) error { return m.err }

func TestReplay_SuccessfulSend(t *testing.T) {
	s := New(10)
	s.Add("ok", []byte(`{}`), 1)
	senders := map[string]Sender{"ok": &mockSender{}}
	s.Replay(context.Background(), senders)
	if s.Len() != 0 {
		t.Fatal("successful replay should leave store empty")
	}
}

func TestReplay_FailedSend_ReQueues(t *testing.T) {
	s := New(10)
	s.Add("bad", []byte(`{}`), 1)
	senders := map[string]Sender{"bad": &mockSender{err: errors.New("timeout")}}
	s.Replay(context.Background(), senders)
	if s.Len() != 1 {
		t.Fatal("failed replay should re-queue entry")
	}
	entries := s.Drain()
	if entries[0].Attempts != 2 {
		t.Fatalf("expected attempts=2, got %d", entries[0].Attempts)
	}
}

func TestReplay_UnknownSink_Dropped(t *testing.T) {
	s := New(10)
	s.Add("ghost", []byte(`{}`), 1)
	s.Replay(context.Background(), map[string]Sender{})
	if s.Len() != 0 {
		t.Fatal("entries for unknown sinks should be dropped")
	}
}
