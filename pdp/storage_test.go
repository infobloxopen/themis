package pdp

import (
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/google/uuid"
)

func TestStorage(t *testing.T) {
	root := &PolicySet{
		id: "test",
		policies: []Evaluable{
			&Policy{
				id:        "first",
				rules:     []*Rule{{id: "permit", effect: EffectPermit}},
				algorithm: firstApplicableEffectRCA{}}},
		algorithm: firstApplicableEffectPCA{}}

	s := NewPolicyStorage(root, nil, nil)
	sr := s.Root()
	if sr != root {
		t.Errorf("Expected stored root policy to be exactly root policy but got different")
	}
}

func TestStorageGetPath(t *testing.T) {
	targetRule := &Rule{id: "permit", effect: EffectPermit}
	targetPolicy := &Policy{
		id:        "first",
		rules:     []*Rule{targetRule},
		algorithm: firstApplicableEffectRCA{}}
	root := &PolicySet{
		id:        "test",
		policies:  []Evaluable{targetPolicy},
		algorithm: firstApplicableEffectPCA{}}
	s := NewPolicyStorage(root, nil, nil)

	expectRule, err := s.GetPath([]string{"test", "first", "permit"})
	if err != nil {
		t.Errorf("Got error %s", err.Error())
	} else if expectRule != targetRule {
		t.Errorf("Expecting iterator %v+, got %v+", targetRule, expectRule)
	}

	expectPolicy, err := s.GetPath([]string{"test", "first"})
	if err != nil {
		t.Errorf("Got error %s", err.Error())
	} else if expectPolicy != targetPolicy {
		t.Errorf("Expecting iterator %v+, got %v+", targetPolicy, expectPolicy)
	}

	expectRoot, err := s.GetPath([]string{"test"})
	if err != nil {
		t.Errorf("Got error %s", err.Error())
	} else if expectRoot != root {
		t.Errorf("Expecting iterator %v+, got %v+", root, expectRoot)
	}

	// expect failures

	// near miss (never match by substring)
	expectNil, err := s.GetPath([]string{"test", "first", "permits"})
	expectError(t, "Queried rule \"permits\" is not found", expectNil, err)

	// bad root
	expectNil, err = s.GetPath([]string{"first", "permit"})
	expectError(t, "Invalid root id or hidden root", expectNil, err)

	// hidden intermediate
	targetPolicy.hidden = true
	expectNil, err = s.GetPath([]string{"test", "first", "permit"})
	expectError(t, "Queried element \"first\" is not found", expectNil, err)

	// hidden root
	root.hidden = true
	expectNil, err = s.GetPath([]string{"test", "first", "permit"})
	expectError(t, "Invalid root id or hidden root", expectNil, err)
}

