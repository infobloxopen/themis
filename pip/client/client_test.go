package client

import (
	"errors"
	"fmt"
	"io"
	"math"
	"net"
	"os"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/infobloxopen/themis/pdp"
	"github.com/infobloxopen/themis/pip/server"
)

func TestNewClient(t *testing.T) {
	c := NewClient()
	assert.NotEmpty(t, c)
}

func TestNewClientWithTooSmallBuffer(t *testing.T) {
	c := NewClient(WithBufferSize(1024))
	if assert.NotEmpty(t, c) {
		if c, ok := c.(*client); assert.True(t, ok) {
			assert.Equal(t, defBufSize, c.opts.bufSize)
		}
	}
}

func TestClientConnect(t *testing.T) {
	var sErr error
	wg := new(sync.WaitGroup)
	defer func() {
		wg.Wait()
		if sErr != server.ErrNotBound {
			assert.NoError(t, sErr)
		}
	}()

	s := server.NewServer()
	if !assert.NoError(t, s.Bind()) {
		assert.FailNow(t, "failed to bind server")
	}
	defer func() {
		assert.NoError(t, s.Stop())
	}()
	wg.Add(1)
	go func() {
		defer wg.Done()
		sErr = s.Serve()
	}()

	c := NewClient()

	if assert.NoError(t, c.Connect()) {
		defer c.Close()

		assert.Equal(t, ErrConnected, c.Connect())
	}
}

func TestClientConnectRoundRobinBalancer(t *testing.T) {
	var s1Err error
	wg1 := new(sync.WaitGroup)
	defer func() {
		wg1.Wait()
		if s1Err != server.ErrNotBound {
			assert.NoError(t, s1Err)
		}
	}()

	s1 := server.NewServer(
		server.WithAddress("127.0.0.1:5601"),
	)
	if !assert.NoError(t, s1.Bind()) {
		assert.FailNow(t, "failed to bind server")
	}
	defer func() {
		assert.NoError(t, s1.Stop())
	}()
	wg1.Add(1)
	go func() {
		defer wg1.Done()
		s1Err = s1.Serve()
	}()

	var s2Err error
	wg2 := new(sync.WaitGroup)
	defer func() {
		wg2.Wait()
		if s2Err != server.ErrNotBound {
			assert.NoError(t, s2Err)
		}
	}()

	s2 := server.NewServer(
		server.WithAddress("127.0.0.1:5602"),
	)
	if !assert.NoError(t, s2.Bind()) {
		assert.FailNow(t, "failed to bind server")
	}
	defer func() {
		assert.NoError(t, s2.Stop())
	}()
	wg2.Add(1)
	go func() {
		defer wg2.Done()
		s2Err = s2.Serve()
	}()

	c := NewClient(
		WithRoundRobinBalancer(
			"127.0.0.1:5601",
			"127.0.0.1:5602",
		),
	)

	if assert.NoError(t, c.Connect()) {
		defer c.Close()
	}
}

func TestClientConnectRoundRobinBalancerInvalidAddress(t *testing.T) {
	var s1Err error
	wg1 := new(sync.WaitGroup)
	defer func() {
		wg1.Wait()
		if s1Err != server.ErrNotBound {
			assert.NoError(t, s1Err)
		}
	}()

	s1 := server.NewServer(
		server.WithAddress("127.0.0.1:5601"),
	)
	if !assert.NoError(t, s1.Bind()) {
		assert.FailNow(t, "failed to bind server")
	}
	defer func() {
		assert.NoError(t, s1.Stop())
	}()
	wg1.Add(1)
	go func() {
		defer wg1.Done()
		s1Err = s1.Serve()
	}()

	var s2Err error
	wg2 := new(sync.WaitGroup)
	defer func() {
		wg2.Wait()
		if s2Err != server.ErrNotBound {
			assert.NoError(t, s2Err)
		}
	}()

	s2 := server.NewServer(
		server.WithAddress("127.0.0.1:5602"),
	)
	if !assert.NoError(t, s2.Bind()) {
		assert.FailNow(t, "failed to bind server")
	}
	defer func() {
		assert.NoError(t, s2.Stop())
	}()
	wg2.Add(1)
	go func() {
		defer wg2.Done()
		s2Err = s2.Serve()
	}()

	c := NewClient(
		WithAddress("/dev/null"),
		WithRoundRobinBalancer(),
	)

	if !assert.Error(t, c.Connect()) {
		defer c.Close()
	}
}

