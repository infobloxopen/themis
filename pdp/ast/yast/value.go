package yast

import (
	"net"
	"strings"

	"github.com/infobloxopen/go-trees/domaintree"
	"github.com/infobloxopen/go-trees/iptree"
	"github.com/infobloxopen/go-trees/strtree"
	"github.com/infobloxopen/themis/pdp"
)

func (ctx context) unmarshalStringValue(v interface{}) (pdp.AttributeValue, boundError) {
	s, err := ctx.validateString(v, "value of string type")
	if err != nil {
		return pdp.AttributeValue{}, err
	}

	return pdp.MakeStringValue(s), nil
}

func (ctx context) unmarshalIntegerValue(v interface{}) (pdp.AttributeValue, boundError) {
	n, err := ctx.validateInteger(v, "value of integer type")
	if err != nil {
		return pdp.AttributeValue{}, err
	}

	return pdp.MakeIntegerValue(n), nil
}

func (ctx context) unmarshalFloatValue(v interface{}) (pdp.AttributeValue, boundError) {
	n, err := ctx.validateFloat(v, "value of float type")
	if err != nil {
		return pdp.AttributeValue{}, err
	}

	return pdp.MakeFloatValue(n), nil
}

func (ctx context) unmarshalAddressValue(v interface{}) (pdp.AttributeValue, boundError) {
	s, err := ctx.validateString(v, "value of address type")
	if err != nil {
		return pdp.AttributeValue{}, err
	}

	a := net.ParseIP(s)
	if a == nil {
		return pdp.AttributeValue{}, newInvalidAddressError(s)
	}

	return pdp.MakeAddressValue(a), nil
}

func (ctx context) unmarshalNetworkValue(v interface{}) (pdp.AttributeValue, boundError) {
	s, err := ctx.validateString(v, "value of network type")
	if err != nil {
		return pdp.AttributeValue{}, err
	}

	_, n, ierr := net.ParseCIDR(s)
	if ierr != nil {
		return pdp.AttributeValue{}, newInvalidNetworkError(s, ierr)
	}

	return pdp.MakeNetworkValue(n), nil
}

func (ctx context) unmarshalDomainValue(v interface{}) (pdp.AttributeValue, boundError) {
	s, err := ctx.validateString(v, "value of domain type")
	if err != nil {
		return pdp.AttributeValue{}, err
	}

	d, ierr := domaintree.MakeWireDomainNameLower(s)
	if ierr != nil {
		return pdp.AttributeValue{}, newInvalidDomainError(s, ierr)
	}

	return pdp.MakeDomainValue(d), nil
}

func (ctx context) unmarshalSetOfStringsValueItem(v interface{}, i int, set *strtree.Tree) boundError {
	s, err := ctx.validateString(v, "element")
	if err != nil {
		return err
	}

	set.InplaceInsert(s, i)
	return nil
}

func (ctx context) unmarshalSetOfStringsValue(v interface{}) (pdp.AttributeValue, boundError) {
	items, err := ctx.validateList(v, "set of strings")
	if err != nil {
		return pdp.AttributeValue{}, err
	}

	set := strtree.NewTree()
	for i, item := range items {
		err = ctx.unmarshalSetOfStringsValueItem(item, i, set)
		if err != nil {
			return pdp.AttributeValue{}, bindError(bindErrorf(err, "%d", i), "set of strings")
		}
	}

	return pdp.MakeSetOfStringsValue(set), nil
}

func (ctx context) unmarshalSetOfNetworksValueItem(v interface{}, i int, set *iptree.Tree) boundError {
	s, err := ctx.validateString(v, "element")
	if err != nil {
		return err
	}

	_, n, ierr := net.ParseCIDR(s)
	if ierr != nil {
		return newInvalidNetworkError(s, ierr)
	}

	set.InplaceInsertNet(n, i)

	return nil
}

