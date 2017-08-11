package jcon

import (
	"strings"
	"testing"

	"github.com/satori/go.uuid"

	"github.com/infobloxopen/themis/pdp"
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
			"keys": ["string", "network", "domain"]
		}
	}
}`

	jsonUpdateStream = `[
  {
    "op": "Add",
    "path": ["first", "update"],
    "entity": {
      "type": "set of strings",
      "keys": ["address", "string"],
      "data": {
        "127.0.0.2": {
          "n": {
            "p": false,
            "q": null
          }
        }
      }
    }
  },
  {
    "op": "Delete",
    "path": ["second", "second", "2001:db8:1000::/36"]
  }
]`
)

func TestUnmarshal(t *testing.T) {
	c, err := Unmarshal(strings.NewReader(jsonStream), nil)
	if err != nil {
		t.Errorf("Expected no error but got (%T):\n\t%s", err, err)
		return
	}

	lc, err := c.Get("missing")
	if err == nil {
		t.Errorf("Expected error but got local content item: %#v", lc)
	}

	lc, err = c.Get("first")
	if err != nil {
		t.Errorf("Expected no error but got (%T):\n\t%s", err, err)
	} else {
		addr, err := pdp.MakeValueFromSting(pdp.TypeAddress, "127.0.0.1")
		if err != nil {
			t.Errorf("Expected no error but got (%T):\n\t%s", err, err)
		} else {
			path := []pdp.Expression{pdp.MakeStringValue("x"), addr, pdp.MakeStringValue("y")}
			r, err := lc.Get(path, nil)
			if err != nil {
				t.Errorf("Expected no error but got (%T):\n\t%s", err, err)
			} else {
				e := "\"z\",\"t\""
				s, err := r.Serialize()
				if err != nil {
					t.Errorf("Expected no error but got (%T):\n\t%s", err, err)
				} else if s != e {
					t.Errorf("Expected [%s] but got [%s]", e, s)
				}
			}
		}

		addr, err = pdp.MakeValueFromSting(pdp.TypeAddress, "127.0.0.2")
		if err != nil {
			t.Errorf("Expected no error but got (%T):\n\t%s", err, err)
		} else {
			path := []pdp.Expression{pdp.MakeStringValue("x"), addr, pdp.MakeStringValue("y")}
			r, err := lc.Get(path, nil)
			if err == nil {
				s, err := r.Serialize()
				if err != nil {
					s = err.Error()
				}
				t.Errorf("Expected error but got result %s", s)
			}
		}
	}

	lc, err = c.Get("second")
	if err != nil {
		t.Errorf("Expected no error but got (%T):\n\t%s", err, err)
	} else {
		n, err := pdp.MakeValueFromSting(pdp.TypeNetwork, "192.0.2.4/30")
		if err != nil {
			t.Errorf("Expected no error but got (%T):\n\t%s", err, err)
		} else {
			path := []pdp.Expression{pdp.MakeStringValue("first"), n, pdp.MakeDomainValue("example.com")}
			r, err := lc.Get(path, nil)
			if err != nil {
				t.Errorf("Expected no error but got (%T):\n\t%s", err, err)
			} else {
				e := "\"2001:db8::/40\",\"2001:db8:100::/40\",\"2001:db8:200::/40\""
				s, err := r.Serialize()
				if err != nil {
					t.Errorf("Expected no error but got (%T):\n\t%s", err, err)
				} else if s != e {
					t.Errorf("Expected [%s] but got [%s]", e, s)
				}
			}
		}

		n, err = pdp.MakeValueFromSting(pdp.TypeNetwork, "2001:db8:1000:1::/64")
		if err != nil {
			t.Errorf("Expected no error but got (%T):\n\t%s", err, err)
		} else {
			path := []pdp.Expression{pdp.MakeStringValue("second"), n, pdp.MakeDomainValue("example.net")}
			r, err := lc.Get(path, nil)
			if err != nil {
				t.Errorf("Expected no error but got (%T):\n\t%s", err, err)
			} else {
				e := "\"2001:db8:3000::/40\",\"2001:db8:3100::/40\",\"2001:db8:3200::/40\""
				s, err := r.Serialize()
				if err != nil {
					t.Errorf("Expected no error but got (%T):\n\t%s", err, err)
				} else if s != e {
					t.Errorf("Expected [%s] but got [%s]", e, s)
				}
			}
		}

		n, err = pdp.MakeValueFromSting(pdp.TypeNetwork, "2001:db8:3000:1::/64")
		if err != nil {
			t.Errorf("Expected no error but got (%T):\n\t%s", err, err)
		} else {
			path := []pdp.Expression{pdp.MakeStringValue("second"), n, pdp.MakeDomainValue("example.net")}
			r, err := lc.Get(path, nil)
			if err == nil {
				s, err := r.Serialize()
				if err != nil {
					s = err.Error()
				}
				t.Errorf("Expected error but got result %s", s)
			}
		}
	}
}

func TestUnmarshalUpdate(t *testing.T) {
	tag := uuid.NewV4()
	c, err := Unmarshal(strings.NewReader(jsonStream), &tag)
	if err != nil {
		t.Errorf("Expected no error but got (%T):\n\t%s", err, err)
		return
	}

	u, err := UnmarshalUpdate(strings.NewReader(jsonUpdateStream), "Test", tag, uuid.NewV4())
	if err != nil {
		t.Errorf("Expected no error but got (%T):\n\t%s", err, err)
		return
	}

	s := pdp.NewLocalContentStorage([]*pdp.LocalContent{c})
	tr, err := s.NewTransaction("Test", &tag)
	if err != nil {
		t.Errorf("Expected no error but got (%T):\n\t%s", err, err)
		return
	}

	err = tr.Apply(u)
	if err != nil {
		t.Errorf("Expected no error but got (%T):\n\t%s", err, err)
		return
	}

	s, err = tr.Commit(s)
	if err != nil {
		t.Errorf("Expected no error but got (%T):\n\t%s", err, err)
		return
	}

	lc, err := s.Get("Test", "first")
	if err != nil {
		t.Errorf("Expected no error but got (%T):\n\t%s", err, err)
	} else {
		addr, err := pdp.MakeValueFromSting(pdp.TypeAddress, "127.0.0.2")
		if err != nil {
			t.Errorf("Expected no error but got (%T):\n\t%s", err, err)
		} else {
			path := []pdp.Expression{pdp.MakeStringValue("update"), addr, pdp.MakeStringValue("n")}
			r, err := lc.Get(path, nil)
			if err != nil {
				t.Errorf("Expected no error but got (%T):\n\t%s", err, err)
			} else {
				e := "\"p\",\"q\""
				s, err := r.Serialize()
				if err != nil {
					t.Errorf("Expected no error but got (%T):\n\t%s", err, err)
				} else if s != e {
					t.Errorf("Expected [%s] but got [%s]", e, s)
				}
			}
		}
	}

	lc, err = s.Get("Test", "second")
	if err != nil {
		t.Errorf("Expected no error but got (%T):\n\t%s", err, err)
	} else {
		n, err := pdp.MakeValueFromSting(pdp.TypeNetwork, "2001:db8:1000:1::/64")
		if err != nil {
			t.Errorf("Expected no error but got (%T):\n\t%s", err, err)
		} else {
			path := []pdp.Expression{pdp.MakeStringValue("second"), n, pdp.MakeDomainValue("example.com")}
			r, err := lc.Get(path, nil)
			if err == nil {
				s, err := r.Serialize()
				if err != nil {
					s = err.Error()
				}
				t.Errorf("Expected error but got result %s", s)
			}
		}
	}
}
