package jast

import (
	"net"
	"strings"

	"github.com/infobloxopen/go-trees/domaintree"
	"github.com/infobloxopen/go-trees/iptree"
	"github.com/infobloxopen/go-trees/strtree"
	"github.com/infobloxopen/themis/pdp"
	"math"
)

func (ctx context) unmarshalStringValueObject(v interface{}) (pdp.AttributeValue, boundError) {
	s, err := ctx.validateString(v, "value of string type")
	if err != nil {
		return pdp.AttributeValue{}, err
	}

	return pdp.MakeStringValue(s), nil
}

func (ctx context) unmarshalIntegerValueObject(v interface{}) (pdp.AttributeValue, boundError) {
	n, err := ctx.validateInteger(v, "value of integer type")
	if err != nil {
		return pdp.AttributeValue{}, err
	}

	return pdp.MakeIntegerValue(n), nil
}

func (ctx context) unmarshalFloatValueObject(v interface{}) (pdp.AttributeValue, boundError) {
	n, err := ctx.validateFloat(v, "value of float type")
	if err != nil {
		return pdp.AttributeValue{}, err
	}

	return pdp.MakeFloatValue(n), nil
}

func (ctx context) unmarshalAddressValueObject(v interface{}) (pdp.AttributeValue, boundError) {
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

func (ctx context) unmarshalNetworkValueObject(v interface{}) (pdp.AttributeValue, boundError) {
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

func (ctx context) unmarshalDomainValueObject(v interface{}) (pdp.AttributeValue, boundError) {
	s, err := ctx.validateString(v, "value of domain type")
	if err != nil {
		return pdp.AttributeValue{}, err
	}

	d, ierr := pdp.AdjustDomainName(s)
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

func (ctx context) unmarshalSetOfStringsValueObject(v interface{}) (pdp.AttributeValue, boundError) {
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

func (ctx context) unmarshalSetOfNetworksValueItemObject(v interface{}, i int, set *iptree.Tree) boundError {
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

func (ctx context) unmarshalSetOfNetworksValueObject(v interface{}) (pdp.AttributeValue, boundError) {
	items, err := ctx.validateList(v, "set of networks")
	if err != nil {
		return pdp.AttributeValue{}, err
	}

	set := iptree.NewTree()
	for i, item := range items {
		err = ctx.unmarshalSetOfNetworksValueItemObject(item, i, set)
		if err != nil {
			return pdp.AttributeValue{}, bindError(bindErrorf(err, "%d", i), "set of networks")
		}
	}

	return pdp.MakeSetOfNetworksValue(set), nil
}

func (ctx context) unmarshalSetOfDomainsValueItemObject(v interface{}, i int, set *domaintree.Node) boundError {
	s, err := ctx.validateString(v, "element")
	if err != nil {
		return err
	}

	d, ierr := pdp.AdjustDomainName(s)
	if ierr != nil {
		return newInvalidDomainError(s, ierr)
	}

	set.InplaceInsert(d, i)

	return nil
}

func (ctx context) unmarshalSetOfDomainsValueObject(v interface{}) (pdp.AttributeValue, boundError) {
	items, err := ctx.validateList(v, "")
	if err != nil {
		return pdp.AttributeValue{}, nil
	}

	set := &domaintree.Node{}
	for i, item := range items {
		err = ctx.unmarshalSetOfDomainsValueItemObject(item, i, set)
		if err != nil {
			return pdp.AttributeValue{}, bindError(bindErrorf(err, "%d", i), "set of domains")
		}
	}

	return pdp.MakeSetOfDomainsValue(set), nil
}

func (ctx context) unmarshalListOfStringsValueItemObject(v interface{}, list []string) ([]string, boundError) {
	s, err := ctx.validateString(v, "element")
	if err != nil {
		return list, err
	}

	return append(list, s), nil
}

func (ctx context) unmarshalListOfStringsValueObject(v interface{}) (pdp.AttributeValue, boundError) {
	items, err := ctx.validateList(v, "list of strings")
	if err != nil {
		return pdp.AttributeValue{}, err
	}

	list := []string{}
	for i, item := range items {
		list, err = ctx.unmarshalListOfStringsValueItemObject(item, list)
		if err != nil {
			return pdp.AttributeValue{}, bindError(bindErrorf(err, "%d", i), "list of strings")
		}
	}

	return pdp.MakeListOfStringsValue(list), nil
}

func (ctx context) unmarshalValueByTypeObject(t int, v interface{}) (pdp.AttributeValue, boundError) {
	switch t {
	case pdp.TypeString:
		return ctx.unmarshalStringValueObject(v)

	case pdp.TypeInteger:
		return ctx.unmarshalIntegerValueObject(v)

	case pdp.TypeFloat:
		return ctx.unmarshalFloatValueObject(v)

	case pdp.TypeAddress:
		return ctx.unmarshalAddressValueObject(v)

	case pdp.TypeNetwork:
		return ctx.unmarshalNetworkValueObject(v)

	case pdp.TypeDomain:
		return ctx.unmarshalDomainValueObject(v)

	case pdp.TypeSetOfStrings:
		return ctx.unmarshalSetOfStringsValueObject(v)

	case pdp.TypeSetOfNetworks:
		return ctx.unmarshalSetOfNetworksValueObject(v)

	case pdp.TypeSetOfDomains:
		return ctx.unmarshalSetOfDomainsValueObject(v)

	case pdp.TypeListOfStrings:
		return ctx.unmarshalListOfStringsValueObject(v)
	}

	return pdp.AttributeValue{}, newNotImplementedValueTypeError(t)
}

func (ctx context) unmarshalValueObject(v interface{}) (pdp.AttributeValue, boundError) {
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

	return ctx.unmarshalValueByTypeObject(t, c)
}

func (ctx context) validateString(v interface{}, desc string) (string, boundError) {
	r, ok := v.(string)
	if !ok {
		return "", newStringError(v, desc)
	}

	return r, nil
}

func (ctx context) extractString(m map[interface{}]interface{}, k string, desc string) (string, boundError) {
	v, ok := m[k]
	if !ok {
		return "", newMissingStringError(desc)
	}

	return ctx.validateString(v, desc)
}

func (ctx context) extractStringOpt(m map[interface{}]interface{}, k string, desc string) (string, bool, boundError) {
	v, ok := m[k]
	if !ok {
		return "", false, nil
	}

	s, err := ctx.validateString(v, desc)
	return s, true, err
}

func (ctx context) validateInteger(v interface{}, desc string) (int64, boundError) {
	switch v := v.(type) {
	case int:
		return int64(v), nil

	case int64:
		return v, nil

	case uint64:
		if v > math.MaxInt64 {
			return 0, newIntegerUint64OverflowError(v, desc)
		}

		return int64(v), nil

	case float64:
		if v < -9007199254740992 || v > 9007199254740992 {
			return 0, newIntegerFloat64OverflowError(v, desc)
		}

		return int64(v), nil
	}

	return 0, newIntegerError(v, desc)
}

func (ctx context) validateFloat(v interface{}, desc string) (float64, boundError) {
	switch v := v.(type) {
	case int:
		return float64(v), nil

	case int64:
		return float64(v), nil

	case uint64:
		return float64(v), nil

	case float64:
		return float64(v), nil
	}

	return 0, newFloatError(v, desc)
}

func (ctx context) validateMap(v interface{}, desc string) (map[interface{}]interface{}, boundError) {
	r, ok := v.(map[interface{}]interface{})
	if !ok {
		return nil, newMapError(v, desc)
	}

	return r, nil
}

func (ctx context) extractMap(m map[interface{}]interface{}, k string, desc string) (map[interface{}]interface{}, boundError) {
	v, ok := m[k]
	if !ok {
		return nil, newMissingMapError(desc)
	}

	return ctx.validateMap(v, desc)
}

func (ctx context) extractMapOpt(m map[interface{}]interface{}, k string, desc string) (map[interface{}]interface{}, bool, boundError) {
	v, ok := m[k]
	if !ok {
		return nil, false, nil
	}

	m, err := ctx.validateMap(v, desc)
	return m, true, err
}

func (ctx context) validateList(v interface{}, desc string) ([]interface{}, boundError) {
	r, ok := v.([]interface{})
	if !ok {
		return nil, newListError(v, desc)
	}

	return r, nil
}
