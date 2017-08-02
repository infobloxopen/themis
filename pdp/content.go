package pdp

import (
	"encoding/json"
	"fmt"
	"net"
	"strings"

	"github.com/infobloxopen/go-trees/domaintree"
	"github.com/infobloxopen/go-trees/iptree"
	"github.com/infobloxopen/go-trees/strtree"
)

var ContentKeyTypes = map[int]bool{
	TypeString:  true,
	TypeAddress: true,
	TypeNetwork: true,
	TypeDomain:  true}

type ContentItem struct {
	r ContentSubItem
	t int
	k []int
}

func MakeContentValueItem(t int, v interface{}) ContentItem {
	return ContentItem{
		r: MakeContentValue(v),
		t: t}
}

func MakeContentMappingItem(t int, k []int, v ContentSubItem) ContentItem {
	return ContentItem{
		r: v,
		t: t,
		k: k}
}

func (c ContentItem) MarshalJSON() ([]byte, error) {
	var err error
	b := []byte("{")

	if len(c.k) > 0 {
		b, err = appendJSONTag(b, "keys")
		if err != nil {
			return nil, err
		}

		keys := make([]string, len(c.k))
		for i, k := range c.k {
			keys[i] = TypeNames[k]
		}

		b, err = appendJSONStringArray(b, keys)
		if err != nil {
			return nil, err
		}

		b = append(b, ',')
	}

	b, err = appendJSONTag(b, "type")
	if err != nil {
		return nil, err
	}

	b, err = appendJSONString(b, TypeNames[c.t])
	if err != nil {
		return nil, err
	}

	b = append(b, ',')

	b, err = appendJSONTag(b, "data")
	if err != nil {
		return nil, err
	}

	b, err = c.r.appendJSON(b, len(c.k), c.t)
	if err != nil {
		return nil, err
	}

	return append(b, '}'), nil
}

func (c ContentItem) get(path []Expression, ctx *Context) (AttributeValue, error) {
	d := len(path)
	if d != len(c.k) {
		return undefinedValue, newInvalidSelectorPathError(c.k, path)
	}

	if d > 0 {
		m := c.r
		loc := []string{""}
		for _, e := range path[:d-1] {
			key, err := e.calculate(ctx)
			if err != nil {
				return undefinedValue, bindError(err, strings.Join(loc, "/"))
			}

			loc = append(loc, key.describe())

			m, err = m.next(key)
			if err != nil {
				return undefinedValue, bindError(err, strings.Join(loc, "/"))
			}
		}

		key, err := path[d-1].calculate(ctx)
		if err != nil {
			return undefinedValue, bindError(err, strings.Join(loc, "/"))
		}

		v, err := m.getValue(key, c.t)
		if err != nil {
			return undefinedValue, bindError(err, strings.Join(append(loc, key.describe()), "/"))
		}

		return v, nil
	}

	return c.r.getValue(undefinedValue, c.t)
}

type ContentSubItem interface {
	getValue(key AttributeValue, t int) (AttributeValue, error)
	next(key AttributeValue) (ContentSubItem, error)
	appendJSON(b []byte, level int, t int) ([]byte, error)
}

type contentStringMap struct {
	tree *strtree.Tree
}

func MakeContentStringMap(tree *strtree.Tree) contentStringMap {
	return contentStringMap{tree: tree}
}

func (m contentStringMap) appendJSON(b []byte, level int, t int) ([]byte, error) {
	var err error
	b = append(b, '{')
	if level > 1 {
		i := 0
		for p := range m.tree.Enumerate() {
			if i > 0 {
				b = append(b, ',')
			}

			b, err = appendJSONTag(b, p.Key)
			if err != nil {
				return nil, err
			}

			n, ok := p.Value.(ContentSubItem)
			if !ok {
				panic(fmt.Errorf("Local selector: Invalid content item map at %q. Expected ContentSubItem but got %T",
					p.Key, p.Value))
			}

			b, err = n.appendJSON(b, level-1, t)
			if err != nil {
				return nil, err
			}

			i++
		}

		return append(b, '}'), nil
	}

	i := 0
	for p := range m.tree.Enumerate() {
		if i > 0 {
			b = append(b, ',')
		}

		b, err = appendJSONTag(b, p.Key)
		if err != nil {
			return nil, err
		}

		b, err = appendJSONValue(b, p.Value, t)
		if err != nil {
			return nil, err
		}

		i++
	}

	return append(b, '}'), nil
}

