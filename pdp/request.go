package pdp

import (
	"encoding/binary"
	"fmt"
	"math"
	"net"
	"reflect"

	"github.com/infobloxopen/go-trees/domain"
	"github.com/infobloxopen/go-trees/domaintree"
	"github.com/infobloxopen/go-trees/iptree"
	"github.com/infobloxopen/go-trees/strtree"
)

const requestVersion = uint16(1)

const (
	requestWireTypeBooleanFalse = iota
	requestWireTypeBooleanTrue
	requestWireTypeString
	requestWireTypeInteger
	requestWireTypeFloat
	requestWireTypeIPv4Address
	requestWireTypeIPv6Address
	requestWireTypeIPv4Network
	requestWireTypeIPv6Network
	requestWireTypeDomain
	requestWireTypeSetOfStrings
	requestWireTypeSetOfNetworks
	requestWireTypeSetOfDomains
	requestWireTypeListOfStrings
	requestWireTypeSetOfFlags

	requestWireTypesTotal
)

var (
	requestWireTypeNames = []string{
		"boolean false",
		"boolean true",
		"string",
		"integer",
		"float",
		"IPv4 address",
		"IPv6 address",
		"IPv4 network",
		"IPv6 network",
		"domain",
		"set of strings",
		"set of networks",
		"set of domains",
		"list of strings",
		"set of flags",
	}

	builtinTypeByWire = []Type{
		TypeBoolean,
		TypeBoolean,
		TypeString,
		TypeInteger,
		TypeFloat,
		TypeAddress,
		TypeAddress,
		TypeNetwork,
		TypeNetwork,
		TypeDomain,
		TypeSetOfStrings,
		TypeSetOfNetworks,
		TypeSetOfDomains,
		TypeListOfStrings,
	}
)

const (
	reqVersionSize = 2

	reqSmallCounterSize = 1
	reqBigCounterSize   = 2

	reqTypeSize = 1

	reqBooleanValueSize     = 0
	reqIntegerValueSize     = 8
	reqFloatValueSize       = 8
	reqIPv4AddressValueSize = 4
	reqIPv6AddressValueSize = 16
	reqNetworkCIDRSize      = 1
)

// MarshalRequestAssignments marshals list of assignments to sequence of bytes.
// It requires each assignment to have immediate value as an expression (which
// can be created with MakeStringValue or similar functions).
func MarshalRequestAssignments(in []AttributeAssignment) ([]byte, error) {
	n, err := calcRequestSize(in)
	if err != nil {
		return nil, err
	}

	b := make([]byte, n)
	_, err = MarshalRequestAssignmentsToBuffer(b, in)
	return b, err
}

// MarshalRequestAssignmentsWithAllocator marshals list of assignments
// to sequence of bytes in the same way as MarshalRequestAssignments. But
// instead of make function it uses given allocator function to obtain buffer.
// The allocator expected to take number of bytes and return slice of bytes
// with given length.
func MarshalRequestAssignmentsWithAllocator(in []AttributeAssignment, f func(n int) ([]byte, error)) ([]byte, error) {
	n, err := calcRequestSize(in)
	if err != nil {
		return nil, err
	}

	b, err := f(n)
	if err != nil {
		return nil, err
	}

	n, err = MarshalRequestAssignmentsToBuffer(b, in)
	if err != nil {
		return nil, err
	}

	return b[:n], nil
}

// MarshalRequestAssignmentsToBuffer marshals list of assignments as a sequence of
// bytes to given buffer. Caller should provide large enough buffer. Function fills
// the buffer and returns number of bytes written.
func MarshalRequestAssignmentsToBuffer(b []byte, in []AttributeAssignment) (int, error) {
	off, err := putRequestVersion(b)
	if err != nil {
		return off, err
	}

	n, err := putAssignmentExpressions(b[off:], in)
	if err != nil {
		return off, err
	}

	return off + n, nil
}

// MarshalRequestReflection marshals set of attributes wrapped with
// reflect.Value to sequence of bytes. For each attribute
// MarshalRequestReflection calls f function with index of the attribute.
// It expects the function to return attribute id, type and value.
// For TypeBoolean MarshalRequestReflectionToBuffer expects bool value,
// for TypeString - string, for TypeInteger - intX, uintX (internally converting
// to int64), TypeFloat - float32 or float64, TypeAddress - net.IP, TypeNetwork
// - net.IPNet or *net.IPNet, TypeDomain - string or domain.Name from
// github.com/infobloxopen/go-trees/domain package, TypeSetOfStrings -
// *strtree.Tree from github.com/infobloxopen/go-trees/strtree package,
// TypeSetOfNetworks - *iptree.Node from
// github.com/infobloxopen/go-trees/iptree, TypeSetOfDomains - *domaintree.Node
// from github.com/infobloxopen/go-trees/domaintree, TypeListOfStrings -
// []string.
func MarshalRequestReflection(c int, f func(i int) (string, Type, reflect.Value, error)) ([]byte, error) {
	n, err := calcRequestSizeFromReflection(c, f)
	if err != nil {
		return nil, err
	}

	b := make([]byte, n)
	_, err = MarshalRequestReflectionToBuffer(b, c, f)
	return b, err
}

// MarshalRequestReflectionWithAllocator marshals set of attributes wrapped with
// reflect.Value to sequence of bytes in the same way
// as MarshalRequestReflection. But instead of make function it uses given
// allocator function to obtain buffer. The allocator expected to take number of
// bytes and return slice of bytes with given length.
func MarshalRequestReflectionWithAllocator(c int, f func(i int) (string, Type, reflect.Value, error), g func(n int) ([]byte, error)) ([]byte, error) {
	n, err := calcRequestSizeFromReflection(c, f)
	if err != nil {
		return nil, err
	}

	b, err := g(n)
	if err != nil {
		return nil, err
	}

	n, err = MarshalRequestReflectionToBuffer(b, c, f)
	if err != nil {
		return nil, err
	}

	return b[:n], nil
}

// MarshalRequestReflectionToBuffer marshals set of attributes wrapped with
// reflect.Value as a sequence of bytes to given buffer similarly to
// MarshalRequestReflection. Caller should provide large enough buffer.
// The function fills given buffer and returns number of bytes written.
func MarshalRequestReflectionToBuffer(b []byte, c int, f func(i int) (string, Type, reflect.Value, error)) (int, error) {
	off, err := putRequestVersion(b)
	if err != nil {
		return off, err
	}

	n, err := putAttributesFromReflection(b[off:], c, f)
	if err != nil {
		return off, err
	}

	return off + n, nil
}

