package integrationTests

import (
	"testing"

	"github.com/infobloxopen/go-trees/strtree"
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

func TestSetOfStringsLen(t *testing.T) {
	ts := testSuite{
		policies: map[policyFormat]string{
			YAML: `# Policy set for Set of Strings Len
attributes:
  a: set of strings
  b: integer

policies:
  id: Test Set of Strings Len
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
    "a": "set of strings",
    "b": "integer"
  },
  "policies": {
    "id": "Test Set of Strings Len",
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
					pdp.MakeSetOfStringsAssignment("a", newStrTree()),
					pdp.MakeIntegerAssignment("b", 0),
				},
				expected: pdp.EffectPermit,
			},
			{
				attrs: []pdp.AttributeAssignment{
					pdp.MakeSetOfStringsAssignment("a", newStrTree("0")),
					pdp.MakeIntegerAssignment("b", 1),
				},
				expected: pdp.EffectPermit,
			},
			{
				attrs: []pdp.AttributeAssignment{
					pdp.MakeSetOfStringsAssignment("a", newStrTree("0", "1", "2")),
					pdp.MakeIntegerAssignment("b", 3),
				},
				expected: pdp.EffectPermit,
			},
		},
	}

	validateTestSuite(ts, t)
}

func TestSetOfStringsContainsIntersect(t *testing.T) {
	ts := testSuite{
		policies: map[policyFormat]string{
			YAML: `# Policy set for Set of Strings Contains and Intersect
attributes:
  a: set of strings
  b: set of strings
  c: string

policies:
  id: Test Set of Strings Contains Intersect
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
    "a": "set of strings",
    "b": "set of strings",
    "c": "string"
  },
  "policies": {
    "id": "Test Set of Strings Contains Intersect",
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
					pdp.MakeSetOfStringsAssignment("a", newStrTree("foo", "bar")),
					pdp.MakeSetOfStringsAssignment("b", newStrTree("boo", "mar", "foo")),
					pdp.MakeStringAssignment("c", "foo"),
				},
				expected: pdp.EffectPermit,
			},
			{
				attrs: []pdp.AttributeAssignment{
					pdp.MakeSetOfStringsAssignment("a", newStrTree("foo", "bar")),
					pdp.MakeSetOfStringsAssignment("b", newStrTree("boo", "mar", "foo")),
					pdp.MakeStringAssignment("c", "boo"),
				},
				expected: pdp.EffectNotApplicable,
			},
			{
				attrs: []pdp.AttributeAssignment{
					pdp.MakeSetOfStringsAssignment("a", newStrTree("a", "e", "f", "b", "c", "g", "d")),
					pdp.MakeSetOfStringsAssignment("b", newStrTree("z", "s", "d", "p", "x", "i", "aa")),
					pdp.MakeStringAssignment("c", "d"),
				},
				expected: pdp.EffectPermit,
			},
		},
	}

	validateTestSuite(ts, t)
}

func TestSetOfStringsLenIntersect(t *testing.T) {
	ts := testSuite{
		policies: map[policyFormat]string{
			YAML: `# Policy set for Set of Strings Len and Intersect
attributes:
  a: set of strings
  b: set of strings
  c: integer

policies:
  id: Test Set of Strings Len Intersect
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
    "a": "set of strings",
    "b": "set of strings",
    "c": "integer"
  },
  "policies": {
    "id": "Test Set of Strings Len Intersect",
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
					pdp.MakeSetOfStringsAssignment("a", newStrTree("foo", "bar", "doo")),
					pdp.MakeSetOfStringsAssignment("b", newStrTree("boo", "mar", "aoo")),
					pdp.MakeIntegerAssignment("c", 0),
				},
				expected: pdp.EffectPermit,
			},
			{
				attrs: []pdp.AttributeAssignment{
					pdp.MakeSetOfStringsAssignment("a", newStrTree("foo", "bar")),
					pdp.MakeSetOfStringsAssignment("b", newStrTree("boo", "mar", "foo")),
					pdp.MakeIntegerAssignment("c", 1),
				},
				expected: pdp.EffectPermit,
			},
			{
				attrs: []pdp.AttributeAssignment{
					pdp.MakeSetOfStringsAssignment("a", newStrTree("foo", "bar", "boo")),
					pdp.MakeSetOfStringsAssignment("b", newStrTree("boo", "mar", "foo")),
					pdp.MakeIntegerAssignment("c", 2),
				},
				expected: pdp.EffectPermit,
			},
		},
	}

	validateTestSuite(ts, t)
}

// TODO decide access level and what to do with value_test.go's values
func newStrTree(args ...string) *strtree.Tree {
	t := strtree.NewTree()
	for i, s := range args {
		t.InplaceInsert(s, i)
	}

	return t
}
