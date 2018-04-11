package pdp

import "strings"

// Type* constants represent all data types PDP can work with.
const (
	// TypeUndefined stands for type of undefined value. The value usually
	// means that evaluation can't be done.
	TypeUndefined = iota
	// TypeBoolean is boolean data type.
	TypeBoolean
	// TypeString is string data type.
	TypeString
	// TypeInteger is integer data type.
	TypeInteger
	// TypeFloat is float data type.
	TypeFloat
	// TypeAddress is IPv4 or IPv6 address data type.
	TypeAddress
	// TypeNetwork is IPv4 or IPv6 network data type.
	TypeNetwork
	// TypeDomain is domain name data type.
	TypeDomain
	// TypeSetOfStrings is set of strings data type (internally stores order
	// in which it was created).
	TypeSetOfStrings
	// TypeSetOfNetworks is set of networks data type (unordered).
	TypeSetOfNetworks
	// TypeSetOfDomains is set of domains data type (unordered).
	TypeSetOfDomains
	// TypeListOfStrings is list of strings data type.
	TypeListOfStrings

	typesTotal
)

// BuiltinType* collections bind type names and IDs.
var (
	// BuiltinTypeNames is list of humanreadable type names. The order must
	// be kept in sync with Type* constants order.
	BuiltinTypeNames = []string{
		"Undefined",
		"Boolean",
		"String",
		"Integer",
		"Float",
		"Address",
		"Network",
		"Domain",
		"Set of Strings",
		"Set of Networks",
		"Set of Domains",
		"List of Strings"}

	// BuiltinTypeKeys maps Type* constants to type IDs. Type ID is all lower
	// case type name. The slice is filled by init function.
	BuiltinTypeKeys []string
	// BuiltinTypeIDs maps type IDs to Type* constants. The map is filled by
	// init function.
	BuiltinTypeIDs = map[string]int{}
)

func init() {
	BuiltinTypeKeys = make([]string, typesTotal)
	for t := 0; t < typesTotal; t++ {
		key := strings.ToLower(BuiltinTypeNames[t])
		BuiltinTypeKeys[t] = key
		BuiltinTypeIDs[key] = t
	}
}
