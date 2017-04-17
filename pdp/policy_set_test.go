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
)

var (
	PolicySetTestAttrs = map[string]AttributeType{
		"x": AttributeType{ID: "x", DataType: DataTypeSetOfStrings},
		"y": AttributeType{ID: "y", DataType: DataTypeString}}
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
