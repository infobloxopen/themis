package jast

import (
	"encoding/json"
	"strings"

	"github.com/infobloxopen/themis/jparser"
	"github.com/infobloxopen/themis/pdp"
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

func (ctx *context) decode(d *json.Decoder) error {
	ok, err := jparser.CheckRootObjectStart(d)
	if err != nil {
		return err
	}

	if !ok {
		return nil
	}

	if err = jparser.UnmarshalObject(d, func(k string, d *json.Decoder) error {
		switch strings.ToLower(k) {
		case yastTagAttributes:
			return ctx.decodeAttributeDeclarations(d)

		case yastTagPolicies:
			return ctx.decodeRootPolicy(d)
		}

		return newUnknownAttributeError(k)
	}, "root"); err != nil {
		return err
	}

	if err = jparser.CheckEOF(d); err != nil {
		return err
	}

	return nil
}
