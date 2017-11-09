package policy

import (
	"encoding/hex"
	"errors"
	"reflect"
	"testing"

	"github.com/coredns/coredns/plugin"
	"github.com/coredns/coredns/plugin/dnstap/taprw"
	dtest "github.com/coredns/coredns/plugin/dnstap/test"
	"github.com/coredns/coredns/plugin/pkg/dnstest"
	"github.com/coredns/coredns/plugin/test"
	ot "github.com/opentracing/opentracing-go"

	pb "github.com/infobloxopen/themis/contrib/coredns/policy/dnstap"
	pdp "github.com/infobloxopen/themis/pdp-service"

	"github.com/miekg/dns"
	"golang.org/x/net/context"
)

var (
	errFakePdp      = errors.New("fake PDP error")
	errFakeResolver = errors.New("fake Resolver error")
)

func TestPolicy(t *testing.T) {
	p := newPolicyPlugin()
	p.next = handler()

	tests := []struct {
		query      string
		queryType  uint16
		response   *pdp.Response
		responseIP *pdp.Response
		errResp    error
		errRespIP  error
		status     int
		err        error
		attrs      []*pdp.Attribute
	}{
		{
			query:     "test.com.",
			queryType: dns.TypeA,
			errResp:   errFakePdp,
			status:    dns.RcodeServerFailure,
			err:       errFakePdp,
		},
		{
			query:     "test.com.",
			queryType: dns.TypeA,
			response:  &pdp.Response{Effect: pdp.INDETERMINATE},
			status:    dns.RcodeServerFailure,
			err:       errInvalidAction,
		},
		{
			query:      "test.com.",
			queryType:  dns.TypeA,
			response:   &pdp.Response{Effect: pdp.PERMIT},
			responseIP: &pdp.Response{Effect: pdp.INDETERMINATE},
			status:     dns.RcodeServerFailure,
			err:        errInvalidAction,
		},
		{
			query:     "test.com.",
			queryType: dns.TypeA,
			response:  &pdp.Response{Effect: pdp.PERMIT},
			errRespIP: errFakePdp,
			status:    dns.RcodeServerFailure,
			err:       errFakePdp,
		},
		{
			query:     "test.com.",
			queryType: dns.TypeA,
			response:  &pdp.Response{Effect: pdp.DENY},
			status:    dns.RcodeNameError,
			err:       nil,
			attrs: []*pdp.Attribute{
				{Id: AttrNameDomainName, Value: "test.com"},
				{Id: AttrNameDNSQtype, Value: "1"},
				{Id: AttrNameSourceIP, Value: "10.240.0.1"},
				{Id: AttrNamePolicyAction, Value: "3"},
				{Id: AttrNameType, Value: TypeValueQuery},
			},
		},
		{
			query:      "test.com.",
			queryType:  dns.TypeA,
			response:   &pdp.Response{Effect: pdp.PERMIT},
			responseIP: &pdp.Response{Effect: pdp.PERMIT},
			status:     dns.RcodeSuccess,
			err:        nil,
		},
		{
			query:      "test.com.",
			queryType:  dns.TypeA,
			response:   &pdp.Response{Effect: pdp.PERMIT},
			responseIP: &pdp.Response{Effect: pdp.DENY},
			status:     dns.RcodeNameError,
			err:        nil,
			attrs: []*pdp.Attribute{
				{Id: AttrNameDomainName, Value: "test.com"},
				{Id: AttrNameDNSQtype, Value: "1"},
				{Id: AttrNameSourceIP, Value: "10.240.0.1"},
				{Id: AttrNameAddress, Value: "10.240.0.1"},
				{Id: AttrNamePolicyAction, Value: "3"},
				{Id: AttrNameType, Value: TypeValueResponse},
			},
		},
		{
			query:     "test.com.",
			queryType: dns.TypeA,
			response: &pdp.Response{Effect: pdp.DENY,
				Obligations: []*pdp.Attribute{{Id: AttrNameRedirectTo, Value: "221.228.88.194"}}},
			status: dns.RcodeSuccess,
			err:    nil,
			attrs: []*pdp.Attribute{
				{Id: AttrNameDomainName, Value: "test.com"},
				{Id: AttrNameDNSQtype, Value: "1"},
				{Id: AttrNameSourceIP, Value: "10.240.0.1"},
				{Id: AttrNameRedirectTo, Value: "221.228.88.194"},
				{Id: AttrNamePolicyAction, Value: "4"},
				{Id: AttrNameType, Value: TypeValueQuery},
			},
		},
		{
			query:     "test.com.",
			queryType: dns.TypeA,
			response: &pdp.Response{Effect: pdp.DENY,
				Obligations: []*pdp.Attribute{{Id: AttrNameRedirectTo, Value: "redirect.biz"}}},
			status: dns.RcodeSuccess,
			err:    nil,
			attrs: []*pdp.Attribute{
				{Id: AttrNameDomainName, Value: "test.com"},
				{Id: AttrNameDNSQtype, Value: "1"},
				{Id: AttrNameSourceIP, Value: "10.240.0.1"},
				{Id: AttrNameRedirectTo, Value: "redirect.biz"},
				{Id: AttrNamePolicyAction, Value: "4"},
				{Id: AttrNameType, Value: TypeValueQuery},
			},
		},
		{
			query:     "test.com.",
			queryType: dns.TypeA,
			response:  &pdp.Response{Effect: pdp.PERMIT},
			responseIP: &pdp.Response{Effect: pdp.DENY,
				Obligations: []*pdp.Attribute{{Id: AttrNameRedirectTo, Value: "221.228.88.194"}}},
			status: dns.RcodeSuccess,
			err:    nil,
			attrs: []*pdp.Attribute{
				{Id: AttrNameDomainName, Value: "test.com"},
				{Id: AttrNameDNSQtype, Value: "1"},
				{Id: AttrNameSourceIP, Value: "10.240.0.1"},
				{Id: AttrNameAddress, Value: "10.240.0.1"},
				{Id: AttrNameRedirectTo, Value: "221.228.88.194"},
				{Id: AttrNamePolicyAction, Value: "4"},
				{Id: AttrNameType, Value: TypeValueResponse},
			},
		},
		{
			query:     "test.com.",
			queryType: dns.TypeA,
			response:  &pdp.Response{Effect: pdp.PERMIT},
			responseIP: &pdp.Response{Effect: pdp.DENY,
				Obligations: []*pdp.Attribute{{Id: AttrNameRedirectTo, Value: "redirect.biz"}}},
			status: dns.RcodeSuccess,
			err:    nil,
			attrs: []*pdp.Attribute{
				{Id: AttrNameDomainName, Value: "test.com"},
				{Id: AttrNameDNSQtype, Value: "1"},
				{Id: AttrNameSourceIP, Value: "10.240.0.1"},
				{Id: AttrNameAddress, Value: "10.240.0.1"},
				{Id: AttrNameRedirectTo, Value: "redirect.biz"},
				{Id: AttrNamePolicyAction, Value: "4"},
				{Id: AttrNameType, Value: TypeValueResponse},
			},
		},
		{
			query:     "test.org.",
			queryType: dns.TypeAAAA,
			response:  &pdp.Response{Effect: pdp.PERMIT},
			responseIP: &pdp.Response{Effect: pdp.DENY,
				Obligations: []*pdp.Attribute{{Id: AttrNameRedirectTo, Value: "2001:db8:0:200:0:0:0:7"}}},
			status: dns.RcodeSuccess,
			err:    nil,
			attrs: []*pdp.Attribute{
				{Id: AttrNameDomainName, Value: "test.org"},
				{Id: AttrNameDNSQtype, Value: "1c"},
				{Id: AttrNameSourceIP, Value: "10.240.0.1"},
				{Id: AttrNameAddress, Value: "21da:d3:0:2f3b:2aa:ff:fe28:9c5a"},
				{Id: AttrNameRedirectTo, Value: "2001:db8:0:200:0:0:0:7"},
				{Id: AttrNamePolicyAction, Value: "4"},
				{Id: AttrNameType, Value: TypeValueResponse},
			},
		},
		{
			query:     "test.com.",
			queryType: dns.TypeA,
			response: &pdp.Response{Effect: pdp.DENY,
				Obligations: []*pdp.Attribute{{Id: AttrNameRedirectTo, Value: "test.net"}}},
			status: dns.RcodeServerFailure,
			err:    errFakeResolver,
		},
		{
			query:     "test.net.",
			queryType: dns.TypeA,
			response:  &pdp.Response{Effect: pdp.PERMIT},
			status:    dns.RcodeServerFailure,
			err:       errFakeResolver,
		},
		{
			query:     "test.org.",
			queryType: dns.TypeAAAA,
			response:  &pdp.Response{Effect: pdp.DENY},
			status:    dns.RcodeNameError,
			err:       nil,
			attrs: []*pdp.Attribute{
				{Id: AttrNameDomainName, Value: "test.org"},
				{Id: AttrNameDNSQtype, Value: "1c"},
				{Id: AttrNameSourceIP, Value: "10.240.0.1"},
				{Id: AttrNamePolicyAction, Value: "3"},
				{Id: AttrNameType, Value: TypeValueQuery},
			},
		},
		{
			query:     "test.org.",
			queryType: dns.TypeAAAA,
			response:  &pdp.Response{Effect: pdp.PERMIT},
			status:    dns.RcodeSuccess,
			err:       nil,
		},
		{
			query:     "test.org.",
			queryType: dns.TypeAAAA,
			response: &pdp.Response{Effect: pdp.DENY,
				Obligations: []*pdp.Attribute{{Id: AttrNameRedirectTo, Value: "redirect.net"}}},
			status: dns.RcodeSuccess,
			err:    nil,
			attrs: []*pdp.Attribute{
				{Id: AttrNameDomainName, Value: "test.org"},
				{Id: AttrNameDNSQtype, Value: "1c"},
				{Id: AttrNameSourceIP, Value: "10.240.0.1"},
				{Id: AttrNamePolicyAction, Value: "4"},
				{Id: AttrNameRedirectTo, Value: "redirect.net"},
				{Id: AttrNameType, Value: TypeValueQuery},
			},
		},
		{
			query:     "test.org.",
			queryType: dns.TypeA,
			response: &pdp.Response{Effect: pdp.DENY,
				Obligations: []*pdp.Attribute{{Id: "refuse", Value: "true"}}},
			status: dns.RcodeRefused,
			err:    nil,
		},
		{
			query:     "test.com.",
			queryType: dns.TypeA,
			response:  &pdp.Response{Effect: pdp.PERMIT},
			responseIP: &pdp.Response{Effect: pdp.DENY,
				Obligations: []*pdp.Attribute{{Id: "refuse", Value: "true"}}},
			status: dns.RcodeRefused,
			err:    nil,
		},
		{
			query:     "nxdomain.org.",
			queryType: dns.TypeA,
			response:  &pdp.Response{Effect: pdp.PERMIT},
			status:    dns.RcodeNameError,
			err:       nil,
		},
	}

	for _, withDnstap := range [2]bool{false, true} {
		var rec dns.ResponseWriter
		sender := &testDnstapSender{}
		if !withDnstap {
			rec = dnstest.NewRecorder(&test.ResponseWriter{})
		} else {
			rec = &taprw.ResponseWriter{
				Query:          new(dns.Msg),
				ResponseWriter: &test.ResponseWriter{},
				Tapper:         &dtest.TrapTapper{Full: true},
			}
			p.tapIO = sender
		}

		// no debug
		for i, test := range tests {
			sender.reset()
			// Make request
			req := new(dns.Msg)
			req.SetQuestion(test.query, test.queryType)
			// Init test mock client
			p.pdp = newTestClientInit(test.response, test.responseIP, test.errResp, test.errRespIP)
			// Handle request
			status, err := p.ServeDNS(context.Background(), rec, req)
			// Check status
			if test.status != status {
				t.Errorf("Case test[%d]: expected status %q but got %q\n", i, test.status, status)
			}
			// Check error
			if test.err != err {
				t.Errorf("Case test[%d]: expected error %v but got %v\n", i, test.err, err)
			}
			if withDnstap {
				// Check Dnstap attributes
				sender.checkAttributes(t, i, test.attrs)
			}
		}

		// debug
		p.debugSuffix = "debug."
		for i, test := range tests {
			sender.reset()
			// Make request
			req := new(dns.Msg)
			req.Question = make([]dns.Question, 1)
			req.Question[0] = dns.Question{test.query + p.debugSuffix, dns.TypeTXT, dns.ClassCHAOS}
			// Init test mock client
			p.pdp = newTestClientInit(test.response, test.responseIP, test.errResp, test.errRespIP)
			// Handle request
			status, err := p.ServeDNS(context.Background(), rec, req)
			var testStatus int
			var testErr error
			if test.err == errFakePdp {
				testErr = errFakePdp
				testStatus = dns.RcodeServerFailure
			}
			if test.status == dns.RcodeRefused {
				testStatus = dns.RcodeRefused
				testErr = nil
			}
			// Check status
			if testStatus != status {
				t.Errorf("Case debug[%d]: expected status %q but got %q\n", i, testStatus, status)
			}
			// Check error
			if testErr != err {
				t.Errorf("Case debug[%d]: expected error %v but got %v\n", i, testErr, err)
			}
			if withDnstap {
				// Check no Dnstap attributes
				sender.checkAttributes(t, i, nil)
			}
		}

	}

}

