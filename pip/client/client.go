// Package client implements client for Policy Information Point (PIP) server.
package client

import "errors"

// ErrorConnected occurs if method connect is called after connection has been established.
var ErrorConnected = errors.New("connection has been already established")

// Client defines abstract PIP service client interface.
type Client interface {
	// Connect establishes connection to given PIP server.
	Connect() error

	// Close terminates previously established connection if any. Close should
	// silently return if connection hasn't been established yet or if it has
	// been already closed.
	Close()
}

// NewClient creates client instance.
func NewClient(opts ...Option) Client {
	o := defaults
	for _, opt := range opts {
		opt(&o)
	}

	return &client{
		opts:  o,
		state: new(uint32),
	}
}

type client struct {
	opts options

	state *uint32
}
