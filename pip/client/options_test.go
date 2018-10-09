package client

import (
	"net"
	"reflect"
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

func TestWithConnErrHandler(t *testing.T) {
	var o options

	f := func(net.Addr, error) {}
	WithConnErrHandler(f)(&o)
	assert.Equal(t, reflect.ValueOf(f).Pointer(), reflect.ValueOf(o.onErr).Pointer())
}

func TestWithWriteInterval(t *testing.T) {
	var o options

	WithWriteInterval(time.Second)(&o)
	assert.Equal(t, time.Second, o.writeInt)

	WithWriteInterval(-1 * time.Second)(&o)
	assert.Equal(t, defWriteInt, o.writeInt)
}

func TestWithResponseTimeout(t *testing.T) {
	var o options

	WithResponseTimeout(time.Second)(&o)
	assert.Equal(t, time.Second, o.timeout)

	WithResponseTimeout(-1 * time.Second)(&o)
	assert.Equal(t, defTimeout, o.timeout)
}

func TestWithResponseCheckInterval(t *testing.T) {
	var o options

	WithResponseCheckInterval(time.Second)(&o)
	assert.Equal(t, time.Second, o.termInt)

	WithResponseCheckInterval(-1 * time.Second)(&o)
	assert.Equal(t, defTermInt, o.termInt)
}

func TestWithTestWriteFlushChannel(t *testing.T) {
	var o options

	withTestWriteFlushChannel(make(chan time.Time))(&o)
	assert.NotZero(t, o.writeFlushCh)
}

func TestWithTestTermFlushChannel(t *testing.T) {
	var o options

	withTestTermFlushChannel(make(chan time.Time))(&o)
	assert.NotZero(t, o.termFlushCh)
}
