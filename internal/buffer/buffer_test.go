package buffer

import (
	"context"
	"testing"
	"time"
)

func TestNew_DefaultsToOne(t *testing.T) {
	b := New(0)
	if b.Cap() != 1 {
		t.Fatalf("expected cap 1, got %d", b.Cap())
	}
}

func TestWrite_WithinCapacity(t *testing.T) {
	b := New(3)
	for i := 0; i < 3; i++ {
		if err := b.Write(Entry{Payload: []byte("x")}); err != nil {
			t.Fatalf("unexpected error on write %d: %v", i, err)
		}
	}
	if b.Len() != 3 {
		t.Fatalf("expected len 3, got %d", b.Len())
	}
}

func TestWrite_Full(t *testing.T) {
	b := New(2)
	_ = b.Write(Entry{Payload: []byte("a")})
	_ = b.Write(Entry{Payload: []byte("b")})
	err := b.Write(Entry{Payload: []byte("c")})
	if err != ErrFull {
		t.Fatalf("expected ErrFull, got %v", err)
	}
}

func TestDrain_ReceivesEntries(t *testing.T) {
	b := New(4)
	payloads := []string{"one", "two", "three"}
	for _, p := range payloads {
		_ = b.Write(Entry{Payload: []byte(p)})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	var got []string
	go func() {
		b.Drain(ctx, func(e Entry) {
			got = append(got, string(e.Payload))
		})
	}()

	time.Sleep(50 * time.Millisecond)
	if len(got) != len(payloads) {
		t.Fatalf("expected %d entries, got %d", len(payloads), len(got))
	}
}

func TestDrain_StopsOnContextCancel(t *testing.T) {
	b := New(10)
	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan struct{})
	go func() {
		b.Drain(ctx, func(Entry) {})
		close(done)
	}()

	cancel()
	select {
	case <-done:
		// ok
	case <-time.After(500 * time.Millisecond):
		t.Fatal("Drain did not stop after context cancel")
	}
}

func TestLen_And_Cap(t *testing.T) {
	b := New(5)
	_ = b.Write(Entry{Payload: []byte("hi")})
	if b.Len() != 1 {
		t.Fatalf("expected len 1, got %d", b.Len())
	}
	if b.Cap() != 5 {
		t.Fatalf("expected cap 5, got %d", b.Cap())
	}
}