// MarshalInfoRequest marshals request for additional information as a sequence
// of bytes to given buffer. The information request is used to get data from
// PIP and consists of a path and a set of attribute values. The path  is used
// to identify specific data source within the same PIP server. Caller should
// provide large enough buffer. The function fills given buffer and returns
// number of bytes written.
func MarshalInfoRequest(b []byte, path string, in []AttributeValue) (int, error) {
	off, err := putRequestVersion(b)
	if err != nil {
		return off, err
	}

	n, err := putRequestString(b[off:], path)
	if err != nil {
		return off, err
	}
	off += n

	n, err = putRequestAttributeCount(b[off:], len(in))
	if err != nil {
		return off, err
	}
	off += n

	for _, v := range in {
		n, err = putRequestAttributeValue(b[off:], v)
		if err != nil {
			return off, err
		}
		off += n
	}

	return off, nil
}

// UnmarshalRequestAssignments parses given sequence of bytes as
// a list of assignments.
func UnmarshalRequestAssignments(b []byte) ([]AttributeAssignment, error) {
	n, err := checkRequestVersion(b)
	if err != nil {
		return nil, err
	}

	return getAssignmentExpressions(b[n:])
}

// UnmarshalRequestAssignmentsWithAllocator parses given sequence of bytes as
// a list of assignments. It uses given allocator to make assignments array.
// The allocator expected to take a number of assignments required and return
// a slice of at least given length.
func UnmarshalRequestAssignmentsWithAllocator(b []byte, f func(n int) ([]AttributeAssignment, error)) ([]AttributeAssignment, error) {
	n, err := checkRequestVersion(b)
	if err != nil {
		return nil, err
	}

	return getAssignmentExpressionsWithAllocator(b[n:], f)
}

// UnmarshalRequestToAssignmentsArray parses given sequence of bytes as
// a list of assignments to given buffer. Caller should provide large enough
// out slice. The function returns number of assignments written.
func UnmarshalRequestToAssignmentsArray(b []byte, out []AttributeAssignment) (int, error) {
	n, err := checkRequestVersion(b)
	if err != nil {
		return 0, err
	}

	return getAssignmentExpressionsToArray(b[n:], out)
}

// UnmarshalRequestReflection parses given sequence of bytes to set of reflected
// values. It calls f function for each attribute extracted from buffer with
// attribute id and type. The f function should return value to set.
// If it returns error UnmarshalRequestReflection stops parsing and exits with
// the error.
func UnmarshalRequestReflection(b []byte, f func(string, Type) (reflect.Value, error)) error {
	n, err := checkRequestVersion(b)
	if err != nil {
		return err
	}
	b = b[n:]

	return getAttributesToReflection(b, f)
}

// UnmarshalInfoRequest unmarshals information request from given buffer.
// It fills given assignment array and returns path and number of attributes.
// Caller should provide large enough array for assignments.
func UnmarshalInfoRequest(b []byte, out []AttributeValue) (string, int, error) {
	n, err := checkRequestVersion(b)
	if err != nil {
		return "", 0, err
	}
	b = b[n:]

	s, n, err := getRequestStringValue(b)
	if err != nil {
		return "", 0, err
	}
	b = b[n:]

	c, n, err := getRequestAttributeCount(b)
	if err != nil {
		return "", 0, err
	}
	b = b[n:]

	if c > len(out) {
		return "", 0, newRequestValuesOverflowError(c, len(out))
	}

	for i := 0; i < c; i++ {
		v, n, err := getRequestAttributeValue(b)
		if err != nil {
			return "", 0, err
		}
		b = b[n:]

		out[i] = v
	}

	return s, c, nil
}

func putRequestVersion(b []byte) (int, error) {
	if len(b) < reqVersionSize {
		return 0, newRequestBufferOverflowError()
	}

	binary.LittleEndian.PutUint16(b, requestVersion)
	return reqVersionSize, nil
}

func checkRequestVersion(b []byte) (int, error) {
	if len(b) < reqVersionSize {
		return 0, newRequestBufferUnderflowError()
	}

	if v := binary.LittleEndian.Uint16(b); v != requestVersion {
		return 0, newRequestVersionError(v, requestVersion)
	}

	return reqVersionSize, nil
}

func putRequestAttributeCount(b []byte, n int) (int, error) {
	if n < 0 {
		return 0, newRequestInvalidAttributeCountError(n)
	}

	if n > math.MaxUint16 {
		return 0, newRequestTooManyAttributesError(n)
	}

	if len(b) < reqBigCounterSize {
		return 0, newRequestBufferOverflowError()
	}

	binary.LittleEndian.PutUint16(b, uint16(n))
	return reqBigCounterSize, nil
}

func getRequestAttributeCount(b []byte) (int, int, error) {
	if len(b) < reqBigCounterSize {
		return 0, 0, newRequestBufferUnderflowError()
	}

	return int(binary.LittleEndian.Uint16(b)), reqBigCounterSize, nil
}

// CheckInfoRequestHeader validates if request for additional information
// has correct header - current version and required number of values.
func CheckInfoRequestHeader(b []byte, count uint16) ([]byte, error) {
	if len(b) < reqVersionSize+reqBigCounterSize {
		return nil, newRequestBufferUnderflowError()
	}

	if v := binary.LittleEndian.Uint16(b); v != requestVersion {
		return nil, newRequestVersionError(v, requestVersion)
	}
	b = b[reqVersionSize:]

	if c := binary.LittleEndian.Uint16(b); c != count {
		return nil, newRequestAttributeCountError(c, count)
	}
	b = b[reqBigCounterSize:]

	return b, nil
}

func putRequestAttribute(b []byte, name string, value AttributeValue) (int, error) {
	t := value.GetResultType()

	switch t {
	case TypeBoolean:
		v, _ := value.boolean()
		return putRequestAttributeBoolean(b, name, v)

	case TypeString:
		v, _ := value.str()
		return putRequestAttributeString(b, name, v)

	case TypeInteger:
		v, _ := value.integer()
		return putRequestAttributeInteger(b, name, v)

	case TypeFloat:
		v, _ := value.float()
		return putRequestAttributeFloat(b, name, v)

	case TypeAddress:
		v, _ := value.address()
		return putRequestAttributeAddress(b, name, v)

	case TypeNetwork:
		v, _ := value.network()
		return putRequestAttributeNetwork(b, name, v)

	case TypeDomain:
		v, _ := value.domain()
		return putRequestAttributeDomain(b, name, v)

	case TypeSetOfStrings:
		v, _ := value.setOfStrings()
		return putRequestAttributeSetOfStrings(b, name, v)

	case TypeSetOfNetworks:
		v, _ := value.setOfNetworks()
		return putRequestAttributeSetOfNetworks(b, name, v)

	case TypeSetOfDomains:
		v, _ := value.setOfDomains()
		return putRequestAttributeSetOfDomains(b, name, v)

	case TypeListOfStrings:
		v, _ := value.listOfStrings()
		return putRequestAttributeListOfStrings(b, name, v)
	}

	return 0, newRequestAttributeMarshallingNotImplementedError(t)
}

