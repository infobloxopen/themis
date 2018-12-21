package pkg

import (
	"fmt"
	"strings"
)

const (
	pipTypeBoolean       = "boolean"
	pipTypeString        = "string"
	pipTypeInteger       = "integer"
	pipTypeFloat         = "float"
	pipTypeAddress       = "address"
	pipTypeNetwork       = "network"
	pipTypeDomain        = "domain"
	pipTypeSetOfStrings  = "set of strings"
	pipTypeSetOfNetworks = "set of networks"
	pipTypeSetOfDomains  = "set of domains"
	pipTypeListOfStrings = "list of strings"

	goTypeBool           = "bool"
	goTypeString         = "string"
	goTypeInt64          = "int64"
	goTypeFloat64        = "float64"
	goTypeNetIP          = "net.IP"
	goTypeNetIPNet       = "*net.IPNet"
	goTypeDomainName     = "domain.Name"
	goTypeStrtree        = "*strtree.Tree"
	goTypeIPTree         = "*iptree.Tree"
	goTypeDomainTree     = "*domaintree.Node"
	goTypeSliceOfStrings = "[]string"
)

type goType struct {
	name string
	zero string
}

var (
	typeMap = map[string]goType{
		pipTypeBoolean: {
			name: goTypeBool,
			zero: "false",
		},
		pipTypeString: {
			name: goTypeString,
			zero: "\"\"",
		},
		pipTypeInteger: {
			name: goTypeInt64,
			zero: "0",
		},
		pipTypeFloat: {
			name: goTypeFloat64,
			zero: "0",
		},
		pipTypeAddress: {
			name: goTypeNetIP,
			zero: "nil",
		},
		pipTypeNetwork: {
			name: goTypeNetIPNet,
			zero: "nil",
		},
		pipTypeDomain: {
			name: goTypeDomainName,
			zero: "domain.Name{}",
		},
		pipTypeSetOfStrings: {
			name: goTypeStrtree,
			zero: "nil",
		},
		pipTypeSetOfNetworks: {
			name: goTypeIPTree,
			zero: "nil",
		},
		pipTypeSetOfDomains: {
			name: goTypeDomainTree,
			zero: "nil",
		},
		pipTypeListOfStrings: {
			name: goTypeSliceOfStrings,
			zero: "nil",
		},
	}
)

func makeGoTypeList(args []string) ([]string, error) {
	out := make([]string, 0, len(args))
	for i, arg := range args {
		t, ok := typeMap[strings.ToLower(arg)]
		if !ok {
			return nil, fmt.Errorf("argument %d: unknown type %q", i, arg)
		}

		out = append(out, t.name)
	}

	return out, nil
}

func getGoType(arg string) (string, string, error) {
	t, ok := typeMap[strings.ToLower(arg)]
	if !ok {
		return "", "", fmt.Errorf("result: unknown type %q", arg)
	}

	return t.name, t.zero, nil
}