func (ctx context) unmarshalSetOfNetworksValue(v interface{}) (pdp.AttributeValue, boundError) {
	items, err := ctx.validateList(v, "set of networks")
	if err != nil {
		return pdp.AttributeValue{}, err
	}

	set := iptree.NewTree()
	for i, item := range items {
		err = ctx.unmarshalSetOfNetworksValueItem(item, i, set)
		if err != nil {
			return pdp.AttributeValue{}, bindError(bindErrorf(err, "%d", i), "set of networks")
		}
	}

	return pdp.MakeSetOfNetworksValue(set), nil
}

func (ctx context) unmarshalSetOfDomainsValueItem(v interface{}, i int, set *domaintree.Node) boundError {
	s, err := ctx.validateString(v, "element")
	if err != nil {
		return err
	}

	set.InplaceInsert(s, i)

	return nil
}

func (ctx context) unmarshalSetOfDomainsValue(v interface{}) (pdp.AttributeValue, boundError) {
	items, err := ctx.validateList(v, "")
	if err != nil {
		return pdp.AttributeValue{}, nil
	}

	set := &domaintree.Node{}
	for i, item := range items {
		err = ctx.unmarshalSetOfDomainsValueItem(item, i, set)
		if err != nil {
			return pdp.AttributeValue{}, bindError(bindErrorf(err, "%d", i), "set of domains")
		}
	}

	return pdp.MakeSetOfDomainsValue(set), nil
}

func (ctx context) unmarshalListOfStringsValueItem(v interface{}, list []string) ([]string, boundError) {
	s, err := ctx.validateString(v, "element")
	if err != nil {
		return list, err
	}

	return append(list, s), nil
}

func (ctx context) unmarshalListOfStringsValue(v interface{}) (pdp.AttributeValue, boundError) {
	items, err := ctx.validateList(v, "list of strings")
	if err != nil {
		return pdp.AttributeValue{}, err
	}

	list := []string{}
	for i, item := range items {
		list, err = ctx.unmarshalListOfStringsValueItem(item, list)
		if err != nil {
			return pdp.AttributeValue{}, bindError(bindErrorf(err, "%d", i), "list of strings")
		}
	}

	return pdp.MakeListOfStringsValue(list), nil
}

func (ctx context) unmarshalValueByType(t int, v interface{}) (pdp.AttributeValue, boundError) {
	switch t {
	case pdp.TypeString:
		return ctx.unmarshalStringValue(v)

	case pdp.TypeInteger:
		return ctx.unmarshalIntegerValue(v)

	case pdp.TypeFloat:
		return ctx.unmarshalFloatValue(v)

	case pdp.TypeAddress:
		return ctx.unmarshalAddressValue(v)

	case pdp.TypeNetwork:
		return ctx.unmarshalNetworkValue(v)

	case pdp.TypeDomain:
		return ctx.unmarshalDomainValue(v)

	case pdp.TypeSetOfStrings:
		return ctx.unmarshalSetOfStringsValue(v)

	case pdp.TypeSetOfNetworks:
		return ctx.unmarshalSetOfNetworksValue(v)

	case pdp.TypeSetOfDomains:
		return ctx.unmarshalSetOfDomainsValue(v)

	case pdp.TypeListOfStrings:
		return ctx.unmarshalListOfStringsValue(v)
	}

	return pdp.AttributeValue{}, newNotImplementedValueTypeError(t)
}

func (ctx context) unmarshalValue(v interface{}) (pdp.AttributeValue, boundError) {
	m, err := ctx.validateMap(v, "value attributes")
	if err != nil {
		return pdp.AttributeValue{}, err
	}

	strT, err := ctx.extractString(m, yastTagType, "type")
	if err != nil {
		return pdp.AttributeValue{}, err
	}

	t, ok := pdp.TypeIDs[strings.ToLower(strT)]
	if !ok {
		return pdp.AttributeValue{}, newUnknownTypeError(strT)
	}

	if t == pdp.TypeUndefined {
		return pdp.AttributeValue{}, newInvalidTypeError(t)
	}

	c, ok := m[yastTagContent]
	if !ok {
		return pdp.AttributeValue{}, newMissingContentError()
	}

	return ctx.unmarshalValueByType(t, c)
}