func TestClientConnectNoServer(t *testing.T) {
	var cErr error
	wg := new(sync.WaitGroup)
	wg.Add(1)
	c := NewClient(
		WithConnTimeout(time.Millisecond),
		WithConnErrHandler(func(addr net.Addr, err error) {
			cErr = err
			wg.Done()
		}),
	).(*client)
	c.opts.connAttemptTimeout = time.Microsecond

	assert.NoError(t, c.Connect())
	defer c.Close()

	wg.Wait()
	if assert.Error(t, cErr) {
		assert.IsType(t, (*net.OpError)(nil), cErr)
	}
}

func TestClientClose(t *testing.T) {
	var sErr error
	wg := new(sync.WaitGroup)
	defer func() {
		wg.Wait()
		if sErr != server.ErrNotBound {
			assert.NoError(t, sErr)
		}
	}()

	s := server.NewServer()
	if !assert.NoError(t, s.Bind()) {
		assert.FailNow(t, "failed to bind server")
	}
	defer func() {
		assert.NoError(t, s.Stop())
	}()
	wg.Add(1)
	go func() {
		defer wg.Done()
		sErr = s.Serve()
	}()

	c := NewClient()

	if assert.NoError(t, c.Connect()) {
		c.Close()
		if cc, ok := c.(*client); assert.True(t, ok) {
			assert.Equal(t, pipClientIdle, atomic.LoadUint32(cc.state))
		}

		c.Close()
		if cc, ok := c.(*client); assert.True(t, ok) {
			assert.Equal(t, pipClientIdle, atomic.LoadUint32(cc.state))
		}
	}
}

func TestClientCloseWithDifferentRequests(t *testing.T) {
	sc, err := net.Listen(defNet, defAddr)
	if !assert.NoError(t, err) {
		assert.FailNow(t, "failed to start server listener")
	}
	defer sc.Close()
	go func() {
		for {
			c, err := sc.Accept()
			if err != nil {
				if err, ok := err.(*net.OpError); ok && err.Err.Error() == "use of closed network connection" {
					return
				}

				panic(fmt.Errorf("failed to listen at %s: %s", sc.Addr(), err))
			}

			go func(c net.Conn) {
				defer c.Close()

				var b [10240]byte
				for {
					_, err := c.Read(b[:])
					if err == io.EOF {
						return
					}

					if err != nil {
						panic(fmt.Errorf("failed to read from %s: %s", c.RemoteAddr(), err))
					}
				}
			}(c)
		}
	}()

	c := NewClient(
		WithMaxRequestSize(16),
		WithMaxQueue(1),
		WithBufferSize(24),
	)

	if !assert.NoError(t, c.Connect()) {
		assert.FailNow(t, "failed to connect to server")
	}

	wg := new(sync.WaitGroup)
	wg.Add(1)
	var err1 error
	go func() {
		defer wg.Done()
		_, err1 = c.Get(pdp.MakeStringAssignment("test", "test"))
	}()

	time.Sleep(50 * time.Millisecond)

	wg.Add(1)
	var err2 error
	go func() {
		defer wg.Done()
		_, err2 = c.Get(pdp.MakeStringAssignment("a", "a"))
	}()

	time.Sleep(50 * time.Millisecond)

	wg.Add(1)
	var err3 error
	go func() {
		defer wg.Done()
		_, err3 = c.Get(pdp.MakeStringAssignment("b", "b"))
	}()

	time.Sleep(50 * time.Millisecond)

	go c.Close()
	wg.Wait()

	assert.Equal(t, errReaderBroken, err1)
	assert.Error(t, err2)
	assert.Equal(t, errWriterBroken, err3)
}

func TestClientGet(t *testing.T) {
	var sErr error
	wg := new(sync.WaitGroup)
	defer func() {
		wg.Wait()
		if sErr != server.ErrNotBound {
			assert.NoError(t, sErr)
		}
	}()

	s := server.NewServer()
	if !assert.NoError(t, s.Bind()) {
		assert.FailNow(t, "failed to bind server")
	}
	defer func() {
		assert.NoError(t, s.Stop())
	}()
	wg.Add(1)
	go func() {
		defer wg.Done()
		sErr = s.Serve()
	}()

	c := NewClient()
	if err := c.Connect(); assert.NoError(t, err) {
		defer c.Close()

		v, err := c.Get()
		assert.Equal(t, pdp.UndefinedValue, v)
		assert.NoError(t, err)
	}
}

func TestClientGetErrNotConnected(t *testing.T) {
	c := NewClient()
	_, err := c.Get()
	assert.Equal(t, ErrNotConnected, err)
}

