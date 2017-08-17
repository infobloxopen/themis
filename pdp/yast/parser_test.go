package yast

import (
	"fmt"
	"testing"

	"github.com/google/uuid"

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

	policyToUpdate = `# Policy to update
attributes:
  a: string
  b: string
  r: string
policies:
  id: Parent policy set
  alg:
    id: mapper
    map:
      attr: a
    default: Deny policy
  policies:
  - id: Deny policy
    alg: FirstApplicableEffect
    rules:
    - effect: Deny
      obligations:
      - r: Default Deny Policy
  - id: Parent policy
    alg:
      id: mapper
      map:
        attr: b
      default: Deny rule
    rules:
    - id: Deny rule
      effect: Deny
      obligations:
      - r: Default Deny rule
    - id: Some rule
      effect: Permit
      obligations:
      - r: Some rule
  - id: Useless policy
    alg: FirstApplicableEffect
    rules:
    - effect: Deny
      obligations:
      - r: Useless policy
`

	simpleUpdate = `# Simple several commands update
- op: add
  path:
  - Parent policy set
  entity:
    id: Policy Set
    alg: FirstApplicableEffect
    policies:
    - id: Permit Policy
      alg: FirstApplicableEffect
      rules:
      - id: Permit Rule
        effect: permit
        obligations:
        - r: First Added Update Item

- op: add
  path:
  - Parent policy set
  entity:
    id: Policy
    alg: FirstApplicableEffect
    rules:
    - id: Permit Rule
      effect: permit
      obligations:
      - r: Second Added Update Item

- op: add
  path:
  - Parent policy set
  - Parent policy
  entity:
    id: Permit Rule
    effect: permit
    obligations:
    - r: Third Added Update Item

- op: delete
  path:
  - Parent policy set
  - Useless policy
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
	tag := uuid.New()
	s, err := Unmarshal([]byte(policyToUpdate), &tag)
	if err != nil {
		t.Errorf("Expected no error but got %T (%s)", err, err)
		return
	}

	attrs := map[string]string{
		"a": "Parent policy",
		"b": "Some rule"}
	assertPolicy(s, attrs, "Some rule", "\"some rule\"", t)

	attrs = map[string]string{"a": "Useless policy"}
	assertPolicy(s, attrs, "Useless policy", "\"useless policy\"", t)

	tr, err := s.NewTransaction(&tag)
	if err != nil {
		t.Errorf("Expected no error but got %T (%s)", err, err)
		return
	}

	u, err := UnmarshalUpdate([]byte(simpleUpdate), tr.Attributes(), tag, uuid.New())
	if err != nil {
		t.Errorf("Expected no error but got %T (%s)", err, err)
		return
	}

	err = tr.Apply(u)
	if err != nil {
		t.Errorf("Expected no error but got %T (%s)", err, err)
		return
	}

	s, err = tr.Commit()
	if err != nil {
		t.Errorf("Expected no error but got %T (%s)", err, err)
		return
	}

	attrs = map[string]string{"a": "Policy Set"}
	assertPolicy(s, attrs, "First Added Update Item", "\"new policy set\"", t)

	attrs = map[string]string{"a": "Policy"}
	assertPolicy(s, attrs, "Second Added Update Item", "\"new policy\"", t)

	attrs = map[string]string{
		"a": "Parent policy",
		"b": "Permit Rule"}
	assertPolicy(s, attrs, "Third Added Update Item", "\"new nested policy set\"", t)

	attrs = map[string]string{"a": "Useless policy"}
	assertPolicy(s, attrs, "Default Deny Policy", "\"deleted useless policy\"", t)
}

func assertPolicy(s *pdp.PolicyStorage, attrs map[string]string, e, desc string, t *testing.T) {
	ctx, err := newStringContext(attrs)
	if err != nil {
		t.Errorf("Expected no error for %s but got %T (%s)", desc, err, err)
		return
	}

	_, o, err := s.Root().Calculate(ctx).Status()
	if err != nil {
		t.Errorf("Expected no error for %s but got %T (%s)", desc, err, err)
		return
	}

	if len(o) < 1 {
		t.Errorf("Expected at least one obligation for %s but got nothing", desc)
		return
	}

	_, _, v, err := o[0].Serialize(ctx)
	if err != nil {
		t.Errorf("Expected no error for %s but got %T (%s)", desc, err, err)
		return
	}

	if v != e {
		t.Errorf("Expected %q for %s but got %q", e, desc, v)
	}
}

func newStringContext(m map[string]string) (*pdp.Context, error) {
	names := make([]string, len(m))
	values := make([]string, len(m))
	i := 0
	for k, v := range m {
		names[i] = k
		values[i] = v
		i++
	}

	return pdp.NewContext(nil, len(m), func(i int) (string, pdp.AttributeValue, error) {
		if i >= len(names) {
			return "", pdp.AttributeValue{}, fmt.Errorf("No attribute name for index %d", i)
		}
		n := names[i]

		if i >= len(values) {
			return "", pdp.AttributeValue{}, fmt.Errorf("No attribute value for index %d", i)
		}
		v := values[i]

		return n, pdp.MakeStringValue(v), nil
	})
}
