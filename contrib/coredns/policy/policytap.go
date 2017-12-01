package policy

import (
	"log"
	"runtime"
	"time"

	"github.com/coredns/coredns/plugin/dnstap"
	"github.com/coredns/coredns/plugin/dnstap/taprw"
	tap "github.com/dnstap/golang-dnstap"
	"github.com/golang/protobuf/proto"
	pb "github.com/infobloxopen/themis/contrib/coredns/policy/dnstap"
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

type DnstapSender interface {
	SendCRExtraMsg(pw *ProxyWriter, ah *attrHolder)
	Stop()
}

type dnstapData struct {
	pw *ProxyWriter
	ah *attrHolder
}

type policyDnstapSender struct {
	ior        dnstap.IORoutine
	workerChan chan dnstapData
}

func NewPolicyDnstapSender(io dnstap.IORoutine) DnstapSender {
	ds := &policyDnstapSender{ior: io, workerChan: make(chan dnstapData, workerQSize)}
	workerCnt := 2 * runtime.NumCPU()
	for i := 0; i < workerCnt; i++ {
		go ds.tapWorker()
	}
	return ds
}

// SendCRExtraMsg creates Client Response (CR) dnstap Message and writes an array
// of extra attributes to Dnstap.Extra field. Then it asynchronously sends the
// message with IORoutine interface
// Parameter tapIO must not be nil
func (s *policyDnstapSender) SendCRExtraMsg(pw *ProxyWriter, ah *attrHolder) {
	if pw == nil || pw.msg == nil {
		log.Printf("[ERROR] Failed to create dnstap CR message - no DNS response message found")
		return
	}
	s.workerChan <- dnstapData{pw: pw, ah: ah}
}

const (
	workerQSize = 10000
	defAttrCnt  = 10
)

type attrBlock struct {
	dAttrs []pb.DnstapAttribute
	pAttrs []*pb.DnstapAttribute
}

func allocateAttrs(a *attrBlock, cnt int) {
	a.dAttrs = make([]pb.DnstapAttribute, cnt)
	a.pAttrs = make([]*pb.DnstapAttribute, cnt)
	for i := range a.dAttrs {
		a.pAttrs[i] = &(a.dAttrs[i])
	}
}

func (s *policyDnstapSender) tapWorker() {
	aBlock := &attrBlock{}
	allocateAttrs(aBlock, defAttrCnt)

	extraMsg := pb.Extra{}
	t := tap.Dnstap_MESSAGE

	for data := range s.workerChan {
		now := time.Now()
		b := data.pw.TapBuilder()
		b.TimeSec = uint64(now.Unix())
		timeNs := uint32(now.Nanosecond())
		err := b.AddrMsg(data.pw.RemoteAddr(), data.pw.msg)
		if err != nil {
			log.Printf("[ERROR] Failed to create dnstap CR message (%v)", err)
			continue
		}
		crMsg := b.ToClientResponse()
		crMsg.ResponseTimeNsec = &timeNs

		var extra []byte
		if data.ah != nil {
			extraMsg.Attrs = convertAttrs(data.ah, aBlock)
			extra, err = proto.Marshal(&extraMsg)
			if err != nil {
				log.Printf("[ERROR] Failed to create extra data for dnstap CR message (%v)", err)
			}
		}
		dnstapMsg := tap.Dnstap{Type: &t, Message: crMsg, Extra: extra}
		s.ior.Dnstap(dnstapMsg)
	}
}

func convertAttrs(ah *attrHolder, a *attrBlock) []*pb.DnstapAttribute {
	length := len(ah.attrsReqDomain) + len(ah.attrsRespDomain) + len(ah.attrsRespRespip) + 2
	if cap(a.dAttrs) < length {
		allocateAttrs(a, length)
	}

	i := 0
	for _, attr := range ah.attrsReqDomain[1:] {
		a.dAttrs[i].Id = attr.Id
		a.dAttrs[i].Value = attr.Value
		i++
	}
	for _, attr := range ah.attrsRespDomain {
		a.dAttrs[i].Id = attr.Id
		a.dAttrs[i].Value = attr.Value
		i++
	}
	if len(ah.address) > 0 {
		a.dAttrs[i].Id = AttrNameAddress
		a.dAttrs[i].Value = ah.address
		i++
	}
	for _, attr := range ah.attrsRespRespip {
		a.dAttrs[i].Id = attr.Id
		a.dAttrs[i].Value = attr.Value
		i++
	}
	a.dAttrs[i].Id = AttrNamePolicyAction
	a.dAttrs[i].Value = actionConvDnstap[ah.action]
	i++
	a.dAttrs[i].Id = AttrNameType
	if len(ah.attrsReqRespip) > 0 {
		a.dAttrs[i].Value = TypeValueResponse
	} else {
		a.dAttrs[i].Value = TypeValueQuery
	}
	i++
	return a.pAttrs[:i]
}

func (s *policyDnstapSender) Stop() {
    close(s.workerChan)
}
