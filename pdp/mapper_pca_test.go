package pdp

import "testing"

func TestMapperPCAOrders(t *testing.T) {
	if totalMapperPCAOrders != len(MapperPCAOrderNames) {
		t.Errorf("Expected total number of order values to be equal to number of their names "+
			"but got totalMapperPCAOrders = %d and len(MapperPCAOrderNames) = %d",
			totalMapperPCAOrders, len(MapperPCAOrderNames))
	}
}

func TestMapperPCAOrdering(t *testing.T) {
	c := &Context{
		a: map[string]interface{}{
			"k": MakeListOfStringsValue([]string{"third", "first", "second"}),
		},
	}

	p := NewPolicySet("test", false, Target{},
		[]Evaluable{
			makeSimplePermitPolicyWithObligations(
				"first",
				makeSingleStringObligation("order", "first"),
			),
			makeSimplePermitPolicyWithObligations(
				"second",
				makeSingleStringObligation("order", "second"),
			),
			makeSimplePermitPolicyWithObligations(
				"third",
				makeSingleStringObligation("order", "third"),
			),
		},
		makeMapperPCA, MapperPCAParams{
			Argument:  AttributeDesignator{a: Attribute{id: "k", t: TypeListOfStrings}},
			Order:     MapperPCAInternalOrder,
			Algorithm: firstApplicableEffectPCA{},
		},
		nil,
	)

	r := p.Calculate(c)
	if len(r.obligations) != 1 {
		t.Errorf("Expected the only obligation got %#v", r.obligations)
	} else {
		ID, ot, s, err := r.obligations[0].Serialize(c)
		if err != nil {
			t.Errorf("Expected no error but got %s", err)
		}

		eID := "order"
		eot := BuiltinTypeKeys[TypeString]
		es := "first"
		if ID != eID || ot != eot || s != es {
			t.Errorf("Expected %q = %q.(%s) obligation but got %q = %q.(%s)", eID, es, eot, ID, s, ot)
		}
	}

	p = NewPolicySet("test", false, Target{},
		[]Evaluable{
			makeSimplePermitPolicyWithObligations(
				"first",
				makeSingleStringObligation("order", "first"),
			),
			makeSimplePermitPolicyWithObligations(
				"second",
				makeSingleStringObligation("order", "second"),
			),
			makeSimplePermitPolicyWithObligations(
				"third",
				makeSingleStringObligation("order", "third"),
			),
		},
		makeMapperPCA, MapperPCAParams{
			Argument:  AttributeDesignator{a: Attribute{id: "k", t: TypeListOfStrings}},
			Order:     MapperPCAExternalOrder,
			Algorithm: firstApplicableEffectPCA{},
		},
		nil,
	)

	r = p.Calculate(c)
	if len(r.obligations) != 1 {
		t.Errorf("Expected the only obligation got %#v", r.obligations)
	} else {
		ID, ot, s, err := r.obligations[0].Serialize(c)
		if err != nil {
			t.Errorf("Expected no error but got %s", err)
		}

		eID := "order"
		eot := BuiltinTypeKeys[TypeString]
		es := "third"
		if ID != eID || ot != eot || s != es {
			t.Errorf("Expected %q = %q.(%s) obligation but got %q = %q.(%s)", eID, es, eot, ID, s, ot)
		}
	}
}