func putRequestAttributeValue(b []byte, value AttributeValue) (int, error) {
	t := value.GetResultType()
	if t, ok := t.(*FlagsType); ok {
		switch t.c {
		case 8:
			v, _ := value.flags8()
			return putRequestSetOfFlags8Value(b, v, t)

		case 16:
			v, _ := value.flags16()
			return putRequestSetOfFlags16Value(b, v, t)

		case 32:
			v, _ := value.flags32()
			return putRequestSetOfFlags32Value(b, v, t)
		}

		v, _ := value.flags64()
		return putRequestSetOfFlags64Value(b, v, t)
	}

	switch t {
	case TypeBoolean:
		v, _ := value.boolean()
		return putRequestBooleanValue(b, v)

	case TypeString:
		v, _ := value.str()
		return putRequestStringValue(b, v)

	case TypeInteger:
		v, _ := value.integer()
		return putRequestIntegerValue(b, v)

	case TypeFloat:
		v, _ := value.float()
		return putRequestFloatValue(b, v)

	case TypeAddress:
		v, _ := value.address()
		return putRequestAddressValue(b, v)

	case TypeNetwork:
		v, _ := value.network()
		return putRequestNetworkValue(b, v)

	case TypeDomain:
		v, _ := value.domain()
		return putRequestDomainValue(b, v)

	case TypeSetOfStrings:
		v, _ := value.setOfStrings()
		return putRequestSetOfStringsValue(b, v)

	case TypeSetOfNetworks:
		v, _ := value.setOfNetworks()
		return putRequestSetOfNetworksValue(b, v)

	case TypeSetOfDomains:
		v, _ := value.setOfDomains()
		return putRequestSetOfDomainsValue(b, v)

	case TypeListOfStrings:
		v, _ := value.listOfStrings()
		return putRequestListOfStringsValue(b, v)
	}

	return 0, newRequestAttributeMarshallingNotImplementedError(t)
}

func getRequestAttribute(b []byte) (string, AttributeValue, int, error) {
	name, off, err := getRequestAttributeName(b)
	if err != nil {
		return "", UndefinedValue, 0, bindError(err, "name")
	}

	v, n, err := getRequestAttributeValue(b[off:])
	if err != nil {
		return "", UndefinedValue, 0, bindError(err, name)
	}

	return name, v, off + n, nil
}

func getRequestAttributeValue(b []byte) (AttributeValue, int, error) {
	t, off, err := getRequestAttributeType(b)
	if err != nil {
		return UndefinedValue, 0, bindError(err, "type")
	}

	v, n, err := getRequestAttributeValueWithType(t, b[off:])
	if err != nil {
		return UndefinedValue, 0, bindError(err, "value")
	}

	return v, off + n, nil
}

func getRequestAttributeValueWithType(t int, b []byte) (AttributeValue, int, error) {
	switch t {
	case requestWireTypeBooleanFalse:
		return MakeBooleanValue(false), 0, nil

	case requestWireTypeBooleanTrue:
		return MakeBooleanValue(true), 0, nil

	case requestWireTypeString:
		s, n, err := getRequestStringValue(b)
		if err != nil {
			return UndefinedValue, 0, err
		}

		return MakeStringValue(s), n, nil

	case requestWireTypeInteger:
		i, n, err := getRequestIntegerValue(b)
		if err != nil {
			return UndefinedValue, 0, err
		}

		return MakeIntegerValue(i), n, nil

	case requestWireTypeFloat:
		f, n, err := getRequestFloatValue(b)
		if err != nil {
			return UndefinedValue, 0, err
		}

		return MakeFloatValue(f), n, nil

	case requestWireTypeIPv4Address:
		a, n, err := getRequestIPv4AddressValue(b)
		if err != nil {
			return UndefinedValue, 0, err
		}

		return MakeAddressValue(a), n, nil

	case requestWireTypeIPv6Address:
		a, n, err := getRequestIPv6AddressValue(b)
		if err != nil {
			return UndefinedValue, 0, err
		}

		return MakeAddressValue(a), n, nil

	case requestWireTypeIPv4Network:
		a, n, err := getRequestIPv4NetworkValue(b)
		if err != nil {
			return UndefinedValue, 0, err
		}

		return MakeNetworkValue(a), n, nil

	case requestWireTypeIPv6Network:
		a, n, err := getRequestIPv6NetworkValue(b)
		if err != nil {
			return UndefinedValue, 0, err
		}

		return MakeNetworkValue(a), n, nil

	case requestWireTypeDomain:
		d, n, err := getRequestDomainValue(b)
		if err != nil {
			return UndefinedValue, 0, err
		}

		return MakeDomainValue(d), n, nil

	case requestWireTypeSetOfStrings:
		ss, n, err := getRequestSetOfStringsValue(b)
		if err != nil {
			return UndefinedValue, 0, err
		}

		return MakeSetOfStringsValue(ss), n, nil

	case requestWireTypeSetOfNetworks:
		sn, n, err := getRequestSetOfNetworksValue(b)
		if err != nil {
			return UndefinedValue, 0, err
		}

		return MakeSetOfNetworksValue(sn), n, nil

	case requestWireTypeSetOfDomains:
		sd, n, err := getRequestSetOfDomainsValue(b)
		if err != nil {
			return UndefinedValue, 0, err
		}

		return MakeSetOfDomainsValue(sd), n, nil

	case requestWireTypeListOfStrings:
		ls, n, err := getRequestListOfStringsValue(b)
		if err != nil {
			return UndefinedValue, 0, err
		}

		return MakeListOfStringsValue(ls), n, nil

	case requestWireTypeSetOfFlags:
		v, s, n, err := getRequestAbstractSetOfFlagsValue(b)
		if err != nil {
			return UndefinedValue, 0, err
		}

		if s <= 0 || s > len(abstractFlagTypes)+1 {
			return UndefinedValue, 0, newRequestAttributeUnmarshallingFlagsSizeError(s)
		}

		ft := abstractFlagTypes[s-1]

		switch {
		case s <= 8:
			return MakeFlagsValue8(uint8(v), ft), n, nil

		case s <= 16:
			return MakeFlagsValue16(uint16(v), ft), n, nil

		case s <= 32:
			return MakeFlagsValue32(uint32(v), ft), n, nil
		}

		return MakeFlagsValue64(v, ft), n, nil
	}

	return UndefinedValue, 0, newRequestAttributeUnmarshallingTypeError(t)
}

func putRequestAttributeName(b []byte, name string) (int, error) {
	if len(name) > math.MaxUint8 {
		return 0, newRequestTooLongAttributeNameError(name)
	}

	n := len(name) + reqSmallCounterSize
	if len(b) < n {
		return 0, newRequestBufferOverflowError()
	}

	b[0] = byte(len(name))
	copy(b[reqSmallCounterSize:], []byte(name))

	return n, nil
}

func getRequestAttributeName(b []byte) (string, int, error) {
	if len(b) < 1 {
		return "", 0, newRequestBufferUnderflowError()
	}

	off := int(b[0]) + 1
	if len(b) < off {
		return "", 0, newRequestBufferUnderflowError()
	}

	return string(b[1:off]), off, nil
}

