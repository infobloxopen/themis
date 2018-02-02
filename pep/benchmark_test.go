package pep

import (
	"bytes"
	"fmt"
	"math/rand"
	"net"
	"strings"
	"testing"
	"time"

	log "github.com/Sirupsen/logrus"

	pb "github.com/infobloxopen/themis/pdp-service"
	"github.com/infobloxopen/themis/pdpserver/server"
)

const (
	oneStageBenchmarkPolicySet = `# Policy set for benchmark
attributes:
  k3: domain
  x: string

policies:
  alg:
    id: mapper
    map:
      selector:
        uri: "local:content/second"
        path:
        - attr: k3
        type: list of strings
    default: DefaultRule
    alg: FirstApplicableEffect
  rules:
  - id: DefaultRule
    effect: Deny
    obligations:
    - x:
       val:
         type: string
         content: DefaultRule
  - id: First
    effect: Permit
    obligations:
    - x:
       val:
         type: string
         content: First
  - id: Second
    effect: Permit
    obligations:
    - x:
       val:
         type: string
         content: Second
  - id: Third
    effect: Permit
    obligations:
    - x:
       val:
         type: string
         content: Third
  - id: Fourth
    effect: Permit
    obligations:
    - x:
       val:
         type: string
         content: Fourth
  - id: Fifth
    effect: Permit
    obligations:
    - x:
       val:
         type: string
         content: Fifth
`

	twoStageBenchmarkPolicySet = `# Policy set for benchmark 2-level nesting policy
attributes:
  k2: string
  k3: domain
  x: string

policies:
  alg:
    id: mapper
    map:
      selector:
        uri: "local:content/first"
        path:
        - attr: k2
        type: string
    default: DefaultPolicy

  policies:
  - id: DefaultPolicy
    alg: FirstApplicableEffect
    rules:
    - effect: Deny
      obligations:
      - x:
         val:
           type: string
           content: DefaultPolicy

  - id: P1
    alg:
      id: mapper
      map:
        selector:
          uri: "local:content/second"
          path:
          - attr: k3
          type: list of strings
      default: DefaultRule
      alg: FirstApplicableEffect
    rules:
    - id: DefaultRule
      effect: Deny
      obligations:
      - x:
         val:
           type: string
           content: P1.DefaultRule
    - id: First
      effect: Permit
      obligations:
      - x:
         val:
           type: string
           content: P1.First
    - id: Second
      effect: Permit
      obligations:
      - x:
         val:
           type: string
           content: P1.Second

  - id: P2
    alg:
      id: mapper
      map:
        selector:
          uri: "local:content/second"
          path:
          - attr: k3
          type: list of strings
      default: DefaultRule
      alg: FirstApplicableEffect
    rules:
    - id: DefaultRule
      effect: Deny
      obligations:
      - x:
         val:
           type: string
           content: P2.DefaultRule
    - id: Second
      effect: Permit
      obligations:
      - x:
         val:
           type: string
           content: P2.Second
    - id: Third
      effect: Permit
      obligations:
      - x:
         val:
           type: string
           content: P2.Third

  - id: P3
    alg:
      id: mapper
      map:
        selector:
          uri: "local:content/second"
          path:
          - attr: k3
          type: list of strings
      default: DefaultRule
      alg: FirstApplicableEffect
    rules:
    - id: DefaultRule
      effect: Deny
      obligations:
      - x:
         val:
           type: string
           content: P3.DefaultRule
    - id: Third
      effect: Permit
      obligations:
      - x:
         val:
           type: string
           content: P3.Third
    - id: Fourth
      effect: Permit
      obligations:
      - x:
         val:
           type: string
           content: P3.Fourth

  - id: P4
    alg:
      id: mapper
      map:
        selector:
          uri: "local:content/second"
          path:
          - attr: k3
          type: list of strings
      default: DefaultRule
      alg: FirstApplicableEffect
    rules:
    - id: DefaultRule
      effect: Deny
      obligations:
      - x:
         val:
           type: string
           content: P4.DefaultRule
    - id: Fourth
      effect: Permit
      obligations:
      - x:
         val:
           type: string
           content: P4.Fourth
    - id: Fifth
      effect: Permit
      obligations:
      - x:
         val:
           type: string
           content: P4.Fifth

  - id: P5
    alg:
      id: mapper
      map:
        selector:
          uri: "local:content/second"
          path:
          - attr: k3
          type: list of strings
      default: DefaultRule
      alg: FirstApplicableEffect
    rules:
    - id: DefaultRule
      effect: Deny
      obligations:
      - x:
         val:
           type: string
           content: P5.DefaultRule
    - id: Fifth
      effect: Permit
      obligations:
      - x:
         val:
           type: string
           content: P5.Fifth
    - id: First
      effect: Permit
      obligations:
      - x:
         val:
           type: string
           content: P5.First
`

	threeStageBenchmarkPolicySet = `# Policy set for benchmark 3-level nesting policy
attributes:
  k1: string
  k2: string
  k3: domain
  x: string

policies:
  alg: FirstApplicableEffect
  policies:
  - target:
    - equal:
      - attr: k1
      - val:
          type: string
          content: "Left"
    alg:
      id: mapper
      map:
        selector:
          uri: "local:content/first"
          path:
          - attr: k2
          type: string
      default: DefaultPolicy

    policies:
    - id: DefaultPolicy
      alg: FirstApplicableEffect
      rules:
      - effect: Deny
        obligations:
        - x:
           val:
             type: string
             content: DefaultPolicy

    - id: P1
      alg:
        id: mapper
        map:
          selector:
            uri: "local:content/second"
            path:
            - attr: k3
            type: list of strings
        default: DefaultRule
        alg: FirstApplicableEffect
      rules:
      - id: DefaultRule
        effect: Deny
        obligations:
        - x:
           val:
             type: string
             content: P1.DefaultRule
      - id: First
        effect: Permit
        obligations:
        - x:
           val:
             type: string
             content: P1.First
      - id: Second
        effect: Permit
        obligations:
        - x:
           val:
             type: string
             content: P1.Second

    - id: P2
      alg:
        id: mapper
        map:
          selector:
            uri: "local:content/second"
            path:
            - attr: k3
            type: list of strings
        default: DefaultRule
        alg: FirstApplicableEffect
      rules:
      - id: DefaultRule
        effect: Deny
        obligations:
        - x:
           val:
             type: string
             content: P2.DefaultRule
      - id: Second
        effect: Permit
        obligations:
        - x:
           val:
             type: string
             content: P2.Second
      - id: Third
        effect: Permit
        obligations:
        - x:
           val:
             type: string
             content: P2.Third

    - id: P3
      alg:
        id: mapper
        map:
          selector:
            uri: "local:content/second"
            path:
            - attr: k3
            type: list of strings
        default: DefaultRule
        alg: FirstApplicableEffect
      rules:
      - id: DefaultRule
        effect: Deny
        obligations:
        - x:
           val:
             type: string
             content: P3.DefaultRule
      - id: Third
        effect: Permit
        obligations:
        - x:
           val:
             type: string
             content: P3.Third
      - id: Fourth
        effect: Permit
        obligations:
        - x:
           val:
             type: string
             content: P3.Fourth

    - id: P4
      alg:
        id: mapper
        map:
          selector:
            uri: "local:content/second"
            path:
            - attr: k3
            type: list of strings
        default: DefaultRule
        alg: FirstApplicableEffect
      rules:
      - id: DefaultRule
        effect: Deny
        obligations:
        - x:
           val:
             type: string
             content: P4.DefaultRule
      - id: Fourth
        effect: Permit
        obligations:
        - x:
           val:
             type: string
             content: P4.Fourth
      - id: Fifth
        effect: Permit
        obligations:
        - x:
           val:
             type: string
             content: P4.Fifth

    - id: P5
      alg:
        id: mapper
        map:
          selector:
            uri: "local:content/second"
            path:
            - attr: k3
            type: list of strings
        default: DefaultRule
        alg: FirstApplicableEffect
      rules:
      - id: DefaultRule
        effect: Deny
        obligations:
        - x:
           val:
             type: string
             content: P5.DefaultRule
      - id: Fifth
        effect: Permit
        obligations:
        - x:
           val:
             type: string
             content: P5.Fifth
      - id: First
        effect: Permit
        obligations:
        - x:
           val:
             type: string
             content: P5.First

  - target:
    - equal:
      - attr: k1
      - val:
          type: string
          content: "Right"
    alg:
      id: mapper
      map:
        selector:
          uri: "local:content/first"
          path:
          - attr: k2
          type: string
      default: DefaultPolicy

    policies:
    - id: DefaultPolicy
      alg: FirstApplicableEffect
      rules:
      - effect: Deny
        obligations:
        - x:
           val:
             type: string
             content: DefaultPolicy

    - id: P1
      alg:
        id: mapper
        map:
          selector:
            uri: "local:content/second"
            path:
            - attr: k3
            type: list of strings
        default: DefaultRule
        alg: FirstApplicableEffect
      rules:
      - id: DefaultRule
        effect: Deny
        obligations:
        - x:
           val:
             type: string
             content: P1.DefaultRule
      - id: First
        effect: Permit
        obligations:
        - x:
           val:
             type: string
             content: P1.First
      - id: Second
        effect: Permit
        obligations:
        - x:
           val:
             type: string
             content: P1.Second

    - id: P2
      alg:
        id: mapper
        map:
          selector:
            uri: "local:content/second"
            path:
            - attr: k3
            type: list of strings
        default: DefaultRule
        alg: FirstApplicableEffect
      rules:
      - id: DefaultRule
        effect: Deny
        obligations:
        - x:
           val:
             type: string
             content: P2.DefaultRule
      - id: Second
        effect: Permit
        obligations:
        - x:
           val:
             type: string
             content: P2.Second
      - id: Third
        effect: Permit
        obligations:
        - x:
           val:
             type: string
             content: P2.Third

    - id: P3
      alg:
        id: mapper
        map:
          selector:
            uri: "local:content/second"
            path:
            - attr: k3
            type: list of strings
        default: DefaultRule
        alg: FirstApplicableEffect
      rules:
      - id: DefaultRule
        effect: Deny
        obligations:
        - x:
           val:
             type: string
             content: P3.DefaultRule
      - id: Third
        effect: Permit
        obligations:
        - x:
           val:
             type: string
             content: P3.Third
      - id: Fourth
        effect: Permit
        obligations:
        - x:
           val:
             type: string
             content: P3.Fourth

    - id: P4
      alg:
        id: mapper
        map:
          selector:
            uri: "local:content/second"
            path:
            - attr: k3
            type: list of strings
        default: DefaultRule
        alg: FirstApplicableEffect
      rules:
      - id: DefaultRule
        effect: Deny
        obligations:
        - x:
           val:
             type: string
             content: P4.DefaultRule
      - id: Fourth
        effect: Permit
        obligations:
        - x:
           val:
             type: string
             content: P4.Fourth
      - id: Fifth
        effect: Permit
        obligations:
        - x:
           val:
             type: string
             content: P4.Fifth

    - id: P5
      alg:
        id: mapper
        map:
          selector:
            uri: "local:content/second"
            path:
            - attr: k3
            type: list of strings
        default: DefaultRule
        alg: FirstApplicableEffect
      rules:
      - id: DefaultRule
        effect: Deny
        obligations:
        - x:
           val:
             type: string
             content: P5.DefaultRule
      - id: Fifth
        effect: Permit
        obligations:
        - x:
           val:
             type: string
             content: P5.Fifth
      - id: First
        effect: Permit
        obligations:
        - x:
           val:
             type: string
             content: P5.First

  - alg: FirstApplicableEffect
    rules:
    - effect: Deny
      obligations:
      - x:
         val:
           type: string
           content: Root Deny
`

	benchmarkContent = `{
    "id": "content",
    "items": {
        "first": {
            "keys": ["string"],
            "type": "string",
            "data": {
                "First": "P1",
                "Second": "P2",
                "Third": "P3",
                "Fourth": "P4",
                "Fifth": "P5",
                "Sixth": "P6",
                "Seventh": "P7"
            }
        },
        "second": {
            "keys": ["domain"],
            "type": "list of strings",
            "data": {
                "first.example.com": ["First", "Third"],
                "second.example.com": ["Second", "Fourth"],
                "third.example.com": ["Third", "Fifth"],
                "first.test.com": ["Fourth", "Sixth"],
                "second.test.com": ["Fifth", "Seventh"],
                "third.test.com": ["Sixth", "First"],
                "first.example.com": ["Seventh", "Second"],
                "second.example.com": ["Firth", "Fourth"],
                "third.example.com": ["Second", "Fifth"],
                "first.test.com": ["Third", "Sixth"],
                "second.test.com": ["Fourth", "Seventh"],
                "third.test.com": ["Fifth", "First"]
            }
        }
    }
}`
)

