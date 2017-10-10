package main

import (
	"fmt"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/infobloxopen/themis/pdp"
	"github.com/infobloxopen/themis/pep"
)

func makeEffect(effect int) (byte, error) {
	switch effect {
	case pdp.EffectDeny:
		return pep.DENY, nil

	case pdp.EffectPermit:
		return pep.PERMIT, nil

	case pdp.EffectNotApplicable:
		return pep.NOTAPPLICABLE, nil

	case pdp.EffectIndeterminate:
		return pep.INDETERMINATE, nil

	case pdp.EffectIndeterminateD:
		return pep.INDETERMINATED, nil

	case pdp.EffectIndeterminateP:
		return pep.INDETERMINATEP, nil

	case pdp.EffectIndeterminateDP:
		return pep.INDETERMINATEDP, nil
	}

	return pep.INDETERMINATE, newUnknownEffectError(effect)
}

func makeFailEffect(effect byte) (byte, error) {
	switch effect {
	case pep.DENY:
		return pep.INDETERMINATED, nil

	case pep.PERMIT:
		return pep.INDETERMINATEP, nil

	case pep.NOTAPPLICABLE, pep.INDETERMINATE, pep.INDETERMINATED, pep.INDETERMINATEP, pep.INDETERMINATEDP:
		return effect, nil
	}

	return pep.INDETERMINATE, newUnknownEffectError(int(effect))
}

type obligations []*pep.Attribute

func (o obligations) String() string {
	if len(o) <= 0 {
		return "no attributes"
	}

	lines := []string{"attributes:"}
	for _, attr := range o {
		lines = append(lines, fmt.Sprintf("- %s.(%s): %q", attr.Id, attr.Type, attr.Value))
	}

	return strings.Join(lines, "\n")
}

func (s *server) newContext(c *pdp.LocalContentStorage, in *pep.Request) (*pdp.Context, error) {
	ctx, err := pdp.NewContext(c, len(in.Attributes), func(i int) (string, pdp.AttributeValue, error) {
		a := in.Attributes[i]

		t, ok := pdp.TypeIDs[strings.ToLower(a.Type)]
		if !ok {
			return "", pdp.AttributeValue{}, bindError(newUnknownAttributeTypeError(a.Type), a.Id)
		}

		v, err := pdp.MakeValueFromString(t, a.Value)
		if err != nil {
			return "", pdp.AttributeValue{}, bindError(err, a.Id)
		}

		return a.Id, v, nil
	})
	if err != nil {
		return nil, newContextCreationError(err)
	}

	return ctx, nil
}

func (s *server) newAttributes(obligations []pdp.AttributeAssignmentExpression, ctx *pdp.Context) ([]*pep.Attribute, error) {
	attrs := make([]*pep.Attribute, len(obligations))
	for i, e := range obligations {
		ID, t, s, err := e.Serialize(ctx)
		if err != nil {
			return attrs[:i], err
		}

		attrs[i] = &pep.Attribute{
			Id:    ID,
			Type:  t,
			Value: s}
	}

	return attrs, nil
}

func (s *server) rawValidate(p *pdp.PolicyStorage, c *pdp.LocalContentStorage, in *pep.Request) (byte, []error, []*pep.Attribute) {
	if p == nil {
		return pep.INDETERMINATE, []error{newMissingPolicyError()}, nil
	}

	ctx, err := s.newContext(c, in)
	if err != nil {
		return pep.INDETERMINATE, []error{err}, nil
	}

	log.WithField("context", ctx).Debug("Request context")

	errs := []error{}

	r := p.Root().Calculate(ctx)
	effect, obligations, err := r.Status()
	if err != nil {
		errs = append(errs, newPolicyCalculationError(err))
	}

	re, err := makeEffect(effect)
	if err != nil {
		errs = append(errs, newEffectTranslationError(err))
	}

	if len(errs) > 0 {
		re, err = makeFailEffect(re)
		if err != nil {
			errs = append(errs, newEffectCombiningError(err))
		}
	}

	attrs, err := s.newAttributes(obligations, ctx)
	if err != nil {
		errs = append(errs, newObligationTranslationError(err))
	}

	return re, errs, attrs
}

func (s *server) Validate(clientAddr string, request interface{}) interface{} {
	s.RLock()
	p := s.p
	c := s.c
	s.RUnlock()

	effect, errs, attrs := s.rawValidate(p, c, request.(*pep.Request))

	status := "Ok"
	if len(errs) > 1 {
		status = newMultiError(errs).Error()
	} else if len(errs) > 0 {
		status = errs[0].Error()
	}

	log.WithFields(log.Fields{
		"effect":     pep.EffectName(effect),
		"reason":     status,
		"obligation": obligations(attrs),
	}).Debug("Response")

	return &pep.Response{
		Effect:      effect,
		Reason:      status,
		Obligations: attrs,
	}
}
