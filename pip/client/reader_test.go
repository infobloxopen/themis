package client

import (
	"io"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestReader(t *testing.T) {
	m := new(sync.Mutex)
	msgs := [][]byte{}
	errs := []error{}

	c := NewClient(
		WithMaxQueue(3),
	).(*client)

	conn := c.newConnection(newTestReaderConn(
		0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01, 0x01, 0x01, 0x01,
		0x08, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x02, 0x02, 0x02, 0x02,
		0x08, 0x00, 0x00, 0x00, 0x02, 0x00, 0x00, 0x00, 0x03, 0x03, 0x03, 0x03,
	))

	for i := 0; i < len(conn.p.p); i++ {
		idx, p := conn.p.alloc()
		conn.w.Add(1)
		go func(idx int, p pipe) {
			defer conn.w.Done()
			defer conn.p.free(idx)

			b, err := p.get()
			m.Lock()
			if b != nil {
				msg := make([]byte, len(b))
				copy(msg, b)

				msgs = append(msgs, msg)
			} else {
				msgs = append(msgs, nil)
			}

			errs = append(errs, err)
			m.Unlock()

			if b != nil {
				c.pool.Put(b)
			}
		}(idx, p)
	}

	conn.w.Add(1)
	conn.reader()
	conn.w.Wait()

	assert.ElementsMatch(t, [][]byte{
		{0x01, 0x01, 0x01, 0x01},
		{0x02, 0x02, 0x02, 0x02},
		{0x03, 0x03, 0x03, 0x03},
	}, msgs)
	assert.Equal(t, []error{nil, nil, nil}, errs)
}

type testReaderConn struct {
	b []byte
}

func newTestReaderConn(b ...byte) *testReaderConn {
	return &testReaderConn{
		b: b,
	}
}

func (r *testReaderConn) Read(b []byte) (int, error) {
	if len(r.b) > 0 {
		n := copy(b, r.b)

		r.b = r.b[n:]
		if len(r.b) > 0 {
			return n, nil
		}

		return n, io.EOF
	}

	return 0, io.EOF
}

func (r *testReaderConn) Close() error {
	return nil
}
func (r *testReaderConn) Write(b []byte) (int, error)        { panic("not implemented") }
func (r *testReaderConn) LocalAddr() net.Addr                { panic("not implemented") }
func (r *testReaderConn) RemoteAddr() net.Addr               { panic("not implemented") }
func (r *testReaderConn) SetDeadline(t time.Time) error      { panic("not implemented") }
func (r *testReaderConn) SetReadDeadline(t time.Time) error  { panic("not implemented") }
func (r *testReaderConn) SetWriteDeadline(t time.Time) error { panic("not implemented") }
