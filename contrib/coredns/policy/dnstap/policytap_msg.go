package policytap

import (
	"log"
	"strings"
	"time"

	tapplg "github.com/coredns/coredns/plugin/dnstap"
	"github.com/coredns/coredns/plugin/dnstap/taprw"
	dnstap "github.com/dnstap/golang-dnstap"
	"github.com/golang/protobuf/proto"
	pb "github.com/infobloxopen/themis/pdp-service"
	"github.com/miekg/dns"
)

// ProxyWriter is designed for intercepting a DNS response being writen to
// ResponseWriter. It also provides access to the base ResponseWriter of type
// "github.com/coredns/coredns/plugin/dnstap/taprw".ResponseWriter
type ProxyWriter struct {
	*taprw.ResponseWriter
	msg *dns.Msg
}

// NewProxyWriter function creates new ProxyWriter from a ResponseWriter derived
// from "github.com/coredns/coredns/plugin/dnstap/taprw".ResponseWriter and
// turns off sending CQ and CR dnstap messages by dnstap plugin
// If ResponseWriter is of other type, NewProxyWriter returns nil
func NewProxyWriter(w dns.ResponseWriter) *ProxyWriter {
	if tapRW, ok := w.(*taprw.ResponseWriter); ok {
		// turn off sending the CQ and CR dnstap messages by dnstap plugin
		tapRW.Send = &taprw.SendOption{}
		return &ProxyWriter{ResponseWriter: tapRW}
	}
	return nil
}

// WriteMsg saves pointer to DNS message and forwards it to base ResponseWriter
func (w *ProxyWriter) WriteMsg(msg *dns.Msg) error {
	w.msg = msg
	return w.ResponseWriter.WriteMsg(msg)
}

// SendCRExtraMsg creates Client Response (CR) dnstap Message and writes an array
// of extra attributes to Dnstap.Extra field. Then it asynchronously sends the
// message with IORoutine interface
// Parameter tapIO must not be nil
func SendCRExtraMsg(tapIO tapplg.IORoutine, t time.Time, pw *ProxyWriter, attrs []*pb.Attribute) {
	go func() {
		if pw == nil || pw.msg == nil {
			log.Printf("[ERROR] Failed to create dnstap CR message - no DNS response message found")
			return
		}
		b := pw.TapBuilder()
		b.TimeSec = uint64(t.Unix())
		timeNs := uint32(t.Nanosecond())
		err := b.AddrMsg(pw.RemoteAddr(), pw.msg)
		if err != nil {
			log.Printf("[ERROR] Failed to create dnstap CR message (%v)", err)
			return
		}
		crMsg := b.ToClientResponse()
		crMsg.ResponseTimeNsec = &timeNs
		t := dnstap.Dnstap_MESSAGE
		extra, err := proto.Marshal(&ExtraAttributes{Attributes: ConvertAttrs(attrs)})
		if err != nil {
			log.Printf("[ERROR] Failed to create extra data for dnstap CR message (%v)", err)
			return
		}
		dnstapMsg := dnstap.Dnstap{Type: &t, Message: crMsg, Extra: extra}
		tapIO.Dnstap(dnstapMsg)
	}()
}

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
