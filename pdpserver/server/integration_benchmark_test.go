package server

import (
	"bytes"
	"fmt"
	"math/rand"
	"net"
	"strings"
	"testing"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/infobloxopen/themis/pdp"
	pb "github.com/infobloxopen/themis/pdp-service"
	_ "github.com/infobloxopen/themis/pdp/selector"
	"github.com/infobloxopen/themis/pep"
)

type decisionRequest struct {
	Direction string `pdp:"k1"`
	Policy    string `pdp:"k2"`
	Domain    string `pdp:"k3,domain"`
}

type decisionResponse struct {
	Effect int    `pdp:"Effect"`
	Reason error  `pdp:"Reason"`
	X      string `pdp:"x"`
}

func (r decisionResponse) String() string {
	if r.Reason != nil {
		return fmt.Sprintf("Effect: %q, Reason: %q, X: %q",
			pdp.EffectNameFromEnum(r.Effect),
			r.Reason,
			r.X,
		)
	}

	return fmt.Sprintf("Effect: %q, X: %q", pdp.EffectNameFromEnum(r.Effect), r.X)
}

var (
	decisionRequests []decisionRequest
	rawRequests      []pb.Msg
)

type testRequest3Keys struct {
	k1 string `pdp:"k1"`
	k2 string `pdp:"k2"`
	k3 string `pdp:"k3,domain"`
}

func init() {
	decisionRequests = make([]decisionRequest, 0x40000)
	for i := range decisionRequests {
		decisionRequests[i] = decisionRequest{
			Direction: directionOpts[rand.Intn(len(directionOpts))],
			Policy:    policySetOpts[rand.Intn(len(policySetOpts))],
			Domain:    domainOpts[rand.Intn(len(domainOpts))],
		}
	}

	rawRequests = make([]pb.Msg, len(decisionRequests))
	for i := range rawRequests {
		b := make([]byte, 128)

		m, err := pep.MakeRequestWithBuffer(testRequest3Keys{
			k1: directionOpts[rand.Intn(len(directionOpts))],
			k2: policySetOpts[rand.Intn(len(policySetOpts))],
			k3: domainOpts[rand.Intn(len(domainOpts))],
		}, b)
		if err != nil {
			panic(fmt.Errorf("failed to create %d raw request: %s", i+1, err))
		}

		rawRequests[i] = m
	}

}

func benchmarkIntPolicySet(name, p string, b *testing.B, opts ...pep.Option) {
	pdpServer, _, c := startPDPServer(p, nil, b, opts...)
	defer func() {
		c.Close()
		if logs := pdpServer.Stop(); len(logs) > 0 {
			b.Logf("server logs:\n%s", logs)
		}
	}()

	b.Run(name, func(b *testing.B) {
		for n := 0; n < b.N; n++ {
			in := decisionRequests[n%len(decisionRequests)]

			var out decisionResponse
			c.Validate(in, &out)

			if (out.Effect != pdp.EffectDeny &&
				out.Effect != pdp.EffectPermit &&
				out.Effect != pdp.EffectNotApplicable) ||
				out.Reason != nil {
				b.Fatalf("unexpected response: %s", out)
			}
		}
	})
}

func BenchmarkStagesPolicySet(b *testing.B) {
	benchmarkIntPolicySet("OneStagePolicySet", oneStageBenchmarkPolicySet, b)
	benchmarkIntPolicySet("TwoStagePolicySet", twoStageBenchmarkPolicySet, b)
	benchmarkIntPolicySet("ThreeStagePolicySet", threeStageBenchmarkPolicySet, b)
	benchmarkIntPolicySet("AutoRequestSize", threeStageBenchmarkPolicySet, b, pep.WithAutoRequestSize(true))
}

