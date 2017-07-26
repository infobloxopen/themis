package pdp

import "testing"

func TestPolicySet(t *testing.T) {
	c := &Context{
		a: map[string]map[int]AttributeValue{
			"missing-type": {
				TypeBoolean: MakeBooleanValue(false)},
			"test-string": {
				TypeString: MakeStringValue("test")},
			"example-string": {
				TypeString: MakeStringValue("example")}}}

	testID := "Test Policy"

	p := PolicySet{
		id:        testID,
		algorithm: firstApplicableEffectPCA{}}
	ID, ok := p.GetID()
	if !ok {
		t.Errorf("Expected policy set ID %q but got hidden policy set", testID)
	} else if ID != testID {
		t.Errorf("Expected policy set ID %q but got %q", testID, ID)
	}

	r := p.Calculate(c)
	if r.Effect != EffectNotApplicable {
		t.Errorf("Expected %q for empty policy set but got %q",
			effectNames[EffectNotApplicable], effectNames[r.Effect])
	}

	p = PolicySet{
		id:        testID,
		target:    makeSimpleStringTarget("missing", "test"),
		algorithm: firstApplicableEffectPCA{}}

	r = p.Calculate(c)
	if r.Effect != EffectNotApplicable {
		t.Errorf("Expected %q for policy set with FirstApplicableEffectPCA and not found attribute but got %q",
			effectNames[EffectNotApplicable], effectNames[r.Effect])
	}

	_, ok = r.status.(*missingAttributeError)
	if !ok {
		t.Errorf("Expected missing attribute status for policy set with FirstApplicableEffectPCA and "+
			"not found attribute but got %T (%s)", r.status, r.status)
	}

	p = PolicySet{
		id:        testID,
		target:    makeSimpleStringTarget("missing-type", "test"),
		algorithm: firstApplicableEffectPCA{}}

	r = p.Calculate(c)
	if r.Effect != EffectNotApplicable {
		t.Errorf("Expected %q for policy set with FirstApplicableEffectPCA and attribute with wrong type but got %q",
			effectNames[EffectNotApplicable], effectNames[r.Effect])
	}

	_, ok = r.status.(*missingAttributeError)
	if !ok {
		t.Errorf("Expected missing attribute status for policy with FirstApplicableEffectPCA and "+
			"attribute with wrong type but got %T (%s)", r.status, r.status)
	}

	p = PolicySet{
		id:        testID,
		target:    makeSimpleStringTarget("example-string", "test"),
		algorithm: firstApplicableEffectPCA{}}

	r = p.Calculate(c)
	if r.Effect != EffectNotApplicable {
		t.Errorf("Expected %q for policy set with FirstApplicableEffectPCA and "+
			"attribute with not maching value but got %q",
			effectNames[EffectNotApplicable], effectNames[r.Effect])
	}

	if r.status != nil {
		t.Errorf("Expected no error status for policy set with FirstApplicableEffectPCA and "+
			"attribute with not maching value but got %T (%s)", r.status, r.status)
	}

	p = PolicySet{
		id:     testID,
		target: makeSimpleStringTarget("test-string", "test"),
		policies: []Evaluable{
			&Policy{
				rules:     []*Rule{{effect: EffectPermit}},
				algorithm: firstApplicableEffectRCA{}}},
		obligations: makeSingleStringObligation("obligation", "test"),
		algorithm:   firstApplicableEffectPCA{}}

	r = p.Calculate(c)
	if r.Effect != EffectPermit {
		t.Errorf("Expected %q for policy with rule and obligations but got %q",
			effectNames[EffectPermit], effectNames[r.Effect])
	}

	if r.status != nil {
		t.Errorf("Expected no error status for policy rule and obligations but got %T (%s)",
			r.status, r.status)
	}

	defaultPolicy := &Policy{
		id:        "Default",
		rules:     []*Rule{{effect: EffectDeny}},
		algorithm: firstApplicableEffectRCA{}}
	errorPolicy := &Policy{
		id:        "Error",
		rules:     []*Rule{{effect: EffectDeny}},
		algorithm: firstApplicableEffectRCA{}}
	permitPolicy := &Policy{
		id:        "Permit",
		rules:     []*Rule{{effect: EffectPermit}},
		algorithm: firstApplicableEffectRCA{}}
	p = PolicySet{
		id:       testID,
		policies: []Evaluable{defaultPolicy, errorPolicy, permitPolicy},
		algorithm: makeMapperPCA(
			[]Evaluable{defaultPolicy, errorPolicy, permitPolicy},
			MapperPCAParams{
				Argument:  AttributeDesignator{a: Attribute{id: "x", t: TypeSetOfStrings}},
				DefOk:     true,
				Def:       "Default",
				ErrOk:     true,
				Err:       "Error",
				Algorithm: makeMapperPCA(
					nil,
					MapperPCAParams{
						Argument: AttributeDesignator{a: Attribute{id: "y", t: TypeString}}})})}

	c = &Context{
		a: map[string]map[int]AttributeValue{
			"x": {TypeSetOfStrings: MakeSetOfStringsValue(newStrTree("Permit", "Default"))},
			"y": {TypeString: MakeStringValue("Permit")}}}

	r = p.Calculate(c)
	if r.Effect != EffectPermit {
		t.Errorf("Expected %q for policy with rule and obligations but got %q",
			effectNames[EffectPermit], effectNames[r.Effect])
	}

	if r.status != nil {
		t.Errorf("Expected no error status for policy rule and obligations but got %T (%s)",
			r.status, r.status)
	}

	c = &Context{
		a: map[string]map[int]AttributeValue{
			"x": {TypeSetOfStrings: MakeSetOfStringsValue(newStrTree("Permit", "Default"))},
			"y": {TypeSetOfStrings: MakeSetOfStringsValue(newStrTree("Permit", "Default"))}}}

	r = p.Calculate(c)
	if r.Effect != EffectIndeterminate {
		t.Errorf("Expected %q for policy with rule and obligations but got %q",
			effectNames[EffectIndeterminate], effectNames[r.Effect])
	}

	_, ok = r.status.(*missingAttributeError)
	if !ok {
		t.Errorf("Expected missing attribute status for policy with rule and obligations but got %T (%s)",
			r.status, r.status)
	}
}