type decisionRequest struct {
	Direction string `pdp:"k1"`
	Policy    string `pdp:"k2"`
	Domain    string `pdp:"k3,domain"`
}

type decisionResponse struct {
	Effect string `pdp:"Effect"`
	Reason string `pdp:"Reason"`
	X      string `pdp:"x"`
}

var (
	directionOpts = []string{
		"Left",
		"Right",
	}

	policySetOpts = []string{
		"First",
		"Second",
		"Third",
		"Fourth",
		"Fifth",
		"Sixth",
		"Seventh",
	}

	domainOpts = []string{
		"first.example.com",
		"second.example.com",
		"third.example.com",
		"first.test.com",
		"second.test.com",
		"third.test.com",
		"first.example.com",
		"second.example.com",
		"third.example.com",
		"first.test.com",
		"second.test.com",
		"third.test.com",
	}

	decisionRequests []decisionRequest
	rawRequests      []pb.Request
)

func init() {
	decisionRequests = make([]decisionRequest, 2000000)
	for i := range decisionRequests {
		decisionRequests[i] = decisionRequest{
			Direction: directionOpts[rand.Intn(len(directionOpts))],
			Policy:    policySetOpts[rand.Intn(len(policySetOpts))],
			Domain:    domainOpts[rand.Intn(len(domainOpts))],
		}
	}

	rawRequests = make([]pb.Request, 2000000)
	for i := range rawRequests {
		rawRequests[i] = pb.Request{
			Attributes: []*pb.Attribute{
				{
					Id:    "k1",
					Type:  "string",
					Value: directionOpts[rand.Intn(len(directionOpts))],
				},
				{
					Id:    "k2",
					Type:  "string",
					Value: policySetOpts[rand.Intn(len(policySetOpts))],
				},
				{
					Id:    "k3",
					Type:  "domain",
					Value: domainOpts[rand.Intn(len(domainOpts))],
				},
			},
		}
	}

}

