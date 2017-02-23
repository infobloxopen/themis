package main

import (
	"fmt"
	"golang.org/x/net/context"
	"net"
	"strconv"
	"strings"

	log "github.com/Sirupsen/logrus"

	pb "github.com/infobloxopen/policy-box/pdp-service"

	"github.com/infobloxopen/policy-box/pdp"
)

type AttrMarshaller func(v pdp.AttributeValueType) (string, error)
type AttrUnmarshaller func(v string) (pdp.AttributeValueType, error)

var (
	Marshallers map[int]AttrMarshaller = map[int]AttrMarshaller {
		pdp.DataTypeUndefined: undefinedMarshaller,
		pdp.DataTypeBoolean:   booleanMarshaller,
		pdp.DataTypeString:    stringMarshaller,
		pdp.DataTypeAddress:   addressMarshaller,
		pdp.DataTypeNetwork:   networkMarshaller,
		pdp.DataTypeDomain:    domainMarshaller}

	Unmarshallers map[int]AttrUnmarshaller = map[int]AttrUnmarshaller{
		pdp.DataTypeUndefined: undefinedUnmarshaller,
		pdp.DataTypeBoolean:   booleanUnmarshaller,
		pdp.DataTypeString:    stringUnmarshaller,
		pdp.DataTypeAddress:   addressUnmarshaller,
		pdp.DataTypeNetwork:   networkUnmarshaller,
		pdp.DataTypeDomain:    domainUnmarshaller}
)

func undefinedMarshaller(v pdp.AttributeValueType) (string, error) {
	return "", fmt.Errorf("Can't marshal value of undefined type into response")
}

func undefinedUnmarshaller(v string) (pdp.AttributeValueType, error) {
	return pdp.AttributeValueType{}, fmt.Errorf("Can't unmarshal undefined type")
}

func booleanMarshaller(v pdp.AttributeValueType) (string, error) {
	return strconv.FormatBool(v.Value.(bool)), nil
}

func booleanUnmarshaller(v string) (pdp.AttributeValueType, error) {
	b, err := strconv.ParseBool(v)
	if err != nil {
		return pdp.AttributeValueType{}, err
	}

	return pdp.AttributeValueType{pdp.DataTypeBoolean, b}, nil
}

func stringMarshaller(v pdp.AttributeValueType) (string, error) {
	return v.Value.(string), nil
}

func stringUnmarshaller(v string) (pdp.AttributeValueType, error) {
	return pdp.AttributeValueType{pdp.DataTypeString, v}, nil
}

func addressMarshaller(v pdp.AttributeValueType) (string, error) {
	return v.Value.(net.IP).String(), nil
}

func addressUnmarshaller(v string) (pdp.AttributeValueType, error) {
	if strings.Contains(v, ":") {
		if strings.Contains(v, "]") {
			v = strings.Split(v, "]")[0][1:]
		} else if strings.Contains(v, ".") {
			v = strings.Split(v, ":")[0]
		}
	}

	ip := net.ParseIP(v)
	if ip == nil {
		return pdp.AttributeValueType{}, fmt.Errorf("Can't treat \"%s\" as address", v)
	}

	return pdp.AttributeValueType{pdp.DataTypeAddress, ip}, nil
}

func networkMarshaller(v pdp.AttributeValueType) (string, error) {
	n := v.Value.(net.IPNet)
	return n.String(), nil
}

func networkUnmarshaller(v string) (pdp.AttributeValueType, error) {
	_, n, err := net.ParseCIDR(v)
	if err != nil {
		return pdp.AttributeValueType{}, err
	}

	return pdp.AttributeValueType{pdp.DataTypeNetwork, *n}, nil
}

func domainMarshaller(v pdp.AttributeValueType) (string, error) {
	return v.Value.(string), nil
}

func domainUnmarshaller(v string) (pdp.AttributeValueType, error) {
	d, err := pdp.AdjustDomainName(v)
	if err != nil {
		return pdp.AttributeValueType{}, err
	}

	return pdp.AttributeValueType{pdp.DataTypeDomain, d}, nil
}

