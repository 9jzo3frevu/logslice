package metrics

import "testing"

func TestCounters_IncrementAndSnap(t *testing.T) {
	c := &Counters{}

	c.Received.Add(10)
	c.Forwarded.Add(7)
	c.Filtered.Add(2)
	c.Errors.Add(1)

	snap := c.Snap()

	if snap.Received != 10 {
		t.Errorf("expected Received=10, got %d", snap.Received)
	}
	if snap.Forwarded != 7 {
		t.Errorf("expected Forwarded=7, got %d", snap.Forwarded)
	}
	if snap.Filtered != 2 {
		t.Errorf("expected Filtered=2, got %d", snap.Filtered)
	}
	if snap.Errors != 1 {
		t.Errorf("expected Errors=1, got %d", snap.Errors)
	}
}

func TestCounters_Reset(t *testing.T) {
	c := &Counters{}
	c.Received.Add(5)
	c.Forwarded.Add(3)
	c.Reset()

	snap := c.Snap()
	if snap.Received != 0 || snap.Forwarded != 0 {
		t.Errorf("expected all zeros after Reset, got %+v", snap)
	}
}

func TestGlobal_NotNil(t *testing.T) {
	if Global == nil {
		t.Fatal("Global metrics instance should not be nil")
	}
}
