package pdp

import "testing"

func TestPolicy(t *testing.T) {
	c := &Context{
		a: map[string]map[int]AttributeValue{
			"missing-type": {
				TypeBoolean: MakeBooleanValue(false)},
			"test-string": {
				TypeString: MakeStringValue("test")},
			"example-string": {
				TypeString: MakeStringValue("example")}}}

	testID := "Test Policy"

	p := Policy{
		id:        testID,
		algorithm: firstApplicableEffectRCA{}}
	ID, ok := p.GetID()
	if !ok {
		t.Errorf("Expected policy ID %q but got hidden policy", testID)
	} else if ID != testID {
		t.Errorf("Expected policy ID %q but got %q", testID, ID)
	}

	r := p.Calculate(c)
	if r.Effect != EffectNotApplicable {
		t.Errorf("Expected %q for empty policy but got %q",
			effectNames[EffectNotApplicable], effectNames[r.Effect])
	}

	p = Policy{
		id:        testID,
		target:    makeSimpleStringTarget("missing", "test"),
		algorithm: firstApplicableEffectRCA{}}

	r = p.Calculate(c)
	if r.Effect != EffectNotApplicable {
		t.Errorf("Expected %q for policy with FirstApplicableEffectRCA and not found attribute but got %q",
			effectNames[EffectNotApplicable], effectNames[r.Effect])
	}

	_, ok = r.status.(*missingAttributeError)
	if !ok {
		t.Errorf("Expected missing attribute status for policy with FirstApplicableEffectRCA and "+
			"not found attribute but got %T (%s)", r.status, r.status)
	}

	p = Policy{
		id:        testID,
		target:    makeSimpleStringTarget("missing-type", "test"),
		algorithm: firstApplicableEffectRCA{}}

	r = p.Calculate(c)
	if r.Effect != EffectNotApplicable {
		t.Errorf("Expected %q for policy with FirstApplicableEffectRCA and attribute with wrong type but got %q",
			effectNames[EffectNotApplicable], effectNames[r.Effect])
	}

	_, ok = r.status.(*missingAttributeError)
	if !ok {
		t.Errorf("Expected missing attribute status for policy with FirstApplicableEffectRCA and "+
			"attribute with wrong type but got %T (%s)", r.status, r.status)
	}

	p = Policy{
		id:        testID,
		target:    makeSimpleStringTarget("example-string", "test"),
		algorithm: firstApplicableEffectRCA{}}

	r = p.Calculate(c)
	if r.Effect != EffectNotApplicable {
		t.Errorf("Expected %q for policy with FirstApplicableEffectRCA and attribute with not maching value but got %q",
			effectNames[EffectNotApplicable], effectNames[r.Effect])
	}

	if r.status != nil {
		t.Errorf("Expected no error status for policy with FirstApplicableEffectRCA and "+
			"attribute with not maching value but got %T (%s)", r.status, r.status)
	}

	p = Policy{
		id:          testID,
		target:      makeSimpleStringTarget("test-string", "test"),
		rules:       []*Rule{{effect: EffectPermit}},
		obligations: makeSingleStringObligation("obligation", "test"),
		algorithm:   firstApplicableEffectRCA{}}

	r = p.Calculate(c)
	if r.Effect != EffectPermit {
		t.Errorf("Expected %q for policy with rule and obligations but got %q",
			effectNames[EffectPermit], effectNames[r.Effect])
	}

	if r.status != nil {
		t.Errorf("Expected no error status for policy with rule and obligations but got %T (%s)",
			r.status, r.status)
	}

	if len(r.obligations) != 1 {
		t.Errorf("Expected single obligation for with rule and obligations but got %#v", r.obligations)
	}

	defaultRule := &Rule{
		id:     "Default",
		effect: EffectDeny}
	errorRule := &Rule{
		id:     "Error",
		effect: EffectDeny}
	permitRule := &Rule{
		id:     "Permit",
		effect: EffectPermit}
	p = Policy{
		id:    testID,
		rules: []*Rule{defaultRule, errorRule, permitRule},
		algorithm: makeMapperRCA(
			[]*Rule{defaultRule, errorRule, permitRule},
			MapperRCAParams{
				Argument: AttributeDesignator{a: Attribute{id: "x", t: TypeSetOfStrings}},
				DefOk:    true,
				Def:      "Default",
				ErrOk:    true,
				Err:      "Error",
				Algorithm: makeMapperRCA(
					nil,
					MapperRCAParams{Argument: AttributeDesignator{a: Attribute{id: "y", t: TypeString}}})})}

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

func makeSimpleStringTarget(ID, value string) Target {
	return Target{a: []AnyOf{{a: []AllOf{{m: []Match{{
		m: functionStringEqual{
			first:  AttributeDesignator{a: Attribute{id: ID, t: TypeString}},
			second: MakeStringValue(value)}}}}}}}}
}

func makeSingleStringObligation(ID, value string) []AttributeAssignmentExpression {
	return []AttributeAssignmentExpression{
		{
			a: Attribute{id: ID, t: TypeString},
			e: MakeStringValue(value)}}
}
