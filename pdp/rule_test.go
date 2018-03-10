package pdp

import (
	"sort"
	"strings"
	"testing"
)

func TestSortRulesByOrder(t *testing.T) {
	rules := []*Rule{
		{
			ord: 1,
			id:  "second",
		},
		{
			ord: 3,
			id:  "fourth",
		},
		{
			ord: 0,
			id:  "first",
		},
		{
			ord: 2,
			id:  "third",
		},
	}

	sort.Sort(byRuleOrder(rules))

	ids := make([]string, len(rules))
	for i, r := range rules {
		ids[i] = r.id
	}
	s := strings.Join(ids, ", ")
	e := "first, second, third, fourth"
	if s != e {
		t.Errorf("Expected rules in order \"%s\" but got \"%s\"", e, s)
	}
}

func TestRuleFindNext(t *testing.T) {
	// public policy test
	rule := makeSimpleRule("test", EffectPermit)

	expectNil, err := rule.FindNext("anything")
	expectError(t, "Rules are always leaves, element \"anything\" not found", expectNil, err)
}

func TestRuleNext(t *testing.T) {
	// public policy test
	rule := makeSimpleRule("test", EffectPermit)

	expectZero := rule.NextSize()
	if expectZero != 0 {
		t.Errorf("Expecting 0 children of rule, but got %d", expectZero)
	}

	expectNil := rule.GetNext(0)
	if expectNil != nil {
		t.Errorf("Expecting nil policy, but got %v+", expectNil)
	}
}
