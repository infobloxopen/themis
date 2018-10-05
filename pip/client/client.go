// Package client implements client for Policy Information Point (PIP) server.
package client

import (
	"errors"
	"net"
	"sync"

	"github.com/infobloxopen/themis/pdp"
)

var (
	// ErrConnected occurs if method connect is called after connection has been
	// established.
	ErrConnected = errors.New("connection has been already established")

	// ErrNotConnected occurs if method get is called before connection has been
	// established.
	ErrNotConnected = errors.New("connection hasn't been established yet")
)

// Client defines abstract PIP service client interface.
type Client interface {
	// Connect establishes connection to given PIP server.
	Connect() error

	// Close terminates previously established connection if any. Close silently
	// returns if connection hasn't been established yet or if it has been
	// already closed.
	Close()

	// Get requests information from PIP.
	Get(args ...pdp.AttributeAssignment) (pdp.AttributeValue, error)
}

// NewClient creates client instance.
func NewClient(opts ...Option) Client {
	o := defaults
	for _, opt := range opts {
		opt(&o)
	}

	if o.maxSize+msgSizeBytes+msgIdxBytes > o.bufSize {
		o.bufSize = defBufSize
	}

	return &client{
		opts: o,

		state: new(uint32),
		pool:  makeBytePool(o.maxSize, false),

		lock: new(sync.RWMutex),
	}
}

type client struct {
	opts options

	state *uint32
	pool  bytePool

	c net.Conn

	lock *sync.RWMutex
	gwg  *sync.WaitGroup
	req  chan request
	wwg  *sync.WaitGroup

	pipes pipes
	dt    chan struct{}
}