func putRequestAttributeType(b []byte, t int) (int, error) {
	if t < 0 || t >= requestWireTypesTotal {
		return 0, newRequestAttributeMarshallingTypeError(t)
	}

	if len(b) < reqTypeSize {
		return 0, newRequestBufferOverflowError()
	}

	b[0] = byte(t)
	return reqTypeSize, nil
}

func getRequestAttributeType(b []byte) (int, int, error) {
	if len(b) < reqTypeSize {
		return 0, 0, newRequestBufferUnderflowError()
	}

	return int(b[0]), reqTypeSize, nil
}

func putRequestAttributeBoolean(b []byte, name string, value bool) (int, error) {
	off, err := putRequestAttributeName(b, name)
	if err != nil {
		return 0, err
	}

	n, err := putRequestBooleanValue(b[off:], value)
	if err != nil {
		return 0, err
	}

	return off + n, err
}

func putRequestBooleanValue(b []byte, value bool) (int, error) {
	if value {
		return putRequestAttributeType(b, requestWireTypeBooleanTrue)
	}

	return putRequestAttributeType(b, requestWireTypeBooleanFalse)
}

// GetInfoRequestBooleanValue extracts boolean value from request for additional
// information.
func GetInfoRequestBooleanValue(b []byte) (bool, []byte, error) {
	if len(b) < reqTypeSize {
		return false, nil, newRequestBufferUnderflowError()
	}

	t := int(b[0])
	switch t {
	case requestWireTypeBooleanFalse:
		return false, b[reqTypeSize:], nil

	case requestWireTypeBooleanTrue:
		return true, b[reqTypeSize:], nil
	}

	return false, nil, newRequestAttributeUnmarshallingBooleanTypeError(t)
}

func putRequestAttributeString(b []byte, name string, value string) (int, error) {
	off, err := putRequestAttributeName(b, name)
	if err != nil {
		return 0, err
	}

	n, err := putRequestStringValue(b[off:], value)
	if err != nil {
		return 0, err
	}

	return off + n, err
}

func putRequestStringValue(b []byte, value string) (int, error) {
	off, err := putRequestAttributeType(b, requestWireTypeString)
	if err != nil {
		return 0, err
	}

	b = b[off:]

	n, err := putRequestString(b, value)
	if err != nil {
		return 0, err
	}

	return off + n, nil
}

func putRequestString(b []byte, value string) (int, error) {
	if len(value) > math.MaxUint16 {
		return 0, newRequestTooLongStringValueError(value)
	}

	n := len(value) + reqBigCounterSize
	if len(b) < n {
		return 0, newRequestBufferOverflowError()
	}

	binary.LittleEndian.PutUint16(b, uint16(len(value)))
	copy(b[reqBigCounterSize:], value)

	return n, nil
}

func getRequestStringValue(b []byte) (string, int, error) {
	if len(b) < reqBigCounterSize {
		return "", 0, newRequestBufferUnderflowError()
	}

	off := int(binary.LittleEndian.Uint16(b)) + reqBigCounterSize
	if len(b) < off {
		return "", 0, newRequestBufferUnderflowError()
	}

	return string(b[reqBigCounterSize:off]), off, nil
}

// GetInfoRequestStringValue extracts string from request for additional
// information.
func GetInfoRequestStringValue(b []byte) (string, []byte, error) {
	if len(b) < reqTypeSize {
		return "", nil, newRequestBufferUnderflowError()
	}

	if t := int(b[0]); t != requestWireTypeString {
		return "", nil, newRequestAttributeUnmarshallingStringTypeError(t)
	}
	b = b[reqTypeSize:]

	v, n, err := getRequestStringValue(b)
	if err != nil {
		return "", nil, err
	}

	return v, b[n:], nil
}

func putRequestAttributeInteger(b []byte, name string, value int64) (int, error) {
	off, err := putRequestAttributeName(b, name)
	if err != nil {
		return 0, err
	}

	n, err := putRequestIntegerValue(b[off:], value)
	if err != nil {
		return 0, err
	}

	return off + n, err
}

func putRequestIntegerValue(b []byte, value int64) (int, error) {
	off, err := putRequestAttributeType(b, requestWireTypeInteger)
	if err != nil {
		return 0, err
	}

	b = b[off:]

	if len(b) < reqIntegerValueSize {
		return 0, newRequestBufferOverflowError()
	}

	binary.LittleEndian.PutUint64(b, uint64(value))
	return off + reqIntegerValueSize, nil
}

func getRequestIntegerValue(b []byte) (int64, int, error) {
	if len(b) < reqIntegerValueSize {
		return 0, 0, newRequestBufferUnderflowError()
	}

	return int64(binary.LittleEndian.Uint64(b)), reqIntegerValueSize, nil
}

// GetInfoRequestIntegerValue extracts integer value from request for additional
// information.
func GetInfoRequestIntegerValue(b []byte) (int64, []byte, error) {
	if len(b) < reqTypeSize {
		return 0, nil, newRequestBufferUnderflowError()
	}

	if t := int(b[0]); t != requestWireTypeInteger {
		return 0, nil, newRequestAttributeUnmarshallingIntegerTypeError(t)
	}
	b = b[reqTypeSize:]

	v, n, err := getRequestIntegerValue(b)
	if err != nil {
		return 0, nil, err
	}

	return v, b[n:], nil
}

func putRequestAttributeFloat(b []byte, name string, value float64) (int, error) {
	off, err := putRequestAttributeName(b, name)
	if err != nil {
		return 0, err
	}

	n, err := putRequestFloatValue(b[off:], value)
	if err != nil {
		return 0, err
	}

	return off + n, err
}

func putRequestFloatValue(b []byte, value float64) (int, error) {
	off, err := putRequestAttributeType(b, requestWireTypeFloat)
	if err != nil {
		return 0, err
	}

	b = b[off:]

	if len(b) < 8 {
		return 0, newRequestBufferOverflowError()
	}

	binary.LittleEndian.PutUint64(b, math.Float64bits(value))
	return off + 8, nil
}

func getRequestFloatValue(b []byte) (float64, int, error) {
	if len(b) < 8 {
		return 0, 0, newRequestBufferUnderflowError()
	}

	return math.Float64frombits(binary.LittleEndian.Uint64(b)), 8, nil
}

// GetInfoRequestFloatValue extracts floating point value from request for
// additional information.
func GetInfoRequestFloatValue(b []byte) (float64, []byte, error) {
	if len(b) < reqTypeSize {
		return 0, nil, newRequestBufferUnderflowError()
	}

	if t := int(b[0]); t != requestWireTypeFloat {
		return 0, nil, newRequestAttributeUnmarshallingFloatTypeError(t)
	}
	b = b[reqTypeSize:]

	v, n, err := getRequestFloatValue(b)
	if err != nil {
		return 0, nil, err
	}

	return v, b[n:], nil
}

func putRequestAttributeAddress(b []byte, name string, value net.IP) (int, error) {
	off, err := putRequestAttributeName(b, name)
	if err != nil {
		return 0, err
	}

	n, err := putRequestAddressValue(b[off:], value)
	if err != nil {
		return 0, err
	}

	return off + n, err
}

