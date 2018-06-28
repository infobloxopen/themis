package policy

import (
	"log"
	"sync"
	"sync/atomic"
	"time"

	"github.com/coredns/coredns/plugin"
	pdp "github.com/infobloxopen/themis/pdp-service"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	BucketCnt = 16
	FreshCnt  = 10
)

// SlicedCounter stores the counter value both as a single number and as a separate
// values per second. The sum of values per second is synchronized with total value
// but not guaranteed to be equal at any moment of time
type SlicedCounter struct {
	oldestValid uint32
	total       uint32
	buckets     [BucketCnt]uint32
}

// NewSlicedCounter creates new SlicedCounter
func NewSlicedCounter(ut uint32) *SlicedCounter {
	return &SlicedCounter{oldestValid: ut}
}

// Total returns the counter value
func (sc *SlicedCounter) Total() uint32 {
	return atomic.LoadUint32(&sc.total)
}

// Inc increments the latest and total counters. Can be called simultaneously
// from different goroutines
func (sc *SlicedCounter) Inc(ut uint32) bool {
	oldest := atomic.LoadUint32(&sc.oldestValid)
	if ut-oldest >= BucketCnt {
		return false
	}
	atomic.AddUint32(&sc.total, 1)
	atomic.AddUint32(&sc.buckets[ut%BucketCnt], 1)
	return true
}

// EraseStale erases the values from stale buckets, decrements the total counter
// by the sum of erased values, and updates the oldestValid time. Should be run
// in single goroutine
func (sc *SlicedCounter) EraseStale(ut uint32) {
	oldest := atomic.LoadUint32(&sc.oldestValid)
	stale := ut - FreshCnt
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

const (
	AttrGaugeStopped = iota
	AttrGaugeStarted
	AttrGaugeStopping
)

const (
	DefaultEraseInterval   = 500 * time.Millisecond
	DefaultMetricsChanSize = 1000
)

func init() {
	initClobalGauge()
}

func initClobalGauge() {
	globalGauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: plugin.Namespace,
		Subsystem: "policy",
		Name:      "recent_queries",
		Help:      "Gauge of recent queries per Attrubute value.",
	}, []string{"attribute", "value"})
	globalPerAttr = make(map[string]map[string]*SlicedCounter)
}

var (
	globalGauge   *prometheus.GaugeVec
	globalPerAttr map[string]map[string]*SlicedCounter
)

// AttrGauge manages GaugeVec for attributes. GaugeVec holds the
// counters for recently received (last FreshCnt seconds) DNS queries
// per attribute/value
type AttrGauge struct {
	perAttr  map[string]map[string]*SlicedCounter
	pgv      *prometheus.GaugeVec
	qChan    chan *pdp.Attribute
	timeFunc func() uint32
	errCnt   uint32
	state    uint32
}

func NewAttrGauge(attrs ...string) *AttrGauge {
	g := &AttrGauge{
		perAttr:  globalPerAttr,
		pgv:      globalGauge,
		qChan:    make(chan *pdp.Attribute),
		timeFunc: unixTime,
	}
	g.AddAttributes(attrs...)
	return g
}

func (g *AttrGauge) AddAttributes(attrNames ...string) {
	for _, attr := range attrNames {
		if g.perAttr[attr] == nil {
			g.perAttr[attr] = make(map[string]*SlicedCounter)
		}
	}
}

func (g *AttrGauge) Start(tickInt time.Duration, chSize int) {
	if atomic.CompareAndSwapUint32(&g.state, AttrGaugeStopped, AttrGaugeStarted) {
		ch := make(chan *pdp.Attribute, chSize)
		g.qChan = ch
		go func() {
			timer := time.NewTimer(tickInt)
			for {
				if atomic.CompareAndSwapUint32(&g.state, AttrGaugeStopping, AttrGaugeStopped) {
					break
				}
				select {
				case attr := <-ch:
					g.synchInc(attr)
				case <-timer.C:
					eCnt := g.Tick()
					if eCnt != 0 {
						log.Printf("[WARN] Policy metrics: %d queries was skipped", eCnt)
					}
					timer.Reset(tickInt)
				}
			}
		}()
	}
}

func (g *AttrGauge) Stop() {
	if g != nil {
		atomic.CompareAndSwapUint32(&g.state, AttrGaugeStarted, AttrGaugeStopping)
	}
}

func (g *AttrGauge) Inc(attr *pdp.Attribute) {
	if g == nil {
		return
	}

	select {
	case g.qChan <- attr:
	default:
		g.ErrorInc()
	}
}

func (g *AttrGauge) synchInc(attr *pdp.Attribute) {
	ut := g.timeFunc()
	sc := g.perAttr[attr.Id][attr.Value]
	if sc == nil {
		sc = NewSlicedCounter(ut)
		g.perAttr[attr.Id][attr.Value] = sc
	}
	if sc.Inc(ut) {
		return
	}
	g.ErrorInc()
}

func (g *AttrGauge) Tick() uint32 {
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

func (g *AttrGauge) ErrorInc() {
	atomic.AddUint32(&g.errCnt, 1)
}

func unixTime() uint32 {
	return uint32(time.Now().Unix())
}

func (pp *policyPlugin) SetupMetrics() bool {
	attrNames := []string{}
	for attr, t := range pp.confAttrs {
		if !t.isMetrics() {
			continue
		}

		attrNames = append(attrNames, attr)

		for _, list := range pp.options {
			for _, opt := range list {
				if opt.name == attr {
					opt.metrics = true
				}
			}
		}
	}
	if len(attrNames) > 0 {
		pp.attrGauges = NewAttrGauge(attrNames...)
		pp.attrGauges.Start(DefaultEraseInterval, DefaultMetricsChanSize)
		return true
	}
	return false
}

var metricsOnce sync.Once
