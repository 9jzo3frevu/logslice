package proxy

import (
	"context"
	"errors"
	"net/http"
	"time"
)

// Server wraps an HTTP server that exposes the log ingestion endpoint.
type Server struct {
	httpServer *http.Server
}

// NewServer creates a Server listening on addr, routing POST /logs to handler.
func NewServer(addr string, handler *Handler) *Server {
	mux := http.NewServeMux()
	mux.Handle("/logs", handler)

	return &Server{
		httpServer: &http.Server{
			Addr:         addr,
			Handler:      mux,
			ReadTimeout:  10 * time.Second,
			WriteTimeout: 10 * time.Second,
			IdleTimeout:  30 * time.Second,
		},
	}
}

// Start begins listening and serving. It blocks until the server stops.
func (s *Server) Start() error {
	if err := s.httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}

// Shutdown gracefully stops the server within the provided context deadline.
func (s *Server) Shutdown(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}
