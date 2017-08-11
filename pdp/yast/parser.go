package yast

import (
	"github.com/satori/go.uuid"
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

	yastTagFirstApplicableEffectAlg = "firstapplicableeffect"
	yastTagDenyOverridesAlg         = "denyoverrides"
)

func Unmarshal(in []byte, tag *uuid.UUID) (*pdp.PolicyStorage, error) {
	m := make(map[interface{}]interface{})
	err := yaml.Unmarshal(in, &m)
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

	p, err := ctx.unmarshalRootPolicy(m)
	if err != nil {
		return nil, err
	}

	if p != nil {
		return pdp.NewPolicyStorage(p, ctx.attrs, tag), nil
	}

	return nil, newRootKeysError(m)
}

func UnmarshalUpdate(in []byte, attrs map[string]pdp.Attribute, oldTag, newTag uuid.UUID) (*pdp.PolicyUpdate, error) {
	a := []interface{}{}
	err := yaml.Unmarshal(in, &a)
	if err != nil {
		return nil, err
	}

	ctx := newContextWithAttributes(attrs)
	u, err := pdp.NewPolicyUpdate(oldTag, newTag)
	if err != nil {
		return nil, err
	}

	for i, item := range a {
		err := ctx.unmarshalCommand(item, u)
		if err != nil {
			return nil, bindErrorf(err, "%d", i)
		}
	}

	return u, nil
}
