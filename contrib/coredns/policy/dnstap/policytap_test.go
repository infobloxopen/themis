package policytap

import (
	"bytes"
	"testing"
	"time"

	"github.com/coredns/coredns/plugin/dnstap/taprw"
	dtest "github.com/coredns/coredns/plugin/dnstap/test"
	"github.com/coredns/coredns/plugin/test"
	dnstap "github.com/dnstap/golang-dnstap"
	"github.com/golang/protobuf/proto"
	pb "github.com/infobloxopen/themis/pdp-service"
	"github.com/miekg/dns"
)

type testIORoutine struct {
	dnstapChan chan dnstap.Dnstap
}

func newIORoutine() testIORoutine {
	ch := make(chan dnstap.Dnstap, 1)
	tapIO := testIORoutine{dnstapChan: ch}
	// close channel by timeout to prevent checker from waiting forever
	go func() {
		time.Sleep(5 * time.Second)
		close(ch)
	}()
	return tapIO
}

func (tapIO testIORoutine) Dnstap(msg dnstap.Dnstap) {
	tapIO.dnstapChan <- msg
}

type commonData struct {
	sec  uint64
	nsec uint32
	qID  uint16
}

type testData struct {
	//input
	tt    PolicyHitMessage_PolicyTriggerType
	qName string
	qType uint16
	attrs []testAttr
	resp  pb.Response

	//output
	act     PolicyHitMessage_PolicyActionType
	polID   string
	rType   uint32
	rData   string
	outAttr map[string]string
}

type testAttr struct {
	id string
	t  string
	v  string
}

func TestSendPolicyHitMsg(t *testing.T) {

	now := time.Now()
	common := commonData{uint64(now.Unix()), uint32(now.Nanosecond()), 12345}

	ttAddress := PolicyHitMessage_POLICY_TRIGGER_ADDRESS
	ttDomain := PolicyHitMessage_POLICY_TRIGGER_DOMAIN

	tests := []testData{
		{
			ttDomain, "test.com", dns.TypeA, []testAttr{
				{"source_ip", "address", "127.0.0.7"},
			}, pb.Response{
				Effect: pb.Response_PERMIT, Reason: "", Obligation: []*pb.Attribute{},
			},
			PolicyHitMessage_POLICY_ACTION_PASSTHROUGH, "", 0, "", map[string]string{
				"source_ip": "127.0.0.7",
			},
		},
		{
			ttAddress, "test.com", dns.TypeAAAA, []testAttr{
				{"customer_id", "hex", "f3dbc96219ee5d44e7e318384088e7d4"},
				{"source_ip", "address", "127.0.0.7"},
				{"address", "address", "69.172.200.235"},
			}, pb.Response{
				Effect: pb.Response_PERMIT, Reason: "", Obligation: []*pb.Attribute{},
			},
			PolicyHitMessage_POLICY_ACTION_PASSTHROUGH, "", 0, "", map[string]string{
				"customer_id": "f3dbc96219ee5d44e7e318384088e7d4",
				"source_ip":   "127.0.0.7",
				"address":     "69.172.200.235",
			},
		},
		{
			ttDomain, "test.com", dns.TypeAAAA, []testAttr{}, pb.Response{
				Effect: pb.Response_PERMIT, Reason: "", Obligation: []*pb.Attribute{
					{"policy_id", "string", "p#123"},
				},
			},
			PolicyHitMessage_POLICY_ACTION_PASSTHROUGH, "p#123", 0, "", map[string]string{},
		},
		{
			ttDomain, "test.com", dns.TypeA, []testAttr{}, pb.Response{
				Effect: pb.Response_PERMIT, Reason: "", Obligation: []*pb.Attribute{
					{"policy_id", "string", "p#123"},
					{"redirect_to", "string", "123.12.0.1"},
				},
			},
			PolicyHitMessage_POLICY_ACTION_PASSTHROUGH, "p#123", 1, "123.12.0.1", map[string]string{},
		},
		{
			ttDomain, "test.com", dns.TypeA, []testAttr{}, pb.Response{
				Effect: pb.Response_DENY, Reason: "", Obligation: []*pb.Attribute{
					{"trigger_category", "string", "spam"},
					{"redirect_to", "string", "fe80::a00:27ff:fe0b:bfde"},
				},
			},
			PolicyHitMessage_POLICY_ACTION_REDIRECT, "", 28, "fe80::a00:27ff:fe0b:bfde", map[string]string{
				"trigger_category": "spam",
			},
		},
		{
			ttDomain, "test.com", dns.TypeA, []testAttr{}, pb.Response{
				Effect: pb.Response_DENY, Reason: "", Obligation: []*pb.Attribute{
					{"refuse", "string", "ok"},
					{"redirect_to", "string", "fe80::a00:27ff:fe0b:bfde"},
				},
			},
			PolicyHitMessage_POLICY_ACTION_REDIRECT, "", 28, "fe80::a00:27ff:fe0b:bfde", map[string]string{},
		},
		{
			ttDomain, "test.com", dns.TypeA, []testAttr{}, pb.Response{
				Effect: pb.Response_DENY, Reason: "", Obligation: []*pb.Attribute{
					{"redirect_to", "string", "test.org"},
				},
			},
			PolicyHitMessage_POLICY_ACTION_REDIRECT, "", 5, "test.org", map[string]string{},
		},
		{
			ttDomain, "test.com", dns.TypeA, []testAttr{}, pb.Response{
				Effect: pb.Response_DENY, Reason: "", Obligation: []*pb.Attribute{
					{"refuse", "string", "ok"},
				},
			},
			PolicyHitMessage_POLICY_ACTION_REFUSE, "", 0, "", map[string]string{},
		},
		{
			ttDomain, "test.com", dns.TypeA, []testAttr{}, pb.Response{
				Effect: pb.Response_NOTAPPLICABLE, Reason: "", Obligation: []*pb.Attribute{},
			},
			PolicyHitMessage_POLICY_ACTION_NXDOMAIN, "", 0, "", map[string]string{},
		},
		{
			ttDomain, "test.com", dns.TypeA, []testAttr{}, pb.Response{
				Effect: pb.Response_DENY, Reason: "", Obligation: []*pb.Attribute{},
			},
			PolicyHitMessage_POLICY_ACTION_NXDOMAIN, "", 0, "", map[string]string{},
		},
	}
	for _, td := range tests {
		tapIO := newIORoutine()
		msg := dns.Msg{}
		msg.SetQuestion(td.qName, td.qType)
		msg.Id = common.qID

		attrs := make([]*DnstapAttribute, 0, len(td.attrs))
		for _, a := range td.attrs {
			attrs = append(attrs, &DnstapAttribute{&a.id, &a.t, &a.v, nil})
		}

		SendPolicyHitMsg(tapIO, now, &msg, td.tt, attrs, &td.resp)
		checkResult(t, tapIO, &common, &td)
	}
}

