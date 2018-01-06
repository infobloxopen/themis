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
	attrs    []*pb.Attribute
	expected int
}

type testSuite struct {
	policy  string
	testSet []testCase
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

func createRequest(tc testCase) *pb.Request {
	return &pb.Request{
		Attributes: tc.attrs,
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
		t.Run(fmt.Sprintf("%v", tc.attrs), func(t *testing.T) {
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
		policy: `# Policy set for Integer Equal Comparison
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
`,
		testSet: []testCase{
			{
				attrs: []*pb.Attribute{
					{
						Id:    "a",
						Type:  "Integer",
						Value: "1",
					},
					{
						Id:    "b",
						Type:  "Integer",
						Value: "1",
					},
				},
				expected: pdp.EffectPermit,
			},
			{
				attrs: []*pb.Attribute{
					{
						Id:    "a",
						Type:  "Integer",
						Value: "0",
					},
					{
						Id:    "b",
						Type:  "Integer",
						Value: "0",
					},
				},
				expected: pdp.EffectPermit,
			},
			{
				attrs: []*pb.Attribute{
					{
						Id:    "a",
						Type:  "Integer",
						Value: "-1",
					},
					{
						Id:    "b",
						Type:  "Integer",
						Value: "-1",
					},
				},
				expected: pdp.EffectPermit,
			},
			{
				attrs: []*pb.Attribute{
					{
						Id:    "a",
						Type:  "Integer",
						Value: "1",
					},
					{
						Id:    "b",
						Type:  "Integer",
						Value: "0",
					},
				},
				expected: pdp.EffectNotApplicable,
			},
			{
				attrs: []*pb.Attribute{
					{
						Id:    "a",
						Type:  "Integer",
						Value: "0",
					},
					{
						Id:    "b",
						Type:  "Integer",
						Value: "1",
					},
				},
				expected: pdp.EffectNotApplicable,
			},
			{
				attrs: []*pb.Attribute{
					{
						Id:    "a",
						Type:  "Integer",
						Value: "1",
					},
					{
						Id:    "b",
						Type:  "Integer",
						Value: "-2",
					},
				},
				expected: pdp.EffectNotApplicable,
			},
		},
	}

	validateTestSuite(ts, t)
}

func TestIntegerGreater(t *testing.T) {
	ts := testSuite{
		policy: `# Policy set for Integer Greater Comparison
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
`,
		testSet: []testCase{
			{
				attrs: []*pb.Attribute{
					{
						Id:    "a",
						Type:  "Integer",
						Value: "1",
					},
					{
						Id:    "b",
						Type:  "Integer",
						Value: "0",
					},
				},
				expected: pdp.EffectPermit,
			},
			{
				attrs: []*pb.Attribute{
					{
						Id:    "a",
						Type:  "Integer",
						Value: "0",
					},
					{
						Id:    "b",
						Type:  "Integer",
						Value: "-1",
					},
				},
				expected: pdp.EffectPermit,
			},
			{
				attrs: []*pb.Attribute{
					{
						Id:    "a",
						Type:  "Integer",
						Value: "1",
					},
					{
						Id:    "b",
						Type:  "Integer",
						Value: "-1",
					},
				},
				expected: pdp.EffectPermit,
			},
			{
				attrs: []*pb.Attribute{
					{
						Id:    "a",
						Type:  "Integer",
						Value: "0",
					},
					{
						Id:    "b",
						Type:  "Integer",
						Value: "0",
					},
				},
				expected: pdp.EffectNotApplicable,
			},
			{
				attrs: []*pb.Attribute{
					{
						Id:    "a",
						Type:  "Integer",
						Value: "-1",
					},
					{
						Id:    "b",
						Type:  "Integer",
						Value: "1",
					},
				},
				expected: pdp.EffectNotApplicable,
			},
		},
	}

	validateTestSuite(ts, t)
}

func TestIntegerAdd(t *testing.T) {
	ts := testSuite{
		policy: `# Policy set for Integer Addition
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
`,
		testSet: []testCase{
			{
				attrs: []*pb.Attribute{
					{
						Id:    "a",
						Type:  "Integer",
						Value: "1",
					},
					{
						Id:    "b",
						Type:  "Integer",
						Value: "1",
					},
					{
						Id:    "c",
						Type:  "Integer",
						Value: "2",
					},
				},
				expected: pdp.EffectPermit,
			},
			{
				attrs: []*pb.Attribute{
					{
						Id:    "a",
						Type:  "Integer",
						Value: "0",
					},
					{
						Id:    "b",
						Type:  "Integer",
						Value: "0",
					},
					{
						Id:    "c",
						Type:  "Integer",
						Value: "0",
					},
				},
				expected: pdp.EffectPermit,
			},
			{
				attrs: []*pb.Attribute{
					{
						Id:    "a",
						Type:  "Integer",
						Value: "-1",
					},
					{
						Id:    "b",
						Type:  "Integer",
						Value: "-1",
					},
					{
						Id:    "c",
						Type:  "Integer",
						Value: "-2",
					},
				},
				expected: pdp.EffectPermit,
			},
			{
				attrs: []*pb.Attribute{
					{
						Id:    "a",
						Type:  "Integer",
						Value: "1",
					},
					{
						Id:    "b",
						Type:  "Integer",
						Value: "0",
					},
					{
						Id:    "c",
						Type:  "Integer",
						Value: "2",
					},
				},
				expected: pdp.EffectNotApplicable,
			},
			{
				attrs: []*pb.Attribute{
					{
						Id:    "a",
						Type:  "Integer",
						Value: "0",
					},
					{
						Id:    "b",
						Type:  "Integer",
						Value: "1",
					},
					{
						Id:    "c",
						Type:  "Integer",
						Value: "2",
					},
				},
				expected: pdp.EffectNotApplicable,
			},
			{
				attrs: []*pb.Attribute{
					{
						Id:    "a",
						Type:  "Integer",
						Value: "1",
					},
					{
						Id:    "b",
						Type:  "Integer",
						Value: "-1",
					},
					{
						Id:    "c",
						Type:  "Integer",
						Value: "2",
					},
				},
				expected: pdp.EffectNotApplicable,
			},
		},
	}

	validateTestSuite(ts, t)
}

func TestIntegerSubtract(t *testing.T) {
	ts := testSuite{
		policy: `# Policy set for Integer Subtraction
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
`,
		testSet: []testCase{
			{
				attrs: []*pb.Attribute{
					{
						Id:    "a",
						Type:  "Integer",
						Value: "1",
					},
					{
						Id:    "b",
						Type:  "Integer",
						Value: "1",
					},
					{
						Id:    "c",
						Type:  "Integer",
						Value: "0",
					},
				},
				expected: pdp.EffectPermit,
			},
			{
				attrs: []*pb.Attribute{
					{
						Id:    "a",
						Type:  "Integer",
						Value: "0",
					},
					{
						Id:    "b",
						Type:  "Integer",
						Value: "0",
					},
					{
						Id:    "c",
						Type:  "Integer",
						Value: "0",
					},
				},
				expected: pdp.EffectPermit,
			},
			{
				attrs: []*pb.Attribute{
					{
						Id:    "a",
						Type:  "Integer",
						Value: "-1",
					},
					{
						Id:    "b",
						Type:  "Integer",
						Value: "-1",
					},
					{
						Id:    "c",
						Type:  "Integer",
						Value: "0",
					},
				},
				expected: pdp.EffectPermit,
			},
			{
				attrs: []*pb.Attribute{
					{
						Id:    "a",
						Type:  "Integer",
						Value: "1",
					},
					{
						Id:    "b",
						Type:  "Integer",
						Value: "0",
					},
					{
						Id:    "c",
						Type:  "Integer",
						Value: "0",
					},
				},
				expected: pdp.EffectNotApplicable,
			},
			{
				attrs: []*pb.Attribute{
					{
						Id:    "a",
						Type:  "Integer",
						Value: "0",
					},
					{
						Id:    "b",
						Type:  "Integer",
						Value: "1",
					},
					{
						Id:    "c",
						Type:  "Integer",
						Value: "0",
					},
				},
				expected: pdp.EffectNotApplicable,
			},
			{
				attrs: []*pb.Attribute{
					{
						Id:    "a",
						Type:  "Integer",
						Value: "1",
					},
					{
						Id:    "b",
						Type:  "Integer",
						Value: "-1",
					},
					{
						Id:    "c",
						Type:  "Integer",
						Value: "0",
					},
				},
				expected: pdp.EffectNotApplicable,
			},
		},
	}

	validateTestSuite(ts, t)
}

func TestIntegerMultiply(t *testing.T) {
	ts := testSuite{
		policy: `# Policy set for Integer Multiplication
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
`,
		testSet: []testCase{
			{
				attrs: []*pb.Attribute{
					{
						Id:    "a",
						Type:  "Integer",
						Value: "1",
					},
					{
						Id:    "b",
						Type:  "Integer",
						Value: "1",
					},
					{
						Id:    "c",
						Type:  "Integer",
						Value: "1",
					},
				},
				expected: pdp.EffectPermit,
			},
			{
				attrs: []*pb.Attribute{
					{
						Id:    "a",
						Type:  "Integer",
						Value: "1",
					},
					{
						Id:    "b",
						Type:  "Integer",
						Value: "0",
					},
					{
						Id:    "c",
						Type:  "Integer",
						Value: "0",
					},
				},
				expected: pdp.EffectPermit,
			},
			{
				attrs: []*pb.Attribute{
					{
						Id:    "a",
						Type:  "Integer",
						Value: "0",
					},
					{
						Id:    "b",
						Type:  "Integer",
						Value: "1",
					},
					{
						Id:    "c",
						Type:  "Integer",
						Value: "0",
					},
				},
				expected: pdp.EffectPermit,
			},
			{
				attrs: []*pb.Attribute{
					{
						Id:    "a",
						Type:  "Integer",
						Value: "-1",
					},
					{
						Id:    "b",
						Type:  "Integer",
						Value: "-1",
					},
					{
						Id:    "c",
						Type:  "Integer",
						Value: "1",
					},
				},
				expected: pdp.EffectPermit,
			},
			{
				attrs: []*pb.Attribute{
					{
						Id:    "a",
						Type:  "Integer",
						Value: "-1",
					},
					{
						Id:    "b",
						Type:  "Integer",
						Value: "1",
					},
					{
						Id:    "c",
						Type:  "Integer",
						Value: "-1",
					},
				},
				expected: pdp.EffectPermit,
			},
			{
				attrs: []*pb.Attribute{
					{
						Id:    "a",
						Type:  "Integer",
						Value: "1",
					},
					{
						Id:    "b",
						Type:  "Integer",
						Value: "0",
					},
					{
						Id:    "c",
						Type:  "Integer",
						Value: "1",
					},
				},
				expected: pdp.EffectNotApplicable,
			},
			{
				attrs: []*pb.Attribute{
					{
						Id:    "a",
						Type:  "Integer",
						Value: "-1",
					},
					{
						Id:    "b",
						Type:  "Integer",
						Value: "1",
					},
					{
						Id:    "c",
						Type:  "Integer",
						Value: "1",
					},
				},
				expected: pdp.EffectNotApplicable,
			},
			{
				attrs: []*pb.Attribute{
					{
						Id:    "a",
						Type:  "Integer",
						Value: "0",
					},
					{
						Id:    "b",
						Type:  "Integer",
						Value: "1",
					},
					{
						Id:    "c",
						Type:  "Integer",
						Value: "1",
					},
				},
				expected: pdp.EffectNotApplicable,
			},
		},
	}

	validateTestSuite(ts, t)
}

func TestIntegerDivide(t *testing.T) {
	ts := testSuite{
		policy: `# Policy set for Integer Division
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
`,
		testSet: []testCase{
			{
				attrs: []*pb.Attribute{
					{
						Id:    "a",
						Type:  "Integer",
						Value: "1",
					},
					{
						Id:    "b",
						Type:  "Integer",
						Value: "1",
					},
					{
						Id:    "c",
						Type:  "Integer",
						Value: "1",
					},
				},
				expected: pdp.EffectPermit,
			},
			{
				attrs: []*pb.Attribute{
					{
						Id:    "a",
						Type:  "Integer",
						Value: "0",
					},
					{
						Id:    "b",
						Type:  "Integer",
						Value: "1",
					},
					{
						Id:    "c",
						Type:  "Integer",
						Value: "0",
					},
				},
				expected: pdp.EffectPermit,
			},
			{
				attrs: []*pb.Attribute{
					{
						Id:    "a",
						Type:  "Integer",
						Value: "4",
					},
					{
						Id:    "b",
						Type:  "Integer",
						Value: "2",
					},
					{
						Id:    "c",
						Type:  "Integer",
						Value: "2",
					},
				},
				expected: pdp.EffectPermit,
			},
			{
				attrs: []*pb.Attribute{
					{
						Id:    "a",
						Type:  "Integer",
						Value: "7",
					},
					{
						Id:    "b",
						Type:  "Integer",
						Value: "2",
					},
					{
						Id:    "c",
						Type:  "Integer",
						Value: "3",
					},
				},
				expected: pdp.EffectPermit,
			},
			{
				attrs: []*pb.Attribute{
					{
						Id:    "a",
						Type:  "Integer",
						Value: "-1",
					},
					{
						Id:    "b",
						Type:  "Integer",
						Value: "1",
					},
					{
						Id:    "c",
						Type:  "Integer",
						Value: "-1",
					},
				},
				expected: pdp.EffectPermit,
			},
			{
				attrs: []*pb.Attribute{
					{
						Id:    "a",
						Type:  "Integer",
						Value: "1",
					},
					{
						Id:    "b",
						Type:  "Integer",
						Value: "-1",
					},
					{
						Id:    "c",
						Type:  "Integer",
						Value: "-1",
					},
				},
				expected: pdp.EffectPermit,
			},
			{
				attrs: []*pb.Attribute{
					{
						Id:    "a",
						Type:  "Integer",
						Value: "2",
					},
					{
						Id:    "b",
						Type:  "Integer",
						Value: "1",
					},
					{
						Id:    "c",
						Type:  "Integer",
						Value: "1",
					},
				},
				expected: pdp.EffectNotApplicable,
			},
			{
				attrs: []*pb.Attribute{
					{
						Id:    "a",
						Type:  "Integer",
						Value: "-1",
					},
					{
						Id:    "b",
						Type:  "Integer",
						Value: "1",
					},
					{
						Id:    "c",
						Type:  "Integer",
						Value: "1",
					},
				},
				expected: pdp.EffectNotApplicable,
			},
			{
				attrs: []*pb.Attribute{
					{
						Id:    "a",
						Type:  "Integer",
						Value: "0",
					},
					{
						Id:    "b",
						Type:  "Integer",
						Value: "1",
					},
					{
						Id:    "c",
						Type:  "Integer",
						Value: "1",
					},
				},
				expected: pdp.EffectNotApplicable,
			},
		},
	}

	validateTestSuite(ts, t)
}

func TestFloatGreater(t *testing.T) {
	ts := testSuite{
		policy: `# Policy set for Float Greater Comparison
attributes:
  a: float
  b: float

policies:
  alg: FirstApplicableEffect
  rules:
  - id: "Test Float Greater"
    condition: # a > b
      greater:
        - attr: a
        - attr: b
    effect: Permit
`,
		testSet: []testCase{
			{
				attrs: []*pb.Attribute{
					{
						Id:    "a",
						Type:  "Float",
						Value: "1.0",
					},
					{
						Id:    "b",
						Type:  "Float",
						Value: "0.9",
					},
				},
				expected: pdp.EffectPermit,
			},
			{
				attrs: []*pb.Attribute{
					{
						Id:    "a",
						Type:  "Float",
						Value: "1.0",
					},
					{
						Id:    "b",
						Type:  "Float",
						Value: "-1.0",
					},
				},
				expected: pdp.EffectPermit,
			},
			{
				attrs: []*pb.Attribute{
					{
						Id:    "a",
						Type:  "Float",
						Value: "0.0",
					},
					{
						Id:    "b",
						Type:  "Float",
						Value: "-1.0",
					},
				},
				expected: pdp.EffectPermit,
			},
			{
				attrs: []*pb.Attribute{
					{
						Id:    "a",
						Type:  "Float",
						Value: "0.8",
					},
					{
						Id:    "b",
						Type:  "Float",
						Value: "0.9",
					},
				},
				expected: pdp.EffectNotApplicable,
			},
			{
				attrs: []*pb.Attribute{
					{
						Id:    "a",
						Type:  "Float",
						Value: "-2.0",
					},
					{
						Id:    "b",
						Type:  "Float",
						Value: "1.0",
					},
				},
				expected: pdp.EffectNotApplicable,
			},
			{
				attrs: []*pb.Attribute{
					{
						Id:    "a",
						Type:  "Float",
						Value: "-1.0",
					},
					{
						Id:    "b",
						Type:  "Float",
						Value: "0.0",
					},
				},
				expected: pdp.EffectNotApplicable,
			},
		},
	}

	validateTestSuite(ts, t)
}

func TestFloatAdd(t *testing.T) {
	ts := testSuite{
		policy: `# Policy set for Integer Addition
attributes:
  a: float
  b: float
  c: float

policies:
  alg: FirstApplicableEffect
  rules:
  - id: "Test Float Addition"
    condition: # a + b == c
      equal:
      - add: # a + b
        - attr: a
        - attr: b
      - attr: c
    effect: Permit
`,
		testSet: []testCase{
			{
				attrs: []*pb.Attribute{
					{
						Id:    "a",
						Type:  "float",
						Value: "1.",
					},
					{
						Id:    "b",
						Type:  "float",
						Value: "1.",
					},
					{
						Id:    "c",
						Type:  "float",
						Value: "2.",
					},
				},
				expected: pdp.EffectPermit,
			},
			{
				attrs: []*pb.Attribute{
					{
						Id:    "a",
						Type:  "float",
						Value: "0.",
					},
					{
						Id:    "b",
						Type:  "float",
						Value: "0.",
					},
					{
						Id:    "c",
						Type:  "float",
						Value: "0.",
					},
				},
				expected: pdp.EffectPermit,
			},
			{
				attrs: []*pb.Attribute{
					{
						Id:    "a",
						Type:  "float",
						Value: "-1.",
					},
					{
						Id:    "b",
						Type:  "float",
						Value: "-1.",
					},
					{
						Id:    "c",
						Type:  "float",
						Value: "-2.",
					},
				},
				expected: pdp.EffectPermit,
			},
			{
				attrs: []*pb.Attribute{
					{
						Id:    "a",
						Type:  "float",
						Value: "1.",
					},
					{
						Id:    "b",
						Type:  "float",
						Value: "0.",
					},
					{
						Id:    "c",
						Type:  "float",
						Value: "0.",
					},
				},
				expected: pdp.EffectNotApplicable,
			},
			{
				attrs: []*pb.Attribute{
					{
						Id:    "a",
						Type:  "float",
						Value: "0.",
					},
					{
						Id:    "b",
						Type:  "float",
						Value: "1.",
					},
					{
						Id:    "c",
						Type:  "float",
						Value: "2.",
					},
				},
				expected: pdp.EffectNotApplicable,
			},
			{
				attrs: []*pb.Attribute{
					{
						Id:    "a",
						Type:  "float",
						Value: "1.",
					},
					{
						Id:    "b",
						Type:  "float",
						Value: "-1.",
					},
					{
						Id:    "c",
						Type:  "float",
						Value: "1.",
					},
				},
				expected: pdp.EffectNotApplicable,
			},
		},
	}

	validateTestSuite(ts, t)
}

func TestFloatSubtract(t *testing.T) {
	ts := testSuite{
		policy: `# Policy set for float Subtraction
attributes:
  a: float
  b: float
  c: float

policies:
  alg: FirstApplicableEffect
  rules:
  - id: "Test Float Subtraction"
    condition: # a - b == c
      equal:
      - subtract:
        - attr: a
        - attr: b
      - attr: c
    effect: Permit
`,
		testSet: []testCase{
			{
				attrs: []*pb.Attribute{
					{
						Id:    "a",
						Type:  "Float",
						Value: "1.",
					},
					{
						Id:    "b",
						Type:  "Float",
						Value: "1.",
					},
					{
						Id:    "c",
						Type:  "Float",
						Value: "0.",
					},
				},
				expected: pdp.EffectPermit,
			},
			{
				attrs: []*pb.Attribute{
					{
						Id:    "a",
						Type:  "Float",
						Value: "0.",
					},
					{
						Id:    "b",
						Type:  "Float",
						Value: "0.",
					},
					{
						Id:    "c",
						Type:  "Float",
						Value: "0.",
					},
				},
				expected: pdp.EffectPermit,
			},
			{
				attrs: []*pb.Attribute{
					{
						Id:    "a",
						Type:  "Float",
						Value: "-1.",
					},
					{
						Id:    "b",
						Type:  "Float",
						Value: "-1.",
					},
					{
						Id:    "c",
						Type:  "Float",
						Value: "0.",
					},
				},
				expected: pdp.EffectPermit,
			},
			{
				attrs: []*pb.Attribute{
					{
						Id:    "a",
						Type:  "Float",
						Value: "1.",
					},
					{
						Id:    "b",
						Type:  "Float",
						Value: "0.",
					},
					{
						Id:    "c",
						Type:  "Float",
						Value: "0.",
					},
				},
				expected: pdp.EffectNotApplicable,
			},
			{
				attrs: []*pb.Attribute{
					{
						Id:    "a",
						Type:  "Float",
						Value: "0.",
					},
					{
						Id:    "b",
						Type:  "Float",
						Value: "1.",
					},
					{
						Id:    "c",
						Type:  "Float",
						Value: "0.",
					},
				},
				expected: pdp.EffectNotApplicable,
			},
			{
				attrs: []*pb.Attribute{
					{
						Id:    "a",
						Type:  "Float",
						Value: "1.",
					},
					{
						Id:    "b",
						Type:  "Float",
						Value: "-1.",
					},
					{
						Id:    "c",
						Type:  "Float",
						Value: "0.",
					},
				},
				expected: pdp.EffectNotApplicable,
			},
		},
	}

	validateTestSuite(ts, t)
}

func TestFloatMultiply(t *testing.T) {
	ts := testSuite{
		policy: `# Policy set for Float Multiplication
attributes:
  a: float
  b: float
  c: float

policies:
  alg: FirstApplicableEffect
  rules:
  - id: "Test Float Multiplication"
    condition: # a * b == c
      equal:
      - multiply: # a * b
        - attr: a
        - attr: b
      - attr: c
    effect: Permit
`,
		testSet: []testCase{
			{
				attrs: []*pb.Attribute{
					{
						Id:    "a",
						Type:  "Float",
						Value: "1.",
					},
					{
						Id:    "b",
						Type:  "Float",
						Value: "1.",
					},
					{
						Id:    "c",
						Type:  "Float",
						Value: "1.",
					},
				},
				expected: pdp.EffectPermit,
			},
			{
				attrs: []*pb.Attribute{
					{
						Id:    "a",
						Type:  "Float",
						Value: "1.",
					},
					{
						Id:    "b",
						Type:  "Float",
						Value: "0.",
					},
					{
						Id:    "c",
						Type:  "Float",
						Value: "0.",
					},
				},
				expected: pdp.EffectPermit,
			},
			{
				attrs: []*pb.Attribute{
					{
						Id:    "a",
						Type:  "Float",
						Value: "0.",
					},
					{
						Id:    "b",
						Type:  "Float",
						Value: "1.",
					},
					{
						Id:    "c",
						Type:  "Float",
						Value: "0.",
					},
				},
				expected: pdp.EffectPermit,
			},
			{
				attrs: []*pb.Attribute{
					{
						Id:    "a",
						Type:  "Float",
						Value: "-1.",
					},
					{
						Id:    "b",
						Type:  "Float",
						Value: "-1.",
					},
					{
						Id:    "c",
						Type:  "Float",
						Value: "1.",
					},
				},
				expected: pdp.EffectPermit,
			},
			{
				attrs: []*pb.Attribute{
					{
						Id:    "a",
						Type:  "Float",
						Value: "-1.",
					},
					{
						Id:    "b",
						Type:  "Float",
						Value: "1.",
					},
					{
						Id:    "c",
						Type:  "Float",
						Value: "-1.",
					},
				},
				expected: pdp.EffectPermit,
			},
			{
				attrs: []*pb.Attribute{
					{
						Id:    "a",
						Type:  "Float",
						Value: "1.",
					},
					{
						Id:    "b",
						Type:  "Float",
						Value: "0.",
					},
					{
						Id:    "c",
						Type:  "Float",
						Value: "1.",
					},
				},
				expected: pdp.EffectNotApplicable,
			},
			{
				attrs: []*pb.Attribute{
					{
						Id:    "a",
						Type:  "Float",
						Value: "-1.",
					},
					{
						Id:    "b",
						Type:  "Float",
						Value: "1.",
					},
					{
						Id:    "c",
						Type:  "Float",
						Value: "1.",
					},
				},
				expected: pdp.EffectNotApplicable,
			},
			{
				attrs: []*pb.Attribute{
					{
						Id:    "a",
						Type:  "Float",
						Value: "0.",
					},
					{
						Id:    "b",
						Type:  "Float",
						Value: "1.",
					},
					{
						Id:    "c",
						Type:  "Float",
						Value: "1.",
					},
				},
				expected: pdp.EffectNotApplicable,
			},
		},
	}

	validateTestSuite(ts, t)
}

func TestFloatDivide(t *testing.T) {
	ts := testSuite{
		policy: `# Policy set for Float Division
attributes:
  a: float
  b: float
  c: float

policies:
  alg: FirstApplicableEffect
  rules:
  - id: "Test Float Division"
    condition: # a / b == c
      equal:
      - divide: # a / b
        - attr: a
        - attr: b
      - attr: c
    effect: Permit
`,
		testSet: []testCase{
			{
				attrs: []*pb.Attribute{
					{
						Id:    "a",
						Type:  "Float",
						Value: "1.",
					},
					{
						Id:    "b",
						Type:  "Float",
						Value: "1.",
					},
					{
						Id:    "c",
						Type:  "Float",
						Value: "1.",
					},
				},
				expected: pdp.EffectPermit,
			},
			{
				attrs: []*pb.Attribute{
					{
						Id:    "a",
						Type:  "Float",
						Value: "0.",
					},
					{
						Id:    "b",
						Type:  "Float",
						Value: "1.",
					},
					{
						Id:    "c",
						Type:  "Float",
						Value: "0.",
					},
				},
				expected: pdp.EffectPermit,
			},
			{
				attrs: []*pb.Attribute{
					{
						Id:    "a",
						Type:  "Float",
						Value: "4.",
					},
					{
						Id:    "b",
						Type:  "Float",
						Value: "2.",
					},
					{
						Id:    "c",
						Type:  "Float",
						Value: "2.",
					},
				},
				expected: pdp.EffectPermit,
			},
			{
				attrs: []*pb.Attribute{
					{
						Id:    "a",
						Type:  "Float",
						Value: "7.",
					},
					{
						Id:    "b",
						Type:  "Float",
						Value: "2.",
					},
					{
						Id:    "c",
						Type:  "Float",
						Value: "3.5",
					},
				},
				expected: pdp.EffectPermit,
			},
			{
				attrs: []*pb.Attribute{
					{
						Id:    "a",
						Type:  "Float",
						Value: "-1.",
					},
					{
						Id:    "b",
						Type:  "Float",
						Value: "1.",
					},
					{
						Id:    "c",
						Type:  "Float",
						Value: "-1.",
					},
				},
				expected: pdp.EffectPermit,
			},
			{
				attrs: []*pb.Attribute{
					{
						Id:    "a",
						Type:  "Float",
						Value: "1.",
					},
					{
						Id:    "b",
						Type:  "Float",
						Value: "-1.",
					},
					{
						Id:    "c",
						Type:  "Float",
						Value: "-1.",
					},
				},
				expected: pdp.EffectPermit,
			},
			{
				attrs: []*pb.Attribute{
					{
						Id:    "a",
						Type:  "Float",
						Value: "2.",
					},
					{
						Id:    "b",
						Type:  "Float",
						Value: "1.",
					},
					{
						Id:    "c",
						Type:  "Float",
						Value: "1.",
					},
				},
				expected: pdp.EffectNotApplicable,
			},
			{
				attrs: []*pb.Attribute{
					{
						Id:    "a",
						Type:  "Float",
						Value: "-1.",
					},
					{
						Id:    "b",
						Type:  "Float",
						Value: "1.",
					},
					{
						Id:    "c",
						Type:  "Float",
						Value: "1.",
					},
				},
				expected: pdp.EffectNotApplicable,
			},
			{
				attrs: []*pb.Attribute{
					{
						Id:    "a",
						Type:  "Float",
						Value: "0.",
					},
					{
						Id:    "b",
						Type:  "Float",
						Value: "1.",
					},
					{
						Id:    "c",
						Type:  "Float",
						Value: "1.",
					},
				},
				expected: pdp.EffectNotApplicable,
			},
		},
	}

	validateTestSuite(ts, t)
}
