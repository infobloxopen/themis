// Package server provides server for Policy Information Point.
package server

import (
	"errors"
	"net"
	"os"
	"strings"
	"syscall"
)

var (
	// ErrBound indicates that server has been already bound to a port or file.
	ErrBound = errors.New("server has been already bound to a port or file")
	// ErrNotStarted indicates that server hasn't been started yet.
	ErrNotStarted = errors.New("server hasn't been started yet")
)

// Server structure represents PIP server.
type Server struct {
	opts options

	ln net.Listener
}

// NewServer creates new Server instance.
func NewServer(opts ...Option) *Server {
	o := defaults
	for _, opt := range opts {
		opt(&o)
	}

	return &Server{
		opts: o,
	}
}

// Bind links server to a port or file.
func (s *Server) Bind() error {
	if s.ln != nil {
		return ErrBound
	}

	nw := strings.ToLower(s.opts.net)
	if nw == "unix" {
		if err := os.Remove(s.opts.addr); err != nil {
			if pErr, ok := err.(*os.PathError); !ok || pErr.Err != syscall.ENOENT {
				return err
			}
		}
	}

	ln, err := net.Listen(nw, s.opts.addr)
	if err != nil {
		return err
	}

	s.ln = ln
	return nil
}

// Serve starts accepting incoming connections.
func (s *Server) Serve() error {
	for {
		c, err := s.ln.Accept()
		if err != nil {
			if opErr, ok := err.(*net.OpError); ok && opErr.Err.Error() == "use of closed network connection" {
				break
			}

			return err
		}

		c.Close()
	}

	return nil
}

// Stop terminates server.
func (s *Server) Stop() error {
	if s.ln == nil {
		return ErrNotStarted
	}

	if err := s.ln.Close(); err != nil {
		return err
	}
	s.ln = nil

	return nil
}