func UnmarshalAttribute(attr *pb.Attribute) (int, pdp.AttributeValueType) {
	dataType, ok := pdp.DataTypeIDs[strings.ToLower(attr.Type)]
	if !ok {
		log.WithFields(log.Fields{
			"type": attr.Type,
			"id":   attr.Id}).Error("Invalid type name")

		return pdp.DataTypeUndefined, pdp.AttributeValueType{}
	}

	unmarshaller, ok := Unmarshallers[dataType]
	if !ok {
		log.WithFields(log.Fields{
			"type": attr.Type,
			"id":   attr.Id}).Error("Unmarshaling for the type from the request hasn't been implemented yet")

		return pdp.DataTypeUndefined, pdp.AttributeValueType{}
	}

	v, err := unmarshaller(attr.Value)
	if err != nil {
		log.WithFields(log.Fields{
			"value": attr.Value,
			"type": attr.Type,
			"id": attr.Id,
			"error": err}).Error("Unmarshaling error")

		return pdp.DataTypeUndefined, pdp.AttributeValueType{}
	}

	return dataType, v
}

func MakeRequestContext(in *pb.Request) pdp.Context {
	ctx := pdp.NewContext()
	for _, attr := range in.Attributes {
		t, v := UnmarshalAttribute(attr)
		if t != pdp.DataTypeUndefined {
			ctx.StoreRawAttribute(attr.Id, t, v)
		}
	}

	return ctx
}

func MarshalAttribute(attr pdp.AttributeValueType) (string, error) {
	marshaller, ok := Marshallers[attr.DataType]
	if !ok {
		return "", fmt.Errorf("Marshaling for type \"%s\" into response hasn't been implemented yet",
			pdp.DataTypeNames[attr.DataType])
	}

	return marshaller(attr)
}

func MarshalAttributes(ctx *pdp.Context) []*pb.Attribute {
	if ctx == nil {
		return nil
	}

	attrs := make([]*pb.Attribute, 0)
	for id, t := range ctx.Attributes {
		for _, v := range t {
			s, err := MarshalAttribute(v)
			if err != nil {
				log.WithFields(log.Fields{
					"type": pdp.DataTypeNames[v.DataType],
					"id": id,
					"error": err}).Error("Marshaling error")

				continue
			}

			attrs = append(attrs, &pb.Attribute{id, pdp.DataTypeNames[v.DataType], s})
		}
	}

	return attrs
}

func serviceReply(e int, s string, ctx *pdp.Context) *pb.Response {
	r := &pb.Response{pb.Response_DENY, s, MarshalAttributes(ctx)}

	if e == pdp.EffectPermit {
		r.Effect = pb.Response_PERMIT
	}

	return r
}

func serviceFail(r pdp.ResponseType, format string, args ...interface{}) *pb.Response {
	s := fmt.Sprintf(format, args...)

	if r.Effect == pdp.EffectDeny || r.Effect == pdp.EffectIndeterminateD {
		return serviceReply(pdp.EffectIndeterminateD, s, nil)
	}

	if r.Effect == pdp.EffectPermit || r.Effect == pdp.EffectIndeterminateP {
		return serviceReply(pdp.EffectIndeterminateP, s, nil)
	}

	if r.Effect == pdp.EffectIndeterminateDP {
		return serviceReply(pdp.EffectIndeterminateP, s, nil)
	}

	return serviceReply(pdp.EffectIndeterminate, s, nil)
}

func MakeResponse(r pdp.ResponseType, ctx *pdp.Context) *pb.Response {
	o := pdp.NewContext()

	err := o.CalculateObligations(r.Obligations, ctx)
	if err != nil {
		return serviceFail(r, "Obligations: %s", err)
	}

	return serviceReply(r.Effect, r.Status, &o)
}

func (s *Server) Validate(server_ctx context.Context, in *pb.Request) (*pb.Response, error) {
	ctx := MakeRequestContext(in)
	log.Info("Validating context")

	s.Lock.RLock()
	p := s.Policy
	s.Lock.RUnlock()

	if p == nil {
		return serviceReply(pdp.EffectIndeterminate, "No policy or policy set defined", nil), nil
	}

	r := p.Calculate(&ctx)
	log.Info("Returning response")

	return MakeResponse(r, &ctx), nil
}
