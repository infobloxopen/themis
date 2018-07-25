// Package pipclient implements gRPC client for Policy Information Point (PIP)
// server. PIPClient package wraps golang gRPC protocol implementation.
// The protocol is defined by github.com/infobloxopen/themis/proto/pip.proto.
// Its golang implementation can be found at
// github.com/infobloxopen/themis/pip-service. PIPClient is able to work with
// single server as well as multiple servers balancing requests using different
// approaches to balance load.
package pipclient

//go:generate bash -c "mkdir -p $GOPATH/src/github.com/infobloxopen/themis/pip-service && protoc -I $GOPATH/src/github.com/infobloxopen/themis/proto/ $GOPATH/src/github.com/infobloxopen/themis/proto/pip.proto --go_out=plugins=grpc:$GOPATH/src/github.com/infobloxopen/themis/pip-service && ls $GOPATH/src/github.com/infobloxopen/themis/pip-service"

import (
	"errors"
	"time"
)

var (
	// ErrorConnected occurs if method connect is called after connection has been established.
	ErrorConnected = errors.New("connection has been already established")
	// ErrorNotConnected indicates that there is no connection established to PIP server.
	ErrorNotConnected = errors.New("no connection")
	// ErrorHotSpotBalancerUnsupported returned by attempt to make unary connection with
	// "hot spot" balancer.
	ErrorHotSpotBalancerUnsupported = errors.New("\"hot spot\" balancer isn't supported by unary gRPC client")
)

// Client defines abstract PIP service client interface.
type Client interface {
	// Connect establishes connection to given PIP server. It ignores address
	// parameter if balancer is provided.
	Connect(addr string) error
	// Close terminates previously established connection if any.
	// Close should silently return if connection hasn't been established yet or
	// if it has been already closed.
	Close()

	// Map sends mapping request to PIP server and fills out response.
	Map(in, out interface{}) error

	// GetCustomData returns data bound to client using WithCustomData option.
	GetCustomData() interface{}
}

// An Option sets such options as balancer, number of streams and so on.
type Option func(*options)

const virtualServerAddress = "pip"

// WithCustomData returns an Option which binds given value to a client.
func WithCustomData(data interface{}) Option {
	return func(o *options) {
		o.customData = data
	}
}

// WithRoundRobinBalancer returns an Option which sets round-robin balancer with given set of servers.
func WithRoundRobinBalancer(addresses ...string) Option {
	return func(o *options) {
		o.balancer = roundRobinBalancer
		o.addresses = addresses
	}
}

// WithHotSpotBalancer returns an Option which sets "hot spot" balancer with given set of servers
// (the balancer can be applied for gRPC streaming connection).
func WithHotSpotBalancer(addresses ...string) Option {
	return func(o *options) {
		o.balancer = hotSpotBalancer
		o.addresses = addresses
	}
}

// WithStreams returns an Option which sets number of gRPC streams to run in parallel.
func WithStreams(n int) Option {
	return func(o *options) {
		o.maxStreams = n
	}
}

// WithConnectionTimeout returns an Option which sets validation timeout
// for the case when no connection can be established. Negative value means
// no timeout. Zero - don't wait for connection, fail immediately.
func WithConnectionTimeout(timeout time.Duration) Option {
	return func(o *options) {
		o.connTimeout = timeout
	}
}

// WithConnectionStateNotification returns an Option which sets connection
// state notification callback. The callback is called before connection
// attempt with state StreamingConnectionConnecting, on successfull connect
// with state StreamingConnectionEstablished. If connection attempt fails
// the callback is called with state StreamingConnectionFailure and with error
// occured during the attempt. State StreamingConnectionBroken is used when
// during request validation connection to any PDP server appears not working.
func WithConnectionStateNotification(callback ConnectionStateNotificationCallback) Option {
	return func(o *options) {
		o.connStateCb = callback
	}
}

// WithAutoRequestSize returns an Option which makes client automatically
// allocate buffer for decision request. By default request size is limited
// by 10KB. When the option is set MaxRequestSize is still used to determine
// cache limit.
func WithAutoRequestSize(b bool) Option {
	return func(o *options) {
		o.autoRequestSize = b
	}
}

// WithMaxRequestSize returns an Option which limits request size in bytes
// to given value. Default 10KB. WithAutoRequestSize overrides the option but
// it still affects cache size.
func WithMaxRequestSize(size uint32) Option {
	return func(o *options) {
		o.maxRequestSize = size
	}
}

// WithNoRequestBufferPool returns an Option which makes client allocate new
// buffer for each request.
func WithNoRequestBufferPool() Option {
	return func(o *options) {
		o.noPool = true
	}
}

// WithCacheTTL returns an Option which adds cache with given TTL for cached
// requests. Cache size isn't limited in the case and can consume all available
// memory on machine.
func WithCacheTTL(ttl time.Duration) Option {
	return func(o *options) {
		o.cache = true
		o.cacheTTL = ttl
	}
}

// WithCacheTTLAndMaxSize returns an Option which adds cache with given TTL
// and size limit for entire cache in MB. When the limit is reached then new
// requests override the oldest ones.
func WithCacheTTLAndMaxSize(ttl time.Duration, size int) Option {
	return func(o *options) {
		o.cache = true
		o.cacheTTL = ttl
		o.cacheMaxSize = size
	}
}

const (
	noBalancer = iota
	roundRobinBalancer
	hotSpotBalancer
)

type options struct {
	addresses       []string
	balancer        int
	maxStreams      int
	connTimeout     time.Duration
	connStateCb     ConnectionStateNotificationCallback
	autoRequestSize bool
	maxRequestSize  uint32
	noPool          bool
	cache           bool
	cacheTTL        time.Duration
	cacheMaxSize    int
	customData      interface{}
}

// NewClient creates client instance using given options.
func NewClient(opts ...Option) Client {
	o := options{
		connTimeout:    -1,
		maxRequestSize: 10240,
	}
	for _, opt := range opts {
		opt(&o)
	}

	if o.maxStreams > 0 {
		return newStreamingClient(o)
	}

	return newUnaryClient(o)
}
