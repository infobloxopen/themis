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

func TestRuleDepthMarshal(t *testing.T) {
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

	// bad depth
	err := rule.DepthMarshal(&buf, -1)
	expectErrMsg := "depth must be >= 0, got -1"
	if err == nil {
		t.Errorf("Expecting error message %s, got nil error", expectErrMsg)
	} else if 0 != strings.Compare(err.Error(), expectErrMsg) {
		t.Errorf("Expecting error message %s, got %s", expectErrMsg, err.Error())
	}

	// good depth, visible rule
	expectMarshal := `{"ord": 32, "id": "one"}`
	err = rule.DepthMarshal(&buf, 0)
	if err != nil {
		t.Errorf("Expecting no error, got %v", err)
	} else {
		gotMarshal := buf.String()
		if 0 != strings.Compare(gotMarshal, expectMarshal) {
			t.Errorf("Expecting marshal output %s, got %s", expectMarshal, gotMarshal)
		}
	}

	// good depth, hidden rule
	err = hiddenRule.DepthMarshal(&buf, 0)
	if err != errHiddenRule {
		t.Errorf("Expecting error %v, got %v", errHiddenRule, err)
	}
}
