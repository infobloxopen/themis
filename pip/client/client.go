// Package client implements client for Policy Information Point (PIP) server.
package client

import (
	"errors"
	"math"
	"net"
	"sync/atomic"
	"time"

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

		state:  new(uint32),
		pool:   makeBytePool(o.maxSize),
		d:      makeDialerTK(o.net, o.connAttemptTimeout, o.keepAlive),
		p:      new(provider),
		autoID: new(uint64),
	}
}

type client struct {
	opts options

	state *uint32
	pool  bytePool

	d dialer
	p *provider

	autoID *uint64
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

	addrs := []string{c.opts.addr}
	var err error
	if c.opts.balancer != balancerTypeSimple {
		if len(c.opts.addrs) > 0 {
			addrs = c.opts.addrs
		} else {
			addrs, err = lookupHostPort(c.opts.addr)
			if err != nil {
				return err
			}
		}
	}

	c.p.start(c, addrs)

	state = pipClientConnected

	return nil
}

func (c *client) Close() {
	if !atomic.CompareAndSwapUint32(c.state, pipClientConnected, pipClientClosing) {
		return
	}
	defer atomic.StoreUint32(c.state, pipClientIdle)

	c.p.stop()
}

func (c *client) Get(args ...pdp.AttributeAssignment) (pdp.AttributeValue, error) {
	conn := c.p.get()
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

func (c *client) nextID() uint64 {
	if atomic.LoadUint64(c.autoID) < math.MaxUint64 {
		return atomic.AddUint64(c.autoID, 1)
	}

	return 0
}

func (c *client) dial(addr string) net.Conn {
	if c.opts.connTimeout > 0 {
		return c.dialIterTimeout(addr)
	}

	return c.dialIter(addr)
}

func (c *client) dialIter(addr string) net.Conn {
	for atomic.LoadUint32(c.state) == pipClientConnected {
		n, err := c.d.dial(addr)
		if err == nil {
			return n
		}

		if isConnRefused(err) {
			time.Sleep(c.opts.connAttemptTimeout)
			continue
		}

		if isConnTimeout(err) {
			continue
		}

		if c.opts.onErr != nil {
			c.opts.onErr(nil, err)
		}

		return nil
	}

	return nil
}

func (c *client) dialIterTimeout(addr string) net.Conn {
	start := time.Now()
	for atomic.LoadUint32(c.state) == pipClientConnected {
		n, err := c.d.dial(addr)
		if err == nil {
			return n
		}

		if time.Since(start) < c.opts.connTimeout {
			if isConnRefused(err) {
				time.Sleep(c.opts.connAttemptTimeout)
				continue
			}

			if isConnTimeout(err) {
				continue
			}
		}

		if c.opts.onErr != nil {
			c.opts.onErr(nil, err)
		}

		return nil
	}

	return nil
}
