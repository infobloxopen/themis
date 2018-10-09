package client

import "sync"

type bytePool struct {
	b *sync.Pool
}

func makeBytePool(size int) bytePool {
	return bytePool{
		b: &sync.Pool{
			New: func() interface{} {
				return make([]byte, size)
			},
		},
	}
}

func (p bytePool) Get() []byte {
	return p.b.Get().([]byte)
}

func (p bytePool) Put(b []byte) {
	p.b.Put(b[:cap(b)])
}
