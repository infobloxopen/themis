package server

import (
	"math"
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

	WithAddress("localhost:5555")(&o)
	assert.Equal(t, "localhost:5555", o.addr)
}

func TestWithMaxConnections(t *testing.T) {
	var o options

	WithMaxConnections(5)(&o)
	assert.Equal(t, 5, o.maxConn)

	WithMaxConnections(-1)(&o)
	assert.Equal(t, 0, o.maxConn)
}

func TestWithConnErrHandler(t *testing.T) {
	var o options

	f := func(net.Addr, error) {}
	WithConnErrHandler(f)(&o)
	assert.Equal(t, reflect.ValueOf(f).Pointer(), reflect.ValueOf(o.onErr).Pointer())
}

func TestWithBufferSize(t *testing.T) {
	var o options

	WithBufferSize(5)(&o)
	assert.Equal(t, 5, o.bufSize)

	WithBufferSize(0)(&o)
	assert.Equal(t, defBufSize, o.bufSize)
}

func TestWithMaxMessageSize(t *testing.T) {
	var o options

	WithMaxMessageSize(5)(&o)
	assert.Equal(t, 5, o.maxMsgSize)

	WithMaxMessageSize(0)(&o)
	assert.Equal(t, defMaxMsgSize, o.maxMsgSize)

	above := math.MaxUint32 + 1
	if above > math.MaxUint32 {
		WithMaxMessageSize(above)(&o)
		assert.Equal(t, defMaxMsgSize, o.maxMsgSize)
	}
}

func TestWithWriteInterval(t *testing.T) {
	var o options

	WithWriteInterval(time.Second)(&o)
	assert.Equal(t, time.Second, o.writeInt)
}
