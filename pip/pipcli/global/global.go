// Package global provides types and methods to parse global options for PIP CLI.
package global

import (
	"flag"
	"fmt"
	"math"
	"os"
	"strings"
	"time"

	"github.com/infobloxopen/themis/pip/client"
)

// Config holds all global options for PIP CLI.
type Config struct {
	// Input holds path to file with requests.
	Input string
	// N is a number of requests to send.
	N int

	// Network stores kind of destination network.
	Network string
	// Address represents destination address.
	Address string

	// Servers is an array of servers used to initialize balancer.
	Servers StringSet
	// RoundRobinBalancer turns on round-robin balancer.
	RoundRobinBalancer bool
	// HotSpotBalancer turns on hot-spot balancer.
	HotSpotBalancer bool

	// DNSRadar turns on discovery by default DNS resolver.
	DNSRadar bool
	// K8sRadar turns on discovery by in-cluster kubernetes tools.
	K8sRadar bool

	// MaxRequestSize limits request size in bytes.
	MaxRequestSize int
	// MaxQueue limits number of requests client can send in parallel.
	MaxQueue int
	// BufferSize defines size of input and output buffers.
	BufferSize int
	// ConnTimeout sets connection timeout.
	ConnTimeout time.Duration
	// WriteInterval is duration after which data from write buffer are sent to
	// network even if write buffer isn't full.
	WriteInterval time.Duration
	// ResponseTimeout sets response timeout.
	ResponseTimeout time.Duration
	// ResponseCheckInterval is an inteval of response timeout checks.
	ResponseCheckInterval time.Duration
	// Requests holds a list of information requests.
	Requests []Request
	// Client is PIP client interface.
	Client client.Client
}

const (
	netUnix  = "unix"
	netTCP   = "tcp"
	netTCPv4 = "tcp4"
	netTCPv6 = "tcp6"

	defMaxSize               = 10 * 1024
	defMaxQueue              = 100
	defBufSize               = 1024 * 1024
	defConnTimeout           = 30 * time.Second
	defWriteInterval         = 50 * time.Microsecond
	defResponseTimeout       = time.Second
	defResponseCheckInterval = 50 * time.Microsecond
)

var validNets = map[string]struct{}{
	netUnix:  {},
	netTCP:   {},
	netTCPv4: {},
	netTCPv6: {},
}

// NewConfigFromCommandLine parses command line arguments and fills Config accordingly.
func NewConfigFromCommandLine(usage func()) *Config {
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	flag.CommandLine.Usage = usage

	conf := new(Config)

	flag.StringVar(&conf.Input, "i", "requests.yaml", "path to input YAML file with requests")
	flag.IntVar(&conf.N, "n", 0, "number of requests to sent (default - all requests from input file)")

	flag.StringVar(&conf.Network, "net", netTCP, "kind of destination network")
	flag.StringVar(&conf.Address, "a", "localhost:5600", "destination address")

	flag.Var(&conf.Servers, "s", "list of servers used to initialize balancer")
	flag.BoolVar(&conf.RoundRobinBalancer, "round-robin", false, "use round-robin balancer")
	flag.BoolVar(&conf.HotSpotBalancer, "hot-spot", false, "use hot-spot balancer")

	flag.BoolVar(&conf.DNSRadar, "dns", false, "use discovery by default DNS resolver")
	flag.BoolVar(&conf.K8sRadar, "k8s", false, "use discovery by in-cluster kubernetes tools")

	flag.IntVar(&conf.MaxRequestSize, "max-request-size", defMaxSize, "limits request size in bytes")
	flag.IntVar(&conf.MaxQueue, "max-queue", defMaxQueue, "number of requests client can send in parallel")
	flag.IntVar(&conf.BufferSize, "buffer-size", defBufSize, "size of input and output buffers")
	flag.DurationVar(&conf.ConnTimeout, "conn-timeout", defConnTimeout, "connection timeout")
	flag.DurationVar(&conf.WriteInterval, "write-interval", defWriteInterval,
		"duration after which data from write buffer are sent to network even if write buffer isn't full")
	flag.DurationVar(&conf.ResponseTimeout, "resp-timeout", defResponseTimeout, "response timeout")
	flag.DurationVar(&conf.ResponseCheckInterval, "check-interval", defResponseCheckInterval,
		"inteval of response timeout checks")

	flag.Parse()

	conf.validateN()
	conf.validateNetwork()
	conf.validateBalancers()
	conf.validateRadars()
	conf.validateMaxRequestSize()
	conf.validateMaxQueue()
	conf.validateBufferSize()
	conf.validateConnTimeout()
	conf.validateWriteInterval()
	conf.validateResponseTimeout()
	conf.validateResponseCheckInterval()

	return conf
}

