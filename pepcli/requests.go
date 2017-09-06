package main

//go:generate bash -c "mkdir -p $GOPATH/src/github.com/infobloxopen/themis/pdp-service && protoc -I $GOPATH/src/github.com/infobloxopen/themis/proto/ $GOPATH/src/github.com/infobloxopen/themis/proto/service.proto --go_out=plugins=grpc:$GOPATH/src/github.com/infobloxopen/themis/pdp-service && ls $GOPATH/src/github.com/infobloxopen/themis/pdp-service"

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"io"
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

	symbols map[string]int
}

type request struct {
	index    int
	position int
	request  pb.Request
	err      error
}

func loadRequests(name string) (*requests, error) {
	b, err := ioutil.ReadFile(name)
	if err != nil {
		return nil, err
	}

	r := &requests{}
	err = yaml.Unmarshal(b, r)
	if err != nil {
		return nil, err
	}

	r.symbols = make(map[string]int)
	for k, v := range r.Attributes {
		t, ok := pdp.TypeIDs[strings.ToLower(v)]
		if !ok {
			return nil, fmt.Errorf("unknown type \"%s\" of \"%s\" attribute", v, k)
		}

		r.symbols[k] = t
	}

	return r, nil
}

func (r *requests) parse(count int) chan request {
	ch := make(chan request)
	go func() {
		defer close(ch)

		length := len(r.Requests)
		if count <= 0 {
			count = length
		}

		for i := 0; i < count; i++ {
			j := i % length
			req := r.Requests[j]

			attrs := []*pb.Attribute{}
			for name, value := range req {
				attr, err := makeAttribute(name, value, r.symbols)
				if err != nil {
					ch <- request{index: i + 1, position: j + 1, err: err}
					return
				}

				attrs = append(attrs, attr)
			}

			ch <- request{
				index:    i + 1,
				position: j + 1,
				request: pb.Request{
					Attributes: attrs,
				},
			}
		}
	}()

	return ch
}

func dumpResponse(r *pb.Response, f io.Writer) error {
	lines := []string{fmt.Sprintf("- effect: %s", r.Effect.String())}
	if len(r.Reason) > 0 {
		lines = append(lines, fmt.Sprintf("  reason: %q", r.Reason))
	}

	if len(r.Obligation) > 0 {
		lines = append(lines, "  obligation:")
		for _, attr := range r.Obligation {
			lines = append(lines, fmt.Sprintf("    - id: %q", attr.Id))
			lines = append(lines, fmt.Sprintf("      type: %q", attr.Type))
			lines = append(lines, fmt.Sprintf("      value: %q", attr.Value))
			lines = append(lines, "")
		}
	} else {
		lines = append(lines, "")
	}

	_, err := fmt.Fprintf(f, "%s\n", strings.Join(lines, "\n"))
	return err
}

type attributeMarshaller func(value interface{}) (string, error)

var marshallers = map[int]attributeMarshaller{
	pdp.TypeBoolean: booleanMarshaller,
	pdp.TypeString:  stringMarshaller,
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
