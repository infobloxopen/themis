package testutil

import (
	"fmt"
	"strings"
	"testing"

	"github.com/infobloxopen/go-trees/domain"
	pb "github.com/infobloxopen/themis/contrib/coredns/policy/dnstap"
	"github.com/infobloxopen/themis/pdp"
)

var emptyCtx, _ = pdp.NewContext(nil, 0, nil)

func MakeTestDomain(s string) domain.Name {
	dn, err := domain.MakeNameFromString(s)
	if err != nil {
		panic(err)
	}

	return dn
}

func AssertPanicWithError(t *testing.T, desc string, f func(), format string, args ...interface{}) {
	defer func() {
		if r := recover(); r != nil {
			e := fmt.Sprintf(format, args...)
			err, ok := r.(error)
			if !ok {
				t.Errorf("excpected error %q on panic for %q but got %T (%#v)", e, desc, r, r)
			} else if err.Error() != e {
				t.Errorf("excpected error %q on panic for %q but got %q", e, desc, r)
			}
		} else {
			t.Errorf("expected panic %q for %q", fmt.Sprintf(format, args...), desc)
		}
	}()

	f()
}

func AssertPanicWithErrorContains(t *testing.T, desc string, f func(), e string) {
	defer func() {
		if r := recover(); r != nil {
			err, ok := r.(error)
			if !ok {
				t.Errorf("excpected error containing %q on panic for %q but got %T (%#v)", e, desc, r, r)
			} else if !strings.Contains(err.Error(), e) {
				t.Errorf("excpected error containing %q on panic for %q but got %q", e, desc, r)
			}
		} else {
			t.Errorf("expected error containing %q on panic for %q", e, desc)
		}
	}()

	f()
}

func AssertValue(t *testing.T, i int, av pdp.AttributeValue, ev pdp.AttributeValue) {
	t.Helper()
	if !ev.GetResultType().Match(av.GetResultType()) {
		t.Errorf("Unexpected type of %d attr, expected %s, actual %s", i, ev.GetResultType(), av.GetResultType())
		return
	}
	if ev.GetResultType().String() == "Undefined" {
		return
	}
	evs, _ := ev.Serialize()
	avs, _ := av.Serialize()
	if evs != avs {
		t.Errorf("Unexpected value of %d attr, expected %s, actual %s", i, evs, avs)
	}
}

func AssertAttr(t *testing.T, i int, a pdp.AttributeAssignment, e pdp.AttributeAssignment) {
	t.Helper()
	en := e.GetID()
	an := a.GetID()
	if en != an {
		t.Errorf("Unexpected name of %d attr, expected %s, actual %s", i, en, an)
		return
	}
	ev, _ := e.GetValue()
	av, _ := a.GetValue()
	AssertValue(t, i, av, ev)
}

func AssertAttrList(t *testing.T, actual []pdp.AttributeAssignment, expected ...pdp.AttributeAssignment) {
	t.Helper()
	if len(actual) != len(expected) {
		t.Errorf("Unexpected attr list size:")
		t.Errorf("    expected: %v", expected)
		t.Errorf("      actual: %v", actual)
		return
	}
	for i, e := range expected {
		AssertAttr(t, i, actual[i], e)
	}
}

func AssertDnstapList(t *testing.T, actual []*pb.DnstapAttribute, expected ...*pb.DnstapAttribute) {
	t.Helper()
	if len(actual) != len(expected) {
		t.Errorf("Unexpected attr list size:")
		t.Errorf("    expected: %v", expected)
		t.Errorf("      actual: %v", actual)
		return
	}
	for i, e := range expected {
		if e.GetId() != actual[i].GetId() {
			t.Errorf("Unexpected value of attr %d: expected %s, actual %s", i, e.GetId(), actual[i].GetId())
		}
		if e.GetValue() != actual[i].GetValue() {
			t.Errorf("Unexpected value of attr %d: expected %s, actual %s", i, e.GetValue(), actual[i].GetValue())
		}
	}
}

func MakeDnOrFail(t *testing.T, qName string) domain.Name {
	t.Helper()
	dn, err := domain.MakeNameFromString(qName)
	if err != nil {
		t.Fatalf("Can't create domain name for %s: %s", qName, err)
	}
	return dn
}

func AssertPdpResponse(t *testing.T, r *pdp.Response, expEff int, o ...pdp.AttributeAssignment) {
	t.Helper()
	if r.Effect != expEff {
		t.Errorf("Unexpected effect: expected %d, actual %d", expEff, r.Effect)
	}
	if len(r.Obligations) != len(o) {
		t.Errorf("Unexpected obligation count: expected %d, actual %d", len(o), len(r.Obligations))
	}
	for i, e := range o {
		AssertAttr(t, i, r.Obligations[i], e)
	}
}
