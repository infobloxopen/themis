package jcon

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/pmezard/go-difflib/difflib"
	"github.com/satori/go.uuid"
)

const (
	jsonStream = `{
	"ID": "Test",
	"Items": {
		"first": {
			"type": "set of strings",
			"keys": ["string", "address", "string"],
			"data": {
				"x": {
					"127.0.0.1": {
						"y": {
							"z": false,
							"t": null
						}
					}
				}
			}
		},
		"second": {
			"data": {
				"first": {
					"192.0.2.0/28": {
						"example.com": ["2001:db8::/40", "2001:db8:0100::/40", "2001:db8:0200::/40"],
						"example.net": ["2001:db8:1000::/40", "2001:db8:1100::/40", "2001:db8:1200::/40"]
					},
					"192.0.2.16/28": {
						"example.com": ["2001:db8:2000::/40", "2001:db8:2100::/40", "2001:db8:2200::/40"],
						"example.net": ["2001:db8:3000::/40", "2001:db8:3100::/40", "2001:db8:3200::/40"]
					},
					"192.0.2.32/28": {
						"example.com": ["2001:db8:4000::/40", "2001:db8:4100::/40", "2001:db8:4200::/40"],
						"example.net": ["2001:db8:5000::/40", "2001:db8:5100::/40", "2001:db8:5200::/40"]
					}
				},
				"second": {
					"2001:db8::/36": {
						"example.com": ["2001:db8::/40", "2001:db8:0100::/40", "2001:db8:0200::/40"],
						"example.net": ["2001:db8:1000::/40", "2001:db8:1100::/40", "2001:db8:1200::/40"]
					},
					"2001:db8:1000::/36": {
						"example.com": ["2001:db8:2000::/40", "2001:db8:2100::/40", "2001:db8:2200::/40"],
						"example.net": ["2001:db8:3000::/40", "2001:db8:3100::/40", "2001:db8:3200::/40"]
					},
					"2001:db8:2000::/36": {
						"example.com": ["2001:db8:4000::/40", "2001:db8:4100::/40", "2001:db8:4200::/40"],
						"example.net": ["2001:db8:5000::/40", "2001:db8:5100::/40", "2001:db8:5200::/40"]
					}
				}
			},
			"type": "set of networks",
			"keys": ["string", "address", "domain"]
		}
	}
}`

	testContentItems = `"first": {
  "keys": [
    "String",
    "Address",
    "String"
  ],
  "type": "Set of Strings",
  "data": {
    "x": {
      "127.0.0.1/32": {
        "y": [
          "z",
          "t"
        ]
      }
    }
  }
}
"second": {
  "keys": [
    "String",
    "Address",
    "Domain"
  ],
  "type": "Set of Networks",
  "data": {
    "first": {
      "192.0.2.0/28": {
        "example.com": [
          "2001:db8::/40",
          "2001:db8:100::/40",
          "2001:db8:200::/40"
        ],
        "example.net": [
          "2001:db8:1000::/40",
          "2001:db8:1100::/40",
          "2001:db8:1200::/40"
        ]
      },
      "192.0.2.16/28": {
        "example.com": [
          "2001:db8:2000::/40",
          "2001:db8:2100::/40",
          "2001:db8:2200::/40"
        ],
        "example.net": [
          "2001:db8:3000::/40",
          "2001:db8:3100::/40",
          "2001:db8:3200::/40"
        ]
      },
      "192.0.2.32/28": {
        "example.com": [
          "2001:db8:4000::/40",
          "2001:db8:4100::/40",
          "2001:db8:4200::/40"
        ],
        "example.net": [
          "2001:db8:5000::/40",
          "2001:db8:5100::/40",
          "2001:db8:5200::/40"
        ]
      }
    },
    "second": {
      "2001:db8::/36": {
        "example.com": [
          "2001:db8::/40",
          "2001:db8:100::/40",
          "2001:db8:200::/40"
        ],
        "example.net": [
          "2001:db8:1000::/40",
          "2001:db8:1100::/40",
          "2001:db8:1200::/40"
        ]
      },
      "2001:db8:1000::/36": {
        "example.com": [
          "2001:db8:2000::/40",
          "2001:db8:2100::/40",
          "2001:db8:2200::/40"
        ],
        "example.net": [
          "2001:db8:3000::/40",
          "2001:db8:3100::/40",
          "2001:db8:3200::/40"
        ]
      },
      "2001:db8:2000::/36": {
        "example.com": [
          "2001:db8:4000::/40",
          "2001:db8:4100::/40",
          "2001:db8:4200::/40"
        ],
        "example.net": [
          "2001:db8:5000::/40",
          "2001:db8:5100::/40",
          "2001:db8:5200::/40"
        ]
      }
    }
  }
}`

	jsonUpdateStream = `[
  {
    "op": "Add",
    "path": ["test", "example"],
    "entity": {
      "type": "set of strings",
      "keys": ["string", "address", "string"],
      "data": {
        "x": {
          "127.0.0.1": {
            "y": {
              "z": false,
              "t": null
            }
          }
        }
      }
    }
  },
  {
    "op": "Delete",
    "path": ["test", "example", "x"]
  }
]`

	testContentUpdate = `[
  {
    "op": "Add",
    "path": [
      "test",
      "example"
    ],
    "entity": {
      "keys": [
        "String",
        "Address",
        "String"
      ],
      "type": "Set of Strings",
      "data": {
        "x": {
          "127.0.0.1/32": {
            "y": [
              "z",
              "t"
            ]
          }
        }
      }
    }
  },
  {
    "op": "Delete",
    "path": [
      "test",
      "example",
      "x"
    ]
  }
]`
)

func TestUnmarshal(t *testing.T) {
	id, items, err := Unmarshal(strings.NewReader(jsonStream))
	if err != nil {
		t.Errorf("Expected no error but got (%T):\n\t%s", err, err)
	} else {
		if id != "Test" {
			t.Errorf("Expected \"Test\" as content id but got %q", id)
		}

		s := []string{}
		for p := range items.Enumerate() {
			b, err := json.MarshalIndent(p.Value, "", "  ")
			if err != nil {
				t.Fatalf("Can't marshal result: %s", err)
			}
			s = append(s, fmt.Sprintf("%q: %s", p.Key, string(b)))
		}

		assertJSON(strings.Join(s, "\n"), testContentItems, "JSON content items", t)
	}
}

func TestUnmarshalUpdate(t *testing.T) {
	u, err := UnmarshalUpdate(strings.NewReader(jsonUpdateStream), "test", uuid.NewV4(), uuid.NewV4())
	if err != nil {
		t.Errorf("Expected no error but got (%T):\n\t%s", err, err)
	} else {
		b, err := json.MarshalIndent(u, "", "  ")
		if err != nil {
			t.Errorf("Expected no error but got (%T):\n\t%s", err, err)
		}

		assertJSON(string(b), testContentUpdate, "JSON content update", t)
	}
}

func assertJSON(v, e string, desc string, t *testing.T) {
	ctx := difflib.ContextDiff{
		A:        difflib.SplitLines(e),
		B:        difflib.SplitLines(v),
		FromFile: "Expected",
		ToFile:   "Got"}

	diff, err := difflib.GetContextDiffString(ctx)
	if err != nil {
		panic(fmt.Errorf("Can't compare \"%s\": %s", desc, err))
	}

	if len(diff) > 0 {
		t.Errorf("\"%s\" doesn't match:\n%s", desc, diff)
	}
}
