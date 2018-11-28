package client

import (
	"math"
	"net"
	"time"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	"github.com/infobloxopen/themis/pdp"
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

// WithDNSRadar returns an Option which turns on DNS discovery (DNS radar).
// The radar periodicaly looks up for IP addresses using system resolver
// (works only for TCP networks). It uses an address given by WithAddress option
// as host name for DNS qurey. If the query gives several IP addresses,
// PIP client balances load using balancer given by With*Balancer option.
// Addresses which come from With*Balancer option are used only as initial
// set of addresses to connect to. If further DNS queries don't return those
// addresses client stops using them.
func WithDNSRadar() Option {
	return func(o *options) {
		o.radar = radarDNS
	}
}

// WithK8sRadar returns an Option which turns on kubernetes discovery.
// The discovery works only inside kubernetes cluster and requires "get",
// "watch" and "list" access to "pods" resource. It treats address from
// WithAddress option as selector encoded in following form
// "<valueN>.<keyN>. ... .<value2>.<key2>.<value1>.<key1>.<namespacea>"
// (it should contain at least one key value pair and namespace). Similarly
// to DNS in case of several pods given balancer is used and addresses coming
// with balancer options become initiall addresses. However the discovery drops
// connection to an initiall address only if kubernetes tells that the pod with
// the IP goes down.
func WithK8sRadar() Option {
	return func(o *options) {
		o.radar = radarK8s
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

// WithCacheTTL returns an Option which adds cache with given TTL for cached
// requests. Cache size isn't limited in the case and can consume all available
// memory on machine.
func WithCacheTTL(d time.Duration) Option {
	return func(o *options) {
		o.cache = true

		if d >= 0 {
			o.cacheTTL = d
		} else {
			o.cacheTTL = 0
		}
	}
}

// WithCacheTTLAndMaxSize returns an Option which adds cache with given TTL
// and size limit for entire cache in MB. When the limit is reached then new
// requests override the oldest ones.
func WithCacheTTLAndMaxSize(d time.Duration, size int) Option {
	return func(o *options) {
		o.cache = true

		if d >= 0 {
			o.cacheTTL = d
		} else {
			o.cacheTTL = 0
		}

		if size >= 0 {
			o.cacheMaxSize = size
		} else {
			o.cacheMaxSize = 0
		}
	}
}

// CacheHitHandler defines a function prototype to call on each cache hit
// if cache has been enabled.
type CacheHitHandler func(path string, args []pdp.AttributeValue, v pdp.AttributeValue, err error)

// WithCacheHitHandler returns an Option which sets handler for cache hits.
func WithCacheHitHandler(h CacheHitHandler) Option {
	return func(o *options) {
		o.onCache = h
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

type options struct {
	maxSize            int
	maxQueue           int
	bufSize            int
	onErr              ConnErrHandler
	cache              bool
	cacheTTL           time.Duration
	cacheMaxSize       int
	onCache            CacheHitHandler
	radar              int
	radarInt           time.Duration
	connTimeout        time.Duration
	connAttemptTimeout time.Duration
	keepAlive          time.Duration
	writeInt           time.Duration
	timeout            time.Duration
	termInt            time.Duration
	k8sClientMaker     func() (kubernetes.Interface, error)

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
	defDNSRadarInt        = time.Second
	defK8sRadarInt        = time.Minute
	defConnTimeout        = 30 * time.Second
	defConnAttemptTimeout = 5 * time.Second
	defKeepAlive          = 5 * time.Second
	defWriteInt           = 50 * time.Microsecond
	defTimeout            = time.Second
	defTermInt            = 50 * time.Microsecond

	defNet  = "tcp"
	defAddr = "localhost:5600"
	defPort = "5600"

	defRadar    = radarNone
	defBalancer = balancerTypeSimple

	unixNet = "unix"
)

const (
	radarNone = iota
	radarDNS
	radarK8s
)

const (
	balancerTypeSimple = iota
	balancerTypeRoundRobin
	balancerTypeHotSpot
)

var (
	defK8sConfig = rest.InClusterConfig

	defaults = options{
		maxSize:            defMaxSize,
		maxQueue:           defMaxQueue,
		bufSize:            defBufSize,
		radar:              defRadar,
		connTimeout:        defConnTimeout,
		connAttemptTimeout: defConnAttemptTimeout,
		keepAlive:          defKeepAlive,
		writeInt:           defWriteInt,
		timeout:            defTimeout,
		termInt:            defTermInt,
		k8sClientMaker:     makeInClusterK8sClient,

		net:  defNet,
		addr: defAddr,

		balancer: defBalancer,
	}
)

func makeOptions(opts []Option) options {
	o := defaults
	for _, opt := range opts {
		opt(&o)
	}

	if o.maxSize+msgSizeBytes+msgIdxBytes > o.bufSize {
		o.bufSize = defBufSize
	}

	switch o.radar {
	case radarDNS:
		if o.radarInt <= 0 {
			o.radarInt = defDNSRadarInt
		}

	case radarK8s:
		if o.radarInt <= 0 {
			o.radarInt = defK8sRadarInt
		}
	}

	return o
}
