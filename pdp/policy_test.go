package pdp

import "testing"

func TestPolicy(t *testing.T) {
	c := &Context{
		a: map[string]map[int]attributeValue{
			"missing-type": {
				typeBoolean: makeBooleanValue(false)},
			"test-string": {
				typeString: makeStringValue("test")},
			"example-string": {
				typeString: makeStringValue("example")}}}

	testID := "Test Policy"

	p := Policy{
		id:        testID,
		algorithm: firstApplicableEffectRCA{}}
	ID := p.GetID()
	if ID != testID {
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

	_, ok := r.status.(*missingAttributeError)
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
		rules:       []*rule{{effect: EffectPermit}},
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

	defaultRule := &rule{
		id:     "Default",
		effect: EffectDeny}
	errorRule := &rule{
		id:     "Error",
		effect: EffectDeny}
	permitRule := &rule{
		id:     "Permit",
		effect: EffectPermit}
	p = Policy{
		id:    testID,
		rules: []*rule{defaultRule, errorRule, permitRule},
		algorithm: mapperRCA{
			argument: attributeDesignator{a: attribute{id: "x", t: typeSetOfStrings}},
			rules: map[string]*rule{
				"Default": defaultRule,
				"Error":   errorRule,
				"Permit":  permitRule},
			def: defaultRule,
			err: errorRule,
			algorithm: mapperRCA{
				argument: attributeDesignator{a: attribute{id: "y", t: typeString}}}}}

	c = &Context{
		a: map[string]map[int]attributeValue{
			"x": {typeSetOfStrings: makeSetOfStringsValue(newStrTree("Permit", "Default"))},
			"y": {typeString: makeStringValue("Permit")}}}

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
		a: map[string]map[int]attributeValue{
			"x": {typeSetOfStrings: makeSetOfStringsValue(newStrTree("Permit", "Default"))},
			"y": {typeSetOfStrings: makeSetOfStringsValue(newStrTree("Permit", "Default"))}}}

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

func makeSimpleStringTarget(ID, value string) target {
	return target{a: []anyOf{{a: []allOf{{m: []match{{
		m: functionStringEqual{
			first:  attributeDesignator{a: attribute{id: ID, t: typeString}},
			second: makeStringValue(value)}}}}}}}}
}

func makeSingleStringObligation(ID, value string) []attributeAssignmentExpression {
	return []attributeAssignmentExpression{
		{
			a: attribute{id: ID, t: typeString},
			e: makeStringValue(value)}}
}
