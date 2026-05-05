package sink

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// Sink represents a log destination.
type Sink struct {
	Name    string
	URL     string
	client  *http.Client
}

// New creates a new Sink with the given name and target URL.
func New(name, url string) (*Sink, error) {
	if name == "" {
		return nil, fmt.Errorf("sink name must not be empty")
	}
	if url == "" {
		return nil, fmt.Errorf("sink url must not be empty")
	}
	return &Sink{
		Name: name,
		URL:  url,
		client: &http.Client{
			Timeout: 5 * time.Second,
		},
	}, nil
}

// Send forwards a log payload (JSON bytes) to the sink's URL via HTTP POST.
func (s *Sink) Send(payload []byte) error {
	resp, err := s.client.Post(s.URL, "application/json", strings.NewReader(string(payload)))
	if err != nil {
		return fmt.Errorf("sink %q: request failed: %w", s.Name, err)
	}
	defer resp.Body.Close()
	io.Copy(io.Discard, resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("sink %q: unexpected status %d", s.Name, resp.StatusCode)
	}
	return nil
}
