package pdp

import (
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
		t.Errorf("Expected stored root policy to be exactly root policy but got different %p != %p", sr, root)
	}
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
		u, err := NewPolicyUpdate(initialTag, newTag)
		if err != nil {
			t.Errorf("Expected no error but got %s", err)
		} else {
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
