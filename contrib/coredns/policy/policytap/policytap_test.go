package policytap

import (
	"bytes"
	"testing"
	"time"

	dnstap "github.com/dnstap/golang-dnstap"
	"github.com/golang/protobuf/proto"
	pdp "github.com/infobloxopen/themis/pdp-service"
	"github.com/miekg/dns"
)

type testIORoutine struct {
	dnstapChan chan dnstap.Dnstap
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
	resp  pdp.Response

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
			}, pdp.Response{
				Effect: pdp.PERMIT, Reason: "", Obligations: []*pdp.Attribute{},
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
			}, pdp.Response{
				Effect: pdp.PERMIT, Reason: "", Obligations: []*pdp.Attribute{},
			},
			PolicyHitMessage_POLICY_ACTION_PASSTHROUGH, "", 0, "", map[string]string{
				"customer_id": "f3dbc96219ee5d44e7e318384088e7d4",
				"source_ip":   "127.0.0.7",
				"address":     "69.172.200.235",
			},
		},
		{
			ttDomain, "test.com", dns.TypeAAAA, []testAttr{}, pdp.Response{
				Effect: pdp.PERMIT, Reason: "", Obligations: []*pdp.Attribute{
					{"policy_id", "string", "p#123"},
				},
			},
			PolicyHitMessage_POLICY_ACTION_PASSTHROUGH, "p#123", 0, "", map[string]string{},
		},
		{
			ttDomain, "test.com", dns.TypeA, []testAttr{}, pdp.Response{
				Effect: pdp.PERMIT, Reason: "", Obligations: []*pdp.Attribute{
					{"policy_id", "string", "p#123"},
					{"redirect_to", "string", "123.12.0.1"},
				},
			},
			PolicyHitMessage_POLICY_ACTION_PASSTHROUGH, "p#123", 1, "123.12.0.1", map[string]string{},
		},
		{
			ttDomain, "test.com", dns.TypeA, []testAttr{}, pdp.Response{
				Effect: pdp.DENY, Reason: "", Obligations: []*pdp.Attribute{
					{"trigger_category", "string", "spam"},
					{"redirect_to", "string", "fe80::a00:27ff:fe0b:bfde"},
				},
			},
			PolicyHitMessage_POLICY_ACTION_REDIRECT, "", 28, "fe80::a00:27ff:fe0b:bfde", map[string]string{
				"trigger_category": "spam",
			},
		},
		{
			ttDomain, "test.com", dns.TypeA, []testAttr{}, pdp.Response{
				Effect: pdp.DENY, Reason: "", Obligations: []*pdp.Attribute{
					{"refuse", "string", "ok"},
					{"redirect_to", "string", "fe80::a00:27ff:fe0b:bfde"},
				},
			},
			PolicyHitMessage_POLICY_ACTION_REDIRECT, "", 28, "fe80::a00:27ff:fe0b:bfde", map[string]string{},
		},
		{
			ttDomain, "test.com", dns.TypeA, []testAttr{}, pdp.Response{
				Effect: pdp.DENY, Reason: "", Obligations: []*pdp.Attribute{
					{"redirect_to", "string", "test.org"},
				},
			},
			PolicyHitMessage_POLICY_ACTION_REDIRECT, "", 5, "test.org", map[string]string{},
		},
		{
			ttDomain, "test.com", dns.TypeA, []testAttr{}, pdp.Response{
				Effect: pdp.DENY, Reason: "", Obligations: []*pdp.Attribute{
					{"refuse", "string", "ok"},
				},
			},
			PolicyHitMessage_POLICY_ACTION_REFUSE, "", 0, "", map[string]string{},
		},
		{
			ttDomain, "test.com", dns.TypeA, []testAttr{}, pdp.Response{
				Effect: pdp.NOTAPPLICABLE, Reason: "", Obligations: []*pdp.Attribute{},
			},
			PolicyHitMessage_POLICY_ACTION_NXDOMAIN, "", 0, "", map[string]string{},
		},
		{
			ttDomain, "test.com", dns.TypeA, []testAttr{}, pdp.Response{
				Effect: pdp.DENY, Reason: "", Obligations: []*pdp.Attribute{},
			},
			PolicyHitMessage_POLICY_ACTION_NXDOMAIN, "", 0, "", map[string]string{},
		},
	}
	for _, td := range tests {
		dnstapChan := make(chan dnstap.Dnstap, 1)
		tapIO := testIORoutine{dnstapChan: dnstapChan}
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
	dnstapMsg := <-tapIO.dnstapChan
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
