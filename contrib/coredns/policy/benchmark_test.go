package policy

import (
	"bytes"
	"fmt"
	"log"
	"math"
	"net"
	"os"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	_ "github.com/infobloxopen/themis/pdp/selector"
	"github.com/infobloxopen/themis/pdpserver/server"
	"github.com/miekg/dns"
	lr "github.com/sirupsen/logrus"
	"golang.org/x/net/context"
)

func BenchmarkPlugin(b *testing.B) {
	endpoint := "127.0.0.1:5555"
	srv := startPDPServerForBenchmark(b, allPermitTestPolicy, endpoint)
	defer func() {
		if logs := srv.Stop(); len(logs) > 0 {
			b.Logf("server logs:\n%s", logs)
		}
	}()

	if err := waitForPortOpened(endpoint); err != nil {
		b.Fatalf("can't connect to PDP server: %s", err)
	}

	b.Run("1-stream", func(b *testing.B) {
		benchSerial(b, newTestPolicyPlugin(endpoint))
	})

	ps := newParStat()
	if b.Run("100-streams", func(b *testing.B) {
		p := newTestPolicyPlugin(endpoint)
		p.streams = 100
		p.hotSpot = true

		benchParallel(b, p, ps)
	}) {
		b.Logf("Parallel stats:\n%s", ps)
	}
}

func benchSerial(b *testing.B, p *policyPlugin) {
	g := newLogGrabber()
	if err := p.connect(); err != nil {
		b.Fatalf("can't connect to PDP: %s\n=== plugin logs ===\n%s--- plugin logs ---", err, g.Release())
	}
	defer p.closeConn()
	g.Release()

	w := newTestAddressedNonwriter("192.0.2.1")

	g = newLogGrabber()
	for n := 0; n < b.N; n++ {
		m := makeTestDNSMsg("example.com", dns.TypeA, dns.ClassINET)
		w.Msg = nil
		rc, err := p.ServeDNS(context.TODO(), w, m)
		if rc != dns.RcodeSuccess || err != nil {
			b.Fatalf("ServeDNS failed with code: %d, error: %s, message:\n%s\n"+
				"=== plugin logs ===\n%s--- plugin logs ---", rc, err, w.Msg, g.Release())
		}
	}
}

func benchParallel(b *testing.B, p *policyPlugin, ps *parStat) {
	g := newLogGrabber()
	if err := p.connect(); err != nil {
		b.Fatalf("can't connect to PDP: %s\n=== plugin logs ===\n%s--- plugin logs ---", err, g.Release())
	}
	defer p.closeConn()
	g.Release()

	var errCnt uint32
	errCntPtr := &errCnt

	var pCnt uint32
	pCntPtr := &pCnt

	g = newLogGrabber()
	b.SetParallelism(25)
	b.RunParallel(func(pb *testing.PB) {
		atomic.AddUint32(pCntPtr, 1)
		w := newTestAddressedNonwriter("192.0.2.1")
		for pb.Next() {
			m := makeTestDNSMsg("example.com", dns.TypeA, dns.ClassINET)
			w.Msg = nil
			rc, err := p.ServeDNS(context.TODO(), w, m)
			if rc != dns.RcodeSuccess || err != nil {
				atomic.AddUint32(errCntPtr, 1)
				return
			}
		}
	})

	logs := g.Release()
	if errCnt > 0 {
		b.Fatalf("parallel failed %d times\n=== plugin logs ===\n%s--- plugin logs ---", errCnt, logs)
	}

	ps.update(pCnt)
}

const allPermitTestPolicy = `# All Permit Policy
policies:
  alg: FirstApplicableEffect
  rules:
  - effect: Permit
`

func newTestPolicyPlugin(endpoints ...string) *policyPlugin {
	p := newPolicyPlugin()
	p.endpoints = endpoints
	p.connTimeout = time.Second
	p.streams = 1

	mp := &mockPlugin{
		ip: net.ParseIP("192.0.2.53"),
		rc: dns.RcodeSuccess,
	}
	p.next = mp

	return p
}

type parStat struct {
	totalRuns   int
	totalParCnt uint32
	maxParCnt   uint32
	minParCnt   uint32
}

func newParStat() *parStat {
	return &parStat{
		minParCnt: math.MaxUint32,
	}
}

func (ps *parStat) update(n uint32) {
	ps.totalRuns++
	ps.totalParCnt += n
	if n < ps.minParCnt {
		ps.minParCnt = n
	}
	if n > ps.maxParCnt {
		ps.maxParCnt = n
	}
}

func (ps *parStat) String() string {
	if ps.minParCnt == ps.maxParCnt {
		return fmt.Sprintf("Routines: %d", ps.minParCnt)
	}

	return fmt.Sprintf("Runs.........: %d\n"+
		"Avg. routines: %g\n"+
		"Min routines.: %d\n"+
		"Max routines.: %d",
		ps.totalRuns,
		float64(ps.totalParCnt)/float64(ps.totalRuns),
		ps.minParCnt,
		ps.maxParCnt,
	)
}

