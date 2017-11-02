package jast

import (
	"fmt"
	"strings"
	"testing"

	"github.com/google/uuid"
)

var testCases = []map[string]string{
	map[string]string{
		"policy": `
{
  "attributes": {
    "a": "sometype"
  },
  "policies": {
    "id": "Default",
    "alg": "FirstApplicableEffect",
    "rules": [
      {
        "effect": "Permit"
      }
    ]
  }
}
`,
		"err": fmt.Sprintf("%T", &attributeTypeError{}),
	},

	map[string]string{
		"policy": `
{
  "attributes": {
    "a": "string"
  },
  "policies": {
    "id": "Default",
    "alg": "FirstApplicableEffect",
    "rules": [
      {
        "effect": "Permit"
      }
    ],
    "policies": [
      {
        "id": "Default",
        "alg": "FirstApplicableEffect",
        "rules": [
          {
            "effect": "Permit"
          }
        ]
      }
    ]
  }
}
`,
		"err": fmt.Sprintf("%T", &policyAmbiguityError{}),
	},

	map[string]string{
		"policy": `
{
  "attributes": {
    "a": "string"
  },
  "policies": {
    "id": "Default",
    "alg": "FirstApplicableEffect"
  }
}
`,
		"err": fmt.Sprintf("%T", &policyMissingKeyError{}),
	},

	map[string]string{
		"policy": `
{
  "attributes": {
    "a": "string"
  },
  "policies": {
    "id": "Default",
    "alg": "SomeAlg",
    "rules": [
      {
        "effect": "Permit"
      }
    ]
  }
}
`,
		"err": fmt.Sprintf("%T", &unknownRCAError{}),
	},

	map[string]string{
		"policy": `
{
  "attributes": {
    "a": "string"
  },
  "policies": {
    "id": "Default",
    "rules": [
      {
        "effect": "Permit"
      }
    ]
  }
}
`,
		"err": fmt.Sprintf("%T", &missingRCAError{}),
	},

	map[string]string{
		"policy": `
{
  "attributes": {
    "a": "string"
  },
  "policies": {
    "id": "Default",
    "alg": {
      "id": "mapper",
      "map": {
        "attr": "a"
      },
      "default": "Default"
    },
    "rules": [
      {
        "id": "Error",
        "effect": "Deny"
      }
    ]
  }
}
`,
		"err": fmt.Sprintf("%T", &missingDefaultRuleRCAError{}),
	},

	map[string]string{
		"policy": `
{
  "attributes": {
    "a": "string"
  },
  "policies": {
    "id": "Default",
    "alg": {
      "id": "mapper",
      "map": {
        "attr": "a"
      },
      "error": "Error"
    },
    "rules": [
      {
        "id": "Default",
        "effect": "Permit"
      }
    ]
  }
}
`,
		"err": fmt.Sprintf("%T", &missingErrorRuleRCAError{}),
	},

	map[string]string{
		"policy": `
{
  "attributes": {
    "a": "string"
  },
  "policies": {
    "id": "Default",
    "alg": "SomeAlg",
    "policies": [
      {
        "id": "Default",
        "alg": "FirstApplicableEffect",
        "rules": [
          {
            "effect": "Permit"
          }
        ]
      }
    ]
  }
}
`,
		"err": fmt.Sprintf("%T", &unknownPCAError{}),
	},

	map[string]string{
		"policy": `
{
  "attributes": {
    "a": "string"
  },
  "policies": {
    "id": "Default",
    "policies": [
      {
        "id": "Default",
        "alg": "FirstApplicableEffect",
        "rules": [
          {
            "effect": "Permit"
          }
        ]
      }
    ]
  }
}
`,
		"err": fmt.Sprintf("%T", &missingPCAError{}),
	},

	map[string]string{
		"policy": `
{
  "attributes": {
    "a": "string"
  },
  "policies": {
    "id": "Default",
    "alg": {
      "id": "mapper",
      "map": {
        "attr": "a"
      },
      "default": "Default"
    },
    "policies": [
      {
        "id": "Error",
        "alg": "FirstApplicableEffect",
        "rules": [
          {
            "effect": "Deny"
          }
        ]
      }
    ]
  }
}
`,
		"err": fmt.Sprintf("%T", &missingDefaultPolicyPCAError{}),
	},

	map[string]string{
		"policy": `
{
  "attributes": {
    "a": "string"
  },
  "policies": {
    "id": "Default",
    "alg": {
      "id": "mapper",
      "map": {
        "attr": "a"
      },
      "error": "Error"
    },
    "policies": [
      {
        "id": "Default",
        "alg": "FirstApplicableEffect",
        "rules": [
          {
            "effect": "Permit"
          }
        ]
      }
    ]
  }
}
`,
		"err": fmt.Sprintf("%T", &missingErrorPolicyPCAError{}),
	},

	map[string]string{
		"policy": `
{
  "attributes": {
    "a": "boolean"
  },
  "policies": {
    "id": "Default",
    "alg": {
      "id": "mapper",
      "map": {
        "attr": "a"
      }
    },
    "rules": [
      {
        "effect": "Permit"
      }
    ]
  }
}
`,
		"err": fmt.Sprintf("%T", &mapperArgumentTypeError{}),
	},

	map[string]string{
		"policy": `
{
  "attributes": {
    "a": "string"
  },
  "policies": {
    "id": "Default",
    "alg": "FirstApplicableEffect",
    "rules": [
      {
        "condition": {
          "attr": "a"
        },
        "effect": "Permit"
      }
    ]
  }
}
`,
		"err": fmt.Sprintf("%T", &conditionTypeError{}),
	},

	map[string]string{
		"policy": `
{
  "attributes": {
    "a": "string"
  },
  "policies": {
    "id": "Default",
    "alg": "FirstApplicableEffect",
    "rules": [
      {
        "effect": "Bye"
      }
    ]
  }
}
`,
		"err": fmt.Sprintf("%T", &unknownEffectError{}),
	},

	map[string]string{
		"policy": `
{
  "attributes": {
    "a": "string"
  },
  "policies": {
    "id": "Default",
    "alg": "FirstApplicableEffect",
    "target": [
      {
        "some": [
          {
            "attr": "a"
          },
          {
            "val": {
              "type": "string",
              "content": "string"
            }
          }
        ]
      }
    ],
    "rules": [
      {
        "effect": "Deny"
      }
    ]
  }
}
`,
		"err": fmt.Sprintf("%T", &unknownMatchFunctionError{}),
	},

	map[string]string{
		"policy": `
{
  "attributes": {
    "a": "boolean"
  },
  "policies": {
    "id": "Default",
    "alg": "FirstApplicableEffect",
    "target": [
      {
        "contains": [
          {
            "attr": "a"
          },
          {
            "val": {
              "type": "string",
              "content": "string"
            }
          }
        ]
      }
    ],
    "rules": [
      {
        "effect": "Deny"
      }
    ]
  }
}
`,
		"err": fmt.Sprintf("%T", &matchFunctionCastError{}),
	},

	map[string]string{
		"policy": `
{
  "attributes": {
    "a": "string"
  },
  "policies": {
    "id": "Default",
    "alg": "FirstApplicableEffect",
    "target": [
      {
        "equal": [
          {
            "attr": "a"
          }
        ]
      }
    ],
    "rules": [
      {
        "effect": "Deny"
      }
    ]
  }
}
`,
		"err": fmt.Sprintf("%T", &matchFunctionArgsNumberError{}),
	},

	map[string]string{
		"policy": `
{
  "attributes": {
    "a": "string",
    "b": "string"
  },
  "policies": {
    "id": "Default",
    "alg": "FirstApplicableEffect",
    "target": [
      {
        "equal": [
          {
            "attr": "a"
          },
          {
            "attr": "b"
          }
        ]
      }
    ],
    "rules": [
      {
        "effect": "Deny"
      }
    ]
  }
}
`,
		"err": fmt.Sprintf("%T", &matchFunctionBothAttrsError{}),
	},

	map[string]string{
		"policy": `
{
  "attributes": {
    "a": "string"
  },
  "policies": {
    "id": "Default",
    "alg": "FirstApplicableEffect",
    "target": [
      {
        "equal": [
          {
            "val": {
              "type": "string",
              "content": "string"
            }
          },
          {
            "val": {
              "type": "string",
              "content": "string"
            }
          }
        ]
      }
    ],
    "rules": [
      {
        "effect": "Deny"
      }
    ]
  }
}
`,
		"err": fmt.Sprintf("%T", &matchFunctionBothValuesError{}),
	},

	map[string]string{
		"policy": `
{
  "attributes": {
    "a": "string",
    "b": "string"
  },
  "policies": {
    "id": "Default",
    "alg": "FirstApplicableEffect",
    "target": [
      {
        "equal": [
          {
            "some": "a"
          },
          {
            "some": "b"
          }
        ]
      }
    ],
    "rules": [
      {
        "effect": "Deny"
      }
    ]
  }
}
`,
		"err": fmt.Sprintf("%T", &unknownFunctionError{}),
	},
}

