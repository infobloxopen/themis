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
	assert.NotEqual(t, nil, NewServer(WithNetwork("invalid")).Bind())
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
	c, err := net.Dial(a.Network(), a.String())
	if assert.NoError(t, err) {
		_, err := c.Read(make([]byte, 256))
		assert.Equal(t, io.EOF, err)
	}

	assert.NoError(t, s.Stop())

	wg.Wait()
	assert.NoError(t, sErr)
}

func TestServerServeError(t *testing.T) {
	s := NewServer()

	err := errors.New("error")
	s.ln = newBrokenListener(err)

	assert.Equal(t, err, s.Serve())
}

func TestServerStop(t *testing.T) {
	s := NewServer()

	if assert.NoError(t, s.Bind()) {
		if assert.NoError(t, s.Stop()) {
			assert.Equal(t, nil, s.ln)
			if assert.NoError(t, s.Bind()) {
				assert.NoError(t, s.Stop())
			}
		}
	}
}

func TestServerStopNotStarted(t *testing.T) {
	assert.Equal(t, ErrNotStarted, NewServer().Stop())
}

func TestServerStopCloseError(t *testing.T) {
	s := NewServer()

	err := errors.New("error")
	s.ln = newBrokenListener(err)
	assert.Equal(t, err, s.Stop())
}

type brokenListener struct {
	err error
}

func newBrokenListener(err error) brokenListener {
	return brokenListener{
		err: err,
	}
}

func (ln brokenListener) Accept() (net.Conn, error) { return nil, ln.err }
func (ln brokenListener) Close() error              { return ln.err }
func (ln brokenListener) Addr() net.Addr            { panic(ln.err) }
