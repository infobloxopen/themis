package pdp

import (
	"encoding/binary"
	"math"
	"net"

	"github.com/infobloxopen/go-trees/domain"
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
)

var requestWireTypeNames = []string{
	"boolean true",
	"boolean false",
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
}

func checkRequestVersion(b []byte) (int, error) {
	if len(b) < 2 {
		return 0, newRequestBufferUnderflow()
	}

	if v := binary.LittleEndian.Uint16(b); v != requestVersion {
		return 0, newRequestVersionError(v, requestVersion)
	}

	return 2, nil
}

func getRequestAttributeCount(b []byte) (int, int, error) {
	if len(b) < 2 {
		return 0, 0, newRequestBufferUnderflow()
	}

	return int(binary.LittleEndian.Uint16(b)), 2, nil
}

func getRequestAttribute(b []byte) (string, AttributeValue, int, error) {
	name, off, err := getRequestAttributeName(b)
	if err != nil {
		return "", UndefinedValue, 0, bindError(err, "name")
	}

	t, n, err := getRequestAttributeType(b[off:])
	if err != nil {
		return "", UndefinedValue, 0, bindError(bindError(err, "type"), name)
	}

	off += n

	switch t {
	case requestWireTypeSetOfStrings, requestWireTypeSetOfNetworks, requestWireTypeSetOfDomains, requestWireTypeListOfStrings:
		return "", UndefinedValue, 0, bindError(newRequestAttributeUnmarshallingNotImplemented(t), name)

	case requestWireTypeBooleanFalse:
		return name, MakeBooleanValue(false), off, nil

	case requestWireTypeBooleanTrue:
		return name, MakeBooleanValue(true), off, nil

	case requestWireTypeString:
		s, n, err := getRequestStringValue(b[off:])
		if err != nil {
			return "", UndefinedValue, 0, bindError(bindError(err, "value"), name)
		}

		return name, MakeStringValue(s), off + n, nil

	case requestWireTypeInteger:
		i, n, err := getRequestIntegerValue(b[off:])
		if err != nil {
			return "", UndefinedValue, 0, bindError(bindError(err, "value"), name)
		}

		return name, MakeIntegerValue(i), off + n, nil

	case requestWireTypeFloat:
		f, n, err := getRequestFloatValue(b[off:])
		if err != nil {
			return "", UndefinedValue, 0, bindError(bindError(err, "value"), name)
		}

		return name, MakeFloatValue(f), off + n, nil

	case requestWireTypeIPv4Address:
		a, n, err := getRequestIPv4AddressValue(b[off:])
		if err != nil {
			return "", UndefinedValue, 0, bindError(bindError(err, "value"), name)
		}

		return name, MakeAddressValue(a), off + n, nil

	case requestWireTypeIPv6Address:
		a, n, err := getRequestIPv6AddressValue(b[off:])
		if err != nil {
			return "", UndefinedValue, 0, bindError(bindError(err, "value"), name)
		}

		return name, MakeAddressValue(a), off + n, nil

	case requestWireTypeIPv4Network:
		a, n, err := getRequestIPv4NetworkValue(b[off:])
		if err != nil {
			return "", UndefinedValue, 0, bindError(bindError(err, "value"), name)
		}

		return name, MakeNetworkValue(a), off + n, nil

	case requestWireTypeIPv6Network:
		a, n, err := getRequestIPv6NetworkValue(b[off:])
		if err != nil {
			return "", UndefinedValue, 0, bindError(bindError(err, "value"), name)
		}

		return name, MakeNetworkValue(a), off + n, nil

	case requestWireTypeDomain:
		d, n, err := getRequestDomainValue(b[off:])
		if err != nil {
			return "", UndefinedValue, 0, bindError(bindError(err, "value"), name)
		}

		return name, MakeDomainValue(d), off + n, nil
	}

	return "", UndefinedValue, 0, bindError(newRequestAttributeTypeError(t), name)
}

func getRequestAttributeName(b []byte) (string, int, error) {
	if len(b) < 1 {
		return "", 0, newRequestBufferUnderflow()
	}

	off := int(b[0]) + 1
	if len(b) < off {
		return "", 0, newRequestBufferUnderflow()
	}

	return string(b[1:off]), off, nil
}

func getRequestAttributeType(b []byte) (int, int, error) {
	if len(b) < 1 {
		return 0, 0, newRequestBufferUnderflow()
	}

	return int(b[0]), 1, nil
}

func getRequestStringValue(b []byte) (string, int, error) {
	if len(b) < 2 {
		return "", 0, newRequestBufferUnderflow()
	}

	off := int(binary.LittleEndian.Uint16(b)) + 2
	if len(b) < off {
		return "", 0, newRequestBufferUnderflow()
	}

	return string(b[2:off]), off, nil
}

func getRequestIntegerValue(b []byte) (int64, int, error) {
	if len(b) < 8 {
		return 0, 0, newRequestBufferUnderflow()
	}

	return int64(binary.LittleEndian.Uint64(b)), 8, nil
}

func getRequestFloatValue(b []byte) (float64, int, error) {
	if len(b) < 8 {
		return 0, 0, newRequestBufferUnderflow()
	}

	return math.Float64frombits(binary.LittleEndian.Uint64(b)), 8, nil
}

func getRequestIPv4AddressValue(b []byte) (net.IP, int, error) {
	if len(b) < 4 {
		return nil, 0, newRequestBufferUnderflow()
	}

	return net.IPv4(b[0], b[1], b[2], b[3]), 4, nil
}

func getRequestIPv6AddressValue(b []byte) (net.IP, int, error) {
	if len(b) < 16 {
		return nil, 0, newRequestBufferUnderflow()
	}

	ip := net.IP(make([]byte, 16))
	copy(ip, b)
	return ip, 16, nil
}

func getRequestIPv4NetworkValue(b []byte) (*net.IPNet, int, error) {
	if len(b) < 5 {
		return nil, 0, newRequestBufferUnderflow()
	}

	mask := net.CIDRMask(int(b[0]), 32)
	if mask == nil {
		return nil, 0, newRequestIPv4InvalidMaskError(b[0])
	}

	return &net.IPNet{
		IP:   net.IPv4(b[1], b[2], b[3], b[4]).Mask(mask),
		Mask: mask,
	}, 5, nil
}

func getRequestIPv6NetworkValue(b []byte) (*net.IPNet, int, error) {
	if len(b) < 17 {
		return nil, 0, newRequestBufferUnderflow()
	}

	mask := net.CIDRMask(int(b[0]), 128)
	if mask == nil {
		return nil, 0, newRequestIPv6InvalidMaskError(b[0])
	}

	ip := net.IP(make([]byte, 16))
	copy(ip, b[1:])

	return &net.IPNet{
		IP:   ip.Mask(mask),
		Mask: mask,
	}, 17, nil
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
