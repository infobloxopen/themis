package client

import (
	"errors"
	"math"
	"sync/atomic"
	"time"
)

var (
	errTimeout      = errors.New("timeout")
	errReaderBroken = errors.New("reader has been broken")
)

type pipes struct {
	t   *int64
	d   int64
	idx chan int
	p   []pipe
}

func makePipes(n int, d int64) pipes {
	idx := make(chan int, n)
	for len(idx) < cap(idx) {
		idx <- len(idx)
	}

	p := make([]pipe, n)
	for i := range p {
		p[i] = makePipe()
	}

	t := new(int64)
	*t = time.Now().UnixNano()

	return pipes{
		t:   t,
		d:   d,
		idx: idx,
		p:   p,
	}
}

func (p pipes) alloc() (int, pipe) {
	i := <-p.idx
	out := p.p[i]
	out.clean(p.t)
	return i, out
}

func (p pipes) free(i int) {
	p.idx <- i
}

func (p pipes) putError(i int, err error) {
	p.p[i].putError(err)
}

func (p pipes) putBytes(i int, b *byteBuffer) bool {
	return p.p[i].putBytes(b)
}

func (p pipes) check(t time.Time) bool {
	n := t.UnixNano()
	atomic.StoreInt64(p.t, n)

	d := p.d
	c := 0
	for _, p := range p.p {
		if pn := atomic.LoadInt64(p.t); pn > math.MinInt64 && n-pn > d {
			p.putError(errTimeout)
			c++
		}
	}

	return c > 0
}

func (p pipes) flush() {
	for _, p := range p.p {
		if pn := atomic.LoadInt64(p.t); pn > math.MinInt64 {
			p.putError(errReaderBroken)
		}
	}
}
