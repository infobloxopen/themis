package pdp

import (
	"encoding/json"
	"fmt"
	"net"
	"strings"

	"github.com/satori/go.uuid"

	"github.com/infobloxopen/go-trees/domaintree"
	"github.com/infobloxopen/go-trees/iptree"
	"github.com/infobloxopen/go-trees/strtree"
)

var ContentKeyTypes = map[int]bool{
	TypeString:  true,
	TypeAddress: true,
	TypeNetwork: true,
	TypeDomain:  true}

type localContentStorage struct {
	r *strtree.Tree
}

func (c *localContentStorage) get(cID, iID string) (*ContentItem, error) {
	v, ok := c.r.Get(cID)
	if !ok {
		return nil, newMissingContentError(cID)
	}

	cnt, ok := v.(*localContent)
	if !ok {
		return nil, newInvalidContentStorageItem(cID, v)
	}

	item, err := cnt.get(iID)
	if err != nil {
		return nil, bindError(err, cID)
	}

	return item, nil
}

func (c *localContentStorage) newTransaction(cID string, tag *uuid.UUID) (*localContentStorageTransaction, error) {
	v, ok := c.r.Get(cID)
	if !ok {
		return nil, newMissingContentError(cID)
	}

	cnt, ok := v.(*localContent)
	if !ok {
		return nil, newInvalidContentStorageItem(cID, v)
	}

	if cnt.tag == nil {
		return nil, newUntaggedContentModificationError(cID)
	}

	if tag == nil {
		return nil, newMissingContentTagError()
	}

	if !uuid.Equal(*cnt.tag, *tag) {
		return nil, newContentTagsNotMatchError(cID, cnt.tag, tag)
	}

	return &localContentStorageTransaction{ID: cID, items: cnt.items}, nil
}

type ContentUpdate struct {
	cID    string
	oldTag uuid.UUID
	newTag uuid.UUID
	cmds   []*command
}

func NewContentUpdate(cID string, oldTag, newTag uuid.UUID) *ContentUpdate {
	return &ContentUpdate{
		cID:    cID,
		oldTag: oldTag,
		newTag: newTag,
		cmds:   []*command{}}
}

func (u *ContentUpdate) Append(op int, path []string, entity *ContentItem) {
	u.cmds = append(u.cmds, &command{op: op, path: path, entity: entity})
}

func (u *ContentUpdate) MarshalJSON() ([]byte, error) {
	b := []byte("[")

	for i, cmd := range u.cmds {
		if i > 0 {
			b = append(b, ',')
		}

		value, err := json.Marshal(cmd)
		if err != nil {
			return nil, err
		}

		b = append(b, value...)
	}

	return append(b, ']'), nil
}

type localContentStorageTransaction struct {
	ID    string
	items *strtree.Tree
}

func (t *localContentStorageTransaction) commit(c *localContentStorage, tag *uuid.UUID) (*localContentStorage, error) {
	if tag == nil {
		return nil, newMissingContentTagError()
	}

	return &localContentStorage{r: c.r.Insert(t.ID, &localContent{tag: tag, items: t.items})}, nil
}

func (t *localContentStorageTransaction) add(ID string, path []AttributeValue, v interface{}) error {
	var (
		c   *ContentItem
		ok  bool
		err error
	)

	if len(path) > 0 {
		v, ok := t.items.Get(ID)
		if !ok {
			return bindError(newMissingContentItemError(ID), t.ID)
		}

		c, ok = v.(*ContentItem)
		if !ok {
			return bindError(bindError(newInvalidContentItemError(v), ID), t.ID)
		}

		c, err = c.add(path, v)
		if err != nil {
			return bindError(bindError(err, ID), t.ID)
		}
	} else {
		c, ok = v.(*ContentItem)
		if !ok {
			return bindError(newInvalidContentItemError(v), ID)
		}
	}

	t.items = t.items.Insert(ID, c)
	return nil
}

func (t *localContentStorageTransaction) del(ID string, path []AttributeValue) error {
	if len(path) > 0 {
		v, ok := t.items.Get(ID)
		if !ok {
			return bindError(newMissingContentItemError(ID), t.ID)
		}

		c, ok := v.(*ContentItem)
		if !ok {
			return bindError(bindError(newInvalidContentItemError(v), ID), t.ID)
		}

		c, err := c.del(path)
		if err != nil {
			return bindError(bindError(err, ID), t.ID)
		}

		t.items = t.items.Insert(ID, c)
		return nil
	}

	items, ok := t.items.Delete(ID)
	if !ok {
		return bindError(newMissingContentItemError(ID), t.ID)
	}

	t.items = items
	return nil
}

type localContent struct {
	tag   *uuid.UUID
	items *strtree.Tree
}

func (c *localContent) get(ID string) (*ContentItem, error) {
	v, ok := c.items.Get(ID)
	if !ok {
		return nil, newMissingContentItemError(ID)
	}

	item, ok := v.(*ContentItem)
	if !ok {
		return nil, bindError(newInvalidContentItemError(v), ID)
	}

	return item, nil
}

