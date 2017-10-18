package jcon

import (
	"encoding/json"
	"net"

	"github.com/infobloxopen/go-trees/domaintree"
	"github.com/infobloxopen/go-trees/iptree"
	"github.com/infobloxopen/go-trees/strtree"
	"github.com/infobloxopen/themis/pdp"
)

type mapUnmarshaller interface {
	get() interface{}
	unmarshal(k string, d *json.Decoder) error
	postProcess(p Pair) error
}

func newTypedMap(c *contentItem, keyIdx int) (mapUnmarshaller, error) {
	t := c.k[keyIdx]

	switch t {
	case pdp.TypeString:
		return &stringMap{
			contentItemLink: contentItemLink{c: c, i: keyIdx},
			m:               strtree.NewTree()}, nil
	case pdp.TypeAddress, pdp.TypeNetwork:
		return &networkMap{
			contentItemLink: contentItemLink{c: c, i: keyIdx},
			m:               iptree.NewTree()}, nil
	case pdp.TypeDomain:
		return &domainMap{
			contentItemLink: contentItemLink{c: c, i: keyIdx},
			m:               &domaintree.Node{}}, nil
	}

	return nil, newInvalidContentKeyTypeError(t, pdp.ContentKeyTypes)
}

type contentItemLink struct {
	c *contentItem
	i int
}

type stringMap struct {
	contentItemLink
	m *strtree.Tree
}

func (m *stringMap) get() interface{} {
	return pdp.MakeContentStringMap(m.m)
}

func (m *stringMap) unmarshal(k string, d *json.Decoder) error {
	v, err := m.c.unmarshalTypedData(d, m.i+1)
	if err != nil {
		return bindError(err, k)
	}

	m.m.InplaceInsert(k, v)

	return nil
}

func (m *stringMap) postProcess(p Pair) error {
	v, err := m.c.postProcess(p.V, m.i+1)
	if err != nil {
		return bindError(err, p.K)
	}

	m.m.InplaceInsert(p.K, v)

	return nil
}

type networkMap struct {
	contentItemLink
	m *iptree.Tree
}

func (m *networkMap) unmarshal(k string, d *json.Decoder) error {
	var (
		n   *net.IPNet
		err error
	)

	a := net.ParseIP(k)
	if a == nil {
		_, n, err = net.ParseCIDR(k)
		if err != nil {
			return newAddressNetworkCastError(k, err)
		}
	}

	v, err := m.c.unmarshalTypedData(d, m.i+1)
	if err != nil {
		return bindError(err, k)
	}

	if a != nil {
		m.m.InplaceInsertIP(a, v)
	} else {
		m.m.InplaceInsertNet(n, v)
	}

	return nil
}

func (m *networkMap) postProcess(p Pair) error {
	var (
		n   *net.IPNet
		err error
	)

	a := net.ParseIP(p.K)
	if a == nil {
		_, n, err = net.ParseCIDR(p.K)
		if err != nil {
			return newAddressNetworkCastError(p.K, err)
		}
	}

	v, err := m.c.postProcess(p.V, m.i+1)
	if err != nil {
		return bindError(err, p.K)
	}

	if a != nil {
		m.m.InplaceInsertIP(a, v)
	} else {
		m.m.InplaceInsertNet(n, v)
	}

	return nil
}

func (m *networkMap) get() interface{} {
	return pdp.MakeContentNetworkMap(m.m)
}

type domainMap struct {
	contentItemLink
	m *domaintree.Node
}

func (m *domainMap) unmarshal(k string, d *json.Decoder) error {
	n, err := pdp.AdjustDomainName(k)
	if err != nil {
		return newDomainCastError(k, err)
	}

	v, err := m.c.unmarshalTypedData(d, m.i+1)
	if err != nil {
		return bindError(err, k)
	}

	m.m.InplaceInsert(n, v)

	return nil
}

func (m *domainMap) postProcess(p Pair) error {
	n, err := pdp.AdjustDomainName(p.K)
	if err != nil {
		return newDomainCastError(p.K, err)
	}

	v, err := m.c.postProcess(p.V, m.i+1)
	if err != nil {
		return bindError(err, p.K)
	}

	m.m.InplaceInsert(n, v)

	return nil
}

func (m *domainMap) get() interface{} {
	return pdp.MakeContentDomainMap(m.m)
}
