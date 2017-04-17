package pdp

import "testing"

const (
	YASTTestAlgPolicySet = `# Test policy set
id: Test
alg:
  id: Mapper
  map:
    attr: x

  alg:
    id: Mapper
    map:
      attr: "y"

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

  - id: Permit
    alg: FirstApplicableEffect
    rules:
      - id: Default
        effect: Permit
`

	YASTTestAlgPolicy = `# Test policy
id: Test
alg:
  id: Mapper
  map:
    attr: x

  alg:
    id: Mapper
    map:
      attr: "y"

    alg: FirstApplicableEffect

  error: Error
  default: Default

rules:
  - id: Default
    effect: Deny

  - id: Error
    effect: Deny

  - id: Permit
    effect: Permit`
)

var (
	YASTPoliciesTestAttrs = map[string]AttributeType{
		"x": AttributeType{ID: "x", DataType: DataTypeSetOfStrings},
		"y": AttributeType{ID: "y", DataType: DataTypeSetOfStrings}}

	YASTPoliciesTestContent = map[string]interface{}{}
)

func TestUnmarshalYASTPolicySetWithNestedMappers(t *testing.T) {
	c, v := prepareTestYAST(YASTTestAlgPolicySet, YASTPoliciesTestAttrs, YASTPoliciesTestContent, t)

	p, err := c.unmarshalItem(v)
	if err != nil {
		t.Fatalf("Expected no errors but got:\n%#v\n\n%s\n", err, err)
	}

	if p == nil {
		t.Fatalf("Expected result but got nothing")
	}

	set, ok := p.(PolicySetType)
	if !ok {
		t.Fatalf("Expected policy set but got %T", p)
	}

	if set.AlgParams == nil {
		t.Fatalf("Expected algorithm parameters but got nothing")
	}

	params, ok := set.AlgParams.(MapperPCAParams)
	if !ok {
		t.Fatalf("Expected mapper algorithm parameters but got %T", set.AlgParams)
	}

	if len(params.PoliciesMap) != len(set.Policies) {
		t.Errorf("Expected filled policies map (with %d policies) but got %d", len(set.Policies), len(params.PoliciesMap))
	}

	ID := params.DefaultPolicy.getID()
	if ID != "Default" {
		t.Errorf("Expected %q policy as default but got %q", "Default", ID)
	}

	ID = params.ErrorPolicy.getID()
	if ID != "Error" {
		t.Errorf("Expected %q policy as \"on error\" policy but got %q", "Error", ID)
	}

	if params.AlgParams == nil {
		t.Fatalf("Expected subalgorithm parameters but got nothing")
	}

	subParams, ok := params.AlgParams.(MapperPCAParams)
	if !ok {
		t.Fatalf("Expected mapper subalgorithm parameters but got %T", params.AlgParams)
	}

	if len(subParams.PoliciesMap) > 0 {
		t.Errorf("Expected empty policies map for subalgorithm but got map with %d elements", len(subParams.PoliciesMap))
	}

	if subParams.DefaultPolicy != nil {
		t.Errorf("Expected no default policy but got %q", subParams.DefaultPolicy.getID())
	}

	if subParams.ErrorPolicy != nil {
		t.Errorf("Expected no \"on error\" policy but got %q", subParams.ErrorPolicy.getID())
	}
}

func TestUnmarshalYASTPolicyWithNestedMappers(t *testing.T) {
	c, v := prepareTestYAST(YASTTestAlgPolicy, YASTPoliciesTestAttrs, YASTPoliciesTestContent, t)

	p, err := c.unmarshalItem(v)
	if err != nil {
		t.Fatalf("Expected no errors but got:\n%#v\n\n%s\n", err, err)
	}

	if p == nil {
		t.Fatalf("Expected result but got nothing")
	}

	pol, ok := p.(PolicyType)
	if !ok {
		t.Fatalf("Expected policy but got %T", p)
	}

	if pol.AlgParams == nil {
		t.Fatalf("Expected algorithm parameters but got nothing")
	}

	params, ok := pol.AlgParams.(MapperRCAParams)
	if !ok {
		t.Fatalf("Expected mapper algorithm parameters but got %T", pol.AlgParams)
	}

	if len(params.RulesMap) != len(pol.Rules) {
		t.Errorf("Expected filled rules map (with %d rules) but got %d", len(pol.Rules), len(params.RulesMap))
	}

	ID := params.DefaultRule.ID
	if ID != "Default" {
		t.Errorf("Expected %q rule as default but got %q", "Default", ID)
	}

	ID = params.ErrorRule.ID
	if ID != "Error" {
		t.Errorf("Expected %q rule as \"on error\" rule but got %q", "Error", ID)
	}

	if params.AlgParams == nil {
		t.Fatalf("Expected subalgorithm parameters but got nothing")
	}

	subParams, ok := params.AlgParams.(MapperRCAParams)
	if !ok {
		t.Fatalf("Expected mapper subalgorithm parameters but got %T", params.AlgParams)
	}

	if len(subParams.RulesMap) > 0 {
		t.Errorf("Expected empty rules map for subalgorithm but got map with %d elements", len(subParams.RulesMap))
	}

	if subParams.DefaultRule != nil {
		t.Errorf("Expected no default rule but got %q", subParams.DefaultRule.ID)
	}

	if subParams.ErrorRule != nil {
		t.Errorf("Expected no \"on error\" rule but got %q", subParams.ErrorRule.ID)
	}
}
