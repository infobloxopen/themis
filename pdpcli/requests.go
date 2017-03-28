package main

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"net"
	"os"
	"strconv"

	pb "github.com/infobloxopen/policy-box/pdp-service"
)

type Requests struct {
	Attributes map[string]string
	Requests   []map[string]interface{}
}

type Request struct {
	Index   int
	Request *pb.Request
	Error   error
}

func LoadRequests(name string) (*Requests, error) {
	f, err := os.Open(name)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	b, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, err
	}

	r := &Requests{}
	return r, yaml.Unmarshal(b, r)
}

func (r *Requests) Parse() (chan Request) {
	ch := make(chan Request)
	go func () {
		defer close(ch)

		for i, req := range r.Requests {
			attrs := []*pb.Attribute{}
			for name, value := range req {
				attr, err := makeAttribute(name, value, r.Attributes)
				if err != nil {
					ch <- Request{Index: i, Error: err}
					return
				}

				attrs = append(attrs, attr)
			}

			ch <- Request{Index: i, Request: &pb.Request{attrs}}
		}
	}()

	return ch
}

const (
	booleanAttribute = "boolean"
	stringAttribute  = "string"
	addressAttribute = "address"
	networkAttribute = "network"
	domainAttribute  = "domain"
)

type attributeMarshaller func(value interface{}) (string, error)

var marshallers = map[string]attributeMarshaller{
	booleanAttribute: booleanMarshaller,
	stringAttribute: stringMarshaller,
	addressAttribute: addressMarshaller,
	networkAttribute: networkMarshaller,
	domainAttribute: domainMarshaller}

func makeAttribute(name string, value interface{}, symbols map[string]string) (*pb.Attribute, error) {
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
		return nil, fmt.Errorf("unknown type \"%s\" of \"%s\" attribute", t, name)
	}

	s, err := marshaller(value)
	if err != nil {
		return nil, fmt.Errorf("can't marshal \"%s\" attribute as \"%s\": %s", name, t, err)
	}

	return &pb.Attribute{name, t, s}, nil
}

func guessType(value interface{}) (string, error) {
	switch value.(type) {
	case bool:
		return booleanAttribute, nil
	case string:
		return stringAttribute, nil
	case net.IP:
		return addressAttribute, nil
	case net.IPNet:
		return networkAttribute, nil
	case *net.IPNet:
		return networkAttribute, nil
	}

	return "", fmt.Errorf("marshaling hasn't been implemented for %T", value)
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
