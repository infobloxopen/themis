package pdp

import "testing"

func TestPolicySet(t *testing.T) {
	c := &Context{
		a: map[string]interface{}{
			"missing-type":   MakeBooleanValue(false),
			"test-string":    MakeStringValue("test"),
			"example-string": MakeStringValue("example")}}

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
				Argument: AttributeDesignator{a: Attribute{id: "x", t: TypeSetOfStrings}},
				DefOk:    true,
				Def:      "Default",
				ErrOk:    true,
				Err:      "Error",
				Algorithm: makeMapperPCA(
					nil,
					MapperPCAParams{
						Argument: AttributeDesignator{a: Attribute{id: "y", t: TypeString}}})})}

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

func TestPolicySetAppend(t *testing.T) {
	testPermitPol := &PolicySet{
		id: "test",
		policies: []Evaluable{
			&Policy{
				id:        "permit",
				rules:     []*Rule{{id: "permit", effect: EffectPermit}},
				algorithm: firstApplicableEffectRCA{}}},
		algorithm: firstApplicableEffectPCA{}}

	p := &PolicySet{id: "test", algorithm: firstApplicableEffectPCA{}}
	newP, err := p.Append([]string{}, testPermitPol)
	if err != nil {
		t.Errorf("Expected no error but got %s", err)
	} else {
		if p == newP {
			t.Errorf("Expected different new policy set but got the same")
		}
	}

	p = &PolicySet{hidden: true, algorithm: firstApplicableEffectPCA{}}
	newP, err = p.Append([]string{}, testPermitPol)
	if err == nil {
		t.Errorf("Expected error but got policy %#v", newP)
	} else if _, ok := err.(*hiddenPolicySetModificationError); !ok {
		t.Errorf("Expected *hiddenPolicySetModificationError but got %T (%s)", err, err)
	}

	p = &PolicySet{id: "test", algorithm: firstApplicableEffectPCA{}}
	newP, err = p.Append([]string{"test"}, testPermitPol)
	if err == nil {
		t.Errorf("Expected error but got policy %#v", newP)
	} else if _, ok := err.(*missingPolicySetChildError); !ok {
		t.Errorf("Expected *missingPolicySetChildError but got %T (%s)", err, err)
	}

	p = &PolicySet{
		id:        "test",
		policies:  []Evaluable{&Policy{id: "test", algorithm: firstApplicableEffectRCA{}}},
		algorithm: firstApplicableEffectPCA{}}
	newP, err = p.Append([]string{"test"}, &Rule{hidden: true, effect: EffectPermit})
	if err == nil {
		t.Errorf("Expected error but got policy %#v", newP)
	} else if _, ok := err.(*hiddenRuleAppendError); !ok {
		t.Errorf("Expected *hiddenRuleAppendError but got %T (%s)", err, err)
	}

	p = &PolicySet{id: "test", algorithm: firstApplicableEffectPCA{}}
	newP, err = p.Append([]string{}, &Rule{id: "permit", effect: EffectPermit})
	if err == nil {
		t.Errorf("Expected error but got policy %#v", newP)
	} else if _, ok := err.(*invalidPolicySetItemTypeError); !ok {
		t.Errorf("Expected *invalidPolicySetItemTypeError but got %T (%s)", err, err)
	}

	p = &PolicySet{id: "test", algorithm: firstApplicableEffectPCA{}}
	newP, err = p.Append([]string{}, &PolicySet{hidden: true, algorithm: firstApplicableEffectPCA{}})
	if err == nil {
		t.Errorf("Expected error but got policy %#v", newP)
	} else if _, ok := err.(*hiddenPolicyAppendError); !ok {
		t.Errorf("Expected *hiddenPolicyAppendError but got %T (%s)", err, err)
	}

	p = &PolicySet{
		id:        "test",
		policies:  []Evaluable{&Policy{id: "test", algorithm: firstApplicableEffectRCA{}}},
		algorithm: firstApplicableEffectPCA{}}
	newP, err = p.Append([]string{"test"}, &Rule{id: "test", effect: EffectPermit})
	if err != nil {
		t.Errorf("Expected no error but got %s", err)
	}

	testFirstPol := &Policy{
		id:        "first",
		rules:     []*Rule{{id: "deny", effect: EffectDeny}},
		algorithm: firstApplicableEffectRCA{}}
	testSecondPol := &Policy{
		id:        "second",
		rules:     []*Rule{{id: "deny", effect: EffectDeny}},
		algorithm: firstApplicableEffectRCA{}}
	testThirdPermitPol := &Policy{
		id:        "third",
		rules:     []*Rule{{id: "permit", effect: EffectPermit}},
		algorithm: firstApplicableEffectRCA{}}
	testThirdDenyPol := &Policy{
		id:        "third",
		rules:     []*Rule{{id: "deny", effect: EffectDeny}},
		algorithm: firstApplicableEffectRCA{}}

	p = &PolicySet{
		id: "test",
		policies: []Evaluable{
			&Policy{
				id:        "first",
				rules:     []*Rule{{id: "permit", effect: EffectPermit}},
				algorithm: firstApplicableEffectRCA{}},
			&Policy{
				id:        "second",
				rules:     []*Rule{{id: "permit", effect: EffectPermit}},
				algorithm: firstApplicableEffectRCA{}}},
		algorithm: firstApplicableEffectPCA{}}
	newP, err = p.Append([]string{}, testThirdPermitPol)
	if err != nil {
		t.Errorf("Expected no error but got %s", err)
	} else {
		if p, ok := newP.(*PolicySet); ok {
			if len(p.policies) == 3 {
				p := p.policies[2]
				if p, ok := p.(*Policy); ok {
					if p.id != "third" {
						t.Errorf("Expected \"third\" policy added to the end but got %q", p.id)
					}
				} else {
					t.Errorf("Expected policy as third item of policy set but got %T (%#v)", p, p)
				}
			} else {
				t.Errorf("Expected three policies after append but got %d", len(p.policies))
			}
		} else {
			t.Errorf("Expected new policy set but got %T (%#v)", newP, newP)
		}
	}

	newP, err = newP.Append([]string{}, testFirstPol)
	if err != nil {
		t.Errorf("Expected no error but got %s", err)
	} else {
		if p, ok := newP.(*PolicySet); ok {
			if len(p.policies) == 3 {
				p := p.policies[0]
				if p, ok := p.(*Policy); ok {
					if p.id != "first" {
						t.Errorf("Expected \"first\" policy replaced at the begining but got %q", p.id)
					} else if p.rules[0].effect != EffectDeny {
						t.Errorf("Expected \"first\" policy became deny but it's still %s",
							effectNames[p.rules[0].effect])
					}
				} else {
					t.Errorf("Expected policy as first item of policy set but got %T (%#v)", p, p)
				}
			} else {
				t.Errorf("Expected three policies after append but got %d", len(p.policies))
			}
		} else {
			t.Errorf("Expected new policy set but got %T (%#v)", newP, newP)
		}
	}

	newP, err = newP.Append([]string{}, testSecondPol)
	if err != nil {
		t.Errorf("Expected no error but got %s", err)
	} else {
		if p, ok := newP.(*PolicySet); ok {
			if len(p.policies) == 3 {
				p := p.policies[1]
				if p, ok := p.(*Policy); ok {
					if p.id != "second" {
						t.Errorf("Expected \"second\" policy replaced at the middle but got %q", p.id)
					} else if p.rules[0].effect != EffectDeny {
						t.Errorf("Expected \"second\" policy became deny but it's still %s",
							effectNames[p.rules[0].effect])
					}
				} else {
					t.Errorf("Expected policy as second item of policy set but got %T (%#v)", p, p)
				}
			} else {
				t.Errorf("Expected three policies after append but got %d", len(p.policies))
			}
		} else {
			t.Errorf("Expected new policy set but got %T (%#v)", newP, newP)
		}
	}

	newP, err = newP.Append([]string{}, testThirdDenyPol)
	if err != nil {
		t.Errorf("Expected no error but got %s", err)
	} else {
		if p, ok := newP.(*PolicySet); ok {
			if len(p.policies) == 3 {
				p := p.policies[2]
				if p, ok := p.(*Policy); ok {
					if p.id != "third" {
						t.Errorf("Expected \"third\" policy replaced at the end but got %q", p.id)
					} else if p.rules[0].effect != EffectDeny {
						t.Errorf("Expected \"third\" policy became deny but it's still %s",
							effectNames[p.rules[0].effect])
					}
				} else {
					t.Errorf("Expected policy as third item of policy set but got %T (%#v)", p, p)
				}
			} else {
				t.Errorf("Expected three policies after append but got %d", len(p.policies))
			}
		} else {
			t.Errorf("Expected new policy but got %T (%#v)", newP, newP)
		}
	}

	testFourthPol := &Policy{
		id:        "fourth",
		rules:     []*Rule{{id: "permit", effect: EffectPermit}},
		algorithm: firstApplicableEffectRCA{}}

	p = NewPolicySet("test", false, Target{},
		[]Evaluable{
			&Policy{
				id:        "first",
				rules:     []*Rule{{id: "permit", effect: EffectPermit}},
				algorithm: firstApplicableEffectRCA{}},
			&Policy{
				id:        "second",
				rules:     []*Rule{{id: "permit", effect: EffectPermit}},
				algorithm: firstApplicableEffectRCA{}},
			&Policy{
				id:        "third",
				rules:     []*Rule{{id: "permit", effect: EffectPermit}},
				algorithm: firstApplicableEffectRCA{}}},
		makeMapperPCA, MapperPCAParams{
			Argument: AttributeDesignator{a: Attribute{id: "k", t: TypeString}},
			DefOk:    true,
			Def:      "first",
			ErrOk:    true,
			Err:      "second"},
		nil)
	newP, err = p.Append([]string{}, testFourthPol)
	if err != nil {
		t.Errorf("Expected no error but got %s", err)
	} else {
		if p, ok := newP.(*PolicySet); ok {
			if len(p.policies) == 4 {
				p := p.policies[3]
				if p, ok := p.(*Policy); ok {
					if p.id != "fourth" {
						t.Errorf("Expected \"fourth\" policy added to the end but got %q", p.id)
					}
				} else {
					t.Errorf("Expected policy as fourth item of policy set but got %T (%#v)", p, p)
				}
			} else {
				t.Errorf("Expected four policies after append but got %d", len(p.policies))
			}

			assertMapperPCAMapKeys(p.algorithm, "after insert \"fourth\"", t, "first", "fourth", "second", "third")
		} else {
			t.Errorf("Expected new policy set but got %T (%#v)", newP, newP)
		}
	}

	newP, err = newP.Append([]string{}, testFirstPol)
	if err != nil {
		t.Errorf("Expected no error but got %s", err)
	} else {
		if p, ok := newP.(*PolicySet); ok {
			if len(p.policies) == 4 {
				p := p.policies[0]
				if p, ok := p.(*Policy); ok {
					if p.id != "first" {
						t.Errorf("Expected \"first\" policy replaced at the begining but got %q", p.id)
					} else if p.rules[0].effect != EffectDeny {
						t.Errorf("Expected \"first\" policy became deny but it's still %s",
							effectNames[p.rules[0].effect])
					}
				} else {
					t.Errorf("Expected policy as first item of policy set but got %T (%#v)", p, p)
				}
			} else {
				t.Errorf("Expected four policies after append but got %d", len(p.policies))
			}

			assertMapperPCAMapKeys(p.algorithm, "after insert \"fourth\"", t, "first", "fourth", "second", "third")
		} else {
			t.Errorf("Expected new policy set but got %T (%#v)", newP, newP)
		}
	}
}

