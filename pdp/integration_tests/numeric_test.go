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

type policyFormat int

const (
	YAML policyFormat = iota
	JSON
)

var policyFormatString = map[policyFormat]string{
	YAML: "YAML",
	JSON: "JSON",
}

func (f policyFormat) String() string {
	return policyFormatString[f]
}

type testCase struct {
	attrs              []*pb.Attribute
	expected           int
	expectedObligation string
	expectedError      string
}

type testSuite struct {
	policies map[policyFormat]string
	testSet  []testCase
}

func init() {
	log.SetLevel(log.ErrorLevel)
}

func loadPolicy(pf policyFormat, ps string) *pdp.PolicyStorage {
	var parser ast.Parser
	switch pf {
	case YAML:
		parser = ast.NewYAMLParser()
	case JSON:
		parser = ast.NewJSONParser()
	}
	policyStorage, err := parser.Unmarshal(strings.NewReader(ps), nil)
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
	for _, tc := range ts.testSet {
		t.Run(fmt.Sprintf("%v", tc.attrs), func(t *testing.T) {
			req := createRequest(tc)
			ctx, err := createContext(req)
			if err != nil {
				t.Fatalf("Expected no error while creating context but got: %s", err)
			}

			for pf, ps := range ts.policies {
				t.Run(fmt.Sprintf("Policy Format: %s", pf), func(t *testing.T) {
					p := loadPolicy(pf, ps)
					r := p.Root().Calculate(ctx)
					effect, obligations, err := r.Status()
					if err != nil {
						if tc.expectedError == "" {
							t.Fatalf("Expected no error while evaluating policy but got: %s", err)
						} else if !strings.Contains(err.Error(), tc.expectedError) {
							t.Fatalf("Expected error while evaluating policy '%s', but got '%s'", tc.expectedError, err)
						}
					}

					if effect != tc.expected {
						t.Fatalf("Expected result of policy evaluation %s, but got %s",
							pdp.EffectNameFromEnum(tc.expected), pdp.EffectNameFromEnum(effect))
					}
					if tc.expectedObligation != "" {
						obLen := len(obligations)
						if obLen != 1 {
							t.Fatalf("Expected result of policy evaluation include 1 obligation, but got %d", obLen)
						}
						_, _, obligationRes, err := obligations[0].Serialize(ctx)
						if err != nil {
							t.Fatalf("Expected when serializing obligation, but got %s", err)
						}
						if obligationRes != tc.expectedObligation {
							t.Fatalf("Expected obligation of policy evaluation %s, but got %s", tc.expectedObligation, obligationRes)
						}
					}
				})
			}
		})
	}
}