func putRequestAddressValue(b []byte, value net.IP) (int, error) {
	t := requestWireTypeIPv4Address

	ip := value.To4()
	if ip == nil {
		t = requestWireTypeIPv6Address

		ip = value.To16()
		if ip == nil {
			return 0, newRequestAddressValueError(value)
		}
	}

	off, err := putRequestAttributeType(b, t)
	if err != nil {
		return 0, err
	}

	b = b[off:]

	if len(b) < len(ip) {
		return 0, newRequestBufferOverflowError()
	}

	copy(b, ip)
	return off + len(ip), nil
}

func getRequestIPv4AddressValue(b []byte) (net.IP, int, error) {
	if len(b) < reqIPv4AddressValueSize {
		return nil, 0, newRequestBufferUnderflowError()
	}

	return net.IPv4(b[0], b[1], b[2], b[3]), reqIPv4AddressValueSize, nil
}

func getRequestIPv6AddressValue(b []byte) (net.IP, int, error) {
	if len(b) < reqIPv6AddressValueSize {
		return nil, 0, newRequestBufferUnderflowError()
	}

	ip := net.IP(make([]byte, reqIPv6AddressValueSize))
	copy(ip, b)
	return ip, reqIPv6AddressValueSize, nil
}

// GetInfoRequestAddressValue extracts IP address from request for additional
// information.
func GetInfoRequestAddressValue(b []byte) (net.IP, []byte, error) {
	if len(b) < reqTypeSize {
		return nil, nil, newRequestBufferUnderflowError()
	}

	t := int(b[0])
	b = b[reqTypeSize:]

	var (
		v   net.IP
		n   int
		err error
	)

	switch t {
	default:
		return nil, nil, newRequestAttributeUnmarshallingAddressTypeError(t)

	case requestWireTypeIPv4Address:
		v, n, err = getRequestIPv4AddressValue(b)

	case requestWireTypeIPv6Address:
		v, n, err = getRequestIPv6AddressValue(b)
	}

	if err != nil {
		return nil, nil, err
	}

	return v, b[n:], nil
}

func putRequestAttributeNetwork(b []byte, name string, value *net.IPNet) (int, error) {
	off, err := putRequestAttributeName(b, name)
	if err != nil {
		return 0, err
	}

	n, err := putRequestNetworkValue(b[off:], value)
	if err != nil {
		return 0, err
	}

	return off + n, err
}

func putRequestNetworkValue(b []byte, value *net.IPNet) (int, error) {
	if value == nil {
		return 0, newRequestInvalidNetworkValueError(value)
	}

	ip := value.IP
	if len(ip) != 4 && len(ip) != 16 {
		return 0, newRequestInvalidNetworkValueError(value)
	}

	t := requestWireTypeIPv4Network
	ones, bits := value.Mask.Size()
	if bits != 32 {
		t = requestWireTypeIPv6Network
		if bits != 128 {
			return 0, newRequestInvalidNetworkValueError(value)
		}
	}

	off, err := putRequestAttributeType(b, t)
	if err != nil {
		return 0, err
	}

	b = b[off:]

	if len(b) < len(ip)+reqNetworkCIDRSize {
		return 0, newRequestBufferOverflowError()
	}

	b[0] = byte(ones)

	copy(b[reqNetworkCIDRSize:], ip)
	return off + len(ip) + reqNetworkCIDRSize, nil
}

func getRequestIPv4NetworkValue(b []byte) (*net.IPNet, int, error) {
	if len(b) < reqNetworkCIDRSize+reqIPv4AddressValueSize {
		return nil, 0, newRequestBufferUnderflowError()
	}

	mask := net.CIDRMask(int(b[0]), 32)
	if mask == nil {
		return nil, 0, newRequestIPv4InvalidMaskError(b[0])
	}

	return &net.IPNet{
		IP: net.IPv4(
			b[reqNetworkCIDRSize],
			b[reqNetworkCIDRSize+1],
			b[reqNetworkCIDRSize+2],
			b[reqNetworkCIDRSize+3],
		).Mask(mask),
		Mask: mask,
	}, reqNetworkCIDRSize + reqIPv4AddressValueSize, nil
}

func getRequestIPv6NetworkValue(b []byte) (*net.IPNet, int, error) {
	if len(b) < reqNetworkCIDRSize+reqIPv6AddressValueSize {
		return nil, 0, newRequestBufferUnderflowError()
	}

	mask := net.CIDRMask(int(b[0]), 128)
	if mask == nil {
		return nil, 0, newRequestIPv6InvalidMaskError(b[0])
	}

	ip := net.IP(make([]byte, 16))
	copy(ip, b[reqNetworkCIDRSize:])

	return &net.IPNet{
		IP:   ip.Mask(mask),
		Mask: mask,
	}, reqNetworkCIDRSize + reqIPv6AddressValueSize, nil
}

// GetInfoRequestNetworkValue extracts IP network from request for additional
// information.
func GetInfoRequestNetworkValue(b []byte) (*net.IPNet, []byte, error) {
	if len(b) < reqTypeSize {
		return nil, nil, newRequestBufferUnderflowError()
	}

	t := int(b[0])
	b = b[reqTypeSize:]

	var (
		v   *net.IPNet
		n   int
		err error
	)

	switch t {
	default:
		return nil, nil, newRequestAttributeUnmarshallingNetworkTypeError(t)

	case requestWireTypeIPv4Network:
		v, n, err = getRequestIPv4NetworkValue(b)

	case requestWireTypeIPv6Network:
		v, n, err = getRequestIPv6NetworkValue(b)
	}

	if err != nil {
		return nil, nil, err
	}

	return v, b[n:], nil
}

func putRequestAttributeDomain(b []byte, name string, value domain.Name) (int, error) {
	off, err := putRequestAttributeName(b, name)
	if err != nil {
		return 0, err
	}

	n, err := putRequestDomainValue(b[off:], value)
	if err != nil {
		return 0, err
	}

	return off + n, err
}

func putRequestDomainValue(b []byte, value domain.Name) (int, error) {
	s := value.String()

	off, err := putRequestAttributeType(b, requestWireTypeDomain)
	if err != nil {
		return 0, err
	}

	b = b[off:]

	n := len(s) + reqBigCounterSize
	if len(b) < n {
		return 0, newRequestBufferOverflowError()
	}

	binary.LittleEndian.PutUint16(b, uint16(len(s)))
	copy(b[reqBigCounterSize:], []byte(s))

	return off + n, nil
}

func getRequestDomainValue(b []byte) (domain.Name, int, error) {
	s, n, err := getRequestStringValue(b)
	if err != nil {
		return domain.Name{}, 0, err
	}

	d, err := domain.MakeNameFromString(s)
	if err != nil {
		return domain.Name{}, 0, err
	}

	return d, n, nil
}

