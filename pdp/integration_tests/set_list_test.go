package integrationTests

import (
	"testing"

	"github.com/infobloxopen/themis/pdp"
)

func TestListOfStringsContains(t *testing.T) {
	ts := testSuite{
		policies: map[policyFormat]string{
			YAML: `# Policy set for List of Strings Contains
attributes:
  a: list of strings
  b: string

policies:
  id: Test List of Strings Contains
  alg: FirstApplicableEffect
  rules:
  - effect: Permit
    condition:
      contains:
      - attr: a
      - attr: b
`,
			JSON: `{
  "attributes": {
    "a": "list of strings",
    "b": "string"
  },
  "policies": {
    "id": "Test List of Strings Contains",
    "alg": "FirstApplicableEffect",
    "rules": [
      {
        "effect": "Permit",
        "condition": {
          "contains": [
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
}`,
		},
		testSet: []testCase{
			{
				attrs: []pdp.AttributeAssignment{
					pdp.MakeListOfStringsAssignment("a", []string{"foo", "bar"}),
					pdp.MakeStringAssignment("b", "foo"),
				},
				expected: pdp.EffectPermit,
			},
			{
				attrs: []pdp.AttributeAssignment{
					pdp.MakeListOfStringsAssignment("a", []string{"foo", "bar"}),
					pdp.MakeStringAssignment("b", "boo"),
				},
				expected: pdp.EffectNotApplicable,
			},
		},
	}

	validateTestSuite(ts, t)
}

func TestListOfStringsLen(t *testing.T) {
	ts := testSuite{
		policies: map[policyFormat]string{
			YAML: `# Policy set for List of Strings Len
attributes:
  a: list of strings
  b: integer

policies:
  id: Test List of Strings Len
  alg: FirstApplicableEffect
  rules:
  - effect: Permit
    condition:
      equal:
      - len:
        - attr: a
      - attr: b
`,
			JSON: `{
  "attributes": {
    "a": "list of strings",
    "b": "integer"
  },
  "policies": {
    "id": "Test List of Strings Len",
    "alg": "FirstApplicableEffect",
    "rules": [
      {
        "effect": "Permit",
        "condition": {
          "equal": [
            {
              "len": [
                {
                  "attr": "a"
                }
              ]
            },
            {
              "attr": "b"
            }
          ]
        }
      }
    ]
  }
}`,
		},
		testSet: []testCase{
			{
				attrs: []pdp.AttributeAssignment{
					pdp.MakeListOfStringsAssignment("a", []string{}),
					pdp.MakeIntegerAssignment("b", 0),
				},
				expected: pdp.EffectPermit,
			},
			{
				attrs: []pdp.AttributeAssignment{
					pdp.MakeListOfStringsAssignment("a", []string{"0"}),
					pdp.MakeIntegerAssignment("b", 1),
				},
				expected: pdp.EffectPermit,
			},
			{
				attrs: []pdp.AttributeAssignment{
					pdp.MakeListOfStringsAssignment("a", []string{"0", "1"}),
					pdp.MakeIntegerAssignment("b", 2),
				},
				expected: pdp.EffectPermit,
			},
		},
	}

	validateTestSuite(ts, t)
}

func TestListOfStringsContainsIntersect(t *testing.T) {
	ts := testSuite{
		policies: map[policyFormat]string{
			YAML: `# Policy set for List of Strings Contains and Intersect
attributes:
  a: list of strings
  b: list of strings
  c: string

policies:
  id: Test List of Strings Contains Intersect
  alg: FirstApplicableEffect
  rules:
  - effect: Permit
    condition:
      contains:
      - intersect:
        - attr: a
        - attr: b
      - attr: c
`,
			JSON: `{
  "attributes": {
    "a": "list of strings",
    "b": "list of strings",
    "c": "string"
  },
  "policies": {
    "id": "Test List of Strings Contains Intersect",
    "alg": "FirstApplicableEffect",
    "rules": [
      {
        "effect": "Permit",
        "condition": {
          "contains": [
            {
              "intersect": [
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
        }
      }
    ]
  }
}`,
		},
		testSet: []testCase{
			{
				attrs: []pdp.AttributeAssignment{
					pdp.MakeListOfStringsAssignment("a", []string{"foo", "bar"}),
					pdp.MakeListOfStringsAssignment("b", []string{"boo", "mar", "foo"}),
					pdp.MakeStringAssignment("c", "foo"),
				},
				expected: pdp.EffectPermit,
			},
			{
				attrs: []pdp.AttributeAssignment{
					pdp.MakeListOfStringsAssignment("a", []string{"foo", "bar"}),
					pdp.MakeListOfStringsAssignment("b", []string{"boo", "mar", "foo"}),
					pdp.MakeStringAssignment("c", "boo"),
				},
				expected: pdp.EffectNotApplicable,
			},
		},
	}

	validateTestSuite(ts, t)
}

func TestListOfStringsLenIntersect(t *testing.T) {
	ts := testSuite{
		policies: map[policyFormat]string{
			YAML: `# Policy set for List of Strings Len and Intersect
attributes:
  a: list of strings
  b: list of strings
  c: integer

policies:
  id: Test List of Strings Len Intersect
  alg: FirstApplicableEffect
  rules:
  - effect: Permit
    condition:
      equal:
      - len:
        - intersect:
          - attr: a
          - attr: b
      - attr: c
`,
			JSON: `{
  "attributes": {
    "a": "list of strings",
    "b": "list of strings",
    "c": "integer"
  },
  "policies": {
    "id": "Test List of Strings Len Intersect",
    "alg": "FirstApplicableEffect",
    "rules": [
      {
        "effect": "Permit",
        "condition": {
          "equal": [
            {
              "len": [
                {
                  "intersect": [
                    {
                      "attr": "a"
                    },
                    {
                      "attr": "b"
                    }
                  ]
                }
              ]
            },
            {
              "attr": "c"
            }
          ]
        }
      }
    ]
  }
}`,
		},
		testSet: []testCase{
			{
				attrs: []pdp.AttributeAssignment{
					pdp.MakeListOfStringsAssignment("a", []string{"foo", "bar", "doo"}),
					pdp.MakeListOfStringsAssignment("b", []string{"boo", "mar", "aoo"}),
					pdp.MakeIntegerAssignment("c", 0),
				},
				expected: pdp.EffectPermit,
			},
			{
				attrs: []pdp.AttributeAssignment{
					pdp.MakeListOfStringsAssignment("a", []string{"foo", "bar"}),
					pdp.MakeListOfStringsAssignment("b", []string{"boo", "mar", "foo"}),
					pdp.MakeIntegerAssignment("c", 1),
				},
				expected: pdp.EffectPermit,
			},
			{
				attrs: []pdp.AttributeAssignment{
					pdp.MakeListOfStringsAssignment("a", []string{"foo", "bar", "boo"}),
					pdp.MakeListOfStringsAssignment("b", []string{"boo", "mar", "foo"}),
					pdp.MakeIntegerAssignment("c", 2),
				},
				expected: pdp.EffectPermit,
			},
		},
	}

	validateTestSuite(ts, t)
}
