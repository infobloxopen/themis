package jcon

import (
	"encoding/json"
	"fmt"
	"net"
	"strings"

	"github.com/infobloxopen/go-trees/domaintree"
	"github.com/infobloxopen/go-trees/iptree"
	"github.com/infobloxopen/go-trees/strtree"
	"github.com/infobloxopen/themis/jparser"
	"github.com/infobloxopen/themis/pdp"
)

type contentItem struct {
	id string

	k      []int
	keysOk bool

	t   int
	tOk bool

	v      interface{}
	vOk    bool
	vReady bool
}

func (c *contentItem) unmarshalTypeField(d *json.Decoder) error {
	if c.tOk {
		return newDuplicateContentItemFieldError("type")
	}

	s, err := jparser.GetString(d, "content item type")
	if err != nil {
		return err
	}

	t, ok := pdp.TypeIDs[strings.ToLower(s)]
	if !ok {
		return newUnknownTypeError(s)
	}

	if t == pdp.TypeUndefined {
		return newInvalidContentItemTypeError(t)
	}

	c.t = t
	c.tOk = true

	return nil
}

func (c *contentItem) unmarshalKeysField(d *json.Decoder) error {
	if c.keysOk {
		return newDuplicateContentItemFieldError("keys")
	}

	err := jparser.CheckArrayStart(d, "content item keys")
	if err != nil {
		return err
	}

	k := []int{}
	i := 1
	for {
		src := fmt.Sprintf("key %d", i)
		t, err := d.Token()
		if err != nil {
			return bindError(err, src)
		}

		switch s := t.(type) {
		default:
			return newStringCastError(t, src)

		case string:
			t, ok := pdp.TypeIDs[strings.ToLower(s)]
			if !ok {
				return bindError(newUnknownTypeError(s), src)
			}

			if _, ok := pdp.ContentKeyTypes[t]; !ok {
				return bindError(newInvalidContentKeyTypeError(t, pdp.ContentKeyTypes), src)
			}

			k = append(k, t)
			i++

		case json.Delim:
			if s.String() != jparser.DelimArrayEnd {
				return newArrayEndDelimiterError(s, jparser.DelimArrayEnd, src)
			}

			c.k = k
			c.keysOk = true

			return nil
		}
	}
}

func (c *contentItem) unmarshalMap(d *json.Decoder, keyIdx int) (interface{}, error) {
	src := fmt.Sprintf("level %d map", keyIdx+1)
	err := jparser.CheckObjectStart(d, src)
	if err != nil {
		return nil, err
	}

	m, err := newTypedMap(c, keyIdx)
	if err != nil {
		return nil, err
	}

	err = jparser.UnmarshalObject(d, m.unmarshal, src)
	if err != nil {
		return nil, err
	}

	return m.get(), nil
}

