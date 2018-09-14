package policy

import (
	"context"
	"fmt"
	"math"
	"net"
	"strconv"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/infobloxopen/themis/contrib/coredns/policy/testutil"
	_ "github.com/infobloxopen/themis/pdp/selector"
	"github.com/infobloxopen/themis/pdpserver/server"
	"github.com/miekg/dns"
)

func BenchmarkPlugin(b *testing.B) {
	endpoint := "127.0.0.1:5555"
	srv := startPDPServerForBenchmark(b, allPermitTestPolicy, endpoint)
	defer func() {
		if logs := srv.Stop(); len(logs) > 0 {
			b.Logf("server logs:\n%s", logs)
		}
	}()

	if err := testutil.WaitForPortOpened(endpoint); err != nil {
		b.Fatalf("can't connect to PDP server: %s", err)
	}

	b.Run("1-stream", func(b *testing.B) {
		benchSerial(b, newTestPolicyPlugin(testutil.MpModeConst, endpoint))
	})

	b.Run("1-stream-cache", func(b *testing.B) {
		p := newTestPolicyPlugin(testutil.MpModeConst, endpoint)
		p.conf.cacheTTL = 10 * time.Minute
		p.conf.cacheLimit = 128

		benchSerial(b, p)
	})

	ps := newParStat()
	if b.Run("100-streams", func(b *testing.B) {
		p := newTestPolicyPlugin(testutil.MpModeConst, endpoint)
		p.conf.streams = 100
		p.conf.hotSpot = true

		benchParallel(b, p, ps)
	}) {
		b.Logf("Parallel stats:\n%s", ps)
	}

	ps = newParStat()
	if b.Run("100-streams-cache-100%-hit", func(b *testing.B) {
		p := newTestPolicyPlugin(testutil.MpModeConst, endpoint)
		p.conf.streams = 100
		p.conf.hotSpot = true
		p.conf.cacheTTL = 10 * time.Minute
		p.conf.cacheLimit = 128

		benchParallel(b, p, ps)
	}) {
		b.Logf("Parallel stats:\n%s", ps)
	}

	ps = newParStat()
	if b.Run("100-streams-cache-50%-hit", func(b *testing.B) {
		p := newTestPolicyPlugin(testutil.MpModeHalfInc, endpoint)
		p.conf.streams = 100
		p.conf.hotSpot = true
		p.conf.cacheTTL = 10 * time.Minute
		p.conf.cacheLimit = 128

		benchParallelHalfHits(b, p, ps)
	}) {
		b.Logf("Parallel stats:\n%s", ps)
	}

	ps = newParStat()
	if b.Run("100-streams-cache-0%-hit", func(b *testing.B) {
		p := newTestPolicyPlugin(testutil.MpModeInc, endpoint)
		p.conf.streams = 100
		p.conf.hotSpot = true
		p.conf.cacheTTL = 10 * time.Minute
		p.conf.cacheLimit = 128

		benchParallelNoHits(b, p, ps)
	}) {
		b.Logf("Parallel stats:\n%s", ps)
	}
}

func benchSerial(b *testing.B, p *policyPlugin) {
	g := testutil.NewLogGrabber()
	if err := p.connect(); err != nil {
		b.Fatalf("can't connect to PDP: %s\n=== plugin logs ===\n%s--- plugin logs ---", err, g.Release())
	}
	defer p.closeConn()
	g.Release()

	w := testutil.NewTestAddressedNonwriter("192.0.2.1")

	g = testutil.NewLogGrabber()
	for n := 0; n < b.N; n++ {
		m := testutil.MakeTestDNSMsg("example.com", dns.TypeA, dns.ClassINET)
		w.Msg = nil
		rc, err := p.ServeDNS(context.TODO(), w, m)
		if rc != dns.RcodeSuccess || err != nil {
			b.Fatalf("ServeDNS failed with code: %d, error: %s, message:\n%s\n"+
				"=== plugin logs ===\n%s--- plugin logs ---", rc, err, w.Msg, g.Release())
		}
	}
}

