package policy

import (
	"errors"
	"sync"
	"sync/atomic"
	"time"

	"github.com/coredns/coredns/core/dnsserver"
	"github.com/coredns/coredns/plugin"
	"github.com/coredns/coredns/plugin/metrics"
	"github.com/infobloxopen/themis/pdp"
	"github.com/mholt/caddy"
	"github.com/prometheus/client_golang/prometheus"
)

// BucketCnt is how many slices the total counter consists of
// DefExpire is default period (in seconds) to keep data in counters
const (
	BucketCnt = 16
	DefExpire = 60
)

// SlicedCounter stores the counter value both as a single number and as a separate
// values per quantum period. The sum of values per quantum is synchronized with total
// value but not guaranteed to be equal at any moment of time
// freshCnt is how many last slices hold actual (not outdated) data
// quantum is the time interval (in seconds) corresponding to a single slice
// freshCnt * quantum = target time period for which data is kept in counters
type SlicedCounter struct {
	oldestValid uint32
	total       uint32
	buckets     [BucketCnt]uint32
	freshCnt    uint16
	quantum     uint16
}

// NewSlicedCounter creates new SlicedCounter
func NewSlicedCounter(ut uint32, expiry uint16) *SlicedCounter {
	q, fc := calcQuantum(expiry)
	sc := &SlicedCounter{freshCnt: fc, quantum: q}
	sc.oldestValid = sc.ut2qt(ut)
	return sc
}

// Total returns the counter value
func (sc *SlicedCounter) Total() uint32 {
	return atomic.LoadUint32(&sc.total)
}

// Inc increments the latest and total counters. Can be called simultaneously
// from different goroutines
func (sc *SlicedCounter) Inc(ut uint32) bool {
	qt := sc.ut2qt(ut)
	oldest := atomic.LoadUint32(&sc.oldestValid)
	if qt-oldest >= BucketCnt {
		return false
	}
	atomic.AddUint32(&sc.total, 1)
	atomic.AddUint32(&sc.buckets[qt%BucketCnt], 1)
	return true
}

// EraseStale erases the values from stale buckets, decrements the total counter
// by the sum of erased values, and updates the oldestValid time. Should be run
// in single goroutine
func (sc *SlicedCounter) EraseStale(ut uint32) {
	oldest := atomic.LoadUint32(&sc.oldestValid)
	stale := sc.ut2qt(ut) - uint32(sc.freshCnt)
	if stale >= oldest+BucketCnt {
		oldest = stale - BucketCnt + 1
		atomic.StoreUint32(&sc.oldestValid, oldest)
	}
	for oldest <= stale {
		cnt := atomic.SwapUint32(&sc.buckets[oldest%BucketCnt], 0)
		atomic.AddUint32(&sc.total, -cnt)
		atomic.AddUint32(&sc.oldestValid, 1)
		oldest++
	}
}

func (sc *SlicedCounter) ut2qt(ut uint32) uint32 {
	return ut / uint32(sc.quantum)
}

func calcQuantum(exp uint16) (q, fc uint16) {
	maxFreshCnt := uint16(BucketCnt - 2)
	q = exp / maxFreshCnt
	if exp%maxFreshCnt != 0 {
		q++
	}
	fc = exp / q
	if exp%q != 0 {
		fc++
	}
	return
}

const (
	AttrGaugeStopped = iota
	AttrGaugeStarted
	AttrGaugeStopping
)

const (
	DefaultQueryChanSize = 1000
)

// AttrGauge manages GaugeVec for attributes. GaugeVec holds the
// counters for recently received (last FreshCnt seconds) DNS queries
// per attribute/value
type AttrGauge struct {
	perAttr  map[string]map[string]*SlicedCounter
	pgv      *prometheus.GaugeVec
	qChan    chan pdp.AttributeAssignment
	nameChan chan string
	timeFunc func() uint32
	errCnt   uint32
	state    uint32
	expire   uint16
}

// NewAttrGauge constructs new AttrGauge object
func NewAttrGauge() *AttrGauge {
	return &AttrGauge{
		perAttr: make(map[string]map[string]*SlicedCounter),
		pgv: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: plugin.Namespace,
			Subsystem: "policy",
			Name:      "recent_queries",
			Help:      "Gauge of recent queries per Attrubute value.",
		}, []string{"attribute", "value"}),
		qChan:    make(chan pdp.AttributeAssignment),
		nameChan: make(chan string),
		timeFunc: unixTime,
		expire:   DefExpire,
	}
}

// globalAttrGauge object is used across all policy plugin instances
// including instances from different corefile blocks and
// new/old instances during graceful restart
var globalAttrGauge = NewAttrGauge()

