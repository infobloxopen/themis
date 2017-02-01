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
)

func UnmarshalYAST(in []byte, dir string, ext map[string]interface{}) (EvaluableType, error) {
	m := make(map[interface{}]interface{})
	err := yaml.Unmarshal(in, &m)
	if err != nil {
		return nil, err
	}

	c := yastCtx{}
	c.nodeSpec = make([]string, 0)
	c.dataDir = dir

	if len(m) > 3 {
		return nil, c.makeRootError(m)
	}

	err = c.unmarshalAttributeDeclarations(m)
	if err != nil {
		return nil, err
	}

	err = c.unmarshalIncludes(m, ext)
	if err != nil {
		return nil, err
	}

	if v, ok := m[yastTagPolicies]; ok {
		c.pushNodeSpec(yastTagPolicies)
		defer c.popNodeSpec()

		p, err := c.unmarshalItem(v)
		if err != nil {
			return nil, err
		}

		return p, nil
	}

	return nil, c.makeRootError(m)
}

func UnmarshalYASTFromFile(name string, dir string) (EvaluableType, error) {
	f, err := findAndOpenFile(name, dir)
	if err != nil {
		return nil, err
	}

	b, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, err
	}

	return UnmarshalYAST(b, dir, nil)
}