func (m contentStringMap) getValue(key AttributeValue, t int) (AttributeValue, error) {
	s, err := key.str()
	if err != nil {
		return undefinedValue, err
	}

	v, ok := m.tree.Get(s)
	if !ok {
		return undefinedValue, newMissingValueError()
	}

	return MakeContentValue(v).getValue(undefinedValue, t)
}

func (m contentStringMap) next(key AttributeValue) (ContentSubItem, error) {
	s, err := key.str()
	if err != nil {
		return nil, err
	}

	v, ok := m.tree.Get(s)
	if !ok {
		return nil, newMissingValueError()
	}

	item, ok := v.(ContentSubItem)
	if !ok {
		return nil, newMapContentSubitemError()
	}

	return item, nil
}

type contentNetworkMap struct {
	tree *iptree.Tree
}

func MakeContentNetworkMap(tree *iptree.Tree) contentNetworkMap {
	return contentNetworkMap{tree: tree}
}

func (m contentNetworkMap) appendJSON(b []byte, level int, t int) ([]byte, error) {
	var err error
	b = append(b, '{')
	if level > 1 {
		i := 0
		for p := range m.tree.Enumerate() {
			if i > 0 {
				b = append(b, ',')
			}

			b, err = appendJSONTag(b, p.Key.String())
			if err != nil {
				return nil, err
			}

			n, ok := p.Value.(ContentSubItem)
			if !ok {
				panic(fmt.Errorf("Local selector: Invalid content item map at %q. Expected ContentSubItem but got %T",
					p.Key, p.Value))
			}

			b, err = n.appendJSON(b, level-1, t)
			if err != nil {
				return nil, err
			}

			i++
		}

		return append(b, '}'), nil
	}

	i := 0
	for p := range m.tree.Enumerate() {
		if i > 0 {
			b = append(b, ',')
		}

		b, err = appendJSONTag(b, p.Key.String())
		if err != nil {
			return nil, err
		}

		b, err = appendJSONValue(b, p.Value, t)
		if err != nil {
			return nil, err
		}

		i++
	}

	return append(b, '}'), nil
}

func (m contentNetworkMap) getValue(key AttributeValue, t int) (AttributeValue, error) {
	a, err := key.address()
	if err != nil {
		return undefinedValue, err
	}

	v, ok := m.tree.GetByIP(a)
	if !ok {
		return undefinedValue, newMissingValueError()
	}

	return MakeContentValue(v).getValue(undefinedValue, t)
}

func (m contentNetworkMap) next(key AttributeValue) (ContentSubItem, error) {
	a, err := key.address()
	if err != nil {
		return nil, err
	}

	v, ok := m.tree.GetByIP(a)
	if !ok {
		return nil, newMissingValueError()
	}

	item, ok := v.(ContentSubItem)
	if !ok {
		return nil, newMapContentSubitemError()
	}

	return item, nil
}

type contentDomainMap struct {
	tree *domaintree.Node
}

func MakeContentDomainMap(tree *domaintree.Node) contentDomainMap {
	return contentDomainMap{tree: tree}
}

func (m contentDomainMap) appendJSON(b []byte, level int, t int) ([]byte, error) {
	var err error
	b = append(b, '{')
	if level > 1 {
		i := 0
		for p := range m.tree.Enumerate() {
			if i > 0 {
				b = append(b, ',')
			}

			b, err = appendJSONTag(b, p.Key)
			if err != nil {
				return nil, err
			}

			n, ok := p.Value.(ContentSubItem)
			if !ok {
				panic(fmt.Errorf("Local selector: Invalid content item map at %q. Expected ContentSubItem but got %T",
					p.Key, p.Value))
			}

			b, err = n.appendJSON(b, level-1, t)
			if err != nil {
				return nil, err
			}

			i++
		}

		return append(b, '}'), nil
	}

	i := 0
	for p := range m.tree.Enumerate() {
		if i > 0 {
			b = append(b, ',')
		}

		b, err = appendJSONTag(b, p.Key)
		if err != nil {
			return nil, err
		}

		b, err = appendJSONValue(b, p.Value, t)
		if err != nil {
			return nil, err
		}

		i++
	}

	return append(b, '}'), nil
}

func (m contentDomainMap) getValue(key AttributeValue, t int) (AttributeValue, error) {
	d, err := key.domain()
	if err != nil {
		return undefinedValue, err
	}

	v, ok := m.tree.Get(d)
	if !ok {
		return undefinedValue, newMissingValueError()
	}

	return MakeContentValue(v).getValue(undefinedValue, t)
}

