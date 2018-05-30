package perf

import (
	"sync"

	"github.com/infobloxopen/themis/pdp"
)

type obligationsPool struct {
	p *sync.Pool
}

func makeObligationsPool(count uint32) obligationsPool {
	return obligationsPool{
		p: &sync.Pool{
			New: func() interface{} {
				return make([]pdp.AttributeAssignmentExpression, count)
			},
		},
	}
}

func (p obligationsPool) get() []pdp.AttributeAssignmentExpression {
	return p.p.Get().([]pdp.AttributeAssignmentExpression)
}

func (p obligationsPool) put(o []pdp.AttributeAssignmentExpression) {
	p.p.Put(o)
}
