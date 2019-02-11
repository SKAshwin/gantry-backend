package http

import (
	"net"
	"net/http"
)

//PreFlightHandler is an interface which provides an adapter to
//handle CORS pre-flight requests
type PreFlightHandler interface {
	Handle(h http.Handler) http.Handler
}

// Server represents an HTTP server.
type Server struct {
	ln net.Listener

	// Handler to serve.
	Handler *Handler

	// Bind address to open.
	Addr string

	//Handles CORS requests
	PreFlightHandler PreFlightHandler
}

// Open opens a socket and serves the HTTP server.
func (s *Server) Open() error {
	// Open socket.
	ln, err := net.Listen("tcp", s.Addr)
	if err != nil {
		return err
	}
	s.ln = ln

	// Start HTTP server.
	go func() { http.Serve(s.ln, s.PreFlightHandler.Handle(s.Handler)) }()

	return nil
}

// Close closes the socket.
func (s *Server) Close() error {
	if s.ln != nil {
		s.ln.Close()
	}
	return nil
}
