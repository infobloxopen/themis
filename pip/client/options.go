package client

import "time"

// An Option allows to set PIP client options.
type Option func(*options)

// WithMaxRequestSize returns an Option which limits request size in bytes
// to given value. Default 10KB.
func WithMaxRequestSize(n int) Option {
	return func(o *options) {
		if n > 0 {
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
		if n > 0 {
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

func withTestWriteFlushChannel(ch <-chan time.Time) Option {
	return func(o *options) {
		o.writeFlushCh = ch
	}
}

type options struct {
	maxSize  int
	maxQueue int
	bufSize  int
	writeInt time.Duration

	net  string
	addr string

	writeFlushCh <-chan time.Time
}

const (
	defMaxSize  = 10 * 1024
	defMaxQueue = 100
	defBufSize  = 1024 * 1024
	defWriteInt = 50 * time.Microsecond
)

var defaults = options{
	maxSize:  defMaxSize,
	maxQueue: defMaxQueue,
	bufSize:  defBufSize,
	writeInt: defWriteInt,

	net:  "tcp",
	addr: "localhost:5600",
}