func startPDPServerForBenchmark(b *testing.B, p, endpoint string) *loggedServer {
	s := newServer(server.WithServiceAt(endpoint))

	if err := s.s.ReadPolicies(strings.NewReader(p)); err != nil {
		b.Fatalf("can't read policies: %s", err)
	}

	if err := waitForPortClosed(endpoint); err != nil {
		b.Fatalf("port still in use: %s", err)
	}

	go func() {
		if err := s.s.Serve(); err != nil {
			b.Fatalf("PDP server failed: %s", err)
		}
	}()

	return s
}

func waitForPortOpened(address string) error {
	var (
		c   net.Conn
		err error
	)

	for i := 0; i < 20; i++ {
		after := time.After(500 * time.Millisecond)
		c, err = net.DialTimeout("tcp", address, 500*time.Millisecond)
		if err == nil {
			return c.Close()
		}

		<-after
	}

	return err
}

func waitForPortClosed(address string) error {
	var (
		c   net.Conn
		err error
	)

	for i := 0; i < 20; i++ {
		after := time.After(500 * time.Millisecond)
		c, err = net.DialTimeout("tcp", address, 500*time.Millisecond)
		if err != nil {
			return nil
		}

		c.Close()
		<-after
	}

	return fmt.Errorf("port at %s hasn't been closed yet", address)
}

type logGrabber struct {
	b *bytes.Buffer
}

func newLogGrabber() *logGrabber {
	b := new(bytes.Buffer)
	log.SetOutput(b)

	return &logGrabber{
		b: b,
	}
}

func (g *logGrabber) Release() string {
	log.SetOutput(os.Stderr)

	return g.b.String()
}

type mockPlugin struct {
	ip  net.IP
	err error
	rc  int
}

// Name implements the plugin.Handler interface.
func (p *mockPlugin) Name() string {
	return "mockPlugin"
}

// ServeDNS implements the plugin.Handler interface.
func (p *mockPlugin) ServeDNS(ctx context.Context, w dns.ResponseWriter, r *dns.Msg) (int, error) {
	if p.err != nil {
		return dns.RcodeServerFailure, p.err
	}

	if r == nil || len(r.Question) <= 0 {
		return dns.RcodeServerFailure, nil
	}

	q := r.Question[0]
	hdr := dns.RR_Header{
		Name:   q.Name,
		Rrtype: q.Qtype,
		Class:  q.Qclass,
	}

	if ipv4 := p.ip.To4(); ipv4 != nil {
		if q.Qtype != dns.TypeA {
			return dns.RcodeSuccess, nil
		}

		m := new(dns.Msg)
		m.SetReply(r)
		m.Authoritative = true
		m.Rcode = p.rc

		if m.Rcode == dns.RcodeSuccess {
			m.Answer = append(m.Answer,
				&dns.A{
					Hdr: hdr,
					A:   ipv4,
				},
			)
		}

		w.WriteMsg(m)
	} else if ipv6 := p.ip.To16(); ipv6 != nil {
		if q.Qtype != dns.TypeAAAA {
			return dns.RcodeSuccess, nil
		}

		m := new(dns.Msg)
		m.SetReply(r)
		m.Authoritative = true
		m.Rcode = p.rc

		if m.Rcode == dns.RcodeSuccess {
			m.Answer = append(m.Answer,
				&dns.AAAA{
					Hdr:  hdr,
					AAAA: ipv6,
				},
			)
		}

		w.WriteMsg(m)
	}

	return p.rc, nil
}

type loggedServer struct {
	s *server.Server
	b *bytes.Buffer
}

func newServer(opts ...server.Option) *loggedServer {
	s := &loggedServer{
		b: new(bytes.Buffer),
	}

	logger := lr.New()
	logger.Out = s.b
	logger.Level = lr.ErrorLevel
	opts = append(opts,
		server.WithLogger(logger),
	)

	s.s = server.NewServer(opts...)
	return s
}

func (s *loggedServer) Stop() string {
	s.s.Stop()
	return s.b.String()
}

func makeTestDNSMsg(n string, t uint16, c uint16) *dns.Msg {
	out := new(dns.Msg)
	out.Question = make([]dns.Question, 1)
	out.Question[0] = dns.Question{
		Name:   dns.Fqdn(n),
		Qtype:  t,
		Qclass: c,
	}
	return out
}

type testAddressedNonwriter struct {
	dns.ResponseWriter
	ra  net.Addr
	Msg *dns.Msg
}

type testUDPAddr struct {
	addr string
}

func newTestAddressedNonwriter(ra string) *testAddressedNonwriter {
	return &testAddressedNonwriter{
		ResponseWriter: nil,
		ra:             newUDPAddr(ra),
	}
}

func newTestAddressedNonwriterWithAddr(ra net.Addr) *testAddressedNonwriter {
	return &testAddressedNonwriter{
		ResponseWriter: nil,
		ra:             ra,
	}
}

func (w *testAddressedNonwriter) RemoteAddr() net.Addr {
	return w.ra
}

func (w *testAddressedNonwriter) WriteMsg(res *dns.Msg) error {
	w.Msg = res
	return nil
}

func newUDPAddr(addr string) *testUDPAddr {
	return &testUDPAddr{
		addr: addr,
	}
}

func (a *testUDPAddr) String() string {
	return a.addr
}

func (a *testUDPAddr) Network() string {
	return "udp"
}
