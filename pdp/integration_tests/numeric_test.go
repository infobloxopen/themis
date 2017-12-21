package integrationTests

import (
	"fmt"
	"strings"
	"testing"

	log "github.com/Sirupsen/logrus"

	"github.com/infobloxopen/themis/pdp"
	pb "github.com/infobloxopen/themis/pdp-service"
	"github.com/infobloxopen/themis/pdp/ast"
)

type testCase struct {
	a        string
	b        string
	c        string
	expected int
}

type testSuite struct {
	policy  string
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

	integerMultiplyPolicySet = `# Policy set for Integer Multiplication
attributes:
  a: integer
  b: integer
  c: integer

policies:
  alg: FirstApplicableEffect
  rules:
  - id: "Test Integer Multiplication"
    condition: # a * b == c
      equal:
      - multiply: # a * b
        - attr: a
        - attr: b
      - attr: c
    effect: Permit
`

	integerDividePolicySet = `# Policy set for Integer Division
attributes:
  a: integer
  b: integer
  c: integer

policies:
  alg: FirstApplicableEffect
  rules:
  - id: "Test Integer Division"
    condition: # a / b == c
      equal:
      - divide: # a / b
        - attr: a
        - attr: b
      - attr: c
    effect: Permit
`
)

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

func createRequest(tc testCase) *pb.Request {
	return &pb.Request{
		Attributes: []*pb.Attribute{
			{
				Id:    "a",
				Type:  "integer",
				Value: tc.a,
			},
			{
				Id:    "b",
				Type:  "integer",
				Value: tc.b,
			},
			{
				Id:    "c",
				Type:  "integer",
				Value: tc.c,
			},
		},
	}
}

func createContext(req *pb.Request) (*pdp.Context, error) {
	ctx, err := pdp.NewContext(nil, len(req.Attributes), func(i int) (string, pdp.AttributeValue, error) {
		a := req.Attributes[i]

		t, ok := pdp.TypeIDs[strings.ToLower(a.Type)]
		if !ok {
			return "", pdp.AttributeValue{}, fmt.Errorf("unknown Attribute Type: %s", a.Type)
		}

		v, err := pdp.MakeValueFromString(t, a.Value)
		if err != nil {
			return "", pdp.AttributeValue{}, fmt.Errorf("error making value from string: %s", err)
		}

		return a.Id, v, nil
	})
	if err != nil {
		return nil, fmt.Errorf("cannot create context: %s", err)
	}

	return ctx, nil
}

func validateTestSuite(ts testSuite, t *testing.T) {
	p := loadPolicy(ts.policy)

	for _, tc := range ts.testSet {
		t.Run(fmt.Sprintf("a=%s,b=%s,c=%s", tc.a, tc.b, tc.c), func(t *testing.T) {
			req := createRequest(tc)
			ctx, err := createContext(req)
			if err != nil {
				t.Fatalf("Expected no error while creating context but got: %s", err)
			}

			r := p.Root().Calculate(ctx)
			effect, _, err := r.Status()
			if err != nil {
				t.Fatalf("Expected no error while evaluating policy but got: %s", err)
			}

			if effect != tc.expected {
				t.Fatalf("Expected result of policy evaluation %d, but got %d", tc.expected, effect)
			}
		})
	}
}

func TestIntegerEqual(t *testing.T) {
	ts := testSuite{
		policy: integerEqualPolicySet,
		testSet: []testCase{
			{
				a:        "1",
				b:        "1",
				c:        "0",
				expected: pdp.EffectPermit,
			},
			{
				a:        "0",
				b:        "0",
				c:        "0",
				expected: pdp.EffectPermit,
			},
			{
				a:        "-1",
				b:        "-1",
				c:        "0",
				expected: pdp.EffectPermit,
			},
			{
				a:        "1",
				b:        "0",
				c:        "0",
				expected: pdp.EffectNotApplicable,
			},
			{
				a:        "0",
				b:        "1",
				c:        "0",
				expected: pdp.EffectNotApplicable,
			},
			{
				a:        "1",
				b:        "-2",
				c:        "0",
				expected: pdp.EffectNotApplicable,
			},
		},
	}

	validateTestSuite(ts, t)
}

func TestIntegerGreater(t *testing.T) {
	ts := testSuite{
		policy: integerGreaterPolicySet,
		testSet: []testCase{
			{
				a:        "1",
				b:        "0",
				c:        "0",
				expected: pdp.EffectPermit,
			},
			{
				a:        "0",
				b:        "-1",
				c:        "0",
				expected: pdp.EffectPermit,
			},
			{
				a:        "1",
				b:        "-1",
				c:        "0",
				expected: pdp.EffectPermit,
			},
			{
				a:        "0",
				b:        "0",
				c:        "0",
				expected: pdp.EffectNotApplicable,
			},
			{
				a:        "0",
				b:        "1",
				c:        "0",
				expected: pdp.EffectNotApplicable,
			},
			{
				a:        "-1",
				b:        "1",
				c:        "0",
				expected: pdp.EffectNotApplicable,
			},
		},
	}

	validateTestSuite(ts, t)
}

