package client

import (
	"testing"

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
