package client

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWithNetwork(t *testing.T) {
	var o options

	WithNetwork("unix")(&o)
	assert.Equal(t, "unix", o.net)
}

func TestWithAddress(t *testing.T) {
	var o options

	WithAddress("/tmp/unix.soket")(&o)
	assert.Equal(t, "/tmp/unix.soket", o.addr)
}
