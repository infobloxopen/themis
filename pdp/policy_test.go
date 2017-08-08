package pdp

import (
	"fmt"
	"testing"

	"github.com/pmezard/go-difflib/difflib"
)

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

func TestPolicyAppend(t *testing.T) {
	p := &Policy{id: "test", algorithm: firstApplicableEffectRCA{}}
	newP, err := p.Append([]string{}, &Rule{id: "test", effect: EffectPermit})
	if err != nil {
		t.Errorf("Expected no error but got %s", err)
	} else {
		if p == newP {
			t.Errorf("Expected different new policy but got the same")
		}
	}

	p = &Policy{hidden: true, algorithm: firstApplicableEffectRCA{}}
	newP, err = p.Append([]string{}, &Rule{id: "test", effect: EffectPermit})
	if err == nil {
		t.Errorf("Expected error but got policy %#v", newP)
	} else if _, ok := err.(*hiddenPolicyModificationError); !ok {
		t.Errorf("Expected *hiddenPolicyModificationError but got %T (%s)", err, err)
	}

	p = &Policy{id: "test", algorithm: firstApplicableEffectRCA{}}
	newP, err = p.Append([]string{"test"}, &Rule{id: "test", effect: EffectPermit})
	if err == nil {
		t.Errorf("Expected error but got policy %#v", newP)
	} else if _, ok := err.(*tooLongPathPolicyModificationError); !ok {
		t.Errorf("Expected *tooLongPathPolicyModificationError but got %T (%s)", err, err)
	}

	p = &Policy{id: "test", algorithm: firstApplicableEffectRCA{}}
	newP, err = p.Append([]string{}, &Policy{id: "test", algorithm: firstApplicableEffectRCA{}})
	if err == nil {
		t.Errorf("Expected error but got policy %#v", newP)
	} else if _, ok := err.(*invalidPolicyItemTypeError); !ok {
		t.Errorf("Expected *invalidPolicyItemTypeError but got %T (%s)", err, err)
	}

	p = &Policy{
		id: "test",
		rules: []*Rule{
			{id: "first", effect: EffectPermit},
			{id: "second", effect: EffectPermit}},
		algorithm: firstApplicableEffectRCA{}}
	newP, err = p.Append([]string{}, &Rule{id: "third", effect: EffectPermit})
	if err != nil {
		t.Errorf("Expected no error but got %s", err)
	} else {
		if p, ok := newP.(*Policy); ok {
			if len(p.rules) == 3 {
				if p.rules[2].id != "third" {
					t.Errorf("Expected \"third\" rule added to the end but got %q", p.rules[2].id)
				}
			} else {
				t.Errorf("Expected three rules after append but got %d", len(p.rules))
			}
		} else {
			t.Errorf("Expected new policy but got %T (%#v)", newP, newP)
		}
	}

	newP, err = newP.Append([]string{}, &Rule{id: "first", effect: EffectDeny})
	if err != nil {
		t.Errorf("Expected no error but got %s", err)
	} else {
		if p, ok := newP.(*Policy); ok {
			if len(p.rules) == 3 {
				if p.rules[0].id != "first" {
					t.Errorf("Expected \"first\" rule replaced at the begining but got %q", p.rules[0].id)
				} else if p.rules[0].effect != EffectDeny {
					t.Errorf("Expected \"first\" rule became deny but it's still %s", effectNames[p.rules[0].effect])
				}
			} else {
				t.Errorf("Expected three rules after append but got %d", len(p.rules))
			}
		} else {
			t.Errorf("Expected new policy but got %T (%#v)", newP, newP)
		}
	}

	newP, err = newP.Append([]string{}, &Rule{id: "second", effect: EffectDeny})
	if err != nil {
		t.Errorf("Expected no error but got %s", err)
	} else {
		if p, ok := newP.(*Policy); ok {
			if len(p.rules) == 3 {
				if p.rules[1].id != "second" {
					t.Errorf("Expected \"second\" rule replaced at the middle but got %q", p.rules[1].id)
				} else if p.rules[1].effect != EffectDeny {
					t.Errorf("Expected \"second\" rule became deny but it's still %s", effectNames[p.rules[1].effect])
				}
			} else {
				t.Errorf("Expected three rules after append but got %d", len(p.rules))
			}
		} else {
			t.Errorf("Expected new policy but got %T (%#v)", newP, newP)
		}
	}

	newP, err = newP.Append([]string{}, &Rule{id: "third", effect: EffectDeny})
	if err != nil {
		t.Errorf("Expected no error but got %s", err)
	} else {
		if p, ok := newP.(*Policy); ok {
			if len(p.rules) == 3 {
				if p.rules[2].id != "third" {
					t.Errorf("Expected \"third\" rule replaced at the end but got %q", p.rules[2].id)
				} else if p.rules[2].effect != EffectDeny {
					t.Errorf("Expected \"third\" rule became deny but it's still %s", effectNames[p.rules[2].effect])
				}
			} else {
				t.Errorf("Expected three rules after append but got %d", len(p.rules))
			}
		} else {
			t.Errorf("Expected new policy but got %T (%#v)", newP, newP)
		}
	}

	p = NewPolicy("test", false, Target{},
		[]*Rule{
			{id: "first", effect: EffectPermit},
			{id: "second", effect: EffectPermit},
			{id: "third", effect: EffectPermit}},
		makeMapperRCA, MapperRCAParams{
			Argument: AttributeDesignator{a: Attribute{id: "k", t: TypeString}},
			DefOk:    true,
			Def:      "first",
			ErrOk:    true,
			Err:      "second"},
		nil)
	newP, err = p.Append([]string{}, &Rule{id: "fourth", effect: EffectDeny})
	if err != nil {
		t.Errorf("Expected no error but got %s", err)
	} else {
		if p, ok := newP.(*Policy); ok {
			if len(p.rules) == 4 {
				if p.rules[3].id != "fourth" {
					t.Errorf("Expected \"fourth\" rule placed at the end but got %q", p.rules[3].id)
				}
			} else {
				t.Errorf("Expected four rules after append but got %d", len(p.rules))
			}

			assertMapperRCAMapKeys(p.algorithm, "after insert \"fourth\"", t, "first", "fourth", "second", "third")
		} else {
			t.Errorf("Expected new policy but got %T (%#v)", newP, newP)
		}
	}

	newP, err = newP.Append([]string{}, &Rule{id: "first", effect: EffectDeny})
	if err != nil {
		t.Errorf("Expected no error but got %s", err)
	} else {
		if p, ok := newP.(*Policy); ok {
			if len(p.rules) == 4 {
				if p.rules[0].id != "first" {
					t.Errorf("Expected \"first\" rule replaced at the begining but got %q", p.rules[0].id)
				} else if p.rules[0].effect != EffectDeny {
					t.Errorf("Expected \"first\" rule became deny but it's still %s", effectNames[p.rules[0].effect])
				}
			} else {
				t.Errorf("Expected four rules after append but got %d", len(p.rules))
			}

			assertMapperRCAMapKeys(p.algorithm, "after insert \"first\"", t, "first", "fourth", "second", "third")
		} else {
			t.Errorf("Expected new policy but got %T (%#v)", newP, newP)
		}
	}
}

