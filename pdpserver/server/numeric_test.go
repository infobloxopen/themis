package server

import (
	"fmt"
	"strings"
	"testing"

	log "github.com/Sirupsen/logrus"

	"github.com/infobloxopen/themis/pdp"
    "github.com/infobloxopen/themis/pdp/ast"
	pb "github.com/infobloxopen/themis/pdp-service"
)

type testCase struct {
	a string
	b string
	c string
	expected pb.Response_Effect
}

type testSuite struct {
	policy string
	testSet []testCase
}

const (
	integerEqualPolicySet = `# Policy set for Integer Equal Comparison
attributes:
  a: integer
  b: integer

policies:
  alg: FirstApplicableEffect
  rules:
  - id: "Test Integer Equal"
    condition: # a == b
      equal:
      - attr: a
      - attr: b
    effect: Permit
`
	
	integerGreaterPolicySet = `# Policy set for Integer Greater Comparison
attributes:
  a: integer
  b: integer

policies:
  alg: FirstApplicableEffect
  rules:
  - id: "Test Integer Greater"
    condition: # a > b
      greater:
      - attr: a
      - attr: b
    effect: Permit
`

	integerAddPolicySet = `# Policy set for Integer Addition
attributes:
  a: integer
  b: integer
  c: integer

policies:
  alg: FirstApplicableEffect
  rules:
  - id: "Test Integer Addition"
    condition: # a + b == c
      equal:
      - add: # a + b
        - attr: a
        - attr: b
      - attr: c
    effect: Permit
`

	integerSubtractPolicySet = `# Policy set for Integer Subtraction
attributes:
  a: integer
  b: integer
  c: integer

policies:
  alg: FirstApplicableEffect
  rules:
  - id: "Test Integer Subtraction"
    condition: # a - b == c
      equal:
      - subtract:
        - attr: a
        - attr: b
      - attr: c
    effect: Permit
`
)

type validateRequestAttributeValue struct {
	k string
	v pdp.AttributeValue
}

func init() {
	log.SetLevel(log.ErrorLevel)
}

func loadPolicy(policyYAML string) *pdp.PolicyStorage {
	parser := ast.NewYAMLParser()
	policyStorage, err := parser.Unmarshal(strings.NewReader(policyYAML), nil)
	if err != nil {
		panic(fmt.Errorf("expected no error while parsing policies but got: %s", err))
	}

	return policyStorage
}

func validateTestSuite(ts testSuite, t *testing.T) {
	s := NewServer()
	s.p = loadPolicy(ts.policy)

	for _, tc := range ts.testSet {
		t.Run(fmt.Sprintf("a=%s,b=%s,c=%s", tc.a, tc.b, tc.c), func(t *testing.T) {
			req := &pb.Request{
				Attributes: []*pb.Attribute{
					{
						Id: "a",
						Type: "integer",
						Value: tc.a,
					},
					{
						Id: "b",
						Type: "integer",
						Value: tc.b,
					},		
					{
						Id: "c",
						Type: "integer",
						Value: tc.c,
					},		
				},
			}

			r, err := s.Validate(nil, req)
			if err != nil {
				t.Fatalf("Expected no error while evaluating policy but got: %s", err)
			}

			if r.Effect != tc.expected {
				t.Fatalf("Expected result of policy evaluation %s, but got %s (%s)",
					tc.expected, pb.Response_Effect_name[int32(r.Effect)], r.Reason)
			}
		})
	}
}

func TestIntegerEqual(t *testing.T) {
	ts := testSuite {
		policy: integerEqualPolicySet,
		testSet: []testCase{
			{
				a: "0",
				b: "0",
 				c: "1",
				expected: pb.Response_PERMIT,
			},
			{
				a: "0",
				b: "0",
 				c: "0",
				expected: pb.Response_PERMIT,
			},
			{
				a: "-1",
				b: "-1",
 				c: "-2",
				expected: pb.Response_PERMIT,
			},
			{
				a: "1",
				b: "0",
 				c: "0",
				expected: pb.Response_NOTAPPLICABLE,
			},
			{
				a: "0",
				b: "1",
 				c: "0",
				expected: pb.Response_NOTAPPLICABLE,
			},
			{
				a: "1",
				b: "-2",
 				c: "0",
				expected: pb.Response_NOTAPPLICABLE,
			},
		},
	}

	validateTestSuite(ts, t)
}

func TestIntegerGreater(t *testing.T) {
	ts := testSuite {
		policy: integerGreaterPolicySet,
		testSet: []testCase{
			{
				a: "1",
				b: "0",
				c: "0",
				expected: pb.Response_PERMIT,
			},
			{
				a: "0",
				b: "-1",
				c: "0",
				expected: pb.Response_PERMIT,
			},
			{
				a: "1",
				b: "-1",
				c: "0",
				expected: pb.Response_PERMIT,
			},
			{
				a: "0",
				b: "0",
				c: "0",
				expected: pb.Response_NOTAPPLICABLE,
			},
			{
				a: "0",
				b: "1",
				c: "0",
				expected: pb.Response_NOTAPPLICABLE,
			},
			{
				a: "-1",
				b: "1",
				c: "0",
				expected: pb.Response_NOTAPPLICABLE,
			},
		},
	}

	validateTestSuite(ts, t)
}

func TestIntegerAdd(t *testing.T) {
	ts := testSuite {
		policy: integerAddPolicySet,
		testSet: []testCase{
			{
				a: "1",
				b: "1",
 				c: "2",
				expected: pb.Response_PERMIT,
			},
			{
				a: "0",
				b: "0",
 				c: "0",
				expected: pb.Response_PERMIT,
			},
			{
				a: "-1",
				b: "-1",
 				c: "-2",
				expected: pb.Response_PERMIT,
			},
			{
				a: "1",
				b: "0",
 				c: "2",
				expected: pb.Response_NOTAPPLICABLE,
			},
			{
				a: "0",
				b: "1",
 				c: "2",
				expected: pb.Response_NOTAPPLICABLE,
			},
			{
				a: "1",
				b: "-1",
 				c: "2",
				expected: pb.Response_NOTAPPLICABLE,
			},
		},
	}

	validateTestSuite(ts, t)
}
	
func TestIntegerSubtract(t *testing.T) {
	ts := testSuite {
		policy: integerSubtractPolicySet,
		testSet: []testCase{
			{
				a: "1",
				b: "1",
 				c: "0",
				expected: pb.Response_PERMIT,
			},
			{
				a: "0",
				b: "0",
 				c: "0",
				expected: pb.Response_PERMIT,
			},
			{
				a: "-1",
				b: "-1",
 				c: "0",
				expected: pb.Response_PERMIT,
			},
			{
				a: "1",
				b: "0",
 				c: "0",
				expected: pb.Response_NOTAPPLICABLE,
			},
			{
				a: "0",
				b: "1",
 				c: "0",
				expected: pb.Response_NOTAPPLICABLE,
			},
			{
				a: "1",
				b: "-1",
 				c: "0",
				expected: pb.Response_NOTAPPLICABLE,
			},
		},
	}

	validateTestSuite(ts, t)
}
	
