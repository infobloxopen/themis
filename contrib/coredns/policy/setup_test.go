package policy

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/mholt/caddy"
)

func TestPolicyConfigParse(t *testing.T) {
	tests := []struct {
		input       string
		endpoints   []string
		errContent  error
		options     map[uint16][]*edns0Map
		debugSuffix *string
		streams     *int
		hotSpot     *bool
		confAttrs   map[string]confAttrType
		ident       *string
		passthrough []string
		connTimeout *time.Duration
	}{
		{
			input: `.:53 {
						log stdout
					}`,
			errContent: errors.New("Policy setup called without keyword 'policy' in Corefile"),
		},
		{
			input: `.:53 {
						policy {
							error option
						}
					}`,
		},
		{
			input: `.:53 {
						policy {
							endpoint
						}
					}`,
			errContent: errors.New("Wrong argument count or unexpected line ending"),
		},
		{
			input: `.:53 {
						policy {
							endpoint 10.2.4.1:5555
						}
					}`,
			endpoints: []string{"10.2.4.1:5555"},
		},
		{
			input: `.:53 {
						policy {
							endpoint 10.2.4.1:5555 10.2.4.2:5555
						}
					}`,
			endpoints: []string{"10.2.4.1:5555", "10.2.4.2:5555"},
		},
		{
			input: `.:53 {
						policy {
							endpoint 10.2.4.1:5555
							edns0 0xfff0 uid hex string wrong_size 0 32
						}
					}`,
			errContent: errors.New("Could not parse EDNS0 data size"),
		},
		{
			input: `.:53 {
						policy {
							endpoint 10.2.4.1:5555
							edns0 0xfff0 uid hex string 32 0 32
						}
					}`,
			options: map[uint16][]*edns0Map{
				0xfff0: {
					&edns0Map{
						name:     "uid",
						dataType: typeEDNS0Hex,
						destType: "string",
						size:     32,
						start:    0,
						end:      32},
				},
			},
			confAttrs: map[string]confAttrType{
				"uid": confAttrEdns,
			},
		},
		{
			input: `.:53 {
						policy {
							endpoint 10.2.4.1:5555
							edns0 wrong_hex uid hex string
						}
					}`,
			errContent: errors.New("Could not parse EDNS0 code"),
		},
		{
			input: `.:53 {
						policy {
							endpoint 10.2.4.1:5555
							edns0 0xfff0 uid hex string 32 wrong_offset 32
						}
					}`,
			errContent: errors.New("Could not parse EDNS0 start index"),
		},
		{
			input: `.:53 {
						policy {
							endpoint 10.2.4.1:5555
							edns0 0xfff0 uid hex string 32 0 wrong_size
						}
					}`,
			errContent: errors.New("Could not parse EDNS0 end index"),
		},
		{
			input: `.:53 {
						policy {
							endpoint 10.2.4.1:5555
							edns0 0xfff0 uid hex string 32 0 16
							edns0 0xfff0 id hex string 32 16 32
						}
					}`,
			options: map[uint16][]*edns0Map{
				0xfff0: {
					&edns0Map{
						name:     "uid",
						dataType: typeEDNS0Hex,
						destType: "string",
						size:     32,
						start:    0,
						end:      16},
					&edns0Map{
						name:     "id",
						dataType: typeEDNS0Hex,
						destType: "string",
						size:     32,
						start:    16,
						end:      32},
				},
			},
			confAttrs: map[string]confAttrType{
				"uid": confAttrEdns,
				"id":  confAttrEdns,
			},
		},
		{
			input: `.:53 {
						policy {
							endpoint 10.2.4.1:5555
							edns0 0xfff0 uid hex string 32 16 15
						}
					}`,
			errContent: errors.New("End index should be > start index"),
		},
		{
			input: `.:53 {
						policy {
							endpoint 10.2.4.1:5555
							edns0 0xfff0 uid hex string 32 0 33
						}
					}`,
			errContent: errors.New("End index should be <= size"),
		},
		{
			input: `.:53 {
						policy {
							endpoint 10.2.4.1:5555
							edns0 0xfff1
						}
					}`,
			errContent: errors.New("Invalid edns0 directive"),
		},
		{
			input: `.:53 {
						policy {
							endpoint 10.2.4.1:5555
							edns0 0xfff1 guid bin bin
						}
					}`,
			errContent: errors.New("Could not add EDNS0 map"),
		},
		{
			input: `.:53 {
						policy {
							endpoint 10.2.4.1:5555
							debug_query_suffix
						}
					}`,
			errContent: errors.New("Wrong argument count or unexpected line ending"),
		},
		{
			input: `.:53 {
						policy {
							endpoint 10.2.4.1:5555
							debug_query_suffix debug.local.
						}
					}`,
			debugSuffix: newStringPtr("debug.local."),
		},
		{
			input: `.:53 {
						policy {
							endpoint 10.2.4.1:5555
							transfer policy_id
						}
					}`,
			confAttrs: map[string]confAttrType{
				"policy_id": confAttrTransfer,
			},
		},
		{
			input: `.:53 {
						policy {
							endpoint 10.2.4.1:5555
							edns0 0xfff0 uid hex string 32 0 16
							edns0 0xfff1 id
							transfer policy_id id
							dnstap policy_id query_id
						}
					}`,
			options: map[uint16][]*edns0Map{
				0xfff0: {
					&edns0Map{
						name:     "uid",
						dataType: typeEDNS0Hex,
						destType: "string",
						size:     32,
						start:    0,
						end:      16},
				},
				0xfff1: {
					&edns0Map{
						name:     "id",
						dataType: typeEDNS0Hex,
						destType: "string",
						size:     0,
						start:    0,
						end:      0},
				},
			},
			confAttrs: map[string]confAttrType{
				"policy_id": confAttrTransfer | confAttrDnstap,
				"id":        confAttrEdns | confAttrTransfer,
				"uid":       confAttrEdns,
				"query_id":  confAttrDnstap,
			},
		},
		{
			input: `.:53 {
						policy {
							endpoint 10.2.4.1:5555
							dnstap
						}
					}`,
			errContent: errors.New("Wrong argument count or unexpected line ending"),
		},
		{
			input: `.:53 {
						policy {
							endpoint 10.2.4.1:5555
							transfer
						}
					}`,
			errContent: errors.New("Wrong argument count or unexpected line ending"),
		},
		{
			input: `.:53 {
						policy {
							endpoint 10.2.4.1:5555
							debug_id corednsinstance
						}
					}`,
			ident: newStringPtr("corednsinstance"),
		},
		{
			input: `.:53 {
						policy {
							endpoint 10.2.4.1:5555
							debug_id
						}
					}`,
			errContent: errors.New("Wrong argument count or unexpected line ending"),
		},
		{
			input: `.:53 {
						policy {
							endpoint 10.2.4.1:5555
							passthrough google.com. facebook.org.
						}
					}`,
			passthrough: []string{
				"google.com.",
				"facebook.org.",
			},
		},
		{
			input: `.:53 {
						policy {
							endpoint 10.2.4.1:5555
							passthrough
						}
					}`,
			errContent: errors.New("Wrong argument count or unexpected line ending"),
		},
		{
			input: `.:53 {
						policy {
							endpoint 10.2.4.1:5555
							log
						}
					}`,
		},
		{
			input: `.:53 {
						policy {
							endpoint 10.2.4.1:5555
							log stdout
						}
					}`,
			errContent: errors.New("Wrong argument count or unexpected line ending"),
		},
	}

	for _, test := range tests {
		c := caddy.NewTestController("dns", test.input)
		mw, err := policyParse(c)
		if err != nil {
			if test.errContent != nil {
				if !strings.Contains(err.Error(), test.errContent.Error()) {
					t.Errorf("Expected error '%v' but got '%v'\n", test.errContent, err)
				}
			} else {
				t.Errorf("Expected no error but got '%v'\n", err)
			}
		} else {
			if test.errContent != nil {
				t.Errorf("Expected error '%v' but got 'nil'\n", test.errContent)
			} else {
				if test.endpoints != nil {
					if len(test.endpoints) != len(mw.endpoints) {
						t.Errorf("Expected endpoints %v but got %v\n", test.endpoints, mw.endpoints)
					} else {
						for i := 0; i < len(test.endpoints); i++ {
							if test.endpoints[i] != mw.endpoints[i] {
								t.Errorf("Expected endpoint '%s' but got '%s'\n", test.endpoints[i], mw.endpoints[i])
							}
						}
					}
				}

				if test.options != nil {
					for k, testOpts := range test.options {
						if mwOpts, ok := mw.options[k]; ok {
							if len(testOpts) != len(mwOpts) {
								t.Errorf("Expected %d EDNS0 options for 0x%04x but got %d",
									len(testOpts), k, len(mwOpts))
							} else {
								for i, testOpt := range testOpts {
									mwOpt := mwOpts[i]
									if testOpt.name != mwOpt.name ||
										testOpt.dataType != mwOpt.dataType ||
										testOpt.destType != mwOpt.destType ||
										testOpt.size != mwOpt.size ||
										testOpt.start != mwOpt.start ||
										testOpt.end != mwOpt.end {
										t.Errorf("Expected EDNS0 option:\n\t\"%#v\""+
											"\nfor 0x%04x at %d but got:\n\t\"%#v\"",
											*testOpt, k, i, *mwOpt)
									}
								}
							}
						} else {
							t.Errorf("Expected EDNS0 options 0x%04x but got nothing", k)
						}
					}

					for k := range mw.options {
						if _, ok := test.options[k]; !ok {
							t.Errorf("Got unexpected options 0x%04x", k)
						}
					}
				}

				if test.debugSuffix != nil && *test.debugSuffix != mw.debugSuffix {
					t.Errorf("Expected debug suffix %q but got %q", *test.debugSuffix, mw.debugSuffix)
				}

				if test.confAttrs != nil {
					for k, et := range test.confAttrs {
						at, ok := mw.confAttrs[k]
						if !ok {
							t.Errorf("Missing conf attribute %q", k)
						} else if et != at {
							t.Errorf("Unexpected type of conf attribute %q; expected=%d, actual=%d", k, et, at)
						}
					}

					for k, at := range mw.confAttrs {
						if _, ok := test.confAttrs[k]; !ok {
							t.Errorf("Unexpected conf attribute %q=%d", k, at)
						}
					}
				}

				if test.ident != nil && *test.ident != mw.ident {
					t.Errorf("Expected debug id %q but got %q", *test.ident, mw.ident)
				}

				if test.passthrough != nil {
					if len(test.passthrough) != len(mw.passthrough) {
						t.Errorf("Expected %d passthrough suffixes but got %d",
							len(test.passthrough), len(mw.passthrough))
					} else {
						for i, s := range test.passthrough {
							if s != mw.passthrough[i] {
								t.Errorf("Expected %q passthrough suffix at %d but got %q",
									s, i, mw.passthrough[i])
							}
						}
					}
				}

				if test.connTimeout != nil && *test.connTimeout != mw.connTimeout {
					t.Errorf("Expected connection timeout %s but got %s", *test.connTimeout, mw.connTimeout)
				}
			}
		}
	}
}

func newStringPtr(s string) *string {
	return &s
}

func newIntPtr(n int) *int {
	return &n
}

func newBoolPtr(b bool) *bool {
	return &b
}

func newDurationPtr(d time.Duration) *time.Duration {
	return &d
}
