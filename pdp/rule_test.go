package pdp

import (
	"bytes"
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

func TestRuleMarshalWithDepth(t *testing.T) {
	var (
		buf  bytes.Buffer
		rule = Rule{
			ord: 32,
			id:  "one",
		}
	)

	// bad depth
	err := rule.MarshalWithDepth(&buf, -1)
	expectErr := newMarshalInvalidDepthError(-1)
	if err == nil {
		t.Errorf("Expecting error %v, got nil error", expectErr)
	} else if err.Error() != expectErr.Error() {
		t.Errorf("Expecting error %v, got %v", expectErr, err)
	}

	// good depth, visible rule
	expectMarshal := `{"ord":32,"id":"one"}`
	err = rule.MarshalWithDepth(&buf, 0)
	if err != nil {
		t.Errorf("Expecting no error, got %v", err)
	} else {
		gotMarshal := buf.String()
		if 0 != strings.Compare(gotMarshal, expectMarshal) {
			t.Errorf("Expecting marshal output %s, got %s", expectMarshal, gotMarshal)
		}
	}
}

func TestRulePathMarshal(t *testing.T) {
	var (
		buf  bytes.Buffer
		rule = Rule{
			ord: 32,
			id:  "one",
		}
		hiddenRule = Rule{
			ord:    32,
			id:     "",
			hidden: true,
		}
	)
	pathfinder, found := rule.PathMarshal("one")
	if !found {
		t.Errorf("Failed to find path to rule one")
	} else if err := pathfinder(&buf); err != nil {
		t.Errorf("Expecting no errors when writing path, got %v", err)
	}
	expectPath := "\"one\""
	if 0 != strings.Compare(buf.String(), expectPath) {
		t.Errorf("Expecting path %s, got %s", buf.String(), expectPath)
	}

	expectNil, found := rule.PathMarshal("two")
	if found {
		t.Errorf("Expecting not to find rule two in rule one")
	} else if expectNil != nil {
		t.Errorf("Expecting nil path callback, got non-nil")
	}
	expectNil, found = hiddenRule.PathMarshal("one")
	if found {
		t.Errorf("Expecting not to find rule one in hidden rule")
	} else if expectNil != nil {
		t.Errorf("Expecting nil path callback, got non-nil")
	}
}