func TestClientGetMarshallingError(t *testing.T) {
	var sErr error
	wg := new(sync.WaitGroup)
	defer func() {
		wg.Wait()
		if sErr != server.ErrNotBound {
			assert.NoError(t, sErr)
		}
	}()

	s := server.NewServer()
	if !assert.NoError(t, s.Bind()) {
		assert.FailNow(t, "failed to bind server")
	}
	defer func() {
		assert.NoError(t, s.Stop())
	}()
	wg.Add(1)
	go func() {
		defer wg.Done()
		sErr = s.Serve()
	}()

	c := NewClient()
	if err := c.Connect(); assert.NoError(t, err) {
		defer c.Close()

		_, err := c.Get(pdp.MakeExpressionAssignment("test", pdp.UndefinedValue))
		assert.Error(t, err)
	}
}

func TestClientNextId(t *testing.T) {
	c := NewClient().(*client)

	i := c.nextID()
	assert.NotZero(t, i)

	j := c.nextID()
	assert.NotEqual(t, i, j)
}

func TestClientNextIdOverflow(t *testing.T) {
	c := NewClient().(*client)
	*c.autoID = math.MaxUint64
	assert.Zero(t, c.nextID())
	assert.Zero(t, c.nextID())
}

func TestClientDial(t *testing.T) {
	var sErr error
	wg := new(sync.WaitGroup)
	defer func() {
		wg.Wait()
		if sErr != server.ErrNotBound {
			assert.NoError(t, sErr)
		}
	}()

	s := server.NewServer()
	if !assert.NoError(t, s.Bind()) {
		assert.FailNow(t, "failed to bind server")
	}
	defer func() {
		assert.NoError(t, s.Stop())
	}()
	wg.Add(1)
	go func() {
		defer wg.Done()
		sErr = s.Serve()
	}()

	c := NewClient().(*client)
	atomic.StoreUint32(c.state, pipClientConnected)

	n := c.dial("127.0.0.1:5600")
	if assert.NotZero(t, n) {
		assert.NoError(t, n.Close())
	}

	c = NewClient(
		WithConnTimeout(0),
	).(*client)
	atomic.StoreUint32(c.state, pipClientConnected)

	n = c.dial("127.0.0.1:5600")
	if assert.NotZero(t, n) {
		assert.NoError(t, n.Close())
	}
}

func TestClientDialIter(t *testing.T) {
	c := NewClient().(*client)
	c.d = makeTestClientDialer(nil)

	n := c.dialIter("127.0.0.1:5600")
	assert.Zero(t, n)

	atomic.StoreUint32(c.state, pipClientConnected)
	n = c.dialIter("127.0.0.1:5600")
	assert.IsType(t, testClientConn{}, n)
}

func TestClientDialIterErrorTimeout(t *testing.T) {
	addrs := []net.Addr{}
	errs := []error{}
	c := NewClient(
		WithConnErrHandler(func(addr net.Addr, err error) {
			addrs = append(addrs, addr)
			errs = append(errs, err)
		}),
	).(*client)
	c.d = makeTestClientDialer(newTestNetTimeoutError())
	c.opts.connAttemptTimeout = time.Microsecond
	atomic.StoreUint32(c.state, pipClientConnected)

	wg := new(sync.WaitGroup)
	wg.Add(1)
	var n net.Conn
	go func() {
		defer wg.Done()

		n = c.dialIter("127.0.0.1:5600")
	}()

	time.Sleep(time.Millisecond)
	atomic.StoreUint32(c.state, pipClientIdle)
	wg.Wait()

	assert.Zero(t, n)
	assert.Equal(t, []net.Addr{}, addrs)
	assert.Equal(t, []error{}, errs)
}

func TestClientDialIterErrorRefused(t *testing.T) {
	addrs := []net.Addr{}
	errs := []error{}
	c := NewClient(
		WithConnErrHandler(func(addr net.Addr, err error) {
			addrs = append(addrs, addr)
			errs = append(errs, err)
		}),
	).(*client)
	c.d = makeTestClientDialer(newTestNetRefusedError())
	c.opts.connAttemptTimeout = time.Microsecond
	atomic.StoreUint32(c.state, pipClientConnected)

	wg := new(sync.WaitGroup)
	wg.Add(1)
	var n net.Conn
	go func() {
		defer wg.Done()

		n = c.dialIter("127.0.0.1:5600")
	}()

	time.Sleep(time.Millisecond)
	atomic.StoreUint32(c.state, pipClientIdle)
	wg.Wait()

	assert.Zero(t, n)
	assert.Equal(t, []net.Addr{}, addrs)
	assert.Equal(t, []error{}, errs)
}

func TestClientDialIterError(t *testing.T) {
	tErr := errors.New("test")
	addrs := []net.Addr{}
	errs := []error{}
	c := NewClient(
		WithConnErrHandler(func(addr net.Addr, err error) {
			addrs = append(addrs, addr)
			errs = append(errs, err)
		}),
	).(*client)
	c.d = makeTestClientDialer(tErr)
	atomic.StoreUint32(c.state, pipClientConnected)

	n := c.dialIter("127.0.0.1:5600")
	assert.Zero(t, n)
	assert.Equal(t, []net.Addr{nil}, addrs)
	assert.Equal(t, []error{tErr}, errs)
}

