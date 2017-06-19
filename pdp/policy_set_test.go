package pdp

import "testing"

const (
	TestAlgPolicySet = `# Test policy set
id: Test
alg:
  id: Mapper
  map:
    attr: x

  alg:
    id: Mapper
    map:
      attr: "y"

  error: Error
  default: Default

policies:
  - id: Default
    alg: FirstApplicableEffect
    rules:
      - id: Default
        effect: Deny

  - id: Error
    alg: FirstApplicableEffect
    rules:
      - id: Default
        effect: Deny

  - id: Permit
    alg: FirstApplicableEffect
    rules:
      - id: Default
        effect: Permit
`

	TestAlgPolicySetListOfStringsSelector = `# Test policy set with list of string selector in mapper
id: Test
alg:
  id: Mapper
  map:
    selector:
      type: List of Strings
      path:
        - test
        - attr: "y"
      content: test
  alg: FirstApplicableEffect

  error: Error
  default: Default

policies:
  - id: Default
    alg: FirstApplicableEffect
    rules:
      - id: Default
        effect: Deny

  - id: Error
    alg: FirstApplicableEffect
    rules:
      - id: Default
        effect: Deny

  - id: Permit-One
    alg: FirstApplicableEffect
    rules:
      - id: Default
        effect: Permit
        obligations:
          - z: One

  - id: Permit-Two
    alg: FirstApplicableEffect
    rules:
      - id: Default
        effect: Permit
        obligations:
          - z: Two
`
)

var (
	PolicySetTestAttrs = map[string]AttributeType{
		"x": {ID: "x", DataType: DataTypeSetOfStrings},
		"y": {ID: "y", DataType: DataTypeString},
		"z": {ID: "z", DataType: DataTypeString}}

	YASTSelectorTestContentWithListOfStrings = map[string]interface{}{
		"test": map[string]interface{}{
			"test": map[string]interface{}{
				"test":    []interface{}{"Permit-One", "Permit-Two"},
				"example": []interface{}{"Permit-Two", "Permit-One"}}}}
)

func TestPolicySetWithNestedMappers(t *testing.T) {
	c, v := prepareTestYAST(YASTTestAlgPolicySet, YASTPoliciesTestAttrs, YASTPoliciesTestContent, t)

	p, err := c.unmarshalItem(v)
	if err != nil {
		t.Errorf("Expected no errors but got:\n%#v\n\n%s\n", err, err)
	} else {
		if p == nil {
			t.Errorf("Expected result but got nothing")
		} else {
			ctx := NewContext()
			ctx.StoreAttribute("x", DataTypeSetOfStrings, map[string]int{"Permit": 0, "Default": 1})
			ctx.StoreAttribute("y", DataTypeSetOfStrings, map[string]int{"Permit": 0})

			r := p.Calculate(&ctx)
			if r.Effect != EffectPermit {
				t.Errorf("Expected %s effect but got %s", EffectNames[EffectPermit], EffectNames[r.Effect])
			}
		}
	}

	c, v = prepareTestYAST(TestAlgPolicySet, PolicySetTestAttrs, YASTPoliciesTestContent, t)

	p, err = c.unmarshalItem(v)
	if err != nil {
		t.Errorf("Expected no errors but got:\n%#v\n\n%s\n", err, err)
	} else {
		if p == nil {
			t.Errorf("Expected result but got nothing")
		} else {
			ctx := NewContext()
			ctx.StoreAttribute("x", DataTypeSetOfStrings, map[string]int{"Permit": 0, "Default": 1})
			ctx.StoreAttribute("y", DataTypeString, "Permit")

			r := p.Calculate(&ctx)
			if r.Effect != EffectPermit {
				t.Errorf("Expected %s effect but got %s", EffectNames[EffectPermit], EffectNames[r.Effect])
			}
		}
	}
}

func TestPolicySetWithSelectorOnListOfStringsMapper(t *testing.T) {
	c, v := prepareTestYAST(TestAlgPolicySetListOfStringsSelector, PolicySetTestAttrs, YASTSelectorTestContentWithListOfStrings, t)

	p, err := c.unmarshalItem(v)
	if err != nil {
		t.Errorf("Expected no errors but got:\n%#v\n\n%s\n", err, err)
	} else {
		if p == nil {
			t.Errorf("Expected result but got nothing")
		} else {
			ctx := NewContext()
			ctx.StoreAttribute("y", DataTypeString, "test")

			r := p.Calculate(&ctx)
			if r.Effect != EffectPermit {
				t.Errorf("Expected %s effect but got %s", EffectNames[EffectPermit], EffectNames[r.Effect])
			}

			assertStringObligation(r, "z", "One", t)

			ctx = NewContext()
			ctx.StoreAttribute("y", DataTypeString, "example")

			r = p.Calculate(&ctx)
			if r.Effect != EffectPermit {
				t.Errorf("Expected %s effect but got %s", EffectNames[EffectPermit], EffectNames[r.Effect])
			}

			assertStringObligation(r, "z", "Two", t)
		}
	}
}

func assertStringObligation(r ResponseType, ID, value string, t *testing.T) {
	if len(r.Obligations) != 1 {
		t.Errorf("Expected one obligation but got %d", len(r.Obligations))
		return
	}

	o := r.Obligations[0]
	if o.Attribute.ID != ID {
		t.Errorf("Expected obligation for %q attribute but got %q", ID, o.Attribute.ID)
	}

	v, ok := o.Expression.(AttributeValueType)
	if !ok {
		t.Errorf("Expected attribute value as obligation expression but got %s",
			o.Expression.describe())
	} else {
		s, err := ExtractStringValue(v, "obligation value")
		if err != nil {
			t.Errorf("Expected no errors but got:\n%#v\n\n%s\n", err, err)
		} else {
			if s != value {
				t.Errorf("Expected %q as obligation value but got %q", value, s)
			}
		}
	}
}
