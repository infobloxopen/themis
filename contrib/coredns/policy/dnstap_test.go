package policy

import (
	"net"
	"testing"

	"github.com/coredns/coredns/plugin/dnstap"
	"github.com/coredns/coredns/plugin/dnstap/taprw"
	dtest "github.com/coredns/coredns/plugin/dnstap/test"
	"github.com/coredns/coredns/plugin/test"
	"github.com/infobloxopen/themis/contrib/coredns/policy/testutil"
	"github.com/infobloxopen/themis/pdp"
	"github.com/miekg/dns"
	"golang.org/x/net/context"

	pb "github.com/infobloxopen/themis/contrib/coredns/policy/dnstap"
)

func TestSendCRExtraNoMsg(t *testing.T) {
	ok := false
	g := testutil.NewLogGrabber()
	defer func() {
		logs := g.Release()
		if ok {
			t.Logf("=== plugin logs ===\n%s--- plugin logs ---", logs)
		}
	}()

	trapper := dtest.TrapTapper{Full: true}
	tapRW := &taprw.ResponseWriter{
		Query:          new(dns.Msg),
		ResponseWriter: &test.ResponseWriter{},
		Tapper:         &trapper,
	}

	io := testutil.NewIORoutine()
	tapIO := newPolicyDnstapSender(io)
	tapIO.sendCRExtraMsg(tapRW, nil, nil)
	if !io.IsEmpty() {
		t.Errorf("Unexpected msg received")
	}
}

func TestSendCRExtraInvalidMsg(t *testing.T) {
	ok := false
	g := testutil.NewLogGrabber()
	defer func() {
		logs := g.Release()
		if ok {
			t.Logf("=== plugin logs ===\n%s--- plugin logs ---", logs)
		}
	}()

	msg := dns.Msg{}
	msg.SetQuestion("test.com.", dns.TypeA)
	msg.Answer = []dns.RR{
		test.A("test.com.       600 IN  A           10.240.0.1"),
	}
	msg.Rcode = -1

	trapper := dtest.TrapTapper{Full: true}
	tapRW := &taprw.ResponseWriter{
		Query:          new(dns.Msg),
		ResponseWriter: &test.ResponseWriter{},
		Tapper:         &trapper,
	}
	tapRW.WriteMsg(&msg)

	io := testutil.NewIORoutine()
	tapIO := newPolicyDnstapSender(io)
	tapIO.sendCRExtraMsg(tapRW, &msg, nil)
	if !io.IsEmpty() {
		t.Errorf("Unexpected msg received")
	}
}

func TestSendCRExtraMsg(t *testing.T) {
	ok := true
	g := testutil.NewLogGrabber()
	defer func() {
		logs := g.Release()
		if !ok {
			t.Logf("=== plugin logs ===\n%s--- plugin logs ---", logs)
		}
	}()

	msg := dns.Msg{}
	msg.SetQuestion("test.com.", dns.TypeA)
	msg.Answer = []dns.RR{
		test.A("test.com.       600 IN  A           10.240.0.1"),
	}

	trapper := dtest.TrapTapper{Full: true}
	tapRW := &taprw.ResponseWriter{
		Query:          new(dns.Msg),
		ResponseWriter: &test.ResponseWriter{},
		Tapper:         &trapper,
		Send:           &taprw.SendOption{Cq: false, Cr: false},
	}
	tapRW.WriteMsg(&msg)

	io := testutil.NewIORoutine()
	tapIO := newPolicyDnstapSender(io)

	testAttrHolder := &attrHolder{
		dnReq: []pdp.AttributeAssignment{
			pdp.MakeStringAssignment(attrNameType, typeValueQuery),
			pdp.MakeDomainAssignment(attrNameDomainName, testutil.MakeTestDomain("test.com")),
			pdp.MakeStringAssignment(attrNameDNSQtype, "1"),
			pdp.MakeAddressAssignment(attrNameSourceIP, net.ParseIP("10.0.0.7")),
			pdp.MakeStringAssignment("option", "option"),
		},
		dnstap: []pdp.AttributeAssignment{
			pdp.MakeStringAssignment("dnstap", "val"),
		},
		action: actionAllow,
	}

	tapIO.sendCRExtraMsg(tapRW, &msg, testAttrHolder)

	ok = testutil.AssertCRExtraResult(t, "sendCRExtraMsg(actionAllow)", io, &msg,
		&pb.DnstapAttribute{Id: attrNameSourceIP, Value: "10.0.0.7"},
		&pb.DnstapAttribute{Id: "option", Value: "option"},
		&pb.DnstapAttribute{Id: "dnstap", Value: "val"},
	)

	if l := len(trapper.Trap); l != 0 {
		t.Fatalf("Dnstap unexpectedly sent %d messages", l)
		ok = false
	}

	testAttrHolder.action = actionBlock

	tapIO.sendCRExtraMsg(tapRW, &msg, testAttrHolder)

	ok = testutil.AssertCRExtraResult(t, "sendCRExtraMsg(actionBlock)", io, &msg,
		&pb.DnstapAttribute{Id: attrNameDomainName, Value: "test.com"},
		&pb.DnstapAttribute{Id: attrNameDNSQtype, Value: "1"},
		&pb.DnstapAttribute{Id: attrNameSourceIP, Value: "10.0.0.7"},
		&pb.DnstapAttribute{Id: "option", Value: "option"},
		&pb.DnstapAttribute{Id: attrNamePolicyAction, Value: "3"},
		&pb.DnstapAttribute{Id: attrNameType, Value: typeValueQuery},
	) && ok

	if l := len(trapper.Trap); l != 0 {
		t.Fatalf("Dnstap unexpectedly sent %d messages", l)
		ok = false
	}
}

func TestRestCqCr(t *testing.T) {
	so := &taprw.SendOption{Cq: true, Cr: true}
	ctx := context.WithValue(context.Background(), dnstap.DnstapSendOption, so)
	resetCqCr(ctx)
	if so.Cq || so.Cr {
		t.Errorf("Failed to reset Cq/Cr flags")
	}
}