func TestPolicySetDelete(t *testing.T) {
	p := &PolicySet{
		id: "test",
		policies: []Evaluable{
			&Policy{
				id:        "first",
				rules:     []*Rule{{id: "permit", effect: EffectPermit}},
				algorithm: firstApplicableEffectRCA{}},
			&Policy{
				id:        "second",
				rules:     []*Rule{{id: "permit", effect: EffectPermit}},
				algorithm: firstApplicableEffectRCA{}},
			&Policy{
				id:        "third",
				rules:     []*Rule{{id: "permit", effect: EffectPermit}},
				algorithm: firstApplicableEffectRCA{}}},
		algorithm: firstApplicableEffectPCA{}}

	newP, err := p.Delete([]string{"second"})
	if err != nil {
		t.Errorf("Expected no error but got %s", err)
	} else if p, ok := newP.(*PolicySet); ok {
		if len(p.policies) == 2 {
			p1, ok1 := p.policies[0].(*Policy)
			p2, ok2 := p.policies[1].(*Policy)
			if ok1 && ok2 {
				if p1.id != "first" || p2.id != "third" {
					t.Errorf("Expected \"first\" and \"third\" policies remaining but got %q and %q", p1.id, p2.id)
				}
			} else {
				t.Errorf("Expected two policies after delete but got %T and %T", p.policies[0], p.policies[1])
			}
		} else {
			t.Errorf("Expected two policies after delete but got %d", len(p.policies))
		}
	} else {
		t.Errorf("Expected new policy set but got %T (%#v)", newP, newP)
	}

	newP, err = p.Delete([]string{"first"})
	if err != nil {
		t.Errorf("Expected no error but got %s", err)
	} else if p, ok := newP.(*PolicySet); ok {
		if len(p.policies) == 2 {
			p1, ok1 := p.policies[0].(*Policy)
			p2, ok2 := p.policies[1].(*Policy)
			if ok1 && ok2 {
				if p1.id != "second" || p2.id != "third" {
					t.Errorf("Expected \"second\" and \"third\" policies remaining but got %q and %q", p1.id, p2.id)
				}
			} else {
				t.Errorf("Expected two policies after delete but got %T and %T", p.policies[0], p.policies[1])
			}
		} else {
			t.Errorf("Expected two policies after delete but got %d", len(p.policies))
		}
	} else {
		t.Errorf("Expected new policy set but got %T (%#v)", newP, newP)
	}

	newP, err = p.Delete([]string{"third"})
	if err != nil {
		t.Errorf("Expected no error but got %s", err)
	} else if p, ok := newP.(*PolicySet); ok {
		if len(p.policies) == 2 {
			p1, ok1 := p.policies[0].(*Policy)
			p2, ok2 := p.policies[1].(*Policy)
			if ok1 && ok2 {
				if p1.id != "first" || p2.id != "second" {
					t.Errorf("Expected \"first\" and \"second\" policies remaining but got %q and %q", p1.id, p2.id)
				}
			} else {
				t.Errorf("Expected two policies after delete but got %T and %T", p.policies[0], p.policies[1])
			}
		} else {
			t.Errorf("Expected two policies after delete but got %d", len(p.policies))
		}
	} else {
		t.Errorf("Expected new policy set but got %T (%#v)", newP, newP)
	}

	newP, err = p.Delete([]string{"first", "permit"})
	if err != nil {
		t.Errorf("Expected no error but got %s", err)
	} else if p, ok := newP.(*PolicySet); ok {
		if len(p.policies) == 3 {
			p := p.policies[0]
			if p, ok := p.(*Policy); ok {
				if p.id == "first" {
					if len(p.rules) > 0 {
						t.Errorf("Expected no rules after nested delete but got %d", len(p.rules))
					}
				} else {
					t.Errorf("Expected \"first\" policy at the beginning but got %q", p.id)
				}
			} else {
				t.Errorf("Expected policy as first item of policy set but got %T (%#v)", p, p)
			}
		} else {
			t.Errorf("Expected three policies after nested delete but got %d", len(p.policies))
		}
	} else {
		t.Errorf("Expected new policy set but got %T (%#v)", newP, newP)
	}

	newP, err = p.Delete([]string{"fourth"})
	if err == nil {
		t.Errorf("Expected error but got new policy %T (%#v)", newP, newP)
	} else if _, ok := err.(*missingPolicySetChildError); !ok {
		t.Errorf("Expected *missingPolicySetChildError but got %T (%s)", err, err)
	}

	newP, err = p.Delete([]string{"fourth", "permit"})
	if err == nil {
		t.Errorf("Expected error but got new policy %T (%#v)", newP, newP)
	} else if _, ok := err.(*missingPolicySetChildError); !ok {
		t.Errorf("Expected *missingPolicySetChildError but got %T (%s)", err, err)
	}

	newP, err = p.Delete([]string{"first", "deny"})
	if err == nil {
		t.Errorf("Expected error but got new policy %T (%#v)", newP, newP)
	} else if _, ok := err.(*missingPolicyChildError); !ok {
		t.Errorf("Expected *missingPolicyChildError but got %T (%s)", err, err)
	}

	p = &PolicySet{
		hidden: true,
		policies: []Evaluable{
			&Policy{
				id:        "first",
				rules:     []*Rule{{id: "permit", effect: EffectPermit}},
				algorithm: firstApplicableEffectRCA{}}},
		algorithm: firstApplicableEffectPCA{}}
	newP, err = p.Delete([]string{"first"})
	if err == nil {
		t.Errorf("Expected error but got new policy %T (%#v)", newP, newP)
	} else if _, ok := err.(*hiddenPolicySetModificationError); !ok {
		t.Errorf("Expected *hiddenPolicySetModificationError but got %T (%s)", err, err)
	}

	p = &PolicySet{
		id: "test",
		policies: []Evaluable{
			&Policy{
				id:        "first",
				rules:     []*Rule{{id: "permit", effect: EffectPermit}},
				algorithm: firstApplicableEffectRCA{}}},
		algorithm: firstApplicableEffectPCA{}}
	newP, err = p.Delete([]string{})
	if err == nil {
		t.Errorf("Expected error but got new policy %T (%#v)", newP, newP)
	} else if _, ok := err.(*tooShortPathPolicySetModificationError); !ok {
		t.Errorf("Expected *tooShortPathPolicySetModificationError but got %T (%s)", err, err)
	}

	p = NewPolicySet("test", false, Target{},
		[]Evaluable{
			&Policy{
				id:        "first",
				rules:     []*Rule{{id: "permit", effect: EffectPermit}},
				algorithm: firstApplicableEffectRCA{}},
			&Policy{
				id:        "second",
				rules:     []*Rule{{id: "permit", effect: EffectPermit}},
				algorithm: firstApplicableEffectRCA{}},
			&Policy{
				id:        "third",
				rules:     []*Rule{{id: "permit", effect: EffectPermit}},
				algorithm: firstApplicableEffectRCA{}}},
		makeMapperPCA, MapperPCAParams{
			Argument: AttributeDesignator{a: Attribute{id: "k", t: TypeString}},
			DefOk:    true,
			Def:      "first",
			ErrOk:    true,
			Err:      "second"},
		nil)
	newP, err = p.Delete([]string{"second"})
	if err != nil {
		t.Errorf("Expected no error but got %s", err)
	} else if p, ok := newP.(*PolicySet); ok {
		if len(p.policies) == 2 {
			p1, ok1 := p.policies[0].(*Policy)
			p2, ok2 := p.policies[1].(*Policy)
			if ok1 && ok2 {
				if p1.id != "first" || p2.id != "third" {
					t.Errorf("Expected \"first\" and \"third\" policies remaining but got %q and %q", p1.id, p2.id)
				}
			} else {
				t.Errorf("Expected two policies after delete but got %T and %T", p.policies[0], p.policies[1])
			}
		} else {
			t.Errorf("Expected two policies after delete but got %d", len(p.policies))
		}

		assertMapperPCAMapKeys(p.algorithm, "after deletion", t, "first", "third")
	} else {
		t.Errorf("Expected new policy set but got %T (%#v)", newP, newP)
	}
}

func assertMapperPCAMapKeys(a PolicyCombiningAlg, desc string, t *testing.T, expected ...string) {
	if m, ok := a.(mapperPCA); ok {
		keys := []string{}
		for p := range m.policies.Enumerate() {
			keys = append(keys, p.Key)
		}

		assertStrings(keys, expected, desc, t)
	} else {
		t.Errorf("Expected mapper as policy combining algorithm but got %T for %s", a, desc)
	}
}
