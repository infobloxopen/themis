package client

import (
	"encoding/binary"
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestWriter(t *testing.T) {
	c := NewClient(
		WithMaxQueue(1),
	).(*client)

	conn := c.newConnection(nil)
	conn.n = makeTestWriterConn(conn.p)

	conn.w.Add(1)
	go conn.writer()

	i, p := conn.p.alloc()
	defer conn.p.free(i)

	conn.r <- request{
		i: i,
		b: []byte{0xef, 0xbe, 0xad, 0xde},
	}

	_, err := p.get()
	assert.NoError(t, err)

	close(conn.r)
	conn.w.Wait()
}

func TestWriterNoTimeout(t *testing.T) {
	c := NewClient(
		WithMaxQueue(1),
		withTestWriteFlushChannel(make(chan time.Time)),
	).(*client)

	conn := c.newConnection(nil)
	conn.n = makeTestWriterConn(conn.p)

	conn.w.Add(1)
	go conn.writer()

	i, p := conn.p.alloc()
	defer conn.p.free(i)

	conn.r <- request{
		i: i,
		b: []byte{0xef, 0xbe, 0xad, 0xde},
	}
	close(conn.r)

	conn.w.Wait()
	_, err := p.get()
	assert.NoError(t, err)
}

type testWriterConn struct {
	p pipes
}

func makeTestWriterConn(p pipes) testWriterConn {
	return testWriterConn{
		p: p,
	}
}

func (c testWriterConn) Write(b []byte) (int, error) {
	n := 0

	for len(b) > 0 {
		if len(b) < msgSizeBytes {
			return n, fmt.Errorf("expected %d bytes for size but got only %d", msgSizeBytes, len(b))
		}
		size := int(binary.LittleEndian.Uint32(b))
		b = b[msgSizeBytes:]
		n += msgSizeBytes
		if len(b) < size {
			return n, fmt.Errorf("expected %d bytes for message but got only %d", size, len(b))
		}

		if size < msgIdxBytes {
			return n, fmt.Errorf("expected %d bytes for index but got only %d", msgIdxBytes, size)
		}
		idx := int(binary.LittleEndian.Uint32(b))
		c.p.putBytes(idx, append(make([]byte, 0, size-msgIdxBytes), b[msgIdxBytes:size]...))

		b = b[size:]
		n += size
	}

	return n, nil
}

func (c testWriterConn) Read(b []byte) (int, error)         { panic("not implemented") }
func (c testWriterConn) Close() error                       { panic("not implemented") }
func (c testWriterConn) LocalAddr() net.Addr                { panic("not implemented") }
func (c testWriterConn) RemoteAddr() net.Addr               { panic("not implemented") }
func (c testWriterConn) SetDeadline(t time.Time) error      { panic("not implemented") }
func (c testWriterConn) SetReadDeadline(t time.Time) error  { panic("not implemented") }
func (c testWriterConn) SetWriteDeadline(t time.Time) error { panic("not implemented") }
