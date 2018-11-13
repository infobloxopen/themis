package main

import (
	"fmt"
	"math"
	"net"
	"os"
	"strings"

	"gopkg.in/yaml.v2"

	"github.com/infobloxopen/go-trees/domain"
	"github.com/infobloxopen/go-trees/domaintree"
	"github.com/infobloxopen/go-trees/iptree"
	"github.com/infobloxopen/go-trees/strtree"
	"github.com/infobloxopen/themis/pdp"
)

type request struct {
	path string
	args []pdp.AttributeValue
}

type rawRequest struct {
	Path string
	Args []interface{}
}

func loadRequests(path string) (out []request, err error) {
	f, err := os.Open(path)
	if err != nil {
		return
	}
	defer func() {
		if cErr := f.Close(); err == nil {
			err = cErr
		}
	}()

	var v []rawRequest
	if err = yaml.NewDecoder(f).Decode(&v); err != nil {
		return
	}

	out, err = parseRequests(v)
	return
}

func parseRequests(in []rawRequest) ([]request, error) {
	out := make([]request, 0, len(in))
	for i, rr := range in {
		r, err := parseRequest(i, rr)
		if err != nil {
			return nil, err
		}

		out = append(out, r)
	}

	return out, nil
}

func parseRequest(i int, in rawRequest) (request, error) {
	args := make([]pdp.AttributeValue, 0, len(in.Args))
	for j, rv := range in.Args {
		v, err := parseValue(i, j, rv)
		if err != nil {
			return request{}, err
		}

		args = append(args, v)
	}

	return request{
		path: in.Path,
		args: args,
	}, nil
}

func parseValue(i, j int, v interface{}) (pdp.AttributeValue, error) {
	if m, ok := v.(map[interface{}]interface{}); ok {
		return parseComplexValue(i, j, m)
	}

	return parseSimpleValue(i, j, v)
}

func parseSimpleValue(i, j int, v interface{}) (pdp.AttributeValue, error) {
	switch v := v.(type) {
	case bool:
		return pdp.MakeBooleanValue(v), nil

	case string:
		return pdp.MakeStringValue(v), nil

	case int:
		if v >= math.MinInt64 && v <= math.MaxInt64 {
			return pdp.MakeIntegerValue(int64(v)), nil
		}

		return pdp.UndefinedValue, fmt.Errorf("%d overflows integer at (%d:%d)", v, i+1, j+1)

	case int8:
		return pdp.MakeIntegerValue(int64(v)), nil

	case int16:
		return pdp.MakeIntegerValue(int64(v)), nil

	case int32:
		return pdp.MakeIntegerValue(int64(v)), nil

	case int64:
		return pdp.MakeIntegerValue(v), nil

	case uint:
		if v <= math.MaxInt64 {
			return pdp.MakeIntegerValue(int64(v)), nil
		}

		return pdp.UndefinedValue, fmt.Errorf("%d overflows integer at (%d:%d)", v, i+1, j+1)

	case uint8:
		return pdp.MakeIntegerValue(int64(v)), nil

	case uint16:
		return pdp.MakeIntegerValue(int64(v)), nil

	case uint32:
		return pdp.MakeIntegerValue(int64(v)), nil

	case uint64:
		if v <= math.MaxInt64 {
			return pdp.MakeIntegerValue(int64(v)), nil
		}

		return pdp.UndefinedValue, fmt.Errorf("%d overflows integer at (%d:%d)", v, i+1, j+1)

	case float32:
		return pdp.MakeFloatValue(float64(v)), nil

	case float64:
		return pdp.MakeFloatValue(v), nil

	case []interface{}:
		ss := make([]string, 0, len(v))
		for k, v := range v {
			s, ok := v.(string)
			if !ok {
				return pdp.UndefinedValue,
					fmt.Errorf("expected string as list of strings item but got %T at (%d:%d:%d)",
						v, i+1, j+1, k+1)
			}

			ss = append(ss, s)
		}

		return pdp.MakeListOfStringsValue(ss), nil
	}

	return pdp.UndefinedValue, fmt.Errorf("expected value but got %T at (%d:%d)", v, i+1, j+1)
}

func parseComplexValue(i, j int, m map[interface{}]interface{}) (pdp.AttributeValue, error) {
	v, ok := m["content"]
	if !ok {
		return pdp.UndefinedValue, fmt.Errorf("missing content for value at (%d:%d)", i+1, j+1)
	}

	if t, ok := m["type"]; ok {
		s, ok := t.(string)
		if !ok {
			return pdp.UndefinedValue,
				fmt.Errorf("expected string as type for value at (%d:%d) but got %T", i+1, j+1, t)
		}

		if f, ok := typedValueParsers[strings.ToLower(s)]; ok {
			return f(i, j, v)
		}

		return pdp.UndefinedValue, fmt.Errorf("unknown type %q for value at (%d:%d)", t, i+1, j+1)
	}

	return parseSimpleValue(i, j, v)
}