// GetInfoRequestDomainValue extracts domain name from request for additional
// information.
func GetInfoRequestDomainValue(b []byte) (domain.Name, []byte, error) {
	if len(b) < reqTypeSize {
		return domain.Name{}, nil, newRequestBufferUnderflowError()
	}

	if t := int(b[0]); t != requestWireTypeDomain {
		return domain.Name{}, nil, newRequestAttributeUnmarshallingDomainTypeError(t)
	}
	b = b[reqTypeSize:]

	v, n, err := getRequestDomainValue(b)
	if err != nil {
		return domain.Name{}, nil, err
	}

	return v, b[n:], nil
}

func putRequestAttributeSetOfStrings(b []byte, name string, value *strtree.Tree) (int, error) {
	off, err := putRequestAttributeName(b, name)
	if err != nil {
		return 0, err
	}

	n, err := putRequestSetOfStringsValue(b[off:], value)
	if err != nil {
		return 0, err
	}

	return off + n, err
}

func putRequestSetOfStringsValue(b []byte, value *strtree.Tree) (int, error) {
	off, err := putRequestAttributeType(b, requestWireTypeSetOfStrings)
	if err != nil {
		return 0, err
	}

	ss := SortSetOfStrings(value)

	if len(ss) > math.MaxUint16 {
		return 0, newRequestTooLongCollectionValueError(TypeSetOfStrings, len(ss))
	}

	total := reqBigCounterSize * (len(ss) + 1)
	for i, s := range ss {
		if len(s) > math.MaxUint16 {
			return 0, bindErrorf(newRequestTooLongStringValueError(s), "%d", i+1)
		}

		total += len(s)
	}

	if len(b[off:]) < total {
		return 0, newRequestBufferOverflowError()
	}

	binary.LittleEndian.PutUint16(b[off:], uint16(len(ss)))
	off += reqBigCounterSize

	for _, s := range ss {
		binary.LittleEndian.PutUint16(b[off:], uint16(len(s)))
		off += reqBigCounterSize

		copy(b[off:], s)
		off += len(s)
	}

	return off, nil
}

func getRequestSetOfStringsValue(b []byte) (*strtree.Tree, int, error) {
	if len(b) < reqBigCounterSize {
		return nil, 0, newRequestBufferUnderflowError()
	}

	ss := strtree.NewTree()

	count := int(binary.LittleEndian.Uint16(b))
	off := reqBigCounterSize

	for i := 0; i < count; i++ {
		s, n, err := getRequestStringValue(b[off:])
		if err != nil {
			return nil, 0, bindErrorf(err, "%d", i+1)
		}

		off += n

		ss.InplaceInsert(s, i)
	}

	return ss, off, nil
}

// GetInfoRequestSetOfStringsValue extracts set of strings from request for
// additional information.
func GetInfoRequestSetOfStringsValue(b []byte) (*strtree.Tree, []byte, error) {
	if len(b) < reqTypeSize {
		return nil, nil, newRequestBufferUnderflowError()
	}

	if t := int(b[0]); t != requestWireTypeSetOfStrings {
		return nil, nil, newRequestAttributeUnmarshallingSetOfStringsTypeError(t)
	}
	b = b[reqTypeSize:]

	v, n, err := getRequestSetOfStringsValue(b)
	if err != nil {
		return nil, nil, err
	}

	return v, b[n:], nil
}

func putRequestAttributeSetOfNetworks(b []byte, name string, value *iptree.Tree) (int, error) {
	off, err := putRequestAttributeName(b, name)
	if err != nil {
		return 0, err
	}

	n, err := putRequestSetOfNetworksValue(b[off:], value)
	if err != nil {
		return 0, err
	}

	return off + n, err
}

func putRequestSetOfNetworksValue(b []byte, value *iptree.Tree) (int, error) {
	off, err := putRequestAttributeType(b, requestWireTypeSetOfNetworks)
	if err != nil {
		return 0, err
	}

	sn := SortSetOfNetworks(value)

	if len(sn) > math.MaxUint16 {
		return 0, newRequestTooLongCollectionValueError(TypeSetOfNetworks, len(sn))
	}

	total := len(sn) + reqBigCounterSize
	for _, n := range sn {
		total += len(n.IP)
	}

	if len(b[off:]) < total {
		return 0, newRequestBufferOverflowError()
	}

	binary.LittleEndian.PutUint16(b[off:], uint16(len(sn)))
	off += reqBigCounterSize

	for _, n := range sn {
		ones, bits := n.Mask.Size()
		if bits == 32 {
			ones += 0xc0
		}

		b[off] = byte(ones)
		copy(b[off+reqNetworkCIDRSize:], n.IP)
		off += len(n.IP) + reqNetworkCIDRSize
	}

	return off, nil
}

func getRequestSetOfNetworksValue(b []byte) (*iptree.Tree, int, error) {
	if len(b) < reqBigCounterSize {
		return nil, 0, newRequestBufferUnderflowError()
	}

	sn := iptree.NewTree()

	count := int(binary.LittleEndian.Uint16(b))
	off := reqBigCounterSize

	var (
		size int
		mask net.IPMask
		n    net.IPNet
	)
	for i := 0; i < count; i++ {
		if len(b) <= off {
			return nil, 0, newRequestBufferUnderflowError()
		}
		ones := b[off]
		off++

		if ones >= 0xc0 {
			size = reqIPv4AddressValueSize

			ones -= 0xc0
			mask = net.CIDRMask(int(ones), 32)
			if mask == nil {
				return nil, 0, bindError(bindErrorf(newRequestIPv4InvalidMaskError(ones), "%d", i+1),
					TypeSetOfNetworks.String())
			}
		} else {
			size = reqIPv6AddressValueSize

			mask = net.CIDRMask(int(ones), 128)
			if mask == nil {
				return nil, 0, bindError(bindErrorf(newRequestIPv6InvalidMaskError(ones), "%d", i+1),
					TypeSetOfNetworks.String())
			}
		}

		if len(b) < off+size {
			return nil, 0, newRequestBufferUnderflowError()
		}
		ip := net.IP(b[off : off+size])
		off += size

		n = net.IPNet{
			IP:   ip.Mask(mask),
			Mask: mask,
		}

		sn.InplaceInsertNet(&n, i)
	}

	return sn, off, nil
}

// GetInfoRequestSetOfNetworksValue extracts set of networks from request for
// additional information.
func GetInfoRequestSetOfNetworksValue(b []byte) (*iptree.Tree, []byte, error) {
	if len(b) < reqTypeSize {
		return nil, nil, newRequestBufferUnderflowError()
	}

	if t := int(b[0]); t != requestWireTypeSetOfNetworks {
		return nil, nil, newRequestAttributeUnmarshallingSetOfNetworksTypeError(t)
	}
	b = b[reqTypeSize:]

	v, n, err := getRequestSetOfNetworksValue(b)
	if err != nil {
		return nil, nil, err
	}

	return v, b[n:], nil
}