func (c *contentItem) unmarshalValue(d *json.Decoder) (interface{}, error) {
	switch c.t {
	case pdp.TypeBoolean:
		return jparser.GetBoolean(d, "value")

	case pdp.TypeString:
		return jparser.GetString(d, "value")

	case pdp.TypeInteger:
		x, err := jparser.GetNumber(d, "value")
		if err != nil {
			return nil, err
		}

		if x < -9007199254740992 || x > 9007199254740992 {
			return nil, newIntegerOverflowError(x)
		}

		return int64(x), nil

	case pdp.TypeFloat:
		x, err := jparser.GetNumber(d, "value")
		if err != nil {
			return nil, err
		}

		return float64(x), nil

	case pdp.TypeAddress:
		s, err := jparser.GetString(d, "address value")
		if err != nil {
			return nil, err
		}

		a := net.ParseIP(s)
		if a == nil {
			return nil, newAddressCastError(s)
		}

		return a, nil

	case pdp.TypeNetwork:
		s, err := jparser.GetString(d, "network value")
		if err != nil {
			return nil, err
		}

		_, n, err := net.ParseCIDR(s)
		if err != nil {
			return nil, newNetworkCastError(s, err)
		}

		return n, nil

	case pdp.TypeDomain:
		s, err := jparser.GetString(d, "domain value")
		if err != nil {
			return nil, err
		}

		d, err := domaintree.MakeWireDomainNameLower(s)
		if err != nil {
			return nil, newDomainCastError(s, err)
		}

		return d, nil

	case pdp.TypeSetOfStrings:
		m := strtree.NewTree()
		i := 0
		err := jparser.GetStringSequence(d, func(idx int, s string) error {
			if _, ok := m.Get(s); !ok {
				m.InplaceInsert(s, i)
				i++
			}

			return nil
		}, "set of strings value")
		if err != nil {
			return nil, err
		}

		return m, nil

	case pdp.TypeSetOfNetworks:
		m := iptree.NewTree()
		err := jparser.GetStringSequence(d, func(idx int, s string) error {
			a := net.ParseIP(s)
			if a != nil {
				m.InplaceInsertIP(a, nil)
			} else {
				_, n, err := net.ParseCIDR(s)
				if err != nil {
					return bindErrorf(newAddressNetworkCastError(s, err), "%d", idx)
				}

				m.InplaceInsertNet(n, nil)
			}

			return nil
		}, "set of networks value")
		if err != nil {
			return nil, err
		}

		return m, nil

	case pdp.TypeSetOfDomains:
		m := &domaintree.Node{}
		err := jparser.GetStringSequence(d, func(idx int, s string) error {
			m.InplaceInsert(s, nil)
			return nil
		}, "set of domains value")
		if err != nil {
			return nil, err
		}

		return m, nil

	case pdp.TypeListOfStrings:
		lst := []string{}
		err := jparser.GetStringSequence(d, func(idx int, s string) error {
			lst = append(lst, s)
			return nil
		}, "list of strings value")
		if err != nil {
			return nil, err
		}

		return lst, nil
	}

	return nil, newInvalidContentItemTypeError(c.t)
}

func (c *contentItem) unmarshalTypedData(d *json.Decoder, keyIdx int) (interface{}, error) {
	if len(c.k) > keyIdx {
		return c.unmarshalMap(d, keyIdx)
	}

	return c.unmarshalValue(d)
}

func (c *contentItem) unmarshalDataField(d *json.Decoder) error {
	if c.vOk {
		return newDuplicateContentItemFieldError("type")
	}

	c.vReady = c.keysOk && c.tOk
	if c.vReady {
		v, err := c.unmarshalTypedData(d, 0)
		if err != nil {
			return err
		}

		if len(c.k) <= 0 {
			c.v = pdp.MakeContentValue(v)
		} else {
			c.v = v
		}
	} else {
		v, err := jparser.GetUndefined(d, "content")
		if err != nil {
			return err
		}

		c.v = v
	}

	c.vOk = true
	return nil
}

func (c *contentItem) unmarshal(k string, d *json.Decoder) error {
	switch strings.ToLower(k) {
	case "type":
		return c.unmarshalTypeField(d)

	case "keys":
		return c.unmarshalKeysField(d)

	case "data":
		return c.unmarshalDataField(d)
	}

	return newUnknownContentItemFieldError(k)
}

func (c *contentItem) adjustValue(v interface{}) pdp.ContentSubItem {
	cv, ok := v.(pdp.ContentSubItem)
	if !ok {
		panic(fmt.Errorf("expected value of type ContentSubItem when item is ready but got %T", v))
	}

	return cv
}

func (c *contentItem) get() (*pdp.ContentItem, error) {
	if !c.vOk {
		return nil, newMissingContentDataError()
	}

	if !c.tOk {
		return nil, newMissingContentTypeError()
	}

	if c.vReady {
		return pdp.MakeContentMappingItem(c.id, c.t, c.k, c.adjustValue(c.v)), nil
	}

	v, err := c.postProcess(c.v, 0)
	if err != nil {
		return nil, err
	}

	if len(c.k) <= 0 {
		return pdp.MakeContentValueItem(c.id, c.t, v), nil
	}

	return pdp.MakeContentMappingItem(c.id, c.t, c.k, c.adjustValue(v)), nil
}

func unmarshalContentItem(id string, d *json.Decoder) (*pdp.ContentItem, error) {
	err := jparser.CheckObjectStart(d, "content item")
	if err != nil {
		return nil, err
	}

	item := &contentItem{id: id}
	err = jparser.UnmarshalObject(d, item.unmarshal, "content item")
	if err != nil {
		return nil, err
	}

	return item.get()
}
