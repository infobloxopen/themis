package client

import "sync"

type byteBufferPool struct {
	b *sync.Pool
}

func makeByteBufferPool(size int) byteBufferPool {
	return byteBufferPool{
		b: &sync.Pool{
			New: func() interface{} {
				return &byteBuffer{
					b: make([]byte, size),
				}
			},
		},
	}
}

func (p byteBufferPool) Get() *byteBuffer {
	return p.b.Get().(*byteBuffer)
}

func (p byteBufferPool) Put(b *byteBuffer) {
	b.b = b.b[:cap(b.b)]
	p.b.Put(b)
}

type byteBuffer struct {
	b []byte
}
