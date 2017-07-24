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
	getValue(t int) (attributeValue, error)
	next(key attributeValue) (contentSubItem, error)
}

type contentStringMap struct {
	tree *strtree.Tree
}

func (m contentStringMap) describe() string {
	return "string map"
}

func (m contentStringMap) getValue(t int) (attributeValue, error) {
	return undefinedValue, newFinalContentSubitemError(m.describe())
}

func (m contentStringMap) next(key attributeValue) (contentSubItem, error) {
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

func (m contentNetworkMap) getValue(t int) (attributeValue, error) {
	return undefinedValue, newFinalContentSubitemError(m.describe())
}

func (m contentNetworkMap) next(key attributeValue) (contentSubItem, error) {
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

func (m contentDomainMap) getValue(t int) (attributeValue, error) {
	return undefinedValue, newFinalContentSubitemError(m.describe())
}

func (m contentDomainMap) next(key attributeValue) (contentSubItem, error) {
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

func (v contentValue) getValue(t int) (attributeValue, error) {
	switch t {
	case typeUndefined:
		panic(fmt.Errorf("Can't convert to value of undefined type"))

	case typeBoolean:
		return makeBooleanValue(v.value.(bool)), nil

	case typeString:
		return makeStringValue(v.value.(string)), nil

	case typeAddress:
		return makeAddressValue(v.value.(net.IP)), nil

	case typeNetwork:
		return makeNetworkValue(v.value.(*net.IPNet)), nil

	case typeDomain:
		return makeDomainValue(v.value.(string)), nil

	case typeSetOfStrings:
		return makeSetOfStringsValue(v.value.(*strtree.Tree)), nil

	case typeSetOfNetworks:
		return makeSetOfNetworksValue(v.value.(*iptree.Tree)), nil

	case typeSetOfDomains:
		return makeSetOfDomainsValue(v.value.(*domaintree.Node)), nil

	case typeListOfStrings:
		return makeListOfStringsValue(v.value.([]string)), nil
	}

	panic(fmt.Errorf("Can't convert to value of unknown type with index %d", t))
}

func (v contentValue) next(key attributeValue) (contentSubItem, error) {
	return nil, newMapContentSubitemError(v.describe())
}
