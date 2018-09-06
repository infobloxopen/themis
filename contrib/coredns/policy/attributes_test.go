package policy

import (
	"testing"

	"github.com/infobloxopen/themis/contrib/coredns/policy/testutil"
	"github.com/infobloxopen/themis/pdp"
)

func TestSerializeOrPanic(t *testing.T) {
	s := serializeOrPanic(pdp.MakeStringAssignment("s", "test"))
	if s != "test" {
		t.Errorf("expected %q but got %q", "test", s)
	}

	testutil.AssertPanicWithErrorContains(t, "serializeOrPanic(expression)", func() {
		serializeOrPanic(pdp.MakeExpressionAssignment("s", pdp.MakeStringDesignator("s")))
	}, "pdp.AttributeDesignator")

	testutil.AssertPanicWithErrorContains(t, "serializeOrPanic(undefined)", func() {
		serializeOrPanic(pdp.MakeExpressionAssignment("s", pdp.UndefinedValue))
	}, "Undefined")
}

func TestCustAttr(t *testing.T) {
	if !custAttr(custAttrEdns).isEdns() {
		t.Errorf("expected %d is EDNS", custAttrEdns)
	}

	if !custAttr(custAttrTransfer).isTransfer() {
		t.Errorf("expected %d is transfer", custAttrTransfer)
	}

	if !custAttr(custAttrDnstap).isDnstap() {
		t.Errorf("expected %d is DNStap", custAttrDnstap)
	}
}
