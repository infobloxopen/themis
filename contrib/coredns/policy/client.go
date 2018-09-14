package policy

import (
	"sync/atomic"

	"github.com/coredns/coredns/plugin/pkg/trace"
	"github.com/infobloxopen/themis/pdp"
	"github.com/infobloxopen/themis/pep"
)

type pepCacheHitHandler struct{}

func (ch *pepCacheHitHandler) Handle(req interface{}, resp interface{}, err error) {
	log.Infof("PEP responding to PDP request from cache %+v", req)
}

func newPepCacheHitHandler() *pepCacheHitHandler {
	return &pepCacheHitHandler{}
}

// connect establishes connection to PDP server.
func (p *policyPlugin) connect() error {
	log.Infof("Connecting %v", p)

	for _, addr := range p.conf.endpoints {
		p.connAttempts[addr] = new(uint32)
	}

	opts := []pep.Option{
		pep.WithConnectionTimeout(p.conf.connTimeout),
		pep.WithConnectionStateNotification(p.connStateCb),
	}

	if p.conf.cacheTTL > 0 {
		if p.conf.cacheLimit > 0 {
			opts = append(opts, pep.WithCacheTTLAndMaxSize(p.conf.cacheTTL, p.conf.cacheLimit))
		} else {
			opts = append(opts, pep.WithCacheTTL(p.conf.cacheTTL))
		}
	}

	if p.conf.streams <= 0 || !p.conf.hotSpot {
		opts = append(opts, pep.WithRoundRobinBalancer(p.conf.endpoints...))
	}

	if p.conf.streams > 0 {
		opts = append(opts, pep.WithStreams(p.conf.streams))
		if p.conf.hotSpot {
			opts = append(opts, pep.WithHotSpotBalancer(p.conf.endpoints...))
		}
	}

	opts = append(opts, pep.WithAutoRequestSize(p.conf.autoReqSize))
	if p.conf.maxReqSize > 0 {
		opts = append(opts, pep.WithMaxRequestSize(uint32(p.conf.maxReqSize)))
	}

	if p.trace != nil {
		if t, ok := p.trace.(trace.Trace); ok {
			opts = append(opts, pep.WithTracer(t.Tracer()))
		}
	}

	if p.conf.log {
		opts = append(opts, pep.WithOnCacheHitHandler(newPepCacheHitHandler()))
	}

	p.pdp = pep.NewClient(opts...)
	return p.pdp.Connect("")
}

// closeConn terminates previously established connection.
func (p *policyPlugin) closeConn() {
	if p.pdp != nil {
		go func() {
			p.wg.Wait()
			p.pdp.Close()
		}()
	}
}

func (p *policyPlugin) validate(buf []pdp.AttributeAssignment, ah *attrHolder, lt int, dbg *dbgMessenger) (bool, error) {
	req := ah.attrList(buf, lt)
	if len(req) <= 0 {
		return true, nil
	}

	if p.conf.log {
		log.Infof("PDP request: %+v", req)
	}

	var respBuffer []pdp.AttributeAssignment
	if !p.conf.autoResAttrs {
		respBuffer = p.attrPool.Get()
		defer p.attrPool.Put(respBuffer)
	}

	res := pdp.Response{Obligations: respBuffer}
	err := p.pdp.Validate(req, &res)

	ah.resetAttribute(attrIndexPolicyAction)
	ah.resetAttribute(attrIndexRedirectTo)
	ah.resetAttribute(attrIndexLog)

	if err != nil {
		log.Errorf("Policy validation failed due to error: %s", err)
		if dbg != nil {
			dbg.appendDefaultDecision(ah.attrList(buf, attrListTypeDefDecision))
		} else {
			ah.resetAttrList(attrListTypeDefDecision)
		}
		if ah.actionValue() != actionInvalid {
			return true, nil
		}
		return false, err
	}

	if p.conf.log {
		log.Infof("PDP response: %+v", res)
	}
	if dbg != nil {
		dbg.appendResponse(&res)
	}

	if res.Effect != pdp.EffectPermit && res.Effect != pdp.EffectDeny {
		log.Errorf("Policy validation failed due to PDP error: %s", res.Status)
		if dbg != nil {
			dbg.appendDefaultDecision(ah.attrList(buf, attrListTypeDefDecision))
		} else {
			ah.resetAttrList(attrListTypeDefDecision)
		}
		if ah.actionValue() != actionInvalid {
			return true, nil
		}
		return false, err
	}

	ah.addAttrList(res.Obligations)

	return res.Effect == pdp.EffectPermit, nil
}

func (p *policyPlugin) connStateCb(addr string, state int, err error) {
	switch state {
	default:
		if err != nil {
			log.Infof("Unknown connection notification %s (%s)", addr, err)
		} else {
			log.Infof("Unknown connection notification %s", addr)
		}

	case pep.StreamingConnectionEstablished:
		ptr, ok := p.connAttempts[addr]
		if !ok {
			ptr = p.unkConnAttempts
		}
		atomic.StoreUint32(ptr, 0)

		log.Infof("Connected to %s", addr)

	case pep.StreamingConnectionBroken:
		log.Errorf("Connection to %s has been broken", addr)

	case pep.StreamingConnectionConnecting:
		ptr, ok := p.connAttempts[addr]
		if !ok {
			ptr = p.unkConnAttempts
		}
		count := atomic.AddUint32(ptr, 1)

		if count <= 1 {
			log.Infof("Connecting to %s", addr)
		}

		if count > 100 {
			log.Errorf("Connecting to %s", addr)
			atomic.StoreUint32(ptr, 1)
		}

	case pep.StreamingConnectionFailure:
		ptr, ok := p.connAttempts[addr]
		if !ok {
			ptr = p.unkConnAttempts
		}
		if atomic.LoadUint32(ptr) <= 1 {
			log.Errorf("Failed to connect to %s (%s)", addr, err)
		}
	}
}
