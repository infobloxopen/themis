package server

import (
	"errors"
	"io"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestHandle(t *testing.T) {
	s := NewServer()

	errs := []error{}
	cc := connWithErrHandler{
		c: makeConnTestErrOnCloseConn(nil),
		h: func(a net.Addr, err error) {
			if err != nil {
				errs = append(errs, err)
			}
		},
	}

	idx := s.conns.put(cc)
	if idx < 0 {
		assert.FailNow(t, "failed to store connection")
	}

	wg := new(sync.WaitGroup)
	wg.Add(1)
	s.handle(wg, cc, idx)
	wg.Wait()
	assert.Equal(t, []error{}, errs)
	assert.Equal(t, 0, len(s.conns.c))
}

func TestHandleError(t *testing.T) {
	s := NewServer()

	errs := []error{}
	err := errors.New("test error")
	cc := connWithErrHandler{
		c: makeConnTestErrOnCloseConn(err),
		h: func(a net.Addr, err error) {
			if err != nil {
				errs = append(errs, err)
			}
		},
	}

	idx := s.conns.put(cc)
	if idx < 0 {
		assert.FailNow(t, "failed to store connection")
	}

	wg := new(sync.WaitGroup)
	wg.Add(1)
	s.handle(wg, cc, idx)
	wg.Wait()
	assert.Equal(t, []error{err}, errs)
	assert.Equal(t, 0, len(s.conns.c))
}

func TestNewConnReg(t *testing.T) {
	c := newConnReg(0)
	assert.NotEqual(t, map[int]connWithErrHandler(nil), c.c)
	assert.Equal(t, 0, c.m)

	c = newConnReg(100)
	assert.NotEqual(t, map[int]connWithErrHandler(nil), c.c)
	assert.Equal(t, 100, c.m)
}

func TestConnRegPut(t *testing.T) {
	c := newConnReg(3)

	idx := c.put(connWithErrHandler{
		c: makeConnTestErrOnCloseConn(nil),
	})
	assert.Equal(t, 0, idx)

	idx = c.put(connWithErrHandler{
		c: makeConnTestErrOnCloseConn(nil),
	})
	assert.Equal(t, 1, idx)

	idx = c.put(connWithErrHandler{
		c: makeConnTestErrOnCloseConn(nil),
	})
	assert.Equal(t, 2, idx)

	idx = c.put(connWithErrHandler{
		c: makeConnTestErrOnCloseConn(nil),
	})
	assert.True(t, idx < 0)
}

func TestConnRegDel(t *testing.T) {
	c := newConnReg(0)

	idx := c.put(connWithErrHandler{
		c: makeConnTestErrOnCloseConn(nil),
	})
	assert.Equal(t, 0, idx)
	assert.Equal(t, 1, len(c.c))

	idx = c.put(connWithErrHandler{
		c: makeConnTestErrOnCloseConn(nil),
	})
	assert.Equal(t, 1, idx)
	assert.Equal(t, 2, len(c.c))

	c.del(idx)
	assert.Equal(t, 1, len(c.c))

	c.del(idx)
	assert.Equal(t, 1, len(c.c))
}

func TestConnRegDelAll(t *testing.T) {
	c := newConnReg(3)

	c.put(connWithErrHandler{
		c: makeConnTestErrOnCloseConn(nil),
	})
	c.put(connWithErrHandler{
		c: makeConnTestErrOnCloseConn(nil),
	})
	c.put(connWithErrHandler{
		c: makeConnTestErrOnCloseConn(nil),
	})
	assert.Equal(t, 3, len(c.c))

	c.delAll()
	assert.Equal(t, 0, len(c.c))
}

func TestConnWithErrHandlerHandle(t *testing.T) {
	errs := []error{}
	cc := connWithErrHandler{
		c: makeConnTestErrOnCloseConn(nil),
		h: func(a net.Addr, err error) {
			if err != nil {
				errs = append(errs, err)
			}
		},
	}

	err := errors.New("test")
	cc.handle(err)
	assert.Equal(t, []error{err}, errs)
}

func TestConnWithErrHandlerClose(t *testing.T) {
	errs := []error{}
	err := errors.New("test")
	cc := connWithErrHandler{
		c: makeConnTestErrOnCloseConn(err),
		h: func(a net.Addr, err error) {
			if err != nil {
				errs = append(errs, err)
			}
		},
	}

	cc.close()
	assert.Equal(t, []error{err}, errs)
}

type connTestErrOnCloseConn struct {
	err error
}

func makeConnTestErrOnCloseConn(err error) net.Conn {
	return connTestErrOnCloseConn{
		err: err,
	}
}

func (c connTestErrOnCloseConn) Close() error { return c.err }
func (c connTestErrOnCloseConn) Read(b []byte) (n int, err error) {
	return 0, io.EOF
}

func (c connTestErrOnCloseConn) RemoteAddr() net.Addr {
	addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
	if err != nil {
		panic(err)
	}

	return addr
}

func (c connTestErrOnCloseConn) Write(b []byte) (n int, err error)  { panic("not implemented") }
func (c connTestErrOnCloseConn) LocalAddr() net.Addr                { panic("not implemented") }
func (c connTestErrOnCloseConn) SetDeadline(t time.Time) error      { panic("not implemented") }
func (c connTestErrOnCloseConn) SetReadDeadline(t time.Time) error  { panic("not implemented") }
func (c connTestErrOnCloseConn) SetWriteDeadline(t time.Time) error { panic("not implemented") }
