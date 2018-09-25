// Package server provides server for Policy Information Point.
package server

import (
	"errors"
	"net"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
)

var (
	// ErrBound indicates that server has been already bound to a port or file.
	ErrBound = errors.New("server has been already bound to a port or file")
	// ErrNotBound indicates that server hasn't been bound yet.
	ErrNotBound = errors.New("server hasn't been bound yet")
	// ErrStarted indicates that server has been already started.
	ErrStarted = errors.New("server has been already started")
	// ErrNotStarted indicates that server hasn't been started yet.
	ErrNotStarted = errors.New("server hasn't been started yet")
)

const (
	srvIdle uint32 = iota
	srvBinding
	srvBound
	srvStarting
	srvStarted
	srvStopping
)

// Server structure represents PIP server.
type Server struct {
	opts options

	state *uint32
	ln    net.Listener

	conns *connReg
}

// NewServer creates new Server instance.
func NewServer(opts ...Option) *Server {
	o := defaults
	for _, opt := range opts {
		opt(&o)
	}

	return &Server{
		opts:  o,
		state: new(uint32),
		conns: newConnReg(o.maxConn),
	}
}

// Bind links server to a port or file.
func (s *Server) Bind() error {
	if !atomic.CompareAndSwapUint32(s.state, srvIdle, srvBinding) {
		return ErrBound
	}
	state := srvIdle
	defer func() {
		atomic.StoreUint32(s.state, state)
	}()

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
	state = srvBound
	return nil
}

// Serve starts accepting incoming connections.
func (s *Server) Serve() error {
	if !atomic.CompareAndSwapUint32(s.state, srvBound, srvStarting) {
		if s := atomic.LoadUint32(s.state); s == srvIdle || s == srvBinding {
			return ErrNotBound
		}

		return ErrStarted
	}

	ln := s.ln

	atomic.StoreUint32(s.state, srvStarted)

	wg := new(sync.WaitGroup)

	for {
		c, err := ln.Accept()
		if err != nil {
			if isLisenerClosed(err) {
				break
			}

			return err
		}

		cc := connWithErrHandler{
			c: c,
			h: s.opts.onErr,
		}
		idx := s.conns.put(cc)
		if idx >= 0 {
			wg.Add(1)
			go s.handle(wg, cc, idx)
		} else {
			cc.handle(c.Close())
		}
	}

	wg.Wait()

	return nil
}

// Stop terminates server.
func (s *Server) Stop() error {
	if atomic.CompareAndSwapUint32(s.state, srvStarted, srvStopping) {
		defer atomic.StoreUint32(s.state, srvIdle)

		s.conns.delAll()
	} else if atomic.CompareAndSwapUint32(s.state, srvBound, srvStopping) {
		defer atomic.StoreUint32(s.state, srvIdle)
	} else {
		return ErrNotStarted
	}

	ln := s.ln
	s.ln = nil

	return ln.Close()
}

const netListenerClosedMsg = "use of closed network connection"

func isLisenerClosed(err error) bool {
	switch err := err.(type) {
	case *net.OpError:
		return err.Err.Error() == netListenerClosedMsg
	}

	return false
}
