// Package testutil provides test helpers for the terrakube client library.
// It MUST NOT import the parent terrakube package to avoid import cycles.
package testutil

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// Server wraps httptest.Server with convenience methods for testing.
type Server struct {
	*httptest.Server
	mux *http.ServeMux
	t   testing.TB
}

// NewServer creates a test server with a fresh ServeMux.
func NewServer(t testing.TB) *Server {
	t.Helper()
	mux := http.NewServeMux()
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return &Server{
		Server: srv,
		mux:    mux,
		t:      t,
	}
}

// HandleFunc registers a handler on the server's mux.
// Pattern follows Go 1.22+ syntax: "METHOD /path".
func (s *Server) HandleFunc(pattern string, handler http.HandlerFunc) {
	s.t.Helper()
	s.mux.HandleFunc(pattern, handler)
}

// Mux returns the underlying ServeMux for direct access.
func (s *Server) Mux() *http.ServeMux {
	return s.mux
}