func TestGetSubtree(t *testing.T) {
	targetRule := &Rule{id: "permit", effect: EffectPermit}
	targetRule2 := &Rule{id: "permit2", effect: EffectPermit}
	targetRule3 := &Rule{id: "permit3", effect: EffectPermit}
	targetPolicy := &Policy{
		id:        "first",
		rules:     []*Rule{targetRule, targetRule2},
		algorithm: firstApplicableEffectRCA{}}
	targetPolicy2 := &Policy{
		id:        "third",
		rules:     []*Rule{targetRule3},
		algorithm: firstApplicableEffectRCA{}}
	targetPolicySet := &PolicySet{
		id:        "second",
		policies:  []Evaluable{targetPolicy2},
		algorithm: firstApplicableEffectPCA{}}
	root := &PolicySet{
		id:        "test",
		policies:  []Evaluable{targetPolicy, targetPolicySet},
		algorithm: firstApplicableEffectPCA{}}

	expectedRes := "{\"id\":\"test\",\"elems\":\"...\"}"
	expectRoot := GetSubtree(root, 0)
	if strings.Compare(expectedRes, expectRoot) != 0 {
		t.Errorf("Expecting json %v, got %v", expectedRes, expectRoot)
	}

	expectedRes = "{\"id\":\"first\",\"elems\":\"...\"}"
	expectPolicy := GetSubtree(targetPolicy, 0)
	if strings.Compare(expectedRes, expectPolicy) != 0 {
		t.Errorf("Expecting json %v, got %v", expectedRes, expectPolicy)
	}

	expectedRes = "{\"id\":\"permit\"}"
	expectRule := GetSubtree(targetRule, 0)
	if strings.Compare(expectedRes, expectRule) != 0 {
		t.Errorf("Expecting json %v, got %v", expectedRes, expectRule)
	}

	expectedRes = "{\"id\":\"test\",\"elems\":" +
		"[{\"id\":\"first\",\"elems\":\"...\"},{\"id\":\"second\",\"elems\":\"...\"}]}"
	expectTopTwo := GetSubtree(root, 1)
	if strings.Compare(expectedRes, expectTopTwo) != 0 {
		t.Errorf("Expecting json %v, got %v", expectedRes, expectTopTwo)
	}

	expectedRes = "{\"id\":\"test\",\"elems\":" +
		"[{\"id\":\"first\",\"elems\":" +
		"[{\"id\":\"permit\"},{\"id\":\"permit2\"}]" +
		"},{\"id\":\"second\",\"elems\":" +
		"[{\"id\":\"third\",\"elems\":\"...\"}]}]}"
	expectTopFive := GetSubtree(root, 2)
	if strings.Compare(expectedRes, expectTopFive) != 0 {
		t.Errorf("Expecting json %v+, got %v+", expectedRes, expectTopFive)
	}

	expectedRes = "{\"id\":\"test\",\"elems\":" +
		"[{\"id\":\"first\",\"elems\":" +
		"[{\"id\":\"permit\"},{\"id\":\"permit2\"}]" +
		"},{\"id\":\"second\",\"elems\":" +
		"[{\"id\":\"third\",\"elems\":[{\"id\":\"permit3\"}]}]}]}"
	expectAll := GetSubtree(root, 5)
	if strings.Compare(expectedRes, expectAll) != 0 {
		t.Errorf("Expecting json %v+, got %v+", expectedRes, expectAll)
	}
}

func TestPathQuery(t *testing.T) {
	targetRule := &Rule{id: "permit", effect: EffectPermit}
	targetRule2 := &Rule{id: "permit2", effect: EffectPermit}
	targetRule3 := &Rule{id: "permit3", effect: EffectPermit}
	targetPolicy := &Policy{
		id:        "first",
		rules:     []*Rule{targetRule, targetRule2},
		algorithm: firstApplicableEffectRCA{}}
	targetPolicy2 := &Policy{
		id:        "third",
		rules:     []*Rule{targetRule3},
		algorithm: firstApplicableEffectRCA{}}
	targetPolicySet := &PolicySet{
		id:        "second",
		policies:  []Evaluable{targetPolicy2},
		algorithm: firstApplicableEffectPCA{}}
	root := &PolicySet{
		id:        "test",
		policies:  []Evaluable{targetPolicy, targetPolicySet},
		algorithm: firstApplicableEffectPCA{}}

	// search from root
	expectPath := []string{"first", "permit2"}
	path, expectRule2, err := PathQuery(root, "permit2")
	if err != nil {
		t.Errorf("Got error %s", err.Error())
	} else if expectRule2 != targetRule2 {
		t.Errorf("Expecting iterator %v+, got %v+", targetRule2, expectRule2)
	} else if !reflect.DeepEqual(expectPath, path) {
		t.Errorf("Expecting path %v, got %v", expectPath, path)
	}

	expectPath = []string{"first", "permit"}
	path, expectRule, err := PathQuery(root, "permit")
	if err != nil {
		t.Errorf("Got error %s", err.Error())
	} else if expectRule != targetRule {
		t.Errorf("Expecting iterator %v+, got %v+", targetRule, expectRule)
	} else if !reflect.DeepEqual(expectPath, path) {
		t.Errorf("Expecting path %v, got %v", expectPath, path)
	}

	expectPath = []string{"second", "third"}
	path, expectPolicy2, err := PathQuery(root, "third")
	if err != nil {
		t.Errorf("Got error %s", err.Error())
	} else if expectPolicy2 != targetPolicy2 {
		t.Errorf("Expecting iterator %v+, got %v+", targetPolicy2, expectPolicy2)
	} else if !reflect.DeepEqual(expectPath, path) {
		t.Errorf("Expecting path %v, got %v", expectPath, path)
	}

	// search from subtree
	expectPath = []string{"third", "permit3"}
	path, expectRule3, err := PathQuery(targetPolicySet, "permit3")
	if err != nil {
		t.Errorf("Got error %s", err.Error())
	} else if expectRule3 != targetRule3 {
		t.Errorf("Expecting iterator %v+, got %v+", targetRule3, expectRule3)
	} else if !reflect.DeepEqual(expectPath, path) {
		t.Errorf("Expecting path %v, got %v", expectPath, path)
	}

	// expect failures

	// search for non-existent node
	_, expectNil, err := PathQuery(root, "non-existent")
	expectError(t, "Element \"non-existent\" not found", expectNil, err)

	// wrong subtree
	_, expectNil, err = PathQuery(targetPolicySet, "permit")
	expectError(t, "Element \"permit\" not found", expectNil, err)
}

