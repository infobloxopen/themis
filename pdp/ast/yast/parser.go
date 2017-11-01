// Package yast implements policies YAML AST (YAST) parser.
package yast

import (
	"io"
	"io/ioutil"

	"github.com/google/uuid"
	"gopkg.in/yaml.v2"

	"github.com/infobloxopen/themis/pdp"
)

const (
	yastTagAttributes = "attributes"
	yastTagID         = "id"
	yastTagTarget     = "target"
	yastTagPolicies   = "policies"
	yastTagRules      = "rules"
	yastTagCondition  = "condition"
	yastTagAlg        = "alg"
	yastTagMap        = "map"
	yastTagDefault    = "default"
	yastTagError      = "error"
	yastTagEffect     = "effect"
	yastTagObligation = "obligations"
	yastTagAny        = "any"
	yastTagAll        = "all"
	yastTagAttribute  = "attr"
	yastTagValue      = "val"
	yastTagSelector   = "selector"
	yastTagType       = "type"
	yastTagContent    = "content"
	yastTagURI        = "uri"
	yastTagPath       = "path"
	yastTagOp         = "op"
	yastTagEntity     = "entity"
)

// Parser is a YAST parser implementation.
type Parser struct{}

// Unmarshal parses policies YAML representation to PDP's internal representation.
func (p Parser) Unmarshal(in io.Reader, tag *uuid.UUID) (*pdp.PolicyStorage, error) {
	b, err := ioutil.ReadAll(in)
	if err != nil {
		return nil, err
	}

	m := make(map[interface{}]interface{})
	err = yaml.Unmarshal(b, &m)
	if err != nil {
		return nil, err
	}

	if len(m) > 2 {
		return nil, newRootKeysError(m)
	}

	ctx := newContext()
	err = ctx.unmarshalAttributeDeclarations(m)
	if err != nil {
		return nil, err
	}

	rp, err := ctx.unmarshalRootPolicy(m)
	if err != nil {
		return nil, err
	}

	if rp != nil {
		return pdp.NewPolicyStorage(rp, ctx.attrs, tag), nil
	}

	return nil, newRootKeysError(m)
}

// UnmarshalUpdate parses policies update YAML representation to PDP's internal representation.
func (p Parser) UnmarshalUpdate(in io.Reader, attrs map[string]pdp.Attribute, oldTag, newTag uuid.UUID) (*pdp.PolicyUpdate, error) {
	b, err := ioutil.ReadAll(in)
	if err != nil {
		return nil, err
	}

	a := []interface{}{}
	err = yaml.Unmarshal(b, &a)
	if err != nil {
		return nil, err
	}

	ctx := newContextWithAttributes(attrs)

	u := pdp.NewPolicyUpdate(oldTag, newTag)
	for i, item := range a {
		err := ctx.unmarshalCommand(item, u)
		if err != nil {
			return nil, bindErrorf(err, "%d", i)
		}
	}

	return u, nil
}