func (conf *Config) validateN() {
	if conf.N < 0 {
		fmt.Fprintf(os.Stderr, "%d is too small for number of requests. using default...\n", conf.N)
	}
}

func (conf *Config) validateNetwork() {
	network := strings.ToLower(conf.Network)
	if _, ok := validNets[network]; !ok {
		fmt.Fprintf(os.Stderr, "unknown kind of network %q\n", conf.Network)
		flag.Usage()
		os.Exit(2)
	}
	conf.Network = network
}

func (conf *Config) validateBalancers() {
	if len(conf.Servers) > 0 && !conf.RoundRobinBalancer && !conf.HotSpotBalancer {
		fmt.Fprintln(os.Stderr, "got list of servers with no balancer. ignoring...")
	}

	if conf.RoundRobinBalancer && conf.HotSpotBalancer {
		fmt.Fprintln(os.Stderr, "both round-robin and hot-spot balancer choosen")
		flag.Usage()
		os.Exit(2)
	}
}

func (conf *Config) validateRadars() {
	if conf.DNSRadar && conf.K8sRadar {
		fmt.Fprintln(os.Stderr, "both DNS and k8s discovery methods choosen")
		flag.Usage()
		os.Exit(2)
	}
}

func (conf *Config) validateMaxRequestSize() {
	if conf.MaxRequestSize < 4 {
		fmt.Fprintf(os.Stderr, "%d is too small for request size limit. using default...\n", conf.MaxRequestSize)
		conf.MaxRequestSize = defMaxSize
	}

	if conf.MaxRequestSize > math.MaxInt32-4 {
		fmt.Fprintf(os.Stderr, "%d is too big for request size limit. using default...\n", conf.MaxRequestSize)
		conf.MaxRequestSize = defMaxSize
	}
}

func (conf *Config) validateMaxQueue() {
	if conf.MaxQueue <= 0 {
		fmt.Fprintf(os.Stderr, "%d is too small for queue size. using default...\n", conf.MaxQueue)
		conf.MaxQueue = defMaxQueue
	}

	if conf.MaxQueue > math.MaxInt32 {
		fmt.Fprintf(os.Stderr, "%d is too big for queue size. using default...\n", conf.MaxQueue)
		conf.MaxQueue = defMaxQueue
	}
}

func (conf *Config) validateBufferSize() {
	if conf.BufferSize <= 0 {
		fmt.Fprintf(os.Stderr, "%d is too small for buffer size. using default...\n", conf.BufferSize)
		conf.BufferSize = defBufSize
	}
}

func (conf *Config) validateConnTimeout() {
	if conf.ConnTimeout < 0 {
		fmt.Fprintf(os.Stderr, "%s is too small for connection timeout. using default...\n", conf.ConnTimeout)
		conf.ConnTimeout = defConnTimeout
	}
}

func (conf *Config) validateWriteInterval() {
	if conf.WriteInterval < 0 {
		fmt.Fprintf(os.Stderr, "%s is too small for write interval. using default...\n", conf.WriteInterval)
		conf.WriteInterval = defWriteInterval
	}
}

func (conf *Config) validateResponseTimeout() {
	if conf.ResponseTimeout < 0 {
		fmt.Fprintf(os.Stderr, "%s is too small for response timeout. using default...\n", conf.ResponseTimeout)
		conf.ResponseTimeout = defResponseTimeout
	}
}

func (conf *Config) validateResponseCheckInterval() {
	if conf.ResponseCheckInterval < 0 {
		fmt.Fprintf(os.Stderr, "%s is too small for timeout check interval. using default...\n",
			conf.ResponseCheckInterval)
		conf.ResponseCheckInterval = defResponseCheckInterval
	}
}
