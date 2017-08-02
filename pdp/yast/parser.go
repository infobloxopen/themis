package yast

import (
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

	yastTagFirstApplicableEffectAlg = "firstapplicableeffect"
	yastTagDenyOverridesAlg         = "denyoverrides"
)

func Unmarshal(in []byte) (pdp.Evaluable, error) {
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
		return p, nil
	}

	return nil, newRootKeysError(m)
}
