package server

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

	WithAddress("localhost:5555")(&o)
	assert.Equal(t, "localhost:5555", o.addr)
}