func checkResult(t *testing.T, tapIO testIORoutine, c *commonData, td *testData) {
	dnstapMsg, ok := <-tapIO.dnstapChan
	if !ok {
		t.Errorf("Receiving Dnstap message was timed out")
		return
	}
	phMsg, err := extractPhm(t, &dnstapMsg)
	if err != nil || phMsg == nil {
		return
	}

	if phMsg.GetTimeSec() != c.sec {
		t.Errorf("Unexpected TimeSec field - expected=%v, actual=%v", c.sec, phMsg.GetTimeSec())
	}
	if phMsg.GetTimeNsec() != c.nsec {
		t.Errorf("Unexpected TimeNsec field - expected=%v, actual=%v", c.nsec, phMsg.GetTimeNsec())
	}
	if phMsg.GetQueryId() != uint32(c.qID) {
		t.Errorf("Unexpected QueryId field - expected=%v, actual=%v", c.qID, phMsg.GetQueryId())
	}
	if phMsg.GetTriggerType() != td.tt {
		t.Errorf("Unexpected TriggerType field - expected=%v, actual=%v", td.tt, phMsg.GetTriggerType())
	}
	if phMsg.GetQueryType() != uint32(td.qType) {
		t.Errorf("Unexpected QueryType field - expected=%v, actual=%v", td.qType, phMsg.GetQueryType())
	}
	if phMsg.GetQueryName() != td.qName {
		t.Errorf("Unexpected QueryName field - expected=%v, actual=%v", td.qName, phMsg.GetQueryName())
	}
	if phMsg.GetPolicyAction() != td.act {
		t.Errorf("Unexpected PolicyAction field - expected=%v, actual=%v", td.act, phMsg.GetPolicyAction())
	}
	if !bytes.Equal(phMsg.GetPolicyId(), []byte(td.polID)) {
		t.Errorf("Unexpected PolicyId field - expected=%v, actual=%v", []byte(td.polID), phMsg.GetPolicyId())
	}
	if phMsg.GetRedirectRrType() != td.rType {
		t.Errorf("Unexpected RedirectRrType field - expected=%v, actual=%v", td.rType, phMsg.GetRedirectRrType())
	}
	if phMsg.GetRedirectRrData() != td.rData {
		t.Errorf("Unexpected RedirectRrData field - expected=%v, actual=%v", td.rType, phMsg.GetRedirectRrData())
	}

	if len(phMsg.Attributes) != len(td.outAttr) {
		t.Errorf("Unexpected count of attributes - expected=%v, actual=%v", len(td.outAttr), len(phMsg.Attributes))
	}
	for _, a := range phMsg.Attributes {
		if exp, ok := td.outAttr[a.GetId()]; !ok {
			t.Errorf("Unexpected attribute - id=%v, value=%v", a.GetId(), a.GetValue())
		} else if a.GetValue() != exp {
			t.Errorf("Unexpected value of attribute %v - expected=%v, actual=%v", a.GetId(), exp, a.GetValue())
		}
	}
}

