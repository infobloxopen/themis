package policytap

import (
	"log"
	"strings"
	"time"

	tapplg "github.com/coredns/coredns/plugin/dnstap"
	dnstap "github.com/dnstap/golang-dnstap"
	"github.com/golang/protobuf/proto"
	pb "github.com/infobloxopen/themis/pdp-service"
	"github.com/miekg/dns"
)

// SendPolicyHitMsg creates PolicyHitMessage and asynchronously sends it to the provided IORoutine
// Parameter tapIO must not be nil
func SendPolicyHitMsg(tapIO tapplg.IORoutine, t time.Time, msg *dns.Msg, tt PolicyHitMessage_PolicyTriggerType,
	attrs []*DnstapAttribute, r *pb.Response) {

	//write message asynchronously
	go func() {
		phm := newPolicyHitMessage(t)
		phm.TriggerType = &tt
		phm.updateFromMessage(msg)
		phm.AddDnstapAttrs(attrs)
		phm.updateFromResponse(r)

		writeMessage(tapIO, phm)
	}()
}

func newPolicyHitMessage(t time.Time) *PolicyHitMessage {
	sec := uint64(t.Unix())
	nsec := uint32(t.Nanosecond())
	return &PolicyHitMessage{TimeSec: &sec, TimeNsec: &nsec}
}

func (phm *PolicyHitMessage) updateFromMessage(msg *dns.Msg) {
	if msg != nil {
		qID := uint32(msg.Id)
		phm.QueryId = &qID
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

func writeMessage(tapIO tapplg.IORoutine, phm *PolicyHitMessage) {
	dnstapType := dnstap.Dnstap_MESSAGE
	dnstapMsg := dnstap.Dnstap{Type: &dnstapType}

	err := proto.SetExtension(&dnstapMsg, E_PolicyHit, phm)
	if err != nil {
		// likely linked with not extendable Dnstap, adding message to XXX_unrecognized
		wrapper := PolicyHitMessageWrapper{PolicyHit: phm}
		rawPhm, err1 := proto.Marshal(&wrapper)
		if err1 == nil {
			dnstapMsg.XXX_unrecognized = rawPhm
			err = nil
		}
	}

	if err == nil {
		tapIO.Dnstap(dnstapMsg)
	} else {
		log.Printf("[ERROR] Failed to pack PolicyHit message (%s)", err)
	}
}