type typedValueParser func(int, int, interface{}) (pdp.AttributeValue, error)

var typedValueParsers = map[string]typedValueParser{
	"boolean":         parseBooleanValue,
	"string":          parseStringValue,
	"integer":         parseIntegerValue,
	"float":           parseFloatValue,
	"address":         parseAddressValue,
	"network":         parseNetworkValue,
	"domain":          parseDomainValue,
	"set of strings":  parseSetOfStringsValue,
	"set of networks": parseSetOfNetworksValue,
	"set of domains":  parseSetOfDomainsValue,
	"list of strings": parseListOfStringsValue,
}

func parseBooleanValue(i, j int, v interface{}) (pdp.AttributeValue, error) {
	switch v := v.(type) {
	case bool:
		return pdp.MakeBooleanValue(v), nil

	case string:
		out, err := pdp.MakeValueFromString(pdp.TypeBoolean, v)
		if err != nil {
			return pdp.UndefinedValue, fmt.Errorf("can't make boolean value from %q at (%d:%d)", v, i+1, j+1)
		}

		return out, nil
	}

	return pdp.UndefinedValue, fmt.Errorf("expected boolean value but got %T at (%d:%d)", v, i+1, j+1)
}

func parseStringValue(i, j int, v interface{}) (pdp.AttributeValue, error) {
	if s, ok := v.(string); ok {
		return pdp.MakeStringValue(s), nil
	}

	return pdp.UndefinedValue, fmt.Errorf("expected string but got %T at (%d:%d)", v, i+1, j+1)
}

func parseIntegerValue(i, j int, v interface{}) (pdp.AttributeValue, error) {
	switch v := v.(type) {
	case int:
		if v >= math.MinInt64 && v <= math.MaxInt64 {
			return pdp.MakeIntegerValue(int64(v)), nil
		}

		return pdp.UndefinedValue, fmt.Errorf("%d overflows integer at (%d:%d)", v, i+1, j+1)

	case int8:
		return pdp.MakeIntegerValue(int64(v)), nil

	case int16:
		return pdp.MakeIntegerValue(int64(v)), nil

	case int32:
		return pdp.MakeIntegerValue(int64(v)), nil

	case int64:
		return pdp.MakeIntegerValue(v), nil

	case uint:
		if v <= math.MaxInt64 {
			return pdp.MakeIntegerValue(int64(v)), nil
		}

		return pdp.UndefinedValue, fmt.Errorf("%d overflows integer at (%d:%d)", v, i+1, j+1)

	case uint8:
		return pdp.MakeIntegerValue(int64(v)), nil

	case uint16:
		return pdp.MakeIntegerValue(int64(v)), nil

	case uint32:
		return pdp.MakeIntegerValue(int64(v)), nil

	case uint64:
		if v <= math.MaxInt64 {
			return pdp.MakeIntegerValue(int64(v)), nil
		}

		return pdp.UndefinedValue, fmt.Errorf("%d overflows integer at (%d:%d)", v, i+1, j+1)

	case string:
		out, err := pdp.MakeValueFromString(pdp.TypeInteger, v)
		if err != nil {
			return pdp.UndefinedValue, fmt.Errorf("can't make integer value from %q at (%d:%d)", v, i+1, j+1)
		}

		return out, nil
	}

	return pdp.UndefinedValue, fmt.Errorf("expected integer value but got %T at (%d:%d)", v, i+1, j+1)
}

func parseFloatValue(i, j int, v interface{}) (pdp.AttributeValue, error) {
	switch v := v.(type) {
	case float32:
		return pdp.MakeFloatValue(float64(v)), nil

	case float64:
		return pdp.MakeFloatValue(v), nil

	case string:
		out, err := pdp.MakeValueFromString(pdp.TypeFloat, v)
		if err != nil {
			return pdp.UndefinedValue, fmt.Errorf("can't make float value from %q at (%d:%d)", v, i+1, j+1)
		}

		return out, nil
	}

	return pdp.UndefinedValue, fmt.Errorf("expected float value but got %T at (%d:%d)", v, i+1, j+1)
}

func parseAddressValue(i, j int, v interface{}) (pdp.AttributeValue, error) {
	if s, ok := v.(string); ok {
		out, err := pdp.MakeValueFromString(pdp.TypeAddress, s)
		if err != nil {
			return pdp.UndefinedValue, fmt.Errorf("can't make address from %q at (%d:%d)", v, i+1, j+1)
		}

		return out, nil
	}

	return pdp.UndefinedValue, fmt.Errorf("expected address but got %T at (%d:%d)", v, i+1, j+1)
}

