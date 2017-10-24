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

func (ctx context) decodeStringValue(d *json.Decoder) (pdp.AttributeValue, error) {
	s, err := jparser.GetString(d, "value of string type")
	if err != nil {
		return pdp.AttributeValue{}, err
	}

	return pdp.MakeStringValue(s), nil
}

func (ctx context) decodeAddressValue(d *json.Decoder) (pdp.AttributeValue, error) {
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

func (ctx context) decodeNetworkValue(d *json.Decoder) (pdp.AttributeValue, error) {
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

func (ctx context) decodeDomainValue(d *json.Decoder) (pdp.AttributeValue, error) {
	s, err := jparser.GetString(d, "value of domain type")
	if err != nil {
		return pdp.AttributeValue{}, err
	}

	dom, ierr := pdp.AdjustDomainName(s)
	if ierr != nil {
		return pdp.AttributeValue{}, newInvalidDomainError(s, ierr)
	}

	return pdp.MakeDomainValue(dom), nil
}

func (ctx context) decodeSetOfStringsValue(d *json.Decoder) (pdp.AttributeValue, error) {
	set := strtree.NewTree()
	if err := jparser.GetStringSequence(d, func(idx int, s string) error {
		set.InplaceInsert(s, idx)
		return nil
	}, "set of strings"); err != nil {
		return pdp.AttributeValue{}, err
	}

	return pdp.MakeSetOfStringsValue(set), nil
}

func (ctx context) decodeSetOfNetworksValueItem(s string, i int, set *iptree.Tree) error {
	_, n, ierr := net.ParseCIDR(s)
	if ierr != nil {
		return newInvalidNetworkError(s, ierr)
	}

	set.InplaceInsertNet(n, i)

	return nil
}

func (ctx context) decodeSetOfNetworksValue(d *json.Decoder) (pdp.AttributeValue, error) {
	set := iptree.NewTree()
	if err := jparser.GetStringSequence(d, func(idx int, s string) error {
		if err := ctx.decodeSetOfNetworksValueItem(s, idx, set); err != nil {
			bindError(bindErrorf(err, "%d", idx), "set of networks")
		}

		return nil
	}, "set of networks"); err != nil {
		return pdp.AttributeValue{}, err
	}

	return pdp.MakeSetOfNetworksValue(set), nil
}

func (ctx context) decodeSetOfDomainsValueItem(s string, i int, set *domaintree.Node) error {
	dom, err := pdp.AdjustDomainName(s)
	if err != nil {
		return newInvalidDomainError(s, err)
	}

	set.InplaceInsert(dom, i)

	return nil
}

func (ctx context) decodeSetOfDomainsValue(d *json.Decoder) (pdp.AttributeValue, error) {
	set := &domaintree.Node{}
	if err := jparser.GetStringSequence(d, func(idx int, s string) error {
		if err := ctx.decodeSetOfDomainsValueItem(s, idx, set); err != nil {
			bindError(bindErrorf(err, "%d", idx), "set of domains")
		}

		return nil
	}, "set of domains"); err != nil {
		return pdp.AttributeValue{}, err
	}

	return pdp.MakeSetOfDomainsValue(set), nil
}

func (ctx context) decodeListOfStringsValue(d *json.Decoder) (pdp.AttributeValue, error) {
	list := []string{}
	if err := jparser.GetStringSequence(d, func(idx int, s string) error {
		list = append(list, s)

		return nil
	}, "list of strings"); err != nil {
		return pdp.AttributeValue{}, err
	}

	return pdp.MakeListOfStringsValue(list), nil
}

func (ctx context) decodeValueByType(t int, d *json.Decoder) (pdp.AttributeValue, error) {
	switch t {
	case pdp.TypeString:
		return ctx.decodeStringValue(d)

	case pdp.TypeAddress:
		return ctx.decodeAddressValue(d)

	case pdp.TypeNetwork:
		return ctx.decodeNetworkValue(d)

	case pdp.TypeDomain:
		return ctx.decodeDomainValue(d)

	case pdp.TypeSetOfStrings:
		return ctx.decodeSetOfStringsValue(d)

	case pdp.TypeSetOfNetworks:
		return ctx.decodeSetOfNetworksValue(d)

	case pdp.TypeSetOfDomains:
		return ctx.decodeSetOfDomainsValue(d)

	case pdp.TypeListOfStrings:
		return ctx.decodeListOfStringsValue(d)
	}

	return pdp.AttributeValue{}, newNotImplementedValueTypeError(t)
}

func (ctx context) decodeValue(d *json.Decoder) (pdp.AttributeValue, error) {
	if err := jparser.CheckObjectStart(d, "value"); err != nil {
		return pdp.AttributeValue{}, err
	}

	var (
		cOk bool
		a   pdp.AttributeValue
		t   int = -1
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
			a, err = ctx.decodeValueByType(t, d)
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