func TestStorageNewTransaction(t *testing.T) {
	initialTag := uuid.New()

	s := &PolicyStorage{tag: &initialTag}
	tr, err := s.NewTransaction(&initialTag)
	if err != nil {
		t.Errorf("Expected no error but got %s", err)
	} else if tr == nil {
		t.Errorf("Expected transaction but got nothing")
	}

	s = &PolicyStorage{}
	tr, err = s.NewTransaction(&initialTag)
	if err == nil {
		t.Errorf("Expected error but got transaction %#v", tr)
	} else if _, ok := err.(*UntaggedPolicyModificationError); !ok {
		t.Errorf("Expected *untaggedPolicyModificationError but got %T (%s)", err, err)
	}

	s = &PolicyStorage{tag: &initialTag}
	tr, err = s.NewTransaction(nil)
	if err == nil {
		t.Errorf("Expected error but got transaction %#v", tr)
	} else if _, ok := err.(*MissingPolicyTagError); !ok {
		t.Errorf("Expected *missingPolicyTagError but got %T (%s)", err, err)
	}

	otherTag := uuid.New()
	s = &PolicyStorage{tag: &initialTag}
	tr, err = s.NewTransaction(&otherTag)
	if err == nil {
		t.Errorf("Expected error but got transaction %#v", tr)
	} else if _, ok := err.(*PolicyTagsNotMatchError); !ok {
		t.Errorf("Expected *policyTagsNotMatchError but got %T (%s)", err, err)
	}
}

func TestStorageCommitTransaction(t *testing.T) {
	initialTag := uuid.New()
	newTag := uuid.New()

	s := &PolicyStorage{tag: &initialTag}
	tr, err := s.NewTransaction(&initialTag)
	if err != nil {
		t.Errorf("Expected no error but got %s", err)
	} else {
		u := NewPolicyUpdate(initialTag, newTag)
		err := tr.Apply(u)
		if err != nil {
			t.Errorf("Expected no error but got %s", err)
		} else {
			newS, err := tr.Commit()
			if err != nil {
				t.Errorf("Expected no error but got %s", err)
			} else {
				if &newS == &s {
					t.Errorf("Expected other storage instance but got the same")
				}

				if newS.tag.String() != newTag.String() {
					t.Errorf("Expected tag %s but got %s", newTag.String(), newS.tag.String())
				}
			}
		}
	}

	tr, err = s.NewTransaction(&initialTag)
	if err != nil {
		t.Errorf("Expected no error but got %s", err)
	} else {
		tr.err = newUnknownPolicyUpdateOperationError(-1)
		s, err := tr.Commit()
		if err == nil {
			t.Errorf("Expected error but got storage %#v", s)
		} else if _, ok := err.(*failedPolicyTransactionError); !ok {
			t.Errorf("Expected *failedPolicyTransactionError but got %T (%s)", err, err)
		}
	}
}

