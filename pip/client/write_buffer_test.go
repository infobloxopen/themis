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
	p := makePipes(1)

	w := newWriteBuffer(makeTestWriteBufferConn(p), 16, p)
	if assert.NotZero(t, w) {
		assert.Equal(t, 16, cap(w.out))
		assert.Empty(t, w.out)
		assert.NotZero(t, w.idx)
		assert.Equal(t, p, w.p)
	}
}

func TestWriteBufferRem(t *testing.T) {
	ps := makePipes(1)
	w := newWriteBuffer(makeTestWriteBufferConn(ps), 16, ps)

	assert.Equal(t, 16, w.rem())

	i, _ := ps.alloc()
	defer ps.free(i)

	w.put(request{
		i: i,
		b: []byte{0xde, 0xc0, 0xad, 0xde},
	})
	assert.Equal(t, 4, w.rem())
}

func TestWriteBufferPut(t *testing.T) {
	wg := new(sync.WaitGroup)
	p := makePipes(2)

	w := newWriteBuffer(makeTestWriteBufferConn(p), 16, p)

	i1, p1 := p.alloc()
	wg.Add(1)
	var b1 []byte
	err1 := errors.New("test1")
	go func() {
		defer wg.Done()
		defer p.free(i1)

		b1, err1 = p1.get()
	}()

	w.put(request{
		i: i1,
		b: []byte{0xde, 0xc0, 0xad, 0xde},
	})

	i2, p2 := p.alloc()
	wg.Add(1)
	var b2 []byte
	err2 := errors.New("test2")
	go func() {
		defer wg.Done()
		defer p.free(i2)

		b2, err2 = p2.get()
	}()

	w.put(request{
		i: i2,
		b: []byte{0xde, 0xc0, 0xad, 0xde, 0xef, 0xeb, 0, 0},
	})

	wg.Wait()
	assert.Equal(t, []byte{0xde, 0xc0, 0xad, 0xde}, b1)
	assert.NoError(t, err1)
	assert.Equal(t, []byte{0xde, 0xc0, 0xad, 0xde, 0xef, 0xeb, 0, 0}, b2)
	assert.NoError(t, err2)
}

func TestWriteBufferFlush(t *testing.T) {
	wg := new(sync.WaitGroup)
	ps := makePipes(1)

	w := newWriteBuffer(makeTestWriteBufferConn(ps), 16, ps)

	i, p := ps.alloc()
	wg.Add(1)
	var b []byte
	err := errors.New("test")
	go func() {
		defer wg.Done()
		defer ps.free(i)

		b, err = p.get()
	}()

	w.put(request{
		i: i,
		b: []byte{0xde, 0xc0, 0xad, 0xde},
	})
	assert.Empty(t, p)

	w.flush()

	wg.Wait()
	assert.Equal(t, []byte{0xde, 0xc0, 0xad, 0xde}, b)
	assert.NoError(t, err)
}

func TestWriteBufferFlushWithError(t *testing.T) {
	wg := new(sync.WaitGroup)
	ps := makePipes(1)

	tErr := errors.New("test")
	w := newWriteBuffer(makeBrokenWriteBufferConn(tErr), 16, ps)

	i, p := ps.alloc()
	wg.Add(1)
	b := []byte{0xde, 0xc0, 0xad, 0xde}
	err := errors.New("test")
	go func() {
		defer wg.Done()
		defer ps.free(i)

		b, err = p.get()
	}()

	w.put(request{
		i: i,
		b: []byte{0xde, 0xc0, 0xad, 0xde},
	})
	assert.Empty(t, p)

	w.flush()

	wg.Wait()
	assert.Empty(t, b)
	assert.Equal(t, tErr, err)
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