type ContentItem struct {
	r ContentSubItem
	t int
	k []int
}

func MakeContentValueItem(t int, v interface{}) *ContentItem {
	return &ContentItem{
		r: MakeContentValue(v),
		t: t}
}

func MakeContentMappingItem(t int, k []int, v ContentSubItem) *ContentItem {
	return &ContentItem{
		r: v,
		t: t,
		k: k}
}

func (c *ContentItem) typeCheck(path []AttributeValue, v interface{}) (ContentSubItem, error) {
	if len(path) < len(c.k) {
		switch c.k[len(path)] {
		default:
			return nil, newInvalidContentKeyTypeError(c.k[len(path)], ContentKeyTypes)

		case TypeString:
			if _, ok := v.(contentStringMap); !ok {
				return nil, newInvalidContentStringMapError(v)
			}

		case TypeAddress, TypeNetwork:
			if _, ok := v.(contentNetworkMap); !ok {
				return nil, newInvalidContentNetworkMapError(v)
			}

		case TypeDomain:
			if _, ok := v.(contentDomainMap); !ok {
				return nil, newInvalidContentDomainMapError(v)
			}
		}

		return v.(ContentSubItem), nil
	}

	subItem, ok := v.(contentValue)
	if !ok {
		return nil, newInvalidContentValueError(v)
	}

	switch c.t {
	default:
		return nil, newUnknownContentItemResultTypeError(c.t)

	case TypeUndefined:
		return nil, newInvalidContentItemResultTypeError(c.t)

	case TypeBoolean:
		if _, ok := subItem.value.(bool); !ok {
			return nil, newInvalidContentValueTypeError(subItem.value, TypeBoolean)
		}

	case TypeString:
		if _, ok := subItem.value.(string); !ok {
			return nil, newInvalidContentValueTypeError(subItem.value, TypeString)
		}

	case TypeAddress:
		if _, ok := subItem.value.(net.IP); !ok {
			return nil, newInvalidContentValueTypeError(subItem.value, TypeAddress)
		}

	case TypeNetwork:
		if _, ok := subItem.value.(*net.IPNet); !ok {
			return nil, newInvalidContentValueTypeError(subItem.value, TypeNetwork)
		}

	case TypeDomain:
		if _, ok := subItem.value.(string); !ok {
			return nil, newInvalidContentValueTypeError(subItem.value, TypeDomain)
		}

	case TypeSetOfStrings:
		if _, ok := subItem.value.(*strtree.Tree); !ok {
			return nil, newInvalidContentValueTypeError(subItem.value, TypeSetOfStrings)
		}

	case TypeSetOfNetworks:
		if _, ok := subItem.value.(*iptree.Tree); !ok {
			return nil, newInvalidContentValueTypeError(subItem.value, TypeSetOfNetworks)
		}

	case TypeSetOfDomains:
		if _, ok := subItem.value.(*domaintree.Node); !ok {
			return nil, newInvalidContentValueTypeError(subItem.value, TypeSetOfDomains)
		}

	case TypeListOfStrings:
		if _, ok := subItem.value.([]string); !ok {
			return nil, newInvalidContentValueTypeError(subItem.value, TypeListOfStrings)
		}
	}

	return subItem, nil
}

func (c *ContentItem) add(path []AttributeValue, v interface{}) (*ContentItem, error) {
	if len(c.k) <= 0 {
		return c, newInvalidContentModificationError()
	}

	if len(path) <= 0 {
		return c, newMissingPathContentModificationError()
	}

	if len(path) > len(c.k) {
		return c, newTooLongPathContentModificationError(c.k, path)
	}

	var err error
	m := c.r

	last := len(path) - 1
	branch := make([]ContentSubItem, last)

	loc := []string{""}

	for i, k := range path[:last] {
		branch[i] = m
		loc = append(loc, k.describe())

		m, err = m.next(k)
		if err != nil {
			return c, bindError(err, strings.Join(loc, "/"))
		}
	}

	k := path[last]
	loc = append(loc, k.describe())

	subItem, err := c.typeCheck(path, v)
	if err != nil {
		return c, bindError(err, strings.Join(loc, "/"))
	}

	m, err = m.put(k, subItem)
	if err != nil {
		return c, bindError(err, strings.Join(loc, "/"))
	}

	for i := len(branch) - 1; i >= 0; i-- {
		p := branch[i]
		m, err = p.put(path[i], m)
		if err != nil {
			return c, bindError(err, strings.Join(loc[:i], "/"))
		}
	}

	return MakeContentMappingItem(c.t, c.k, m), nil
}

