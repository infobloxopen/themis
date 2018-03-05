package pep

import (
	"bytes"
	"fmt"
	"math/rand"
	"net"
	"strings"
	"testing"
	"time"

	log "github.com/sirupsen/logrus"

	pb "github.com/infobloxopen/themis/pdp-service"
	"github.com/infobloxopen/themis/pdpserver/server"
)

const (
	policySet = `# Policy set for benchmark 3-level nesting policy
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

	rawRequests []pb.Request
)

func init() {
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

func benchmarkValidate(b *testing.B) {
	pdp, _, c := startPDPServer(policySet, []string{"127.0.0.1:5555"}, b)
	defer func() {
		c.Close()
		if logs := pdp.Stop(); len(logs) > 0 {
			b.Logf("server logs:\n%s", logs)
		}
	}()

	b.Run("BenchmarkValidate", func(b *testing.B) {
		for n := 0; n < b.N; n++ {
			in := rawRequests[n%len(rawRequests)]

			out, err := c.Validate(&in)

			if err != nil {
				b.Fatalf("unexpected error: %#v", err)
			}

			if (out.Effect != pb.Response_DENY &&
				out.Effect != pb.Response_PERMIT &&
				out.Effect != pb.Response_NOTAPPLICABLE) ||
				out.Reason != "Ok" {
				b.Fatalf("unexpected response: %#v", out)
			}
		}
	})
}

func benchmarkClient(name string, addrs []string, b *testing.B) {
	if len(addrs) != 1 && len(addrs) != 2 {
		b.Fatalf("only 1 or 2 endpoints supported but got %d", len(addrs))
	}

	pdp, pdpAlt, c := startPDPServer(policySet, addrs, b)
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

		th := make(chan int, 100)
		for n := 0; n < b.N; n++ {
			th <- 0
			go func(i int) {
				defer func() { <-th }()

				in := rawRequests[n%len(rawRequests)]

				out, err := c.Validate(&in)

				if err != nil {
					b.Fatalf("unexpected error: %#v", err)
				}

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

func Benchmark(b *testing.B) {
	benchmarkValidate(b)
	/*benchmarkClient("BenchmarkClient",
		[]string{
			"127.0.0.1:5556",
		},
		b,
	)*/
	benchmarkClient("BenchmarkTwoClients",
		[]string{
			"127.0.0.1:5557",
			"127.0.0.1:5558",
		},
		b,
	)
}

func isOpen(address string) bool {
	server, err := net.Listen("tcp", address)
	if err != nil {
		return true
	}
	server.Close()
	return false
}

func waitForPortOpened(address string) error {
	for i := 0; i < 20; i++ {
		if isOpen(address) {
			return nil
		}
		<-time.After(500 * time.Millisecond)
	}

	return fmt.Errorf("port at %s hasn't been opened yet", address)
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

func startPDPServer(p string, addrs []string, b *testing.B) (*loggedServer, *loggedServer, Client) {
	var (
		primary   *loggedServer
		secondary *loggedServer
	)

	addr := addrs[0]

	primary = newServer(
		server.WithServiceAt(addr),
	)

	if err := primary.s.ReadPolicies(strings.NewReader(p)); err != nil {
		b.Fatalf("can't read policies: %s", err)
	}

	if err := primary.s.ReadContent(strings.NewReader(benchmarkContent)); err != nil {
		b.Fatalf("can't read content: %s", err)
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

	if len(addrs) > 1 {
		addr := addrs[1]
		secondary = newServer(
			server.WithServiceAt(addr),
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

	c := NewClient(addrs, 0, 0)
	if err := c.Connect(); err != nil {
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