func TestIntegerEqual(t *testing.T) {
	ts := testSuite{
		policies: map[policyFormat]string{
			YAML: `# Policy set for Integer Equal Comparison
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
			JSON: `{
  "attributes": {
    "a": "integer",
    "b": "integer"
  },
  "policies": {
    "id": "Test Integer Equal Policies",
    "alg": "FirstApplicableEffect",
    "rules": [
      {
        "id": "Test Integer Equal Rule",
        "condition": {
          "equal": [
            {
              "attr": "a"
            },
            {
              "attr": "b"
            }
          ]
        },
        "effect": "Permit"
      }
    ]
  }
}`,
		},
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
		policies: map[policyFormat]string{
			YAML: `# Policy set for Integer Greater Comparison
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
			JSON: `{
  "attributes": {
    "a": "integer",
    "b": "integer"
  },
  "policies": {
    "id": "Test Integer Greater Policies",
    "alg": "FirstApplicableEffect",
    "rules": [
      {
        "id": "Test Integer Greater Rule",
        "condition": {
          "greater": [
            {
              "attr": "a"
            },
            {
              "attr": "b"
            }
          ]
        },
        "effect": "Permit"
      }
    ]
  }
}`,
		},
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
		policies: map[policyFormat]string{
			YAML: `# Policy set for Integer Addition
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
			JSON: `{
  "attributes": {
    "a": "integer",
    "b": "integer",
    "c": "integer"
  },
  "policies": {
    "id": "Test Integer Addition Policies",
    "alg": "FirstApplicableEffect",
    "rules": [
      {
        "id": "Test Integer Addition Rule",
        "condition": {
          "equal": [
            {
              "add": [
                {
                  "attr": "a"
                },
                {
                  "attr": "b"
                }
              ]
            },
            {
              "attr": "c"
            }
          ]
        },
        "effect": "Permit"
      }
    ]
  }
}`,
		},
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
		policies: map[policyFormat]string{
			YAML: `# Policy set for Integer Subtraction
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
			JSON: `{
  "attributes": {
    "a": "integer",
    "b": "integer",
    "c": "integer"
  },
  "policies": {
    "id": "Test Integer Subtraction Policies",
    "alg": "FirstApplicableEffect",
    "rules": [
      {
        "id": "Test Integer Subtraction Rule",
        "condition": {
          "equal": [
            {
              "subtract": [
                {
                  "attr": "a"
                },
                {
                  "attr": "b"
                }
              ]
            },
            {
              "attr": "c"
            }
          ]
        },
        "effect": "Permit"
      }
    ]
  }
}`,
		},
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
		policies: map[policyFormat]string{
			YAML: `# Policy set for Integer Multiplication
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
			JSON: `{
  "attributes": {
    "a": "integer",
    "b": "integer",
    "c": "integer"
  },
  "policies": {
    "id": "Test Integer Multiplication Policies",
    "alg": "FirstApplicableEffect",
    "rules": [
      {
        "id": "Test Integer Multiplication Rule",
        "condition": {
          "equal": [
            {
              "multiply": [
                {
                  "attr": "a"
                },
                {
                  "attr": "b"
                }
              ]
            },
            {
              "attr": "c"
            }
          ]
        },
        "effect": "Permit"
      }
    ]
  }
}`,
		},
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
		policies: map[policyFormat]string{
			YAML: `# Policy set for Integer Division
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
			JSON: `{
  "attributes": {
    "a": "integer",
    "b": "integer",
    "c": "integer"
  },
  "policies": {
    "id": "Test Integer Division Policies",
    "alg": "FirstApplicableEffect",
    "rules": [
      {
        "id": "Test Integer Division Rule",
        "condition": {
          "equal": [
            {
              "divide": [
                {
                  "attr": "a"
                },
                {
                  "attr": "b"
                }
              ]
            },
            {
              "attr": "c"
            }
          ]
        },
        "effect": "Permit"
      }
    ]
  }
}`,
		},
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
				expected:      pdp.EffectIndeterminateP,
				expectedError: "Integer divisor has a value of 0",
			},
		},
	}

	validateTestSuite(ts, t)
}

func TestFloatGreater(t *testing.T) {
	ts := testSuite{
		policies: map[policyFormat]string{
			YAML: `# Policy set for Float Greater Comparison
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
			JSON: `{
  "attributes": {
    "a": "float",
    "b": "float"
  },
  "policies": {
    "id": "Test Float Greater Policies",
    "alg": "FirstApplicableEffect",
    "rules": [
      {
        "id": "Test Float Greater Rule",
        "condition": {
          "greater": [
            {
              "attr": "a"
            },
            {
              "attr": "b"
            }
          ]
        },
        "effect": "Permit"
      }
    ]
  }
}`,
		},
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
		policies: map[policyFormat]string{
			YAML: `# Policy set for Integer Addition
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
			JSON: `{
  "attributes": {
    "a": "float",
    "b": "float",
    "c": "float"
  },
  "policies": {
    "id": "Test Float Addition Policies",
    "alg": "FirstApplicableEffect",
    "rules": [
      {
        "id": "Test Float Addition Rule",
        "condition": {
          "equal": [
            {
              "add": [
                {
                  "attr": "a"
                },
                {
                  "attr": "b"
                }
              ]
            },
            {
              "attr": "c"
            }
          ]
        },
        "effect": "Permit"
      }
    ]
  }
}`,
		},
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
		policies: map[policyFormat]string{
			YAML: `# Policy set for float Subtraction
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
			JSON: `{
  "attributes": {
    "a": "float",
    "b": "float",
    "c": "float"
  },
  "policies": {
    "id": "Test Float Subtraction Policies",
    "alg": "FirstApplicableEffect",
    "rules": [
      {
        "id": "Test Float Subtraction Rule",
        "condition": {
          "equal": [
            {
              "subtract": [
                {
                  "attr": "a"
                },
                {
                  "attr": "b"
                }
              ]
            },
            {
              "attr": "c"
            }
          ]
        },
        "effect": "Permit"
      }
    ]
  }
}`,
		},
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
		policies: map[policyFormat]string{
			YAML: `# Policy set for Float Multiplication
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
			JSON: `
{
  "attributes": {
    "a": "float",
    "b": "float",
    "c": "float"
  },
  "policies": {
    "id": "Test Float Multiplication Policies",
    "alg": "FirstApplicableEffect",
    "rules": [
      {
        "id": "Test Float Multiplication Rule",
        "condition": {
          "equal": [
            {
              "multiply": [
                {
                  "attr": "a"
                },
                {
                  "attr": "b"
                }
              ]
            },
            {
                "attr": "c"
            }
          ]
        },
        "effect": "Permit"
      }
    ]
  }
}
`,
		},
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
						Value: "1.9e+200",
					},
					{
						Id:    "b",
						Type:  "Float",
						Value: "1.9e+233",
					},
					{
						Id:    "c",
						Type:  "Float",
						Value: "1.",
					},
				},
				expected:      pdp.EffectIndeterminateP,
				expectedError: "Float result has a value of Inf",
			},
		},
	}

	validateTestSuite(ts, t)
}

