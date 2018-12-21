package client

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	core "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func TestNewK8sRadar(t *testing.T) {
	r, err := newK8sRadar("pip.app.namespace:5600", fake.NewSimpleClientset(), time.Minute)
	assert.NoError(t, err)
	if assert.NotZero(t, r) {
		r.Lock()
		assert.False(t, r.started)
		assert.NotZero(t, r.done)
		assert.NotZero(t, r.sii)
		r.Unlock()
	}
}

func TestNewK8sRadarWithInvalidAddress(t *testing.T) {
	r, err := newK8sRadar(":::", fake.NewSimpleClientset(), time.Minute)
	if assert.Error(t, err) {
		assert.Equal(t, "address :::: too many colons in address", err.Error())
	}
	assert.Zero(t, r)

	r, err = newK8sRadar("app.namespace:5600", fake.NewSimpleClientset(), time.Minute)
	assert.Equal(t, errK8sNameTooShort, err)
	assert.Zero(t, r)
}

func TestK8sRadarStartStop(t *testing.T) {
	r, err := newK8sRadar("pip.app.namespace:5600", fake.NewSimpleClientset(), time.Minute)
	assert.NoError(t, err)
	if assert.NotZero(t, r) {
		r.stop()
		r.Lock()
		assert.NotZero(t, r.done)
		r.Unlock()

		ch := r.start(nil)
		r.Lock()
		assert.True(t, r.started)
		r.Unlock()

		assert.Zero(t, r.start(nil))

		r.stop()
		_, ok := <-ch
		r.Lock()
		assert.False(t, ok)
		assert.False(t, r.started)
		r.Unlock()

		r.stop()

		assert.Zero(t, r.start(nil))
	}
}

func TestK8sRadarOnAdd(t *testing.T) {
	r, err := newK8sRadar("pip.app.namespace:5600", fake.NewSimpleClientset(), time.Minute)
	assert.NoError(t, err)
	if assert.NotZero(t, r) {
		ch := make(chan addrUpdate, 1024)
		if f := r.OnAdd(ch); assert.NotZero(t, f) {
			f(makeTestK8sPod(true, "127.0.0.1", "app", "pip"))
			assert.Equal(t, addrUpdate{
				op:   addrUpdateOpAdd,
				addr: "127.0.0.1:5600",
			}, <-ch)
		}
	}
}

func TestK8sRadarOnUpdate(t *testing.T) {
	r, err := newK8sRadar("pip.app.namespace:5600", fake.NewSimpleClientset(), time.Minute)
	assert.NoError(t, err)
	if assert.NotZero(t, r) {
		ch := make(chan addrUpdate, 1024)
		if f := r.OnUpdate(ch); assert.NotZero(t, f) {
			f(makeTestK8sPod(false, "127.0.0.1", "app", "pip"), makeTestK8sPod(true, "127.0.0.1", "app", "pip"))
			assert.Equal(t, addrUpdate{
				op:   addrUpdateOpAdd,
				addr: "127.0.0.1:5600",
			}, <-ch)

			f(makeTestK8sPod(true, "127.0.0.1", "app", "pip"), makeTestK8sPod(false, "127.0.0.1", "app", "pip"))
			assert.Equal(t, addrUpdate{
				op:   addrUpdateOpDel,
				addr: "127.0.0.1:5600",
			}, <-ch)

			f(makeTestK8sPod(true, "127.0.0.1", "app", "pip"), makeTestK8sPod(true, "127.0.0.1", "app", "other"))
			assert.Equal(t, addrUpdate{
				op:   addrUpdateOpDel,
				addr: "127.0.0.1:5600",
			}, <-ch)
		}
	}
}

func TestK8sRadarOnDelete(t *testing.T) {
	r, err := newK8sRadar("pip.app.namespace:5600", fake.NewSimpleClientset(), time.Minute)
	assert.NoError(t, err)
	if assert.NotZero(t, r) {
		ch := make(chan addrUpdate, 1024)
		if f := r.OnDelete(ch); assert.NotZero(t, f) {
			f(makeTestK8sPod(false, "127.0.0.1", "app", "pip"))
			assert.Equal(t, addrUpdate{
				op:   addrUpdateOpDel,
				addr: "127.0.0.1:5600",
			}, <-ch)
		}
	}
}

func makeTestK8sPod(ready bool, ip string, s ...string) *core.Pod {
	p := new(core.Pod)
	p.Labels = make(map[string]string)
	for i := 0; i < len(s)/2; i++ {
		p.Labels[s[2*i]] = s[2*i+1]
	}

	if ready {
		p.Status.Conditions = []core.PodCondition{
			{
				Type:   core.PodReady,
				Status: core.ConditionTrue,
			},
		}
	}

	p.Status.PodIP = ip

	return p
}
