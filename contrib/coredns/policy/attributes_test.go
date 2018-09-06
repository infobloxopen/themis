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

func TestNewAttrsConfig(t *testing.T) {
	cfg := newAttrsConfig()
	if len(cfg.attrInds) != attrIndexCount {
		t.Errorf("Incorrect initialization of attribute map")
	}
}

func TestProvideIndex(t *testing.T) {
	cfg := newAttrsConfig()
	stdlen := len(cfg.attrInds)

	tcs := []struct {
		name string
		ind  int
	}{
		{attrNamePolicyAction, attrIndexPolicyAction},
		{attrNameSourceIP, attrIndexSourceIP},
		{attrNameLog, attrIndexLog},
		{attrNameDomainName, attrIndexDomainName},
		{attrNameAddress, attrIndexAddress},
		{attrNameRedirectTo, attrIndexRedirectTo},
		{attrNameDNSQtype, attrIndexDNSQtype},
		{"NewAttribute1", stdlen},
		{"NewAttribute2", stdlen + 1},
		{"NewAttribute1", stdlen},
	}
	for _, tc := range tcs {
		if ind := cfg.provideIndex(tc.name); ind != tc.ind {
			t.Errorf("Incorrect mapping for %q: expected %d, got %d", tc.name, tc.ind, ind)
		}
	}
}

func TestParseAttrArg(t *testing.T) {
	tcs := []struct {
		input  string
		name   string
		stype  string
		svalue string
		err    bool
	}{
		{"attr1", "attr1", "Undefined", "", false},
		{"attr2=`strVal1`", "attr2", "String", "strVal1", false},
		{"attr3=\"strVal2\"", "attr3", "String", "strVal2", false},
		{"attr4=strVal3", "", "Undefined", "", true}, // quotes are obligatory for strings
		{"attr5=127.0.0.1", "attr5", "Address", "127.0.0.1", false},
		{"attr6=1245:abcd::77fe", "attr6", "Address", "1245:abcd::77fe", false},
		{"attr7=::baba:77fe", "attr7", "Address", "::baba:77fe", false},
		{"attr8=:baba::77fe", "", "Undefined", "", true},
		{"attr9=123456789123456789", "attr9", "Integer", "123456789123456789", false},
		{"attrA=0x2694", "attrA", "Integer", "9876", false},
		{"attrB=02322", "attrB", "Integer", "1234", false},
		{"attrC=568word", "", "Undefined", "", true},
		{"attrD=568.321", "", "Undefined", "", true}, //float is not supported for now
		{"attrE=-333", "attrE", "Integer", "-333", false},
		{"attrF=`str\"Val1\"`", "attrF", "String", "str\"Val1\"", false},
	}
	for i, tc := range tcs {
		n, v, e := parseAttrArg(tc.input)
		if (e != nil) != tc.err {
			t.Errorf("TC#%d: incorrect error status", i)
			continue
		}
		if n != tc.name {
			t.Errorf("TC#%d: incorrect name: expected %s, got %s", i, tc.name, n)
		}
		vt := v.GetResultType().String()
		if vt != tc.stype {
			t.Errorf("TC#%d: incorrect type: expected %s, got %s", i, tc.stype, vt)
		}
		if vt == "Undefined" {
			continue
		}
		vs, _ := v.Serialize()
		if vs != tc.svalue {
			t.Errorf("TC#%d: incorrect value: expected %s, got %s", i, tc.svalue, vs)
		}
	}
}

func TestParseAttrList(t *testing.T) {
	cfg := newAttrsConfig()
	stdlen := len(cfg.attrInds)

	// incorrect input
	err := cfg.parseAttrList(attrListTypeDefDecision, "incorrect=input")
	if err == nil {
		t.Errorf("Unexpected successful call")
	}
	if len(cfg.confLists[attrListTypeDefDecision]) != 0 {
		t.Errorf("Config list is not empty")
	}

	// correct input
	err = cfg.parseAttrList(attrListTypeDefDecision, "correct=123", attrNameLog)
	if err != nil {
		t.Errorf("Unexpected failed call")
	}
	exp := 2
	if l := len(cfg.confLists[attrListTypeDefDecision]); l != exp {
		t.Errorf("Unexpected size of config list; expected %d, got %d", exp, l)
	}
	for i, expInd := range []int{stdlen, attrIndexLog} {
		if ind := cfg.confLists[attrListTypeDefDecision][i].index; ind != expInd {
			t.Errorf("Unexpected index in config list; expected %d, got %d", expInd, ind)
		}
	}

	// incorrect list id
	testutil.AssertPanicWithErrorContains(t, "serializeOrPanic(expression)", func() {
		cfg.parseAttrList(10000, "incorrect_list_id")
	}, "index out of range")
}
