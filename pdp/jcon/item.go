package jcon

import (
	"encoding/json"
	"fmt"
	"net"
	"strings"

	"github.com/infobloxopen/go-trees/domaintree"
	"github.com/infobloxopen/go-trees/iptree"
	"github.com/infobloxopen/go-trees/strtree"
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

	s, err := GetString(d, "content item type")
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

	err := CheckArrayStart(d, "content item keys")
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
			if s.String() != delimArrayEnd {
				return newArrayEndDelimiterError(s, delimArrayEnd, src)
			}

			c.k = k
			c.keysOk = true

			return nil
		}
	}
}

func (c *contentItem) unmarshalMap(d *json.Decoder, keyIdx int) (interface{}, error) {
	src := fmt.Sprintf("level %d map", keyIdx+1)
	err := CheckObjectStart(d, src)
	if err != nil {
		return nil, err
	}

	m, err := newTypedMap(c, keyIdx)
	if err != nil {
		return nil, err
	}

	err = UnmarshalObject(d, m.unmarshal, src)
	if err != nil {
		return nil, err
	}

	return m.get(), nil
}

func (c *contentItem) unmarshalValue(d *json.Decoder) (interface{}, error) {
	switch c.t {
	case pdp.TypeBoolean:
		return getBoolean(d, "value")

	case pdp.TypeString:
		return GetString(d, "value")

	case pdp.TypeAddress:
		s, err := GetString(d, "address value")
		if err != nil {
			return nil, err
		}

		a := net.ParseIP(s)
		if a == nil {
			return nil, newAddressCastError(s)
		}

		return a, nil

	case pdp.TypeNetwork:
		s, err := GetString(d, "network value")
		if err != nil {
			return nil, err
		}

		_, n, err := net.ParseCIDR(s)
		if err != nil {
			return nil, newNetworkCastError(s, err)
		}

		return n, nil

	case pdp.TypeDomain:
		s, err := GetString(d, "domain value")
		if err != nil {
			return nil, err
		}

		d, err := pdp.AdjustDomainName(s)
		if err != nil {
			return nil, newDomainCastError(s, err)
		}

		return d, nil

	case pdp.TypeSetOfStrings:
		m := strtree.NewTree()
		i := 0
		err := GetStringSequence(d, "set of strings value", func(s string) error {
			if _, ok := m.Get(s); !ok {
				m.InplaceInsert(s, i)
				i++
			}

			return nil
		})
		if err != nil {
			return nil, err
		}

		return m, nil

	case pdp.TypeSetOfNetworks:
		m := iptree.NewTree()
		err := GetStringSequence(d, "set of networks value", func(s string) error {
			a := net.ParseIP(s)
			if a != nil {
				m.InplaceInsertIP(a, nil)
			} else {
				_, n, err := net.ParseCIDR(s)
				if err != nil {
					return newAddressNetworkCastError(s, err)
				}

				m.InplaceInsertNet(n, nil)
			}

			return nil
		})
		if err != nil {
			return nil, err
		}

		return m, nil

	case pdp.TypeSetOfDomains:
		m := &domaintree.Node{}
		err := GetStringSequence(d, "set of domains value", func(s string) error {
			d, err := pdp.AdjustDomainName(s)
			if err != nil {
				return newDomainCastError(s, err)
			}

			m.InplaceInsert(d, nil)

			return nil
		})
		if err != nil {
			return nil, err
		}

		return m, nil

	case pdp.TypeListOfStrings:
		lst := []string{}
		err := GetStringSequence(d, "list of strings value", func(s string) error {
			lst = append(lst, s)
			return nil
		})
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
			return nil
		}

		if len(c.k) <= 0 {
			c.v = pdp.MakeContentValue(v)
		} else {
			c.v = v
		}
	} else {
		v, err := GetUndefined(d, "content")
		if err != nil {
			return nil
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
	err := CheckObjectStart(d, "content item")
	if err != nil {
		return nil, err
	}

	item := &contentItem{id: id}
	err = UnmarshalObject(d, item.unmarshal, "content item")
	if err != nil {
		return nil, err
	}

	return item.get()
}