func putRequestAttributeSetOfDomains(b []byte, name string, value *domaintree.Node) (int, error) {
	off, err := putRequestAttributeName(b, name)
	if err != nil {
		return 0, err
	}

	n, err := putRequestSetOfDomainsValue(b[off:], value)
	if err != nil {
		return 0, err
	}

	return off + n, err
}

func putRequestSetOfDomainsValue(b []byte, value *domaintree.Node) (int, error) {
	off, err := putRequestAttributeType(b, requestWireTypeSetOfDomains)
	if err != nil {
		return 0, err
	}

	sd := SortSetOfDomains(value)

	if len(sd) > math.MaxUint16 {
		return 0, newRequestTooLongCollectionValueError(TypeSetOfDomains, len(sd))
	}

	total := reqBigCounterSize * (len(sd) + 1)
	for _, s := range sd {
		total += len(s)
	}

	if len(b[off:]) < total {
		return 0, newRequestBufferOverflowError()
	}

	binary.LittleEndian.PutUint16(b[off:], uint16(len(sd)))
	off += reqBigCounterSize

	for _, s := range sd {
		binary.LittleEndian.PutUint16(b[off:], uint16(len(s)))
		off += reqBigCounterSize

		copy(b[off:], s)
		off += len(s)
	}

	return off, nil
}

func getRequestSetOfDomainsValue(b []byte) (*domaintree.Node, int, error) {
	if len(b) < reqBigCounterSize {
		return nil, 0, newRequestBufferUnderflowError()
	}

	sd := new(domaintree.Node)

	count := int(binary.LittleEndian.Uint16(b))
	off := reqBigCounterSize

	for i := 0; i < count; i++ {
		d, n, err := getRequestDomainValue(b[off:])
		if err != nil {
			return nil, 0, bindErrorf(err, "%d", i+1)
		}

		off += n

		sd.InplaceInsert(d, i)
	}

	return sd, off, nil
}

// GetInfoRequestSetOfDomainsValue extracts set of domains from request for
// additional information.
func GetInfoRequestSetOfDomainsValue(b []byte) (*domaintree.Node, []byte, error) {
	if len(b) < reqTypeSize {
		return nil, nil, newRequestBufferUnderflowError()
	}

	if t := int(b[0]); t != requestWireTypeSetOfDomains {
		return nil, nil, newRequestAttributeUnmarshallingSetOfDomainsTypeError(t)
	}
	b = b[reqTypeSize:]

	v, n, err := getRequestSetOfDomainsValue(b)
	if err != nil {
		return nil, nil, err
	}

	return v, b[n:], nil
}

func putRequestAttributeListOfStrings(b []byte, name string, value []string) (int, error) {
	off, err := putRequestAttributeName(b, name)
	if err != nil {
		return 0, err
	}

	n, err := putRequestListOfStringsValue(b[off:], value)
	if err != nil {
		return 0, err
	}

	return off + n, err
}

func putRequestListOfStringsValue(b []byte, value []string) (int, error) {
	off, err := putRequestAttributeType(b, requestWireTypeListOfStrings)
	if err != nil {
		return 0, err
	}

	if len(value) > math.MaxUint16 {
		return 0, newRequestTooLongCollectionValueError(TypeListOfStrings, len(value))
	}

	total := reqBigCounterSize * (len(value) + 1)
	for i, s := range value {
		if len(s) > math.MaxUint16 {
			return 0, bindErrorf(newRequestTooLongStringValueError(s), "%d", i+1)
		}

		total += len(s)
	}

	if len(b[off:]) < total {
		return 0, newRequestBufferOverflowError()
	}

	binary.LittleEndian.PutUint16(b[off:], uint16(len(value)))
	off += reqBigCounterSize

	for _, s := range value {
		binary.LittleEndian.PutUint16(b[off:], uint16(len(s)))
		off += reqBigCounterSize

		copy(b[off:], s)
		off += len(s)
	}

	return off, nil
}

func putRequestSetOfFlags8Value(b []byte, value uint8, t *FlagsType) (int, error) {
	off, err := putRequestAttributeType(b, requestWireTypeSetOfFlags)
	if err != nil {
		return 0, err
	}

	total := reqSmallCounterSize + 1
	if len(b[off:]) < total {
		return 0, newRequestBufferOverflowError()
	}

	b[off] = byte(len(t.b))
	b[off+1] = value

	return off + total, nil
}

func putRequestSetOfFlags16Value(b []byte, value uint16, t *FlagsType) (int, error) {
	off, err := putRequestAttributeType(b, requestWireTypeSetOfFlags)
	if err != nil {
		return 0, err
	}

	total := reqSmallCounterSize + 2
	if len(b[off:]) < total {
		return 0, newRequestBufferOverflowError()
	}

	b[off] = byte(len(t.b))
	binary.LittleEndian.PutUint16(b[off+1:], value)

	return off + total, nil
}

func putRequestSetOfFlags32Value(b []byte, value uint32, t *FlagsType) (int, error) {
	off, err := putRequestAttributeType(b, requestWireTypeSetOfFlags)
	if err != nil {
		return 0, err
	}

	total := reqSmallCounterSize + 4
	if len(b[off:]) < total {
		return 0, newRequestBufferOverflowError()
	}

	b[off] = byte(len(t.b))
	binary.LittleEndian.PutUint32(b[off+1:], value)

	return off + total, nil
}

func putRequestSetOfFlags64Value(b []byte, value uint64, t *FlagsType) (int, error) {
	off, err := putRequestAttributeType(b, requestWireTypeSetOfFlags)
	if err != nil {
		return 0, err
	}

	total := reqSmallCounterSize + 8
	if len(b[off:]) < total {
		return 0, newRequestBufferOverflowError()
	}

	b[off] = byte(len(t.b))
	binary.LittleEndian.PutUint64(b[off+1:], value)

	return off + total, nil
}

func getRequestListOfStringsValue(b []byte) ([]string, int, error) {
	if len(b) < reqBigCounterSize {
		return nil, 0, newRequestBufferUnderflowError()
	}

	count := int(binary.LittleEndian.Uint16(b))
	off := reqBigCounterSize

	ls := make([]string, count)

	for i := 0; i < count; i++ {
		s, n, err := getRequestStringValue(b[off:])
		if err != nil {
			return nil, 0, bindErrorf(err, "%d", i+1)
		}

		off += n

		ls[i] = s
	}

	return ls, off, nil
}

// GetInfoRequestListOfStringsValue extracts list of strings from request for
// additional information.
func GetInfoRequestListOfStringsValue(b []byte) ([]string, []byte, error) {
	if len(b) < reqTypeSize {
		return nil, nil, newRequestBufferUnderflowError()
	}

	if t := int(b[0]); t != requestWireTypeListOfStrings {
		return nil, nil, newRequestAttributeUnmarshallingListOfStringsTypeError(t)
	}
	b = b[reqTypeSize:]

	v, n, err := getRequestListOfStringsValue(b)
	if err != nil {
		return nil, nil, err
	}

	return v, b[n:], nil
}

