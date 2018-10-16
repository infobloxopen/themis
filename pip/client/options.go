package client

import (
	"math"
	"net"
	"time"
)

// An Option allows to set such options as address, balancer and so on.
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

// WithRoundRobinBalancer returns an Option which sets round robin balancer.
// If no addresses provided with the option client connects to all IP addresses
// it can get by host name from WithAddress option. For "unix" network
// the option is ignored.
func WithRoundRobinBalancer(addrs ...string) Option {
	return func(o *options) {
		o.balancer = balancerTypeRoundRobin
		if len(addrs) > 0 {
			o.addrs = addrs
		}
	}
}

// WithHotSpotBalancer returns an Option which sets hot spot balancer.
// The balancer puts requests to the same connection until its queue is full
// and then goes to the next connection. If no addresses provided with
// the option client connects to all IP addresses it can get by host name from
// WithAddress option. For "unix" network the option is ignored.
func WithHotSpotBalancer(addrs ...string) Option {
	return func(o *options) {
		o.balancer = balancerTypeHotSpot
		if len(addrs) > 0 {
			o.addrs = addrs
		}
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

// ConnErrHandler is a function to process errors within a connection.
type ConnErrHandler func(net.Addr, error)

// WithConnErrHandler returns an Option which sets custom handler for transport
// errors.
func WithConnErrHandler(f ConnErrHandler) Option {
	return func(o *options) {
		o.onErr = f
	}
}

// WithConnTimeout returns an Option which sets connection timeout.
func WithConnTimeout(d time.Duration) Option {
	return func(o *options) {
		o.connTimeout = d
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
	maxSize            int
	maxQueue           int
	bufSize            int
	onErr              ConnErrHandler
	connTimeout        time.Duration
	connAttemptTimeout time.Duration
	keepAlive          time.Duration
	writeInt           time.Duration
	timeout            time.Duration
	termInt            time.Duration

	net  string
	addr string

	addrs    []string
	balancer int

	writeFlushCh <-chan time.Time
	termFlushCh  <-chan time.Time
}

const (
	defMaxSize            = 10 * 1024
	defMaxQueue           = 100
	defBufSize            = 1024 * 1024
	defConnTimeout        = 30 * time.Second
	defConnAttemptTimeout = 5 * time.Second
	defKeepAlive          = 5 * time.Second
	defWriteInt           = 50 * time.Microsecond
	defTimeout            = time.Second
	defTermInt            = 50 * time.Microsecond

	defNet  = "tcp"
	defAddr = "localhost:5600"
	defPort = "5600"

	defBalancer = balancerTypeSimple
)

const (
	balancerTypeSimple = iota
	balancerTypeRoundRobin
	balancerTypeHotSpot
)

var defaults = options{
	maxSize:            defMaxSize,
	maxQueue:           defMaxQueue,
	bufSize:            defBufSize,
	connTimeout:        defConnTimeout,
	connAttemptTimeout: defConnAttemptTimeout,
	keepAlive:          defKeepAlive,
	writeInt:           defWriteInt,
	timeout:            defTimeout,
	termInt:            defTermInt,

	net:  defNet,
	addr: defAddr,

	balancer: defBalancer,
}
