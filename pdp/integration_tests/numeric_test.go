package integrationTests

import (
	"testing"

	"github.com/infobloxopen/themis/pdp"
)

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
				attrs: []pdp.AttributeAssignment{
					pdp.MakeIntegerAssignment("a", 1),
					pdp.MakeIntegerAssignment("b", 1),
				},
				expected: pdp.EffectPermit,
			},
			{
				attrs: []pdp.AttributeAssignment{
					pdp.MakeIntegerAssignment("a", 0),
					pdp.MakeIntegerAssignment("b", 0),
				},
				expected: pdp.EffectPermit,
			},
			{
				attrs: []pdp.AttributeAssignment{
					pdp.MakeIntegerAssignment("a", -1),
					pdp.MakeIntegerAssignment("b", -1),
				},
				expected: pdp.EffectPermit,
			},
			{
				attrs: []pdp.AttributeAssignment{
					pdp.MakeIntegerAssignment("a", 1),
					pdp.MakeIntegerAssignment("b", 0),
				},
				expected: pdp.EffectNotApplicable,
			},
			{
				attrs: []pdp.AttributeAssignment{
					pdp.MakeIntegerAssignment("a", 0),
					pdp.MakeIntegerAssignment("b", 1),
				},
				expected: pdp.EffectNotApplicable,
			},
			{
				attrs: []pdp.AttributeAssignment{
					pdp.MakeIntegerAssignment("a", 1),
					pdp.MakeIntegerAssignment("b", -2),
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
				attrs: []pdp.AttributeAssignment{
					pdp.MakeIntegerAssignment("a", 1),
					pdp.MakeIntegerAssignment("b", 0),
				},
				expected: pdp.EffectPermit,
			},
			{
				attrs: []pdp.AttributeAssignment{
					pdp.MakeIntegerAssignment("a", 0),
					pdp.MakeIntegerAssignment("b", -1),
				},
				expected: pdp.EffectPermit,
			},
			{
				attrs: []pdp.AttributeAssignment{
					pdp.MakeIntegerAssignment("a", 1),
					pdp.MakeIntegerAssignment("b", -1),
				},
				expected: pdp.EffectPermit,
			},
			{
				attrs: []pdp.AttributeAssignment{
					pdp.MakeIntegerAssignment("a", 0),
					pdp.MakeIntegerAssignment("b", 0),
				},
				expected: pdp.EffectNotApplicable,
			},
			{
				attrs: []pdp.AttributeAssignment{
					pdp.MakeIntegerAssignment("a", -1),
					pdp.MakeIntegerAssignment("b", 1),
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
				attrs: []pdp.AttributeAssignment{
					pdp.MakeIntegerAssignment("a", 1),
					pdp.MakeIntegerAssignment("b", 1),
					pdp.MakeIntegerAssignment("c", 2),
				},
				expected: pdp.EffectPermit,
			},
			{
				attrs: []pdp.AttributeAssignment{
					pdp.MakeIntegerAssignment("a", 0),
					pdp.MakeIntegerAssignment("b", 0),
					pdp.MakeIntegerAssignment("c", 0),
				},
				expected: pdp.EffectPermit,
			},
			{
				attrs: []pdp.AttributeAssignment{
					pdp.MakeIntegerAssignment("a", -1),
					pdp.MakeIntegerAssignment("b", -1),
					pdp.MakeIntegerAssignment("c", -2),
				},
				expected: pdp.EffectPermit,
			},
			{
				attrs: []pdp.AttributeAssignment{
					pdp.MakeIntegerAssignment("a", 1),
					pdp.MakeIntegerAssignment("b", 0),
					pdp.MakeIntegerAssignment("c", 2),
				},
				expected: pdp.EffectNotApplicable,
			},
			{
				attrs: []pdp.AttributeAssignment{
					pdp.MakeIntegerAssignment("a", 0),
					pdp.MakeIntegerAssignment("b", 1),
					pdp.MakeIntegerAssignment("c", 2),
				},
				expected: pdp.EffectNotApplicable,
			},
			{
				attrs: []pdp.AttributeAssignment{
					pdp.MakeIntegerAssignment("a", 1),
					pdp.MakeIntegerAssignment("b", -1),
					pdp.MakeIntegerAssignment("c", 2),
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
				attrs: []pdp.AttributeAssignment{
					pdp.MakeIntegerAssignment("a", 1),
					pdp.MakeIntegerAssignment("b", 1),
					pdp.MakeIntegerAssignment("c", 0),
				},
				expected: pdp.EffectPermit,
			},
			{
				attrs: []pdp.AttributeAssignment{
					pdp.MakeIntegerAssignment("a", 0),
					pdp.MakeIntegerAssignment("b", 0),
					pdp.MakeIntegerAssignment("c", 0),
				},
				expected: pdp.EffectPermit,
			},
			{
				attrs: []pdp.AttributeAssignment{
					pdp.MakeIntegerAssignment("a", -1),
					pdp.MakeIntegerAssignment("b", -1),
					pdp.MakeIntegerAssignment("c", 0),
				},
				expected: pdp.EffectPermit,
			},
			{
				attrs: []pdp.AttributeAssignment{
					pdp.MakeIntegerAssignment("a", 1),
					pdp.MakeIntegerAssignment("b", 0),
					pdp.MakeIntegerAssignment("c", 0),
				},
				expected: pdp.EffectNotApplicable,
			},
			{
				attrs: []pdp.AttributeAssignment{
					pdp.MakeIntegerAssignment("a", 0),
					pdp.MakeIntegerAssignment("b", 1),
					pdp.MakeIntegerAssignment("c", 0),
				},
				expected: pdp.EffectNotApplicable,
			},
			{
				attrs: []pdp.AttributeAssignment{
					pdp.MakeIntegerAssignment("a", 1),
					pdp.MakeIntegerAssignment("b", -1),
					pdp.MakeIntegerAssignment("c", 0),
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
				attrs: []pdp.AttributeAssignment{
					pdp.MakeIntegerAssignment("a", 1),
					pdp.MakeIntegerAssignment("b", 1),
					pdp.MakeIntegerAssignment("c", 1),
				},
				expected: pdp.EffectPermit,
			},
			{
				attrs: []pdp.AttributeAssignment{
					pdp.MakeIntegerAssignment("a", 1),
					pdp.MakeIntegerAssignment("b", 0),
					pdp.MakeIntegerAssignment("c", 0),
				},
				expected: pdp.EffectPermit,
			},
			{
				attrs: []pdp.AttributeAssignment{
					pdp.MakeIntegerAssignment("a", 0),
					pdp.MakeIntegerAssignment("b", 1),
					pdp.MakeIntegerAssignment("c", 0),
				},
				expected: pdp.EffectPermit,
			},
			{
				attrs: []pdp.AttributeAssignment{
					pdp.MakeIntegerAssignment("a", -1),
					pdp.MakeIntegerAssignment("b", -1),
					pdp.MakeIntegerAssignment("c", 1),
				},
				expected: pdp.EffectPermit,
			},
			{
				attrs: []pdp.AttributeAssignment{
					pdp.MakeIntegerAssignment("a", -1),
					pdp.MakeIntegerAssignment("b", 1),
					pdp.MakeIntegerAssignment("c", -1),
				},
				expected: pdp.EffectPermit,
			},
			{
				attrs: []pdp.AttributeAssignment{
					pdp.MakeIntegerAssignment("a", 1),
					pdp.MakeIntegerAssignment("b", 0),
					pdp.MakeIntegerAssignment("c", 1),
				},
				expected: pdp.EffectNotApplicable,
			},
			{
				attrs: []pdp.AttributeAssignment{
					pdp.MakeIntegerAssignment("a", -1),
					pdp.MakeIntegerAssignment("b", 1),
					pdp.MakeIntegerAssignment("c", 1),
				},
				expected: pdp.EffectNotApplicable,
			},
			{
				attrs: []pdp.AttributeAssignment{
					pdp.MakeIntegerAssignment("a", 0),
					pdp.MakeIntegerAssignment("b", 1),
					pdp.MakeIntegerAssignment("c", 1),
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
				attrs: []pdp.AttributeAssignment{
					pdp.MakeIntegerAssignment("a", 1),
					pdp.MakeIntegerAssignment("b", 1),
					pdp.MakeIntegerAssignment("c", 1),
				},
				expected: pdp.EffectPermit,
			},
			{
				attrs: []pdp.AttributeAssignment{
					pdp.MakeIntegerAssignment("a", 0),
					pdp.MakeIntegerAssignment("b", 1),
					pdp.MakeIntegerAssignment("c", 0),
				},
				expected: pdp.EffectPermit,
			},
			{
				attrs: []pdp.AttributeAssignment{
					pdp.MakeIntegerAssignment("a", 4),
					pdp.MakeIntegerAssignment("b", 2),
					pdp.MakeIntegerAssignment("c", 2),
				},
				expected: pdp.EffectPermit,
			},
			{
				attrs: []pdp.AttributeAssignment{
					pdp.MakeIntegerAssignment("a", 7),
					pdp.MakeIntegerAssignment("b", 2),
					pdp.MakeIntegerAssignment("c", 3),
				},
				expected: pdp.EffectPermit,
			},
			{
				attrs: []pdp.AttributeAssignment{
					pdp.MakeIntegerAssignment("a", -1),
					pdp.MakeIntegerAssignment("b", 1),
					pdp.MakeIntegerAssignment("c", -1),
				},
				expected: pdp.EffectPermit,
			},
			{
				attrs: []pdp.AttributeAssignment{
					pdp.MakeIntegerAssignment("a", 1),
					pdp.MakeIntegerAssignment("b", -1),
					pdp.MakeIntegerAssignment("c", -1),
				},
				expected: pdp.EffectPermit,
			},
			{
				attrs: []pdp.AttributeAssignment{
					pdp.MakeIntegerAssignment("a", 2),
					pdp.MakeIntegerAssignment("b", 1),
					pdp.MakeIntegerAssignment("c", 1),
				},
				expected: pdp.EffectNotApplicable,
			},
			{
				attrs: []pdp.AttributeAssignment{
					pdp.MakeIntegerAssignment("a", -1),
					pdp.MakeIntegerAssignment("b", 1),
					pdp.MakeIntegerAssignment("c", 1),
				},
				expected: pdp.EffectNotApplicable,
			},
			{
				attrs: []pdp.AttributeAssignment{
					pdp.MakeIntegerAssignment("a", 0),
					pdp.MakeIntegerAssignment("b", 1),
					pdp.MakeIntegerAssignment("c", 1),
				},
				expected: pdp.EffectNotApplicable,
			},
			{
				attrs: []pdp.AttributeAssignment{
					pdp.MakeIntegerAssignment("a", 1),
					pdp.MakeIntegerAssignment("b", 0),
					pdp.MakeIntegerAssignment("c", 1),
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
				attrs: []pdp.AttributeAssignment{
					pdp.MakeFloatAssignment("a", 1.0),
					pdp.MakeFloatAssignment("b", 0.9),
				},
				expected: pdp.EffectPermit,
			},
			{
				attrs: []pdp.AttributeAssignment{
					pdp.MakeFloatAssignment("a", 1.0),
					pdp.MakeFloatAssignment("b", -1.0),
				},
				expected: pdp.EffectPermit,
			},
			{
				attrs: []pdp.AttributeAssignment{
					pdp.MakeFloatAssignment("a", 0.0),
					pdp.MakeFloatAssignment("b", -1.0),
				},
				expected: pdp.EffectPermit,
			},
			{
				attrs: []pdp.AttributeAssignment{
					pdp.MakeFloatAssignment("a", 0.8),
					pdp.MakeFloatAssignment("b", 0.9),
				},
				expected: pdp.EffectNotApplicable,
			},
			{
				attrs: []pdp.AttributeAssignment{
					pdp.MakeFloatAssignment("a", -2.0),
					pdp.MakeFloatAssignment("b", 1.0),
				},
				expected: pdp.EffectNotApplicable,
			},
			{
				attrs: []pdp.AttributeAssignment{
					pdp.MakeFloatAssignment("a", -1.0),
					pdp.MakeFloatAssignment("b", 0.0),
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
				attrs: []pdp.AttributeAssignment{
					pdp.MakeFloatAssignment("a", 1.0),
					pdp.MakeFloatAssignment("b", 1.0),
					pdp.MakeFloatAssignment("c", 2.0),
				},
				expected: pdp.EffectPermit,
			},
			{
				attrs: []pdp.AttributeAssignment{
					pdp.MakeFloatAssignment("a", 0.0),
					pdp.MakeFloatAssignment("b", 0.0),
					pdp.MakeFloatAssignment("c", 0.0),
				},
				expected: pdp.EffectPermit,
			},
			{
				attrs: []pdp.AttributeAssignment{
					pdp.MakeFloatAssignment("a", -1.0),
					pdp.MakeFloatAssignment("b", -1.0),
					pdp.MakeFloatAssignment("c", -2.0),
				},
				expected: pdp.EffectPermit,
			},
			{
				attrs: []pdp.AttributeAssignment{
					pdp.MakeFloatAssignment("a", 1.0),
					pdp.MakeFloatAssignment("b", 0.0),
					pdp.MakeFloatAssignment("c", 0.0),
				},
				expected: pdp.EffectNotApplicable,
			},
			{
				attrs: []pdp.AttributeAssignment{
					pdp.MakeFloatAssignment("a", 0.0),
					pdp.MakeFloatAssignment("b", 1.0),
					pdp.MakeFloatAssignment("c", 2.0),
				},
				expected: pdp.EffectNotApplicable,
			},
			{
				attrs: []pdp.AttributeAssignment{
					pdp.MakeFloatAssignment("a", 1.0),
					pdp.MakeFloatAssignment("b", -1.0),
					pdp.MakeFloatAssignment("c", 1.0),
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
				attrs: []pdp.AttributeAssignment{
					pdp.MakeFloatAssignment("a", 1.0),
					pdp.MakeFloatAssignment("b", 1.0),
					pdp.MakeFloatAssignment("c", 0.0),
				},
				expected: pdp.EffectPermit,
			},
			{
				attrs: []pdp.AttributeAssignment{
					pdp.MakeFloatAssignment("a", 0.0),
					pdp.MakeFloatAssignment("b", 0.0),
					pdp.MakeFloatAssignment("c", 0.0),
				},
				expected: pdp.EffectPermit,
			},
			{
				attrs: []pdp.AttributeAssignment{
					pdp.MakeFloatAssignment("a", -1.0),
					pdp.MakeFloatAssignment("b", -1.0),
					pdp.MakeFloatAssignment("c", 0.0),
				},
				expected: pdp.EffectPermit,
			},
			{
				attrs: []pdp.AttributeAssignment{
					pdp.MakeFloatAssignment("a", 1.0),
					pdp.MakeFloatAssignment("b", 0.0),
					pdp.MakeFloatAssignment("c", 0.0),
				},
				expected: pdp.EffectNotApplicable,
			},
			{
				attrs: []pdp.AttributeAssignment{
					pdp.MakeFloatAssignment("a", 0.0),
					pdp.MakeFloatAssignment("b", 1.0),
					pdp.MakeFloatAssignment("c", 0.0),
				},
				expected: pdp.EffectNotApplicable,
			},
			{
				attrs: []pdp.AttributeAssignment{
					pdp.MakeFloatAssignment("a", 1.0),
					pdp.MakeFloatAssignment("b", -1.0),
					pdp.MakeFloatAssignment("c", 0.0),
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
				attrs: []pdp.AttributeAssignment{
					pdp.MakeFloatAssignment("a", 1.0),
					pdp.MakeFloatAssignment("b", 1.0),
					pdp.MakeFloatAssignment("c", 1.0),
				},
				expected: pdp.EffectPermit,
			},
			{
				attrs: []pdp.AttributeAssignment{
					pdp.MakeFloatAssignment("a", 1.0),
					pdp.MakeFloatAssignment("b", 0.0),
					pdp.MakeFloatAssignment("c", 0.0),
				},
				expected: pdp.EffectPermit,
			},
			{
				attrs: []pdp.AttributeAssignment{
					pdp.MakeFloatAssignment("a", 0.0),
					pdp.MakeFloatAssignment("b", 1.0),
					pdp.MakeFloatAssignment("c", 0.0),
				},
				expected: pdp.EffectPermit,
			},
			{
				attrs: []pdp.AttributeAssignment{
					pdp.MakeFloatAssignment("a", -1.0),
					pdp.MakeFloatAssignment("b", -1.0),
					pdp.MakeFloatAssignment("c", 1.0),
				},
				expected: pdp.EffectPermit,
			},
			{
				attrs: []pdp.AttributeAssignment{
					pdp.MakeFloatAssignment("a", -1.0),
					pdp.MakeFloatAssignment("b", 1.0),
					pdp.MakeFloatAssignment("c", -1.0),
				},
				expected: pdp.EffectPermit,
			},
			{
				attrs: []pdp.AttributeAssignment{
					pdp.MakeFloatAssignment("a", 1.0),
					pdp.MakeFloatAssignment("b", 0.0),
					pdp.MakeFloatAssignment("c", 1.0),
				},
				expected: pdp.EffectNotApplicable,
			},
			{
				attrs: []pdp.AttributeAssignment{
					pdp.MakeFloatAssignment("a", -1.0),
					pdp.MakeFloatAssignment("b", 1.0),
					pdp.MakeFloatAssignment("c", 1.0),
				},
				expected: pdp.EffectNotApplicable,
			},
			{
				attrs: []pdp.AttributeAssignment{
					pdp.MakeFloatAssignment("a", 1.9e+200),
					pdp.MakeFloatAssignment("b", 1.9e+233),
					pdp.MakeFloatAssignment("c", 1.0),
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
				attrs: []pdp.AttributeAssignment{
					pdp.MakeFloatAssignment("a", 1.0),
					pdp.MakeFloatAssignment("b", 1.0),
					pdp.MakeFloatAssignment("c", 1.0),
				},
				expected: pdp.EffectPermit,
			},
			{
				attrs: []pdp.AttributeAssignment{
					pdp.MakeFloatAssignment("a", 0.0),
					pdp.MakeFloatAssignment("b", 1.0),
					pdp.MakeFloatAssignment("c", 0.0),
				},
				expected: pdp.EffectPermit,
			},
			{
				attrs: []pdp.AttributeAssignment{
					pdp.MakeFloatAssignment("a", 4.0),
					pdp.MakeFloatAssignment("b", 2.0),
					pdp.MakeFloatAssignment("c", 2.0),
				},
				expected: pdp.EffectPermit,
			},
			{
				attrs: []pdp.AttributeAssignment{
					pdp.MakeFloatAssignment("a", 7.0),
					pdp.MakeFloatAssignment("b", 2.0),
					pdp.MakeFloatAssignment("c", 3.5),
				},
				expected: pdp.EffectPermit,
			},
			{
				attrs: []pdp.AttributeAssignment{
					pdp.MakeFloatAssignment("a", -1.0),
					pdp.MakeFloatAssignment("b", 1.0),
					pdp.MakeFloatAssignment("c", -1.0),
				},
				expected: pdp.EffectPermit,
			},
			{
				attrs: []pdp.AttributeAssignment{
					pdp.MakeFloatAssignment("a", 1.0),
					pdp.MakeFloatAssignment("b", -1.0),
					pdp.MakeFloatAssignment("c", -1.0),
				},
				expected: pdp.EffectPermit,
			},
			{
				attrs: []pdp.AttributeAssignment{
					pdp.MakeFloatAssignment("a", 2.0),
					pdp.MakeFloatAssignment("b", 1.0),
					pdp.MakeFloatAssignment("c", 1.0),
				},
				expected: pdp.EffectNotApplicable,
			},
			{
				attrs: []pdp.AttributeAssignment{
					pdp.MakeFloatAssignment("a", -1.0),
					pdp.MakeFloatAssignment("b", 1.0),
					pdp.MakeFloatAssignment("c", 1.0),
				},
				expected: pdp.EffectNotApplicable,
			},
			{
				attrs: []pdp.AttributeAssignment{
					pdp.MakeFloatAssignment("a", 1.0),
					pdp.MakeFloatAssignment("b", 0.0),
					pdp.MakeFloatAssignment("c", 1.0),
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
				attrs: []pdp.AttributeAssignment{
					pdp.MakeIntegerAssignment("a", 1),
					pdp.MakeFloatAssignment("b", 1.0),
				},
				expected: pdp.EffectPermit,
			},
			{
				attrs: []pdp.AttributeAssignment{
					pdp.MakeIntegerAssignment("a", 0),
					pdp.MakeFloatAssignment("b", 0.0),
				},
				expected: pdp.EffectPermit,
			},
			{
				attrs: []pdp.AttributeAssignment{
					pdp.MakeIntegerAssignment("a", -1),
					pdp.MakeFloatAssignment("b", -1.0),
				},
				expected: pdp.EffectPermit,
			},
			{
				attrs: []pdp.AttributeAssignment{
					pdp.MakeIntegerAssignment("a", 1),
					pdp.MakeFloatAssignment("b", 0.0),
				},
				expected: pdp.EffectNotApplicable,
			},
			{
				attrs: []pdp.AttributeAssignment{
					pdp.MakeIntegerAssignment("a", 0),
					pdp.MakeFloatAssignment("b", 1.0),
				},
				expected: pdp.EffectNotApplicable,
			},
			{
				attrs: []pdp.AttributeAssignment{
					pdp.MakeIntegerAssignment("a", 1),
					pdp.MakeFloatAssignment("b", -2.0),
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
				attrs: []pdp.AttributeAssignment{
					pdp.MakeIntegerAssignment("a", 1),
					pdp.MakeFloatAssignment("b", 0.0),
				},
				expected: pdp.EffectPermit,
			},
			{
				attrs: []pdp.AttributeAssignment{
					pdp.MakeIntegerAssignment("a", 0),
					pdp.MakeFloatAssignment("b", -1.0),
				},
				expected: pdp.EffectPermit,
			},
			{
				attrs: []pdp.AttributeAssignment{
					pdp.MakeIntegerAssignment("a", 1),
					pdp.MakeFloatAssignment("b", -1.0),
				},
				expected: pdp.EffectPermit,
			},
			{
				attrs: []pdp.AttributeAssignment{
					pdp.MakeIntegerAssignment("a", 0),
					pdp.MakeFloatAssignment("b", 0.0),
				},
				expected: pdp.EffectNotApplicable,
			},
			{
				attrs: []pdp.AttributeAssignment{
					pdp.MakeIntegerAssignment("a", -1),
					pdp.MakeFloatAssignment("b", 1.0),
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
				attrs: []pdp.AttributeAssignment{
					pdp.MakeIntegerAssignment("a", 1),
					pdp.MakeFloatAssignment("b", 1.3),
					pdp.MakeFloatAssignment("c", 2.3),
				},
				expected: pdp.EffectPermit,
			},

			{
				attrs: []pdp.AttributeAssignment{
					pdp.MakeIntegerAssignment("a", 0),
					pdp.MakeFloatAssignment("b", 0.0),
					pdp.MakeFloatAssignment("c", 0.0),
				},
				expected: pdp.EffectPermit,
			},
			{
				attrs: []pdp.AttributeAssignment{
					pdp.MakeIntegerAssignment("a", -1),
					pdp.MakeFloatAssignment("b", -1.1),
					pdp.MakeFloatAssignment("c", -2.1),
				},
				expected: pdp.EffectPermit,
			},
			{
				attrs: []pdp.AttributeAssignment{
					pdp.MakeIntegerAssignment("a", 1),
					pdp.MakeFloatAssignment("b", 0.0),
					pdp.MakeFloatAssignment("c", 2.0),
				},
				expected: pdp.EffectNotApplicable,
			},
			{
				attrs: []pdp.AttributeAssignment{
					pdp.MakeIntegerAssignment("a", 0),
					pdp.MakeFloatAssignment("b", 1.0),
					pdp.MakeFloatAssignment("c", 2.0),
				},
				expected: pdp.EffectNotApplicable,
			},
			{
				attrs: []pdp.AttributeAssignment{
					pdp.MakeIntegerAssignment("a", 1),
					pdp.MakeFloatAssignment("b", -1.0),
					pdp.MakeFloatAssignment("c", 2.0),
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
				attrs: []pdp.AttributeAssignment{
					pdp.MakeIntegerAssignment("a", 1),
					pdp.MakeFloatAssignment("b", 1.0),
					pdp.MakeFloatAssignment("c", 0.0),
				},
				expected: pdp.EffectPermit,
			},
			{
				attrs: []pdp.AttributeAssignment{
					pdp.MakeIntegerAssignment("a", 0),
					pdp.MakeFloatAssignment("b", 0.0),
					pdp.MakeFloatAssignment("c", 0.0),
				},
				expected: pdp.EffectPermit,
			},
			{
				attrs: []pdp.AttributeAssignment{
					pdp.MakeIntegerAssignment("a", -1),
					pdp.MakeFloatAssignment("b", -1.0),
					pdp.MakeFloatAssignment("c", 0.0),
				},
				expected: pdp.EffectPermit,
			},
			{
				attrs: []pdp.AttributeAssignment{
					pdp.MakeIntegerAssignment("a", 1),
					pdp.MakeFloatAssignment("b", 0.0),
					pdp.MakeFloatAssignment("c", 0.0),
				},
				expected: pdp.EffectNotApplicable,
			},
			{
				attrs: []pdp.AttributeAssignment{
					pdp.MakeIntegerAssignment("a", 0),
					pdp.MakeFloatAssignment("b", 1.0),
					pdp.MakeFloatAssignment("c", 0.0),
				},
				expected: pdp.EffectNotApplicable,
			},
			{
				attrs: []pdp.AttributeAssignment{
					pdp.MakeIntegerAssignment("a", 1),
					pdp.MakeFloatAssignment("b", -1.0),
					pdp.MakeFloatAssignment("c", 0.0),
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
				attrs: []pdp.AttributeAssignment{
					pdp.MakeIntegerAssignment("a", 1),
					pdp.MakeFloatAssignment("b", 1.0),
					pdp.MakeFloatAssignment("c", 1.0),
				},
				expected: pdp.EffectPermit,
			},
			{
				attrs: []pdp.AttributeAssignment{
					pdp.MakeIntegerAssignment("a", 1),
					pdp.MakeFloatAssignment("b", 0.0),
					pdp.MakeFloatAssignment("c", 0.0),
				},
				expected: pdp.EffectPermit,
			},
			{
				attrs: []pdp.AttributeAssignment{
					pdp.MakeIntegerAssignment("a", 0),
					pdp.MakeFloatAssignment("b", 1.0),
					pdp.MakeFloatAssignment("c", 0.0),
				},
				expected: pdp.EffectPermit,
			},
			{
				attrs: []pdp.AttributeAssignment{
					pdp.MakeIntegerAssignment("a", -1),
					pdp.MakeFloatAssignment("b", -1.0),
					pdp.MakeFloatAssignment("c", 1.0),
				},
				expected: pdp.EffectPermit,
			},
			{
				attrs: []pdp.AttributeAssignment{
					pdp.MakeIntegerAssignment("a", -1),
					pdp.MakeFloatAssignment("b", 1.0),
					pdp.MakeFloatAssignment("c", -1.0),
				},
				expected: pdp.EffectPermit,
			},
			{
				attrs: []pdp.AttributeAssignment{
					pdp.MakeIntegerAssignment("a", 1),
					pdp.MakeFloatAssignment("b", 0.0),
					pdp.MakeFloatAssignment("c", 1.0),
				},
				expected: pdp.EffectNotApplicable,
			},
			{
				attrs: []pdp.AttributeAssignment{
					pdp.MakeIntegerAssignment("a", -1),
					pdp.MakeFloatAssignment("b", 1.0),
					pdp.MakeFloatAssignment("c", 1.0),
				},
				expected: pdp.EffectNotApplicable,
			},
			{
				attrs: []pdp.AttributeAssignment{
					pdp.MakeIntegerAssignment("a", 0),
					pdp.MakeFloatAssignment("b", 1.0),
					pdp.MakeFloatAssignment("c", 1.0),
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
				attrs: []pdp.AttributeAssignment{
					pdp.MakeIntegerAssignment("a", 1),
					pdp.MakeFloatAssignment("b", 1.0),
					pdp.MakeFloatAssignment("c", 1.0),
				},
				expected: pdp.EffectPermit,
			},
			{
				attrs: []pdp.AttributeAssignment{
					pdp.MakeIntegerAssignment("a", 0),
					pdp.MakeFloatAssignment("b", 1.0),
					pdp.MakeFloatAssignment("c", 0.0),
				},
				expected: pdp.EffectPermit,
			},
			{
				attrs: []pdp.AttributeAssignment{
					pdp.MakeIntegerAssignment("a", 4),
					pdp.MakeFloatAssignment("b", 2.0),
					pdp.MakeFloatAssignment("c", 2.0),
				},
				expected: pdp.EffectPermit,
			},
			{
				attrs: []pdp.AttributeAssignment{
					pdp.MakeIntegerAssignment("a", 7),
					pdp.MakeFloatAssignment("b", 2.0),
					pdp.MakeFloatAssignment("c", 3.5),
				},
				expected: pdp.EffectPermit,
			},
			{
				attrs: []pdp.AttributeAssignment{
					pdp.MakeIntegerAssignment("a", -1),
					pdp.MakeFloatAssignment("b", 1.0),
					pdp.MakeFloatAssignment("c", -1.0),
				},
				expected: pdp.EffectPermit,
			},
			{
				attrs: []pdp.AttributeAssignment{
					pdp.MakeIntegerAssignment("a", 1),
					pdp.MakeFloatAssignment("b", -1.0),
					pdp.MakeFloatAssignment("c", -1.0),
				},
				expected: pdp.EffectPermit,
			},
			{
				attrs: []pdp.AttributeAssignment{
					pdp.MakeIntegerAssignment("a", 2),
					pdp.MakeFloatAssignment("b", 1.0),
					pdp.MakeFloatAssignment("c", 1.0),
				},
				expected: pdp.EffectNotApplicable,
			},
			{
				attrs: []pdp.AttributeAssignment{
					pdp.MakeIntegerAssignment("a", -1),
					pdp.MakeFloatAssignment("b", 1.0),
					pdp.MakeFloatAssignment("c", 1.0),
				},
				expected: pdp.EffectNotApplicable,
			},
			{
				attrs: []pdp.AttributeAssignment{
					pdp.MakeIntegerAssignment("a", 0),
					pdp.MakeFloatAssignment("b", 1.0),
					pdp.MakeFloatAssignment("c", 1.0),
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
				attrs: []pdp.AttributeAssignment{
					pdp.MakeFloatAssignment("a", 1.0),
					pdp.MakeFloatAssignment("b", 5.0),
					pdp.MakeFloatAssignment("c", 0.0),
				},
				expected:           pdp.EffectPermit,
				expectedObligation: "Below",
			},
			{
				attrs: []pdp.AttributeAssignment{
					pdp.MakeFloatAssignment("a", 1.0),
					pdp.MakeFloatAssignment("b", 5.0),
					pdp.MakeFloatAssignment("c", 10.0),
				},
				expected:           pdp.EffectPermit,
				expectedObligation: "Above",
			},
			{
				attrs: []pdp.AttributeAssignment{
					pdp.MakeFloatAssignment("a", 1.0),
					pdp.MakeFloatAssignment("b", 5.0),
					pdp.MakeFloatAssignment("c", 3.3),
				},
				expected:           pdp.EffectPermit,
				expectedObligation: "Within",
			},
		},
	}

	validateTestSuite(ts, t)
}