func (c *ContentItem) del(path []AttributeValue) (*ContentItem, error) {
	if len(c.k) <= 0 {
		return c, newInvalidContentModificationError()
	}

	if len(path) <= 0 {
		return c, newMissingPathContentModificationError()
	}

	if len(path) > len(c.k) {
		return c, newTooLongPathContentModificationError(c.k, path)
	}

	var err error
	m := c.r

	last := len(path) - 1
	branch := make([]ContentSubItem, last)

	loc := []string{""}

	for i, k := range path[:last] {
		branch[i] = m
		loc = append(loc, k.describe())

		m, err = m.next(k)
		if err != nil {
			return c, bindError(err, strings.Join(loc, "/"))
		}
	}

	k := path[last]
	loc = append(loc, k.describe())
	m, err = m.del(k)
	if err != nil {
		return c, bindError(err, strings.Join(loc, "/"))
	}

	for i := len(branch) - 1; i >= 0; i-- {
		p := branch[i]
		m, err = p.put(path[i], m)
		if err != nil {
			return c, bindError(err, strings.Join(loc[:i], "/"))
		}
	}

	return MakeContentMappingItem(c.t, c.k, m), nil
}

func (c *ContentItem) MarshalJSON() ([]byte, error) {
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

func (c *ContentItem) get(path []Expression, ctx *Context) (AttributeValue, error) {
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
	put(key AttributeValue, v ContentSubItem) (ContentSubItem, error)
	del(key AttributeValue) (ContentSubItem, error)
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

func (m contentStringMap) put(key AttributeValue, value ContentSubItem) (ContentSubItem, error) {
	k, err := key.str()
	if err != nil {
		return m, err
	}

	if v, ok := value.(contentValue); ok {
		return MakeContentStringMap(m.tree.Insert(k, v.value)), nil
	}

	return MakeContentStringMap(m.tree.Insert(k, value)), nil
}

func (m contentStringMap) del(key AttributeValue) (ContentSubItem, error) {
	k, err := key.str()
	if err != nil {
		return m, err
	}

	t, ok := m.tree.Delete(k)
	if !ok {
		return m, newMissingValueError()
	}

	return MakeContentStringMap(t), nil
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

func (m contentNetworkMap) getByAttribute(key AttributeValue) (interface{}, error) {
	if a, err := key.address(); err == nil {
		if v, ok := m.tree.GetByIP(a); ok {
			return v, nil
		}

		return undefinedValue, newMissingValueError()
	}

	if n, err := key.network(); err == nil {
		if v, ok := m.tree.GetByNet(n); ok {
			return v, nil
		}

		return undefinedValue, newMissingValueError()
	}

	return undefinedValue, newNetworkMapKeyValueTypeError(key.GetResultType())
}

func (m contentNetworkMap) getValue(key AttributeValue, t int) (AttributeValue, error) {
	v, err := m.getByAttribute(key)
	if err != nil {
		return undefinedValue, err
	}

	return MakeContentValue(v).getValue(undefinedValue, t)
}

func (m contentNetworkMap) next(key AttributeValue) (ContentSubItem, error) {
	v, err := m.getByAttribute(key)
	if err != nil {
		return nil, err
	}

	item, ok := v.(ContentSubItem)
	if !ok {
		return nil, newMapContentSubitemError()
	}

	return item, nil
}

func (m contentNetworkMap) put(key AttributeValue, value ContentSubItem) (ContentSubItem, error) {
	if a, err := key.address(); err == nil {
		if v, ok := value.(contentValue); ok {
			return MakeContentNetworkMap(m.tree.InsertIP(a, v.value)), nil
		}

		return MakeContentNetworkMap(m.tree.InsertIP(a, value)), nil
	}

	if n, err := key.network(); err == nil {
		if v, ok := value.(contentValue); ok {
			return MakeContentNetworkMap(m.tree.InsertNet(n, v.value)), nil
		}

		return MakeContentNetworkMap(m.tree.InsertNet(n, value)), nil
	}

	return nil, newNetworkMapKeyValueTypeError(key.GetResultType())
}

func (m contentNetworkMap) del(key AttributeValue) (ContentSubItem, error) {
	if a, err := key.address(); err == nil {
		if t, ok := m.tree.DeleteByIP(a); ok {
			return MakeContentNetworkMap(t), nil
		}

		return m, newMissingValueError()
	}

	if n, err := key.network(); err == nil {
		if t, ok := m.tree.DeleteByNet(n); ok {
			return MakeContentNetworkMap(t), nil
		}

		return m, newMissingValueError()
	}

	return nil, newNetworkMapKeyValueTypeError(key.GetResultType())
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

func (m contentDomainMap) put(key AttributeValue, value ContentSubItem) (ContentSubItem, error) {
	d, err := key.domain()
	if err != nil {
		return m, err
	}

	if v, ok := value.(contentValue); ok {
		return MakeContentDomainMap(m.tree.Insert(d, v.value)), nil
	}

	return MakeContentDomainMap(m.tree.Insert(d, value)), nil
}

func (m contentDomainMap) del(key AttributeValue) (ContentSubItem, error) {
	d, err := key.domain()
	if err != nil {
		return m, err
	}

	t, ok := m.tree.Delete(d)
	if !ok {
		return m, newMissingValueError()
	}

	return MakeContentDomainMap(t), nil
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

func (v contentValue) put(key AttributeValue, value ContentSubItem) (ContentSubItem, error) {
	return v, newInvalidContentValueModificationError()
}

func (v contentValue) del(key AttributeValue) (ContentSubItem, error) {
	return v, newInvalidContentValueModificationError()
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
