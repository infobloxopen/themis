package jcon

import (
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/infobloxopen/go-trees/domaintree"
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
    "path": ["second", "second", "2001:db8:1000::/36", "example.net"]
  },
  {
    "op": "Delete",
    "path": ["second", "second", "2001:db8:1000::/36"]
  }
]`

	jsonAllMapsStream = `{
	"ID": "AllMaps",
	"Items": {
		"str-map": {
			"type": "string",
			"keys": ["string"],
			"data": {
				"key-1": "value-1",
				"key-2": "value-2",
				"key-3": "value-3"
			}
		},
		"net-map": {
			"type": "string",
			"keys": ["network"],
			"data": {
				"192.0.2.16/28": "value-1",
				"192.0.2.32/28": "value-2",
				"192.0.2.48/28": "value-3"
			}
		},
		"dom-map": {
			"type": "string",
			"keys": ["domain"],
			"data": {
				"example.com": "value-1",
				"example.net": "value-2",
				"example.org": "value-3"
			}
		}
	}
}`

	jsonPostprocessAllMapsStream = `{
	"ID": "AllMaps",
	"Items": {
		"str-map": {
			"data": {
				"key-1": "value-1",
				"key-2": "value-2",
				"key-3": "value-3"
			},
			"type": "string",
			"keys": ["string"]
		},
		"net-map": {
			"data": {
				"192.0.2.16/28": "value-1",
				"192.0.2.32/28": "value-2",
				"192.0.2.48/28": "value-3"
			},
			"type": "string",
			"keys": ["network"]
		},
		"dom-map": {
			"data": {
				"example.com": "value-1",
				"example.net": "value-2",
				"example.org": "value-3"
			},
			"type": "string",
			"keys": ["domain"]
		}
	}
}`

	jsonAllValuesStream = `{
	"ID": "AllValues",
	"Items": {
		"boolean": {
			"type": "boolean",
			"keys": ["string"],
			"data": {
				"key": true
			}
		},
		"string": {
			"type": "string",
			"keys": ["string"],
			"data": {
				"key": "value"
			}
		},
        "integer": {
            "type": "integer",
            "keys": ["string"],
            "data": {
                "key": 9.007199254740992e+15
            }
        },
		"address": {
			"type": "address",
			"keys": ["string"],
			"data": {
				"key": "192.0.2.1"
			}
		},
		"network": {
			"type": "network",
			"keys": ["string"],
			"data": {
				"key": "192.0.2.0/24"
			}
		},
		"domain": {
			"type": "domain",
			"keys": ["string"],
			"data": {
				"key": "example.com"
			}
		},
		"[]set of strings": {
			"type": "set of strings",
			"keys": ["string"],
			"data": {
				"key": [
					"1-first",
					"2-second",
					"3-third"
				]
			}
		},
		"{}set of strings": {
			"type": "set of strings",
			"keys": ["string"],
			"data": {
				"key": {
					"1-first": "skip me",
					"2-second": {"skip": "me"},
					"3-third": ["skip", "me"]
				}
			}
		},
		"set of networks": {
			"type": "set of networks",
			"keys": ["string"],
			"data": {
				"key": [
					"192.0.2.16/28",
					"192.0.2.32/28",
					"2001:db8::/32"
				]
			}
		},
		"set of domains": {
			"type": "set of domains",
			"keys": ["string"],
			"data": {
				"key": [
					"example.com",
					"example.net",
					"example.org"
				]
			}
		},
		"list of strings": {
			"type": "list of strings",
			"keys": ["string"],
			"data": {
				"key": [
					"1-first",
					"2-second",
					"3-third"
				]
			}
		}
	}
}`

	jsonPostprocessAllValuesStream = `{
	"ID": "AllValues",
	"Items": {
		"boolean": {
			"data": {
				"key": true
			},
			"type": "boolean",
			"keys": ["string"]
		},
		"string": {
			"data": {
				"key": "value"
			},
			"type": "string",
			"keys": ["string"]
		},
        "integer": {
            "type": "integer",
            "keys": ["string"],
            "data": {
                "key": 9.007199254740992e+15
            }
        },
		"address": {
			"data": {
				"key": "192.0.2.1"
			},
			"type": "address",
			"keys": ["string"]
		},
		"network": {
			"data": {
				"key": "192.0.2.0/24"
			},
			"type": "network",
			"keys": ["string"]
		},
		"domain": {
			"data": {
				"key": "example.com"
			},
			"type": "domain",
			"keys": ["string"]
		},
		"[]set of strings": {
			"data": {
				"key": [
					"1-first",
					"2-second",
					"3-third"
				]
			},
			"type": "set of strings",
			"keys": ["string"]
		},
		"{}set of strings": {
			"data": {
				"key": {
					"1-first": "skip me",
					"2-second": {"skip": "me"},
					"3-third": ["skip", "me"]
				}
			},
			"type": "set of strings",
			"keys": ["string"]
		},
		"set of networks": {
			"data": {
				"key": [
					"192.0.2.16/28",
					"192.0.2.32/28",
					"2001:db8::/32"
				]
			},
			"type": "set of networks",
			"keys": ["string"]
		},
		"set of domains": {
			"data": {
				"key": [
					"example.com",
					"example.net",
					"example.org"
				]
			},
			"type": "set of domains",
			"keys": ["string"]
		},
		"list of strings": {
			"data": {
				"key": [
					"1-first",
					"2-second",
					"3-third"
				]
			},
			"type": "list of strings",
			"keys": ["string"]
		}
	}
}`
)

func TestUnmarshal(t *testing.T) {
	c, err := Unmarshal(strings.NewReader(jsonStream), nil)
	if err != nil {
		t.Errorf("Expected no error but got (%T):\n\t%s", err, err)
	} else {
		lc, err := c.Get("missing")
		if err == nil {
			t.Errorf("Expected error but got local content item: %#v", lc)
		}

		lc, err = c.Get("first")
		if err != nil {
			t.Errorf("Expected no error but got (%T):\n\t%s", err, err)
		} else {
			addr, err := pdp.MakeValueFromString(pdp.TypeAddress, "127.0.0.1")
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

			addr, err = pdp.MakeValueFromString(pdp.TypeAddress, "127.0.0.2")
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
			n, err := pdp.MakeValueFromString(pdp.TypeNetwork, "192.0.2.4/30")
			if err != nil {
				t.Errorf("Expected no error but got (%T):\n\t%s", err, err)
			} else {
				path := []pdp.Expression{
					pdp.MakeStringValue("first"),
					n,
					pdp.MakeDomainValue(domaintree.WireDomainNameLower("\x07example\x03com\x00")),
				}
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

			n, err = pdp.MakeValueFromString(pdp.TypeNetwork, "2001:db8:1000:1::/64")
			if err != nil {
				t.Errorf("Expected no error but got (%T):\n\t%s", err, err)
			} else {
				path := []pdp.Expression{
					pdp.MakeStringValue("second"),
					n,
					pdp.MakeDomainValue(domaintree.WireDomainNameLower("\x07example\x03net\x00")),
				}
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

			n, err = pdp.MakeValueFromString(pdp.TypeNetwork, "2001:db8:3000:1::/64")
			if err != nil {
				t.Errorf("Expected no error but got (%T):\n\t%s", err, err)
			} else {
				path := []pdp.Expression{
					pdp.MakeStringValue("second"),
					n,
					pdp.MakeDomainValue(domaintree.WireDomainNameLower("\x07example\x03net\x00")),
				}
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

	c, err = Unmarshal(strings.NewReader(jsonAllMapsStream), nil)
	if err != nil {
		t.Errorf("Expected no error but got (%T):\n\t%s", err, err)
	} else {
		lc, err := c.Get("str-map")
		if err != nil {
			t.Errorf("Expected no error but got (%T):\n\t%s", err, err)
		} else {
			path := []pdp.Expression{pdp.MakeStringValue("key-2")}
			r, err := lc.Get(path, nil)
			if err != nil {
				t.Errorf("Expected no error but got (%T):\n\t%s", err, err)
			} else {
				e := "value-2"
				s, err := r.Serialize()
				if err != nil {
					t.Errorf("Expected no error but got (%T):\n\t%s", err, err)
				} else if s != e {
					t.Errorf("Expected %q but got %q", e, s)
				}
			}
		}

		lc, err = c.Get("net-map")
		if err != nil {
			t.Errorf("Expected no error but got (%T):\n\t%s", err, err)
		} else {
			n, err := pdp.MakeValueFromString(pdp.TypeNetwork, "192.0.2.44/30")
			if err != nil {
				t.Errorf("Expected no error but got (%T):\n\t%s", err, err)
			} else {
				path := []pdp.Expression{n}
				r, err := lc.Get(path, nil)
				if err != nil {
					t.Errorf("Expected no error but got (%T):\n\t%s", err, err)
				} else {
					e := "value-2"
					s, err := r.Serialize()
					if err != nil {
						t.Errorf("Expected no error but got (%T):\n\t%s", err, err)
					} else if s != e {
						t.Errorf("Expected %q but got %q", e, s)
					}
				}
			}
		}

		lc, err = c.Get("dom-map")
		if err != nil {
			t.Errorf("Expected no error but got (%T):\n\t%s", err, err)
		} else {
			path := []pdp.Expression{
				pdp.MakeDomainValue(domaintree.WireDomainNameLower("\x07example\x03net\x00")),
			}
			r, err := lc.Get(path, nil)
			if err != nil {
				t.Errorf("Expected no error but got (%T):\n\t%s", err, err)
			} else {
				e := "value-2"
				s, err := r.Serialize()
				if err != nil {
					t.Errorf("Expected no error but got (%T):\n\t%s", err, err)
				} else if s != e {
					t.Errorf("Expected %q but got %q", e, s)
				}
			}
		}
	}

	c, err = Unmarshal(strings.NewReader(jsonAllMapsStream), nil)
	if err != nil {
		t.Errorf("Expected no error but got (%T):\n\t%s", err, err)
	} else {
		lc, err := c.Get("str-map")
		if err != nil {
			t.Errorf("Expected no error but got (%T):\n\t%s", err, err)
		} else {
			path := []pdp.Expression{pdp.MakeStringValue("key-2")}
			r, err := lc.Get(path, nil)
			if err != nil {
				t.Errorf("Expected no error but got (%T):\n\t%s", err, err)
			} else {
				e := "value-2"
				s, err := r.Serialize()
				if err != nil {
					t.Errorf("Expected no error but got (%T):\n\t%s", err, err)
				} else if s != e {
					t.Errorf("Expected %q but got %q", e, s)
				}
			}
		}

		lc, err = c.Get("net-map")
		if err != nil {
			t.Errorf("Expected no error but got (%T):\n\t%s", err, err)
		} else {
			n, err := pdp.MakeValueFromString(pdp.TypeNetwork, "192.0.2.44/30")
			if err != nil {
				t.Errorf("Expected no error but got (%T):\n\t%s", err, err)
			} else {
				path := []pdp.Expression{n}
				r, err := lc.Get(path, nil)
				if err != nil {
					t.Errorf("Expected no error but got (%T):\n\t%s", err, err)
				} else {
					e := "value-2"
					s, err := r.Serialize()
					if err != nil {
						t.Errorf("Expected no error but got (%T):\n\t%s", err, err)
					} else if s != e {
						t.Errorf("Expected %q but got %q", e, s)
					}
				}
			}
		}

		lc, err = c.Get("dom-map")
		if err != nil {
			t.Errorf("Expected no error but got (%T):\n\t%s", err, err)
		} else {
			path := []pdp.Expression{
				pdp.MakeDomainValue(domaintree.WireDomainNameLower("\x07example\x03net\x00")),
			}
			r, err := lc.Get(path, nil)
			if err != nil {
				t.Errorf("Expected no error but got (%T):\n\t%s", err, err)
			} else {
				e := "value-2"
				s, err := r.Serialize()
				if err != nil {
					t.Errorf("Expected no error but got (%T):\n\t%s", err, err)
				} else if s != e {
					t.Errorf("Expected %q but got %q", e, s)
				}
			}
		}
	}

	c, err = Unmarshal(strings.NewReader(jsonAllValuesStream), nil)
	if err != nil {
		t.Errorf("Expected no error but got (%T):\n\t%s", err, err)
	} else {
		path := []pdp.Expression{pdp.MakeStringValue("key")}

		lc, err := c.Get("boolean")
		if err != nil {
			t.Errorf("Expected no error but got (%T):\n\t%s", err, err)
		} else {
			r, err := lc.Get(path, nil)
			if err != nil {
				t.Errorf("Expected no error but got (%T):\n\t%s", err, err)
			} else {
				e := "true"
				s, err := r.Serialize()
				if err != nil {
					t.Errorf("Expected no error but got (%T):\n\t%s", err, err)
				} else if s != e {
					t.Errorf("Expected %q but got %q", e, s)
				}
			}
		}

		lc, err = c.Get("string")
		if err != nil {
			t.Errorf("Expected no error but got (%T):\n\t%s", err, err)
		} else {
			r, err := lc.Get(path, nil)
			if err != nil {
				t.Errorf("Expected no error but got (%T):\n\t%s", err, err)
			} else {
				e := "value"
				s, err := r.Serialize()
				if err != nil {
					t.Errorf("Expected no error but got (%T):\n\t%s", err, err)
				} else if s != e {
					t.Errorf("Expected %q but got %q", e, s)
				}
			}
		}

		lc, err = c.Get("address")
		if err != nil {
			t.Errorf("Expected no error but got (%T):\n\t%s", err, err)
		} else {
			r, err := lc.Get(path, nil)
			if err != nil {
				t.Errorf("Expected no error but got (%T):\n\t%s", err, err)
			} else {
				e := "192.0.2.1"
				s, err := r.Serialize()
				if err != nil {
					t.Errorf("Expected no error but got (%T):\n\t%s", err, err)
				} else if s != e {
					t.Errorf("Expected %q but got %q", e, s)
				}
			}
		}

		lc, err = c.Get("network")
		if err != nil {
			t.Errorf("Expected no error but got (%T):\n\t%s", err, err)
		} else {
			r, err := lc.Get(path, nil)
			if err != nil {
				t.Errorf("Expected no error but got (%T):\n\t%s", err, err)
			} else {
				e := "192.0.2.0/24"
				s, err := r.Serialize()
				if err != nil {
					t.Errorf("Expected no error but got (%T):\n\t%s", err, err)
				} else if s != e {
					t.Errorf("Expected %q but got %q", e, s)
				}
			}
		}

		lc, err = c.Get("domain")
		if err != nil {
			t.Errorf("Expected no error but got (%T):\n\t%s", err, err)
		} else {
			r, err := lc.Get(path, nil)
			if err != nil {
				t.Errorf("Expected no error but got (%T):\n\t%s", err, err)
			} else {
				e := "example.com"
				s, err := r.Serialize()
				if err != nil {
					t.Errorf("Expected no error but got (%T):\n\t%s", err, err)
				} else if s != e {
					t.Errorf("Expected %q but got %q", e, s)
				}
			}
		}

		lc, err = c.Get("[]set of strings")
		if err != nil {
			t.Errorf("Expected no error but got (%T):\n\t%s", err, err)
		} else {
			r, err := lc.Get(path, nil)
			if err != nil {
				t.Errorf("Expected no error but got (%T):\n\t%s", err, err)
			} else {
				e := "\"1-first\",\"2-second\",\"3-third\""
				s, err := r.Serialize()
				if err != nil {
					t.Errorf("Expected no error but got (%T):\n\t%s", err, err)
				} else if s != e {
					t.Errorf("Expected %q but got %q", e, s)
				}
			}
		}

		lc, err = c.Get("{}set of strings")
		if err != nil {
			t.Errorf("Expected no error but got (%T):\n\t%s", err, err)
		} else {
			r, err := lc.Get(path, nil)
			if err != nil {
				t.Errorf("Expected no error but got (%T):\n\t%s", err, err)
			} else {
				e := "\"1-first\",\"2-second\",\"3-third\""
				s, err := r.Serialize()
				if err != nil {
					t.Errorf("Expected no error but got (%T):\n\t%s", err, err)
				} else if s != e {
					t.Errorf("Expected %q but got %q", e, s)
				}
			}
		}

		lc, err = c.Get("set of networks")
		if err != nil {
			t.Errorf("Expected no error but got (%T):\n\t%s", err, err)
		} else {
			r, err := lc.Get(path, nil)
			if err != nil {
				t.Errorf("Expected no error but got (%T):\n\t%s", err, err)
			} else {
				e := "\"192.0.2.16/28\",\"192.0.2.32/28\",\"2001:db8::/32\""
				s, err := r.Serialize()
				if err != nil {
					t.Errorf("Expected no error but got (%T):\n\t%s", err, err)
				} else if s != e {
					t.Errorf("Expected %q but got %q", e, s)
				}
			}
		}

		lc, err = c.Get("set of domains")
		if err != nil {
			t.Errorf("Expected no error but got (%T):\n\t%s", err, err)
		} else {
			r, err := lc.Get(path, nil)
			if err != nil {
				t.Errorf("Expected no error but got (%T):\n\t%s", err, err)
			} else {
				e := "\"example.com\",\"example.net\",\"example.org\""
				s, err := r.Serialize()
				if err != nil {
					t.Errorf("Expected no error but got (%T):\n\t%s", err, err)
				} else if s != e {
					t.Errorf("Expected %q but got %q", e, s)
				}
			}
		}

		lc, err = c.Get("list of strings")
		if err != nil {
			t.Errorf("Expected no error but got (%T):\n\t%s", err, err)
		} else {
			r, err := lc.Get(path, nil)
			if err != nil {
				t.Errorf("Expected no error but got (%T):\n\t%s", err, err)
			} else {
				e := "\"1-first\",\"2-second\",\"3-third\""
				s, err := r.Serialize()
				if err != nil {
					t.Errorf("Expected no error but got (%T):\n\t%s", err, err)
				} else if s != e {
					t.Errorf("Expected %q but got %q", e, s)
				}
			}
		}
	}

	c, err = Unmarshal(strings.NewReader(jsonPostprocessAllValuesStream), nil)
	if err != nil {
		t.Errorf("Expected no error but got (%T):\n\t%s", err, err)
	} else {
		path := []pdp.Expression{pdp.MakeStringValue("key")}

		lc, err := c.Get("boolean")
		if err != nil {
			t.Errorf("Expected no error but got (%T):\n\t%s", err, err)
		} else {
			r, err := lc.Get(path, nil)
			if err != nil {
				t.Errorf("Expected no error but got (%T):\n\t%s", err, err)
			} else {
				e := "true"
				s, err := r.Serialize()
				if err != nil {
					t.Errorf("Expected no error but got (%T):\n\t%s", err, err)
				} else if s != e {
					t.Errorf("Expected %q but got %q", e, s)
				}
			}
		}

		lc, err = c.Get("string")
		if err != nil {
			t.Errorf("Expected no error but got (%T):\n\t%s", err, err)
		} else {
			r, err := lc.Get(path, nil)
			if err != nil {
				t.Errorf("Expected no error but got (%T):\n\t%s", err, err)
			} else {
				e := "value"
				s, err := r.Serialize()
				if err != nil {
					t.Errorf("Expected no error but got (%T):\n\t%s", err, err)
				} else if s != e {
					t.Errorf("Expected %q but got %q", e, s)
				}
			}
		}

		lc, err = c.Get("address")
		if err != nil {
			t.Errorf("Expected no error but got (%T):\n\t%s", err, err)
		} else {
			r, err := lc.Get(path, nil)
			if err != nil {
				t.Errorf("Expected no error but got (%T):\n\t%s", err, err)
			} else {
				e := "192.0.2.1"
				s, err := r.Serialize()
				if err != nil {
					t.Errorf("Expected no error but got (%T):\n\t%s", err, err)
				} else if s != e {
					t.Errorf("Expected %q but got %q", e, s)
				}
			}
		}

		lc, err = c.Get("network")
		if err != nil {
			t.Errorf("Expected no error but got (%T):\n\t%s", err, err)
		} else {
			r, err := lc.Get(path, nil)
			if err != nil {
				t.Errorf("Expected no error but got (%T):\n\t%s", err, err)
			} else {
				e := "192.0.2.0/24"
				s, err := r.Serialize()
				if err != nil {
					t.Errorf("Expected no error but got (%T):\n\t%s", err, err)
				} else if s != e {
					t.Errorf("Expected %q but got %q", e, s)
				}
			}
		}

		lc, err = c.Get("domain")
		if err != nil {
			t.Errorf("Expected no error but got (%T):\n\t%s", err, err)
		} else {
			r, err := lc.Get(path, nil)
			if err != nil {
				t.Errorf("Expected no error but got (%T):\n\t%s", err, err)
			} else {
				e := "example.com"
				s, err := r.Serialize()
				if err != nil {
					t.Errorf("Expected no error but got (%T):\n\t%s", err, err)
				} else if s != e {
					t.Errorf("Expected %q but got %q", e, s)
				}
			}
		}

		lc, err = c.Get("[]set of strings")
		if err != nil {
			t.Errorf("Expected no error but got (%T):\n\t%s", err, err)
		} else {
			r, err := lc.Get(path, nil)
			if err != nil {
				t.Errorf("Expected no error but got (%T):\n\t%s", err, err)
			} else {
				e := "\"1-first\",\"2-second\",\"3-third\""
				s, err := r.Serialize()
				if err != nil {
					t.Errorf("Expected no error but got (%T):\n\t%s", err, err)
				} else if s != e {
					t.Errorf("Expected %q but got %q", e, s)
				}
			}
		}

		lc, err = c.Get("{}set of strings")
		if err != nil {
			t.Errorf("Expected no error but got (%T):\n\t%s", err, err)
		} else {
			r, err := lc.Get(path, nil)
			if err != nil {
				t.Errorf("Expected no error but got (%T):\n\t%s", err, err)
			} else {
				e := "\"1-first\",\"2-second\",\"3-third\""
				s, err := r.Serialize()
				if err != nil {
					t.Errorf("Expected no error but got (%T):\n\t%s", err, err)
				} else if s != e {
					t.Errorf("Expected %q but got %q", e, s)
				}
			}
		}

		lc, err = c.Get("set of networks")
		if err != nil {
			t.Errorf("Expected no error but got (%T):\n\t%s", err, err)
		} else {
			r, err := lc.Get(path, nil)
			if err != nil {
				t.Errorf("Expected no error but got (%T):\n\t%s", err, err)
			} else {
				e := "\"192.0.2.16/28\",\"192.0.2.32/28\",\"2001:db8::/32\""
				s, err := r.Serialize()
				if err != nil {
					t.Errorf("Expected no error but got (%T):\n\t%s", err, err)
				} else if s != e {
					t.Errorf("Expected %q but got %q", e, s)
				}
			}
		}

		lc, err = c.Get("set of domains")
		if err != nil {
			t.Errorf("Expected no error but got (%T):\n\t%s", err, err)
		} else {
			r, err := lc.Get(path, nil)
			if err != nil {
				t.Errorf("Expected no error but got (%T):\n\t%s", err, err)
			} else {
				e := "\"example.com\",\"example.net\",\"example.org\""
				s, err := r.Serialize()
				if err != nil {
					t.Errorf("Expected no error but got (%T):\n\t%s", err, err)
				} else if s != e {
					t.Errorf("Expected %q but got %q", e, s)
				}
			}
		}

		lc, err = c.Get("list of strings")
		if err != nil {
			t.Errorf("Expected no error but got (%T):\n\t%s", err, err)
		} else {
			r, err := lc.Get(path, nil)
			if err != nil {
				t.Errorf("Expected no error but got (%T):\n\t%s", err, err)
			} else {
				e := "\"1-first\",\"2-second\",\"3-third\""
				s, err := r.Serialize()
				if err != nil {
					t.Errorf("Expected no error but got (%T):\n\t%s", err, err)
				} else if s != e {
					t.Errorf("Expected %q but got %q", e, s)
				}
			}
		}
	}
}

func TestUnmarshalUpdate(t *testing.T) {
	tag := uuid.New()
	c, err := Unmarshal(strings.NewReader(jsonStream), &tag)
	if err != nil {
		t.Errorf("Expected no error but got (%T):\n\t%s", err, err)
		return
	}

	u, err := UnmarshalUpdate(strings.NewReader(jsonUpdateStream), "Test", tag, uuid.New())
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
		addr, err := pdp.MakeValueFromString(pdp.TypeAddress, "127.0.0.2")
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
		n, err := pdp.MakeValueFromString(pdp.TypeNetwork, "2001:db8:1000:1::/64")
		if err != nil {
			t.Errorf("Expected no error but got (%T):\n\t%s", err, err)
		} else {
			path := []pdp.Expression{
				pdp.MakeStringValue("second"),
				n,
				pdp.MakeDomainValue(domaintree.WireDomainNameLower("\x07example\x03com\x00")),
			}
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
