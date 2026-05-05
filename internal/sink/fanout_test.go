package sink

import (
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
)

func TestFanout_SendAll(t *testing.T) {
	var hits int32

	makeServer := func() *httptest.Server {
		return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			atomic.AddInt32(&hits, 1)
			w.WriteHeader(http.StatusOK)
		}))
	}

	s1 := makeServer()
	s2 := makeServer()
	s3 := makeServer()
	defer s1.Close()
	defer s2.Close()
	defer s3.Close()

	sink1, _ := New("s1", s1.URL)
	sink2, _ := New("s2", s2.URL)
	sink3, _ := New("s3", s3.URL)

	fo := NewFanout(sink1, sink2, sink3)
	if fo.Len() != 3 {
		t.Fatalf("expected 3 sinks, got %d", fo.Len())
	}

	if err := fo.Send([]byte(`{"level":"info"}`)); err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if atomic.LoadInt32(&hits) != 3 {
		t.Errorf("expected 3 hits, got %d", hits)
	}
}

func TestFanout_PartialFailure(t *testing.T) {
	good := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer good.Close()

	sinkGood, _ := New("good", good.URL)
	sinkBad, _ := New("bad", "http://127.0.0.1:19998")

	fo := NewFanout(sinkGood, sinkBad)
	if err := fo.Send([]byte(`{}`)); err == nil {
		t.Error("expected error when one sink fails")
	}
}

func TestFanout_Empty(t *testing.T) {
	fo := NewFanout()
	if err := fo.Send([]byte(`{}`)); err != nil {
		t.Errorf("unexpected error with no sinks: %v", err)
	}
}
