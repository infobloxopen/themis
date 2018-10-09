package client

import (
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewReadBuffer(t *testing.T) {
	ctx := makeReadBufferContext(1024, 8, 1)
	if assert.NotZero(t, ctx.r) {
		assert.Equal(t, 1024, len(ctx.r.in))
		assert.Equal(t, msgSizeBytes, cap(ctx.r.buf))
		assert.Zero(t, ctx.r.msgBuf)
		assert.Zero(t, ctx.r.size)
		assert.Equal(t, uint32(8), ctx.r.max)
		assert.Equal(t, -1, ctx.r.idx)
		assert.Equal(t, ctx.pool, ctx.r.pool)
		assert.Equal(t, ctx.p, ctx.r.p)
	}
}

func TestReadBufferRead(t *testing.T) {
	ctx := makeReadBufferContext(1024, 8, 1)
	err := ctx.r.read(newTestReadBufferReadCloser(
		[]byte{
			0x08, 0x00, 0x00, 0x00,
			0x00, 0x00, 0x00, 0x00,
			0xde, 0xc0, 0xad, 0xde,
		},
	))
	assert.Equal(t, io.EOF, err)

	b, err := ctx.p.p[0].get()
	defer ctx.pool.Put(b)

	assert.NoError(t, err)
	assert.Equal(t, []byte{0xde, 0xc0, 0xad, 0xde}, b)
}

func TestReadBufferReadWithInvalidIdx(t *testing.T) {
	ctx := makeReadBufferContext(1024, 8, 1)
	err := ctx.r.read(newTestReadBufferReadCloser(
		[]byte{
			0x08, 0x00, 0x00, 0x00,
			0x01, 0x00, 0x00, 0x00,
			0xde, 0xc0, 0xad, 0xde,
		},
	))
	assert.Equal(t, errMsgInvalidIndex, err)
}

func TestReadBufferExtractData(t *testing.T) {
	ctx := makeReadBufferContext(1024, 8, 1)

	b := []byte{
		0x08, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00,
		0xde, 0xc0, 0xad, 0xde,
	}
	n, err := ctx.r.extractData(b)
	assert.Equal(t, 4, n)
	assert.NoError(t, err)
	assert.Equal(t, 8, ctx.r.size)
	assert.Equal(t, 0, len(ctx.r.buf))
	assert.Equal(t, msgSizeBytes, cap(ctx.r.buf))

	n, err = ctx.r.extractData(b[4:])
	assert.Equal(t, 4, n)
	assert.NoError(t, err)
	assert.Equal(t, 4, ctx.r.size)
	assert.Equal(t, 0, ctx.r.idx)
	assert.Equal(t, 0, len(ctx.r.buf))
	assert.Equal(t, msgIdxBytes, cap(ctx.r.buf))
	assert.NotZero(t, ctx.r.msgBuf)
	assert.Equal(t, 8, cap(ctx.r.msgBuf))

	n, err = ctx.r.extractData(b[8:])
	assert.Equal(t, 4, n)
	assert.NoError(t, err)
	assert.Equal(t, 0, ctx.r.size)
	assert.Equal(t, -1, ctx.r.idx)
	assert.Zero(t, ctx.r.msgBuf)

	b, err = ctx.p.p[0].get()
	defer ctx.pool.Put(b)

	assert.NoError(t, err)
	assert.Equal(t, []byte{0xde, 0xc0, 0xad, 0xde}, b)
}

func TestReadBufferFillSize(t *testing.T) {
	ctx := makeReadBufferContext(1024, 8, 1)

	n, err := ctx.r.fillSize([]byte{
		0x08, 0x00, 0x00, 0x00,
	})
	assert.Equal(t, 4, n)
	assert.NoError(t, err)
	assert.Equal(t, 8, ctx.r.size)
	assert.Equal(t, 0, len(ctx.r.buf))
	assert.Equal(t, msgSizeBytes, cap(ctx.r.buf))
}

func TestReadBufferFillSizePartial(t *testing.T) {
	ctx := makeReadBufferContext(1024, 8, 1)

	n, err := ctx.r.fillSize([]byte{
		0x08, 0x00, 0x00,
	})
	assert.Equal(t, 3, n)
	assert.NoError(t, err)
	assert.Equal(t, 0, ctx.r.size)
	assert.Equal(t, 3, len(ctx.r.buf))
	assert.Equal(t, msgSizeBytes, cap(ctx.r.buf))
}

func TestReadBufferFillSizeTooBig(t *testing.T) {
	ctx := makeReadBufferContext(1024, 8, 1)

	n, err := ctx.r.fillSize([]byte{
		0xff, 0xff, 0xff, 0xff,
	})
	assert.Equal(t, 4, n)
	assert.Equal(t, errMsgOverflow, err)
	assert.Equal(t, 0, ctx.r.size)
	assert.Equal(t, 0, len(ctx.r.buf))
	assert.Equal(t, msgSizeBytes, cap(ctx.r.buf))
}

func TestReadBufferFillSizeTooSmall(t *testing.T) {
	ctx := makeReadBufferContext(1024, 8, 1)

	n, err := ctx.r.fillSize([]byte{
		0x02, 0x00, 0x00, 0x00,
	})
	assert.Equal(t, 4, n)
	assert.Equal(t, errMsgUnderflow, err)
	assert.Equal(t, 2, ctx.r.size)
	assert.Equal(t, 0, len(ctx.r.buf))
	assert.Equal(t, msgSizeBytes, cap(ctx.r.buf))
}

func TestReadBufferFillIdx(t *testing.T) {
	ctx := makeReadBufferContext(1024, 8, 1)

	n, err := ctx.r.fillIdx([]byte{
		0x00, 0x00, 0x00, 0x00,
	})
	assert.Equal(t, 4, n)
	assert.NoError(t, err)
	assert.Equal(t, 0, ctx.r.idx)
	assert.Equal(t, 0, len(ctx.r.buf))
	assert.Equal(t, msgIdxBytes, cap(ctx.r.buf))
}

func TestReadBufferFillIdxPartial(t *testing.T) {
	ctx := makeReadBufferContext(1024, 8, 1)

	n, err := ctx.r.fillIdx([]byte{
		0x00, 0x00, 0x00,
	})
	assert.Equal(t, 3, n)
	assert.NoError(t, err)
	assert.Equal(t, -1, ctx.r.idx)
	assert.Equal(t, 3, len(ctx.r.buf))
	assert.Equal(t, msgIdxBytes, cap(ctx.r.buf))
}

func TestReadBufferFillIdxTooBig(t *testing.T) {
	ctx := makeReadBufferContext(1024, 8, 1)

	n, err := ctx.r.fillIdx([]byte{
		0xff, 0xff, 0xff, 0xff,
	})
	assert.Equal(t, 4, n)
	assert.Equal(t, errMsgInvalidIndex, err)
	assert.Equal(t, -1, ctx.r.idx)
	assert.Equal(t, 0, len(ctx.r.buf))
	assert.Equal(t, msgIdxBytes, cap(ctx.r.buf))
}

func TestReadBufferFillMsg(t *testing.T) {
	ctx := makeReadBufferContext(1024, 8, 1)
	ctx.r.size = 4
	ctx.r.idx = 0
	ctx.r.msgBuf = ctx.pool.Get()[:0]

	n, err := ctx.r.fillMsg([]byte{
		0xde, 0xc0, 0xad, 0xde,
	})
	assert.Equal(t, 4, n)
	assert.NoError(t, err)
	assert.Equal(t, 0, ctx.r.size)
	assert.Equal(t, -1, ctx.r.idx)
	assert.Zero(t, ctx.r.msgBuf)

	b, err := ctx.p.p[0].get()
	defer ctx.pool.Put(b)

	assert.NoError(t, err)
	assert.Equal(t, []byte{0xde, 0xc0, 0xad, 0xde}, b)
}

func TestReadBufferFillMsgPartial(t *testing.T) {
	ctx := makeReadBufferContext(1024, 8, 1)
	ctx.r.size = 4
	ctx.r.idx = 0
	ctx.r.msgBuf = ctx.pool.Get()[:0]

	n, err := ctx.r.fillMsg([]byte{
		0xde, 0xc0,
	})
	assert.Equal(t, 2, n)
	assert.NoError(t, err)
	assert.Equal(t, 2, ctx.r.size)
	assert.Equal(t, 0, ctx.r.idx)
	assert.Equal(t, []byte{0xde, 0xc0}, ctx.r.msgBuf)
}

func TestReadBufferFinalize(t *testing.T) {
	ctx := makeReadBufferContext(1024, 8, 1)
	ctx.r.size = 4
	ctx.r.idx = 0
	ctx.r.msgBuf = ctx.pool.Get()[:0]

	n, err := ctx.r.fillMsg([]byte{
		0xde, 0xc0,
	})
	assert.Equal(t, 2, n)
	assert.NoError(t, err)
	assert.Equal(t, 2, ctx.r.size)
	assert.Equal(t, 0, ctx.r.idx)
	assert.Equal(t, []byte{0xde, 0xc0}, ctx.r.msgBuf)

	ctx.r.finalize()
}

type readBufferContext struct {
	r    *readBuffer
	p    pipes
	pool bytePool
}

func makeReadBufferContext(n, m, q int) readBufferContext {
	out := readBufferContext{
		p:    makePipes(q, defTimeout.Nanoseconds()),
		pool: makeBytePool(m, false),
	}

	out.r = newReadBuffer(n, m, out.pool, out.p)
	return out
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
