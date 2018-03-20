package pdp

import (
	"fmt"
	"strings"
)

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

// Type* collections bind type names and IDs.
var (
	// TypeNames is list of humanreadable type names. The order must be kept
	// in sync with Type* constants order.
	TypeNames = []string{
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

	// TypeKeys maps Type* constants to type IDs. Type ID is all lower case
	// type name. The slice is filled by init function.
	TypeKeys []string
	// TypeIDs maps type IDs to Type* constants. The map is filled by init
	// function.
	TypeIDs = map[string]int{}
)

var undefinedValue = AttributeValue{}

func init() {
	TypeKeys = make([]string, typesTotal)
	for t := 0; t < typesTotal; t++ {
		key := strings.ToLower(TypeNames[t])
		TypeKeys[t] = key
		TypeIDs[key] = t
	}
}

// Attribute represents attribute definition which binds attribute name
// and type.
type Attribute struct {
	id string
	t  int
}

// MakeAttribute creates new attribute instance. It requires attribute name
// as "ID" argument and type as "t" argument. Value of "t" should be one of
// Type* constants.
func MakeAttribute(ID string, t int) Attribute {
	return Attribute{id: ID, t: t}
}

// GetType returns attribute type.
func (a Attribute) GetType() int {
	return a.t
}

func (a Attribute) describe() string {
	return fmt.Sprintf("attr(%s.%s)", a.id, TypeNames[a.t])
}
