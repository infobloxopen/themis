package jast

import (
	"encoding/json"
	"net"
	"strings"

	"github.com/infobloxopen/themis/jparser"
	"github.com/infobloxopen/themis/pdp"

	"github.com/infobloxopen/go-trees/domaintree"
	"github.com/infobloxopen/go-trees/iptree"
	"github.com/infobloxopen/go-trees/strtree"
)

func (ctx context) unmarshalStringValue(d *json.Decoder) (pdp.AttributeValue, error) {
	s, err := jparser.GetString(d, "value of string type")
	if err != nil {
		return pdp.AttributeValue{}, err
	}

	return pdp.MakeStringValue(s), nil
}

func (ctx context) unmarshalIntegerValue(d *json.Decoder) (pdp.AttributeValue, error) {
	x, err := jparser.GetNumber(d, "value of integer type")
	if err != nil {
		return pdp.AttributeValue{}, err
	}

	if x < -9007199254740992 || x > 9007199254740992 {
		return pdp.AttributeValue{}, newIntegerOverflowError(x)
	}

	return pdp.MakeIntegerValue(int64(x)), nil
}

func (ctx context) unmarshalFloatValue(d *json.Decoder) (pdp.AttributeValue, error) {
	x, err := jparser.GetNumber(d, "value of float type")
	if err != nil {
		return pdp.AttributeValue{}, err
	}

	return pdp.MakeFloatValue(float64(x)), nil
}

func (ctx context) unmarshalAddressValue(d *json.Decoder) (pdp.AttributeValue, error) {
	s, err := jparser.GetString(d, "value of address type")
	if err != nil {
		return pdp.AttributeValue{}, err
	}

	a := net.ParseIP(s)
	if a == nil {
		return pdp.AttributeValue{}, newInvalidAddressError(s)
	}

	return pdp.MakeAddressValue(a), nil
}

func (ctx context) unmarshalNetworkValue(d *json.Decoder) (pdp.AttributeValue, error) {
	s, err := jparser.GetString(d, "value of network type")
	if err != nil {
		return pdp.AttributeValue{}, err
	}

	_, n, ierr := net.ParseCIDR(s)
	if ierr != nil {
		return pdp.AttributeValue{}, newInvalidNetworkError(s, ierr)
	}

	return pdp.MakeNetworkValue(n), nil
}

func (ctx context) unmarshalDomainValue(d *json.Decoder) (pdp.AttributeValue, error) {
	s, err := jparser.GetString(d, "value of domain type")
	if err != nil {
		return pdp.AttributeValue{}, err
	}

	dom, ierr := domaintree.MakeWireDomainNameLower(s)
	if ierr != nil {
		return pdp.AttributeValue{}, newInvalidDomainError(s, ierr)
	}

	return pdp.MakeDomainValue(dom), nil
}

func (ctx context) unmarshalSetOfStringsValue(d *json.Decoder) (pdp.AttributeValue, error) {
	set := strtree.NewTree()
	if err := jparser.GetStringSequence(d, func(idx int, s string) error {
		set.InplaceInsert(s, idx)
		return nil
	}, "set of strings"); err != nil {
		return pdp.AttributeValue{}, err
	}

	return pdp.MakeSetOfStringsValue(set), nil
}

func (ctx context) unmarshalSetOfNetworksValueItem(s string, i int, set *iptree.Tree) error {
	_, n, ierr := net.ParseCIDR(s)
	if ierr != nil {
		return newInvalidNetworkError(s, ierr)
	}

	set.InplaceInsertNet(n, i)

	return nil
}

func (ctx context) unmarshalSetOfNetworksValue(d *json.Decoder) (pdp.AttributeValue, error) {
	set := iptree.NewTree()
	if err := jparser.GetStringSequence(d, func(idx int, s string) error {
		if err := ctx.unmarshalSetOfNetworksValueItem(s, idx, set); err != nil {
			bindError(bindErrorf(err, "%d", idx), "set of networks")
		}

		return nil
	}, "set of networks"); err != nil {
		return pdp.AttributeValue{}, err
	}

	return pdp.MakeSetOfNetworksValue(set), nil
}

func (ctx context) unmarshalSetOfDomainsValueItem(s string, i int, set *domaintree.Node) error {
	set.InplaceInsert(s, i)

	return nil
}

func (ctx context) unmarshalSetOfDomainsValue(d *json.Decoder) (pdp.AttributeValue, error) {
	set := &domaintree.Node{}
	if err := jparser.GetStringSequence(d, func(idx int, s string) error {
		if err := ctx.unmarshalSetOfDomainsValueItem(s, idx, set); err != nil {
			bindError(bindErrorf(err, "%d", idx), "set of domains")
		}

		return nil
	}, "set of domains"); err != nil {
		return pdp.AttributeValue{}, err
	}

	return pdp.MakeSetOfDomainsValue(set), nil
}

func (ctx context) unmarshalListOfStringsValue(d *json.Decoder) (pdp.AttributeValue, error) {
	list := []string{}
	if err := jparser.GetStringSequence(d, func(idx int, s string) error {
		list = append(list, s)

		return nil
	}, "list of strings"); err != nil {
		return pdp.AttributeValue{}, err
	}

	return pdp.MakeListOfStringsValue(list), nil
}

func (ctx context) unmarshalValueByType(t int, d *json.Decoder) (pdp.AttributeValue, error) {
	switch t {
	case pdp.TypeString:
		return ctx.unmarshalStringValue(d)

	case pdp.TypeInteger:
		return ctx.unmarshalIntegerValue(d)

	case pdp.TypeFloat:
		return ctx.unmarshalFloatValue(d)

	case pdp.TypeAddress:
		return ctx.unmarshalAddressValue(d)

	case pdp.TypeNetwork:
		return ctx.unmarshalNetworkValue(d)

	case pdp.TypeDomain:
		return ctx.unmarshalDomainValue(d)

	case pdp.TypeSetOfStrings:
		return ctx.unmarshalSetOfStringsValue(d)

	case pdp.TypeSetOfNetworks:
		return ctx.unmarshalSetOfNetworksValue(d)

	case pdp.TypeSetOfDomains:
		return ctx.unmarshalSetOfDomainsValue(d)

	case pdp.TypeListOfStrings:
		return ctx.unmarshalListOfStringsValue(d)
	}

	return pdp.AttributeValue{}, newNotImplementedValueTypeError(t)
}

func (ctx context) unmarshalValue(d *json.Decoder) (pdp.AttributeValue, error) {
	if err := jparser.CheckObjectStart(d, "value"); err != nil {
		return pdp.AttributeValue{}, err
	}

	var (
		cOk bool
		a   pdp.AttributeValue
		t   = -1
	)

	if err := jparser.UnmarshalObject(d, func(k string, d *json.Decoder) error {
		k = strings.ToLower(k)

		switch k {
		case yastTagType:
			s, err := jparser.GetString(d, "value type")
			if err != nil {
				return err
			}

			var ok bool
			t, ok = pdp.TypeIDs[strings.ToLower(s)]
			if !ok {
				return newUnknownTypeError(s)
			}

			if t == pdp.TypeUndefined {
				return newInvalidTypeError(t)
			}

			return nil

		case yastTagContent:
			if t == -1 {
				return newMissingContentTypeError()
			}

			cOk = true
			var err error
			a, err = ctx.unmarshalValueByType(t, d)
			return err
		}

		return newUnknownAttributeError(k)
	}, "value"); err != nil {
		return pdp.AttributeValue{}, err
	}

	if !cOk {
		return pdp.AttributeValue{}, newMissingContentError()
	}

	return a, nil
}