func TestPPConnect(t *testing.T) {
	p := newPolicyPlugin()
	p.endpoints = []string{
		"127.0.0.1:5555",
		"127.0.0.2:5555",
	}

	// Test if closing not connected plugin doesn't panic.
	p.closeConn()

	// Test connection.
	err := p.connect()
	if err != nil {
		t.Errorf("Expected no error but got %s", err)
	}

	p.closeConn()
}

func handler() plugin.Handler {
	return plugin.HandlerFunc(func(ctx context.Context, w dns.ResponseWriter, r *dns.Msg) (int, error) {
		q := r.Question[0].Name
		switch q {
		case "test.com.":
			r.Answer = []dns.RR{
				test.A("test.com.		600	IN	A			10.240.0.1"),
			}
		case "test.org.":
			r.Answer = []dns.RR{
				test.AAAA("test.org.	600	IN	AAAA		21da:d3:0:2f3b:2aa:ff:fe28:9c5a"),
			}
		case "redirect.biz.":
			r.Answer = []dns.RR{
				test.A("redirect.biz.	600	IN	A			221.228.88.194"),
			}
		case "redirect.net.":
			r.Answer = []dns.RR{
				test.AAAA("redirect.net.	600	IN	AAAA		2001:db8:0:200:0:0:0:7"),
			}
		case "nxdomain.org.":
			w.WriteMsg(r)
			return dns.RcodeNameError, nil
		default:
			return dns.RcodeServerFailure, errFakeResolver
		}
		w.WriteMsg(r)
		return dns.RcodeSuccess, nil
	})
}