func getRequestAbstractSetOfFlagsValue(b []byte) (uint64, int, int, error) {
	if len(b) < reqSmallCounterSize {
		return 0, 0, 0, newRequestBufferUnderflowError()
	}

	s := int(b[0])
	b = b[1:]

	switch {
	case s <= 8:
		if len(b) < 1 {
			return 0, 0, 0, newRequestBufferUnderflowError()
		}

		return uint64(b[0]), s, reqSmallCounterSize + 1, nil

	case s <= 16:
		if len(b) < 2 {
			return 0, 0, 0, newRequestBufferUnderflowError()
		}

		return uint64(binary.LittleEndian.Uint16(b)), s, reqSmallCounterSize + 2, nil

	case s <= 32:
		if len(b) < 4 {
			return 0, 0, 0, newRequestBufferUnderflowError()
		}

		return uint64(binary.LittleEndian.Uint32(b)), s, reqSmallCounterSize + 4, nil
	}

	if len(b) < 8 {
		return 0, 0, 0, newRequestBufferUnderflowError()
	}

	return binary.LittleEndian.Uint64(b), s, reqSmallCounterSize + 8, nil
}

func calcRequestSize(in []AttributeAssignment) (int, error) {
	s, err := calcAssignmentExpressionsSize(in)
	if err != nil {
		return 0, err
	}

	return reqVersionSize + s, nil
}

func calcRequestSizeFromReflection(c int, f func(i int) (string, Type, reflect.Value, error)) (int, error) {
	s, err := calcAttributesSizeFromReflection(c, f)
	if err != nil {
		return 0, err
	}

	return reqVersionSize + s, nil
}

func calcRequestAttributeSize(value AttributeValue) (int, error) {
	var (
		s   int
		err error
	)

	t := value.GetResultType()
	switch t {
	default:
		return 0, newRequestAttributeMarshallingNotImplementedError(t)

	case TypeBoolean:
		break

	case TypeString:
		v, _ := value.str()
		s, err = calcRequestAttributeStringSize(v)

	case TypeInteger:
		v, _ := value.integer()
		s, err = calcRequestAttributeIntegerSize(v)

	case TypeFloat:
		v, _ := value.float()
		s, err = calcRequestAttributeFloatSize(v)

	case TypeAddress:
		v, _ := value.address()
		s, err = calcRequestAttributeAddressSize(v)

	case TypeNetwork:
		v, _ := value.network()
		s, err = calcRequestAttributeNetworkSize(v)

	case TypeDomain:
		v, _ := value.domain()
		s, err = calcRequestAttributeDomainSize(v)

	case TypeSetOfStrings:
		v, _ := value.setOfStrings()
		s, err = calcRequestAttributeSetOfStringsSize(v)

	case TypeSetOfNetworks:
		v, _ := value.setOfNetworks()
		s, err = calcRequestAttributeSetOfNetworksSize(v)

	case TypeSetOfDomains:
		v, _ := value.setOfDomains()
		s, err = calcRequestAttributeSetOfDomainsSize(v)

	case TypeListOfStrings:
		v, _ := value.listOfStrings()
		s, err = calcRequestAttributeListOfStringsSize(v)
	}

	return reqTypeSize + s, err
}

func calcRequestAttributeNameSize(name string) (int, error) {
	if len(name) > math.MaxUint8 {
		return 0, newRequestTooLongAttributeNameError(name)
	}

	return reqSmallCounterSize + len(name), nil
}

func calcRequestAttributeStringSize(value string) (int, error) {
	if len(value) > math.MaxUint16 {
		return 0, newRequestTooLongStringValueError(value)
	}

	return reqBigCounterSize + len(value), nil
}

func calcRequestAttributeIntegerSize(value int64) (int, error) {
	return reqIntegerValueSize, nil
}

func calcRequestAttributeFloatSize(value float64) (int, error) {
	return reqFloatValueSize, nil
}

func calcRequestAttributeAddressSize(value net.IP) (int, error) {
	if ip := value.To4(); ip != nil {
		return len(ip), nil
	}

	if ip := value.To16(); ip != nil {
		return len(ip), nil
	}

	return 0, newRequestAddressValueError(value)
}

func calcRequestAttributeNetworkSize(value *net.IPNet) (int, error) {
	if value == nil {
		return 0, newRequestInvalidNetworkValueError(value)
	}

	ip := value.IP
	if len(ip) != 4 && len(ip) != 16 {
		return 0, newRequestInvalidNetworkValueError(value)
	}

	_, bits := value.Mask.Size()
	if bits == 32 || bits == 128 {
		return reqNetworkCIDRSize + len(ip), nil
	}

	return 0, newRequestInvalidNetworkValueError(value)
}

func calcRequestAttributeDomainSize(value domain.Name) (int, error) {
	return calcRequestAttributeStringSize(value.String())
}

func calcRequestAttributeSetOfStringsSize(value *strtree.Tree) (int, error) {
	ss := SortSetOfStrings(value)

	if len(ss) > math.MaxUint16 {
		return 0, newRequestTooLongCollectionValueError(TypeSetOfStrings, len(ss))
	}

	total := reqBigCounterSize * (len(ss) + 1)
	for i, s := range ss {
		if len(s) > math.MaxUint16 {
			return 0, bindErrorf(newRequestTooLongStringValueError(s), "%d", i+1)
		}

		total += len(s)
	}

	return total, nil
}

func calcRequestAttributeSetOfNetworksSize(value *iptree.Tree) (int, error) {
	sn := SortSetOfNetworks(value)

	if len(sn) > math.MaxUint16 {
		return 0, newRequestTooLongCollectionValueError(TypeSetOfNetworks, len(sn))
	}

	total := len(sn) + reqBigCounterSize
	for _, n := range sn {
		total += len(n.IP)
	}

	return total, nil
}

func calcRequestAttributeSetOfDomainsSize(value *domaintree.Node) (int, error) {
	sd := SortSetOfDomains(value)

	if len(sd) > math.MaxUint16 {
		return 0, newRequestTooLongCollectionValueError(TypeSetOfDomains, len(sd))
	}

	total := reqBigCounterSize * (len(sd) + 1)
	for _, s := range sd {
		total += len(s)
	}

	return total, nil
}

func calcRequestAttributeListOfStringsSize(value []string) (int, error) {
	if len(value) > math.MaxUint16 {
		return 0, newRequestTooLongCollectionValueError(TypeListOfStrings, len(value))
	}

	total := reqBigCounterSize * (len(value) + 1)
	for i, s := range value {
		if len(s) > math.MaxUint16 {
			return 0, bindErrorf(newRequestTooLongStringValueError(s), "%d", i+1)
		}

		total += len(s)
	}

	return total, nil
}

func getRequestWireTypeName(t int) string {
	if t < 0 || t >= len(requestWireTypeNames) {
		return fmt.Sprintf("unknown (%d)", t)
	}

	return requestWireTypeNames[t]
}
