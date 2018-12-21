package client

import (
	"math"
	"sync/atomic"
)

type pipe struct {
	t  *int64
	ch chan response
}

func makePipe() pipe {
	t := new(int64)
	*t = math.MinInt64

	return pipe{
		t:  t,
		ch: make(chan response, 1),
	}
}

func (p pipe) clean(t *int64) {
	for {
		select {
		default:
			atomic.StoreInt64(p.t, atomic.LoadInt64(t))
			return

		case <-p.ch:
		}
	}
}

func (p pipe) get() (*byteBuffer, error) {
	r := <-p.ch
	return r.b, r.err
}

func (p pipe) putError(err error) {
	select {
	default:
	case p.ch <- response{err: err}:
	}

	atomic.StoreInt64(p.t, math.MinInt64)
}

func (p pipe) putBytes(b *byteBuffer) (ok bool) {
	select {
	default:
	case p.ch <- response{b: b}:
		ok = true
	}

	atomic.StoreInt64(p.t, math.MinInt64)
	return
}
