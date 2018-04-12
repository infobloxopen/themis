package pdp

import (
	"sort"
	"strconv"
	"strings"
)

// Type* values represent all built-in data types PDP can work with.
var (
	// TypeUndefined stands for type of undefined value. The value usually
	// means that evaluation can't be done.
	TypeUndefined = newBuiltinType("Undefined")
	// TypeBoolean is boolean data type.
	TypeBoolean = newBuiltinType("Boolean")
	// TypeString is string data type.
	TypeString = newBuiltinType("String")
	// TypeInteger is integer data type.
	TypeInteger = newBuiltinType("Integer")
	// TypeFloat is float data type.
	TypeFloat = newBuiltinType("Float")
	// TypeAddress is IPv4 or IPv6 address data type.
	TypeAddress = newBuiltinType("Address")
	// TypeNetwork is IPv4 or IPv6 network data type.
	TypeNetwork = newBuiltinType("Network")
	// TypeDomain is domain name data type.
	TypeDomain = newBuiltinType("Domain")
	// TypeSetOfStrings is set of strings data type (internally stores order
	// in which it was created).
	TypeSetOfStrings = newBuiltinType("Set of Strings")
	// TypeSetOfNetworks is set of networks data type (unordered).
	TypeSetOfNetworks = newBuiltinType("Set of Networks")
	// TypeSetOfDomains is set of domains data type (unordered).
	TypeSetOfDomains = newBuiltinType("Set of Domains")
	// TypeListOfStrings is list of strings data type.
	TypeListOfStrings = newBuiltinType("List of Strings")

	// BuiltinTypeIDs maps type keys to Type* constants.
	BuiltinTypes = make(map[string]Type)
)

// Type is generic data type.
type Type interface {
	// String returns human readable type name.
	String() string
	// GetKey returns case insensitive (always lowercase) type key.
	GetKey() string
}

type builtinType struct {
	n string
	k string
}

func newBuiltinType(s string) Type {
	t := &builtinType{
		n: s,
		k: strings.ToLower(s),
	}

	BuiltinTypes[t.GetKey()] = t

	return t
}

func (t *builtinType) String() string {
	return t.n
}

func (t *builtinType) GetKey() string {
	return t.k
}

// Signature is an ordered sequence of types.
type Signature []Type

// MakeSignature function creates signature from given types.
func MakeSignature(t ...Type) Signature {
	return t
}

// String method returns a string containing list of types separated by slash.
func (s Signature) String() string {
	if len(s) > 0 {
		seq := make([]string, len(s))
		for i, t := range s {
			seq[i] = strconv.Quote(t.String())
		}

		return strings.Join(seq, "/")
	}

	return "empty"
}

// TypeSet represent an unordered set of types.
type TypeSet map[Type]struct{}

func makeTypeSet(t ...Type) TypeSet {
	s := make(TypeSet, len(t))
	for _, t := range t {
		s[t] = struct{}{}
	}
	return s
}

// Contains method checks whether the set contains a type.
func (s TypeSet) Contains(t Type) bool {
	_, ok := s[t]
	return ok
}

// String method returns a string containing type names separated by comma.
func (s TypeSet) String() string {
	if len(s) > 0 {
		seq := make([]string, len(s))
		i := 0
		for t := range s {
			seq[i] = strconv.Quote(t.String())
			i++
		}

		sort.Strings(seq)
		return strings.Join(seq, ", ")
	}

	return "empty"
}
