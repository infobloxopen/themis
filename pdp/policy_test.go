package pdp

import (
	"fmt"
	"testing"

	"github.com/pmezard/go-difflib/difflib"
)

func TestPolicy(t *testing.T) {
	c := &Context{
		a: map[string]interface{}{
			"missing-type":   MakeBooleanValue(false),
			"test-string":    MakeStringValue("test"),
			"example-string": MakeStringValue("example")}}

	testID := "Test Policy"

	p := makeSimplePolicy(testID)
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

	p = &Policy{
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

	p = &Policy{
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

	p = &Policy{
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

	p = &Policy{
		id:          testID,
		target:      makeSimpleStringTarget("test-string", "test"),
		rules:       []*Rule{makeSimpleHiddenRule(EffectPermit)},
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

	defaultRule := makeSimpleRule("Default", EffectDeny)
	errorRule := makeSimpleRule("Error", EffectDeny)
	permitRule := makeSimpleRule("Permit", EffectPermit)
	p = &Policy{
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
		a: map[string]interface{}{
			"x": MakeSetOfStringsValue(newStrTree("Permit", "Default")),
			"y": MakeStringValue("Permit")}}

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
		a: map[string]interface{}{
			"x": MakeSetOfStringsValue(newStrTree("Permit", "Default")),
			"y": MakeSetOfStringsValue(newStrTree("Permit", "Default"))}}

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
	p := makeSimplePolicy("test")
	p.ord = 5

	newE, err := p.Append([]string{}, makeSimpleRule("test", EffectPermit))
	if err != nil {
		t.Errorf("Expected no error but got %s", err)
	} else if newP, ok := newE.(*Policy); ok {
		if p == newP {
			t.Errorf("Expected different new policy but got the same")
		}

		if newP.ord != p.ord {
			t.Errorf("Expected unchanged order %d but got %d", p.ord, newP.ord)
		}

		if len(newP.rules) == 1 {
			r := newP.rules[0]
			if r.ord != 0 {
				t.Errorf("Expected index of the only rule to be 0 but got %d", r.ord)
			}
		} else {
			t.Errorf("Expected only appended rule but got %d rules", len(newP.rules))
		}
	} else {
		t.Errorf("Expected new policy but got %T (%#v)", newE, newE)
	}

	p = makeSimpleHiddenPolicy()
	newE, err = p.Append([]string{}, makeSimpleRule("test", EffectPermit))
	if err == nil {
		t.Errorf("Expected error but got policy %#v", newE)
	} else if _, ok := err.(*hiddenPolicyModificationError); !ok {
		t.Errorf("Expected *hiddenPolicyModificationError but got %T (%s)", err, err)
	}

	p = makeSimplePolicy("test")
	newE, err = p.Append([]string{"test"}, makeSimpleRule("test", EffectPermit))
	if err == nil {
		t.Errorf("Expected error but got policy %#v", newE)
	} else if _, ok := err.(*tooLongPathPolicyModificationError); !ok {
		t.Errorf("Expected *tooLongPathPolicyModificationError but got %T (%s)", err, err)
	}

	p = makeSimplePolicy("test")
	newE, err = p.Append([]string{}, makeSimplePolicy("test"))
	if err == nil {
		t.Errorf("Expected error but got policy %#v", newE)
	} else if _, ok := err.(*invalidPolicyItemTypeError); !ok {
		t.Errorf("Expected *invalidPolicyItemTypeError but got %T (%s)", err, err)
	}

	p = makeSimplePolicy("test",
		makeSimpleRule("first", EffectPermit),
		makeSimpleRule("second", EffectPermit),
	)
	if len(p.rules) == 2 {
		for i, r := range p.rules {
			if r.ord != i {
				t.Errorf("Expected %q rule to get %d order but got %d", r.id, i, r.ord)
			}
		}
	} else {
		t.Errorf("Expected 2 rules in the policy but got %d", len(p.rules))
	}

	newE, err = p.Append([]string{}, makeSimpleRule("third", EffectPermit))
	if err != nil {
		t.Errorf("Expected no error but got %s", err)
	} else if newP, ok := newE.(*Policy); ok {
		if len(newP.rules) == 3 {
			r := newP.rules[2]
			if r.id != "third" {
				t.Errorf("Expected \"third\" rule added to the end but got %q", r.id)
			}

			if r.ord != 2 {
				t.Errorf("Expected the last rule to get order 2 but got %d", r.ord)
			}
		} else {
			t.Errorf("Expected three rules after append but got %d", len(newP.rules))
		}
	} else {
		t.Errorf("Expected new policy but got %T (%#v)", newE, newE)
	}

	newE, err = newE.Append([]string{}, makeSimpleRule("first", EffectDeny))
	if err != nil {
		t.Errorf("Expected no error but got %s", err)
	} else if newP, ok := newE.(*Policy); ok {
		if len(newP.rules) == 3 {
			r := newP.rules[0]
			if r.id != "first" {
				t.Errorf("Expected \"first\" rule replaced at the begining but got %q", r.id)
			} else if r.effect != EffectDeny {
				t.Errorf("Expected \"first\" rule became deny but it's still %s", effectNames[r.effect])
			}

			if r.ord != 0 {
				t.Errorf("Expected the first rule to keep order 0 but got %d", r.ord)
			}
		} else {
			t.Errorf("Expected three rules after append but got %d", len(newP.rules))
		}
	} else {
		t.Errorf("Expected new policy but got %T (%#v)", newE, newE)
	}

	newE, err = newE.Append([]string{}, makeSimpleRule("second", EffectDeny))
	if err != nil {
		t.Errorf("Expected no error but got %s", err)
	} else if newP, ok := newE.(*Policy); ok {
		if len(newP.rules) == 3 {
			r := newP.rules[1]
			if r.id != "second" {
				t.Errorf("Expected \"second\" rule replaced at the middle but got %q", r.id)
			} else if r.effect != EffectDeny {
				t.Errorf("Expected \"second\" rule became deny but it's still %s", effectNames[r.effect])
			}

			if r.ord != 1 {
				t.Errorf("Expected second rule to keep order 1 but got %d", r.ord)
			}
		} else {
			t.Errorf("Expected three rules after append but got %d", len(newP.rules))
		}
	} else {
		t.Errorf("Expected new policy but got %T (%#v)", newE, newE)
	}

	newE, err = newE.Append([]string{}, makeSimpleRule("third", EffectDeny))
	if err != nil {
		t.Errorf("Expected no error but got %s", err)
	} else if newP, ok := newE.(*Policy); ok {
		if len(newP.rules) == 3 {
			r := newP.rules[2]
			if r.id != "third" {
				t.Errorf("Expected \"third\" rule replaced at the end but got %q", r.id)
			} else if r.effect != EffectDeny {
				t.Errorf("Expected \"third\" rule became deny but it's still %s", effectNames[r.effect])
			}

			if r.ord != 2 {
				t.Errorf("Expected third rule to keep order 2 but got %d", r.ord)
			}
		} else {
			t.Errorf("Expected three rules after append but got %d", len(newP.rules))
		}
	} else {
		t.Errorf("Expected new policy but got %T (%#v)", newE, newE)
	}

	p = NewPolicy("test", false, Target{},
		[]*Rule{
			makeSimpleRule("first", EffectPermit),
			makeSimpleRule("second", EffectPermit),
			makeSimpleRule("third", EffectPermit),
		},
		makeMapperRCA, MapperRCAParams{
			Argument: AttributeDesignator{a: Attribute{id: "k", t: TypeString}},
			DefOk:    true,
			Def:      "first",
			ErrOk:    true,
			Err:      "second"},
		nil)
	if len(p.rules) == 3 {
		for i, r := range p.rules {
			if r.ord != i {
				t.Errorf("Expected %q rule to get %d order but got %d", r.id, i, r.ord)
			}
		}
	} else {
		t.Errorf("Expected 3 rules in the policy but got %d", len(p.rules))
	}

	newE, err = p.Append([]string{}, makeSimpleRule("fourth", EffectDeny))
	if err != nil {
		t.Errorf("Expected no error but got %s", err)
	} else if newP, ok := newE.(*Policy); ok {
		if len(newP.rules) == 4 {
			r := newP.rules[3]
			if r.id != "fourth" {
				t.Errorf("Expected \"fourth\" rule placed at the end but got %q", r.id)
			}

			if r.ord != 3 {
				t.Errorf("Expected fourth rule to get order 3 but got %d", r.ord)
			}
		} else {
			t.Errorf("Expected four rules after append but got %d", len(newP.rules))
		}

		assertMapperRCAMapKeys(newP.algorithm, "after insert \"fourth\"", t, "first", "fourth", "second", "third")
	} else {
		t.Errorf("Expected new policy but got %T (%#v)", newE, newE)
	}

	newE, err = newE.Append([]string{}, &Rule{id: "first", effect: EffectDeny})
	if err != nil {
		t.Errorf("Expected no error but got %s", err)
	} else if newP, ok := newE.(*Policy); ok {
		if len(newP.rules) == 4 {
			r := newP.rules[0]
			if r.id != "first" {
				t.Errorf("Expected \"first\" rule replaced at the begining but got %q", r.id)
			} else if r.effect != EffectDeny {
				t.Errorf("Expected \"first\" rule became deny but it's still %s", effectNames[r.effect])
			}

			if r.ord != 0 {
				t.Errorf("Expected first rule to keep order 0 but got %d", r.ord)
			}
		} else {
			t.Errorf("Expected four rules after append but got %d", len(newP.rules))
		}

		assertMapperRCAMapKeys(newP.algorithm, "after insert \"first\"", t, "first", "fourth", "second", "third")
	} else {
		t.Errorf("Expected new policy but got %T (%#v)", newE, newE)
	}
}

func TestPolicyDelete(t *testing.T) {
	p := makeSimplePolicy("test",
		makeSimpleRule("first", EffectPermit),
		makeSimpleRule("second", EffectPermit),
		makeSimpleRule("third", EffectPermit),
	)
	if len(p.rules) == 3 {
		for i, r := range p.rules {
			if r.ord != i {
				t.Errorf("Expected %q rule to get %d order but got %d", r.id, i, r.ord)
			}
		}
	} else {
		t.Errorf("Expected 3 rules in the policy but got %d", len(p.rules))
	}

	newE, err := p.Delete([]string{"second"})
	if err != nil {
		t.Errorf("Expected no error but got %s", err)
	} else if newP, ok := newE.(*Policy); ok {
		if len(newP.rules) == 2 {
			r1 := newP.rules[0]
			r3 := newP.rules[1]
			if r1.id != "first" || r3.id != "third" {
				t.Errorf("Expected \"first\" and \"third\" rules remaining but got %q and %q", r1.id, r3.id)
			}

			if r1.ord != 0 || r3.ord != 2 {
				t.Errorf("Expected remaining rules to keep their orders but got %d and %d", r1.ord, r3.ord)
			}
		} else {
			t.Errorf("Expected two rules after delete but got %d", len(newP.rules))
		}
	} else {
		t.Errorf("Expected new policy but got %T (%#v)", newE, newE)
	}

	newE, err = p.Delete([]string{"first"})
	if err != nil {
		t.Errorf("Expected no error but got %s", err)
	} else if newP, ok := newE.(*Policy); ok {
		if len(newP.rules) == 2 {
			r2 := newP.rules[0]
			r3 := newP.rules[1]
			if r2.id != "second" || r3.id != "third" {
				t.Errorf("Expected \"second\" and \"third\" rules remaining but got %q and %q", r2.id, r3.id)
			}

			if r2.ord != 1 || r3.ord != 2 {
				t.Errorf("Expected remaining rules to keep their orders but got %d and %d", r2.ord, r3.ord)
			}
		} else {
			t.Errorf("Expected two rules after delete but got %d", len(newP.rules))
		}
	} else {
		t.Errorf("Expected new policy but got %T (%#v)", newE, newE)
	}

	newE, err = p.Delete([]string{"third"})
	if err != nil {
		t.Errorf("Expected no error but got %s", err)
	} else if newP, ok := newE.(*Policy); ok {
		if len(newP.rules) == 2 {
			r1 := newP.rules[0]
			r2 := newP.rules[1]
			if r1.id != "first" || r2.id != "second" {
				t.Errorf("Expected \"first\" and \"second\" rules remaining but got %q and %q", r1.id, r2.id)
			}

			if r1.ord != 0 || r2.ord != 1 {
				t.Errorf("Expected remaining rules to keep their orders but got %d and %d", r1.ord, r2.ord)
			}
		} else {
			t.Errorf("Expected two rules after delete but got %d", len(newP.rules))
		}
	} else {
		t.Errorf("Expected new policy but got %T (%#v)", newE, newE)
	}

	newE, err = p.Delete([]string{"fourth"})
	if err == nil {
		t.Errorf("Expected error but got new policy %T (%#v)", newE, newE)
	} else if _, ok := err.(*missingPolicyChildError); !ok {
		t.Errorf("Expected *missingPolicyChildError but got %T (%s)", err, err)
	}

	p = makeSimpleHiddenPolicy(makeSimpleRule("test", EffectPermit))
	newE, err = p.Delete([]string{"test"})
	if err == nil {
		t.Errorf("Expected error but got new policy %T (%#v)", newE, newE)
	} else if _, ok := err.(*hiddenPolicyModificationError); !ok {
		t.Errorf("Expected *hiddenPolicyModificationError but got %T (%s)", err, err)
	}

	p = makeSimplePolicy("test", makeSimpleRule("test", EffectPermit))
	newE, err = p.Delete([]string{})
	if err == nil {
		t.Errorf("Expected error but got new policy %T (%#v)", newE, newE)
	} else if _, ok := err.(*tooShortPathPolicyModificationError); !ok {
		t.Errorf("Expected *tooShortPathPolicyModificationError but got %T (%s)", err, err)
	}

	newE, err = p.Delete([]string{"test", "example"})
	if err == nil {
		t.Errorf("Expected error but got new policy %T (%#v)", newE, newE)
	} else if _, ok := err.(*tooLongPathPolicyModificationError); !ok {
		t.Errorf("Expected *tooLongPathPolicyModificationError but got %T (%s)", err, err)
	}

	p = NewPolicy("test", false, Target{},
		[]*Rule{
			makeSimpleRule("first", EffectPermit),
			makeSimpleRule("second", EffectPermit),
			makeSimpleRule("third", EffectPermit),
		},
		makeMapperRCA, MapperRCAParams{
			Argument: AttributeDesignator{a: Attribute{id: "k", t: TypeString}},
			DefOk:    true,
			Def:      "first",
			ErrOk:    true,
			Err:      "second"},
		nil)
	if len(p.rules) == 3 {
		for i, r := range p.rules {
			if r.ord != i {
				t.Errorf("Expected %q rule to get %d order but got %d", r.id, i, r.ord)
			}
		}
	} else {
		t.Errorf("Expected 3 rules in the policy but got %d", len(p.rules))
	}

	newE, err = p.Delete([]string{"second"})
	if err != nil {
		t.Errorf("Expected no error but got %s", err)
	} else if newP, ok := newE.(*Policy); ok {
		if len(newP.rules) == 2 {
			r1 := newP.rules[0]
			r3 := newP.rules[1]
			if r1.id != "first" || r3.id != "third" {
				t.Errorf("Expected \"first\" and \"third\" rules remaining but got %q and %q", r1.id, r3.id)
			}

			if r1.ord != 0 || r3.ord != 2 {
				t.Errorf("Expected remaining rules to keep their orders but got %d and %d", r1.ord, r3.ord)
			}
		} else {
			t.Errorf("Expected two rules after delete but got %d", len(newP.rules))
		}

		assertMapperRCAMapKeys(newP.algorithm, "after deletion", t, "first", "third")
	} else {
		t.Errorf("Expected new policy but got %T (%#v)", newE, newE)
	}
}

func makeSimplePolicy(ID string, rules ...*Rule) *Policy {
	return NewPolicy(
		ID, false,
		Target{},
		rules,
		makeFirstApplicableEffectRCA,
		nil,
		nil,
	)
}

func makeSimpleHiddenPolicy(rules ...*Rule) *Policy {
	return NewPolicy(
		"", true,
		Target{},
		rules,
		makeFirstApplicableEffectRCA,
		nil,
		nil,
	)
}

func makeSimpleRule(ID string, effect int) *Rule {
	return NewRule(
		ID, false,
		Target{},
		nil,
		effect,
		nil,
	)
}

func makeSimpleHiddenRule(effect int) *Rule {
	return NewRule(
		"", true,
		Target{},
		nil,
		effect,
		nil,
	)
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

func assertMapperRCAMapKeys(a RuleCombiningAlg, desc string, t *testing.T, expected ...string) {
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
		panic(fmt.Errorf("can't compare \"%s\": %s", desc, err))
	}

	if len(diff) > 0 {
		t.Errorf("\"%s\" doesn't match:\n%s", desc, diff)
	}
}