func BenchmarkUnaryRaw(b *testing.B) {
	pdpServer, _, c := startPDPServer(threeStageBenchmarkPolicySet, nil, b)
	defer func() {
		c.Close()
		if logs := pdpServer.Stop(); len(logs) > 0 {
			b.Logf("server logs:\n%s", logs)
		}
	}()

	name := "UnaryRaw"

	b.Run(name, func(b *testing.B) {
		var (
			out        pdp.Response
			assignment [16]pdp.AttributeAssignment
		)
		for n := 0; n < b.N; n++ {
			in := rawRequests[n%len(rawRequests)]

			out.Obligations = assignment[:]
			c.Validate(in, &out)

			err := assertBenchMsg(&out, "%q request %d", name, n)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

func BenchmarkUnaryWithCache(b *testing.B) {
	pdpServer, _, c := startPDPServer(threeStageBenchmarkPolicySet, nil, b,
		pep.WithMaxRequestSize(128),
		pep.WithCacheTTL(15*time.Minute),
	)
	defer func() {
		c.Close()
		if logs := pdpServer.Stop(); len(logs) > 0 {
			b.Logf("server logs:\n%s", logs)
		}
	}()

	name := "UnaryWithCache"

	cc := 10
	var (
		out        pdp.Response
		assignment [16]pdp.AttributeAssignment
	)
	for n := 0; n < cc; n++ {
		in := rawRequests[n%cc]

		out.Obligations = assignment[:]
		c.Validate(in, &out)

		err := assertBenchMsg(&out, "%q request %d", name, n)
		if err != nil {
			b.Fatal(err)
		}
	}

	b.Run(name, func(b *testing.B) {
		var (
			out        pdp.Response
			assignment [16]pdp.AttributeAssignment
		)
		for n := 0; n < b.N; n++ {
			in := rawRequests[n%cc]

			out.Obligations = assignment[:]
			c.Validate(in, &out)

			err := assertBenchMsg(&out, "%q request %d", name, n)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

func benchmarkStreamingClient(name string, b *testing.B, opts ...pep.Option) {
	streams := 96

	opts = append(opts,
		pep.WithStreams(streams),
	)
	pdpSrv, _, c := startPDPServer(threeStageBenchmarkPolicySet, nil, b, opts...)
	defer func() {
		c.Close()
		if logs := pdpSrv.Stop(); len(logs) > 0 {
			b.Logf("primary server logs:\n%s", logs)
		}
	}()

	b.Run(name, func(b *testing.B) {
		assignments := make(chan []pdp.AttributeAssignment, streams)
		for i := 0; i < cap(assignments); i++ {
			assignments <- make([]pdp.AttributeAssignment, 16)
		}

		th := make(chan int, streams)
		for n := 0; n < b.N; n++ {
			th <- 0
			go func(i int) {
				defer func() { <-th }()

				var out pdp.Response

				assignment := <-assignments
				defer func() { assignments <- assignment }()
				out.Obligations = assignment

				c.Validate(decisionRequests[i%len(decisionRequests)], &out)
				if err := assertBenchMsg(&out, "%q request %d", name, i); err != nil {
					panic(err)
				}
			}(n)
		}
	})
}

func BenchmarkStreamingRequestSize(b *testing.B) {
	benchmarkStreamingClient("StreamingFixedRequestSize", b)
	benchmarkStreamingClient("StreamingAutoRequestSize", b, pep.WithAutoRequestSize(true))
}

func benchmarkRawStreamingClient(name string, ports []uint16, b *testing.B, opts ...pep.Option) {
	if len(ports) != 0 && len(ports) != 2 {
		b.Fatalf("only 0 for single PDP and 2 for 2 PDP ports supported but got %d", len(ports))
	}

	streams := 96

	opts = append(opts,
		pep.WithStreams(streams),
	)
	pdpSrv, pdpSrvAlt, c := startPDPServer(threeStageBenchmarkPolicySet, ports, b, opts...)
	defer func() {
		c.Close()
		if logs := pdpSrv.Stop(); len(logs) > 0 {
			b.Logf("primary server logs:\n%s", logs)
		}
		if pdpSrvAlt != nil {
			if logs := pdpSrvAlt.Stop(); len(logs) > 0 {
				b.Logf("secondary server logs:\n%s", logs)
			}
		}
	}()

	b.Run(name, func(b *testing.B) {
		assignments := make(chan []pdp.AttributeAssignment, streams)
		for i := 0; i < cap(assignments); i++ {
			assignments <- make([]pdp.AttributeAssignment, 16)
		}

		th := make(chan int, streams)
		for n := 0; n < b.N; n++ {
			th <- 0
			go func(i int) {
				defer func() { <-th }()

				var out pdp.Response

				assignment := <-assignments
				defer func() { assignments <- assignment }()
				out.Obligations = assignment

				c.Validate(rawRequests[i%len(rawRequests)], &out)
				if err := assertBenchMsg(&out, "%q request %d", name, i); err != nil {
					panic(err)
				}
			}(n)
		}
	})
}

func BenchmarkRawStreamingClient(b *testing.B) {
	benchmarkRawStreamingClient("StreamingClient", nil, b)
	benchmarkRawStreamingClient("RoundRobinStreamingClient",
		[]uint16{
			5555,
			5556,
		},
		b,
		pep.WithRoundRobinBalancer("127.0.0.1:5555", "127.0.0.1:5556"),
	)
	benchmarkRawStreamingClient("HotSpotStreamingClient",
		[]uint16{
			5555,
			5556,
		},
		b,
		pep.WithHotSpotBalancer("127.0.0.1:5555", "127.0.0.1:5556"),
	)
}

func BenchmarkRawStreamingClientWithCache(b *testing.B) {
	streams := 96

	pdpServer, _, c := startPDPServer(threeStageBenchmarkPolicySet, nil, b,
		pep.WithStreams(streams),
		pep.WithMaxRequestSize(128),
		pep.WithCacheTTL(15*time.Minute),
	)
	defer func() {
		c.Close()
		if logs := pdpServer.Stop(); len(logs) > 0 {
			b.Logf("server logs:\n%s", logs)
		}
	}()

	name := "RawStreamingClientWithCache"

	cc := 10 * streams
	var (
		out        pdp.Response
		assignment [16]pdp.AttributeAssignment
	)
	for n := 0; n < cc; n++ {
		in := rawRequests[n%cc]

		out.Obligations = assignment[:]
		c.Validate(in, &out)

		err := assertBenchMsg(&out, "%q request %d", name, n)
		if err != nil {
			b.Fatal(err)
		}
	}

	b.Run(name, func(b *testing.B) {
		assignments := make(chan []pdp.AttributeAssignment, streams)
		for i := 0; i < cap(assignments); i++ {
			assignments <- make([]pdp.AttributeAssignment, 16)
		}

		th := make(chan int, streams)
		for n := 0; n < b.N; n++ {
			th <- 0
			go func(i int) {
				defer func() { <-th }()

				var out pdp.Response

				assignment := <-assignments
				defer func() { assignments <- assignment }()
				out.Obligations = assignment

				c.Validate(rawRequests[i%cc], &out)
				if err := assertBenchMsg(&out, "%q request %d", name, i); err != nil {
					panic(err)
				}
			}(n)
		}
	})
}

func benchmarkStreamingClientWithCache(name string, b *testing.B, opts ...pep.Option) {
	streams := 96

	opts = append(opts,
		pep.WithStreams(streams),
		pep.WithMaxRequestSize(128),
		pep.WithCacheTTL(15*time.Minute),
	)

	pdpServer, _, c := startPDPServer(threeStageBenchmarkPolicySet, nil, b, opts...)
	defer func() {
		c.Close()
		if logs := pdpServer.Stop(); len(logs) > 0 {
			b.Logf("server logs:\n%s", logs)
		}
	}()

	cc := 10 * streams
	var (
		out        pdp.Response
		assignment [16]pdp.AttributeAssignment
	)
	for n := 0; n < cc; n++ {
		in := decisionRequests[n%cc]

		out.Obligations = assignment[:]
		c.Validate(in, &out)

		err := assertBenchMsg(&out, "%q request %d", name, n)
		if err != nil {
			b.Fatal(err)
		}
	}

	b.Run(name, func(b *testing.B) {
		assignments := make(chan []pdp.AttributeAssignment, streams)
		for i := 0; i < cap(assignments); i++ {
			assignments <- make([]pdp.AttributeAssignment, 16)
		}

		th := make(chan int, streams)
		for n := 0; n < b.N; n++ {
			th <- 0
			go func(i int) {
				defer func() { <-th }()

				var out pdp.Response

				assignment := <-assignments
				defer func() { assignments <- assignment }()
				out.Obligations = assignment

				c.Validate(decisionRequests[i%cc], &out)
				if err := assertBenchMsg(&out, "%q request %d", name, i); err != nil {
					panic(err)
				}
			}(n)
		}
	})
}

func BenchmarkStreamingClientWithCache(b *testing.B) {
	benchmarkStreamingClientWithCache("FixedRequestSize", b)
	benchmarkStreamingClientWithCache("AutoRequestSize", b, pep.WithAutoRequestSize(true))
}

func benchmarkAutoResponseServer(name string, b *testing.B, opts ...pep.Option) {
	streams := 96

	opts = append(opts,
		pep.WithStreams(streams),
	)
	pdpSrv, c := startPDPServerWithAutoResponse(threeStageBenchmarkPolicySet, b, opts...)
	defer func() {
		c.Close()
		if logs := pdpSrv.Stop(); len(logs) > 0 {
			b.Logf("primary server logs:\n%s", logs)
		}
	}()

	b.Run(name, func(b *testing.B) {
		assignments := make(chan []pdp.AttributeAssignment, streams)
		for i := 0; i < cap(assignments); i++ {
			assignments <- make([]pdp.AttributeAssignment, 16)
		}

		th := make(chan int, streams)
		for n := 0; n < b.N; n++ {
			th <- 0
			go func(i int) {
				defer func() { <-th }()

				var out pdp.Response

				assignment := <-assignments
				defer func() { assignments <- assignment }()
				out.Obligations = assignment

				c.Validate(rawRequests[i%len(rawRequests)], &out)
				if err := assertBenchMsg(&out, "%q request %d", name, i); err != nil {
					panic(err)
				}
			}(n)
		}
	})
}

func BenchmarkAutoResponseServer(b *testing.B) {
	benchmarkAutoResponseServer("StreamingClient", b)
	benchmarkAutoResponseServer("RoundRobinStreamingClient", b,
		pep.WithRoundRobinBalancer("127.0.0.1:5555"),
	)
	benchmarkAutoResponseServer("HotSpotStreamingClient", b,
		pep.WithHotSpotBalancer("127.0.0.1:5555"),
	)
}

func BenchmarkUnaryRouting(b *testing.B) {
	router, srv, c := startRoutingPDPServers(b, 0)
	defer func() {
		c.Close()
		if logs := router.Stop(); len(logs) > 0 {
			b.Logf("router logs:\n%s", logs)
		}
		if logs := srv.Stop(); len(logs) > 0 {
			b.Logf("server logs:\n%s", logs)
		}
	}()

	name := "UnaryRaw"

	b.Run(name, func(b *testing.B) {
		var (
			out        pdp.Response
			assignment [16]pdp.AttributeAssignment
		)
		for n := 0; n < b.N; n++ {
			in := rawRequests[n%len(rawRequests)]

			out.Obligations = assignment[:]
			c.Validate(in, &out)

			err := assertBenchMsg(&out, "%q request %d", name, n)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

func BenchmarkStreamingRouting(b *testing.B) {
	streams := 96

	router, srv, c := startRoutingPDPServers(b, streams, pep.WithStreams(streams))
	defer func() {
		c.Close()
		if logs := router.Stop(); len(logs) > 0 {
			b.Logf("router logs:\n%s", logs)
		}
		if logs := srv.Stop(); len(logs) > 0 {
			b.Logf("server logs:\n%s", logs)
		}
	}()

	name := "StreamingRaw"

	b.Run(name, func(b *testing.B) {
		assignments := make(chan []pdp.AttributeAssignment, streams)
		for i := 0; i < cap(assignments); i++ {
			assignments <- make([]pdp.AttributeAssignment, 16)
		}

		th := make(chan int, streams)
		for n := 0; n < b.N; n++ {
			th <- 0
			go func(i int) {
				defer func() { <-th }()

				var out pdp.Response

				assignment := <-assignments
				defer func() { assignments <- assignment }()
				out.Obligations = assignment

				c.Validate(decisionRequests[i%len(decisionRequests)], &out)
				if err := assertBenchMsg(&out, "%q request %d", name, i); err != nil {
					panic(err)
				}
			}(n)
		}
	})
}

func waitForPortOpened(address string) error {
	var (
		c   net.Conn
		err error
	)

	for i := 0; i < 20; i++ {
		after := time.After(500 * time.Millisecond)
		c, err = net.DialTimeout("tcp", address, 500*time.Millisecond)
		if err == nil {
			return c.Close()
		}

		<-after
	}

	return err
}

func waitForPortClosed(address string) error {
	var (
		c   net.Conn
		err error
	)

	for i := 0; i < 20; i++ {
		after := time.After(500 * time.Millisecond)
		c, err = net.DialTimeout("tcp", address, 500*time.Millisecond)
		if err != nil {
			return nil
		}

		c.Close()
		<-after
	}

	return fmt.Errorf("port at %s hasn't been closed yet", address)
}

type loggedServer struct {
	s *Server
	b *bytes.Buffer
}

func newServer(opts ...Option) *loggedServer {
	s := &loggedServer{
		b: new(bytes.Buffer),
	}

	logger := log.New()
	logger.Out = s.b
	logger.Level = log.ErrorLevel
	opts = append(opts,
		WithLogger(logger),
	)

	s.s = NewServer(opts...)
	return s
}

func (s *loggedServer) Stop() string {
	s.s.Stop()
	return s.b.String()
}

func startPDPServerWithAutoResponse(p string, b *testing.B, opts ...pep.Option) (*loggedServer, pep.Client) {
	service := "127.0.0.1:5555"

	s := newServer(
		WithServiceAt(service),
		WithAutoResponseSize(true),
	)

	if err := s.s.ReadPolicies(strings.NewReader(p)); err != nil {
		b.Fatalf("can't read policies: %s", err)
	}

	if err := s.s.ReadContent(strings.NewReader(benchmarkContent)); err != nil {
		b.Fatalf("can't read content: %s", err)
	}

	if err := waitForPortClosed(service); err != nil {
		b.Fatalf("port still in use: %s", err)
	}
	go func() {
		if err := s.s.Serve(); err != nil {
			b.Fatalf("primary server failed: %s", err)
		}
	}()

	if err := waitForPortOpened(service); err != nil {
		if logs := s.Stop(); len(logs) > 0 {
			b.Logf("primary server logs:\n%s", logs)
		}

		b.Fatalf("can't connect to PDP server: %s", err)
	}

	c := pep.NewClient(opts...)
	if err := c.Connect(service); err != nil {
		if logs := s.Stop(); len(logs) > 0 {
			b.Logf("primary server logs:\n%s", logs)
		}

		b.Fatalf("can't connect to PDP server: %s", err)
	}

	return s, c
}

func startPDPServer(p string, ports []uint16, b *testing.B, opts ...pep.Option) (*loggedServer, *loggedServer, pep.Client) {
	var (
		primary   *loggedServer
		secondary *loggedServer
	)

	service := "127.0.0.1:5555"
	if len(ports) > 0 {
		service = fmt.Sprintf("127.0.0.1:%d", ports[0])
	}

	primary = newServer(
		WithServiceAt(service),
	)

	if err := primary.s.ReadPolicies(strings.NewReader(p)); err != nil {
		b.Fatalf("can't read policies: %s", err)
	}

	if err := primary.s.ReadContent(strings.NewReader(benchmarkContent)); err != nil {
		b.Fatalf("can't read content: %s", err)
	}

	if err := waitForPortClosed(service); err != nil {
		b.Fatalf("port still in use: %s", err)
	}
	go func() {
		if err := primary.s.Serve(); err != nil {
			b.Fatalf("primary server failed: %s", err)
		}
	}()

	if err := waitForPortOpened(service); err != nil {
		if logs := primary.Stop(); len(logs) > 0 {
			b.Logf("primary server logs:\n%s", logs)
		}

		b.Fatalf("can't connect to PDP server: %s", err)
	}

	if len(ports) > 1 {
		service := fmt.Sprintf("127.0.0.1:%d", ports[1])
		secondary = newServer(
			WithServiceAt(service),
		)

		if err := secondary.s.ReadPolicies(strings.NewReader(p)); err != nil {
			if logs := primary.Stop(); len(logs) > 0 {
				b.Logf("primary server logs:\n%s", logs)
			}
			b.Fatalf("can't read policies: %s", err)
		}

		if err := secondary.s.ReadContent(strings.NewReader(benchmarkContent)); err != nil {
			if logs := primary.Stop(); len(logs) > 0 {
				b.Logf("primary server logs:\n%s", logs)
			}
			b.Fatalf("can't read content: %s", err)
		}

		if err := waitForPortClosed(service); err != nil {
			b.Fatalf("port still in use: %s", err)
		}
		go func() {
			if err := secondary.s.Serve(); err != nil {
				b.Fatalf("secondary server failed: %s", err)
			}
		}()

		if err := waitForPortOpened(service); err != nil {
			if logs := secondary.Stop(); len(logs) > 0 {
				b.Logf("secondary server logs:\n%s", logs)
			}
			if logs := primary.Stop(); len(logs) > 0 {
				b.Logf("primary server logs:\n%s", logs)
			}

			b.Fatalf("can't connect to PDP server: %s", err)
		}
	}

	c := pep.NewClient(opts...)
	if err := c.Connect(service); err != nil {
		if secondary != nil {
			if logs := secondary.Stop(); len(logs) > 0 {
				b.Logf("secondary server logs:\n%s", logs)
			}
		}
		if logs := primary.Stop(); len(logs) > 0 {
			b.Logf("primary server logs:\n%s", logs)
		}

		b.Fatalf("can't connect to PDP server: %s", err)
	}

	return primary, secondary, c
}

func startRoutingPDPServers(b *testing.B, streams int, opts ...pep.Option) (*loggedServer, *loggedServer, pep.Client) {
	router := newServer(
		WithServiceAt("127.0.0.1:5555"),
		WithShards(
			Shard{
				Name:      "A",
				Streams:   streams,
				Addresses: []string{"127.0.0.1:5556"},
			},
		),
	)

	if err := router.s.ReadPolicies(strings.NewReader(oneStageBenchmarkPolicySet)); err != nil {
		b.Fatalf("can't read policies: %s", err)
	}

	if err := router.s.ReadContent(strings.NewReader(benchmarkContent)); err != nil {
		b.Fatalf("can't read content: %s", err)
	}

	if err := waitForPortClosed("127.0.0.1:5555"); err != nil {
		b.Fatalf("port still in use: %s", err)
	}
	go func() {
		if err := router.s.Serve(); err != nil {
			b.Fatalf("router server failed: %s", err)
		}
	}()

	if err := waitForPortOpened("127.0.0.1:5555"); err != nil {
		if logs := router.Stop(); len(logs) > 0 {
			b.Logf("router server logs:\n%s", logs)
		}

		b.Fatalf("can't connect to PDP server: %s", err)
	}

	primary := newServer(
		WithServiceAt("127.0.0.1:5556"),
	)

	if err := primary.s.ReadPolicies(strings.NewReader(oneStageBenchmarkPolicySet)); err != nil {
		if logs := router.Stop(); len(logs) > 0 {
			b.Logf("router server logs:\n%s", logs)
		}

		b.Fatalf("can't read policies: %s", err)
	}

	if err := primary.s.ReadContent(strings.NewReader(benchmarkContent)); err != nil {
		if logs := router.Stop(); len(logs) > 0 {
			b.Logf("router server logs:\n%s", logs)
		}

		b.Fatalf("can't read content: %s", err)
	}

	if err := waitForPortClosed("127.0.0.1.5556"); err != nil {
		if logs := router.Stop(); len(logs) > 0 {
			b.Logf("router server logs:\n%s", logs)
		}

		b.Fatalf("port still in use: %s", err)
	}
	go func() {
		if err := primary.s.Serve(); err != nil {
			b.Fatalf("primary server failed: %s", err)
		}
	}()

	if err := waitForPortOpened("127.0.0.1:5556"); err != nil {
		if logs := router.Stop(); len(logs) > 0 {
			b.Logf("router server logs:\n%s", logs)
		}

		if logs := primary.Stop(); len(logs) > 0 {
			b.Logf("primary server logs:\n%s", logs)
		}

		b.Fatalf("can't connect to PDP server: %s", err)
	}

	c := pep.NewClient(opts...)
	if err := c.Connect("127.0.0.1:5555"); err != nil {
		if logs := router.Stop(); len(logs) > 0 {
			b.Logf("router server logs:\n%s", logs)
		}

		if logs := primary.Stop(); len(logs) > 0 {
			b.Logf("primary server logs:\n%s", logs)
		}

		b.Fatalf("can't connect to PDP server: %s", err)
	}

	return router, primary, c
}

func assertBenchMsg(r *pdp.Response, s string, args ...interface{}) error {
	if r.Effect != pdp.EffectDeny && r.Effect != pdp.EffectPermit && r.Effect != pdp.EffectNotApplicable {
		desc := fmt.Sprintf(s, args...)
		return fmt.Errorf("unexpected response effect for %s: %s", desc, pdp.EffectNameFromEnum(r.Effect))
	}

	if r.Status != nil {
		desc := fmt.Sprintf(s, args...)
		return fmt.Errorf("unexpected response status for %s: %s", desc, r.Status)
	}

	return nil
}