func TestPolicyDelete(t *testing.T) {
	p := &Policy{
		id: "test",
		rules: []*Rule{
			{id: "first", effect: EffectPermit},
			{id: "second", effect: EffectPermit},
			{id: "third", effect: EffectPermit}},
		algorithm: firstApplicableEffectRCA{}}

	newP, err := p.Delete([]string{"second"})
	if err != nil {
		t.Errorf("Expected no error but got %s", err)
	} else if p, ok := newP.(*Policy); ok {
		if len(p.rules) == 2 {
			if p.rules[0].id != "first" || p.rules[1].id != "third" {
				t.Errorf("Expected \"first\" and \"third\" rules remaining but got %q and %q",
					p.rules[0].id, p.rules[1].id)
			}
		} else {
			t.Errorf("Expected two rules after delete but got %d", len(p.rules))
		}
	} else {
		t.Errorf("Expected new policy but got %T (%#v)", newP, newP)
	}

	newP, err = p.Delete([]string{"first"})
	if err != nil {
		t.Errorf("Expected no error but got %s", err)
	} else if p, ok := newP.(*Policy); ok {
		if len(p.rules) == 2 {
			if p.rules[0].id != "second" || p.rules[1].id != "third" {
				t.Errorf("Expected \"second\" and \"third\" rules remaining but got %q and %q",
					p.rules[0].id, p.rules[1].id)
			}
		} else {
			t.Errorf("Expected two rules after delete but got %d", len(p.rules))
		}
	} else {
		t.Errorf("Expected new policy but got %T (%#v)", newP, newP)
	}

	newP, err = p.Delete([]string{"third"})
	if err != nil {
		t.Errorf("Expected no error but got %s", err)
	} else if p, ok := newP.(*Policy); ok {
		if len(p.rules) == 2 {
			if p.rules[0].id != "first" || p.rules[1].id != "second" {
				t.Errorf("Expected \"first\" and \"second\" rules remaining but got %q and %q",
					p.rules[0].id, p.rules[1].id)
			}
		} else {
			t.Errorf("Expected two rules after delete but got %d", len(p.rules))
		}
	} else {
		t.Errorf("Expected new policy but got %T (%#v)", newP, newP)
	}

	newP, err = p.Delete([]string{"fourth"})
	if err == nil {
		t.Errorf("Expected error but got new policy %T (%#v)", newP, newP)
	} else if _, ok := err.(*missingPolicyChildError); !ok {
		t.Errorf("Expected *missingPolicyChildError but got %T (%s)", err, err)
	}

	p = &Policy{
		hidden:    true,
		rules:     []*Rule{{id: "test", effect: EffectPermit}},
		algorithm: firstApplicableEffectRCA{}}
	newP, err = p.Delete([]string{"test"})
	if err == nil {
		t.Errorf("Expected error but got new policy %T (%#v)", newP, newP)
	} else if _, ok := err.(*hiddenPolicyModificationError); !ok {
		t.Errorf("Expected *hiddenPolicyModificationError but got %T (%s)", err, err)
	}

	p = &Policy{
		id:        "test",
		rules:     []*Rule{{id: "test", effect: EffectPermit}},
		algorithm: firstApplicableEffectRCA{}}
	newP, err = p.Delete([]string{})
	if err == nil {
		t.Errorf("Expected error but got new policy %T (%#v)", newP, newP)
	} else if _, ok := err.(*tooShortPathPolicyModificationError); !ok {
		t.Errorf("Expected *tooShortPathPolicyModificationError but got %T (%s)", err, err)
	}

	newP, err = p.Delete([]string{"test", "example"})
	if err == nil {
		t.Errorf("Expected error but got new policy %T (%#v)", newP, newP)
	} else if _, ok := err.(*tooLongPathPolicyModificationError); !ok {
		t.Errorf("Expected *tooLongPathPolicyModificationError but got %T (%s)", err, err)
	}

	p = NewPolicy("test", false, Target{},
		[]*Rule{
			{id: "first", effect: EffectPermit},
			{id: "second", effect: EffectPermit},
			{id: "third", effect: EffectPermit}},
		makeMapperRCA, MapperRCAParams{
			Argument: AttributeDesignator{a: Attribute{id: "k", t: TypeString}},
			DefOk:    true,
			Def:      "first",
			ErrOk:    true,
			Err:      "second"},
		nil)
	newP, err = p.Delete([]string{"second"})
	if err != nil {
		t.Errorf("Expected no error but got %s", err)
	} else if p, ok := newP.(*Policy); ok {
		if len(p.rules) == 2 {
			if p.rules[0].id != "first" || p.rules[1].id != "third" {
				t.Errorf("Expected \"first\" and \"third\" rules remaining but got %q and %q",
					p.rules[0].id, p.rules[1].id)
			}
		} else {
			t.Errorf("Expected two rules after delete but got %d", len(p.rules))
		}

		assertMapperRCAMapKeys(p.algorithm, "after deletion", t, "first", "third")
	} else {
		t.Errorf("Expected new policy but got %T (%#v)", newP, newP)
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

func assertMapperRCAMapKeys(a ruleCombiningAlg, desc string, t *testing.T, expected ...string) {
	if m, ok := a.(mapperRCA); ok {
		keys := []string{}
		for p := range m.rules.Enumerate() {
			keys = append(keys, p.Key)
		}

		assertStrings(keys, expected, desc, t)
	} else {
		t.Errorf("Expected mapper as rule combining algorithm but got %T for %s", a, desc)
	}
}

func assertStrings(v, e []string, desc string, t *testing.T) {
	veol := make([]string, len(v))
	for i, s := range v {
		veol[i] = s + "\n"
	}

	eeol := make([]string, len(e))
	for i, s := range e {
		eeol[i] = s + "\n"
	}

	ctx := difflib.ContextDiff{
		A:        eeol,
		B:        veol,
		FromFile: "Expected",
		ToFile:   "Got"}

	diff, err := difflib.GetContextDiffString(ctx)
	if err != nil {
		panic(fmt.Errorf("Can't compare \"%s\": %s", desc, err))
	}

	if len(diff) > 0 {
		t.Errorf("\"%s\" doesn't match:\n%s", desc, diff)
	}
}
