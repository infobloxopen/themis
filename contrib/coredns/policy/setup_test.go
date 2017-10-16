package policy

import (
	"strings"
	"testing"

	"github.com/mholt/caddy"
)

func TestPolicyConfigParse(t *testing.T) {
	tests := []struct {
		input      string
		endpoints  []string
		errContent string
	}{
		{
			input: `.:53 {
						log stdout
					}`,
			errContent: "Policy setup called without keyword 'policy' in Corefile",
		},
		{
			input: `.:53 {
						policy {
							endpoint
						}
					}`,
			errContent: "Wrong argument count or unexpected line ending",
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
			endpoints:  []string{"10.2.4.1:5555"},
			errContent: "Could not parse EDNS0 data size",
		},
		{
			input: `.:53 {
						policy {
							endpoint 10.2.4.1:5555
							edns0 0xfff0 uid hex string 32 0 32
						}
					}`,
			endpoints: []string{"10.2.4.1:5555"},
		},
		{
			input: `.:53 {
						policy {
							endpoint 10.2.4.1:5555
							edns0 wrong_hex uid hex string
						}
					}`,
			endpoints:  []string{"10.2.4.1:5555"},
			errContent: "Could not parse EDNS0 code",
		},
		{
			input: `.:53 {
						policy {
							endpoint 10.2.4.1:5555
							edns0 0xfff0 uid hex string 32 wrong_offset 32
						}
					}`,
			endpoints:  []string{"10.2.4.1:5555"},
			errContent: "Could not parse EDNS0 start index",
		},
		{
			input: `.:53 {
						policy {
							endpoint 10.2.4.1:5555
							edns0 0xfff0 uid hex string 32 0 wrong_size
						}
					}`,
			endpoints:  []string{"10.2.4.1:5555"},
			errContent: "Could not parse EDNS0 end index",
		},
		{
			input: `.:53 {
						policy {
							endpoint 10.2.4.1:5555
							edns0 0xfff0 uid hex string 32 0 16
							edns0 0xfff0 id hex string 32 16 32
						}
					}`,
			endpoints: []string{"10.2.4.1:5555"},
		},
		{
			input: `.:53 {
						policy {
							endpoint 10.2.4.1:5555
							edns0 0xfff0 uid hex string 32 16 15
						}
					}`,
			endpoints:  []string{"10.2.4.1:5555"},
			errContent: "End index should be > start index",
		},
		{
			input: `.:53 {
						policy {
							endpoint 10.2.4.1:5555
							edns0 0xfff0 uid hex string 32 0 33
						}
					}`,
			endpoints:  []string{"10.2.4.1:5555"},
			errContent: "End index should be <= size",
		},
		{
			input: `.:53 {
						policy {
							endpoint 10.2.4.1:5555
							edns0 0xfff1
						}
					}`,
			endpoints:  []string{"10.2.4.1:5555"},
			errContent: "Invalid edns0 directive",
		},
		{
			input: `.:53 {
						policy {
							endpoint 10.2.4.1:5555
							edns0 0xfff1 guid bin bin
						}
					}`,
			endpoints:  []string{"10.2.4.1:5555"},
			errContent: "Could not add EDNS0 map",
		},
		{
			input: `.:53 {
						policy {
							endpoint 10.2.4.1:5555
							debug_query_suffix
						}
					}`,
			endpoints:  []string{"10.2.4.1:5555"},
			errContent: "Wrong argument count or unexpected line ending",
		},
		{
			input: `.:53 {
						policy {
							endpoint 10.2.4.1:5555
							debug_query_suffix debug.local.
						}
					}`,
			endpoints: []string{"10.2.4.1:5555"},
		},
		{
			input: `.:53 {
						policy {
							endpoint 10.2.4.1:5555
							streams 10
						}
					}`,
			endpoints: []string{"10.2.4.1:5555"},
		},
		{
			input: `.:53 {
						policy {
							endpoint 10.2.4.1:5555
							streams Ten
						}
					}`,
			endpoints:  []string{"10.2.4.1:5555"},
			errContent: "Could not parse number of streams",
		},
		{
			input: `.:53 {
						policy {
							endpoint 10.2.4.1:5555
							streams -1
						}
					}`,
			endpoints:  []string{"10.2.4.1:5555"},
			errContent: "Expected at least one stream got -1",
		},
	}

	for _, test := range tests {
		c := caddy.NewTestController("dns", test.input)
		mw, err := policyParse(c)
		if err != nil {
			if len(test.errContent) > 0 {
				if !strings.Contains(err.Error(), test.errContent) {
					t.Errorf("Expected error '%v' but got '%v'\n", test.errContent, err)
				}
			} else {
				t.Errorf("Expected no error but got '%v'\n", err)
			}
		} else {
			if len(test.errContent) > 0 {
				t.Errorf("Expected error '%v' but got 'nil'\n", test.errContent)
			} else if len(test.endpoints) != len(mw.endpoints) {
				t.Errorf("Expected endpoints %v but got %v\n", test.endpoints, mw.endpoints)
			} else {
				for i := 0; i < len(test.endpoints); i++ {
					if test.endpoints[i] != mw.endpoints[i] {
						t.Errorf("Expected endpoint '%s' but got '%s'\n", test.endpoints[i], mw.endpoints[i])
					}
				}
			}
		}
	}
}
