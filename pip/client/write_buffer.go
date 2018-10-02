package client

import "encoding/binary"

const (
	msgSizeBytes = 4
	msgIdxBytes  = 4
)

type writeBuffer struct {
	out []byte
	idx []int
	p   pipes
}

func newWriteBuffer(n int, p pipes) *writeBuffer {
	return &writeBuffer{
		out: make([]byte, 0, n),
		idx: make([]int, n/(msgSizeBytes+msgIdxBytes)),
		p:   p,
	}
}

func (w *writeBuffer) isEmpty() bool {
	return len(w.out) <= 0
}

func (w *writeBuffer) rem() int {
	return cap(w.out) - len(w.out)
}

func (w *writeBuffer) put(r request) {
	size := msgIdxBytes + len(r.b)
	if w.rem() < msgSizeBytes+size {
		w.flush()
	}

	i := len(w.out)
	w.out = append(w.out, 0, 0, 0, 0, 0, 0, 0, 0)
	binary.LittleEndian.PutUint32(w.out[i:], uint32(size))
	binary.LittleEndian.PutUint32(w.out[i+msgSizeBytes:], uint32(r.i))
	w.out = append(w.out, r.b...)
	w.idx = append(w.idx, r.i)

	if w.rem() <= 0 {
		w.flush()
	}
}

func (w *writeBuffer) flush() {
	for _, i := range w.idx {
		w.p.putBytes(i, nil)
	}

	w.out = w.out[:0]
	w.idx = w.idx[:0]
}
