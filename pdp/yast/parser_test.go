package yast

import (
	"testing"

	"github.com/satori/go.uuid"

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

	simpleUpdate = `# Simple several commands update
- op: add
  path:
  - path
  - to
  - parent policy set
  entity:
    id: Policy Set
    alg: FirstApplicableEffect
    policies:
    - id: Permit Policy
      alg: FirstApplicableEffect
      rules:
      - id: Permit Rule
        effect: permit

- op: add
  path:
  - path
  - to
  - parent policy set
  entity:
    id: Policy
    alg: FirstApplicableEffect
    rules:
    - id: Permit Rule
      effect: permit

- op: add
  path:
  - path
  - to
  - parent policy
  entity:
    id: Permit Rule
    effect: permit

- op: delete
  path:
  - path
  - to
  - useless policy
`
)

func TestUnmarshal(t *testing.T) {
	_, err := Unmarshal([]byte(invalidYAML), nil)
	if err == nil {
		t.Errorf("Expected error for invalid YAML but got nothing")
	}

	_, err = Unmarshal([]byte(invalidRootKeysPolicy), nil)
	if err == nil {
		t.Errorf("Expected error for policy with invalid keys but got nothing")
	} else {
		_, ok := err.(*rootKeysError)
		if !ok {
			t.Errorf("Expected *rootTagsError for policy with invalid keys but got %T (%s)", err, err)
		}
	}

	s, err := Unmarshal([]byte(simpleAllPermitPolicy), nil)
	if err != nil {
		t.Errorf("Expected no error but got %T (%s)", err, err)
	} else {
		p, ok := s.Root().(*pdp.Policy)
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

		r := s.Root().Calculate(&pdp.Context{})
		if r.Effect != pdp.EffectPermit {
			t.Errorf("Expected permit as a response for Simple All Permit Policy but got %d", r.Effect)
		}
	}
}

func TestUnmarshalUpdate(t *testing.T) {
	_, err := UnmarshalUpdate([]byte(simpleUpdate), map[string]pdp.Attribute{}, uuid.NewV4(), uuid.NewV4())
	if err != nil {
		t.Errorf("Expected no error but got %T (%s)", err, err)
	}
}
