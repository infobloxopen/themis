// Package client implements client for Policy Information Point (PIP) server.
package client

import (
	"errors"
	"sync/atomic"

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
		pool:  makeBytePool(o.maxSize),
		b:     new(simpleBalancer),
	}
}

type client struct {
	opts options

	state *uint32
	pool  bytePool

	b balancer
}

const (
	pipClientIdle uint32 = iota
	pipClientConnecting
	pipClientConnected
	pipClientClosing
)

func (c *client) Connect() error {
	if !atomic.CompareAndSwapUint32(c.state, pipClientIdle, pipClientConnecting) {
		return ErrConnected
	}

	state := pipClientIdle
	defer func() {
		atomic.StoreUint32(c.state, state)
	}()

	if err := c.b.start(c); err != nil {
		return err
	}

	state = pipClientConnected

	return nil
}

func (c *client) Close() {
	if !atomic.CompareAndSwapUint32(c.state, pipClientConnected, pipClientClosing) {
		return
	}
	defer atomic.StoreUint32(c.state, pipClientIdle)

	c.b.stop()
}

func (c *client) Get(args ...pdp.AttributeAssignment) (pdp.AttributeValue, error) {
	conn := c.b.get()
	if conn == nil {
		return pdp.UndefinedValue, ErrNotConnected
	}
	defer conn.g.Done()

	b := c.pool.Get()
	defer func() {
		if b != nil {
			c.pool.Put(b)
		}
	}()

	n, err := pdp.MarshalRequestAssignmentsToBuffer(b, args)
	if err != nil {
		return pdp.UndefinedValue, err
	}

	b, err = conn.get(b[:n])
	return pdp.UndefinedValue, err
}
