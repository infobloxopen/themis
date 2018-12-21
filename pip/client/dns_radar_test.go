package client

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewDNSRadar(t *testing.T) {
	r := newDNSRadar("localhost:5600", time.Millisecond)
	assert.Equal(t, "localhost:5600", r.addr)
	assert.Equal(t, time.Millisecond, r.d)
	assert.NotZero(t, r.done)
}

func TestDNSRadarStartStop(t *testing.T) {
	r := newDNSRadar("localhost:5600", time.Millisecond)

	ch := r.start(nil)
	assert.NotZero(t, ch)

	r.stop()
	for range ch {
	}

	assert.Zero(t, r.done)

	r.stop()

	assert.Zero(t, r.start(nil))
}

func TestRunDNSRadar(t *testing.T) {
	done := make(chan struct{})
	tch := make(chan time.Time)
	ch := make(chan addrUpdate, 1024)

	go runDNSRadar(done, ch, tch, "localhost:5600", []string{"127.0.0.2:5600"})

	tch <- time.Now()
	close(done)

	for u := range ch {
		if assert.NoError(t, u.err) {
			if u.addr == "127.0.0.2:5600" {
				assert.Equal(t, addrUpdateOpDel, u.op)
			} else {
				assert.Equal(t, addrUpdateOpAdd, u.op)
			}
		}
	}
}

func TestLookupDNSRadar(t *testing.T) {
	ch := make(chan addrUpdate, 1024)

	idx := lookupDNSRadar(ch, nil, "localhost:5600")
	iAddrs := make([]string, 0, len(idx))
	for addr := range idx {
		iAddrs = append(iAddrs, addr)
	}

	cAddrs := make([]string, len(iAddrs))
	for i := 0; i < len(iAddrs); i++ {
		if assert.NotEmpty(t, ch) {
			u := <-ch
			if assert.NoError(t, u.err) && assert.Equal(t, addrUpdateOpAdd, u.op) {
				cAddrs[i] = u.addr
			}
		}
	}

	assert.ElementsMatch(t, iAddrs, cAddrs)
}

func TestLookupDNSRadarWithError(t *testing.T) {
	ch := make(chan addrUpdate, 1024)

	idx := lookupDNSRadar(ch, nil, ":::")
	assert.Empty(t, idx)
	select {
	default:
		assert.Fail(t, "no update")
	case u := <-ch:
		assert.Error(t, u.err)
	}
}

func TestDispatchDNSRadar(t *testing.T) {
	ch := make(chan addrUpdate, 1024)

	idx := dispatchDNSRadar(ch, map[string]struct{}{}, []string{"127.0.0.1:5600"})
	assert.Equal(t, map[string]struct{}{"127.0.0.1:5600": {}}, idx)
	assert.Equal(t, addrUpdate{
		op:   addrUpdateOpAdd,
		addr: "127.0.0.1:5600",
	}, <-ch)

	idx = dispatchDNSRadar(ch, map[string]struct{}{"127.0.0.1:5600": {}}, []string{})
	assert.Empty(t, idx)
	assert.Equal(t, addrUpdate{
		op:   addrUpdateOpDel,
		addr: "127.0.0.1:5600",
	}, <-ch)
}