func makeRequestWithEDNS0(code uint16, hexstring string, nonlocal bool) *dns.Msg {
	req := new(dns.Msg)
	o := new(dns.OPT)
	o.Hdr.Name = "."
	o.Hdr.Rrtype = dns.TypeOPT
	if nonlocal {
		e := new(dns.EDNS0_SUBNET)
		o.Option = append(o.Option, e)
	} else {
		e := new(dns.EDNS0_LOCAL)
		e.Code = code
		src := []byte(hexstring)
		dst := make([]byte, hex.DecodedLen(len(src)))
		hex.Decode(dst, src)
		e.Data = dst
		o.Option = append(o.Option, e)
	}
	req.Extra = append(req.Extra, o)
	return req
}

func TestEdns(t *testing.T) {
	p := newPolicyPlugin()

	// Add EDNS mapping
	AttrClientId := "client_id"
	if err := p.addEDNS0Map("0xfffa", AttrClientId, "hex", "string", "32", "0", "16"); err != nil {
		t.Errorf("Expected error 'nil' but got %v\n", err)
	}
	AttrGroupId := "group_id"
	if err := p.addEDNS0Map("0xfffa", AttrGroupId, "hex", "string", "32", "16", "32"); err != nil {
		t.Errorf("Expected error 'nil' but got %v\n", err)
	}
	if err := p.addEDNS0Map("0xfffb", AttrNameSourceIP, "address", "address", "0", "0", "0"); err != nil {
		t.Errorf("Expected error 'nil' but got %v\n", err)
	}
	AttrClientName := "client_name"
	if err := p.addEDNS0Map("0xfffc", AttrClientName, "bytes", "string", "0", "0", "0"); err != nil {
		t.Errorf("Expected error 'nil' but got %v\n", err)
	}
	AttrClientUID := "client_uid"
	if err := p.addEDNS0Map("0xfffd", AttrClientUID, "hex", "string", "0", "0", "0"); err != nil {
		t.Errorf("Expected error 'nil' but got %v\n", err)
	}
	AttrHexName := "hex_name"
	if err := p.addEDNS0Map("0xfffe", AttrHexName, "hex", "string", "0", "2", "0"); err != nil {
		t.Errorf("Expected error 'nil' but got %v\n", err)
	}
	AttrVar := "var"
	if err := p.addEDNS0Map("0xffff", AttrVar, "hex", "string", "0", "2", "6"); err != nil {
		t.Errorf("Expected error 'nil' but got %v\n", err)
	}

	tests := []struct {
		name     string
		code     uint16
		data     string
		ip       string
		attr     map[string]*pdp.Attribute
		nonlocal bool
	}{
		{
			name: "Test different than EDNS0_LOCAL option",
			ip:   "192.168.0.1",
			attr: map[string]*pdp.Attribute{
				AttrNameType:       {Id: AttrNameType, Type: "string", Value: "query"},
				AttrNameDNSQtype:   {Id: AttrNameDNSQtype, Type: "string", Value: "1"},
				AttrNameDomainName: {Id: AttrNameDomainName, Type: "domain", Value: "test.com"},
				AttrNameSourceIP:   {Id: AttrNameSourceIP, Type: "address", Value: "192.168.0.1"},
			},
			nonlocal: true,
		},
		{
			name: "Test option that not in config mapping",
			code: 0xfff9,
			data: "cafecafe",
			ip:   "192.168.0.2",
			attr: map[string]*pdp.Attribute{
				AttrNameType:       {Id: AttrNameType, Type: "string", Value: "query"},
				AttrNameDNSQtype:   {Id: AttrNameDNSQtype, Type: "string", Value: "1"},
				AttrNameDomainName: {Id: AttrNameDomainName, Type: "domain", Value: "test.com"},
				AttrNameSourceIP:   {Id: AttrNameSourceIP, Type: "address", Value: "192.168.0.2"},
			},
		},
		{
			name: "Test option handled as hex",
			code: 0xfffa,
			data: "4e7e318384088e7d4f3dbc96219ee5d4" + "318384088e7d4f3dbc96219ee5d44e7e",
			ip:   "192.168.0.3",
			attr: map[string]*pdp.Attribute{
				AttrNameType:       {Id: AttrNameType, Type: "string", Value: "query"},
				AttrNameDNSQtype:   {Id: AttrNameDNSQtype, Type: "string", Value: "1"},
				AttrNameDomainName: {Id: AttrNameDomainName, Type: "domain", Value: "test.com"},
				AttrNameSourceIP:   {Id: AttrNameSourceIP, Type: "address", Value: "192.168.0.3"},
				AttrClientId:       {Id: AttrClientId, Type: "string", Value: "4e7e318384088e7d4f3dbc96219ee5d4"},
				AttrGroupId:        {Id: AttrGroupId, Type: "string", Value: "318384088e7d4f3dbc96219ee5d44e7e"},
			},
		},
		{
			name: "Test option 'source_ip' handled as address",
			code: 0xfffb,
			data: "aca80001", // 172.168.0.1 in hex
			ip:   "192.168.0.4",
			attr: map[string]*pdp.Attribute{
				AttrNameType:          {Id: AttrNameType, Type: "string", Value: "query"},
				AttrNameDNSQtype:      {Id: AttrNameDNSQtype, Type: "string", Value: "1"},
				AttrNameDomainName:    {Id: AttrNameDomainName, Type: "domain", Value: "test.com"},
				AttrNameSourceIP:      {Id: AttrNameSourceIP, Type: "address", Value: "172.168.0.1"},
				AttrNameProxySourceIP: {Id: AttrNameProxySourceIP, Type: AttrNameAddress, Value: "192.168.0.4"},
			},
		},
		{
			name: "Test option handled as bytes",
			code: 0xfffc,
			data: "637573746f6d6572", // "customer" in hex
			ip:   "192.168.0.5",
			attr: map[string]*pdp.Attribute{
				AttrNameType:       {Id: AttrNameType, Type: "string", Value: "query"},
				AttrNameDNSQtype:   {Id: AttrNameDNSQtype, Type: "string", Value: "1"},
				AttrNameDomainName: {Id: AttrNameDomainName, Type: "domain", Value: "test.com"},
				AttrNameSourceIP:   {Id: AttrNameSourceIP, Type: "address", Value: "192.168.0.5"},
				AttrClientName:     {Id: AttrClientName, Type: "string", Value: "customer"},
			},
		},
		{
			name: "Test option handled as hex, start = 0 and end = 0 (whole data)",
			code: 0xfffd,
			data: "96219ee5d44e7e318384088e7d4f3dbc",
			ip:   "192.168.0.6",
			attr: map[string]*pdp.Attribute{
				AttrNameType:       {Id: AttrNameType, Type: "string", Value: "query"},
				AttrNameDNSQtype:   {Id: AttrNameDNSQtype, Type: "string", Value: "1"},
				AttrNameDomainName: {Id: AttrNameDomainName, Type: "domain", Value: "test.com"},
				AttrNameSourceIP:   {Id: AttrNameSourceIP, Type: "address", Value: "192.168.0.6"},
				AttrClientUID:      {Id: AttrClientUID, Type: "string", Value: "96219ee5d44e7e318384088e7d4f3dbc"},
			},
		},
		{
			name: "Test option handled as hex, start = 2 and end = 0 (from start to end of data)",
			code: 0xfffe,
			data: "8e7d" + "4f3dbc96219ee5d44e7e31838408",
			ip:   "192.168.0.7",
			attr: map[string]*pdp.Attribute{
				AttrNameType:       {Id: AttrNameType, Type: "string", Value: "query"},
				AttrNameDNSQtype:   {Id: AttrNameDNSQtype, Type: "string", Value: "1"},
				AttrNameDomainName: {Id: AttrNameDomainName, Type: "domain", Value: "test.com"},
				AttrNameSourceIP:   {Id: AttrNameSourceIP, Type: "address", Value: "192.168.0.7"},
				AttrHexName:        {Id: AttrHexName, Type: "string", Value: "4f3dbc96219ee5d44e7e31838408"},
			},
		},
		{
			name: "Test skip option with wrong size",
			code: 0xfffa,
			data: "8e7d4f3dbc96219ee5d44e7e31838408",
			ip:   "192.168.0.8",
			attr: map[string]*pdp.Attribute{
				AttrNameType:       {Id: AttrNameType, Type: "string", Value: "query"},
				AttrNameDNSQtype:   {Id: AttrNameDNSQtype, Type: "string", Value: "1"},
				AttrNameDomainName: {Id: AttrNameDomainName, Type: "domain", Value: "test.com"},
				AttrNameSourceIP:   {Id: AttrNameSourceIP, Type: "address", Value: "192.168.0.8"},
			},
		},
		{
			name: "Test skip option if start >= size",
			code: 0xffff,
			data: "0011",
			ip:   "192.168.0.9",
			attr: map[string]*pdp.Attribute{
				AttrNameType:       {Id: AttrNameType, Type: "string", Value: "query"},
				AttrNameDNSQtype:   {Id: AttrNameDNSQtype, Type: "string", Value: "1"},
				AttrNameDomainName: {Id: AttrNameDomainName, Type: "domain", Value: "test.com"},
				AttrNameSourceIP:   {Id: AttrNameSourceIP, Type: "address", Value: "192.168.0.9"},
			},
		},
		{
			name: "Test skip option if end > size",
			code: 0xffff,
			data: "00112233",
			ip:   "192.168.0.10",
			attr: map[string]*pdp.Attribute{
				AttrNameType:       {Id: AttrNameType, Type: "string", Value: "query"},
				AttrNameDNSQtype:   {Id: AttrNameDNSQtype, Type: "string", Value: "1"},
				AttrNameDomainName: {Id: AttrNameDomainName, Type: "domain", Value: "test.com"},
				AttrNameSourceIP:   {Id: AttrNameSourceIP, Type: "address", Value: "192.168.0.10"},
			},
		},
	}

	for _, test := range tests {
		req := makeRequestWithEDNS0(test.code, test.data, test.nonlocal)
		ah := newAttrHolder("test.com", 1)
		p.getAttrsFromEDNS0(ah, req, test.ip)
		mapAttr := make(map[string]*pdp.Attribute)
		for _, a := range ah.attrsReqDomain {
			mapAttr[a.Id] = a
		}
		if len(ah.attrsReqDomain) != len(mapAttr) {
			t.Errorf("%s: array %q transformed to map %q\n", test.name, ah.attrsReqDomain, mapAttr)
		}
		if !reflect.DeepEqual(test.attr, mapAttr) {
			t.Errorf("%s: expected attributes %q but got %q\n", test.name, test.attr, mapAttr)
		}
	}
}