func TestStorageModifications(t *testing.T) {
	tag := uuid.New()

	s := &PolicyStorage{
		tag: &tag,
		policies: &Policy{
			id:        "test",
			algorithm: firstApplicableEffectRCA{}}}
	tr, err := s.NewTransaction(&tag)
	if err != nil {
		t.Errorf("Expected no error but got %s", err)
	} else {
		err := tr.appendItem([]string{"test"}, &Rule{id: "permit", effect: EffectPermit})
		if err != nil {
			t.Errorf("Expected no error but got %s", err)
		}

		err = tr.appendItem(nil, &Rule{id: "permit", effect: EffectPermit})
		if err == nil {
			t.Errorf("Expected error but got nothing")
		} else if _, ok := err.(*invalidRootPolicyItemTypeError); !ok {
			t.Errorf("Expected *invalidRootPolicyItemTypeError but got %T (%s)", err, err)
		}

		err = tr.appendItem(nil, &Policy{hidden: true})
		if err == nil {
			t.Errorf("Expected error but got nothing")
		} else if _, ok := err.(*hiddenRootPolicyAppendError); !ok {
			t.Errorf("Expected *hiddenRootPolicyAppendError but got %T (%s)", err, err)
		}

		err = tr.appendItem(nil, &Policy{
			id:        "test",
			rules:     []*Rule{{id: "permit", effect: EffectPermit}},
			algorithm: firstApplicableEffectRCA{}})
		if err != nil {
			t.Errorf("Expected no error but got %s", err)
		}

		err = tr.appendItem([]string{"example"}, &Rule{id: "permit", effect: EffectPermit})
		if err == nil {
			t.Errorf("Expected error but got nothing")
		} else if _, ok := err.(*invalidRootPolicyError); !ok {
			t.Errorf("Expected *invalidRootPolicyError but got %T (%s)", err, err)
		}

		err = tr.appendItem([]string{"test"}, &Rule{hidden: true})
		if err == nil {
			t.Errorf("Expected error but got nothing")
		} else if _, ok := err.(*hiddenRuleAppendError); !ok {
			t.Errorf("Expected *hiddenRuleAppendError but got %T (%s)", err, err)
		}

		err = tr.del(nil)
		if err == nil {
			t.Errorf("Expected error but got nothing")
		} else if _, ok := err.(*emptyPathModificationError); !ok {
			t.Errorf("Expected *emptyPathModificationError but got %T (%s)", err, err)
		}

		err = tr.del([]string{"example"})
		if err == nil {
			t.Errorf("Expected error but got nothing")
		} else if _, ok := err.(*invalidRootPolicyError); !ok {
			t.Errorf("Expected *invalidRootPolicyError but got %T (%s)", err, err)
		}

		err = tr.del([]string{"test", "example"})
		if err == nil {
			t.Errorf("Expected error but got nothing")
		} else if _, ok := err.(*missingPolicyChildError); !ok {
			t.Errorf("Expected *missingPolicyChildError but got %T (%s)", err, err)
		}

		err = tr.del([]string{"test", "permit"})
		if err != nil {
			t.Errorf("Expected no error but got %s", err)
		}

		err = tr.del([]string{"test"})
		if err != nil {
			t.Errorf("Expected no error but got %s", err)
		}

		if tr.policies != nil {
			t.Errorf("Expected no root policy but got %#v", tr.policies)
		}
	}
}

