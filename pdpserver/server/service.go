package server

import (
	"fmt"
	"strings"

	log "github.com/sirupsen/logrus"
	"golang.org/x/net/context"

	"github.com/infobloxopen/themis/pdp"
	pb "github.com/infobloxopen/themis/pdp-service"
)

func (s *Server) newContext(c *pdp.LocalContentStorage, in []byte) (*pdp.Context, error) {
	ctx, err := pdp.NewContextFromBytes(c, in)
	if err != nil {
		return nil, newContextCreationError(err)
	}

	return ctx, nil
}

func makeFailureResponse(b []byte, err error) []byte {
	n, err := pdp.MakeIndeterminateResponse(b, err)
	if err != nil {
		panic(err)
	}

	return b[:n]
}

func (s *Server) rawValidate(p *pdp.PolicyStorage, c *pdp.LocalContentStorage, in []byte, out []byte) []byte {
	if out == nil {
		out = make([]byte, s.opts.maxResponseSize)
	}

	if p == nil {
		return makeFailureResponse(out, newMissingPolicyError())
	}

	ctx, err := s.newContext(c, in)
	if err != nil {
		return makeFailureResponse(out, err)
	}

	if s.opts.logger.Level >= log.DebugLevel {
		s.opts.logger.WithField("context", ctx).Debug("Request context")
	}

	r := p.Root().Calculate(ctx)

	if s.opts.logger.Level >= log.DebugLevel {
		s.opts.logger.WithFields(log.Fields{
			"effect": pdp.EffectNameFromEnum(r.Effect),
			"reason": r.Status,
			"obligations": obligations{
				ctx: ctx,
				o:   r.Obligations,
			},
		}).Debug("Response")
	}

	n, err := r.Marshal(out, ctx)
	if err != nil {
		panic(err)
	}

	return out[:n]
}

// Validate is a server handler for gRPC call
// It handles PDP decision requests
func (s *Server) Validate(ctx context.Context, in *pb.Msg) (*pb.Msg, error) {
	s.RLock()
	p := s.p
	c := s.c
	s.RUnlock()

	return &pb.Msg{
		Body: s.rawValidate(p, c, in.Body, nil),
	}, nil
}

type obligations struct {
	ctx *pdp.Context
	o   []pdp.AttributeAssignmentExpression
}

func (o obligations) String() string {
	if len(o.o) <= 0 {
		return "no attributes"
	}

	lines := make([]string, len(o.o)+1)
	lines[0] = "attributes:"
	for i, e := range o.o {
		id, t, v, err := e.Serialize(o.ctx)
		if err != nil {
			lines[i+1] = fmt.Sprintf("- %d: %s", i+1, err)
		} else {
			lines[i+1] = fmt.Sprintf("- %s.(%s): %q", id, t, v)
		}
	}

	return strings.Join(lines, "\n")
}