func TestUnmarshalErrors(t *testing.T) {
	p := Parser{}

	for _, tc := range testCases {
		_, err := p.Unmarshal(strings.NewReader(tc["policy"]), nil)
		if err == nil {
			t.Errorf("Expected %s error but got nothing", tc["err"])
		} else if e := fmt.Sprintf("%T", err); e != tc["err"] {
			t.Errorf("Expected %s error but got %s", tc["err"], e)
		}
	}
}

var testCasesUpdate = []map[string]string{
	map[string]string{
		"update": `
[
  {
    "op": "some",
    "path": [
      "Parent policy set"
    ],
    "entity": {
      "id": "Policy Set",
      "alg": "FirstApplicableEffect",
      "rules": {
        "effect": "Permit"
      }
    }
  }
]
`,
		"err": fmt.Sprintf("%T", &unknownPolicyUpdateOperationError{}),
	},

	map[string]string{
		"update": `
[
  {
    "op": "add",
    "path": [
      "Parent policy set"
    ],
    "entity": {
      "id": "Policy Set"
    }
  }
]
`,
		"err": fmt.Sprintf("%T", &entityMissingKeyError{}),
	},

	map[string]string{
		"update": `
[
  {
    "op": "add",
    "path": [
      "Parent policy set"
    ],
    "entity": {
      "id": "Policy Set",
      "alg": "FirstApplicableEffect",
      "rules": [
        {
          "effect": "Permit"
        }
      ],
      "policies": [
        {
          "id": "Permit Policy",
          "alg": "FirstApplicableEffect",
          "rules": [
            {
              "effect": "Permit"
            }
          ]
        }
      ]
    }
  }
]
`,
		"err": fmt.Sprintf("%T", &entityAmbiguityError{}),
	},
}

func TestUnmarshalUpdateErrors(t *testing.T) {
	p := Parser{}
	tag := uuid.New()
	s, err := p.Unmarshal(strings.NewReader(policyToUpdate), &tag)
	if err != nil {
		t.Errorf("Expected no error but got %T (%s)", err, err)
		return
	}

	for _, tc := range testCasesUpdate {
		tr, err := s.NewTransaction(&tag)
		if err != nil {
			t.Errorf("Expected no error but got %T (%s)", err, err)
			return
		}

		_, err = p.UnmarshalUpdate(strings.NewReader(tc["update"]), tr.Attributes(), tag, uuid.New())
		if err == nil {
			t.Errorf("Expected %s error but got nothing", tc["err"])
			return
		}

		if e := fmt.Sprintf("%T", err); e != tc["err"] {
			t.Errorf("Expected %s error but got %s", tc["err"], e)
		}
	}
}
