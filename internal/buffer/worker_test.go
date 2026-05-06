package buffer

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"
)

type mockSender struct {
	mu      sync.Mutex
	recv    [][]byte
	failAll bool
}

func (m *mockSender) Send(p []byte) error {
	if m.failAll {
		return errors.New("mock send failure")
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	cp := make([]byte, len(p))
	copy(cp, p)
	m.recv = append(m.recv, cp)
	return nil
}

func (m *mockSender) Received() [][]byte {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.recv
}

func TestWorker_ForwardsEntries(t *testing.T) {
	b := New(8)
	sender := &mockSender{}
	w := NewWorker(b, sender)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go w.Run(ctx)

	payloads := []string{"alpha", "beta", "gamma"}
	for _, p := range payloads {
		_ = b.Write(Entry{Payload: []byte(p)})
	}

	time.Sleep(80 * time.Millisecond)
	got := sender.Received()
	if len(got) != len(payloads) {
		t.Fatalf("expected %d forwarded, got %d", len(payloads), len(got))
	}
}

func TestWorker_ContinuesOnSendError(t *testing.T) {
	b := New(4)
	sender := &mockSender{failAll: true}
	w := NewWorker(b, sender)

	ctx, cancel := context.WithTimeout(context.Background(), 150*time.Millisecond)
	defer cancel()

	_ = b.Write(Entry{Payload: []byte("will-fail")})
	_ = b.Write(Entry{Payload: []byte("also-fail")})

	// Should not panic or block — just log errors.
	w.Run(ctx)
}
