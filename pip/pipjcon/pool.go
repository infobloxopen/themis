package main

import (
	"sync"

	"github.com/infobloxopen/themis/pdp"
)

type argsPool struct {
	a *sync.Pool
}

func makeArgsPool(size int) argsPool {
	return argsPool{
		a: &sync.Pool{
			New: func() interface{} {
				return &argsBuffer{
					a: make([]pdp.AttributeValue, size),
				}
			},
		},
	}
}

func (p argsPool) Get() *argsBuffer {
	return p.a.Get().(*argsBuffer)
}

func (p argsPool) Put(a *argsBuffer) {
	a.a = a.a[:cap(a.a)]
	p.a.Put(a)
}

type argsBuffer struct {
	a []pdp.AttributeValue
}
