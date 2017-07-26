package pdp

import (
	"fmt"
	"net"

	"github.com/infobloxopen/go-trees/domaintree"
	"github.com/infobloxopen/go-trees/iptree"
	"github.com/infobloxopen/go-trees/strtree"
)

type contentItem struct {
	r contentSubItem
	t int
}

type contentSubItem interface {
	getValue(t int) (AttributeValue, error)
	next(key AttributeValue) (contentSubItem, error)
}

type contentStringMap struct {
	tree *strtree.Tree
}

func (m contentStringMap) describe() string {
	return "string map"
}

func (m contentStringMap) getValue(t int) (AttributeValue, error) {
	return undefinedValue, newFinalContentSubitemError(m.describe())
}

func (m contentStringMap) next(key AttributeValue) (contentSubItem, error) {
	s, err := key.str()
	if err != nil {
		return nil, err
	}

	v, ok := m.tree.Get(s)
	if !ok {
		return nil, &missingValueError{}
	}

	item, ok := v.(contentSubItem)
	if !ok {
		return contentValue{v}, nil
	}

	return item, nil
}

type contentNetworkMap struct {
	tree *iptree.Tree
}

func (m contentNetworkMap) describe() string {
	return "network map"
}

func (m contentNetworkMap) getValue(t int) (AttributeValue, error) {
	return undefinedValue, newFinalContentSubitemError(m.describe())
}

func (m contentNetworkMap) next(key AttributeValue) (contentSubItem, error) {
	a, err := key.address()
	if err != nil {
		return nil, err
	}

	v, ok := m.tree.GetByIP(a)
	if !ok {
		return nil, &missingValueError{}
	}

	item, ok := v.(contentSubItem)
	if !ok {
		return contentValue{v}, nil
	}

	return item, nil
}

type contentDomainMap struct {
	tree *domaintree.Node
}

func (m contentDomainMap) describe() string {
	return "domain map"
}

func (m contentDomainMap) getValue(t int) (AttributeValue, error) {
	return undefinedValue, newFinalContentSubitemError(m.describe())
}

func (m contentDomainMap) next(key AttributeValue) (contentSubItem, error) {
	d, err := key.domain()
	if err != nil {
		return nil, err
	}

	v, ok := m.tree.Get(d)
	if !ok {
		return nil, &missingValueError{}
	}

	item, ok := v.(contentSubItem)
	if !ok {
		return contentValue{v}, nil
	}

	return item, nil
}

type contentValue struct {
	value interface{}
}

func (v contentValue) describe() string {
	return fmt.Sprintf("value")
}

func (v contentValue) getValue(t int) (AttributeValue, error) {
	switch t {
	case TypeUndefined:
		panic(fmt.Errorf("Can't convert to value of undefined type"))

	case TypeBoolean:
		return MakeBooleanValue(v.value.(bool)), nil

	case TypeString:
		return MakeStringValue(v.value.(string)), nil

	case TypeAddress:
		return MakeAddressValue(v.value.(net.IP)), nil

	case TypeNetwork:
		return MakeNetworkValue(v.value.(*net.IPNet)), nil

	case TypeDomain:
		return MakeDomainValue(v.value.(string)), nil

	case TypeSetOfStrings:
		return MakeSetOfStringsValue(v.value.(*strtree.Tree)), nil

	case TypeSetOfNetworks:
		return MakeSetOfNetworksValue(v.value.(*iptree.Tree)), nil

	case TypeSetOfDomains:
		return MakeSetOfDomainsValue(v.value.(*domaintree.Node)), nil

	case TypeListOfStrings:
		return MakeListOfStringsValue(v.value.([]string)), nil
	}

	panic(fmt.Errorf("Can't convert to value of unknown type with index %d", t))
}

func (v contentValue) next(key AttributeValue) (contentSubItem, error) {
	return nil, newMapContentSubitemError(v.describe())
}
