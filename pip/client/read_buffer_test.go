package client

import (
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewReadBuffer(t *testing.T) {
	p := makePipes(1)
	pool := makeBytePool(8, false)

	r := newReadBuffer(1024, 8, pool, p)
	if assert.NotZero(t, r) {
		assert.Equal(t, 1024, len(r.in))
		assert.Equal(t, msgSizeBytes, cap(r.buf))
		assert.Zero(t, r.msgBuf)
		assert.Zero(t, r.size)
		assert.Equal(t, uint32(8), r.max)
		assert.Equal(t, -1, r.idx)
		assert.Equal(t, pool, r.pool)
		assert.Equal(t, p, r.p)
	}
}

func TestReadBufferRead(t *testing.T) {
	p := makePipes(1)
	pool := makeBytePool(8, false)

	r := newReadBuffer(1024, 8, pool, p)
	assert.False(t, r.read(newTestReadBufferReadCloser(
		[]byte{
			0x08, 0x00, 0x00, 0x00,
			0x00, 0x00, 0x00, 0x00,
			0xde, 0xc0, 0xad, 0xde,
		},
	)))

	b, err := p.p[0].get()
	defer pool.Put(b)

	assert.NoError(t, err)
	assert.Equal(t, []byte{0xde, 0xc0, 0xad, 0xde}, b)
}

func TestReadBufferReadWithInvalidIdx(t *testing.T) {
	p := makePipes(1)
	pool := makeBytePool(8, false)

	r := newReadBuffer(1024, 8, pool, p)
	assert.False(t, r.read(newTestReadBufferReadCloser(
		[]byte{
			0x08, 0x00, 0x00, 0x00,
			0x01, 0x00, 0x00, 0x00,
			0xde, 0xc0, 0xad, 0xde,
		},
	)))
}

func TestReadBufferExtractData(t *testing.T) {
	p := makePipes(1)
	pool := makeBytePool(8, false)

	r := newReadBuffer(1024, 8, pool, p)

	b := []byte{
		0x08, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00,
		0xde, 0xc0, 0xad, 0xde,
	}
	n, ok := r.extractData(b)
	assert.Equal(t, 4, n)
	assert.True(t, ok)
	assert.Equal(t, 8, r.size)
	assert.Equal(t, 0, len(r.buf))
	assert.Equal(t, msgSizeBytes, cap(r.buf))

	n, ok = r.extractData(b[4:])
	assert.Equal(t, 4, n)
	assert.True(t, ok)
	assert.Equal(t, 4, r.size)
	assert.Equal(t, 0, r.idx)
	assert.Equal(t, 0, len(r.buf))
	assert.Equal(t, msgIdxBytes, cap(r.buf))
	assert.NotZero(t, r.msgBuf)
	assert.Equal(t, 8, cap(r.msgBuf))

	n, ok = r.extractData(b[8:])
	assert.Equal(t, 4, n)
	assert.True(t, ok)
	assert.Equal(t, 0, r.size)
	assert.Equal(t, -1, r.idx)
	assert.Zero(t, r.msgBuf)

	b, err := p.p[0].get()
	defer pool.Put(b)

	assert.NoError(t, err)
	assert.Equal(t, []byte{0xde, 0xc0, 0xad, 0xde}, b)
}

func TestReadBufferFillSize(t *testing.T) {
	p := makePipes(1)
	pool := makeBytePool(8, false)

	r := newReadBuffer(1024, 8, pool, p)

	n, ok := r.fillSize([]byte{
		0x08, 0x00, 0x00, 0x00,
	})
	assert.Equal(t, 4, n)
	assert.True(t, ok)
	assert.Equal(t, 8, r.size)
	assert.Equal(t, 0, len(r.buf))
	assert.Equal(t, msgSizeBytes, cap(r.buf))
}

func TestReadBufferFillSizePartial(t *testing.T) {
	p := makePipes(1)
	pool := makeBytePool(8, false)

	r := newReadBuffer(1024, 8, pool, p)

	n, ok := r.fillSize([]byte{
		0x08, 0x00, 0x00,
	})
	assert.Equal(t, 3, n)
	assert.True(t, ok)
	assert.Equal(t, 0, r.size)
	assert.Equal(t, 3, len(r.buf))
	assert.Equal(t, msgSizeBytes, cap(r.buf))
}

func TestReadBufferFillSizeTooBig(t *testing.T) {
	p := makePipes(1)
	pool := makeBytePool(8, false)

	r := newReadBuffer(1024, 8, pool, p)

	n, ok := r.fillSize([]byte{
		0xff, 0xff, 0xff, 0xff,
	})
	assert.Equal(t, 4, n)
	assert.False(t, ok)
	assert.Equal(t, 0, r.size)
	assert.Equal(t, 0, len(r.buf))
	assert.Equal(t, msgSizeBytes, cap(r.buf))
}

func TestReadBufferFillSizeTooSmall(t *testing.T) {
	p := makePipes(1)
	pool := makeBytePool(8, false)

	r := newReadBuffer(1024, 8, pool, p)

	n, ok := r.fillSize([]byte{
		0x02, 0x00, 0x00, 0x00,
	})
	assert.Equal(t, 4, n)
	assert.False(t, ok)
	assert.Equal(t, 2, r.size)
	assert.Equal(t, 0, len(r.buf))
	assert.Equal(t, msgSizeBytes, cap(r.buf))
}

func TestReadBufferFillIdx(t *testing.T) {
	p := makePipes(1)
	pool := makeBytePool(8, false)

	r := newReadBuffer(1024, 8, pool, p)

	n, ok := r.fillIdx([]byte{
		0x00, 0x00, 0x00, 0x00,
	})
	assert.Equal(t, 4, n)
	assert.True(t, ok)
	assert.Equal(t, 0, r.idx)
	assert.Equal(t, 0, len(r.buf))
	assert.Equal(t, msgIdxBytes, cap(r.buf))
}

func TestReadBufferFillIdxPartial(t *testing.T) {
	p := makePipes(1)
	pool := makeBytePool(8, false)

	r := newReadBuffer(1024, 8, pool, p)

	n, ok := r.fillIdx([]byte{
		0x00, 0x00, 0x00,
	})
	assert.Equal(t, 3, n)
	assert.True(t, ok)
	assert.Equal(t, -1, r.idx)
	assert.Equal(t, 3, len(r.buf))
	assert.Equal(t, msgIdxBytes, cap(r.buf))
}

func TestReadBufferFillIdxTooBig(t *testing.T) {
	p := makePipes(1)
	pool := makeBytePool(8, false)

	r := newReadBuffer(1024, 8, pool, p)

	n, ok := r.fillIdx([]byte{
		0xff, 0xff, 0xff, 0xff,
	})
	assert.Equal(t, 4, n)
	assert.False(t, ok)
	assert.Equal(t, -1, r.idx)
	assert.Equal(t, 0, len(r.buf))
	assert.Equal(t, msgIdxBytes, cap(r.buf))
}

func TestReadBufferFillMsg(t *testing.T) {
	p := makePipes(1)
	pool := makeBytePool(8, false)

	r := newReadBuffer(1024, 8, pool, p)
	r.size = 4
	r.idx = 0
	r.msgBuf = pool.Get()[:0]

	n, ok := r.fillMsg([]byte{
		0xde, 0xc0, 0xad, 0xde,
	})
	assert.Equal(t, 4, n)
	assert.True(t, ok)
	assert.Equal(t, 0, r.size)
	assert.Equal(t, -1, r.idx)
	assert.Zero(t, r.msgBuf)

	b, err := p.p[0].get()
	defer pool.Put(b)

	assert.NoError(t, err)
	assert.Equal(t, []byte{0xde, 0xc0, 0xad, 0xde}, b)
}

func TestReadBufferFillMsgPartial(t *testing.T) {
	p := makePipes(1)
	pool := makeBytePool(8, false)

	r := newReadBuffer(1024, 8, pool, p)
	r.size = 4
	r.idx = 0
	r.msgBuf = pool.Get()[:0]

	n, ok := r.fillMsg([]byte{
		0xde, 0xc0,
	})
	assert.Equal(t, 2, n)
	assert.True(t, ok)
	assert.Equal(t, 2, r.size)
	assert.Equal(t, 0, r.idx)
	assert.Equal(t, []byte{0xde, 0xc0}, r.msgBuf)
}

func TestReadBufferClean(t *testing.T) {
	p := makePipes(1)
	pool := makeBytePool(8, false)

	r := newReadBuffer(1024, 8, pool, p)
	r.size = 4
	r.idx = 0
	r.msgBuf = pool.Get()[:0]

	n, ok := r.fillMsg([]byte{
		0xde, 0xc0,
	})
	assert.Equal(t, 2, n)
	assert.True(t, ok)
	assert.Equal(t, 2, r.size)
	assert.Equal(t, 0, r.idx)
	assert.Equal(t, []byte{0xde, 0xc0}, r.msgBuf)

	r.clean()
	assert.Equal(t, 0, len(r.buf))
	assert.Equal(t, msgSizeBytes, cap(r.buf))
	assert.Zero(t, r.msgBuf)
	assert.Equal(t, 0, r.size)
	assert.Equal(t, -1, r.idx)
}

type testReadBufferReadCloser struct {
	b []byte
}

func newTestReadBufferReadCloser(b []byte) *testReadBufferReadCloser {
	return &testReadBufferReadCloser{
		b: b,
	}
}

func (r *testReadBufferReadCloser) Read(b []byte) (int, error) {
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

func (r *testReadBufferReadCloser) Close() error {
	return nil
}
