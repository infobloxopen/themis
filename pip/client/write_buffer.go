package client

import (
	"encoding/binary"
	"errors"
	"net"
)

var errWriterBroken = errors.New("writer has been broken")

const (
	msgSizeBytes = 4
	msgIdxBytes  = 4
)

type writeBuffer struct {
	c   net.Conn
	out []byte
	idx []int
	p   pipes
}

func newWriteBuffer(c net.Conn, n int, p pipes) *writeBuffer {
	return &writeBuffer{
		c:   c,
		out: make([]byte, 0, n),
		idx: make([]int, 0, n/(msgSizeBytes+msgIdxBytes)),
		p:   p,
	}
}

func (w *writeBuffer) rem() int {
	return cap(w.out) - len(w.out)
}

func (w *writeBuffer) put(r request) {
	if w.c == nil {
		w.p.putError(r.i, errWriterBroken)
		return
	}

	size := msgIdxBytes + len(r.b)
	if w.rem() < msgSizeBytes+size {
		w.rawFlush()
	}

	i := len(w.out)
	w.out = append(w.out, 0, 0, 0, 0, 0, 0, 0, 0)
	binary.LittleEndian.PutUint32(w.out[i:], uint32(size))
	binary.LittleEndian.PutUint32(w.out[i+msgSizeBytes:], uint32(r.i))
	w.out = append(w.out, r.b...)
	w.idx = append(w.idx, r.i)

	if w.rem() <= 0 {
		w.rawFlush()
	}
}

func (w *writeBuffer) flush() {
	if len(w.out) > 0 {
		w.rawFlush()
	}
}

func (w *writeBuffer) rawFlush() {
	if w.c != nil {
		if _, err := w.c.Write(w.out); err != nil {
			for _, i := range w.idx {
				w.p.putError(i, err)
			}

			w.c = nil
		}
	} else {
		for _, i := range w.idx {
			w.p.putError(i, errWriterBroken)
		}
	}

	w.out = w.out[:0]
	w.idx = w.idx[:0]
}
