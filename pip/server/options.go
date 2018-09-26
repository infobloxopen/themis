package server

import (
	"math"
	"net"
	"time"
)

// Option configures how we set up PIP server.
type Option func(*options)

// WithNetwork returns an Option which sets service network. It supports "tcp", "tcp4", "tcp6" and "unix" netwroks.
func WithNetwork(net string) Option {
	return func(o *options) {
		o.net = net
	}
}

// WithAddress returns an Option which sets service endpoint.
func WithAddress(addr string) Option {
	return func(o *options) {
		o.addr = addr
	}
}

// WithMaxConnections returns an Option which limits number of simultaneous connections. If n <= 0 server doesn't limit incoming connections.
func WithMaxConnections(n int) Option {
	return func(o *options) {
		if n > 0 {
			o.maxConn = n
		} else {
			o.maxConn = 0
		}
	}
}

// ConnErrHandler is a function to process errors within a connection.
type ConnErrHandler func(net.Addr, error)

// WithConnErrHandler returns an Option which sets error handler for communication errors within a connection.
func WithConnErrHandler(f ConnErrHandler) Option {
	return func(o *options) {
		o.onErr = f
	}
}

// WithBufferSize returns an Option which sets size of input and output buffers. By default or if n <= 0 it is 1 MB.
func WithBufferSize(n int) Option {
	return func(o *options) {
		if n > 0 {
			o.bufSize = n
		} else {
			o.bufSize = defBufSize
		}
	}
}

// WithMaxMessageSize returns an Option which sets limit on message size. Default 10 KB.
func WithMaxMessageSize(n int) Option {
	return func(o *options) {
		if n > 0 && n <= math.MaxUint32 {
			o.maxMsgSize = n
		} else {
			o.maxMsgSize = defMaxMsgSize
		}
	}
}

// WithWriteInterval returns an Option which sets duration after which data from write buffer are sent to network even if write buffer isn't full.
func WithWriteInterval(d time.Duration) Option {
	return func(o *options) {
		o.writeInt = d
	}
}

type options struct {
	net        string
	addr       string
	maxConn    int
	onErr      ConnErrHandler
	bufSize    int
	maxMsgSize int
	writeInt   time.Duration
}

const (
	defBufSize    = 1024 * 1024
	defMaxMsgSize = 10 * 1024
	defWriteInt   = 50 * time.Microsecond
)

var defaults = options{
	net:        "tcp",
	addr:       "localhost:5600",
	bufSize:    defBufSize,
	maxMsgSize: defMaxMsgSize,
	writeInt:   defWriteInt,
}
