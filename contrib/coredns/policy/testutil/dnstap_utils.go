package testutil

import (
	"fmt"
	"testing"

	dtest "github.com/coredns/coredns/plugin/dnstap/test"
	tap "github.com/dnstap/golang-dnstap"
	"github.com/golang/protobuf/proto"
	"github.com/miekg/dns"
	"github.com/pmezard/go-difflib/difflib"

	pb "github.com/infobloxopen/themis/contrib/coredns/policy/dnstap"
)

type testIORoutine struct {
	msg tap.Dnstap
}

func NewIORoutine() *testIORoutine {
	return &testIORoutine{}
}

func (t *testIORoutine) Dnstap(msg tap.Dnstap) {
	t.msg = msg
}

func (t *testIORoutine) IsEmpty() bool {
	return t.msg.Message == nil || t.msg.Type == nil
}

func AssertCRExtraResult(t *testing.T, desc string, io *testIORoutine, eMsg *dns.Msg, e ...*pb.DnstapAttribute) bool {
	dnstapMsg := io.msg

	extra := &pb.Extra{}
	err := proto.Unmarshal(dnstapMsg.Extra, extra)
	if err != nil {
		t.Errorf("Failed to unmarshal Extra for %q (%v)", desc, err)
		return false
	}

	ok := AssertDnstapAttributes(t, desc, extra.GetAttrs(), e...)
	ok = assertCRMessage(t, desc, dnstapMsg.Message, eMsg) && ok
	return ok
}

func assertCRMessage(t *testing.T, desc string, msg *tap.Message, e *dns.Msg) bool {
	if e == nil {
		t.Errorf("Expected CR message for %q not found", desc)
		return false
	}

	if msg == nil {
		t.Errorf("CR message for %q not found", desc)
		return false
	}

	bin, err := e.Pack()
	if err != nil {
		t.Errorf("Failed to pack message for %q (%v)", desc, err)
		return false
	}

	d := dtest.TestingData()
	d.Packed = bin
	eMsg, _ := d.ToClientResponse()
	if !dtest.MsgEqual(eMsg, msg) {
		t.Errorf("Unexpected message for %q: expected:\n%v\nactual:\n%v", desc, eMsg, msg)
		return false
	}

	return true
}

func serializeDnstapAttributesForAssert(a []*pb.DnstapAttribute) []string {
	out := make([]string, len(a))
	for i, a := range a {
		out[i] = fmt.Sprintf("%q = %q\n", a.Id, a.Value)
	}

	return out
}

func AssertDnstapAttributes(t *testing.T, desc string, a []*pb.DnstapAttribute, e ...*pb.DnstapAttribute) bool {
	ctx := difflib.ContextDiff{
		A:        serializeDnstapAttributesForAssert(e),
		B:        serializeDnstapAttributesForAssert(a),
		FromFile: "Expected",
		ToFile:   "Got"}

	diff, err := difflib.GetContextDiffString(ctx)
	if err != nil {
		panic(fmt.Errorf("can't compare \"%s\": %s", desc, err))
	}

	if len(diff) > 0 {
		t.Errorf("\"%s\" doesn't match:\n%s", desc, diff)
		return false
	}

	return true
}
