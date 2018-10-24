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

	WithNetwork(unixNet)(&o)
	assert.Equal(t, unixNet, o.net)
}

func TestWithAddress(t *testing.T) {
	var o options

	WithAddress("localhost:0")(&o)
	assert.Equal(t, "localhost:0", o.addr)
}

func TestWithRoundRobinBalancer(t *testing.T) {
	var o options

	WithRoundRobinBalancer()(&o)
	assert.Equal(t, balancerTypeRoundRobin, o.balancer)
	assert.Empty(t, o.addrs)

	WithRoundRobinBalancer("127.0.0.1:5600", "[::1]:5600")(&o)
	assert.Equal(t, balancerTypeRoundRobin, o.balancer)
	assert.Equal(t, []string{"127.0.0.1:5600", "[::1]:5600"}, o.addrs)
}

func TestWithHotSpotBalancer(t *testing.T) {
	var o options

	WithHotSpotBalancer()(&o)
	assert.Equal(t, balancerTypeHotSpot, o.balancer)
	assert.Empty(t, o.addrs)

	WithHotSpotBalancer("127.0.0.1:5600", "[::1]:5600")(&o)
	assert.Equal(t, balancerTypeHotSpot, o.balancer)
	assert.Equal(t, []string{"127.0.0.1:5600", "[::1]:5600"}, o.addrs)
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

func TestWithConnTimeout(t *testing.T) {
	var o options

	WithConnTimeout(time.Second)(&o)
	assert.Equal(t, time.Second, o.connTimeout)
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

func TestMakeOptions(t *testing.T) {
	o := makeOptions([]Option{
		WithNetwork(unixNet),
		WithAddress("/var/run/pip.socket"),
	})

	assert.Equal(t, unixNet, o.net)
	assert.Equal(t, "/var/run/pip.socket", o.addr)
	assert.Equal(t, defMaxSize, o.maxSize)
}

func TestMakeOptionsWithTooSmallBufSize(t *testing.T) {
	o := makeOptions([]Option{
		WithBufferSize(1024),
	})
	assert.Equal(t, defBufSize, o.bufSize)

	o = makeOptions([]Option{
		WithMaxRequestSize(256),
		WithBufferSize(1024),
	})
	assert.Equal(t, 1024, o.bufSize)
}

func TestMakeOptionsWithDefRadarInt(t *testing.T) {
	o := makeOptions([]Option{
		WithDNSRadar(),
	})
	assert.Equal(t, defDNSRadarInt, o.radarInt)

	o = makeOptions([]Option{
		WithK8sRadar(),
	})
	assert.Equal(t, defK8sRadarInt, o.radarInt)
}