type testDnstapSender struct {
	attrs []*pb.DnstapAttribute
}

func (s *testDnstapSender) reset() {
	s.attrs = nil
}

func (s *testDnstapSender) SendCRExtraMsg(pw *ProxyWriter, ah *attrHolder) {
	if ah != nil {
		s.attrs = ah.convertAttrs()
	}
}

func (s *testDnstapSender) checkAttributes(t *testing.T, i int, attrs []*pdp.Attribute) {
	if attrs == nil {
		if s.attrs != nil {
			t.Errorf("SendCRExtraMsg was unexpectedly called in test %d", i)
		}
		return
	}
	if s.attrs == nil {
		t.Errorf("SendCRExtraMsg was unexpectedly not called in test %d", i)
		return
	}

	if len(attrs) != len(s.attrs) {
		t.Errorf("Not expected number of attributes in test %d\n, expected %q\n, found %q\n", i, attrs, s.attrs)
	}

checkAttr:
	for _, a := range s.attrs {
		for _, e := range attrs {
			if e.Id == a.Id {
				if e.Value != a.Value {
					t.Errorf("Attribute %s: expected %q, found %q in test %d", e.Id, e.Value, a.Value, i)
					return
				}
				continue checkAttr
			}
		}
		t.Errorf("Unexpected attribute found %q in test %d", a, i)
	}
}

type testTracerHandler struct{}

func (t *testTracerHandler) ServeDNS(context.Context, dns.ResponseWriter, *dns.Msg) (int, error) {
	return 0, nil
}

func (t *testTracerHandler) Name() string {
	return "testTracerHandler"
}

func (t *testTracerHandler) Tracer() ot.Tracer {
	return nil
}