func TestStorageTransactionalUpdate(t *testing.T) {
	tag := uuid.New()

	root := &PolicySet{
		id: "test",
		policies: []Evaluable{
			&Policy{
				id:     "first",
				target: makeSimpleStringTarget("s", "test"),
				rules: []*Rule{
					{
						id:          "permit",
						effect:      EffectPermit,
						obligations: makeSingleStringObligation("s", "permit")}},
				algorithm: denyOverridesRCA{}},
			&Policy{
				id: "del",
				rules: []*Rule{
					{
						id:          "permit",
						effect:      EffectPermit,
						obligations: makeSingleStringObligation("s", "del-permit")}},
				algorithm: firstApplicableEffectRCA{}}},
		algorithm: firstApplicableEffectPCA{}}

	s := NewPolicyStorage(root, map[string]Attribute{"s": MakeAttribute("s", TypeString)}, &tag)

	newTag := uuid.New()

	u := NewPolicyUpdate(tag, newTag)
	u.Append(UOAdd, []string{"test", "first"}, &Rule{
		id:          "deny",
		effect:      EffectDeny,
		obligations: makeSingleStringObligation("s", "deny")})
	u.Append(UODelete, []string{"test", "del"}, nil)

	eUpd := fmt.Sprintf("policy update: %s - %s\n"+
		"commands:\n"+
		"- Add path (\"test\"/\"first\")\n"+
		"- Delete path (\"test\"/\"del\")", tag.String(), newTag.String())
	sUpd := u.String()
	if sUpd != eUpd {
		t.Errorf("Expected:\n%s\n\nupdate but got:\n%s\n\n", eUpd, sUpd)
	}

	tr, err := s.NewTransaction(&tag)
	if err != nil {
		t.Fatalf("Expected no error but got %T (%s)", err, err)
	}

	attrs := tr.Attributes()
	if len(attrs) != 1 {
		t.Fatalf("Expected one attribute but got %#v", attrs)
	}
	if _, ok := attrs["s"]; !ok {
		t.Errorf("Expected %q attribute but got %#v", "s", attrs)
	}

	err = tr.Apply(u)
	if err != nil {
		t.Fatalf("Expected no error but got %T (%s)", err, err)
	}

	s, err = tr.Commit()
	if err != nil {
		t.Fatalf("Expected no error but got %T (%s)", err, err)
	}

	ctx, err := NewContext(nil, 1, func(i int) (string, AttributeValue, error) {
		return "s", MakeStringValue("test"), nil
	})
	if err != nil {
		t.Errorf("Expected no error but got %T (%s)", err, err)
	} else {
		r := s.Root().Calculate(ctx)
		effect, o, err := r.Status()
		if err != nil {
			t.Errorf("Expected no error but got %T (%s)", err, err)
		}

		if effect != EffectDeny {
			t.Errorf("Expected deny effect but got %d", effect)
		}

		if len(o) < 1 {
			t.Error("Expected at least one obligation")
		} else {
			_, _, v, err := o[0].Serialize(ctx)
			if err != nil {
				t.Errorf("Expected no error but got %T (%s)", err, err)
			} else {
				e := "deny"
				if v != e {
					t.Errorf("Expected %q obligation but got %q", e, v)
				}
			}
		}
	}

	ctx, err = NewContext(nil, 1, func(i int) (string, AttributeValue, error) {
		return "s", MakeStringValue("no test"), nil
	})
	if err != nil {
		t.Errorf("Expected no error but got %T (%s)", err, err)
	} else {
		r := s.Root().Calculate(ctx)
		effect, _, err := r.Status()
		if err != nil {
			t.Errorf("Expected no error but got %T (%s)", err, err)
		}

		if effect != EffectNotApplicable {
			t.Errorf("Expected \"not applicable\" effect but got %d", effect)
		}
	}
}
