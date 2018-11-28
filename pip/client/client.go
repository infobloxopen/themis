// Package client implements client for Policy Information Point (PIP) server.
package client

import (
	"errors"
	"math"
	"net"
	"sync/atomic"
	"time"

	"github.com/allegro/bigcache"

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
	Get(path string, args []pdp.AttributeValue) (pdp.AttributeValue, error)
}

// NewClient creates client instance.
func NewClient(opts ...Option) Client {
	o := makeOptions(opts)

	return &client{
		opts: o,

		state:  new(uint32),
		pool:   makeByteBufferPool(o.maxSize),
		d:      makeDialerTK(o.net, o.connAttemptTimeout, o.keepAlive),
		p:      new(provider),
		autoID: new(uint64),
	}
}

type client struct {
	opts options

	state *uint32
	pool  byteBufferPool

	d     dialer
	r     radar
	p     *provider
	cache *bigcache.BigCache

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

	cache, err := newCacheFromOptions(c.opts)
	if err != nil {
		return err
	}

	addrs, r, err := c.newAddressesAndRadar()
	if err != nil {
		return err
	}

	c.r = r
	c.cache = cache
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

	if c.cache != nil {
		c.cache.Reset()
		c.cache = nil
	}
}

func (c *client) Get(path string, args []pdp.AttributeValue) (pdp.AttributeValue, error) {
	for atomic.LoadUint32(c.state) == pipClientConnected {
		v, ok, err := c.tryGet(path, args)
		if !ok || err == nil {
			return v, err
		}
	}

	return pdp.UndefinedValue, ErrNotConnected
}

func (c *client) tryGet(path string, args []pdp.AttributeValue) (pdp.AttributeValue, bool, error) {
	conn := c.p.get()
	if conn == nil {
		return pdp.UndefinedValue, false, ErrNotConnected
	}
	defer conn.g.Done()

	b := c.pool.Get()
	defer func() {
		if b != nil {
			c.pool.Put(b)
		}
	}()

	n, err := pdp.MarshalInfoRequest(b.b, path, args)
	if err != nil {
		return pdp.UndefinedValue, false, err
	}
	b.b = b.b[:n]

	var (
		key   string
		cache *bigcache.BigCache
	)

	if c.cache != nil {
		cache = c.cache
		key = string(b.b)

		var out []byte
		if out, err = cache.Get(key); err == nil {
			var v pdp.AttributeValue
			v, err = pdp.UnmarshalInfoResponse(out)
			if err != nil {
				return pdp.UndefinedValue, false, err
			}

			if c.opts.onCache != nil {
				c.opts.onCache(path, args, v, err)
			}

			return v, false, nil
		}
	}

	b, err = conn.get(b)
	if err != nil {
		c.p.report(conn)
		return pdp.UndefinedValue, true, err
	}

	v, err := pdp.UnmarshalInfoResponse(b.b)
	if err != nil {
		if err, ok := err.(*pdp.ResponseServerError); ok {
			return pdp.UndefinedValue, false, err
		}

		c.p.report(conn)
		return pdp.UndefinedValue, true, err
	}

	if cache != nil {
		cache.Set(key, b.b)
	}

	return v, false, nil
}

func (c *client) nextID() uint64 {
	if atomic.LoadUint64(c.autoID) < math.MaxUint64 {
		return atomic.AddUint64(c.autoID, 1)
	}

	return 0
}

func (c *client) dial(addr string, ch <-chan struct{}) net.Conn {
	if c.opts.connTimeout > 0 {
		return c.dialIterTimeout(addr, ch)
	}

	return c.dialIter(addr, ch)
}

func (c *client) dialIter(addr string, ch <-chan struct{}) net.Conn {
	for atomic.LoadUint32(c.state) == pipClientConnected {
		select {
		default:
		case _, ok := <-ch:
			if !ok {
				return nil
			}
		}

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

func (c *client) dialIterTimeout(addr string, ch <-chan struct{}) net.Conn {
	start := time.Now()
	for atomic.LoadUint32(c.state) == pipClientConnected {
		select {
		default:
		case _, ok := <-ch:
			if !ok {
				return nil
			}
		}

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
