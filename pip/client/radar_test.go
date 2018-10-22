package client

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
)

func TestNewAddressesAndRadar(t *testing.T) {
	c := NewClient().(*client)
	addrs, r, err := c.newAddressesAndRadar()
	assert.Equal(t, []string{defAddr}, addrs)
	assert.Zero(t, r)
	assert.NoError(t, err)

	c = NewClient(
		WithNetwork(unixNet),
		WithAddress("/var/run/pip.socket"),
		WithRoundRobinBalancer(
			"/var/run/pip.1.socket",
			"/var/run/pip.2.socket",
			"/var/run/pip.3.socket",
		),
	).(*client)
	addrs, r, err = c.newAddressesAndRadar()
	assert.Equal(t, []string{"/var/run/pip.socket"}, addrs)
	assert.Zero(t, r)
	assert.NoError(t, err)

	c = NewClient(
		WithRoundRobinBalancer(
			"127.0.0.1:5601",
			"127.0.0.1:5602",
			"127.0.0.1:5603",
		),
	).(*client)
	addrs, r, err = c.newAddressesAndRadar()
	assert.Equal(t, []string{
		"127.0.0.1:5601",
		"127.0.0.1:5602",
		"127.0.0.1:5603",
	}, addrs)
	assert.Zero(t, r)
	assert.NoError(t, err)

	if lAddrs, err := lookupHostPort(defAddr); assert.NoError(t, err) {
		c = NewClient(
			WithRoundRobinBalancer(),
		).(*client)
		addrs, r, err = c.newAddressesAndRadar()
		assert.ElementsMatch(t, lAddrs, addrs)
		assert.Zero(t, r)
		assert.NoError(t, err)
	}

	c = NewClient(
		WithRoundRobinBalancer(
			"127.0.0.1:5600",
			"127.0.0.2:5600",
			"127.0.0.3:5600",
		),
		WithDNSRadar(),
	).(*client)
	addrs, r, err = c.newAddressesAndRadar()
	assert.Equal(t, []string{
		"127.0.0.1:5600",
		"127.0.0.2:5600",
		"127.0.0.3:5600",
	}, addrs)
	assert.IsType(t, &dNSRadar{}, r)
	assert.NoError(t, err)
}

func TestNewRadarWithDNS(t *testing.T) {
	c := NewClient(
		WithDNSRadar(),
	).(*client)

	r, err := c.newRadar()
	assert.IsType(t, &dNSRadar{}, r)
	assert.NoError(t, err)
}

func TestNewRadarWithK8s(t *testing.T) {
	c := NewClient(
		WithK8sRadar(),
		WithAddress("value.key.namespace:5600"),
		withTestK8sClient(func() (kubernetes.Interface, error) {
			return fake.NewSimpleClientset(), nil
		}),
	).(*client)

	r, err := c.newRadar()
	assert.IsType(t, &k8sRadar{}, r)
	assert.NoError(t, err)
}

func TestNewRadarWithK8sWithBrokenClientMaker(t *testing.T) {
	tErr := errors.New("test")

	c := NewClient(
		WithK8sRadar(),
		WithAddress("value.key.namespace:5600"),
		withTestK8sClient(func() (kubernetes.Interface, error) {
			return nil, tErr
		}),
	).(*client)

	r, err := c.newRadar()
	assert.Zero(t, r)
	assert.Equal(t, tErr, err)
}

func TestNewRadarWithK8sWithShortName(t *testing.T) {
	c := NewClient(
		WithK8sRadar(),
		WithAddress("key.namespace:5600"),
		withTestK8sClient(func() (kubernetes.Interface, error) {
			return fake.NewSimpleClientset(), nil
		}),
	).(*client)

	r, err := c.newRadar()
	assert.Zero(t, r)
	assert.Equal(t, errK8sNameTooShort, err)
}
