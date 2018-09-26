package server

import (
	"errors"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestWrite(t *testing.T) {
	errs := []error{}
	c := newWTestConn()
	cc := connWithErrHandler{
		c: c,
		h: func(a net.Addr, err error) {
			if err != nil {
				errs = append(errs, err)
			}
		},
	}

	p := makePool(2, 8)
	ch := make(chan []byte, 2)
	wg := new(sync.WaitGroup)

	wg.Add(1)

	ch <- append(p.get(), 0xef, 0xbe, 0xad, 0xde, 0x00, 0xde, 0xc0)
	go func() {
		defer wg.Done()
		write(cc, ch, p, defBufSize, defWriteInt)
	}()

	time.Sleep(1000 * defWriteInt)
	ch <- append(p.get(), 0xde, 0xc0, 0xad, 0x0b)
	close(ch)

	wg.Wait()

	assert.Equal(t, []error{}, errs)
	assert.Equal(t, [][]byte{
		{0x7, 0x0, 0x0, 0x0, 0xef, 0xbe, 0xad, 0xde, 0x0, 0xde, 0xc0},
		{0x4, 0x0, 0x0, 0x0, 0xde, 0xc0, 0xad, 0x0b},
	}, c.data)
}

func TestWriteSmallBuffer(t *testing.T) {
	errs := []error{}
	c := newWTestConn()
	cc := connWithErrHandler{
		c: c,
		h: func(a net.Addr, err error) {
			if err != nil {
				errs = append(errs, err)
			}
		},
	}

	p := makePool(2, 8)
	ch := make(chan []byte, 2)
	wg := new(sync.WaitGroup)

	wg.Add(1)

	ch <- append(p.get(), 0xef, 0xbe, 0xad, 0xde, 0x00, 0xde, 0xc0)
	go func() {
		defer wg.Done()
		write(cc, ch, p, 2, defWriteInt)
	}()

	time.Sleep(5 * defWriteInt)
	ch <- append(p.get(), 0xde, 0xc0, 0xad, 0x0b)
	close(ch)

	wg.Wait()

	assert.Equal(t, []error{}, errs)
	assert.Equal(t, [][]byte{
		{0x7, 0x0}, {0x0, 0x0}, {0xef, 0xbe}, {0xad, 0xde}, {0x0, 0xde}, {0xc0},
		{0x4, 0x0}, {0x0, 0x0}, {0xde, 0xc0}, {0xad, 0x0b},
	}, c.data)
}

func TestWriteErrorOnSize(t *testing.T) {
	errs := []error{}
	err := errors.New("test")
	c := newWTestConn(err)
	cc := connWithErrHandler{
		c: c,
		h: func(a net.Addr, err error) {
			if err != nil {
				errs = append(errs, err)
			}
		},
	}

	p := makePool(2, 8)
	ch := make(chan []byte, 2)
	wg := new(sync.WaitGroup)

	wg.Add(1)

	ch <- append(p.get(), 0xef, 0xbe, 0xad, 0xde)
	go func() {
		defer wg.Done()
		write(cc, ch, p, 4, defWriteInt)
	}()

	close(ch)

	wg.Wait()

	assert.Equal(t, []error{err}, errs)
	assert.Equal(t, [][]byte{{0x4, 0x0, 0x0, 0x0}}, c.data)
}

func TestWriteErrorOnData(t *testing.T) {
	errs := []error{}
	err := errors.New("test")
	c := newWTestConn(nil, err)
	cc := connWithErrHandler{
		c: c,
		h: func(a net.Addr, err error) {
			if err != nil {
				errs = append(errs, err)
			}
		},
	}

	p := makePool(2, 8)
	ch := make(chan []byte, 2)
	wg := new(sync.WaitGroup)

	wg.Add(1)

	ch <- append(p.get(), 0xef, 0xbe, 0xad, 0xde)
	go func() {
		defer wg.Done()
		write(cc, ch, p, 4, defWriteInt)
	}()

	close(ch)

	wg.Wait()

	assert.Equal(t, []error{err}, errs)
	assert.Equal(t, [][]byte{
		{0x4, 0x0, 0x0, 0x0},
		{0xef, 0xbe, 0xad, 0xde},
	}, c.data)
}

func TestWriteErrorOnWait(t *testing.T) {
	errs := []error{}
	err := errors.New("test")
	c := newWTestConn(nil, nil, err)
	cc := connWithErrHandler{
		c: c,
		h: func(a net.Addr, err error) {
			if err != nil {
				errs = append(errs, err)
			}
		},
	}

	p := makePool(2, 8)
	ch := make(chan []byte, 2)
	wg := new(sync.WaitGroup)

	wg.Add(1)

	ch <- append(p.get(), 0xef, 0xbe, 0xad, 0xde, 0xde, 0xc0)
	go func() {
		defer wg.Done()
		write(cc, ch, p, 4, defWriteInt)
	}()

	time.Sleep(5 * defWriteInt)
	close(ch)

	wg.Wait()

	assert.Equal(t, []error{err}, errs)
	assert.Equal(t, [][]byte{
		{0x6, 0x0, 0x0, 0x0},
		{0xef, 0xbe, 0xad, 0xde},
		{0xde, 0xc0},
	}, c.data)
}

func TestFlush(t *testing.T) {
	errs := []error{}
	c := newWTestConn()
	cc := connWithErrHandler{
		c: c,
		h: func(a net.Addr, err error) {
			if err != nil {
				errs = append(errs, err)
			}
		},
	}

	assert.True(t, flush(cc, []byte{0xef, 0xbe, 0xad, 0xde}))
	assert.Equal(t, []error{}, errs)
	assert.Equal(t, [][]byte{{0xef, 0xbe, 0xad, 0xde}}, c.data)
}

func TestFlushWriteError(t *testing.T) {
	errs := []error{}
	err := errors.New("test")
	c := newWTestConn(err)
	cc := connWithErrHandler{
		c: c,
		h: func(a net.Addr, err error) {
			if err != nil {
				errs = append(errs, err)
			}
		},
	}

	assert.False(t, flush(cc, []byte{0xef, 0xbe, 0xad, 0xde}))
	assert.Equal(t, []error{err}, errs)
	assert.Equal(t, [][]byte{{0xef, 0xbe, 0xad, 0xde}}, c.data)
}

func TestIgnore(t *testing.T) {
	p := makePool(3, 4)

	ch := make(chan []byte, 3)
	ch <- p.get()
	ch <- p.get()
	ch <- p.get()
	close(ch)

	ignore(ch, p)

	assert.Equal(t, 0, len(ch))
	assert.Equal(t, 3, len(p.ch))
}

type wTestConn struct {
	errs []error
	data [][]byte
}

func newWTestConn(errs ...error) *wTestConn {
	return &wTestConn{
		errs: errs,
		data: [][]byte{},
	}
}

func (c *wTestConn) Write(b []byte) (int, error) {
	c.data = append(c.data, append([]byte{}, b...))

	if len(c.errs) > 0 {
		err := c.errs[0]
		c.errs = c.errs[1:]

		if err != nil {
			return 0, err
		}
	}

	return len(b), nil
}

func (c *wTestConn) RemoteAddr() net.Addr {
	addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
	if err != nil {
		panic(err)
	}

	return addr
}

func (c *wTestConn) Close() error                       { panic("not implemented") }
func (c *wTestConn) Read(b []byte) (int, error)         { panic("not implemented") }
func (c *wTestConn) LocalAddr() net.Addr                { panic("not implemented") }
func (c *wTestConn) SetDeadline(t time.Time) error      { panic("not implemented") }
func (c *wTestConn) SetReadDeadline(t time.Time) error  { panic("not implemented") }
func (c *wTestConn) SetWriteDeadline(t time.Time) error { panic("not implemented") }
