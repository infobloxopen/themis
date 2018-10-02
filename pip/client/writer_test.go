package client

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestWriter(t *testing.T) {
	wg := new(sync.WaitGroup)
	req := make(chan request)
	ps := makePipes(1)

	c := NewClient().(*client)

	wg.Add(1)
	go c.writer(wg, req, ps)

	i, p := ps.alloc()
	req <- request{
		i: i,
		b: []byte{0xef, 0xbe, 0xad, 0xde},
	}

	_, err := p.get()
	assert.NoError(t, err)

	close(req)
	wg.Wait()
}

func TestWriterNoTimeout(t *testing.T) {
	wg := new(sync.WaitGroup)
	req := make(chan request)
	ps := makePipes(1)

	c := NewClient(
		withTestWriteFlushChannel(make(chan time.Time)),
	).(*client)

	wg.Add(1)
	go c.writer(wg, req, ps)

	i, p := ps.alloc()
	req <- request{
		i: i,
		b: []byte{0xef, 0xbe, 0xad, 0xde},
	}
	close(req)

	wg.Wait()
	_, err := p.get()
	assert.NoError(t, err)
}
