package server

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestRead(t *testing.T) {
	errs := []error{}
	c := connWithErrHandler{
		c: newRTestConn(
			[]byte{0x04, 0x00, 0x00, 0x00, 0xef, 0xbe, 0xad, 0xde},
		),
		h: func(a net.Addr, err error) {
			if err != nil {
				errs = append(errs, err)
			}
		},
	}

	out := []uint32{}
	msgs := makePool(1, 10)
	for msg := range read(c, msgs, 2) {
		assert.Equal(t, 4, len(msg), "message %d", len(out)+1)
		out = append(out, binary.LittleEndian.Uint32(msg))
	}

	assert.Equal(t, []error{}, errs)
	assert.Equal(t, []uint32{0xdeadbeef}, out)
}

func TestReadWithMsgBufferOverflow(t *testing.T) {
	errs := []error{}
	c := connWithErrHandler{
		c: newRTestConn(
			[]byte{0x04, 0x00, 0x00, 0x00, 0xef, 0xbe, 0xad, 0xde},
		),
		h: func(a net.Addr, err error) {
			if err != nil {
				errs = append(errs, err)
			}
		},
	}

	out := []uint32{}
	msgs := makePool(1, 2)
	for msg := range read(c, msgs, 2) {
		assert.Equal(t, 4, len(msg), "message %d", len(out)+1)
		out = append(out, binary.LittleEndian.Uint32(msg))
	}

	assert.Equal(t, []error{ErrMsgOverflow}, errs)
	assert.Equal(t, []uint32{}, out)
}

func TestReadError(t *testing.T) {
	errs := []error{}
	err := errors.New("test")
	c := connWithErrHandler{
		c: newRTestConn(
			[]byte{0x04, 0x00, 0x00, 0x00, 0xef, 0xbe, 0xad, 0xde},
			err,
		),
		h: func(a net.Addr, err error) {
			if err != nil {
				errs = append(errs, err)
			}
		},
	}

	out := []uint32{}
	msgs := makePool(1, 10)
	for msg := range read(c, msgs, 2) {
		assert.Equal(t, 4, len(msg), "message %d", len(out)+1)
		out = append(out, binary.LittleEndian.Uint32(msg))
	}

	assert.Equal(t, []error{err}, errs)
	assert.Equal(t, []uint32{0xdeadbeef}, out)
}

type rTestConn struct {
	i    int
	j    int
	data []interface{}
}

func newRTestConn(data ...interface{}) net.Conn {
	return &rTestConn{
		data: data,
	}
}

func (c *rTestConn) Read(b []byte) (int, error) {
	if c.i >= len(c.data) {
		return 0, io.EOF
	}

	v := c.data[c.i]
	switch v := v.(type) {
	case []byte:
		if c.j < len(v) {
			n := copy(b, v[c.j:])
			c.j += n
			return n, nil
		}

		c.j = 0
		c.i++
		return c.Read(b)

	case error:
		return 0, v
	}

	panic(fmt.Errorf("failed to return %T (%#v) from Conn.Read", v, v))
}

func (c *rTestConn) RemoteAddr() net.Addr {
	addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
	if err != nil {
		panic(err)
	}

	return addr
}

func (c *rTestConn) Close() error                       { panic("not implemented") }
func (c *rTestConn) Write(b []byte) (n int, err error)  { panic("not implemented") }
func (c *rTestConn) LocalAddr() net.Addr                { panic("not implemented") }
func (c *rTestConn) SetDeadline(t time.Time) error      { panic("not implemented") }
func (c *rTestConn) SetReadDeadline(t time.Time) error  { panic("not implemented") }
func (c *rTestConn) SetWriteDeadline(t time.Time) error { panic("not implemented") }
