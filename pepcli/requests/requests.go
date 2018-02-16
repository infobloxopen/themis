// Package requests provides loader for YAML formatted authorization requests file.
package requests

//go:generate bash -c "mkdir -p $GOPATH/src/github.com/infobloxopen/themis/pdp-service && protoc -I $GOPATH/src/github.com/infobloxopen/themis/proto/ $GOPATH/src/github.com/infobloxopen/themis/proto/service.proto --go_out=plugins=grpc:$GOPATH/src/github.com/infobloxopen/themis/pdp-service && ls $GOPATH/src/github.com/infobloxopen/themis/pdp-service"

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"net"
	"strconv"
	"strings"

	"github.com/infobloxopen/themis/pdp"
	pb "github.com/infobloxopen/themis/pdp-service"
)

type requests struct {
	Attributes map[string]string
	Requests   []map[string]interface{}
}

// Load reads given YAML file and porduces list of requests to run.
func Load(name string) ([]pb.Request, error) {
	b, err := ioutil.ReadFile(name)
	if err != nil {
		return nil, err
	}

	in := &requests{}
	err = yaml.Unmarshal(b, in)
	if err != nil {
		return nil, err
	}

	symbols := make(map[string]int)
	for k, v := range in.Attributes {
		t, ok := pdp.TypeIDs[strings.ToLower(v)]
		if !ok {
			return nil, fmt.Errorf("unknown type \"%s\" of \"%s\" attribute", v, k)
		}

		symbols[k] = t
	}

	out := make([]pb.Request, len(in.Requests))
	for i, r := range in.Requests {
		attrs := make([]*pb.Attribute, len(r))
		j := 0
		for k, v := range r {
			a, err := makeAttribute(k, v, symbols)
			if err != nil {
				return nil, fmt.Errorf("invalid attribute in request %d: %s", i+1, err)
			}

			attrs[j] = a
			j++
		}

		out[i] = pb.Request{Attributes: attrs}
	}

	return out, nil
}

type attributeMarshaller func(value interface{}) (string, error)

var marshallers = map[int]attributeMarshaller{
	pdp.TypeBoolean: booleanMarshaller,
	pdp.TypeString:  stringMarshaller,
	pdp.TypeInteger: integerMarshaller,
	pdp.TypeFloat:   floatMarshaller,
	pdp.TypeAddress: addressMarshaller,
	pdp.TypeNetwork: networkMarshaller,
	pdp.TypeDomain:  domainMarshaller}

func makeAttribute(name string, value interface{}, symbols map[string]int) (*pb.Attribute, error) {
	t, ok := symbols[name]
	var err error
	if !ok {
		t, err = guessType(value)
		if err != nil {
			return nil, fmt.Errorf("type of \"%s\" attribute isn't defined and can't be derived: %s", name, err)
		}
	}

	marshaller, ok := marshallers[t]
	if !ok {
		return nil, fmt.Errorf("marshaling hasn't been implemented for type \"%s\" of \"%s\" attribute",
			pdp.TypeNames[t], name)
	}

	s, err := marshaller(value)
	if err != nil {
		return nil, fmt.Errorf("can't marshal \"%s\" attribute as \"%s\": %s", name, pdp.TypeNames[t], err)
	}

	return &pb.Attribute{
		Id:    name,
		Type:  pdp.TypeKeys[t],
		Value: s,
	}, nil
}

func guessType(value interface{}) (int, error) {
	switch value.(type) {
	case bool:
		return pdp.TypeBoolean, nil
	case string:
		return pdp.TypeString, nil
	case net.IP:
		return pdp.TypeAddress, nil
	case net.IPNet:
		return pdp.TypeNetwork, nil
	case *net.IPNet:
		return pdp.TypeNetwork, nil
	}

	return 0, fmt.Errorf("marshaling hasn't been implemented for %T", value)
}

func booleanMarshaller(value interface{}) (string, error) {
	switch value := value.(type) {
	case bool:
		return strconv.FormatBool(value), nil
	case string:
		_, err := strconv.ParseBool(value)
		if err != nil {
			return "", fmt.Errorf("can't marshal \"%s\" as boolean", value)
		}

		return value, nil
	}

	return "", fmt.Errorf("can't marshal %T as boolean", value)
}

func stringMarshaller(value interface{}) (string, error) {
	s, ok := value.(string)
	if !ok {
		return "", fmt.Errorf("can't marshal %T as string", value)
	}

	return s, nil
}

func integerMarshaller(value interface{}) (string, error) {
	var err error
	switch value := value.(type) {
	case int:
		return strconv.FormatInt(int64(value), 10), nil
	case int64:
		return strconv.FormatInt(value, 10), nil
	case uint:
		return strconv.FormatInt(int64(value), 10), nil
	case uint64:
		return strconv.FormatInt(int64(value), 10), nil
	case float64:
		if value > -9007199254740992 && value < 9007199254740992 {
			return strconv.FormatInt(int64(value), 10), nil
		}
		err = fmt.Errorf("can't marshal %T as int64", value)

	case string:
		_, err = strconv.ParseInt(value, 10, 64)
		if err == nil {
			return value, nil
		}
		err = fmt.Errorf("can't marshal \"%s\" as int64", value)
	}

	return "", err
}

func floatMarshaller(value interface{}) (string, error) {
	var err error
	switch value := value.(type) {
	case int:
		return strconv.FormatFloat(float64(value), 'g', -1, 64), nil
	case int64:
		return strconv.FormatFloat(float64(value), 'g', -1, 64), nil
	case uint:
		return strconv.FormatFloat(float64(value), 'g', -1, 64), nil
	case uint64:
		return strconv.FormatFloat(float64(value), 'g', -1, 64), nil
	case float64:
		return strconv.FormatFloat(value, 'g', -1, 64), nil
	case string:
		_, err = strconv.ParseFloat(value, 64)
		if err == nil {
			return value, nil
		}
		err = fmt.Errorf("can't marshal \"%s\" as float64", value)
	}

	return "", err
}

func addressMarshaller(value interface{}) (string, error) {
	switch value := value.(type) {
	case net.IP:
		return value.String(), nil
	case string:
		addr := net.ParseIP(value)
		if addr == nil {
			return "", fmt.Errorf("can't marshal \"%s\" as IP address", value)
		}

		return value, nil
	}

	return "", fmt.Errorf("can't marshal %T as IP address", value)
}

func networkMarshaller(value interface{}) (string, error) {
	switch value := value.(type) {
	case net.IPNet:
		return (&value).String(), nil
	case *net.IPNet:
		return value.String(), nil
	case string:
		_, _, err := net.ParseCIDR(value)
		if err != nil {
			return "", fmt.Errorf("can't marshal \"%s\" as network", value)
		}

		return value, nil
	}

	return "", fmt.Errorf("can't marshal %T as network", value)
}

func domainMarshaller(value interface{}) (string, error) {
	s, ok := value.(string)
	if !ok {
		return "", fmt.Errorf("can't marshal %T as domain", value)
	}

	return s, nil
}
