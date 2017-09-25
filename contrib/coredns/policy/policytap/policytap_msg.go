package policytap

import (
	"io"
	"log"
	"strings"
	"time"

	dnstap "github.com/dnstap/golang-dnstap"
	"github.com/golang/protobuf/proto"
	pb "github.com/infobloxopen/themis/pdp-service"
	"github.com/miekg/dns"
)

// ToDnstap creates PolicyHitMessage and writes it to the privided Writer
func ToDnstap(w io.Writer, t time.Time, msg *dns.Msg, tt PolicyHitMessage_PolicyTriggerType,
	attrs []*DnstapAttribute, r *pb.Response) error {

	if w == nil {
		return nil
	}

	phm := newPolicyHitMessage(t)
	phm.TriggerType = &tt
	phm.updateFromMessage(msg)
	phm.AddDnstapAttrs(attrs)
	phm.updateFromResponse(r)

	return writeMessage(w, phm)
}

func newPolicyHitMessage(t time.Time) *PolicyHitMessage {
	sec := uint64(t.Unix())
	nsec := uint32(t.Nanosecond())
	return &PolicyHitMessage{TimeSec: &sec, TimeNsec: &nsec}
}

func (phm *PolicyHitMessage) updateFromMessage(msg *dns.Msg) {
	if msg != nil {
		qId := uint32(msg.Id)
		phm.QueryId = &qId
		if len(msg.Question) > 0 {
			q := msg.Question[0]
			qType := uint32(q.Qtype)
			qName := strings.TrimRight(q.Name, ".")
			phm.QueryType = &qType
			phm.QueryName = &qName
		}
	}
}

func (phm *PolicyHitMessage) updateFromResponse(r *pb.Response) {
	act := PolicyHitMessage_POLICY_ACTION_NXDOMAIN
	switch r.Effect {
	case pb.Response_PERMIT:
		act = PolicyHitMessage_POLICY_ACTION_PASSTHROUGH
	}
	phm.PolicyAction = &act
	phm.AddPdpAttrs(r.Obligation)
}

func writeMessage(w io.Writer, phm *PolicyHitMessage) error {
	dnstapType := dnstap.Dnstap_MESSAGE
	dnstapMsg := dnstap.Dnstap{Type: &dnstapType}

	err := proto.SetExtension(&dnstapMsg, E_PolicyHit, phm)
	if err != nil {
		// likely linked with not extendable Dnstap, adding message to XXX_unrecognized
		wraper := PolicyHitMessageWraper{PolicyHit: phm}
		rawPhm, err1 := proto.Marshal(&wraper)
		if err1 != nil {
			return err
		}
		dnstapMsg.XXX_unrecognized = rawPhm
	}

	rawMsg, err := proto.Marshal(&dnstapMsg)
	if err != nil {
		log.Printf("[ERROR] %v", err)
		return err
	}
	_, err = w.Write(rawMsg)
	return err
}
