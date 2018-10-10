package client

import (
	"fmt"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestLookupHostPort(t *testing.T) {
	addrs, err := lookupHostPort("localhost:5600")
	assert.NoError(t, err)
	if len(addrs) > 1 {
		assert.ElementsMatch(t, []string{"127.0.0.1:5600", "[::1]:5600"}, addrs)
	} else {
		assert.Equal(t, []string{"127.0.0.1:5600"}, addrs)
	}
}

func TestLookupHostPortNoPort(t *testing.T) {
	addrs, err := lookupHostPort("127.0.0.1")
	assert.NoError(t, err)
	assert.Equal(t, []string{"127.0.0.1:" + defPort}, addrs)
}

func TestLookupHostPortInvalidAddress(t *testing.T) {
	addrs, err := lookupHostPort("127.0.0.1::5600")
	assert.Error(t, err, "got addresses: %#v", addrs)
}

func TestLookupHostPortUnknownAddress(t *testing.T) {
	addrs, err := lookupHostPort("example.zone-which-should-not-exist:5600")
	assert.Error(t, err, "got addresses: %#v", addrs)
}

func TestJoinAddrsPort(t *testing.T) {
	addrs := joinAddrsPort([]string{
		"127.0.0.1",
		"::1",
		"localhost",
	}, defPort)
	assert.ElementsMatch(t, []string{"127.0.0.1:5600", "[::1]:5600"}, addrs)
}

func TestDialTimeout(t *testing.T) {
	wg := new(sync.WaitGroup)
	defer wg.Wait()

	sc1, err := net.Listen(defNet, "127.0.0.1:5601")
	if !assert.NoError(t, err) {
		assert.FailNow(t, "failed to start server listener")
	}
	defer sc1.Close()

	wg.Add(1)
	go ignoreIncommingConnections(wg, sc1)

	sc2, err := net.Listen(defNet, "127.0.0.1:5602")
	if !assert.NoError(t, err) {
		assert.FailNow(t, "failed to start server listener")
	}
	defer sc2.Close()

	wg.Add(1)
	go ignoreIncommingConnections(wg, sc2)

	conns, err := dialTimeout(defNet, []string{"127.0.0.1:5601", "127.0.0.1:5602"}, time.Second)
	if assert.NoError(t, err) {
		for i, c := range conns {
			if assert.NotZero(t, c, "connection: %d", i+1) {
				c.Close()
			}
		}
	}
}

func TestDialTimeoutNoConnection(t *testing.T) {
	conns, err := dialTimeout(defNet, []string{"127.0.0.1:5601", "127.0.0.1:5602"}, time.Second)
	assert.Error(t, err, "got connections: %#v", conns)
}

func ignoreIncommingConnections(wg *sync.WaitGroup, sc net.Listener) {
	defer wg.Done()

	for {
		c, err := sc.Accept()
		if err != nil {
			if !isConnClosed(err) {
				panic(fmt.Errorf("failed to listen at %s: %s", sc.Addr(), err))
			}

			return
		}

		c.Close()
	}
}
