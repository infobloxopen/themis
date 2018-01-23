package policy

import (
	"log"
	"time"

	"github.com/coredns/coredns/plugin/dnstap"
	"github.com/coredns/coredns/plugin/dnstap/taprw"
	tap "github.com/dnstap/golang-dnstap"
	"github.com/golang/protobuf/proto"
	pb "github.com/infobloxopen/themis/contrib/coredns/policy/dnstap"
	"github.com/miekg/dns"
)

// proxyWriter is designed for intercepting a DNS response being writen to
// ResponseWriter. It also provides access to the base ResponseWriter of type
// "github.com/coredns/coredns/plugin/dnstap/taprw".ResponseWriter
type proxyWriter struct {
	*taprw.ResponseWriter
	msg *dns.Msg
}

// newProxyWriter function creates new proxyWriter from a ResponseWriter derived
// from "github.com/coredns/coredns/plugin/dnstap/taprw".ResponseWriter and
// turns off sending CQ and CR dnstap messages by dnstap plugin
// If ResponseWriter is of other type, newProxyWriter returns nil
func newProxyWriter(w dns.ResponseWriter) *proxyWriter {
	if tapRW, ok := w.(*taprw.ResponseWriter); ok {
		// turn off sending the CQ and CR dnstap messages by dnstap plugin
		tapRW.Send = &taprw.SendOption{}
		return &proxyWriter{ResponseWriter: tapRW}
	}
	return nil
}

// WriteMsg saves pointer to DNS message and forwards it to base ResponseWriter
func (w *proxyWriter) WriteMsg(msg *dns.Msg) error {
	w.msg = msg
	return w.ResponseWriter.WriteMsg(msg)
}

type dnstapSender interface {
	sendCRExtraMsg(pw *proxyWriter, ah *attrHolder)
}

type policyDnstapSender struct {
	ior dnstap.IORoutine
}

func newPolicyDnstapSender(io dnstap.IORoutine) dnstapSender {
	return &policyDnstapSender{ior: io}
}

// sendCRExtraMsg creates Client Response (CR) dnstap Message and writes an array
// of extra attributes to Dnstap.Extra field. Then it asynchronously sends the
// message with IORoutine interface
// Parameter tapIO must not be nil
func (s *policyDnstapSender) sendCRExtraMsg(pw *proxyWriter, ah *attrHolder) {
	if pw == nil || pw.msg == nil {
		log.Printf("[ERROR] Failed to create dnstap CR message - no DNS response message found")
		return
	}
	now := time.Now()
	b := pw.TapBuilder()
	b.TimeSec = uint64(now.Unix())
	timeNs := uint32(now.Nanosecond())
	err := b.AddrMsg(pw.RemoteAddr(), pw.msg)
	if err != nil {
		log.Printf("[ERROR] Failed to create dnstap CR message (%v)", err)
		return
	}
	crMsg := b.ToClientResponse()
	crMsg.ResponseTimeNsec = &timeNs
	t := tap.Dnstap_MESSAGE

	var extra []byte
	if ah != nil {
		extra, err = proto.Marshal(&pb.Extra{Attrs: ah.convertAttrs()})
		if err != nil {
			log.Printf("[ERROR] Failed to create extra data for dnstap CR message (%v)", err)
		}
	}
	dnstapMsg := tap.Dnstap{Type: &t, Message: crMsg, Extra: extra}
	s.ior.Dnstap(dnstapMsg)
}
