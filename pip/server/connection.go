package server

import (
	"net"
	"sync"
)

func (s *Server) handle(wg *sync.WaitGroup, c connWithErrHandler, idx int) {
	defer func() {
		wg.Done()
		s.conns.del(idx)
	}()

	msgs := makePool(1, s.opts.maxMsgSize)
	write(c, read(c, msgs, s.opts.bufSize), msgs, s.opts.bufSize, s.opts.writeInt)
}

type connReg struct {
	sync.Mutex

	m int
	i int
	c map[int]connWithErrHandler
}

func newConnReg(max int) *connReg {
	if max > 0 {
		return &connReg{
			m: max,
			c: make(map[int]connWithErrHandler, max),
		}
	}

	return &connReg{
		c: make(map[int]connWithErrHandler),
	}
}

func (r *connReg) put(c connWithErrHandler) int {
	r.Lock()
	defer r.Unlock()

	if r.m > 0 && r.i >= r.m {
		return -1
	}

	i := r.i
	r.i++

	r.c[i] = c
	return i
}

func (r *connReg) del(i int) {
	r.Lock()
	defer r.Unlock()

	if c, ok := r.c[i]; ok {
		delete(r.c, i)
		c.close()
	}
}

func (r *connReg) delAll() {
	r.Lock()
	defer r.Unlock()

	for i, c := range r.c {
		delete(r.c, i)
		c.close()
	}
}

type connWithErrHandler struct {
	c net.Conn
	h ConnErrHandler
}

func (c connWithErrHandler) handle(err error) {
	if c.h != nil && err != nil {
		c.h(c.c.RemoteAddr(), err)
	}
}

func (c connWithErrHandler) close() {
	c.handle(c.c.Close())
}
