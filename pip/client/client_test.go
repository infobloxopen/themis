package client

import (
	"fmt"
	"io"
	"net"
	"sync"
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
	s := server.NewServer()
	if !assert.NoError(t, s.Bind()) {
		assert.FailNow(t, "failed to bind server")
	}
	defer func() {
		assert.NoError(t, s.Stop())
	}()
	var sErr error
	go func() {
		sErr = s.Serve()
	}()
	defer func() {
		assert.NoError(t, sErr)
	}()

	c := NewClient()

	if assert.NoError(t, c.Connect()) {
		assert.Equal(t, ErrConnected, c.Connect())
	}
}

func TestClientConnectNoServer(t *testing.T) {
	c := NewClient()

	err := c.Connect()
	if assert.Error(t, err) {
		assert.IsType(t, (*net.OpError)(nil), err)
	}
}

func TestClientClose(t *testing.T) {
	s := server.NewServer()
	if !assert.NoError(t, s.Bind()) {
		assert.FailNow(t, "failed to bind server")
	}
	defer func() {
		assert.NoError(t, s.Stop())
	}()
	var sErr error
	go func() {
		sErr = s.Serve()
	}()
	defer func() {
		assert.NoError(t, sErr)
	}()

	c := NewClient()

	if assert.NoError(t, c.Connect()) {
		c.Close()
		if cc, ok := c.(*client); assert.True(t, ok) {
			assert.Equal(t, pipClientIdle, *cc.state)
		}

		c.Close()
		if cc, ok := c.(*client); assert.True(t, ok) {
			assert.Equal(t, pipClientIdle, *cc.state)
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
		withTestTermFlushChannel(make(chan time.Time)),
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

	c.Close()
	wg.Wait()

	assert.Equal(t, errReaderBroken, err1)
	assert.Error(t, err2)
	assert.Equal(t, errWriterBroken, err3)
}

func TestClientGet(t *testing.T) {
	s := server.NewServer()
	if !assert.NoError(t, s.Bind()) {
		assert.FailNow(t, "failed to bind server")
	}
	defer func() {
		assert.NoError(t, s.Stop())
	}()
	var sErr error
	go func() {
		sErr = s.Serve()
	}()
	defer func() {
		assert.NoError(t, sErr)
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
	s := server.NewServer()
	if !assert.NoError(t, s.Bind()) {
		assert.FailNow(t, "failed to bind server")
	}
	defer func() {
		assert.NoError(t, s.Stop())
	}()
	var sErr error
	go func() {
		sErr = s.Serve()
	}()
	defer func() {
		assert.NoError(t, sErr)
	}()

	c := NewClient()
	if err := c.Connect(); assert.NoError(t, err) {
		defer c.Close()

		_, err := c.Get(pdp.MakeExpressionAssignment("test", pdp.UndefinedValue))
		assert.Error(t, err)
	}
}
