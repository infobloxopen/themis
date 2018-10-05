package client

import (
	"math"
	"time"
)

// An Option allows to set PIP client options.
type Option func(*options)

// WithNetwork returns an Option which sets destination network. The client
// supports "tcp", "tcp4", "tcp6" and "unix" netwroks.
func WithNetwork(n string) Option {
	return func(o *options) {
		o.net = n
	}
}

// WithAddress returns an Option which sets destination address.
func WithAddress(a string) Option {
	return func(o *options) {
		o.addr = a
	}
}

// WithMaxRequestSize returns an Option which limits request size in bytes
// to given value. Default 10KB.
func WithMaxRequestSize(n int) Option {
	return func(o *options) {
		if n > 0 && n <= math.MaxInt32-msgIdxBytes {
			o.maxSize = n
		} else {
			o.maxSize = defMaxSize
		}
	}
}

// WithMaxQueue returns an Option which limits number of requests client can
// send in parallel.
func WithMaxQueue(n int) Option {
	return func(o *options) {
		if n > 0 && n <= math.MaxInt32 {
			o.maxQueue = n
		} else {
			o.maxQueue = defMaxQueue
		}
	}
}

// WithBufferSize returns an Option which sets size of input and output
// buffers. By default it is 1 MB.
func WithBufferSize(n int) Option {
	return func(o *options) {
		if n > 0 {
			o.bufSize = n
		} else {
			o.bufSize = defBufSize
		}
	}
}

// WithWriteInterval returns an Option which sets duration after which data
// from write buffer are sent to network even if write buffer isn't full.
// Default 50 us.
func WithWriteInterval(d time.Duration) Option {
	return func(o *options) {
		if d > 0 {
			o.writeInt = d
		} else {
			o.writeInt = defWriteInt
		}
	}
}

// WithResponseTimeout returns an Option which sets timeout for a response.
// If client gets no response within the interval it drops connection.
func WithResponseTimeout(d time.Duration) Option {
	return func(o *options) {
		if d > 0 {
			o.timeout = d
		} else {
			o.timeout = defTimeout
		}
	}
}

// WithResponseCheckInterval returns an Option which sets inteval of
// timeout checks.
func WithResponseCheckInterval(d time.Duration) Option {
	return func(o *options) {
		if d > 0 {
			o.termInt = d
		} else {
			o.termInt = defTermInt
		}
	}
}

func withTestWriteFlushChannel(ch <-chan time.Time) Option {
	return func(o *options) {
		o.writeFlushCh = ch
	}
}

func withTestTermFlushChannel(ch <-chan time.Time) Option {
	return func(o *options) {
		o.termFlushCh = ch
	}
}

type options struct {
	maxSize  int
	maxQueue int
	bufSize  int
	writeInt time.Duration
	timeout  time.Duration
	termInt  time.Duration

	net  string
	addr string

	writeFlushCh <-chan time.Time
	termFlushCh  <-chan time.Time
}

const (
	defMaxSize  = 10 * 1024
	defMaxQueue = 100
	defBufSize  = 1024 * 1024
	defWriteInt = 50 * time.Microsecond
	defTimeout  = time.Second
	defTermInt  = 50 * time.Microsecond

	defNet  = "tcp"
	defAddr = "localhost:5600"
)

var defaults = options{
	maxSize:  defMaxSize,
	maxQueue: defMaxQueue,
	bufSize:  defBufSize,
	writeInt: defWriteInt,
	timeout:  defTimeout,
	termInt:  defTermInt,

	net:  defNet,
	addr: defAddr,
}