func benchmarkPolicySet(name, p string, b *testing.B) {
	pdp, _, c := startPDPServer(p, nil, b)
	defer func() {
		c.Close()
		if logs := pdp.Stop(); len(logs) > 0 {
			b.Logf("server logs:\n%s", logs)
		}
	}()

	b.Run(name, func(b *testing.B) {
		for n := 0; n < b.N; n++ {
			in := decisionRequests[n%len(decisionRequests)]

			var out decisionResponse
			c.Validate(in, &out)

			if (out.Effect != "DENY" && out.Effect != "PERMIT" && out.Effect != "NOTAPPLICABLE") ||
				out.Reason != "Ok" {
				b.Fatalf("unexpected response: %#v", out)
			}
		}
	})
}

func BenchmarkOneStagePolicySet(b *testing.B) {
	benchmarkPolicySet("BenchmarkOneStagePolicySet", oneStageBenchmarkPolicySet, b)
}

func BenchmarkTwoStagePolicySet(b *testing.B) {
	benchmarkPolicySet("BenchmarkTwoStagePolicySet", twoStageBenchmarkPolicySet, b)
}

func BenchmarkThreeStagePolicySet(b *testing.B) {
	benchmarkPolicySet("BenchmarkThreeStagePolicySet", threeStageBenchmarkPolicySet, b)
}