func (m contentDomainMap) next(key AttributeValue) (ContentSubItem, error) {
	d, err := key.domain()
	if err != nil {
		return nil, err
	}

	v, ok := m.tree.Get(d)
	if !ok {
		return nil, newMissingValueError()
	}

	item, ok := v.(ContentSubItem)
	if !ok {
		return nil, newMapContentSubitemError()
	}

	return item, nil
}

type contentValue struct {
	value interface{}
}

func MakeContentValue(value interface{}) contentValue {
	return contentValue{value: value}
}

func (v contentValue) appendJSON(b []byte, level int, t int) ([]byte, error) {
	return appendJSONValue(b, v.value, t)
}

func (v contentValue) getValue(key AttributeValue, t int) (AttributeValue, error) {
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

func (v contentValue) next(key AttributeValue) (ContentSubItem, error) {
	return nil, newMapContentSubitemError()
}

func appendJSONTag(b []byte, tag string) ([]byte, error) {
	t, err := json.Marshal(tag)
	if err != nil {
		return nil, err
	}

	return append(append(b, t...), ':'), nil
}

func appendJSONString(b []byte, s string) ([]byte, error) {
	v, err := json.Marshal(s)
	if err != nil {
		return nil, err
	}

	return append(b, v...), nil
}

func appendJSONStringArray(b []byte, s []string) ([]byte, error) {
	a, err := json.Marshal(s)
	if err != nil {
		return nil, err
	}

	return append(b, a...), nil
}

func appendJSONValue(b []byte, v interface{}, t int) ([]byte, error) {
	var (
		value []byte
		err   error
	)

	switch t {
	default:
		panic(fmt.Errorf("Can't marshal to JSON of unknown type with index %d", t))

	case TypeUndefined:
		panic(fmt.Errorf("Can't marshal value of %s type to JSON", TypeNames[t]))

	case TypeBoolean:
		bVal, ok := v.(bool)
		if !ok {
			panic(fmt.Errorf("Can't marshal %T to JSON as %s", v, TypeNames[t]))
		}

		value, err = json.Marshal(bVal)

	case TypeString:
		s, ok := v.(string)
		if !ok {
			panic(fmt.Errorf("Can't marshal %T to JSON as %s", v, TypeNames[t]))
		}

		value, err = json.Marshal(s)

	case TypeAddress:
		a, ok := v.(net.IP)
		if !ok {
			panic(fmt.Errorf("Can't marshal %T to JSON as %s", v, TypeNames[t]))
		}

		value, err = json.Marshal(a.String())

	case TypeNetwork:
		n, ok := v.(*net.IPNet)
		if !ok {
			panic(fmt.Errorf("Can't marshal %T to JSON as %s", v, TypeNames[t]))
		}

		value, err = json.Marshal(n.String())

	case TypeDomain:
		d, ok := v.(string)
		if !ok {
			panic(fmt.Errorf("Can't marshal %T to JSON as %s", v, TypeNames[t]))
		}

		value, err = json.Marshal(d)

	case TypeSetOfStrings:
		s, ok := v.(*strtree.Tree)
		if !ok {
			panic(fmt.Errorf("Can't marshal %T to JSON as %s", v, TypeNames[t]))
		}

		value, err = json.Marshal(sortSetOfStrings(s))

	case TypeSetOfNetworks:
		s, ok := v.(*iptree.Tree)
		if !ok {
			panic(fmt.Errorf("Can't marshal %T to JSON as %s", v, TypeNames[t]))
		}

		strs := []string{}
		for p := range s.Enumerate() {
			strs = append(strs, p.Key.String())
		}

		value, err = json.Marshal(strs)

	case TypeSetOfDomains:
		s, ok := v.(*domaintree.Node)
		if !ok {
			panic(fmt.Errorf("Can't marshal %T to JSON as %s", v, TypeNames[t]))
		}

		strs := []string{}
		for p := range s.Enumerate() {
			strs = append(strs, p.Key)
		}

		value, err = json.Marshal(strs)

	case TypeListOfStrings:
		lst, ok := v.([]string)
		if !ok {
			panic(fmt.Errorf("Can't marshal %T to JSON as %s", v, TypeNames[t]))
		}

		value, err = json.Marshal(lst)
	}

	if err != nil {
		return nil, err
	}

	return append(b, value...), nil
}
