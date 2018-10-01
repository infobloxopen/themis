package client

import "sync/atomic"

const (
	pipClientIdle uint32 = iota
	pipClientConnected
)

func (c *client) Connect() error {
	if !atomic.CompareAndSwapUint32(c.state, pipClientIdle, pipClientConnected) {
		return ErrorConnected
	}

	return nil
}

func (c *client) Close() {
	atomic.CompareAndSwapUint32(c.state, pipClientConnected, pipClientIdle)
}