func BenchmarkThreeStagePolicySetRaw(b *testing.B) {
	pdp, _, c := startPDPServer(threeStageBenchmarkPolicySet, nil, b)
	defer func() {
		c.Close()
		if logs := pdp.Stop(); len(logs) > 0 {
			b.Logf("server logs:\n%s", logs)
		}
	}()

	b.Run("BenchmarkThreeStagePolicySetRaw", func(b *testing.B) {
		var out pb.Response
		for n := 0; n < b.N; n++ {
			in := rawRequests[n%len(rawRequests)]

			c.Validate(in, &out)

			if (out.Effect != pb.Response_DENY &&
				out.Effect != pb.Response_PERMIT &&
				out.Effect != pb.Response_NOTAPPLICABLE) ||
				out.Reason != "Ok" {
				b.Fatalf("unexpected response: %#v", out)
			}
		}
	})
}

func benchmarkStreamingClient(name string, ports []uint16, b *testing.B, opts ...Option) {
	if len(ports) != 0 && len(ports) != 2 {
		b.Fatalf("only 0 for single PDP and 2 for 2 PDP ports supported but got %d", len(ports))
	}

	streams := 96

	opts = append(opts,
		WithStreams(streams),
	)
	pdp, pdpAlt, c := startPDPServer(threeStageBenchmarkPolicySet, ports, b, opts...)
	defer func() {
		c.Close()
		if logs := pdp.Stop(); len(logs) > 0 {
			b.Logf("primary server logs:\n%s", logs)
		}
		if pdpAlt != nil {
			if logs := pdpAlt.Stop(); len(logs) > 0 {
				b.Logf("secondary server logs:\n%s", logs)
			}
		}
	}()

	b.Run(name, func(b *testing.B) {
		errs := make([]error, b.N)

		th := make(chan int, streams)
		for n := 0; n < b.N; n++ {
			th <- 0
			go func(i int) {
				defer func() { <-th }()

				var out pb.Response
				c.Validate(rawRequests[i%len(rawRequests)], &out)

				if (out.Effect != pb.Response_DENY &&
					out.Effect != pb.Response_PERMIT &&
					out.Effect != pb.Response_NOTAPPLICABLE) ||
					out.Reason != "Ok" {
					errs[i] = fmt.Errorf("unexpected response: %#v", out)
				}
			}(n)
		}

		for i, err := range errs {
			if err != nil {
				b.Fatalf("request %d failed: %s", i, err)
			}
		}
	})
}

