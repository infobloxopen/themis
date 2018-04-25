package jcon

import (
	"encoding/json"
	"net"

	"github.com/infobloxopen/go-trees/domaintree"
	"github.com/infobloxopen/go-trees/domaintree16"
	"github.com/infobloxopen/go-trees/domaintree32"
	"github.com/infobloxopen/go-trees/domaintree64"
	"github.com/infobloxopen/go-trees/domaintree8"
	"github.com/infobloxopen/go-trees/iptree"
	"github.com/infobloxopen/go-trees/strtree"
	"github.com/infobloxopen/themis/jparser"
	"github.com/infobloxopen/themis/pdp"
)

type mapUnmarshaller interface {
	get() interface{}
	unmarshal(k string, d *json.Decoder) error
	postProcess(p jparser.Pair) error
}

func newTypedMap(c *contentItem, keyIdx int) (mapUnmarshaller, error) {
	t := c.k[keyIdx]

	switch t {
	case pdp.TypeString:
		if _, ok := c.t.(*pdp.FlagsType); ok {
			return nil, newInvalidContentValueTypeError(c.t)
		}

		return &stringMap{
			contentItemLink: contentItemLink{c: c, i: keyIdx},
			m:               strtree.NewTree()}, nil
	case pdp.TypeAddress, pdp.TypeNetwork:
		if _, ok := c.t.(*pdp.FlagsType); ok {
			return nil, newInvalidContentValueTypeError(c.t)
		}

		return &networkMap{
			contentItemLink: contentItemLink{c: c, i: keyIdx},
			m:               iptree.NewTree()}, nil
	case pdp.TypeDomain:
		if t, ok := c.t.(*pdp.FlagsType); ok {
			switch t.Capacity() {
			case 8:
				return &domain8Map{
					contentItemLink: contentItemLink{c: c, i: keyIdx},
					m:               &domaintree8.Node{}}, nil

			case 16:
				return &domain16Map{
					contentItemLink: contentItemLink{c: c, i: keyIdx},
					m:               &domaintree16.Node{}}, nil

			case 32:
				return &domain32Map{
					contentItemLink: contentItemLink{c: c, i: keyIdx},
					m:               &domaintree32.Node{}}, nil
			}

			return &domain64Map{
				contentItemLink: contentItemLink{c: c, i: keyIdx},
				m:               &domaintree64.Node{}}, nil
		}

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

func (m *stringMap) postProcess(p jparser.Pair) error {
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

func (m *networkMap) postProcess(p jparser.Pair) error {
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
	v, err := m.c.unmarshalTypedData(d, m.i+1)
	if err != nil {
		return bindError(err, k)
	}

	m.m.InplaceInsert(k, v)

	return nil
}

func (m *domainMap) postProcess(p jparser.Pair) error {
	v, err := m.c.postProcess(p.V, m.i+1)
	if err != nil {
		return bindError(err, p.K)
	}

	m.m.InplaceInsert(p.K, v)

	return nil
}

func (m *domainMap) get() interface{} {
	return pdp.MakeContentDomainMap(m.m)
}

type domain8Map struct {
	contentItemLink
	m *domaintree8.Node
}

func (m *domain8Map) unmarshal(k string, d *json.Decoder) error {
	v, err := m.c.unmarshalFlags8Value(d)
	if err != nil {
		return bindError(err, k)
	}

	m.m.InplaceInsert(k, v)

	return nil
}

func (m *domain8Map) postProcess(p jparser.Pair) error {
	v, err := m.c.postProcessFlags8Value(p.V)
	if err != nil {
		return bindError(err, p.K)
	}

	m.m.InplaceInsert(p.K, v)

	return nil
}

func (m *domain8Map) get() interface{} {
	return pdp.MakeContentDomainFlags8Map(m.m)
}

type domain16Map struct {
	contentItemLink
	m *domaintree16.Node
}

func (m *domain16Map) unmarshal(k string, d *json.Decoder) error {
	v, err := m.c.unmarshalFlags16Value(d)
	if err != nil {
		return bindError(err, k)
	}

	m.m.InplaceInsert(k, v)

	return nil
}

func (m *domain16Map) postProcess(p jparser.Pair) error {
	v, err := m.c.postProcessFlags16Value(p.V)
	if err != nil {
		return bindError(err, p.K)
	}

	m.m.InplaceInsert(p.K, v)

	return nil
}

func (m *domain16Map) get() interface{} {
	return pdp.MakeContentDomainFlags16Map(m.m)
}

type domain32Map struct {
	contentItemLink
	m *domaintree32.Node
}

func (m *domain32Map) unmarshal(k string, d *json.Decoder) error {
	v, err := m.c.unmarshalFlags32Value(d)
	if err != nil {
		return bindError(err, k)
	}

	m.m.InplaceInsert(k, v)

	return nil
}

func (m *domain32Map) postProcess(p jparser.Pair) error {
	v, err := m.c.postProcessFlags32Value(p.V)
	if err != nil {
		return bindError(err, p.K)
	}

	m.m.InplaceInsert(p.K, v)

	return nil
}

func (m *domain32Map) get() interface{} {
	return pdp.MakeContentDomainFlags32Map(m.m)
}

type domain64Map struct {
	contentItemLink
	m *domaintree64.Node
}

func (m *domain64Map) unmarshal(k string, d *json.Decoder) error {
	v, err := m.c.unmarshalFlags64Value(d)
	if err != nil {
		return bindError(err, k)
	}

	m.m.InplaceInsert(k, v)

	return nil
}

func (m *domain64Map) postProcess(p jparser.Pair) error {
	v, err := m.c.postProcessFlags64Value(p.V)
	if err != nil {
		return bindError(err, p.K)
	}

	m.m.InplaceInsert(p.K, v)

	return nil
}

func (m *domain64Map) get() interface{} {
	return pdp.MakeContentDomainFlags64Map(m.m)
}
