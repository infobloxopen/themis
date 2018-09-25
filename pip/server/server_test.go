package server

import (
	"errors"
	"io"
	"io/ioutil"
	"net"
	"os"
	"path"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewServer(t *testing.T) {
	s := NewServer(
		WithNetwork("tcp4"),
		WithAddress("127.0.0.1:5555"),
	)

	assert.Equal(t, s.opts.net, "tcp4")
	assert.Equal(t, s.opts.addr, "127.0.0.1:5555")
}

func TestServerBind(t *testing.T) {
	s := NewServer(
		WithAddress("127.0.0.1:0"),
	)

	if assert.NoError(t, s.Bind()) {
		assert.NotEqual(t, nil, s.ln)
		assert.NoError(t, s.Stop())
	}
}

func TestServerBindUnix(t *testing.T) {
	dir, err := ioutil.TempDir("", "pip.server.test.")
	if err != nil {
		assert.FailNow(t, "failed to create temporary directory", "ioutil.TempDir error: %s", err)
	}
	defer func() {
		if err = os.RemoveAll(dir); err != nil {
			assert.Fail(t, "failed to cleanup temporary directory", "os.RemoveAll error: %s", err)
		}
	}()

	tmp := path.Join(dir, "test.socket")

	s := NewServer(
		WithNetwork("unix"),
		WithAddress(tmp),
	)

	if assert.NoError(t, s.Bind()) {
		assert.NotEqual(t, nil, s.ln)
		assert.NoError(t, s.Stop())
	}
}

func TestServerBindTwice(t *testing.T) {
	s := NewServer(
		WithAddress("127.0.0.1:0"),
	)

	if assert.NoError(t, s.Bind()) {
		assert.Equal(t, ErrBound, s.Bind())
	}
}

func TestServerBindInvalidNetwork(t *testing.T) {
	s := NewServer(
		WithNetwork("invalid"),
	)

	assert.NotEqual(t, nil, s.Bind())
}

func TestServerBindUnixExistingROFile(t *testing.T) {
	dir, err := ioutil.TempDir("", "pip.server.test.")
	if err != nil {
		assert.FailNow(t, "failed to create temporary directory", "ioutil.TempDir error: %s", err)
	}
	defer func() {
		if err = os.RemoveAll(dir); err != nil {
			assert.Fail(t, "failed to cleanup temporary directory", "os.RemoveAll error: %s", err)
		}
	}()

	tmp := path.Join(dir, "test.socket")
	f, err := os.Create(tmp)
	if err != nil {
		assert.FailNow(t, "failed to create file", "os.Create error: %s", err)
	}
	defer f.Close()

	if err = os.Chmod(dir, 0555); err != nil {
		assert.FailNow(t, "failed to make directory read-only", "os.Chmod error: %s", err)
	}
	defer os.Chmod(dir, 0777)

	s := NewServer(
		WithNetwork("unix"),
		WithAddress(tmp),
	)

	assert.NotEqual(t, nil, s.Bind())
}

func TestServerServe(t *testing.T) {
	s := NewServer()
	if err := s.Bind(); err != nil {
		assert.FailNow(t, "failed to bind server", "s.Bind error: %s", err)
	}
	defer s.Stop()

	var sErr error
	wg := new(sync.WaitGroup)
	wg.Add(1)
	go func() {
		defer wg.Done()
		sErr = s.Serve()
	}()

	a := s.ln.Addr()
	tcpAddr, err := net.ResolveTCPAddr(a.Network(), a.String())
	if err != nil {
		assert.FailNow(t, "failed to make address", "net.ResolveTCPAddr error: %s", err)
	}

	c, err := net.DialTCP(tcpAddr.Network(), nil, tcpAddr)
	if assert.NoError(t, err) {
		defer func(c net.Conn) {
			assert.NoError(t, c.Close())
		}(c)
		assert.NoError(t, c.CloseWrite())

		_, err = c.Read(make([]byte, 256))
		assert.Equal(t, io.EOF, err)
	}

	assert.NoError(t, s.Stop())

	wg.Wait()
	assert.NoError(t, sErr)
}

func TestServerServeNotBoundError(t *testing.T) {
	s := NewServer()
	assert.Equal(t, ErrNotBound, s.Serve())
}

func TestServerServeTwice(t *testing.T) {
	s := NewServer()
	if err := s.Bind(); err != nil {
		assert.FailNow(t, "failed to bind server", "s.Bind error: %s", err)
	}
	defer s.Stop()

	var sErr error
	wg := new(sync.WaitGroup)
	wg.Add(1)
	go func() {
		defer wg.Done()
		sErr = s.Serve()
	}()

	a := s.ln.Addr()
	tcpAddr, err := net.ResolveTCPAddr(a.Network(), a.String())
	if err != nil {
		assert.FailNow(t, "failed to make address", "net.ResolveTCPAddr error: %s", err)
	}

	c, err := net.DialTCP(tcpAddr.Network(), nil, tcpAddr)
	if assert.NoError(t, err) {
		defer func(c net.Conn) {
			assert.NoError(t, c.Close())
		}(c)
		assert.NoError(t, c.CloseWrite())

		_, err = c.Read(make([]byte, 256))
		assert.Equal(t, io.EOF, err)
	}

	assert.Equal(t, ErrStarted, s.Serve())

	assert.NoError(t, s.Stop())

	wg.Wait()
	assert.NoError(t, sErr)
}

func TestServerServeListenError(t *testing.T) {
	s := NewServer()
	if err := s.Bind(); err != nil {
		assert.FailNow(t, "failed to bind server", "s.Bind error: %s", err)
	}
	assert.NoError(t, s.ln.Close())

	err := errors.New("error")
	s.ln = makeBrokenListener(err)

	assert.Equal(t, err, s.Serve())
}

func TestServerServeWithConnectionLimit(t *testing.T) {
	s := NewServer(
		WithMaxConnections(1),
	)
	if err := s.Bind(); err != nil {
		assert.FailNow(t, "failed to bind server", "s.Bind error: %s", err)
	}
	defer s.Stop()

	var sErr error
	wg := new(sync.WaitGroup)
	wg.Add(1)
	go func() {
		defer wg.Done()
		sErr = s.Serve()
	}()

	a := s.ln.Addr()
	tcpAddr, err := net.ResolveTCPAddr(a.Network(), a.String())
	if err != nil {
		assert.FailNow(t, "failed to make address", "net.ResolveTCPAddr error: %s", err)
	}

	primary, err := net.DialTCP(tcpAddr.Network(), nil, tcpAddr)
	if assert.NoError(t, err) {
		defer func(c net.Conn) {
			assert.NoError(t, c.Close())
		}(primary)

		if _, err = primary.Write([]byte{4, 0, 0, 0, 0xef, 0xbe, 0xad, 0xde}); err != nil {
			assert.FailNow(t, "failed to write to connection", "TCPConn.Write error: %s", err)
		}

		secondary, err := net.DialTCP(tcpAddr.Network(), nil, tcpAddr)
		if assert.NoError(t, err) {
			defer func(c net.Conn) {
				assert.NoError(t, c.Close())
			}(secondary)

			_, err = secondary.Read(make([]byte, 256))
			assert.Equal(t, io.EOF, err)
		}

		assert.NoError(t, primary.CloseWrite())

		b := make([]byte, 256)
		n, err := primary.Read(b)
		assert.Equal(t, nil, err)
		assert.Equal(t, []byte{4, 0, 0, 0, 0xef, 0xbe, 0xad, 0xde}, b[:n])

		_, err = primary.Read(b)
		assert.Equal(t, io.EOF, err)
	}

	assert.NoError(t, s.Stop())

	wg.Wait()
	assert.NoError(t, sErr)
}

func TestServerServeWithConnectionLimitOnCloseError(t *testing.T) {
	errs := []error{}
	s := NewServer(
		WithMaxConnections(1),
		WithConnErrHandler(func(a net.Addr, err error) {
			if err != nil {
				errs = append(errs, err)
			}
		}),
	)
	if err := s.Bind(); err != nil {
		assert.FailNow(t, "failed to bind server", "s.Bind error: %s", err)
	}
	s.ln.Close()
	defer s.Stop()

	err := errors.New("test")
	c := makeTestWaitingConn()
	s.ln = newTestListener(c, makeSrvTestErrOnCloseConn(err))

	var sErr error
	wg := new(sync.WaitGroup)
	wg.Add(1)
	go func() {
		defer wg.Done()
		sErr = s.Serve()
	}()

	<-c.n

	assert.NoError(t, s.Stop())

	wg.Wait()
	assert.NoError(t, sErr)
	assert.Equal(t, errs, []error{err})
}

func TestServerStop(t *testing.T) {
	s := NewServer()

	if assert.NoError(t, s.Bind()) {
		s.ln.Close()

		c := makeTestWaitingConn()
		s.ln = newTestListener(c)

		var sErr error
		wg := new(sync.WaitGroup)
		wg.Add(1)
		go func() {
			defer wg.Done()
			sErr = s.Serve()
		}()

		<-c.n

		if assert.NoError(t, s.Stop()) {
			wg.Wait()

			assert.Equal(t, nil, s.ln)
			assert.NoError(t, sErr)
			if assert.NoError(t, s.Bind()) {
				assert.NoError(t, s.Stop())
			}
		}
	}
}

func TestServerStopNotStarted(t *testing.T) {
	assert.Equal(t, ErrNotStarted, NewServer().Stop())
}

type brokenListener struct {
	err error
}

func makeBrokenListener(err error) brokenListener {
	return brokenListener{
		err: err,
	}
}

func (ln brokenListener) Accept() (net.Conn, error) { return nil, ln.err }
func (ln brokenListener) Close() error              { return ln.err }
func (ln brokenListener) Addr() net.Addr            { panic(ln.err) }

type testListener struct {
	c []net.Conn
	d chan struct{}
}

func newTestListener(c ...net.Conn) *testListener {
	return &testListener{
		c: c,
		d: make(chan struct{}),
	}
}

var errClosedListener = errors.New(netListenerClosedMsg)

func (ln *testListener) Accept() (net.Conn, error) {
	if len(ln.c) > 0 {
		c := ln.c[0]
		ln.c = ln.c[1:]
		return c, nil
	}

	<-ln.d
	return nil, &net.OpError{
		Err: errClosedListener,
	}
}

func (ln *testListener) Close() error {
	close(ln.d)
	return nil
}

func (ln *testListener) Addr() net.Addr { panic("not implemented") }

type testWaitingConn struct {
	n chan struct{}
	d chan struct{}
}

func makeTestWaitingConn() testWaitingConn {
	return testWaitingConn{
		n: make(chan struct{}),
		d: make(chan struct{}),
	}
}

func (c testWaitingConn) Close() error {
	close(c.d)
	return nil
}

func (c testWaitingConn) Read(b []byte) (n int, err error) {
	if c.n != nil {
		close(c.n)
	}

	<-c.d
	return 0, io.EOF
}

func (c testWaitingConn) RemoteAddr() net.Addr {
	addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
	if err != nil {
		panic(err)
	}

	return addr
}

func (c testWaitingConn) Write(b []byte) (n int, err error)  { panic("not implemented") }
func (c testWaitingConn) LocalAddr() net.Addr                { panic("not implemented") }
func (c testWaitingConn) SetDeadline(t time.Time) error      { panic("not implemented") }
func (c testWaitingConn) SetReadDeadline(t time.Time) error  { panic("not implemented") }
func (c testWaitingConn) SetWriteDeadline(t time.Time) error { panic("not implemented") }

type srvTestErrOnCloseConn struct {
	err error
}

func makeSrvTestErrOnCloseConn(err error) net.Conn {
	return srvTestErrOnCloseConn{
		err: err,
	}
}

func (c srvTestErrOnCloseConn) Close() error { return c.err }
func (c srvTestErrOnCloseConn) RemoteAddr() net.Addr {
	addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
	if err != nil {
		panic(err)
	}

	return addr
}

func (c srvTestErrOnCloseConn) Read(b []byte) (n int, err error)   { panic("not implemented") }
func (c srvTestErrOnCloseConn) Write(b []byte) (n int, err error)  { panic("not implemented") }
func (c srvTestErrOnCloseConn) LocalAddr() net.Addr                { panic("not implemented") }
func (c srvTestErrOnCloseConn) SetDeadline(t time.Time) error      { panic("not implemented") }
func (c srvTestErrOnCloseConn) SetReadDeadline(t time.Time) error  { panic("not implemented") }
func (c srvTestErrOnCloseConn) SetWriteDeadline(t time.Time) error { panic("not implemented") }