func BenchmarkStreamingClient(b *testing.B) {
	benchmarkStreamingClient("BenchmarkStreamingClient", nil, b)
}

func BenchmarkRoundRobinStreamingClient(b *testing.B) {
	benchmarkStreamingClient("BenchmarkRoundRobinStreamingClient",
		[]uint16{
			5555,
			5556,
		},
		b,
		WithRoundRobinBalancer("127.0.0.1:5555", "127.0.0.1:5556"),
	)
}

func BenchmarkHotSpotStreamingClient(b *testing.B) {
	benchmarkStreamingClient("BenchmarkHotSpotStreamingClient",
		[]uint16{
			5555,
			5556,
		},
		b,
		WithHotSpotBalancer("127.0.0.1:5555", "127.0.0.1:5556"),
	)
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
	s *server.Server
	b *bytes.Buffer
}

func newServer(opts ...server.Option) *loggedServer {
	s := &loggedServer{
		b: new(bytes.Buffer),
	}

	logger := log.New()
	logger.Out = s.b
	logger.Level = log.ErrorLevel
	opts = append(opts,
		server.WithLogger(logger),
	)

	s.s = server.NewServer(opts...)
	return s
}

func (s *loggedServer) Stop() string {
	s.s.Stop()
	return s.b.String()
}

func startPDPServer(p string, ports []uint16, b *testing.B, opts ...Option) (*loggedServer, *loggedServer, Client) {
	var (
		primary   *loggedServer
		secondary *loggedServer
	)

	service := ":5555"
	if len(ports) > 0 {
		service = fmt.Sprintf(":%d", ports[0])
	}
	addr := "127.0.0.1" + service

	primary = newServer(
		server.WithServiceAt(service),
	)

	if err := primary.s.ReadPolicies(strings.NewReader(p)); err != nil {
		b.Fatalf("can't read policies: %s", err)
	}

	if err := primary.s.ReadContent(strings.NewReader(benchmarkContent)); err != nil {
		b.Fatalf("can't read content: %s", err)
	}

	if err := waitForPortClosed(addr); err != nil {
		b.Fatalf("port still in use: %s", err)
	}
	go func() {
		if err := primary.s.Serve(); err != nil {
			b.Fatalf("primary server failed: %s", err)
		}
	}()

	if err := waitForPortOpened(addr); err != nil {
		if logs := primary.Stop(); len(logs) > 0 {
			b.Logf("primary server logs:\n%s", logs)
		}

		b.Fatalf("can't connect to PDP server: %s", err)
	}

	if len(ports) > 1 {
		service := fmt.Sprintf(":%d", ports[1])
		secondary = newServer(
			server.WithServiceAt(service),
		)
		addr := "127.0.0.1" + service

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

		if err := waitForPortClosed(addr); err != nil {
			b.Fatalf("port still in use: %s", err)
		}
		go func() {
			if err := secondary.s.Serve(); err != nil {
				b.Fatalf("secondary server failed: %s", err)
			}
		}()

		if err := waitForPortOpened(addr); err != nil {
			if logs := secondary.Stop(); len(logs) > 0 {
				b.Logf("secondary server logs:\n%s", logs)
			}
			if logs := primary.Stop(); len(logs) > 0 {
				b.Logf("primary server logs:\n%s", logs)
			}

			b.Fatalf("can't connect to PDP server: %s", err)
		}
	}

	c := NewClient(opts...)
	if err := c.Connect(addr); err != nil {
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
