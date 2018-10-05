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

	pool := makeBytePool(defMaxSize, false)

	wg := new(sync.WaitGroup)
	ps := makePipes(3, defTimeout.Nanoseconds())
	for i := 0; i < len(ps.p); i++ {
		idx, p := ps.alloc()
		wg.Add(1)
		go func(wg *sync.WaitGroup, idx int, p pipe) {
			defer wg.Done()
			defer ps.free(idx)

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
				pool.Put(b[:cap(b)])
			}
		}(wg, idx, p)
	}

	c := NewClient().(*client)

	wg.Add(1)
	c.reader(wg, newTestReaderConn(
		0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01, 0x01, 0x01, 0x01,
		0x08, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x02, 0x02, 0x02, 0x02,
		0x08, 0x00, 0x00, 0x00, 0x02, 0x00, 0x00, 0x00, 0x03, 0x03, 0x03, 0x03,
	), ps)

	wg.Wait()

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
