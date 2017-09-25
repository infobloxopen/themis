package policytap

import (
	"bytes"
	"testing"
	"time"

	dnstap "github.com/dnstap/golang-dnstap"
	"github.com/golang/protobuf/proto"
	pb "github.com/infobloxopen/themis/pdp-service"
	"github.com/miekg/dns"
)

type testWriter struct {
	data *[]byte
}

func (w testWriter) Write(data []byte) (int, error) {
	*w.data = data
	return len(data), nil
}

type commonData struct {
	sec  uint64
	nsec uint32
	qId  uint16
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
	pId     string
	rType   uint32
	rData   string
	outAttr map[string]string
}

type testAttr struct {
	id string
	t  string
	v  string
}

func TestToDnstap(t *testing.T) {

	now := time.Now()
	common := commonData{uint64(now.Unix()), uint32(now.Nanosecond()), 12345}

	ttAddress := PolicyHitMessage_POLICY_TRIGGER_ADDRESS
	ttDomain := PolicyHitMessage_POLICY_TRIGGER_DOMAIN

	tests := []testData{
		{
			ttDomain, "test.com", dns.TypeA, []testAttr{
				{"source_ip", "address", "127.0.0.7"},
			}, pb.Response{
				pb.Response_PERMIT, "", []*pb.Attribute{},
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
				pb.Response_PERMIT, "", []*pb.Attribute{},
			},
			PolicyHitMessage_POLICY_ACTION_PASSTHROUGH, "", 0, "", map[string]string{
				"customer_id": "f3dbc96219ee5d44e7e318384088e7d4",
				"source_ip":   "127.0.0.7",
				"address":     "69.172.200.235",
			},
		},
		{
			ttDomain, "test.com", dns.TypeAAAA, []testAttr{}, pb.Response{
				pb.Response_PERMIT, "", []*pb.Attribute{
					{"policy_id", "string", "p#123"},
				},
			},
			PolicyHitMessage_POLICY_ACTION_PASSTHROUGH, "p#123", 0, "", map[string]string{},
		},
		{
			ttDomain, "test.com", dns.TypeA, []testAttr{}, pb.Response{
				pb.Response_PERMIT, "", []*pb.Attribute{
					{"policy_id", "string", "p#123"},
					{"redirect_to", "string", "123.12.0.1"},
				},
			},
			PolicyHitMessage_POLICY_ACTION_PASSTHROUGH, "p#123", 1, "123.12.0.1", map[string]string{},
		},
		{
			ttDomain, "test.com", dns.TypeA, []testAttr{}, pb.Response{
				pb.Response_DENY, "", []*pb.Attribute{
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
				pb.Response_DENY, "", []*pb.Attribute{
					{"refuse", "string", "ok"},
					{"redirect_to", "string", "fe80::a00:27ff:fe0b:bfde"},
				},
			},
			PolicyHitMessage_POLICY_ACTION_REDIRECT, "", 28, "fe80::a00:27ff:fe0b:bfde", map[string]string{},
		},
		{
			ttDomain, "test.com", dns.TypeA, []testAttr{}, pb.Response{
				pb.Response_DENY, "", []*pb.Attribute{
					{"redirect_to", "string", "test.org"},
				},
			},
			PolicyHitMessage_POLICY_ACTION_REDIRECT, "", 5, "test.org", map[string]string{},
		},
		{
			ttDomain, "test.com", dns.TypeA, []testAttr{}, pb.Response{
				pb.Response_DENY, "", []*pb.Attribute{
					{"refuse", "string", "ok"},
				},
			},
			PolicyHitMessage_POLICY_ACTION_REFUSE, "", 0, "", map[string]string{},
		},
		{
			ttDomain, "test.com", dns.TypeA, []testAttr{}, pb.Response{
				pb.Response_NOTAPPLICABLE, "", []*pb.Attribute{},
			},
			PolicyHitMessage_POLICY_ACTION_NXDOMAIN, "", 0, "", map[string]string{},
		},
		{
			ttDomain, "test.com", dns.TypeA, []testAttr{}, pb.Response{
				pb.Response_DENY, "", []*pb.Attribute{},
			},
			PolicyHitMessage_POLICY_ACTION_NXDOMAIN, "", 0, "", map[string]string{},
		},
	}
	for _, td := range tests {
		rawMsg := []byte{}
		w := testWriter{&rawMsg}
		msg := dns.Msg{}
		msg.SetQuestion(td.qName, td.qType)
		msg.Id = common.qId

		attrs := make([]*DnstapAttribute, 0, len(td.attrs))
		for _, a := range td.attrs {
			attrs = append(attrs, &DnstapAttribute{&a.id, &a.t, &a.v, nil})
		}

		err := ToDnstap(w, now, &msg, td.tt, attrs, &td.resp)
		checkResult(t, err, w, &common, &td)
	}
}

func checkResult(t *testing.T, err error, tw testWriter, c *commonData, td *testData) {
	if err != nil {
		t.Errorf("ToDnstap returned error (%v)", err)
	} else {
		phMsg, err := unmarshalPhm(t, *tw.data)
		if err != nil || phMsg == nil {
			return
		}

		if phMsg.GetTimeSec() != c.sec {
			t.Errorf("Unexpected TimeSec field - expected=%v, actual=%v", c.sec, phMsg.GetTimeSec())
		}
		if phMsg.GetTimeNsec() != c.nsec {
			t.Errorf("Unexpected TimeNsec field - expected=%v, actual=%v", c.nsec, phMsg.GetTimeNsec())
		}
		if phMsg.GetQueryId() != uint32(c.qId) {
			t.Errorf("Unexpected QueryId field - expected=%v, actual=%v", c.qId, phMsg.GetQueryId())
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
		if !bytes.Equal(phMsg.GetPolicyId(), []byte(td.pId)) {
			t.Errorf("Unexpected PolicyId field - expected=%v, actual=%v", []byte(td.pId), phMsg.GetPolicyId())
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
}

func unmarshalPhm(t *testing.T, data []byte) (*PolicyHitMessage, error) {
	dMsg := &dnstap.Dnstap{}
	err := proto.Unmarshal(data, dMsg)
	if err != nil {
		t.Errorf("Failed to unmarshal Dnstap message (%v); data=[%x]", err, data)
		return nil, err
	}
	phMsg := &PolicyHitMessage{}
	ext, err := proto.GetExtension(dMsg, E_PolicyHit)
	if err == nil {
		phMsg = ext.(*PolicyHitMessage)
	} else {
		//Dnstap message is not extendable? Checking XXX_unrecognized
		if dMsg.XXX_unrecognized != nil && len(dMsg.XXX_unrecognized) > 0 {
			wraper := &PolicyHitMessageWraper{}
			err = proto.Unmarshal(dMsg.XXX_unrecognized, wraper)
			if err == nil {
				phMsg = wraper.GetPolicyHit()
			}
		}
	}
	if err != nil {
		t.Errorf("Failed to unmarshal PolicyHit message (%v)", err)
		return nil, err
	}
	return phMsg, nil
}
