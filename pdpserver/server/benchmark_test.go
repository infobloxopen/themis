package server

import (
	"fmt"
	"math/rand"
	"strings"
	"testing"

	log "github.com/Sirupsen/logrus"

	"github.com/infobloxopen/themis/pdp"
	pb "github.com/infobloxopen/themis/pdp-service"
	"github.com/infobloxopen/themis/pdp/ast"
	"github.com/infobloxopen/themis/pdp/jcon"
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

type requestAttributeValue struct {
	k string
	v pdp.AttributeValue
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

	benchmarkContentStorage          *pdp.LocalContentStorage
	oneStageBenchmarkPolicyStorage   *pdp.PolicyStorage
	twoStageBenchmarkPolicyStorage   *pdp.PolicyStorage
	threeStageBenchmarkPolicyStorage *pdp.PolicyStorage

	benchmarkRequests []*pb.Request
)

func init() {
	log.SetLevel(log.ErrorLevel)

	c, err := jcon.Unmarshal(strings.NewReader(benchmarkContent), nil)
	if err != nil {
		panic(fmt.Errorf("expected no error while parsing content but got: %s", err))
	}

	benchmarkContentStorage = pdp.NewLocalContentStorage([]*pdp.LocalContent{c})
	parser := ast.NewYAMLParser()

	oneStageBenchmarkPolicyStorage, err = parser.Unmarshal(strings.NewReader(oneStageBenchmarkPolicySet), nil)
	if err != nil {
		panic(fmt.Errorf("expected no error while parsing policies but got: %s", err))
	}

	twoStageBenchmarkPolicyStorage, err = parser.Unmarshal(strings.NewReader(twoStageBenchmarkPolicySet), nil)
	if err != nil {
		panic(fmt.Errorf("expected no error while parsing policies but got: %s", err))
	}

	threeStageBenchmarkPolicyStorage, err = parser.Unmarshal(strings.NewReader(threeStageBenchmarkPolicySet), nil)
	if err != nil {
		panic(fmt.Errorf("expected no error while parsing policies but got: %s", err))
	}

	benchmarkRequests = make([]*pb.Request, 2000000)
	for i := range benchmarkRequests {
		benchmarkRequests[i] = &pb.Request{
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

func benchmarkPolicySet(p *pdp.PolicyStorage, b *testing.B) {
	s := NewServer()
	s.p = p
	s.c = benchmarkContentStorage

	for n := 0; n < b.N; n++ {
		r, err := s.Validate(nil, benchmarkRequests[n%len(benchmarkRequests)])
		if err != nil {
			b.Fatalf("Expected no error while evaluating policies at %d iteration but got: %s", n+1, err)
		}

		if r.Effect >= pb.Response_INDETERMINATE {
			b.Fatalf("Expected specific result of policy evaluation at %d iteration but got %s (%s)",
				n+1, pb.Response_Effect_name[int32(r.Effect)], r.Reason)
		}
	}
}

func BenchmarkOneStagePolicySet(b *testing.B) {
	benchmarkPolicySet(oneStageBenchmarkPolicyStorage, b)
}

func BenchmarkTwoStagePolicySet(b *testing.B) {
	benchmarkPolicySet(twoStageBenchmarkPolicyStorage, b)
}

func BenchmarkThreeStagePolicySet(b *testing.B) {
	benchmarkPolicySet(threeStageBenchmarkPolicyStorage, b)
}
