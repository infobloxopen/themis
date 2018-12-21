package client

import (
	"encoding/binary"
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

func TestClientConnect(t *testing.T) {
	s := newTestServerForClient(t,
		server.WithHandler(testServerForClientHandler),
	)
	defer s.stop(t)

	c := NewClient()
	if assert.NoError(t, c.Connect()) {
		defer c.Close()

		assert.Equal(t, ErrConnected, c.Connect())
	}
}

func TestClientConnectRoundRobinBalancer(t *testing.T) {
	s1 := newTestServerForClient(t,
		server.WithAddress("127.0.0.1:5601"),
		server.WithHandler(testServerForClientHandler),
	)
	defer s1.stop(t)

	s2 := newTestServerForClient(t,
		server.WithAddress("127.0.0.1:5602"),
		server.WithHandler(testServerForClientHandler),
	)
	defer s2.stop(t)

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
	s1 := newTestServerForClient(t,
		server.WithAddress("127.0.0.1:5601"),
		server.WithHandler(testServerForClientHandler),
	)
	defer s1.stop(t)

	s2 := newTestServerForClient(t,
		server.WithAddress("127.0.0.1:5602"),
		server.WithHandler(testServerForClientHandler),
	)
	defer s2.stop(t)

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
	s := newTestServerForClient(t,
		server.WithHandler(testServerForClientHandler),
	)
	defer s.stop(t)

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

func TestClientCloseWithCache(t *testing.T) {
	s := newTestServerForClient(t,
		server.WithHandler(testServerForClientHandler),
	)
	defer s.stop(t)

	c := NewClient(
		WithCacheTTL(time.Minute),
	)

	if assert.NoError(t, c.Connect()) {
		if cc, ok := c.(*client); assert.True(t, ok) {
			assert.NotZero(t, cc.cache)
		}

		c.Close()
		if cc, ok := c.(*client); assert.True(t, ok) {
			assert.Equal(t, pipClientIdle, atomic.LoadUint32(cc.state))
		}

		if cc, ok := c.(*client); assert.True(t, ok) {
			assert.Zero(t, cc.cache)
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
		WithMaxRequestSize(22),
		WithMaxQueue(1),
		WithBufferSize(30),
	).(*client)

	if !assert.NoError(t, c.Connect()) {
		assert.FailNow(t, "failed to connect to server")
	}

	wg := new(sync.WaitGroup)
	wg.Add(1)
	var err1 error
	go func() {
		defer wg.Done()
		_, _, err1 = c.tryGet("test", []pdp.AttributeValue{
			pdp.MakeStringValue("test"),
		})
	}()

	time.Sleep(50 * time.Millisecond)

	wg.Add(1)
	var err2 error
	go func() {
		defer wg.Done()
		_, _, err2 = c.tryGet("test", []pdp.AttributeValue{
			pdp.MakeStringValue("a"),
		})
	}()

	time.Sleep(50 * time.Millisecond)

	wg.Add(1)
	var err3 error
	go func() {
		defer wg.Done()
		_, _, err3 = c.tryGet("test", []pdp.AttributeValue{
			pdp.MakeStringValue("b"),
		})
	}()

	time.Sleep(50 * time.Millisecond)

	go c.Close()
	wg.Wait()

	assert.Equal(t, errReaderBroken, err1)
	assert.Error(t, err2)
	assert.Equal(t, errWriterBroken, err3)
}

func TestClientGet(t *testing.T) {
	s := newTestServerForClient(t,
		server.WithHandler(testServerForClientHandler),
	)
	defer s.stop(t)

	c := NewClient()
	if err := c.Connect(); assert.NoError(t, err) {
		defer c.Close()

		v, err := c.Get("test", []pdp.AttributeValue{pdp.MakeStringValue("test")})
		assert.Equal(t, pdp.MakeStringValue("test"), v)
		assert.NoError(t, err)
	}
}

func TestClientGetErrNotConnected(t *testing.T) {
	c := NewClient()
	_, err := c.Get("test", nil)
	assert.Equal(t, ErrNotConnected, err)
}

func TestClientTryGet(t *testing.T) {
	s := newTestServerForClient(t,
		server.WithHandler(testServerForClientHandler),
	)
	defer s.stop(t)

	c := NewClient().(*client)
	if err := c.Connect(); assert.NoError(t, err) {
		defer c.Close()

		v, ok, err := c.tryGet("test", []pdp.AttributeValue{pdp.MakeStringValue("test")})
		assert.Equal(t, pdp.MakeStringValue("test"), v)
		assert.False(t, ok)
		assert.NoError(t, err)
	}
}

func TestClientTryGetWithCache(t *testing.T) {
	s := newTestServerForClient(t,
		server.WithHandler(testServerForClientHandler),
	)
	defer s.stop(t)

	hits := 0
	c := NewClient(
		WithCacheTTL(time.Minute),
		WithCacheHitHandler(func(string, []pdp.AttributeValue, pdp.AttributeValue, error) {
			hits++
		}),
	).(*client)
	if err := c.Connect(); assert.NoError(t, err) {
		defer c.Close()

		assert.Zero(t, c.cache.Len())

		v, ok, err := c.tryGet("test", []pdp.AttributeValue{pdp.MakeStringValue("test")})
		assert.Equal(t, pdp.MakeStringValue("test"), v)
		assert.False(t, ok)
		assert.NoError(t, err)
		assert.Equal(t, 1, c.cache.Len())
		assert.Zero(t, hits)

		v, ok, err = c.tryGet("test", []pdp.AttributeValue{pdp.MakeStringValue("test")})
		assert.Equal(t, pdp.MakeStringValue("test"), v)
		assert.False(t, ok)
		assert.NoError(t, err)
		assert.Equal(t, 1, c.cache.Len())
		assert.Equal(t, 1, hits)
	}
}

func TestClientTryGettErrNotConnected(t *testing.T) {
	c := NewClient().(*client)
	_, ok, err := c.tryGet("test", nil)
	assert.False(t, ok)
	assert.Equal(t, ErrNotConnected, err)
}

func TestClientTryGetMarshallingError(t *testing.T) {
	s := newTestServerForClient(t,
		server.WithHandler(testServerForClientHandler),
	)
	defer s.stop(t)

	c := NewClient().(*client)
	if err := c.Connect(); assert.NoError(t, err) {
		defer c.Close()

		_, ok, err := c.tryGet("test", []pdp.AttributeValue{
			pdp.UndefinedValue,
		})
		assert.False(t, ok)
		assert.Error(t, err)
	}
}

func TestClientTryGetResponseServerError(t *testing.T) {
	s := newTestServerForClient(t,
		server.WithHandler(constTestServerForClientHandler(
			1, 0, 4, 0, 't', 'e', 's', 't',
		)),
	)
	defer s.stop(t)

	c := NewClient().(*client)
	if err := c.Connect(); assert.NoError(t, err) {
		defer c.Close()

		_, ok, err := c.tryGet("test", []pdp.AttributeValue{pdp.MakeStringValue("test")})
		assert.False(t, ok)
		if assert.Error(t, err) {
			assert.IsType(t, &pdp.ResponseServerError{}, err)
		}
	}
}

func TestClientTryGetBadResponse(t *testing.T) {
	s := newTestServerForClient(t,
		server.WithHandler(constTestServerForClientHandler(
			2, 0,
		)),
	)
	defer s.stop(t)

	c := NewClient().(*client)
	if err := c.Connect(); assert.NoError(t, err) {
		defer c.Close()

		_, ok, err := c.tryGet("test", []pdp.AttributeValue{pdp.MakeStringValue("test")})
		assert.True(t, ok)
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
	s := newTestServerForClient(t,
		server.WithHandler(testServerForClientHandler),
	)
	defer s.stop(t)

	c := NewClient().(*client)
	atomic.StoreUint32(c.state, pipClientConnected)

	n := c.dial("127.0.0.1:5600", make(chan struct{}))
	if assert.NotZero(t, n) {
		assert.NoError(t, n.Close())
	}

	c = NewClient(
		WithConnTimeout(0),
	).(*client)
	atomic.StoreUint32(c.state, pipClientConnected)

	n = c.dial("127.0.0.1:5600", make(chan struct{}))
	if assert.NotZero(t, n) {
		assert.NoError(t, n.Close())
	}
}

func TestClientDialIter(t *testing.T) {
	c := NewClient().(*client)
	c.d = makeTestClientDialer(nil)

	n := c.dialIter("127.0.0.1:5600", make(chan struct{}))
	assert.Zero(t, n)

	atomic.StoreUint32(c.state, pipClientConnected)
	n = c.dialIter("127.0.0.1:5600", make(chan struct{}))
	assert.IsType(t, testClientConn{}, n)

	c.opts.connAttemptTimeout = time.Millisecond
	c.d = makeTestClientDialer(newTestNetRefusedError())
	ch := make(chan struct{})
	wg := new(sync.WaitGroup)
	wg.Add(1)
	go func() {
		defer wg.Done()

		n = c.dialIter("127.0.0.1:5600", ch)
	}()

	time.Sleep(time.Millisecond)
	close(ch)
	wg.Wait()

	assert.Zero(t, n)
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

		n = c.dialIter("127.0.0.1:5600", make(chan struct{}))
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

		n = c.dialIter("127.0.0.1:5600", make(chan struct{}))
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

	n := c.dialIter("127.0.0.1:5600", make(chan struct{}))
	assert.Zero(t, n)
	assert.Equal(t, []net.Addr{nil}, addrs)
	assert.Equal(t, []error{tErr}, errs)
}

func TestClientDialIterTimeout(t *testing.T) {
	c := NewClient().(*client)
	c.d = makeTestClientDialer(nil)

	n := c.dialIterTimeout("127.0.0.1:5600", make(chan struct{}))
	assert.Zero(t, n)

	atomic.StoreUint32(c.state, pipClientConnected)
	n = c.dialIterTimeout("127.0.0.1:5600", make(chan struct{}))
	assert.IsType(t, testClientConn{}, n)

	c.opts.connAttemptTimeout = time.Millisecond
	c.d = makeTestClientDialer(newTestNetRefusedError())
	ch := make(chan struct{})
	wg := new(sync.WaitGroup)
	wg.Add(1)
	go func() {
		defer wg.Done()

		n = c.dialIterTimeout("127.0.0.1:5600", ch)
	}()

	time.Sleep(time.Millisecond)
	close(ch)
	wg.Wait()

	assert.Zero(t, n)
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

		n = c.dialIterTimeout("127.0.0.1:5600", make(chan struct{}))
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

		n = c.dialIterTimeout("127.0.0.1:5600", make(chan struct{}))
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

	n := c.dialIterTimeout("127.0.0.1:5600", make(chan struct{}))
	assert.Zero(t, n)
	assert.Equal(t, []net.Addr{nil}, addrs)
	assert.Equal(t, []error{tErr}, errs)
}

type testServerForClient struct {
	sync.WaitGroup
	s   *server.Server
	err error
}

var testServerForClientInfoResponseHeader = []byte{1, 0, 0, 0, 2}

func newTestServerForClient(t *testing.T, opts ...server.Option) *testServerForClient {
	s := new(testServerForClient)

	s.s = server.NewServer(opts...)
	if !assert.NoError(t, s.s.Bind()) {
		assert.FailNow(t, "failed to bind server")
	}

	s.Add(1)
	go func() {
		defer s.Done()

		s.err = s.s.Serve()
	}()

	return s
}

func (s *testServerForClient) stop(t *testing.T) {
	if assert.NoError(t, s.s.Stop()) {
		s.Wait()

		if s.err != server.ErrNotBound {
			assert.NoError(t, s.err)
		}
	}
}

func testServerForClientHandler(b []byte) []byte {
	if len(b) < 4 {
		panic("too short input buffer")
	}

	in := testServerForClientHandlerCheckVersion(b[4:])
	in = testServerForClientHandlerSkipPath(in)
	in = testServerForClientHandlerCheckValuesNumber(in)
	in = testServerForClientHandlerCheckFirstValueType(in)
	in = testServerForClientHandlerGetFirstValueBytes(in)

	out := b
	m := 4
	m += copy(out[m:], testServerForClientInfoResponseHeader)
	binary.LittleEndian.PutUint16(out[m:], uint16(len(in)))
	m += 2
	m += copy(out[m:], in)

	return out[:m]
}

func testServerForClientHandlerCheckVersion(b []byte) []byte {
	if len(b) < 2 {
		panic("too short input buffer")
	}

	if v := binary.LittleEndian.Uint16(b); v != 1 {
		panic(fmt.Errorf("invalid information request version %d (expected %d)", v, 1))
	}

	return b[2:]
}

func testServerForClientHandlerSkipPath(b []byte) []byte {
	if len(b) < 2 {
		panic("too short input buffer")
	}

	n := int(binary.LittleEndian.Uint16(b))
	b = b[2:]

	if len(b) < n {
		panic("too short input buffer")
	}

	return b[n:]
}

func testServerForClientHandlerCheckValuesNumber(b []byte) []byte {
	if len(b) < 2 {
		panic("too short input buffer")
	}

	if c := int(binary.LittleEndian.Uint16(b)); c < 1 {
		panic(fmt.Errorf("expected at least one value but got %d", c))
	}

	return b[2:]
}

func testServerForClientHandlerCheckFirstValueType(b []byte) []byte {
	if len(b) < 1 {
		panic("too short input buffer")
	}

	if t := b[0]; t != 2 {
		panic(fmt.Errorf("expected value of type %d (string) but got %d", 2, t))
	}

	return b[1:]
}

func testServerForClientHandlerGetFirstValueBytes(b []byte) []byte {
	if len(b) < 2 {
		panic("too short input buffer")
	}

	n := int(binary.LittleEndian.Uint16(b))
	b = b[2:]

	if len(b) < n {
		panic(fmt.Errorf("too short input buffer %d < %d", len(b), n))
	}

	return b[:n]
}

func constTestServerForClientHandler(payload ...byte) func([]byte) []byte {
	return func(b []byte) []byte {
		if len(b) < 4 {
			panic(fmt.Errorf("too short input buffer %d (missing request id)", len(b)))
		}

		if len(b)-4 < len(payload) {
			panic(fmt.Errorf("too short input buffer %d (reply don't fit)", len(b)))
		}

		m := 4
		m += copy(b[m:], payload)

		return b[:m]
	}
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
