package client

import (
	"encoding/binary"
	"io"
)

type readBuffer struct {
	in     []byte
	buf    []byte
	msgBuf []byte
	size   int
	max    uint32
	idx    int
	pool   bytePool
	p      pipes
}

func newReadBuffer(n, m int, pool bytePool, p pipes) *readBuffer {
	return &readBuffer{
		in:   make([]byte, n),
		buf:  make([]byte, 0, msgSizeBytes),
		max:  uint32(m),
		idx:  -1,
		pool: pool,
		p:    p,
	}
}

func (rb *readBuffer) finalize() {
	if rb.msgBuf != nil {
		rb.pool.Put(rb.msgBuf[:cap(rb.msgBuf)])
	}
}

func (rb *readBuffer) read(r io.ReadCloser) bool {
	n, err := r.Read(rb.in)
	if n > 0 {
		b := rb.in[:n]
		for len(b) > 0 {
			m, ok := rb.extractData(b)
			if !ok {
				r.Close()
				return false
			}

			b = b[m:]
		}
	}

	return err == nil
}

func (rb *readBuffer) extractData(b []byte) (int, bool) {
	if rb.idx >= 0 {
		return rb.fillMsg(b)
	}

	if rb.size > 0 {
		return rb.fillIdx(b)
	}

	return rb.fillSize(b)
}

func (rb *readBuffer) fillSize(b []byte) (int, bool) {
	a := rb.buf
	n := msgSizeBytes - len(a)
	if n > len(b) {
		rb.buf = append(a, b...)
		return len(b), true
	}

	size := binary.LittleEndian.Uint32(append(a, b[:n]...))
	rb.buf = a[:0]
	if size > rb.max {
		return n, false
	}

	rb.size = int(size)
	if rb.size < msgIdxBytes {
		return n, false
	}

	return n, true
}

func (rb *readBuffer) fillIdx(b []byte) (int, bool) {
	a := rb.buf
	n := msgIdxBytes - len(a)
	if n > len(b) {
		rb.buf = append(a, b...)
		rb.size -= len(b)
		return len(b), true
	}

	idx := binary.LittleEndian.Uint32(append(a, b[:n]...))
	rb.buf = a[:0]
	if idx >= uint32(len(rb.p.p)) {
		return n, false
	}

	rb.size -= n
	rb.idx = int(idx)
	rb.msgBuf = rb.pool.Get()[:0]

	return n, true
}

func (rb *readBuffer) fillMsg(b []byte) (int, bool) {
	a := rb.msgBuf
	n := rb.size
	if n > len(b) {
		rb.msgBuf = append(a, b...)
		rb.size -= len(b)
		return len(b), true
	}

	rb.p.putBytes(rb.idx, append(a, b[:n]...))
	rb.size, rb.idx, rb.msgBuf = 0, -1, nil

	return n, true
}
