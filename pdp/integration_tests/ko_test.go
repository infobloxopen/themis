package integrationTests

import (
	"testing"

	"github.com/infobloxopen/themis/pdp"
	pb "github.com/infobloxopen/themis/pdp-service"
)

func TestKo(t *testing.T) {
	ts := testSuite{
		policies: map[policyFormat]string{
			YAML: `# Policy set for Integer Equal Comparison
attributes:
  a: integer
  b: integer
  r: string

policies:
  alg: FirstApplicableEffect
  rules:
  - id: "Test Integer Equal"
    condition: # a == b
       equal:
       - attr: a
       - attr: b
    effect: Permit
    obligations:
      - r: "All Good"
`,
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

func TestKo2(t *testing.T) {
	ts := testSuite{
		policies: map[policyFormat]string{
			YAML: `# Policy set for Integer Equal Comparison
attributes:
  a: integer
  b: integer
  r: integer

policies:
  alg: FirstApplicableEffect
  rules:
  - id: "Test Integer Equal"
    condition: # a == b
       equal:
       - attr: a
       - attr: b
    effect: Permit
    obligations:
      - r:
         add:
           - attr: a
           - attr: b
`,
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

func TestKo3(t *testing.T) {
	ts := testSuite{
		policies: map[policyFormat]string{
			JSON: `{
  "attributes": {
    "a": "integer",
    "b": "float",
    "c": "float",
    "r": "float"
  },
  "policies": {
    "id": "Test Ko 3 Policies",
    "alg": "FirstApplicableEffect",
    "rules": [
      {
        "id": "Test Ko 3 Rule",
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
        "effect": "Permit",
        "Obligations": [
           {
               "r": {
                  "add": [
                     {
                       "attr": "a"
                     },
                     {
                       "attr": "b"
                     }
                  ]
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
				expected:           pdp.EffectPermit,
				expectedObligation: "2",
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
