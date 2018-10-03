package client

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestWithNetwork(t *testing.T) {
	var o options

	WithNetwork("unix")(&o)
	assert.Equal(t, "unix", o.net)
}

func TestWithAddress(t *testing.T) {
	var o options

	WithAddress("localhost:0")(&o)
	assert.Equal(t, "localhost:0", o.addr)
}

func TestWithMaxRequestSize(t *testing.T) {
	var o options

	WithMaxRequestSize(1024)(&o)
	assert.Equal(t, 1024, o.maxSize)

	WithMaxRequestSize(-1)(&o)
	assert.Equal(t, defMaxSize, o.maxSize)
}

func TestWithMaxQueue(t *testing.T) {
	var o options

	WithMaxQueue(1000)(&o)
	assert.Equal(t, 1000, o.maxQueue)

	WithMaxQueue(-1)(&o)
	assert.Equal(t, defMaxQueue, o.maxQueue)
}

func TestWithBufferSize(t *testing.T) {
	var o options

	WithBufferSize(1000)(&o)
	assert.Equal(t, 1000, o.bufSize)

	WithBufferSize(-1)(&o)
	assert.Equal(t, defBufSize, o.bufSize)
}

func TestWithWriteInterval(t *testing.T) {
	var o options

	WithWriteInterval(time.Second)(&o)
	assert.Equal(t, time.Second, o.writeInt)

	WithWriteInterval(-1 * time.Second)(&o)
	assert.Equal(t, defWriteInt, o.writeInt)
}

func TestWithTestWriteFlushChannel(t *testing.T) {
	var o options

	withTestWriteFlushChannel(make(chan time.Time))(&o)
	assert.NotZero(t, o.writeFlushCh)
}