// Start starts goroutine which reads and handles data from channels
func (g *AttrGauge) Start(tickInt time.Duration, chSize int) {
	if atomic.CompareAndSwapUint32(&g.state, AttrGaugeStopped, AttrGaugeStarted) {
		ch := make(chan pdp.AttributeAssignment, chSize)
		g.qChan = ch
		go func() {
			timer := time.NewTimer(tickInt)
			for {
				if atomic.CompareAndSwapUint32(&g.state, AttrGaugeStopping, AttrGaugeStopped) {
					break
				}
				select {
				case name := <-g.nameChan:
					g.addAttribute(name)
				case attr := <-ch:
					g.synchInc(attr)
				case <-timer.C:
					eCnt := g.tick()
					if eCnt != 0 {
						log.Warningf("Policy metrics: %d queries was skipped", eCnt)
					}
					timer.Reset(tickInt)
				}
			}
		}()
	}
}

// Stop stops goroutine which reads and handles data from channels
func (g *AttrGauge) Stop() {
	if g == nil {
		return
	}
	if !atomic.CompareAndSwapUint32(&g.state, AttrGaugeStarted, AttrGaugeStopping) {
		return
	}
	for atomic.LoadUint32(&g.state) != AttrGaugeStopped {
		time.Sleep(10 * time.Millisecond)
	}
}

// AddAttribute adds new attribute names to gauge. It's safe to call it from
// any goroutine. The AttrGauge should be started before calling AddAttributes
func (g *AttrGauge) AddAttributes(attrNames ...string) {
	for _, name := range attrNames {
		g.nameChan <- name
	}
}

// addAttribute adds new attribute name to gauge. Should be called synchronously
func (g *AttrGauge) addAttribute(attrName string) {
	if g.perAttr[attrName] == nil {
		g.perAttr[attrName] = make(map[string]*SlicedCounter)
	}
}

// Inc increments the counter corresponding to the attr. It's safe
// to call it from any goroutine. The AttrGauge should be started before
// calling Inc
func (g *AttrGauge) Inc(attr pdp.AttributeAssignment) {
	if g == nil {
		return
	}

	select {
	case g.qChan <- attr:
	default:
		g.ErrorInc()
	}
}

// synchInc increments internal counter corresponding to the attr.
// The actual prometheus value is not updated in this method.
// Should be called synchronously
func (g *AttrGauge) synchInc(attr pdp.AttributeAssignment) {
	ut := g.timeFunc()
	id := attr.GetID()
	v := serializeOrPanic(attr)
	sc := g.perAttr[id][v]
	if sc == nil {
		sc = NewSlicedCounter(ut, g.expire)
		g.perAttr[id][v] = sc
	}
	if sc.Inc(ut) {
		return
	}
	g.ErrorInc()
}

// tick synchronises prometheus gauge with internal counters.
// Should be called synchronously
func (g *AttrGauge) tick() uint32 {
	ut := g.timeFunc()
	for attr, amap := range g.perAttr {
		for val, sc := range amap {
			sc.EraseStale(ut)
			total := sc.Total()
			if total > 0 {
				g.pgv.WithLabelValues(attr, val).Set(float64(total))
				continue
			}
			g.pgv.DeleteLabelValues(attr, val)
			delete(amap, val)
		}
		g.pgv.WithLabelValues(attr, "VALUES_COUNT").Set(float64(len(amap)))
	}
	return atomic.SwapUint32(&g.errCnt, 0)
}

// ErrorInc increments error counter
func (g *AttrGauge) ErrorInc() {
	atomic.AddUint32(&g.errCnt, 1)
}

// unixTime returns number of seconds since Unix epoch
func unixTime() uint32 {
	return uint32(time.Now().Unix())
}

// SetupMetrics checks for configured metrics attributes and starts and
// configures globalAttrGauge as needed
func (pp *policyPlugin) SetupMetrics(c *caddy.Controller) error {
	attrNames := []string{}
	for _, conf := range pp.conf.attrs.confLists[attrListTypeMetrics] {
		attrNames = append(attrNames, conf.name)
	}
	if len(attrNames) > 0 {
		if mh := dnsserver.GetConfig(c).Handler("prometheus"); mh != nil {
			if m, ok := mh.(*metrics.Metrics); ok {
				m.MustRegister(globalAttrGauge.pgv)
				metricsOnce.Do(func() {
					// The globalAttrGauge is started once and is not stopped
					// until process termination
					q, _ := calcQuantum(globalAttrGauge.expire)
					globalAttrGauge.Start(time.Duration(q)*time.Second, DefaultQueryChanSize)
				})
				globalAttrGauge.AddAttributes(attrNames...)
				pp.attrGauges = globalAttrGauge
				return nil
			}
		}
		return errors.New("can't find prometheus plugin")
	}
	return nil
}

var metricsOnce sync.Once
