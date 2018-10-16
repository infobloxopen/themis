package client

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMakeDialerTK(t *testing.T) {
	d := makeDialerTK("tcp", defConnTimeout, defKeepAlive)
	assert.Equal(t, "tcp", d.n)
	if assert.NotZero(t, d.d) {
		assert.Equal(t, defConnTimeout, d.d.Timeout)
		assert.Equal(t, defKeepAlive, d.d.KeepAlive)
	}
}

func TestDialerTKDial(t *testing.T) {
	d := makeDialerTK("tcp", defConnTimeout, defKeepAlive)
	c, err := d.dial("/dev/null")
	assert.Zero(t, c)
	assert.Error(t, err)
}
