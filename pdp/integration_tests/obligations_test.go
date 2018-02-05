package integrationTests

import (
	"testing"

	"github.com/infobloxopen/themis/pdp"
	pb "github.com/infobloxopen/themis/pdp-service"
)

func TestExpression(t *testing.T) {
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
      - r:
         val:
           type: string
           content: "All Good"
`,
			JSON: `{
  "attributes": {
    "a": "integer",
    "b": "integer",
    "r": "string"
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
        "effect": "Permit",
        "obligations": [
          {
            "r": {
              "val": {
                "type": "string",
                "content": "All Good"
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
						Type:  "Integer",
						Value: "1",
					},
					{
						Id:    "b",
						Type:  "Integer",
						Value: "1",
					},
				},
				expected:           pdp.EffectPermit,
				expectedObligation: "All Good",
			},
		},
	}

	validateTestSuite(ts, t)
}
