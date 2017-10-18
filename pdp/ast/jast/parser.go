// Package jast implements policies JSON AST (JAST) parser.
package jast

import (
	"encoding/json"
	"io"

	"github.com/infobloxopen/themis/pdp"

	"github.com/google/uuid"
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

type Parser struct{}

func (p Parser) Unmarshal(in io.Reader, tag *uuid.UUID) (*pdp.PolicyStorage, error) {
	ctx := newContext()
	if err := ctx.decode(json.NewDecoder(in)); err != nil {
		return nil, err
	}

	return pdp.NewPolicyStorage(ctx.rootPolicy, ctx.attrs, tag), nil
}

func (p Parser) UnmarshalUpdate(in io.Reader, attrs map[string]pdp.Attribute, oldTag, newTag uuid.UUID) (*pdp.PolicyUpdate, error) {
	ctx := newContextWithAttributes(attrs)
	u := pdp.NewPolicyUpdate(oldTag, newTag)
	if err := ctx.decodeCommands(json.NewDecoder(in), u); err != nil {
		return nil, err
	}

	return u, nil
}
