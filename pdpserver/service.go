package main

import (
	"fmt"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/infobloxopen/themis/pdp"
	rpc "github.com/infobloxopen/themis/pdp-service"
)

func makeEffect(effect int) (byte, error) {
	switch effect {
	case pdp.EffectDeny:
		return rpc.DENY, nil

	case pdp.EffectPermit:
		return rpc.PERMIT, nil

	case pdp.EffectNotApplicable:
		return rpc.NOTAPPLICABLE, nil

	case pdp.EffectIndeterminate:
		return rpc.INDETERMINATE, nil

	case pdp.EffectIndeterminateD:
		return rpc.INDETERMINATED, nil

	case pdp.EffectIndeterminateP:
		return rpc.INDETERMINATEP, nil

	case pdp.EffectIndeterminateDP:
		return rpc.INDETERMINATEDP, nil
	}

	return rpc.INDETERMINATE, newUnknownEffectError(effect)
}

func makeFailEffect(effect byte) (byte, error) {
	switch effect {
	case rpc.DENY:
		return rpc.INDETERMINATED, nil

	case rpc.PERMIT:
		return rpc.INDETERMINATEP, nil

	case rpc.NOTAPPLICABLE, rpc.INDETERMINATE, rpc.INDETERMINATED, rpc.INDETERMINATEP, rpc.INDETERMINATEDP:
		return effect, nil
	}

	return rpc.INDETERMINATE, newUnknownEffectError(int(effect))
}

type obligations []*rpc.Attribute

func (o obligations) String() string {
	if len(o) <= 0 {
		return "no attributes"
	}

	lines := []string{"attributes:"}
	for _, attr := range o {
		lines = append(lines, fmt.Sprintf("- %s.(%s): %q", attr.Id(), attr.Type(), attr.Value()))
	}

	return strings.Join(lines, "\n")
}

func (s *server) newContext(c *pdp.LocalContentStorage, in *rpc.Request) (*pdp.Context, error) {
	ctx, err := pdp.NewContext(c, len(in.Attributes), func(i int) (string, pdp.AttributeValue, error) {
		a := in.Attributes[i]

		t, ok := pdp.TypeIDs[strings.ToLower(a.Type())]
		if !ok {
			return "", pdp.AttributeValue{}, bindError(newUnknownAttributeTypeError(a.Type()), a.Id())
		}

		v, err := pdp.MakeValueFromString(t, a.Value())
		if err != nil {
			return "", pdp.AttributeValue{}, bindError(err, a.Id())
		}

		return a.Id(), v, nil
	})
	if err != nil {
		return nil, newContextCreationError(err)
	}

	return ctx, nil
}

func (s *server) newAttributes(obligations []pdp.AttributeAssignmentExpression, ctx *pdp.Context) ([]*rpc.Attribute, error) {
	attrs := make([]*rpc.Attribute, len(obligations))
	for i, e := range obligations {
		ID, t, s, err := e.Serialize(ctx)
		if err != nil {
			return attrs[:i], err
		}

		attrs[i] = &rpc.Attribute{ID, t, s}
	}

	return attrs, nil
}

func (s *server) rawValidate(p *pdp.PolicyStorage, c *pdp.LocalContentStorage, in *rpc.Request) (byte, []error, []*rpc.Attribute) {
	if p == nil {
		return rpc.INDETERMINATE, []error{newMissingPolicyError()}, nil
	}

	ctx, err := s.newContext(c, in)
	if err != nil {
		return rpc.INDETERMINATE, []error{err}, nil
	}

	if s.logLevel >= log.DebugLevel {
		log.WithField("context", ctx).Debug("Request context")
	}

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

	effect, errs, attrs := s.rawValidate(p, c, request.(*rpc.Request))

	status := "Ok"
	if len(errs) > 1 {
		status = newMultiError(errs).Error()
	} else if len(errs) > 0 {
		status = errs[0].Error()
	}

	if s.logLevel >= log.DebugLevel {
		log.WithFields(log.Fields{
			"effect":     rpc.EffectName(effect),
			"reason":     status,
			"obligation": obligations(attrs),
		}).Debug("Response")
	}

	return &rpc.Response{
		Effect:      effect,
		Reason:      status,
		Obligations: attrs,
	}
}