func extractPhm(t *testing.T, dMsg *dnstap.Dnstap) (*PolicyHitMessage, error) {
	phMsg := &PolicyHitMessage{}
	ext, err := proto.GetExtension(dMsg, E_PolicyHit)
	if err == nil {
		phMsg = ext.(*PolicyHitMessage)
	} else {
		//Dnstap message is not extendable? Checking XXX_unrecognized
		if dMsg.XXX_unrecognized != nil && len(dMsg.XXX_unrecognized) > 0 {
			wrapper := &PolicyHitMessageWrapper{}
			err = proto.Unmarshal(dMsg.XXX_unrecognized, wrapper)
			if err == nil {
				phMsg = wrapper.GetPolicyHit()
			}
		}
	}
	if err != nil {
		t.Errorf("Failed to extract PolicyHit message (%v)", err)
		return nil, err
	}
	return phMsg, nil
}

func TestSendCRExtraMsg(t *testing.T) {
	now := time.Now()
	tapIO := newIORoutine()

	msg := dns.Msg{}
	msg.SetQuestion("test.com.", dns.TypeA)
	msg.Answer = []dns.RR{
		test.A("test.com.		600	IN	A			10.240.0.1"),
	}

	trapper := dtest.TrapTapper{Full: true}
	tapRW := taprw.ResponseWriter{
		Query:          new(dns.Msg),
		ResponseWriter: &test.ResponseWriter{},
		Tapper:         &trapper,
	}
	proxyRW := NewProxyWriter(&tapRW)
	proxyRW.WriteMsg(&msg)

	attrs := []*pb.Attribute{
		{Id: "attr1", Type: "address", Value: "10.240.0.1"},
		{Id: "attr2", Type: "string", Value: "value2"},
	}

	SendCRExtraMsg(tapIO, now, proxyRW, attrs)
	checkCRExtraResult(t, tapIO, now, proxyRW, attrs)

	if l := len(trapper.Trap); l != 0 {
		t.Errorf("Dnstap unexpectedly sent %d messages", l)
		return
	}
}

func checkCRExtraResult(t *testing.T, tapIO testIORoutine, crTime time.Time, proxyRW *ProxyWriter, attrs []*pb.Attribute) {
	dnstapMsg, ok := <-tapIO.dnstapChan
	if !ok {
		t.Errorf("Receiving Dnstap message was timed out")
		return
	}
	extra := &ExtraAttributes{}
	err := proto.Unmarshal(dnstapMsg.Extra, extra)
	if err != nil {
		t.Errorf("Failed to unmarshal Extra (%v)", err)
		return
	}

	checkExtraAttrs(t, extra.GetAttributes(), attrs)
	checkCRMessage(t, dnstapMsg.Message, proxyRW)
}

func checkExtraAttrs(t *testing.T, actual []*DnstapAttribute, expected []*pb.Attribute) {
	if len(actual) != len(expected) {
		t.Errorf("Expected %d attributes, found %d", len(expected), len(actual))
		return
	}

checkAttr:
	for _, ea := range actual {
		for _, a := range expected {
			if ea.GetId() == a.GetId() {
				if ea.GetValue() != a.GetValue() || ea.GetType() != a.GetType() {
					t.Errorf("Attribute %s: expected %v , found %v", ea.GetId(), a, ea)
					return
				}
				continue checkAttr
			}
		}
		t.Errorf("Unexpected attribute found %v", ea)
	}
}

func checkCRMessage(t *testing.T, msg *dnstap.Message, proxyRW *ProxyWriter) {
	if msg == nil {
		t.Errorf("CR message not found")
		return
	}

	d := dtest.TestingData()
	bin, err := proxyRW.msg.Pack()
	if err != nil {
		t.Errorf("Failed to pack message (%v)", err)
		return
	}
	d.Packed = bin
	expMsg := d.ToClientResponse()
	if !dtest.MsgEqual(expMsg, msg) {
		t.Errorf("Unexpected message: expected: %v\nactual: %v", expMsg, msg)
	}
}