func TestFloatDivide(t *testing.T) {
	ts := testSuite{
		policies: map[policyFormat]string{
			YAML: `# Policy set for Float Division
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
			JSON: `
{
  "attributes": {
    "a": "float",
    "b": "float",
    "c": "float"
  },
  "policies": {
    "id": "Test Float Division Policies",
    "alg": "FirstApplicableEffect",
    "rules": [
      {
        "id": "Test Float Division Rule",
        "condition": {
          "equal": [
            {
              "divide": [
                {
                  "attr": "a"
                },
                {
                  "attr": "b"
                }
              ]
            },
            {
                "attr": "c"
            }
          ]
        },
        "effect": "Permit"
      }
    ]
  }
}
`,
		},
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
				expected:      pdp.EffectIndeterminateP,
				expectedError: "Float divisor has a value of 0",
			},
		},
	}

	validateTestSuite(ts, t)
}

func TestFloatIntegerEqual(t *testing.T) {
	ts := testSuite{
		policies: map[policyFormat]string{
			YAML: `# Policy set for Float Integer Equal Comparison
attributes:
  a: integer
  b: float

policies:
  alg: FirstApplicableEffect
  rules:
  - id: "Test Float Integer Equal"
    condition: # a == b
       equal:
       - attr: a
       - attr: b
    effect: Permit
`,
			JSON: `{
  "attributes": {
    "a": "integer",
    "b": "float"
  },
  "policies": {
    "id": "Test Float Integer Equal Policies",
    "alg": "FirstApplicableEffect",
    "rules": [
      {
        "id": "Test Float Integer Equal Rule",
        "condition": {
          "equal": [
            {
              "attr": "a"
            },
            {
              "attr": "b"
            }
          ]
        },
        "effect": "Permit"
      }
    ]
  }
}`,
		},
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
						Type:  "Integer",
						Value: "0",
					},
					{
						Id:    "b",
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
						Type:  "Integer",
						Value: "-1",
					},
					{
						Id:    "b",
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
						Type:  "Integer",
						Value: "1",
					},
					{
						Id:    "b",
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
						Type:  "Integer",
						Value: "0",
					},
					{
						Id:    "b",
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
						Type:  "Integer",
						Value: "1",
					},
					{
						Id:    "b",
						Type:  "Float",
						Value: "-2.",
					},
				},
				expected: pdp.EffectNotApplicable,
			},
		},
	}

	validateTestSuite(ts, t)
}

func TestFloatIntegerGreater(t *testing.T) {
	ts := testSuite{
		policies: map[policyFormat]string{
			YAML: `# Policy set for Float Integer Greater Comparison
attributes:
  a: integer
  b: float

policies:
  alg: FirstApplicableEffect
  rules:
  - id: "Test Float Integer Greater"
    condition: # a > b
      greater:
      - attr: a
      - attr: b
    effect: Permit
`,
			JSON: `{
  "attributes": {
    "a": "integer",
    "b": "float"
  },
  "policies": {
    "id": "Test Float Integer Greater Policies",
    "alg": "FirstApplicableEffect",
    "rules": [
      {
        "id": "Test Float Integer Greater Rule",
        "condition": {
          "greater": [
            {
              "attr": "a"
            },
            {
              "attr": "b"
            }
          ]
        },
        "effect": "Permit"
      }
    ]
  }
}`,
		},
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
						Type:  "Float",
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
						Type:  "Float",
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
						Type:  "Float",
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
						Type:  "Float",
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
						Type:  "Float",
						Value: "1",
					},
				},
				expected: pdp.EffectNotApplicable,
			},
		},
	}

	validateTestSuite(ts, t)
}

func TestFloatIntegerAdd(t *testing.T) {
	ts := testSuite{
		policies: map[policyFormat]string{
			YAML: `# Policy set for Float Integer Addition
attributes:
  a: integer
  b: float
  c: float

policies:
  alg: FirstApplicableEffect
  rules:
  - id: "Test Float Integer Addition"
    condition: # a + b == c
      equal:
      - add: # a + b
        - attr: a
        - attr: b
      - attr: c
    effect: Permit
`,
			JSON: `{
  "attributes": {
    "a": "integer",
    "b": "float",
    "c": "float"
  },
  "policies": {
    "id": "Test Float Integer Addition Policies",
    "alg": "FirstApplicableEffect",
    "rules": [
      {
        "id": "Test Float Integer Addition Rule",
        "condition": {
          "equal": [
            {
              "add": [
                {
                  "attr": "a"
                },
                {
                  "attr": "b"
                }
              ]
            },
            {
              "attr": "c"
            }
          ]
        },
        "effect": "Permit"
      }
    ]
  }
}`,
		},
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
						Type:  "float",
						Value: "1.3",
					},
					{
						Id:    "c",
						Type:  "float",
						Value: "2.3",
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
						Type:  "Integer",
						Value: "-1",
					},
					{
						Id:    "b",
						Type:  "float",
						Value: "-1.1",
					},
					{
						Id:    "c",
						Type:  "float",
						Value: "-2.1",
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
						Type:  "float",
						Value: "0.",
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
						Type:  "Integer",
						Value: "0",
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
						Type:  "Integer",
						Value: "1",
					},
					{
						Id:    "b",
						Type:  "float",
						Value: "-1.",
					},
					{
						Id:    "c",
						Type:  "float",
						Value: "2.",
					},
				},
				expected: pdp.EffectNotApplicable,
			},
		},
	}
	validateTestSuite(ts, t)
}

func TestFloatIntegerSubtract(t *testing.T) {
	ts := testSuite{
		policies: map[policyFormat]string{
			YAML: `# Policy set for Float Integer Subtraction
attributes:
  a: integer
  b: float
  c: float

policies:
  alg: FirstApplicableEffect
  rules:
  - id: "Test Float Integer Subtraction"
    condition: # a - b == c
      equal:
      - subtract:
        - attr: a
        - attr: b
      - attr: c
    effect: Permit
`,
			JSON: `{
  "attributes": {
    "a": "integer",
    "b": "float",
    "c": "float"
  },
  "policies": {
    "id": "Test Float Integer Subtraction Policies",
    "alg": "FirstApplicableEffect",
    "rules": [
      {
        "id": "Test Float Integer Subtraction Rule",
        "condition": {
          "equal": [
            {
              "subtract": [
                {
                  "attr": "a"
                },
                {
                  "attr": "b"
                }
              ]
            },
            {
              "attr": "c"
            }
          ]
        },
        "effect": "Permit"
      }
    ]
  }
}`,
		},
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
						Type:  "Integer",
						Value: "0",
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
						Type:  "Integer",
						Value: "-1",
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
						Type:  "Integer",
						Value: "1",
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
						Type:  "Integer",
						Value: "0",
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
						Type:  "Integer",
						Value: "1",
					},
					{
						Id:    "b",
						Type:  "Float",
						Value: "-1.",
					},
					{
						Id:    "c",
						Type:  "Float",
						Value: "0",
					},
				},
				expected: pdp.EffectNotApplicable,
			},
		},
	}

	validateTestSuite(ts, t)
}

func TestFloatIntegerMultiply(t *testing.T) {
	ts := testSuite{
		policies: map[policyFormat]string{
			YAML: `# Policy set for Float Integer Multiplication
attributes:
  a: integer
  b: float
  c: float

policies:
  alg: FirstApplicableEffect
  rules:
  - id: "Test Float Integer Multiplication"
    condition: # a * b == c
      equal:
      - multiply: # a * b
        - attr: a
        - attr: b
      - attr: c
    effect: Permit
`,
			JSON: `{
  "attributes": {
    "a": "integer",
    "b": "float",
    "c": "float"
  },
  "policies": {
    "id": "Test Float Integer Multiplication Policies",
    "alg": "FirstApplicableEffect",
    "rules": [
      {
        "id": "Test Float Integer Multiplication Rule",
        "condition": {
          "equal": [
            {
              "multiply": [
                {
                  "attr": "a"
                },
                {
                  "attr": "b"
                }
              ]
            },
            {
              "attr": "c"
            }
          ]
        },
        "effect": "Permit"
      }
    ]
  }
}`,
		},
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
						Type:  "Integer",
						Value: "1",
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
						Type:  "Integer",
						Value: "0",
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
						Type:  "Integer",
						Value: "-1",
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
						Type:  "Integer",
						Value: "-1",
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
						Type:  "Integer",
						Value: "1",
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
						Type:  "Integer",
						Value: "-1",
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
						Type:  "Integer",
						Value: "0",
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

func TestFloatIntegerDivide(t *testing.T) {
	ts := testSuite{
		policies: map[policyFormat]string{
			YAML: `# Policy set for Float Integer Division
attributes:
  a: integer
  b: float
  c: float

policies:
  alg: FirstApplicableEffect
  rules:
  - id: "Test Float Integer Division"
    condition: # a / b == c
      equal:
      - divide: # a / b
        - attr: a
        - attr: b
      - attr: c
    effect: Permit
`,
			JSON: `{
  "attributes": {
    "a": "integer",
    "b": "float",
    "c": "float"
  },
  "policies": {
    "id": "Test Float Integer Division Policies",
    "alg": "FirstApplicableEffect",
    "rules": [
      {
        "id": "Test Float Integer Division Rule",
        "condition": {
          "equal": [
            {
              "divide": [
                {
                  "attr": "a"
                },
                {
                  "attr": "b"
                }
              ]
            },
            {
              "attr": "c"
            }
          ]
        },
        "effect": "Permit"
      }
    ]
  }
}`,
		},
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
						Type:  "Integer",
						Value: "0",
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
						Type:  "Integer",
						Value: "4",
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
						Type:  "Integer",
						Value: "7",
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
						Type:  "Integer",
						Value: "-1",
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
						Type:  "Integer",
						Value: "1",
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
						Type:  "Integer",
						Value: "2",
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
						Type:  "Integer",
						Value: "-1",
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
						Type:  "Integer",
						Value: "0",
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

func TestFloatRange(t *testing.T) {
	ts := testSuite{
		policies: map[policyFormat]string{
			YAML: `# Policy set for Float Range
attributes:
  a: float
  b: float
  c: float
  r: string

policies:
  alg:
    id: Mapper
    map:
      range:
        - attr: a
        - attr: b
        - attr: c
    alg: FirstApplicableEffect
  rules:
  - id: Below
    effect: Permit
    obligations:
    - r:
       val:
         type: string
         content: Below

  - id: Above
    effect: Permit
    obligations:
    - r:
       val:
         type: string
         content: Above

  - id: Within
    effect: Permit
    obligations:
    - r:
       val:
         type: string
         content: Within
`,
			JSON: `{
  "attributes": {
    "a": "float",
    "b": "float",
    "c": "float",
    "r": "string"
  },
  "policies": {
    "id": "Test Float Range Policies",
    "alg": {
       "id": "mapper",
       "map": {
          "range": [
            {
               "attr": "a"
            },
            {
               "attr": "b"
            },
            {
               "attr": "c"
            }
          ]
        }
     },
    "rules": [
      {
        "id": "Below",
        "effect": "Permit",
        "obligations": [
           {
              "r": {
                 "val": {
                     "type": "string",
                     "content": "Below"
                 }
              }
           }
        ]
      },
      {
        "id": "Above",
        "effect": "Permit",
        "obligations": [
           {
              "r": {
                 "val": {
                     "type": "string",
                     "content": "Above"
                 }
              }
           }
        ]
      },
      {
        "id": "Within",
        "effect": "Permit",
        "obligations": [
           {
              "r": {
                 "val": {
                     "type": "string",
                     "content": "Within"
                 }
              }
           }
        ]
      }
    ]
  }
}`,
		},
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
						Value: "5.",
					},
					{
						Id:    "c",
						Type:  "Float",
						Value: "0.",
					},
				},
				expected:           pdp.EffectPermit,
				expectedObligation: "Below",
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
						Value: "5.",
					},
					{
						Id:    "c",
						Type:  "Float",
						Value: "10.",
					},
				},
				expected:           pdp.EffectPermit,
				expectedObligation: "Above",
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
						Value: "5.",
					},
					{
						Id:    "c",
						Type:  "Float",
						Value: "3.3",
					},
				},
				expected:           pdp.EffectPermit,
				expectedObligation: "Within",
			},
		},
	}

	validateTestSuite(ts, t)
}