func benchParallel(b *testing.B, p *policyPlugin, ps *parStat) {
	g := testutil.NewLogGrabber()
	if err := p.connect(); err != nil {
		b.Fatalf("can't connect to PDP: %s\n=== plugin logs ===\n%s--- plugin logs ---", err, g.Release())
	}
	defer p.closeConn()
	g.Release()

	var errCnt uint32
	errCntPtr := &errCnt

	var pCnt uint32
	pCntPtr := &pCnt

	g = testutil.NewLogGrabber()
	b.SetParallelism(25)
	b.RunParallel(func(pb *testing.PB) {
		atomic.AddUint32(pCntPtr, 1)
		w := testutil.NewTestAddressedNonwriter("192.0.2.1")
		for pb.Next() {
			m := testutil.MakeTestDNSMsg("example.com", dns.TypeA, dns.ClassINET)
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

func benchParallelNoHits(b *testing.B, p *policyPlugin, ps *parStat) {
	g := testutil.NewLogGrabber()
	if err := p.connect(); err != nil {
		b.Fatalf("can't connect to PDP: %s\n=== plugin logs ===\n%s--- plugin logs ---", err, g.Release())
	}
	defer p.closeConn()
	g.Release()

	var errCnt uint32
	errCntPtr := &errCnt

	var pCnt uint32
	pCntPtr := &pCnt

	g = testutil.NewLogGrabber()
	b.SetParallelism(25)
	b.RunParallel(func(pb *testing.PB) {
		i := int(atomic.AddUint32(pCntPtr, 1))
		w := testutil.NewTestAddressedNonwriter("192.0.2.1")

		j := 0
		for pb.Next() {
			j++

			m := testutil.MakeTestDNSMsg(strconv.Itoa(i)+"."+strconv.Itoa(j)+".example.com", dns.TypeA, dns.ClassINET)
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

func benchParallelHalfHits(b *testing.B, p *policyPlugin, ps *parStat) {
	g := testutil.NewLogGrabber()
	if err := p.connect(); err != nil {
		b.Fatalf("can't connect to PDP: %s\n=== plugin logs ===\n%s--- plugin logs ---", err, g.Release())
	}
	defer p.closeConn()
	g.Release()

	var errCnt uint32
	errCntPtr := &errCnt

	var pCnt uint32
	pCntPtr := &pCnt

	g = testutil.NewLogGrabber()
	b.SetParallelism(25)
	b.RunParallel(func(pb *testing.PB) {
		i := int(atomic.AddUint32(pCntPtr, 1))
		w := testutil.NewTestAddressedNonwriter("192.0.2.1")

		j := 0
		for pb.Next() {
			j++

			dn := "example.com"
			if j&1 == 0 {
				dn = strconv.Itoa(i) + "." + strconv.Itoa(j) + "." + dn
			}

			m := testutil.MakeTestDNSMsg(dn, dns.TypeA, dns.ClassINET)
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
attributes:
  policy_action: integer
policies:
  alg: FirstApplicableEffect
  rules:
  - effect: Permit
    obligations:
    - policy_action:
        val:
          type: integer
          content: 2
`

func newTestPolicyPlugin(mpMode int, endpoints ...string) *policyPlugin {
	p := newPolicyPlugin()
	p.conf.endpoints = endpoints
	p.conf.connTimeout = time.Second
	p.conf.streams = 1
	p.conf.maxReqSize = 256
	p.conf.autoResAttrs = true
	p.conf.attrs.parseAttrList(attrListTypeVal1, attrNameDomainName)
	p.conf.attrs.parseAttrList(attrListTypeVal2, attrNameAddress)

	mp := &testutil.MockPlugin{
		Ip: net.ParseIP("192.0.2.53"),
		Rc: dns.RcodeSuccess,
	}

	mp.Mode = mpMode
	if mp.Mode != testutil.MpModeConst {
		var cnt uint32
		mp.Cnt = &cnt
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

func startPDPServerForBenchmark(b *testing.B, p, endpoint string) *testutil.LoggedServer {
	s := testutil.NewServer(server.WithServiceAt(endpoint))

	if err := s.S.ReadPolicies(strings.NewReader(p)); err != nil {
		b.Fatalf("can't read policies: %s", err)
	}

	if err := testutil.WaitForPortClosed(endpoint); err != nil {
		b.Fatalf("port still in use: %s", err)
	}

	go func() {
		if err := s.S.Serve(); err != nil {
			b.Fatalf("PDP server failed: %s", err)
		}
	}()

	return s
}
