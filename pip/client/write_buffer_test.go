package client

import (
	"encoding/binary"
	"errors"
	"fmt"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewWriteBuffer(t *testing.T) {
	ctx := makeWriteBufferContext(16, 1, nil)
	if assert.NotZero(t, ctx.w) {
		assert.Equal(t, 16, cap(ctx.w.out))
		assert.Empty(t, ctx.w.out)
		assert.NotZero(t, ctx.w.idx)
		assert.Equal(t, ctx.ps, ctx.w.p)
		assert.Equal(t, ctx.inc, ctx.w.inc)
	}
}

func TestWriteBufferRem(t *testing.T) {
	ctx := makeWriteBufferContext(16, 1, nil)

	assert.Equal(t, 16, ctx.w.rem())

	i, _ := ctx.ps.alloc()
	defer ctx.ps.free(i)

	ctx.w.put(request{
		i: i,
		b: []byte{0xde, 0xc0, 0xad, 0xde},
	})
	assert.Equal(t, 4, ctx.w.rem())
}

func TestWriteBufferPut(t *testing.T) {
	ctx := makeWriteBufferContext(16, 2, nil)

	i1, _, b1, err1 := ctx.put([]byte{0xde, 0xc0, 0xad, 0xde}, true)
	i2, _, b2, err2 := ctx.put([]byte{0xde, 0xc0, 0xad, 0xde, 0xef, 0xeb, 0, 0}, true)

	ctx.wg.Wait()

	assert.Equal(t, []byte{0xde, 0xc0, 0xad, 0xde}, *b1)
	assert.NoError(t, *err1)
	assert.Equal(t, i1, <-ctx.inc)

	assert.Equal(t, []byte{0xde, 0xc0, 0xad, 0xde, 0xef, 0xeb, 0, 0}, *b2)
	assert.NoError(t, *err2)
	assert.Equal(t, i2, <-ctx.inc)
}

func TestWriteBufferPutAfterError(t *testing.T) {
	tErr := errors.New("test")
	ctx := makeWriteBufferContext(16, 1, makeBrokenWriteBufferConn(tErr))

	_, _, _, err := ctx.put([]byte{0, 1, 2, 3, 4, 5, 6, 7}, false)

	ctx.wg.Wait()
	assert.Equal(t, tErr, *err)

	_, _, _, err = ctx.put([]byte{0, 1, 2, 3, 4, 5, 6, 7}, false)

	ctx.wg.Wait()
	assert.Equal(t, errWriterBroken, *err)
}

func TestWriteBufferFlush(t *testing.T) {
	ctx := makeWriteBufferContext(16, 1, nil)

	_, p, b, err := ctx.put([]byte{0xde, 0xc0, 0xad, 0xde}, true)
	assert.Empty(t, p)

	ctx.w.flush()
	ctx.wg.Wait()

	assert.Equal(t, []byte{0xde, 0xc0, 0xad, 0xde}, *b)
	assert.NoError(t, *err)
}

func TestWriteBufferFlushWithError(t *testing.T) {
	tErr := errors.New("test")
	ctx := makeWriteBufferContext(16, 1, makeBrokenWriteBufferConn(tErr))

	_, p, b, err := ctx.put([]byte{0xde, 0xc0, 0xad, 0xde}, true)
	assert.Empty(t, p)

	ctx.w.flush()
	ctx.wg.Wait()

	assert.Empty(t, *b)
	assert.Equal(t, tErr, *err)
}

func TestWriteBufferFlushAfterError(t *testing.T) {
	tErr := errors.New("test")
	ctx := makeWriteBufferContext(16, 1, makeBrokenWriteBufferConn(tErr))

	_, p, b, err := ctx.put([]byte{0xde, 0xc0, 0xad, 0xde}, true)
	assert.Empty(t, p)

	ctx.w.flush()
	ctx.wg.Wait()

	assert.Empty(t, *b)
	assert.Equal(t, tErr, *err)

	i, p, b, err := ctx.startReceiver(true)

	ctx.w.out = append(ctx.w.out,
		append(
			[]byte{byte(msgIdxBytes + len(*b)), 0x00, 0x00, 0x00, byte(i), 0x00, 0x00, 0x00},
			*b...,
		)...,
	)
	ctx.w.idx = append(ctx.w.idx, i)

	ctx.w.flush()
	ctx.wg.Wait()

	assert.Empty(t, *b)
	assert.Equal(t, errWriterBroken, *err)
}

func TestWriteBufferFinalize(t *testing.T) {
	ctx := makeWriteBufferContext(16, 1, nil)

	i, p, b, err := ctx.put([]byte{0xde, 0xc0, 0xad, 0xde}, true)
	assert.Empty(t, p)

	ctx.w.finalize()
	ctx.wg.Wait()

	assert.Equal(t, []byte{0xde, 0xc0, 0xad, 0xde}, *b)
	assert.NoError(t, *err)
	assert.Equal(t, i, <-ctx.w.inc)

	select {
	default:
		assert.Fail(t, "channel inc hasn't been closed yet")

	case _, ok := <-ctx.w.inc:
		assert.False(t, ok)
	}
}

type writeBufferContext struct {
	w  *writeBuffer
	wg *sync.WaitGroup
	ps pipes
	inc chan int
}

func makeWriteBufferContext(n, q int, c net.Conn) writeBufferContext {
	out := writeBufferContext{
		wg:  new(sync.WaitGroup),
		ps:  makePipes(q),
		inc: make(chan int, q),
	}

	if c == nil {
		c = makeTestWriteBufferConn(out.ps)
	}

	out.w = newWriteBuffer(c, n, out.ps, out.inc)
	return out
}

func (ctx writeBufferContext) startReceiver(fill bool) (int, pipe, *[]byte, *error) {
	var (
		out []byte
		err error
	)

	if fill {
		out = []byte{0xff}
		err = errors.New("artificial")
	}

	i, p := ctx.ps.alloc()
	ctx.wg.Add(1)
	go func() {
		defer ctx.wg.Done()
		defer ctx.ps.free(i)

		out, err = p.get()
	}()

	return i, p, &out, &err
}

func (ctx writeBufferContext) put(in []byte, fill bool) (int, pipe, *[]byte, *error) {
	i, p, out, err := ctx.startReceiver(fill)

	ctx.w.put(request{
		i: i,
		b: in,
	})

	return i, p, out, err
}

type testWriteBufferConn struct {
	p pipes
}

func makeTestWriteBufferConn(p pipes) testWriteBufferConn {
	return testWriteBufferConn{
		p: p,
	}
}

func (c testWriteBufferConn) Write(b []byte) (int, error) {
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

func (c testWriteBufferConn) Read(b []byte) (int, error)         { panic("not implemented") }
func (c testWriteBufferConn) Close() error                       { panic("not implemented") }
func (c testWriteBufferConn) LocalAddr() net.Addr                { panic("not implemented") }
func (c testWriteBufferConn) RemoteAddr() net.Addr               { panic("not implemented") }
func (c testWriteBufferConn) SetDeadline(t time.Time) error      { panic("not implemented") }
func (c testWriteBufferConn) SetReadDeadline(t time.Time) error  { panic("not implemented") }
func (c testWriteBufferConn) SetWriteDeadline(t time.Time) error { panic("not implemented") }

type brokenWriteBufferConn struct {
	err error
}

func makeBrokenWriteBufferConn(err error) brokenWriteBufferConn {
	return brokenWriteBufferConn{
		err: err,
	}
}

func (c brokenWriteBufferConn) Write(b []byte) (int, error) {
	return 0, c.err
}

func (c brokenWriteBufferConn) Read(b []byte) (int, error)         { panic("not implemented") }
func (c brokenWriteBufferConn) Close() error                       { panic("not implemented") }
func (c brokenWriteBufferConn) LocalAddr() net.Addr                { panic("not implemented") }
func (c brokenWriteBufferConn) RemoteAddr() net.Addr               { panic("not implemented") }
func (c brokenWriteBufferConn) SetDeadline(t time.Time) error      { panic("not implemented") }
func (c brokenWriteBufferConn) SetReadDeadline(t time.Time) error  { panic("not implemented") }
func (c brokenWriteBufferConn) SetWriteDeadline(t time.Time) error { panic("not implemented") }
