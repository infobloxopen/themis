package pdp

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"
)

const (
	yastTagAttributes = "attributes"
	yastTagInclude    = "include"
	yastTagPolicies   = "policies"
	yastTagID         = "id"
	yastTagAlg        = "alg"
	yastTagRules      = "rules"
	yastTagTarget     = "target"
	yastTagCondition  = "condition"
	yastTagObligation = "obligations"
	yastTagEffect     = "effect"
	yastTagAttribute  = "attr"
	yastTagValue      = "val"
	yastTagType       = "type"
	yastTagContent    = "content"
	yastTagSelector   = "selector"
	yastTagPath       = "path"
	yastTagMap        = "map"
	yastTagDefault    = "default"
	yastTagError      = "error"

	yastTagDataTypeUndefined     = "undefined"
	yastTagDataTypeBoolean       = "boolean"
	yastTagDataTypeString        = "string"
	yastTagDataTypeAddress       = "address"
	yastTagDataTypeNetwork       = "network"
	yastTagDataTypeDomain        = "domain"
	yastTagDataTypeSetOfStrings  = "set of strings"
	yastTagDataTypeSetOfNetworks = "set of networks"
	yastTagDataTypeSetOfDomains  = "set of domains"

	yastTagFirstApplicableEffectAlg = "firstapplicableeffect"
	yastTagDenyOverridesAlg         = "denyoverrides"
	yastTagMapperAlg                = "mapper"
	yastTagDefaultAlg               = yastTagFirstApplicableEffectAlg

	yastExpressionEqual    = "equal"
	yastExpressionContains = "contains"
	yastExpressionNot      = "not"
	yastExpressionOr       = "or"
	yastExpressionAnd      = "and"
)

func (ctx *YastCtx) UnmarshalYAST(in []byte, ext map[string]interface{}) (EvaluableType, error) {
	m := make(map[interface{}]interface{})
	err := yaml.Unmarshal(in, &m)
	if err != nil {
		return nil, err
	}

	if len(m) > 3 {
		return nil, ctx.makeRootError(m)
	}

	err = ctx.unmarshalAttributeDeclarations(m)
	if err != nil {
		return nil, err
	}

	err = ctx.unmarshalIncludes(m, ext)
	if err != nil {
		return nil, err
	}

	if v, ok := m[yastTagPolicies]; ok {
		ctx.pushNodeSpec(yastTagPolicies)
		defer ctx.popNodeSpec()

		p, err := ctx.unmarshalItem(v)
		if err != nil {
			return nil, err
		}

		return p, nil
	}

	return nil, ctx.makeRootError(m)
}

func (ctx *YastCtx) UnmarshalYASTFromFile(name string) (EvaluableType, error) {
	f, err := findAndOpenFile(name, ctx.dataDir)
	if err != nil {
		return nil, err
	}

	b, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, err
	}

	return ctx.UnmarshalYAST(b, nil)
}