func TestClientDialIterTimeout(t *testing.T) {
	c := NewClient().(*client)
	c.d = makeTestClientDialer(nil)

	n := c.dialIterTimeout("127.0.0.1:5600")
	assert.Zero(t, n)

	atomic.StoreUint32(c.state, pipClientConnected)
	n = c.dialIterTimeout("127.0.0.1:5600")
	assert.IsType(t, testClientConn{}, n)
}

func TestClientDialIterTimeoutErrorTimeout(t *testing.T) {
	addrs := []net.Addr{}
	errs := []error{}
	c := NewClient(
		WithConnTimeout(time.Millisecond),
		WithConnErrHandler(func(addr net.Addr, err error) {
			addrs = append(addrs, addr)
			errs = append(errs, err)
		}),
	).(*client)
	tErr := newTestNetTimeoutError()
	c.d = makeTestClientDialer(tErr)
	c.opts.connAttemptTimeout = time.Microsecond
	atomic.StoreUint32(c.state, pipClientConnected)

	wg := new(sync.WaitGroup)
	wg.Add(1)
	var n net.Conn
	go func() {
		defer wg.Done()

		n = c.dialIterTimeout("127.0.0.1:5600")
	}()
	wg.Wait()

	assert.Zero(t, n)
	assert.Equal(t, []net.Addr{nil}, addrs)
	assert.Equal(t, []error{tErr}, errs)
}

func TestClientDialIterTimeoutErrorRefused(t *testing.T) {
	addrs := []net.Addr{}
	errs := []error{}
	c := NewClient(
		WithConnTimeout(time.Millisecond),
		WithConnErrHandler(func(addr net.Addr, err error) {
			addrs = append(addrs, addr)
			errs = append(errs, err)
		}),
	).(*client)
	tErr := newTestNetRefusedError()
	c.d = makeTestClientDialer(tErr)
	c.opts.connAttemptTimeout = time.Microsecond
	atomic.StoreUint32(c.state, pipClientConnected)

	wg := new(sync.WaitGroup)
	wg.Add(1)
	var n net.Conn
	go func() {
		defer wg.Done()

		n = c.dialIterTimeout("127.0.0.1:5600")
	}()
	wg.Wait()

	assert.Zero(t, n)
	assert.Equal(t, []net.Addr{nil}, addrs)
	assert.Equal(t, []error{tErr}, errs)
}

func TestClientDialIterTimeoutError(t *testing.T) {
	tErr := errors.New("test")
	addrs := []net.Addr{}
	errs := []error{}
	c := NewClient(
		WithConnErrHandler(func(addr net.Addr, err error) {
			addrs = append(addrs, addr)
			errs = append(errs, err)
		}),
	).(*client)
	c.d = makeTestClientDialer(tErr)
	atomic.StoreUint32(c.state, pipClientConnected)

	n := c.dialIterTimeout("127.0.0.1:5600")
	assert.Zero(t, n)
	assert.Equal(t, []net.Addr{nil}, addrs)
	assert.Equal(t, []error{tErr}, errs)
}

type testClientDialer struct {
	err error
}

func makeTestClientDialer(err error) testClientDialer {
	return testClientDialer{
		err: err,
	}
}

func (d testClientDialer) dial(a string) (net.Conn, error) {
	if d.err != nil {
		return nil, d.err
	}

	return testClientConn{}, nil
}

type testClientConn struct{}

func (c testClientConn) Write(b []byte) (int, error)        { panic("not implemented") }
func (c testClientConn) Read(b []byte) (int, error)         { panic("not implemented") }
func (c testClientConn) Close() error                       { panic("not implemented") }
func (c testClientConn) LocalAddr() net.Addr                { panic("not implemented") }
func (c testClientConn) RemoteAddr() net.Addr               { panic("not implemented") }
func (c testClientConn) SetDeadline(t time.Time) error      { panic("not implemented") }
func (c testClientConn) SetReadDeadline(t time.Time) error  { panic("not implemented") }
func (c testClientConn) SetWriteDeadline(t time.Time) error { panic("not implemented") }

func newTestNetTimeoutError() error {
	return &net.OpError{
		Err: testTimeoutError{},
	}
}

func newTestNetRefusedError() error {
	return &net.OpError{
		Err: os.NewSyscallError("test", errors.New(netConnRefusedMsg)),
	}
}

type testTimeoutError struct{}

func (err testTimeoutError) Error() string {
	return "test timeout"
}

func (err testTimeoutError) Timeout() bool {
	return true
}
