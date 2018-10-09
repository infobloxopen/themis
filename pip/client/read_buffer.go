package client

import (
	"encoding/binary"
	"errors"
	"io"
)

var (
	errMsgOverflow     = errors.New("message is too big")
	errMsgUnderflow    = errors.New("message is too short")
	errMsgInvalidIndex = errors.New("invalid message index")
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
		rb.pool.Put(rb.msgBuf)
	}
}

func (rb *readBuffer) read(r io.ReadCloser) error {
	n, err := r.Read(rb.in)
	if n > 0 {
		b := rb.in[:n]
		for len(b) > 0 {
			m, err := rb.extractData(b)
			if err != nil {
				return err
			}

			b = b[m:]
		}
	}

	return err
}

func (rb *readBuffer) extractData(b []byte) (int, error) {
	if rb.idx >= 0 {
		return rb.fillMsg(b)
	}

	if rb.size > 0 {
		return rb.fillIdx(b)
	}

	return rb.fillSize(b)
}

func (rb *readBuffer) fillSize(b []byte) (int, error) {
	a := rb.buf
	n := msgSizeBytes - len(a)
	if n > len(b) {
		rb.buf = append(a, b...)
		return len(b), nil
	}

	size := binary.LittleEndian.Uint32(append(a, b[:n]...))
	rb.buf = a[:0]
	if size > rb.max {
		return n, errMsgOverflow
	}

	rb.size = int(size)
	if rb.size < msgIdxBytes {
		return n, errMsgUnderflow
	}

	return n, nil
}

func (rb *readBuffer) fillIdx(b []byte) (int, error) {
	a := rb.buf
	n := msgIdxBytes - len(a)
	if n > len(b) {
		rb.buf = append(a, b...)
		rb.size -= len(b)
		return len(b), nil
	}

	idx := binary.LittleEndian.Uint32(append(a, b[:n]...))
	rb.buf = a[:0]
	if idx >= uint32(len(rb.p.p)) {
		return n, errMsgInvalidIndex
	}

	rb.size -= n
	rb.idx = int(idx)
	rb.msgBuf = rb.pool.Get()[:0]

	return n, nil
}

func (rb *readBuffer) fillMsg(b []byte) (int, error) {
	a := rb.msgBuf
	n := rb.size
	if n > len(b) {
		rb.msgBuf = append(a, b...)
		rb.size -= len(b)
		return len(b), nil
	}

	if !rb.p.putBytes(rb.idx, append(a, b[:n]...)) {
		rb.pool.Put(rb.msgBuf)
	}
	rb.size, rb.idx, rb.msgBuf = 0, -1, nil

	return n, nil
}