func parseNetworkValue(i, j int, v interface{}) (pdp.AttributeValue, error) {
	if s, ok := v.(string); ok {
		out, err := pdp.MakeValueFromString(pdp.TypeNetwork, s)
		if err != nil {
			return pdp.UndefinedValue, fmt.Errorf("can't make network from %q at (%d:%d)", v, i+1, j+1)
		}

		return out, nil
	}

	return pdp.UndefinedValue, fmt.Errorf("expected network but got %T at (%d:%d)", v, i+1, j+1)
}

func parseDomainValue(i, j int, v interface{}) (pdp.AttributeValue, error) {
	if s, ok := v.(string); ok {
		out, err := pdp.MakeValueFromString(pdp.TypeDomain, s)
		if err != nil {
			return pdp.UndefinedValue, fmt.Errorf("can't make domain name from %q at (%d:%d)", v, i+1, j+1)
		}

		return out, nil
	}

	return pdp.UndefinedValue, fmt.Errorf("expected domain name but got %T at (%d:%d)", v, i+1, j+1)
}

func parseSetOfStringsValue(i, j int, v interface{}) (pdp.AttributeValue, error) {
	if v, ok := v.([]interface{}); ok {
		t := strtree.NewTree()
		for k, v := range v {
			s, ok := v.(string)
			if !ok {
				return pdp.UndefinedValue,
					fmt.Errorf("expected string as set of strings item but got %T at (%d:%d:%d)",
						v, i+1, j+1, k+1)
			}

			t.InplaceInsert(s, k)
		}

		return pdp.MakeSetOfStringsValue(t), nil
	}

	return pdp.UndefinedValue, fmt.Errorf("expected set of strings but got %T at (%d:%d)", v, i+1, j+1)
}

func parseSetOfNetworksValue(i, j int, v interface{}) (pdp.AttributeValue, error) {
	if v, ok := v.([]interface{}); ok {
		t := iptree.NewTree()
		for k, v := range v {
			s, ok := v.(string)
			if !ok {
				return pdp.UndefinedValue,
					fmt.Errorf("expected address or network as set of networks item but got %T at (%d:%d:%d)",
						v, i+1, j+1, k+1)
			}

			if ip := net.ParseIP(s); ip != nil {
				t.InplaceInsertIP(ip, k)
			} else {
				_, n, err := net.ParseCIDR(s)
				if err != nil {
					return pdp.UndefinedValue,
						fmt.Errorf("expected address or network as set of networks item but got %q at (%d:%d:%d)",
							s, i+1, j+1, k+1)
				}

				t.InplaceInsertNet(n, k)
			}
		}

		return pdp.MakeSetOfNetworksValue(t), nil
	}

	return pdp.UndefinedValue, fmt.Errorf("expected set of networks but got %T at (%d:%d)", v, i+1, j+1)
}

func parseSetOfDomainsValue(i, j int, v interface{}) (pdp.AttributeValue, error) {
	if v, ok := v.([]interface{}); ok {
		t := new(domaintree.Node)
		for k, v := range v {
			s, ok := v.(string)
			if !ok {
				return pdp.UndefinedValue,
					fmt.Errorf("expected domain as set of domains item but got %T at (%d:%d:%d)",
						v, i+1, j+1, k+1)
			}

			d, err := domain.MakeNameFromString(s)
			if err != nil {
				return pdp.UndefinedValue,
					fmt.Errorf("expected domain as set of domains item but got %q at (%d:%d:%d)",
						s, i+1, j+1, k+1)
			}

			t.InplaceInsert(d, k)
		}

		return pdp.MakeSetOfDomainsValue(t), nil
	}

	return pdp.UndefinedValue, fmt.Errorf("expected set of domains but got %T at (%d:%d)", v, i+1, j+1)
}

func parseListOfStringsValue(i, j int, v interface{}) (pdp.AttributeValue, error) {
	if v, ok := v.([]interface{}); ok {
		ss := make([]string, 0, len(v))
		for k, v := range v {
			s, ok := v.(string)
			if !ok {
				return pdp.UndefinedValue,
					fmt.Errorf("expected string as list of strings item but got %T at (%d:%d:%d)",
						v, i+1, j+1, k+1)
			}

			ss = append(ss, s)
		}

		return pdp.MakeListOfStringsValue(ss), nil
	}

	return pdp.UndefinedValue, fmt.Errorf("expected list of strings but got %T at (%d:%d)", v, i+1, j+1)
}
