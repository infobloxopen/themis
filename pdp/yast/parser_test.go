package yast

import (
	"testing"

	"github.com/infobloxopen/themis/pdp"
)

const (
	invalidYAML = `# Invalid YAML
x:
- one
+ two
- three
`

	invalidRootKeysPolicy = `# Policy with invalid keys
attributes:
  x: string

invalid:
- first

policies:
  id: Default
  alg: FirstApplicableEffect
  rules:
  - effect: Permit
`

	simpleAllPermitPolicy = `# Simple All Permit Policy
policies:
  id: Default
  alg: FirstApplicableEffect
  rules:
  - effect: Permit
`
)

func TestUnmarshalYAST(t *testing.T) {
	_, err := UnmarshalYAST([]byte(invalidYAML))
	if err == nil {
		t.Errorf("Expected error for invalid YAML but got nothing")
	}

	_, err = UnmarshalYAST([]byte(invalidRootKeysPolicy))
	if err == nil {
		t.Errorf("Expected error for policy with invalid keys but got nothing")
	} else {
		_, ok := err.(*rootKeysError)
		if !ok {
			t.Errorf("Expected *rootTagsError for policy with invalid keys but got %T (%s)", err, err)
		}
	}

	v, err := UnmarshalYAST([]byte(simpleAllPermitPolicy))
	if err != nil {
		t.Errorf("Expected no error but got %T (%s)", err, err)
	} else {
		p, ok := v.(*pdp.Policy)
		if !ok {
			t.Errorf("Expected policy as root item in Simple All Permit Policy but got %T", p)
		} else {
			PID, ok := p.GetID()
			if !ok {
				t.Errorf("Expected %q as Simple All Permit Policy ID but got hidden policy", "Default")
			} else if PID != "Default" {
				t.Errorf("Expected %q as Simple All Permit Policy ID but got %q", "Default", PID)
			}
		}
	}

	r := v.Calculate(&pdp.Context{})
	if r.Effect != pdp.EffectPermit {
		t.Errorf("Expected permit as a response for Simple All Permit Policy but got %d", r.Effect)
	}
}