func TestIntegerAdd(t *testing.T) {
	ts := testSuite{
		policy: integerAddPolicySet,
		testSet: []testCase{
			{
				a:        "1",
				b:        "1",
				c:        "2",
				expected: pdp.EffectPermit,
			},
			{
				a:        "0",
				b:        "0",
				c:        "0",
				expected: pdp.EffectPermit,
			},
			{
				a:        "-1",
				b:        "-1",
				c:        "-2",
				expected: pdp.EffectPermit,
			},
			{
				a:        "1",
				b:        "0",
				c:        "2",
				expected: pdp.EffectNotApplicable,
			},
			{
				a:        "0",
				b:        "1",
				c:        "2",
				expected: pdp.EffectNotApplicable,
			},
			{
				a:        "1",
				b:        "-1",
				c:        "2",
				expected: pdp.EffectNotApplicable,
			},
		},
	}

	validateTestSuite(ts, t)
}

func TestIntegerSubtract(t *testing.T) {
	ts := testSuite{
		policy: integerSubtractPolicySet,
		testSet: []testCase{
			{
				a:        "1",
				b:        "1",
				c:        "0",
				expected: pdp.EffectPermit,
			},
			{
				a:        "0",
				b:        "0",
				c:        "0",
				expected: pdp.EffectPermit,
			},
			{
				a:        "-1",
				b:        "-1",
				c:        "0",
				expected: pdp.EffectPermit,
			},
			{
				a:        "1",
				b:        "0",
				c:        "0",
				expected: pdp.EffectNotApplicable,
			},
			{
				a:        "0",
				b:        "1",
				c:        "0",
				expected: pdp.EffectNotApplicable,
			},
			{
				a:        "1",
				b:        "-1",
				c:        "0",
				expected: pdp.EffectNotApplicable,
			},
		},
	}

	validateTestSuite(ts, t)
}

func TestIntegerMultiply(t *testing.T) {
	ts := testSuite{
		policy: integerMultiplyPolicySet,
		testSet: []testCase{
			{
				a:        "1",
				b:        "1",
				c:        "1",
				expected: pdp.EffectPermit,
			},
			{
				a:        "1",
				b:        "0",
				c:        "0",
				expected: pdp.EffectPermit,
			},
			{
				a:        "0",
				b:        "1",
				c:        "0",
				expected: pdp.EffectPermit,
			},
			{
				a:        "-1",
				b:        "-1",
				c:        "1",
				expected: pdp.EffectPermit,
			},
			{
				a:        "-1",
				b:        "1",
				c:        "-1",
				expected: pdp.EffectPermit,
			},
			{
				a:        "1",
				b:        "0",
				c:        "1",
				expected: pdp.EffectNotApplicable,
			},
			{
				a:        "-1",
				b:        "1",
				c:        "1",
				expected: pdp.EffectNotApplicable,
			},
			{
				a:        "0",
				b:        "1",
				c:        "1",
				expected: pdp.EffectNotApplicable,
			},
		},
	}

	validateTestSuite(ts, t)
}

func TestIntegerDivide(t *testing.T) {
	ts := testSuite{
		policy: integerDividePolicySet,
		testSet: []testCase{
			{
				a:        "1",
				b:        "1",
				c:        "1",
				expected: pdp.EffectPermit,
			},
			{
				a:        "0",
				b:        "1",
				c:        "0",
				expected: pdp.EffectPermit,
			},
			{
				a:        "4",
				b:        "2",
				c:        "2",
				expected: pdp.EffectPermit,
			},
			{
				a:        "7",
				b:        "2",
				c:        "3",
				expected: pdp.EffectPermit,
			},
			{
				a:        "-1",
				b:        "1",
				c:        "-1",
				expected: pdp.EffectPermit,
			},
			{
				a:        "1",
				b:        "-1",
				c:        "-1",
				expected: pdp.EffectPermit,
			},
			{
				a:        "2",
				b:        "1",
				c:        "1",
				expected: pdp.EffectNotApplicable,
			},
			{
				a:        "-1",
				b:        "1",
				c:        "1",
				expected: pdp.EffectNotApplicable,
			},
			{
				a:        "0",
				b:        "1",
				c:        "1",
				expected: pdp.EffectNotApplicable,
			},
		},
	}

	validateTestSuite(ts, t)
}
