package jast

import (
	"encoding/json"
	"strings"

	"github.com/infobloxopen/themis/pdp"
	"github.com/infobloxopen/themis/pdp/jcon"
)

type context struct {
	attrs      map[string]pdp.Attribute
	rootPolicy pdp.Evaluable
}

func newContext() *context {
	return &context{attrs: make(map[string]pdp.Attribute)}
}

func newContextWithAttributes(attrs map[string]pdp.Attribute) *context {
	return &context{attrs: attrs}
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

func (ctx context) extractList(m map[interface{}]interface{}, k, desc string) ([]interface{}, boundError) {
	v, ok := m[k]
	if !ok {
		return nil, newMissingListError(desc)
	}

	return ctx.validateList(v, desc)
}

func (ctx context) extractListOpt(m map[interface{}]interface{}, k, desc string) ([]interface{}, bool, boundError) {
	v, ok := m[k]
	if !ok {
		return nil, false, nil
	}

	l, err := ctx.validateList(v, desc)
	return l, true, err
}

func (ctx context) validateList(v interface{}, desc string) ([]interface{}, boundError) {
	r, ok := v.([]interface{})
	if !ok {
		return nil, newListError(v, desc)
	}

	return r, nil
}

func (ctx context) getSingleMapPair(m map[interface{}]interface{}, desc string) (interface{}, interface{}, boundError) {
	if len(m) > 1 {
		return nil, nil, newTooManySMPItemsError(desc, len(m))
	}

	for k, v := range m {
		return k, v, nil
	}

	return nil, nil, newNoSMPItemsError(desc, len(m))
}

func (ctx *context) decode(d *json.Decoder) error {
	ok, err := jcon.CheckRootObjectStart(d)
	if err != nil {
		return err
	}

	if !ok {
		return nil
	}

	err = jcon.UnmarshalObject(d, func(k string, d *json.Decoder) error {
		switch strings.ToLower(k) {
		case yastTagAttributes:
			return ctx.decodeAttributeDeclarations(d)

		case yastTagPolicies:
			return ctx.decodeRootPolicy(d)
		}

		return newUnknownAttributeError(k)
	}, "root")
	if err != nil {
		return err
	}

	err = jcon.CheckEOF(d)
	if err != nil {
		return err
	}

	return nil
}

func pairs2map(pairs []jcon.Pair) map[interface{}]interface{} {
	m := make(map[interface{}]interface{}, len(pairs))
	for _, p := range pairs {
		switch v := p.V.(type) {
		case []jcon.Pair:
			m[p.K] = pairs2map(v)
		case []interface{}:
			for i, item := range v {
				if pairs, ok := item.([]jcon.Pair); ok {
					v[i] = pairs2map(pairs)
				}
			}
			m[p.K] = v
		default:
			m[p.K] = v
		}
	}
	return m
}

func (ctx *context) decodeObject(d *json.Decoder, desc string) (map[interface{}]interface{}, error) {
	pairs, err := jcon.GetObject(d, desc)
	if err != nil {
		return nil, err
	}

	return pairs2map(pairs), nil
}

func (ctx *context) decodeArray(d *json.Decoder, desc string) ([]interface{}, error) {
	arr, err := jcon.GetArray(d, desc)
	if err != nil {
		return nil, err
	}

	for i, item := range arr {
		if pairs, ok := item.([]jcon.Pair); ok {
			arr[i] = pairs2map(pairs)
		}
	}

	return arr, nil
}

func (ctx *context) decodeUndefined(d *json.Decoder, desc string) (interface{}, error) {
	v, err := jcon.GetUndefined(d, desc)
	if err != nil {
		return nil, err
	}

	switch v := v.(type) {
	case []jcon.Pair:
		return pairs2map(v), nil
	case []interface{}:
		for i, item := range v {
			if pairs, ok := item.([]jcon.Pair); ok {
				v[i] = pairs2map(pairs)
			}
		}
		return v, nil
	default:
		return v, nil
	}
}
