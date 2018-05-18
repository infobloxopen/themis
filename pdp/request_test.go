package pdp

import (
	"bytes"
	"errors"
	"fmt"
	"math"
	"net"
	"reflect"
	"testing"

	"github.com/infobloxopen/go-trees/domain"
)

var (
	testWireRequest = []byte{
		1, 0, 3, 0,
		6, 's', 't', 'r', 'i', 'n', 'g', byte(requestWireTypeString), 4, 0, 't', 'e', 's', 't',
		7, 'b', 'o', 'o', 'l', 'e', 'a', 'n', byte(requestWireTypeBooleanTrue),
		7, 'i', 'n', 't', 'e', 'g', 'e', 'r', byte(requestWireTypeInteger),
		0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x7f,
	}

	testRequestAssignments = []AttributeAssignmentExpression{
		MakeAttributeAssignmentExpression(
			MakeAttribute("string", TypeString),
			MakeStringValue("test"),
		),
		MakeAttributeAssignmentExpression(
			MakeAttribute("boolean", TypeBoolean),
			MakeBooleanValue(true),
		),
		MakeAttributeAssignmentExpression(
			MakeAttribute("integer", TypeInteger),
			MakeIntegerValue(9223372036854775807),
		),
	}
)

func TestRequestWireTypesTotal(t *testing.T) {
	if requestWireTypesTotal != len(requestWireTypeNames) {
		t.Errorf("Expected number of wire type names %d to be equal to total number of wire types %d",
			len(requestWireTypeNames), requestWireTypesTotal)
	}
}

func TestMarshalRequestAssignments(t *testing.T) {
	var b [44]byte
	n, err := MarshalRequestAssignments(b[:], testRequestAssignments)
	assertRequestBytesBuffer(t, "MarshalRequestAssignments", err, b[:], n, testWireRequest...)

	n, err = MarshalRequestAssignments([]byte{}, testRequestAssignments)
	assertRequestBufferOverflow(t, "MarshalRequestAssignments", err, n)

	n, err = MarshalRequestAssignments(b[:2], testRequestAssignments)
	if err == nil {
		t.Errorf("expected no data put to small buffer but got %d", n)
	} else if _, ok := err.(*requestBufferOverflowError); !ok {
		t.Errorf("expected *requestBufferOverflowError but got %T (%s)", err, err)
	}

	n, err = MarshalRequestAssignments(b[:], []AttributeAssignmentExpression{
		MakeAttributeAssignmentExpression(
			MakeAttribute("boolean", TypeBoolean),
			makeFunctionBooleanNot([]Expression{MakeBooleanValue(true)}),
		),
	})
	if err == nil {
		t.Errorf("expected no data put to buffer for invalid expression but got %d", n)
	} else if _, ok := err.(*requestInvalidExpressionError); !ok {
		t.Errorf("expected *requestInvalidExpressionError but got %T (%s)", err, err)
	}

	n, err = MarshalRequestAssignments(b[:12], testRequestAssignments)
	if err == nil {
		t.Errorf("expected no data put to small buffer but got %d", n)
	} else if _, ok := err.(*requestBufferOverflowError); !ok {
		t.Errorf("expected *requestBufferOverflowError but got %T (%s)", err, err)
	}
}

func TestMarshalRequestReflection(t *testing.T) {
	var b [111]byte

	f := func(i int) (string, Type, reflect.Value, error) {
		switch i {
		case 0:
			return "boolean", TypeBoolean, reflect.ValueOf(true), nil

		case 1:
			return "string", TypeString, reflect.ValueOf("test"), nil

		case 2:
			return "integer", TypeInteger, reflect.ValueOf(int64(9223372036854775807)), nil

		case 3:
			return "float", TypeFloat, reflect.ValueOf(float64(math.Pi)), nil

		case 4:
			return "address", TypeAddress, reflect.ValueOf(net.ParseIP("192.0.2.1")), nil

		case 5:
			return "network", TypeNetwork, reflect.ValueOf(makeTestNetwork("192.0.2.0/24")), nil

		case 6:
			return "domain", TypeDomain, reflect.ValueOf(makeTestDomain("www.example.com")), nil
		}

		return "", TypeUndefined, reflect.ValueOf(nil), fmt.Errorf("unexpected intex %d", i)
	}
	n, err := MarshalRequestReflection(b[:], 7, f)
	assertRequestBytesBuffer(t, "MarshalRequestReflection", err, b[:], n,
		1, 0, 7, 0,
		7, 'b', 'o', 'o', 'l', 'e', 'a', 'n', byte(requestWireTypeBooleanTrue),
		6, 's', 't', 'r', 'i', 'n', 'g', byte(requestWireTypeString), 4, 0, 't', 'e', 's', 't',
		7, 'i', 'n', 't', 'e', 'g', 'e', 'r', byte(requestWireTypeInteger),
		0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x7f,
		5, 'f', 'l', 'o', 'a', 't', byte(requestWireTypeFloat), 24, 45, 68, 84, 251, 33, 9, 64,
		7, 'a', 'd', 'd', 'r', 'e', 's', 's', byte(requestWireTypeIPv4Address), 192, 0, 2, 1,
		7, 'n', 'e', 't', 'w', 'o', 'r', 'k', byte(requestWireTypeIPv4Network), 24, 192, 0, 2, 0,
		6, 'd', 'o', 'm', 'a', 'i', 'n', byte(requestWireTypeDomain),
		15, 0, 'w', 'w', 'w', '.', 'e', 'x', 'a', 'm', 'p', 'l', 'e', '.', 'c', 'o', 'm',
	)

	n, err = MarshalRequestReflection([]byte{}, 1, f)
	assertRequestBufferOverflow(t, "MarshalRequestReflection", err, n)

	n, err = MarshalRequestReflection(b[:2], 1, f)
	if err == nil {
		t.Errorf("expected no data put to small buffer but got %d", n)
	} else if _, ok := err.(*requestBufferOverflowError); !ok {
		t.Errorf("expected *requestBufferOverflowError but got %T (%s)", err, err)
	}

	testFuncErr := errors.New("test function error")
	n, err = MarshalRequestReflection(b[:], 1, func(i int) (string, Type, reflect.Value, error) {
		return "", TypeUndefined, reflect.ValueOf(nil), testFuncErr
	})
	if err == nil {
		t.Errorf("expected no data put to buffer for broken function but got %d", n)
	} else if err != testFuncErr {
		t.Errorf("expected testFuncErr but got %T (%s)", err, err)
	}

	n, err = MarshalRequestReflection(b[:], 1, func(i int) (string, Type, reflect.Value, error) {
		return "undefined", TypeUndefined, reflect.ValueOf(nil), nil
	})
	if err == nil {
		t.Errorf("expected no data put to buffer for undefined value but got %d", n)
	} else if _, ok := err.(*requestAttributeMarshallingNotImplemented); !ok {
		t.Errorf("expected *requestAttributeMarshallingNotImplemented but got %T (%s)", err, err)
	}

	n, err = MarshalRequestReflection(b[:12], 1, f)
	if err == nil {
		t.Errorf("expected no data put to small buffer but got %d", n)
	} else if _, ok := err.(*requestBufferOverflowError); !ok {
		t.Errorf("expected *requestBufferOverflowError but got %T (%s)", err, err)
	}
}

func TestUnmarshalRequestAssignments(t *testing.T) {
	var a [3]AttributeAssignmentExpression

	n, err := UnmarshalRequestAssignments(testWireRequest, a[:])
	assertRequestAssignmentExpressions(t, "UnmarshalRequestAssignments", err, a[:], n, testRequestAssignments...)

	n, err = UnmarshalRequestAssignments([]byte{}, a[:])
	if err == nil {
		t.Errorf("expected *requestBufferUnderflowError but got %d bytes", n)
	} else if _, ok := err.(*requestBufferUnderflowError); !ok {
		t.Errorf("expected *requestBufferUnderflowError but got %T (%s)", err, err)
	}

	n, err = UnmarshalRequestAssignments([]byte{
		1, 0,
	}, a[:])
	if err == nil {
		t.Errorf("expected *requestBufferUnderflowError but got %d bytes", n)
	} else if _, ok := err.(*requestBufferUnderflowError); !ok {
		t.Errorf("expected *requestBufferUnderflowError but got %T (%s)", err, err)
	}

	n, err = UnmarshalRequestAssignments([]byte{
		1, 0, 255, 255,
	}, a[:])
	if err == nil {
		t.Errorf("expected *requestAssignmentsOverflowError but got %d bytes", n)
	} else if _, ok := err.(*requestAssignmentsOverflowError); !ok {
		t.Errorf("expected *requestAssignmentsOverflowError but got %T (%s)", err, err)
	}

	n, err = UnmarshalRequestAssignments([]byte{
		1, 0, 1, 0,
	}, a[:])
	if err == nil {
		t.Errorf("expected *requestBufferUnderflowError but got got %d bytes", n)
	} else if _, ok := err.(*requestBufferUnderflowError); !ok {
		t.Errorf("expected *requestBufferUnderflowError but got %T (%s)", err, err)
	}
}

func TestUnmarshalRequestReflection(t *testing.T) {
	var (
		names [10]string
	)

	i := 0

	booleanFalse := true
	booleanTrue := false

	var (
		str string
		num int64
		flt float64
		av4 net.IP
		av6 net.IP
		nv4 *net.IPNet
		nv6 *net.IPNet
		dn  domain.Name
	)

	values := []reflect.Value{
		reflect.Indirect(reflect.ValueOf(&booleanFalse)),
		reflect.Indirect(reflect.ValueOf(&booleanTrue)),
		reflect.Indirect(reflect.ValueOf(&str)),
		reflect.Indirect(reflect.ValueOf(&num)),
		reflect.Indirect(reflect.ValueOf(&flt)),
		reflect.Indirect(reflect.ValueOf(&av4)),
		reflect.Indirect(reflect.ValueOf(&av6)),
		reflect.Indirect(reflect.ValueOf(&nv4)),
		reflect.Indirect(reflect.ValueOf(&nv6)),
		reflect.Indirect(reflect.ValueOf(&dn)),
	}

	err := UnmarshalRequestReflection([]byte{
		1, 0, byte(len(names)), 0,
		12, 'b', 'o', 'o', 'l', 'e', 'a', 'n', 'F', 'a', 'l', 's', 'e', byte(requestWireTypeBooleanFalse),
		11, 'b', 'o', 'o', 'l', 'e', 'a', 'n', 'T', 'r', 'u', 'e', byte(requestWireTypeBooleanTrue),
		6, 's', 't', 'r', 'i', 'n', 'g', byte(requestWireTypeString), 4, 0, 't', 'e', 's', 't',
		7, 'i', 'n', 't', 'e', 'g', 'e', 'r', byte(requestWireTypeInteger), 1, 0, 0, 0, 0, 0, 0, 0,
		5, 'f', 'l', 'o', 'a', 't', byte(requestWireTypeFloat), 24, 45, 68, 84, 251, 33, 9, 64,
		8, 'a', 'd', 'd', 'r', 'e', 's', 's', '4', byte(requestWireTypeIPv4Address), 192, 0, 2, 1,
		8, 'a', 'd', 'd', 'r', 'e', 's', 's', '6', byte(requestWireTypeIPv6Address),
		32, 1, 13, 184, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1,
		8, 'n', 'e', 't', 'w', 'o', 'r', 'k', '4', byte(requestWireTypeIPv4Network), 24, 192, 0, 2, 0,
		8, 'n', 'e', 't', 'w', 'o', 'r', 'k', '6', byte(requestWireTypeIPv6Network),
		32, 32, 1, 13, 184, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		6, 'd', 'o', 'm', 'a', 'i', 'n', byte(requestWireTypeDomain),
		15, 0, 'w', 'w', 'w', '.', 'e', 'x', 'a', 'm', 'p', 'l', 'e', '.', 'c', 'o', 'm',
	}, func(id string, t Type) (reflect.Value, bool, error) {
		if i >= len(names) || i >= len(values) || i >= len(builtinTypeByWire) {
			return reflect.ValueOf(nil), false, fmt.Errorf("requested invalid value number: %d", i)
		}

		if et := builtinTypeByWire[i]; t != et {
			return reflect.ValueOf(nil), false, fmt.Errorf("expected %q for %d but got %q", et, i, t)
		}

		names[i] = id
		v := values[i]
		i++

		return v, true, nil
	})

	a := []AttributeAssignmentExpression{
		MakeAttributeAssignmentExpression(
			MakeAttribute(names[0], TypeBoolean),
			MakeBooleanValue(booleanFalse),
		),
		MakeAttributeAssignmentExpression(
			MakeAttribute(names[1], TypeBoolean),
			MakeBooleanValue(booleanTrue),
		),
		MakeAttributeAssignmentExpression(
			MakeAttribute(names[2], TypeString),
			MakeStringValue(str),
		),
		MakeAttributeAssignmentExpression(
			MakeAttribute(names[3], TypeInteger),
			MakeIntegerValue(num),
		),
		MakeAttributeAssignmentExpression(
			MakeAttribute(names[4], TypeFloat),
			MakeFloatValue(flt),
		),
		MakeAttributeAssignmentExpression(
			MakeAttribute(names[5], TypeAddress),
			MakeAddressValue(av4),
		),
		MakeAttributeAssignmentExpression(
			MakeAttribute(names[6], TypeAddress),
			MakeAddressValue(av6),
		),
		MakeAttributeAssignmentExpression(
			MakeAttribute(names[7], TypeNetwork),
			MakeNetworkValue(nv4),
		),
		MakeAttributeAssignmentExpression(
			MakeAttribute(names[8], TypeNetwork),
			MakeNetworkValue(nv6),
		),
		MakeAttributeAssignmentExpression(
			MakeAttribute(names[9], TypeDomain),
			MakeDomainValue(dn),
		),
	}

	assertRequestAssignmentExpressions(t, "UnmarshalRequestReflection", err, a, i,
		MakeAttributeAssignmentExpression(
			MakeAttribute("booleanFalse", TypeBoolean),
			MakeBooleanValue(false),
		),
		MakeAttributeAssignmentExpression(
			MakeAttribute("booleanTrue", TypeBoolean),
			MakeBooleanValue(true),
		),
		MakeAttributeAssignmentExpression(
			MakeAttribute("string", TypeString),
			MakeStringValue("test"),
		),
		MakeAttributeAssignmentExpression(
			MakeAttribute("integer", TypeInteger),
			MakeIntegerValue(1),
		),
		MakeAttributeAssignmentExpression(
			MakeAttribute("float", TypeFloat),
			MakeFloatValue(float64(math.Pi)),
		),
		MakeAttributeAssignmentExpression(
			MakeAttribute("address4", TypeAddress),
			MakeAddressValue(net.ParseIP("192.0.2.1")),
		),
		MakeAttributeAssignmentExpression(
			MakeAttribute("address6", TypeAddress),
			MakeAddressValue(net.ParseIP("2001:db8::1")),
		),
		MakeAttributeAssignmentExpression(
			MakeAttribute("network4", TypeNetwork),
			MakeNetworkValue(makeTestNetwork("192.0.2.0/24")),
		),
		MakeAttributeAssignmentExpression(
			MakeAttribute("network6", TypeNetwork),
			MakeNetworkValue(makeTestNetwork("2001:db8::/32")),
		),
		MakeAttributeAssignmentExpression(
			MakeAttribute("domain", TypeDomain),
			MakeDomainValue(makeTestDomain("www.example.com")),
		),
	)

	err = UnmarshalRequestReflection([]byte{}, func(id string, t Type) (reflect.Value, bool, error) {
		return reflect.ValueOf(nil), false, fmt.Errorf("in unreacheable place with id %q and type %q", id, t)
	})
	if err == nil {
		t.Error("expected *requestBufferUnderflowError but got nothing")
	} else if _, ok := err.(*requestBufferUnderflowError); !ok {
		t.Errorf("expected *requestBufferUnderflowError but got %T (%s)", err, err)
	}

	err = UnmarshalRequestReflection([]byte{
		1, 0,
	}, func(id string, t Type) (reflect.Value, bool, error) {
		return reflect.ValueOf(nil), false, fmt.Errorf("in unreacheable place with id %q and type %q", id, t)
	})
	if err == nil {
		t.Error("expected *requestBufferUnderflowError but got nothing")
	} else if _, ok := err.(*requestBufferUnderflowError); !ok {
		t.Errorf("expected *requestBufferUnderflowError but got %T (%s)", err, err)
	}

	err = UnmarshalRequestReflection([]byte{
		1, 0, 1, 0,
	}, func(id string, t Type) (reflect.Value, bool, error) {
		return reflect.ValueOf(nil), false, fmt.Errorf("in unreacheable place with id %q and type %q", id, t)
	})
	if err == nil {
		t.Error("expected *requestBufferUnderflowError but got nothing")
	} else if _, ok := err.(*requestBufferUnderflowError); !ok {
		t.Errorf("expected *requestBufferUnderflowError but got %T (%s)", err, err)
	}

	err = UnmarshalRequestReflection([]byte{
		1, 0, 1, 0,
		7, 's', 't', 'r', 'i', 'n', 'g', 's',
	}, func(id string, t Type) (reflect.Value, bool, error) {
		return reflect.ValueOf(nil), false, fmt.Errorf("in unreacheable place with id %q and type %q", id, t)
	})
	if err == nil {
		t.Error("expected *requestBufferUnderflowError but got nothing")
	} else if _, ok := err.(*requestBufferUnderflowError); !ok {
		t.Errorf("expected *requestBufferUnderflowError but got %T (%s)", err, err)
	}

	err = UnmarshalRequestReflection([]byte{
		1, 0, 1, 0,
		7, 's', 't', 'r', 'i', 'n', 'g', 's', byte(requestWireTypeSetOfStrings),
	}, func(id string, t Type) (reflect.Value, bool, error) {
		return reflect.ValueOf(nil), false, fmt.Errorf("in unreacheable place with id %q and type %q", id, t)
	})
	if err == nil {
		t.Error("expected *requestAttributeUnmarshallingNotImplemented but got nothing")
	} else if _, ok := err.(*requestAttributeUnmarshallingNotImplemented); !ok {
		t.Errorf("expected *requestAttributeUnmarshallingNotImplemented but got %T (%s)", err, err)
	}

	err = UnmarshalRequestReflection([]byte{
		1, 0, 1, 0,
		7, 's', 't', 'r', 'i', 'n', 'g', 's', 255,
	}, func(id string, t Type) (reflect.Value, bool, error) {
		return reflect.ValueOf(nil), false, fmt.Errorf("in unreacheable place with id %q and type %q", id, t)
	})
	if err == nil {
		t.Error("expected *requestAttributeUnmarshallingTypeError but got nothing")
	} else if _, ok := err.(*requestAttributeUnmarshallingTypeError); !ok {
		t.Errorf("expected *requestAttributeUnmarshallingTypeError but got %T (%s)", err, err)
	}

	testFuncErr := errors.New("test function error")
	err = UnmarshalRequestReflection(testWireRequest, func(id string, t Type) (reflect.Value, bool, error) {
		return reflect.ValueOf(nil), false, testFuncErr
	})
	if err == nil {
		t.Error("expected testFuncErr but got nothing")
	} else if err != testFuncErr {
		t.Errorf("expected testFuncErr but got %T (%s)", err, err)
	}

	err = UnmarshalRequestReflection([]byte{
		1, 0, 1, 0,
		7, 's', 't', 'r', 'i', 'n', 'g', 's', byte(requestWireTypeString), 4, 0, 't', 'e',
	}, func(id string, t Type) (reflect.Value, bool, error) {
		return values[2], true, nil
	})
	if err == nil {
		t.Errorf("expected *requestBufferUnderflowError but got %s", str)
	} else if _, ok := err.(*requestBufferUnderflowError); !ok {
		t.Errorf("expected *requestBufferUnderflowError but got %T (%s)", err, err)
	}

	err = UnmarshalRequestReflection([]byte{
		1, 0, 1, 0,
		7, 'i', 'n', 't', 'e', 'g', 'e', 'r', byte(requestWireTypeInteger), 1, 0, 0, 0,
	}, func(id string, t Type) (reflect.Value, bool, error) {
		return values[3], true, nil
	})
	if err == nil {
		t.Errorf("expected *requestBufferUnderflowError but got %s", str)
	} else if _, ok := err.(*requestBufferUnderflowError); !ok {
		t.Errorf("expected *requestBufferUnderflowError but got %T (%s)", err, err)
	}

	var i8 int8
	v := reflect.Indirect(reflect.ValueOf(&i8))

	err = UnmarshalRequestReflection([]byte{
		1, 0, 1, 0,
		7, 'i', 'n', 't', 'e', 'g', 'e', 'r', byte(requestWireTypeInteger), 128, 0, 0, 0, 0, 0, 0, 0,
	}, func(id string, t Type) (reflect.Value, bool, error) {
		return v, true, nil
	})
	if err == nil {
		t.Errorf("expected *requestUnmarshalIntegerOverflowError but got %s", str)
	} else if _, ok := err.(*requestUnmarshalIntegerOverflowError); !ok {
		t.Errorf("expected *requestUnmarshalIntegerOverflowError but got %T (%s)", err, err)
	}

	err = UnmarshalRequestReflection([]byte{
		1, 0, 1, 0,
		5, 'f', 'l', 'o', 'a', 't', byte(requestWireTypeFloat), 24, 45, 68, 84,
	}, func(id string, t Type) (reflect.Value, bool, error) {
		return values[4], true, nil
	})
	if err == nil {
		t.Errorf("expected *requestBufferUnderflowError but got %s", str)
	} else if _, ok := err.(*requestBufferUnderflowError); !ok {
		t.Errorf("expected *requestBufferUnderflowError but got %T (%s)", err, err)
	}

	err = UnmarshalRequestReflection([]byte{
		1, 0, 1, 0,
		8, 'a', 'd', 'd', 'r', 'e', 's', 's', '4', byte(requestWireTypeIPv4Address), 192, 0,
	}, func(id string, t Type) (reflect.Value, bool, error) {
		return values[5], true, nil
	})
	if err == nil {
		t.Errorf("expected *requestBufferUnderflowError but got %s", str)
	} else if _, ok := err.(*requestBufferUnderflowError); !ok {
		t.Errorf("expected *requestBufferUnderflowError but got %T (%s)", err, err)
	}

	err = UnmarshalRequestReflection([]byte{
		1, 0, 1, 0,
		8, 'a', 'd', 'd', 'r', 'e', 's', 's', '6', byte(requestWireTypeIPv6Address), 32, 1, 13, 184, 0, 0, 0, 0,
	}, func(id string, t Type) (reflect.Value, bool, error) {
		return values[6], true, nil
	})
	if err == nil {
		t.Errorf("expected *requestBufferUnderflowError but got %s", str)
	} else if _, ok := err.(*requestBufferUnderflowError); !ok {
		t.Errorf("expected *requestBufferUnderflowError but got %T (%s)", err, err)
	}

	err = UnmarshalRequestReflection([]byte{
		1, 0, 1, 0,
		8, 'n', 'e', 't', 'w', 'o', 'r', 'k', '4', byte(requestWireTypeIPv4Network), 24, 192, 0,
	}, func(id string, t Type) (reflect.Value, bool, error) {
		return values[7], true, nil
	})
	if err == nil {
		t.Errorf("expected *requestBufferUnderflowError but got %s", str)
	} else if _, ok := err.(*requestBufferUnderflowError); !ok {
		t.Errorf("expected *requestBufferUnderflowError but got %T (%s)", err, err)
	}

	err = UnmarshalRequestReflection([]byte{
		1, 0, 1, 0,
		8, 'n', 'e', 't', 'w', 'o', 'r', 'k', '6', byte(requestWireTypeIPv6Network), 32, 32, 1, 13, 184, 0, 0, 0, 0,
	}, func(id string, t Type) (reflect.Value, bool, error) {
		return values[8], true, nil
	})
	if err == nil {
		t.Errorf("expected *requestBufferUnderflowError but got %s", str)
	} else if _, ok := err.(*requestBufferUnderflowError); !ok {
		t.Errorf("expected *requestBufferUnderflowError but got %T (%s)", err, err)
	}

	err = UnmarshalRequestReflection([]byte{
		1, 0, 1, 0,
		6, 'd', 'o', 'm', 'a', 'i', 'n', byte(requestWireTypeDomain),
		15, 0, 'w', 'w', 'w', '.', 'e', 'x', 'a', 'm', 'p', 'l', 'e',
	}, func(id string, t Type) (reflect.Value, bool, error) {
		return values[9], true, nil
	})
	if err == nil {
		t.Errorf("expected *requestBufferUnderflowError but got %s", str)
	} else if _, ok := err.(*requestBufferUnderflowError); !ok {
		t.Errorf("expected *requestBufferUnderflowError but got %T (%s)", err, err)
	}
}

func TestCheckRequestVersion(t *testing.T) {
	n, err := checkRequestVersion(testWireRequest)
	if err != nil {
		t.Error(err)
	} else if n <= 0 {
		t.Errorf("expected some bytes consumed but got %d", n)
	} else if n > len(testWireRequest) {
		t.Errorf("not expected more bytes consumed (%d) than buffer has (%d)", n, len(testWireRequest))
	}

	_, err = checkRequestVersion([]byte{})
	if err == nil {
		t.Error("expected *requestBufferUnderflowError but got nothing")
	} else if _, ok := err.(*requestBufferUnderflowError); !ok {
		t.Errorf("expected *requestBufferUnderflowError but got %T (%s)", err, err)
	}

	_, err = checkRequestVersion([]byte{2, 0})
	if err == nil {
		t.Error("expected *requestVersionError but got nothing")
	} else if _, ok := err.(*requestVersionError); !ok {
		t.Errorf("expected *requestVersionError but got %T (%s)", err, err)
	}
}

func TestGetRequestAttributeCount(t *testing.T) {
	off, err := checkRequestVersion(testWireRequest)
	if err != nil {
		t.Fatal(err)
	}

	c, n, err := getRequestAttributeCount(testWireRequest[off:])
	if err != nil {
		t.Error(err)
	} else if n <= 0 {
		t.Errorf("expected some bytes consumed but got %d", n)
	} else if n > len(testWireRequest) {
		t.Errorf("not expected more bytes consumed (%d) than buffer has (%d)", n, len(testWireRequest))
	} else if c != 3 {
		t.Errorf("expected %d as attribute count but got %d", 1, c)
	}

	c, _, err = getRequestAttributeCount([]byte{})
	if err == nil {
		t.Errorf("expected *requestBufferUnderflowError but got count %d", c)
	} else if _, ok := err.(*requestBufferUnderflowError); !ok {
		t.Errorf("expected *requestBufferUnderflowError but got %T (%s)", err, err)
	}
}

func TestGetRequestAttributeName(t *testing.T) {
	off, err := checkRequestVersion(testWireRequest)
	if err != nil {
		t.Fatal(err)
	}

	c, n, err := getRequestAttributeCount(testWireRequest[off:])
	if err != nil {
		t.Fatal(err)
	} else if c != 3 {
		t.Fatalf("expected %d as attribute count but got %d", 1, c)
	}

	off += n

	name, n, err := getRequestAttributeName(testWireRequest[off:])
	if err != nil {
		t.Error(err)
	} else if n <= 0 {
		t.Errorf("expected some bytes consumed but got %d", n)
	} else if n > len(testWireRequest) {
		t.Errorf("not expected more bytes consumed (%d) than buffer has (%d)", n, len(testWireRequest))
	} else if name != "string" {
		t.Errorf("expected %q as attribute name but got %q", "test", name)
	}

	name, _, err = getRequestAttributeName([]byte{})
	if err == nil {
		t.Errorf("expected *requestBufferUnderflowError but got name %q", name)
	} else if _, ok := err.(*requestBufferUnderflowError); !ok {
		t.Errorf("expected *requestBufferUnderflowError but got %T (%s)", err, err)
	}

	name, _, err = getRequestAttributeName([]byte{4, 't', 'e'})
	if err == nil {
		t.Errorf("expected *requestBufferUnderflowError but got name %q", name)
	} else if _, ok := err.(*requestBufferUnderflowError); !ok {
		t.Errorf("expected *requestBufferUnderflowError but got %T (%s)", err, err)
	}
}

func TestGetRequestAttributeType(t *testing.T) {
	off, err := checkRequestVersion(testWireRequest)
	if err != nil {
		t.Fatal(err)
	}

	c, n, err := getRequestAttributeCount(testWireRequest[off:])
	if err != nil {
		t.Fatal(err)
	} else if c != 3 {
		t.Fatalf("expected %d as attribute count but got %d", 1, c)
	}

	off += n

	name, n, err := getRequestAttributeName(testWireRequest[off:])
	if err != nil {
		t.Fatal(err)
	} else if name != "string" {
		t.Fatalf("expected %q as attribute name but got %q", "test", name)
	}

	off += n

	at, n, err := getRequestAttributeType(testWireRequest[off:])
	if err != nil {
		t.Error(err)
	} else if n <= 0 {
		t.Errorf("expected some bytes consumed but got %d", n)
	} else if n > len(testWireRequest) {
		t.Errorf("not expected more bytes consumed (%d) than buffer has (%d)", n, len(testWireRequest))
	} else if at != requestWireTypeString {
		tn := "unknown"
		if at >= 0 || at < len(requestWireTypeNames) {
			tn = requestWireTypeNames[at]
		}

		t.Errorf("expected %q (%d) as attribute type but got %q (%d)",
			requestWireTypeNames[requestWireTypeString], requestWireTypeString, tn, at)
	}

	at, _, err = getRequestAttributeType([]byte{})
	if err == nil {
		tn := "unknown"
		if at >= 0 || at < len(requestWireTypeNames) {
			tn = requestWireTypeNames[at]
		}

		t.Errorf("expected *requestBufferUnderflowError but got type %q (%d)", tn, at)
	} else if _, ok := err.(*requestBufferUnderflowError); !ok {
		t.Errorf("expected *requestBufferUnderflowError but got %T (%s)", err, err)
	}
}

func TestGetRequestStringValue(t *testing.T) {
	off, err := checkRequestVersion(testWireRequest)
	if err != nil {
		t.Fatal(err)
	}

	c, n, err := getRequestAttributeCount(testWireRequest[off:])
	if err != nil {
		t.Fatal(err)
	} else if c != 3 {
		t.Fatalf("expected %d as attribute count but got %d", 1, c)
	}

	off += n

	name, n, err := getRequestAttributeName(testWireRequest[off:])
	if err != nil {
		t.Fatal(err)
	} else if name != "string" {
		t.Fatalf("expected %q as attribute name but got %q", "test", name)
	}

	off += n

	at, n, err := getRequestAttributeType(testWireRequest[off:])
	if err != nil {
		t.Fatal(err)
	} else if at != requestWireTypeString {
		tn := "unknown"
		if at >= 0 || at < len(requestWireTypeNames) {
			tn = requestWireTypeNames[at]
		}

		t.Fatalf("expected %q (%d) as attribute type but got %q (%d)",
			requestWireTypeNames[requestWireTypeString], requestWireTypeString, tn, at)
	}

	off += n

	v, n, err := getRequestStringValue(testWireRequest[off:])
	if err != nil {
		t.Error(err)
	} else if n <= 0 {
		t.Errorf("expected some bytes consumed but got %d", n)
	} else if off+n > len(testWireRequest) {
		t.Errorf("not expected more bytes consumed (%d) than buffer has (%d)", n, len(testWireRequest))
	} else if v != "test" {
		t.Errorf("expected string %q as attribute value but got %q", "test value", v)
	}

	v, _, err = getRequestStringValue([]byte{})
	if err == nil {
		t.Errorf("expected *requestBufferUnderflowError but got string %q", v)
	} else if _, ok := err.(*requestBufferUnderflowError); !ok {
		t.Errorf("expected *requestBufferUnderflowError but got %T (%s)", err, err)
	}

	v, _, err = getRequestStringValue([]byte{10, 0, 't', 'e', 's', 't'})
	if err == nil {
		t.Errorf("expected *requestBufferUnderflowError but got string %q", v)
	} else if _, ok := err.(*requestBufferUnderflowError); !ok {
		t.Errorf("expected *requestBufferUnderflowError but got %T (%s)", err, err)
	}
}

func TestGetRequestIntegerValue(t *testing.T) {
	testWireIntegerValue := []byte{
		0, 0, 0, 0, 0, 0, 0, 128,
	}
	v, n, err := getRequestIntegerValue(testWireIntegerValue)
	if err != nil {
		t.Error(err)
	} else if n != len(testWireIntegerValue) {
		t.Errorf("expected whole buffer consumed (%d) but got (%d)", len(testWireIntegerValue), n)
	} else if v != -9223372036854775808 {
		t.Errorf("expected integer %d as attribute value but got %d", -9223372036854775808, v)
	}

	v, _, err = getRequestIntegerValue([]byte{})
	if err == nil {
		t.Errorf("expected *requestBufferUnderflowError but got integer %d", v)
	} else if _, ok := err.(*requestBufferUnderflowError); !ok {
		t.Errorf("expected *requestBufferUnderflowError but got %T (%s)", err, err)
	}
}

func TestGetRequestFloatValue(t *testing.T) {
	testWireFloatValue := []byte{
		24, 45, 68, 84, 251, 33, 9, 64,
	}
	v, n, err := getRequestFloatValue(testWireFloatValue)
	if err != nil {
		t.Error(err)
	} else if n != len(testWireFloatValue) {
		t.Errorf("expected whole buffer consumed (%d) but got (%d)", len(testWireFloatValue), n)
	} else if v != float64(math.Pi) {
		t.Errorf("expected float %g as attribute value but got %g", float64(math.Pi), v)
	}

	v, _, err = getRequestFloatValue([]byte{})
	if err == nil {
		t.Errorf("expected *requestBufferUnderflowError but got float %g", v)
	} else if _, ok := err.(*requestBufferUnderflowError); !ok {
		t.Errorf("expected *requestBufferUnderflowError but got %T (%s)", err, err)
	}
}

func TestGetRequestIPv4AddressValue(t *testing.T) {
	testWireIPv4AddressValue := []byte{
		192, 0, 2, 1,
	}
	v, n, err := getRequestIPv4AddressValue(testWireIPv4AddressValue)
	if err != nil {
		t.Error(err)
	} else if n != len(testWireIPv4AddressValue) {
		t.Errorf("expected whole buffer consumed (%d) but got (%d)", len(testWireIPv4AddressValue), n)
	} else if !v.Equal(net.ParseIP("192.0.2.1")) {
		t.Errorf("expected IPv4 address %q as attribute value but got %q", net.ParseIP("192.0.2.1"), v)
	}

	v, _, err = getRequestIPv4AddressValue([]byte{})
	if err == nil {
		t.Errorf("expected *requestBufferUnderflowError but got IPv4 address %q", v)
	} else if _, ok := err.(*requestBufferUnderflowError); !ok {
		t.Errorf("expected *requestBufferUnderflowError but got %T (%s)", err, err)
	}
}

func TestGetRequestIPv6AddressValue(t *testing.T) {
	testWireIPv6AddressValue := []byte{
		32, 1, 13, 184, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1,
	}
	v, n, err := getRequestIPv6AddressValue(testWireIPv6AddressValue)
	if err != nil {
		t.Error(err)
	} else if n != len(testWireIPv6AddressValue) {
		t.Errorf("expected whole buffer consumed (%d) but got (%d)", len(testWireIPv6AddressValue), n)
	} else if !v.Equal(net.ParseIP("2001:db8::1")) {
		t.Errorf("expected IPv6 address %q as attribute value but got %q", net.ParseIP("2001:db8::1"), v)
	}

	v, _, err = getRequestIPv6AddressValue([]byte{})
	if err == nil {
		t.Errorf("expected *requestBufferUnderflowError but got IPv6 address %q", v)
	} else if _, ok := err.(*requestBufferUnderflowError); !ok {
		t.Errorf("expected *requestBufferUnderflowError but got %T (%s)", err, err)
	}
}

func TestGetRequestIPv4NetworkValue(t *testing.T) {
	testWireIPv4NetworkValue := []byte{
		24, 192, 0, 2, 1,
	}
	v, n, err := getRequestIPv4NetworkValue(testWireIPv4NetworkValue)
	if err != nil {
		t.Error(err)
	} else if n != len(testWireIPv4NetworkValue) {
		t.Errorf("expected whole buffer consumed (%d) but got (%d)", len(testWireIPv4NetworkValue), n)
	} else if v.String() != "192.0.2.0/24" {
		t.Errorf("expected IPv4 network %q as attribute value but got %q", "192.0.2.0/24", v)
	}

	v, _, err = getRequestIPv4NetworkValue([]byte{})
	if err == nil {
		t.Errorf("expected *requestBufferUnderflowError but got IPv4 network %q", v)
	} else if _, ok := err.(*requestBufferUnderflowError); !ok {
		t.Errorf("expected *requestBufferUnderflowError but got %T (%s)", err, err)
	}

	v, _, err = getRequestIPv4NetworkValue([]byte{
		255, 192, 0, 2, 1,
	})
	if err == nil {
		t.Errorf("expected *requestIPv4InvalidMaskError but got IPv4 network %q", v)
	} else if _, ok := err.(*requestIPv4InvalidMaskError); !ok {
		t.Errorf("expected *requestIPv4InvalidMaskError but got %T (%s)", err, err)
	}
}

func TestGetRequestIPv6NetworkValue(t *testing.T) {
	testWireIPv6NetworkValue := []byte{
		32, 32, 1, 13, 184, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1,
	}
	v, n, err := getRequestIPv6NetworkValue(testWireIPv6NetworkValue)
	if err != nil {
		t.Error(err)
	} else if n != len(testWireIPv6NetworkValue) {
		t.Errorf("expected whole buffer consumed (%d) but got (%d)", len(testWireIPv6NetworkValue), n)
	} else if v.String() != "2001:db8::/32" {
		t.Errorf("expected IPv6 network %q as attribute value but got %q", "2001:db8::/32", v)
	}

	v, _, err = getRequestIPv6NetworkValue([]byte{})
	if err == nil {
		t.Errorf("expected *requestBufferUnderflowError but got IPv6 network %q", v)
	} else if _, ok := err.(*requestBufferUnderflowError); !ok {
		t.Errorf("expected *requestBufferUnderflowError but got %T (%s)", err, err)
	}

	v, _, err = getRequestIPv6NetworkValue([]byte{
		255, 32, 1, 13, 184, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1,
	})
	if err == nil {
		t.Errorf("expected *requestIPv6InvalidMaskError but got IPv6 network %q", v)
	} else if _, ok := err.(*requestIPv6InvalidMaskError); !ok {
		t.Errorf("expected *requestIPv6InvalidMaskError but got %T (%s)", err, err)
	}
}

func TestGetRequestDomainValue(t *testing.T) {
	testWireDomainValue := []byte{
		8, 0, 't', 'e', 's', 't', '.', 'c', 'o', 'm',
	}
	v, n, err := getRequestDomainValue(testWireDomainValue)
	if err != nil {
		t.Error(err)
	} else if n != len(testWireDomainValue) {
		t.Errorf("expected whole buffer consumed (%d) but got (%d)", len(testWireDomainValue), n)
	} else if v.String() != "test.com" {
		t.Errorf("expected domain %q as attribute value but got %q", "test.com", v)
	}

	v, _, err = getRequestDomainValue([]byte{})
	if err == nil {
		t.Errorf("expected *requestBufferUnderflowError but got domain %q", v)
	} else if _, ok := err.(*requestBufferUnderflowError); !ok {
		t.Errorf("expected *requestBufferUnderflowError but got %T (%s)", err, err)
	}

	v, _, err = getRequestDomainValue([]byte{
		3, 0, '.', '.', '.',
	})
	if err == nil {
		t.Errorf("expected domain.ErrEmptyLabel error but got domain %q", v)
	} else if err != domain.ErrEmptyLabel {
		t.Errorf("expected domain.ErrEmptyLabel error but got %T (%s)", err, err)
	}
}

func TestGetRequestAttribute(t *testing.T) {
	testWireBooleanFalseAttribute := []byte{
		2, 'n', 'o', byte(requestWireTypeBooleanFalse),
	}

	name, v, n, err := getRequestAttribute(testWireBooleanFalseAttribute)
	if err != nil {
		t.Error(err)
	} else if n != len(testWireBooleanFalseAttribute) {
		t.Errorf("expected whole buffer consumed (%d) but got (%d)", len(testWireBooleanFalseAttribute), n)
	} else if name != "no" {
		t.Errorf("expected %q as attribute name but got %q", "no", name)
	} else if vt := v.GetResultType(); vt != TypeBoolean {
		t.Errorf("expected value of %q type but got %q %s", TypeBoolean, vt, v.describe())
	} else {
		s, err := v.Serialize()
		if err != nil {
			t.Error(err)
		} else {
			e := "false"
			if s != e {
				t.Errorf("expected %q but got %q", e, s)
			}
		}
	}

	testWireBooleanTrueAttribute := []byte{
		3, 'y', 'e', 's', byte(requestWireTypeBooleanTrue),
	}

	name, v, n, err = getRequestAttribute(testWireBooleanTrueAttribute)
	if err != nil {
		t.Error(err)
	} else if n != len(testWireBooleanTrueAttribute) {
		t.Errorf("expected whole buffer consumed (%d) but got (%d)", len(testWireBooleanTrueAttribute), n)
	} else if name != "yes" {
		t.Errorf("expected %q as attribute name but got %q", "yes", name)
	} else if vt := v.GetResultType(); vt != TypeBoolean {
		t.Errorf("expected value of %q type but got %q %s", TypeBoolean, vt, v.describe())
	} else {
		s, err := v.Serialize()
		if err != nil {
			t.Error(err)
		} else {
			e := "true"
			if s != e {
				t.Errorf("expected %q but got %q", e, s)
			}
		}
	}

	testWireStringAttribute := []byte{
		6, 's', 't', 'r', 'i', 'n', 'g', byte(requestWireTypeString), 4, 0, 't', 'e', 's', 't',
	}

	name, v, n, err = getRequestAttribute(testWireStringAttribute)
	if err != nil {
		t.Error(err)
	} else if n != len(testWireStringAttribute) {
		t.Errorf("expected whole buffer consumed (%d) but got (%d)", len(testWireStringAttribute), n)
	} else if name != "string" {
		t.Errorf("expected %q as attribute name but got %q", "string", name)
	} else if vt := v.GetResultType(); vt != TypeString {
		t.Errorf("expected value of %q type but got %q %s", TypeString, vt, v.describe())
	} else {
		s, err := v.Serialize()
		if err != nil {
			t.Error(err)
		} else {
			e := "test"
			if s != e {
				t.Errorf("expected %q but got %q", e, s)
			}
		}
	}

	testWireIntegerAttribute := []byte{
		7, 'i', 'n', 't', 'e', 'g', 'e', 'r', byte(requestWireTypeInteger), 0, 0, 0, 0, 0, 0, 0, 128,
	}

	name, v, n, err = getRequestAttribute(testWireIntegerAttribute)
	if err != nil {
		t.Error(err)
	} else if n != len(testWireIntegerAttribute) {
		t.Errorf("expected whole buffer consumed (%d) but got (%d)", len(testWireIntegerAttribute), n)
	} else if name != "integer" {
		t.Errorf("expected %q as attribute name but got %q", "integer", name)
	} else if vt := v.GetResultType(); vt != TypeInteger {
		t.Errorf("expected value of %q type but got %q %s", TypeInteger, vt, v.describe())
	} else {
		s, err := v.Serialize()
		if err != nil {
			t.Error(err)
		} else {
			e := "-9223372036854775808"
			if s != e {
				t.Errorf("expected %q but got %q", e, s)
			}
		}
	}

	testWireFloatAttribute := []byte{
		5, 'f', 'l', 'o', 'a', 't', byte(requestWireTypeFloat), 24, 45, 68, 84, 251, 33, 9, 64,
	}

	name, v, n, err = getRequestAttribute(testWireFloatAttribute)
	if err != nil {
		t.Error(err)
	} else if n != len(testWireFloatAttribute) {
		t.Errorf("expected whole buffer consumed (%d) but got (%d)", len(testWireFloatAttribute), n)
	} else if name != "float" {
		t.Errorf("expected %q as attribute name but got %q", "float", name)
	} else if vt := v.GetResultType(); vt != TypeFloat {
		t.Errorf("expected value of %q type but got %q %s", TypeFloat, vt, v.describe())
	} else {
		s, err := v.Serialize()
		if err != nil {
			t.Error(err)
		} else {
			e := "3.141592653589793"
			if s != e {
				t.Errorf("expected %q but got %q", e, s)
			}
		}
	}

	testWireIPv4AddressAttribute := []byte{
		4, 'I', 'P', 'v', '4', byte(requestWireTypeIPv4Address), 192, 0, 2, 1,
	}

	name, v, n, err = getRequestAttribute(testWireIPv4AddressAttribute)
	if err != nil {
		t.Error(err)
	} else if n != len(testWireIPv4AddressAttribute) {
		t.Errorf("expected whole buffer consumed (%d) but got (%d)", len(testWireIPv4AddressAttribute), n)
	} else if name != "IPv4" {
		t.Errorf("expected %q as attribute name but got %q", "IPv4", name)
	} else if vt := v.GetResultType(); vt != TypeAddress {
		t.Errorf("expected value of %q type but got %q %s", TypeAddress, vt, v.describe())
	} else {
		s, err := v.Serialize()
		if err != nil {
			t.Error(err)
		} else {
			e := "192.0.2.1"
			if s != e {
				t.Errorf("expected %q but got %q", e, s)
			}
		}
	}

	testWireIPv6AddressAttribute := []byte{
		4, 'I', 'P', 'v', '6', byte(requestWireTypeIPv6Address), 32, 1, 13, 184, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1,
	}

	name, v, n, err = getRequestAttribute(testWireIPv6AddressAttribute)
	if err != nil {
		t.Error(err)
	} else if n != len(testWireIPv6AddressAttribute) {
		t.Errorf("expected whole buffer consumed (%d) but got (%d)", len(testWireIPv6AddressAttribute), n)
	} else if name != "IPv6" {
		t.Errorf("expected %q as attribute name but got %q", "IPv6", name)
	} else if vt := v.GetResultType(); vt != TypeAddress {
		t.Errorf("expected value of %q type but got %q %s", TypeAddress, vt, v.describe())
	} else {
		s, err := v.Serialize()
		if err != nil {
			t.Error(err)
		} else {
			e := "2001:db8::1"
			if s != e {
				t.Errorf("expected %q but got %q", e, s)
			}
		}
	}

	testWireIPv4NetworkAttribute := []byte{
		11, 'I', 'P', 'v', '4', 'N', 'e', 't', 'w', 'o', 'r', 'k', byte(requestWireTypeIPv4Network), 24, 192, 0, 2, 1,
	}

	name, v, n, err = getRequestAttribute(testWireIPv4NetworkAttribute)
	if err != nil {
		t.Error(err)
	} else if n != len(testWireIPv4NetworkAttribute) {
		t.Errorf("expected whole buffer consumed (%d) but got (%d)", len(testWireIPv4NetworkAttribute), n)
	} else if name != "IPv4Network" {
		t.Errorf("expected %q as attribute name but got %q", "IPv4Network", name)
	} else if vt := v.GetResultType(); vt != TypeNetwork {
		t.Errorf("expected value of %q type but got %q %s", TypeNetwork, vt, v.describe())
	} else {
		s, err := v.Serialize()
		if err != nil {
			t.Error(err)
		} else {
			e := "192.0.2.0/24"
			if s != e {
				t.Errorf("expected %q but got %q", e, s)
			}
		}
	}

	testWireIPv6NetworkAttribute := []byte{
		11, 'I', 'P', 'v', '6', 'N', 'e', 't', 'w', 'o', 'r', 'k', byte(requestWireTypeIPv6Network),
		32, 32, 1, 13, 184, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1,
	}

	name, v, n, err = getRequestAttribute(testWireIPv6NetworkAttribute)
	if err != nil {
		t.Error(err)
	} else if n != len(testWireIPv6NetworkAttribute) {
		t.Errorf("expected whole buffer consumed (%d) but got (%d)", len(testWireIPv6NetworkAttribute), n)
	} else if name != "IPv6Network" {
		t.Errorf("expected %q as attribute name but got %q", "IPv6Network", name)
	} else if vt := v.GetResultType(); vt != TypeNetwork {
		t.Errorf("expected value of %q type but got %q %s", TypeNetwork, vt, v.describe())
	} else {
		s, err := v.Serialize()
		if err != nil {
			t.Error(err)
		} else {
			e := "2001:db8::/32"
			if s != e {
				t.Errorf("expected %q but got %q", e, s)
			}
		}
	}

	testWireDomainAttribute := []byte{
		6, 'd', 'o', 'm', 'a', 'i', 'n', byte(requestWireTypeDomain), 8, 0, 't', 'e', 's', 't', '.', 'c', 'o', 'm',
	}

	name, v, n, err = getRequestAttribute(testWireDomainAttribute)
	if err != nil {
		t.Error(err)
	} else if n != len(testWireDomainAttribute) {
		t.Errorf("expected whole buffer consumed (%d) but got (%d)", len(testWireDomainAttribute), n)
	} else if name != "domain" {
		t.Errorf("expected %q as attribute name but got %q", "domain", name)
	} else if vt := v.GetResultType(); vt != TypeDomain {
		t.Errorf("expected value of %q type but got %q %s", TypeDomain, vt, v.describe())
	} else {
		s, err := v.Serialize()
		if err != nil {
			t.Error(err)
		} else {
			e := "test.com"
			if s != e {
				t.Errorf("expected %q but got %q", e, s)
			}
		}
	}

	name, v, _, err = getRequestAttribute([]byte{})
	if err == nil {
		t.Errorf("expected *requestBufferUnderflowError but got attribute %q = %s", name, v.describe())
	} else if _, ok := err.(*requestBufferUnderflowError); !ok {
		t.Errorf("expected *requestBufferUnderflowError but got %T (%s)", err, err)
	}

	name, v, _, err = getRequestAttribute([]byte{
		6, 'n', 'o', 't', 'y', 'p', 'e',
	})
	if err == nil {
		t.Errorf("expected *requestBufferUnderflowError but got attribute %q = %s", name, v.describe())
	} else if _, ok := err.(*requestBufferUnderflowError); !ok {
		t.Errorf("expected *requestBufferUnderflowError but got %T (%s)", err, err)
	}

	name, v, _, err = getRequestAttribute([]byte{
		7, 'u', 'n', 'k', 'n', 'o', 'w', 'n', 255,
	})
	if err == nil {
		t.Errorf("expected *requestAttributeUnmarshallingTypeError but got attribute %q = %s", name, v.describe())
	} else if _, ok := err.(*requestAttributeUnmarshallingTypeError); !ok {
		t.Errorf("expected *requestAttributeUnmarshallingTypeError but got %T (%s)", err, err)
	}

	name, v, _, err = getRequestAttribute([]byte{
		7, 'n', 'o', 't', 'i', 'm', 'p', 'l', byte(requestWireTypeListOfStrings),
	})
	if err == nil {
		t.Errorf("expected *requestAttributeUnmarshallingNotImplemented error but got attribute %q = %s",
			name, v.describe())
	} else if _, ok := err.(*requestAttributeUnmarshallingNotImplemented); !ok {
		t.Errorf("expected *requestAttributeUnmarshallingNotImplemented error but got %T (%s)", err, err)
	}

	name, v, _, err = getRequestAttribute([]byte{
		6, 's', 't', 'r', 'i', 'n', 'g', byte(requestWireTypeString), 4, 0, 't', 'e',
	})
	if err == nil {
		t.Errorf("expected *requestBufferUnderflowError but got attribute %q = %s", name, v.describe())
	} else if _, ok := err.(*requestBufferUnderflowError); !ok {
		t.Errorf("expected *requestBufferUnderflowError but got %T (%s)", err, err)
	}

	name, v, _, err = getRequestAttribute([]byte{
		7, 'i', 'n', 't', 'e', 'g', 'e', 'r', byte(requestWireTypeInteger), 0, 0, 0, 0,
	})
	if err == nil {
		t.Errorf("expected *requestBufferUnderflowError but got attribute %q = %s", name, v.describe())
	} else if _, ok := err.(*requestBufferUnderflowError); !ok {
		t.Errorf("expected *requestBufferUnderflowError but got %T (%s)", err, err)
	}

	name, v, _, err = getRequestAttribute([]byte{
		5, 'f', 'l', 'o', 'a', 't', byte(requestWireTypeFloat), 24, 45, 68, 84,
	})
	if err == nil {
		t.Errorf("expected *requestBufferUnderflowError but got attribute %q = %s", name, v.describe())
	} else if _, ok := err.(*requestBufferUnderflowError); !ok {
		t.Errorf("expected *requestBufferUnderflowError but got %T (%s)", err, err)
	}

	name, v, _, err = getRequestAttribute([]byte{
		4, 'I', 'P', 'v', '4', byte(requestWireTypeIPv4Address), 192, 0,
	})
	if err == nil {
		t.Errorf("expected *requestBufferUnderflowError but got attribute %q = %s", name, v.describe())
	} else if _, ok := err.(*requestBufferUnderflowError); !ok {
		t.Errorf("expected *requestBufferUnderflowError but got %T (%s)", err, err)
	}

	name, v, _, err = getRequestAttribute([]byte{
		4, 'I', 'P', 'v', '6', byte(requestWireTypeIPv6Address), 32, 1, 13, 184, 0, 0, 0, 0,
	})
	if err == nil {
		t.Errorf("expected *requestBufferUnderflowError but got attribute %q = %s", name, v.describe())
	} else if _, ok := err.(*requestBufferUnderflowError); !ok {
		t.Errorf("expected *requestBufferUnderflowError but got %T (%s)", err, err)
	}

	name, v, _, err = getRequestAttribute([]byte{
		11, 'I', 'P', 'v', '4', 'N', 'e', 't', 'w', 'o', 'r', 'k', byte(requestWireTypeIPv4Network), 192, 0, 2, 1,
	})
	if err == nil {
		t.Errorf("expected *requestBufferUnderflowError but got attribute %q = %s", name, v.describe())
	} else if _, ok := err.(*requestBufferUnderflowError); !ok {
		t.Errorf("expected *requestBufferUnderflowError but got %T (%s)", err, err)
	}

	name, v, _, err = getRequestAttribute([]byte{
		11, 'I', 'P', 'v', '6', 'N', 'e', 't', 'w', 'o', 'r', 'k', byte(requestWireTypeIPv6Network),
		32, 1, 13, 184, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1,
	})
	if err == nil {
		t.Errorf("expected *requestBufferUnderflowError but got attribute %q = %s", name, v.describe())
	} else if _, ok := err.(*requestBufferUnderflowError); !ok {
		t.Errorf("expected *requestBufferUnderflowError but got %T (%s)", err, err)
	}

	name, v, _, err = getRequestAttribute([]byte{
		6, 'd', 'o', 'm', 'a', 'i', 'n', byte(requestWireTypeDomain), 8, 0, 't', 'e', 's', 't',
	})
	if err == nil {
		t.Errorf("expected *requestBufferUnderflowError but got attribute %q = %s", name, v.describe())
	} else if _, ok := err.(*requestBufferUnderflowError); !ok {
		t.Errorf("expected *requestBufferUnderflowError but got %T (%s)", err, err)
	}
}

func TestPutRequestVersion(t *testing.T) {
	var b [2]byte

	n, err := putRequestVersion(b[:])
	assertRequestBytesBuffer(t, "putRequestVersion", err, b[:], n, 1, 0)

	n, err = putRequestVersion(nil)
	assertRequestBufferOverflow(t, "putRequestVersion", err, n)
}

func TestPutRequestAttributeCount(t *testing.T) {
	var b [2]byte

	n, err := putRequestAttributeCount(b[:], 0)
	assertRequestBytesBuffer(t, "putRequestAttributeCount", err, b[:], n, 0, 0)

	n, err = putRequestAttributeCount(nil, 0)
	assertRequestBufferOverflow(t, "putRequestAttributeCount", err, n)

	n, err = putRequestAttributeCount(b[:], -1)
	if err == nil {
		t.Errorf("expected *requestInvalidAttributeCountError for negative count but got %d bytes in buffer", n)
	} else if _, ok := err.(*requestInvalidAttributeCountError); !ok {
		t.Errorf("expected *requestInvalidAttributeCountError but got %T (%s)", err, err)
	}

	n, err = putRequestAttributeCount(b[:], math.MaxUint16+1)
	if err == nil {
		t.Errorf("expected *requestTooManyAttributesError for large count but got %d bytes in buffer", n)
	} else if _, ok := err.(*requestTooManyAttributesError); !ok {
		t.Errorf("expected *requestTooManyAttributesError but got %T (%s)", err, err)
	}
}

func TestPutRequestAttributeName(t *testing.T) {
	var b [5]byte

	n, err := putRequestAttributeName(b[:], "test")
	assertRequestBytesBuffer(t, "putRequestAttributeName", err, b[:], n, 4, 't', 'e', 's', 't')

	n, err = putRequestAttributeName(nil, "test")
	assertRequestBufferOverflow(t, "putRequestAttributeName", err, n)

	n, err = putRequestAttributeName(b[:],
		"01234567890123456789012345678901234567890123456789012345678901234567890123456789"+
			"01234567890123456789012345678901234567890123456789012345678901234567890123456789"+
			"01234567890123456789012345678901234567890123456789012345678901234567890123456789"+
			"0123456789012345",
	)
	if err == nil {
		t.Errorf("expected *requestTooLongAttributeNameError for long name but got %d bytes in buffer", n)
	} else if _, ok := err.(*requestTooLongAttributeNameError); !ok {
		t.Errorf("expected *requestTooLongAttributeNameError but got %T (%s)", err, err)
	}
}

func TestPutRequestAttributeType(t *testing.T) {
	var b [1]byte

	n, err := putRequestAttributeType(b[:], requestWireTypeString)
	assertRequestBytesBuffer(t, "putRequestAttributeType", err, b[:], n, byte(requestWireTypeString))

	n, err = putRequestAttributeType(nil, requestWireTypeString)
	assertRequestBufferOverflow(t, "putRequestAttributeType", err, n)

	n, err = putRequestAttributeType(b[:], 2147483647)
	if err == nil {
		t.Errorf("expected *requestAttributeMarshallingTypeError for long name but got %d bytes in buffer", n)
	} else if _, ok := err.(*requestAttributeMarshallingTypeError); !ok {
		t.Errorf("expected *requestAttributeMarshallingTypeError but got %T (%s)", err, err)
	}
}

func TestPutRequestAttribute(t *testing.T) {
	var b [25]byte

	n, err := putRequestAttribute(b[:9], "boolean", MakeBooleanValue(true))
	assertRequestBytesBuffer(t, "putRequestAttribute(boolean)", err, b[:9], n,
		7, 'b', 'o', 'o', 'l', 'e', 'a', 'n', byte(requestWireTypeBooleanTrue),
	)

	n, err = putRequestAttribute(b[:14], "string", MakeStringValue("test"))
	assertRequestBytesBuffer(t, "putRequestAttribute(string)", err, b[:14], n,
		6, 's', 't', 'r', 'i', 'n', 'g', byte(requestWireTypeString), 4, 0, 't', 'e', 's', 't',
	)

	n, err = putRequestAttribute(b[:17], "integer", MakeIntegerValue(-9223372036854775808))
	assertRequestBytesBuffer(t, "putRequestAttribute(integer)", err, b[:17], n,
		7, 'i', 'n', 't', 'e', 'g', 'e', 'r', byte(requestWireTypeInteger), 0, 0, 0, 0, 0, 0, 0, 0x80,
	)

	n, err = putRequestAttribute(b[:15], "float", MakeFloatValue(math.Pi))
	assertRequestBytesBuffer(t, "putRequestAttribute(float)", err, b[:15], n,
		5, 'f', 'l', 'o', 'a', 't', byte(requestWireTypeFloat), 24, 45, 68, 84, 251, 33, 9, 64,
	)

	n, err = putRequestAttribute(b[:13], "address", MakeAddressValue(net.ParseIP("192.0.2.1")))
	assertRequestBytesBuffer(t, "putRequestAttribute(address)", err, b[:13], n,
		7, 'a', 'd', 'd', 'r', 'e', 's', 's', byte(requestWireTypeIPv4Address), 192, 0, 2, 1,
	)

	n, err = putRequestAttribute(b[:14], "network", MakeNetworkValue(makeTestNetwork("192.0.2.0/24")))
	assertRequestBytesBuffer(t, "putRequestAttribute(network)", err, b[:14], n,
		7, 'n', 'e', 't', 'w', 'o', 'r', 'k', byte(requestWireTypeIPv4Network), 24, 192, 0, 2, 0,
	)

	n, err = putRequestAttribute(b[:25], "domain", MakeDomainValue(makeTestDomain("www.example.com")))
	assertRequestBytesBuffer(t, "putRequestAttribute(domain)", err, b[:25], n,
		6, 'd', 'o', 'm', 'a', 'i', 'n', byte(requestWireTypeDomain),
		15, 0, 'w', 'w', 'w', '.', 'e', 'x', 'a', 'm', 'p', 'l', 'e', '.', 'c', 'o', 'm',
	)

	n, err = putRequestAttribute(b[:], "undefined", UndefinedValue)
	if err == nil {
		t.Errorf("expected no data put to buffer for undefined value but got %d", n)
	} else if _, ok := err.(*requestAttributeMarshallingNotImplemented); !ok {
		t.Errorf("expected *requestAttributeMarshallingNotImplemented but got %T (%s)", err, err)
	}
}

func TestPutRequestAttributeBoolean(t *testing.T) {
	var b [9]byte

	n, err := putRequestAttributeBoolean(b[:], "boolean", true)
	assertRequestBytesBuffer(t, "putRequestAttributeBoolean", err, b[:], n,
		7, 'b', 'o', 'o', 'l', 'e', 'a', 'n', byte(requestWireTypeBooleanTrue),
	)

	n, err = putRequestAttributeBoolean(b[:5], "boolean", true)
	if err == nil {
		t.Errorf("expected no data put to small buffer but got %d", n)
	} else if _, ok := err.(*requestBufferOverflowError); !ok {
		t.Errorf("expected *requestBufferOverflowError but got %T (%s)", err, err)
	}

	n, err = putRequestAttributeBoolean(b[:8], "boolean", true)
	if err == nil {
		t.Errorf("expected no data put to small buffer but got %d", n)
	} else if _, ok := err.(*requestBufferOverflowError); !ok {
		t.Errorf("expected *requestBufferOverflowError but got %T (%s)", err, err)
	}
}

func TestPutRequestAttributeString(t *testing.T) {
	var b [14]byte

	n, err := putRequestAttributeString(b[:], "string", "test")
	assertRequestBytesBuffer(t, "putRequestAttributeString", err, b[:], n,
		6, 's', 't', 'r', 'i', 'n', 'g', byte(requestWireTypeString), 4, 0, 't', 'e', 's', 't',
	)

	n, err = putRequestAttributeString(b[:5], "string", "test")
	if err == nil {
		t.Errorf("expected no data put to small buffer but got %d", n)
	} else if _, ok := err.(*requestBufferOverflowError); !ok {
		t.Errorf("expected *requestBufferOverflowError but got %T (%s)", err, err)
	}

	n, err = putRequestAttributeString(b[:9], "string", "test")
	if err == nil {
		t.Errorf("expected no data put to small buffer but got %d", n)
	} else if _, ok := err.(*requestBufferOverflowError); !ok {
		t.Errorf("expected *requestBufferOverflowError but got %T (%s)", err, err)
	}
}

func TestPutRequestAttributeInteger(t *testing.T) {
	var b [17]byte

	n, err := putRequestAttributeInteger(b[:], "integer", -9223372036854775808)
	assertRequestBytesBuffer(t, "putRequestAttributeInteger", err, b[:], n,
		7, 'i', 'n', 't', 'e', 'g', 'e', 'r', byte(requestWireTypeInteger), 0, 0, 0, 0, 0, 0, 0, 0x80,
	)

	n, err = putRequestAttributeInteger(b[:5], "integer", 0)
	if err == nil {
		t.Errorf("expected no data put to small buffer but got %d", n)
	} else if _, ok := err.(*requestBufferOverflowError); !ok {
		t.Errorf("expected *requestBufferOverflowError but got %T (%s)", err, err)
	}

	n, err = putRequestAttributeInteger(b[:9], "integer", 0)
	if err == nil {
		t.Errorf("expected no data put to small buffer but got %d", n)
	} else if _, ok := err.(*requestBufferOverflowError); !ok {
		t.Errorf("expected *requestBufferOverflowError but got %T (%s)", err, err)
	}
}

func TestPutRequestAttributeFloat(t *testing.T) {
	var b [15]byte

	n, err := putRequestAttributeFloat(b[:], "float", math.Pi)
	assertRequestBytesBuffer(t, "putRequestAttributeFloat", err, b[:], n,
		5, 'f', 'l', 'o', 'a', 't', byte(requestWireTypeFloat), 24, 45, 68, 84, 251, 33, 9, 64,
	)

	n, err = putRequestAttributeFloat(b[:4], "float", 0)
	if err == nil {
		t.Errorf("expected no data put to small buffer but got %d", n)
	} else if _, ok := err.(*requestBufferOverflowError); !ok {
		t.Errorf("expected *requestBufferOverflowError but got %T (%s)", err, err)
	}

	n, err = putRequestAttributeFloat(b[:7], "float", 0)
	if err == nil {
		t.Errorf("expected no data put to small buffer but got %d", n)
	} else if _, ok := err.(*requestBufferOverflowError); !ok {
		t.Errorf("expected *requestBufferOverflowError but got %T (%s)", err, err)
	}
}

func TestPutRequestAttributeAddress(t *testing.T) {
	var b [13]byte

	n, err := putRequestAttributeAddress(b[:], "address", net.ParseIP("192.0.2.1"))
	assertRequestBytesBuffer(t, "putRequestAttributeFloat", err, b[:], n,
		7, 'a', 'd', 'd', 'r', 'e', 's', 's', byte(requestWireTypeIPv4Address), 192, 0, 2, 1,
	)

	n, err = putRequestAttributeAddress(b[:4], "address", net.ParseIP("192.0.2.1"))
	if err == nil {
		t.Errorf("expected no data put to small buffer but got %d", n)
	} else if _, ok := err.(*requestBufferOverflowError); !ok {
		t.Errorf("expected *requestBufferOverflowError but got %T (%s)", err, err)
	}

	n, err = putRequestAttributeAddress(b[:9], "address", net.ParseIP("192.0.2.1"))
	if err == nil {
		t.Errorf("expected no data put to small buffer but got %d", n)
	} else if _, ok := err.(*requestBufferOverflowError); !ok {
		t.Errorf("expected *requestBufferOverflowError but got %T (%s)", err, err)
	}
}

func TestPutRequestAttributeNetwork(t *testing.T) {
	var b [14]byte

	n, err := putRequestAttributeNetwork(b[:], "network", makeTestNetwork("192.0.2.0/24"))
	assertRequestBytesBuffer(t, "putRequestAttributeFloat", err, b[:], n,
		7, 'n', 'e', 't', 'w', 'o', 'r', 'k', byte(requestWireTypeIPv4Network), 24, 192, 0, 2, 0,
	)

	n, err = putRequestAttributeNetwork(b[:4], "network", makeTestNetwork("192.0.2.0/24"))
	if err == nil {
		t.Errorf("expected no data put to small buffer but got %d", n)
	} else if _, ok := err.(*requestBufferOverflowError); !ok {
		t.Errorf("expected *requestBufferOverflowError but got %T (%s)", err, err)
	}

	n, err = putRequestAttributeNetwork(b[:10], "network", makeTestNetwork("192.0.2.0/24"))
	if err == nil {
		t.Errorf("expected no data put to small buffer but got %d", n)
	} else if _, ok := err.(*requestBufferOverflowError); !ok {
		t.Errorf("expected *requestBufferOverflowError but got %T (%s)", err, err)
	}
}

func TestPutRequestAttributeDomain(t *testing.T) {
	var b [25]byte

	n, err := putRequestAttributeDomain(b[:], "domain", makeTestDomain("www.example.com"))
	assertRequestBytesBuffer(t, "putRequestAttributeFloat", err, b[:], n,
		6, 'd', 'o', 'm', 'a', 'i', 'n', byte(requestWireTypeDomain),
		15, 0, 'w', 'w', 'w', '.', 'e', 'x', 'a', 'm', 'p', 'l', 'e', '.', 'c', 'o', 'm',
	)

	n, err = putRequestAttributeDomain(b[:4], "domain", makeTestDomain("www.example.com"))
	if err == nil {
		t.Errorf("expected no data put to small buffer but got %d", n)
	} else if _, ok := err.(*requestBufferOverflowError); !ok {
		t.Errorf("expected *requestBufferOverflowError but got %T (%s)", err, err)
	}

	n, err = putRequestAttributeDomain(b[:10], "domain", makeTestDomain("www.example.com"))
	if err == nil {
		t.Errorf("expected no data put to small buffer but got %d", n)
	} else if _, ok := err.(*requestBufferOverflowError); !ok {
		t.Errorf("expected *requestBufferOverflowError but got %T (%s)", err, err)
	}
}

func TestPutRequestBooleanValue(t *testing.T) {
	var b [1]byte

	n, err := putRequestBooleanValue(b[:], true)
	assertRequestBytesBuffer(t, "putRequestBooleanValue(true)", err, b[:], n, byte(requestWireTypeBooleanTrue))

	n, err = putRequestBooleanValue(b[:], false)
	assertRequestBytesBuffer(t, "putRequestBooleanValue(false)", err, b[:], n, byte(requestWireTypeBooleanFalse))
}

func TestPutRequestStringValue(t *testing.T) {
	var b [7]byte

	n, err := putRequestStringValue(b[:], "test")
	assertRequestBytesBuffer(t, "putRequestStringValue", err, b[:], n,
		byte(requestWireTypeString), 4, 0, 't', 'e', 's', 't',
	)

	n, err = putRequestStringValue(nil, "test")
	assertRequestBufferOverflow(t, "putRequestStringValue", err, n)

	n, err = putRequestStringValue(b[:], string(make([]byte, math.MaxUint16+1)))
	if err == nil {
		t.Errorf("expected no data put to buffer for large string value but got %d", n)
	} else if _, ok := err.(*requestTooLongStringValueError); !ok {
		t.Errorf("expected *requestTooLongStringValueError but got %T (%s)", err, err)
	}

	n, err = putRequestStringValue(b[:3], "test")
	if err == nil {
		t.Errorf("expected no data put to small buffer but got %d", n)
	} else if _, ok := err.(*requestBufferOverflowError); !ok {
		t.Errorf("expected *requestBufferOverflowError but got %T (%s)", err, err)
	}
}

func TestPutRequestIntegerValue(t *testing.T) {
	var b [9]byte

	n, err := putRequestIntegerValue(b[:], -9223372036854775808)
	assertRequestBytesBuffer(t, "putRequestIntegerValue", err, b[:], n,
		byte(requestWireTypeInteger), 0, 0, 0, 0, 0, 0, 0, 0x80,
	)

	n, err = putRequestIntegerValue(nil, 0)
	assertRequestBufferOverflow(t, "putRequestIntegerValue", err, n)

	n, err = putRequestIntegerValue(b[:5], 0)
	if err == nil {
		t.Errorf("expected no data put to small buffer but got %d", n)
	} else if _, ok := err.(*requestBufferOverflowError); !ok {
		t.Errorf("expected *requestBufferOverflowError but got %T (%s)", err, err)
	}
}

func TestPutRequestFloatValue(t *testing.T) {
	var b [9]byte

	n, err := putRequestFloatValue(b[:], math.Pi)
	assertRequestBytesBuffer(t, "putRequestFloatValue", err, b[:], n,
		byte(requestWireTypeFloat), 24, 45, 68, 84, 251, 33, 9, 64,
	)

	n, err = putRequestFloatValue(nil, 0)
	assertRequestBufferOverflow(t, "putRequestFloatValue", err, n)

	n, err = putRequestFloatValue(b[:5], 0)
	if err == nil {
		t.Errorf("expected no data put to small buffer but got %d", n)
	} else if _, ok := err.(*requestBufferOverflowError); !ok {
		t.Errorf("expected *requestBufferOverflowError but got %T (%s)", err, err)
	}
}

func TestPutRequestAddressValue(t *testing.T) {
	var b [17]byte

	n, err := putRequestAddressValue(b[:5], net.ParseIP("192.0.2.1"))
	assertRequestBytesBuffer(t, "putRequestAddressValue(IPv4)", err, b[:5], n,
		byte(requestWireTypeIPv4Address), 192, 0, 2, 1,
	)

	n, err = putRequestAddressValue(b[:17], net.ParseIP("2001:db8::1"))
	assertRequestBytesBuffer(t, "putRequestAddressValue(IPv6)", err, b[:17], n,
		byte(requestWireTypeIPv6Address), 32, 1, 13, 184, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1,
	)

	n, err = putRequestAddressValue(b[:], []byte{192, 0, 2, 1, 0})
	if err == nil {
		t.Errorf("expected no data put to buffer for invalid IP address but got %d", n)
	} else if _, ok := err.(*requestAddressValueError); !ok {
		t.Errorf("expected *requestAddressValueError but got %T (%s)", err, err)
	}

	n, err = putRequestAddressValue(nil, net.ParseIP("192.0.2.1"))
	assertRequestBufferOverflow(t, "putRequestAddressValue", err, n)

	n, err = putRequestAddressValue(b[:3], net.ParseIP("192.0.2.1"))
	if err == nil {
		t.Errorf("expected no data put to small buffer but got %d", n)
	} else if _, ok := err.(*requestBufferOverflowError); !ok {
		t.Errorf("expected *requestBufferOverflowError but got %T (%s)", err, err)
	}
}

func TestPutRequestNetworkValue(t *testing.T) {
	var b [18]byte

	n, err := putRequestNetworkValue(b[:6], makeTestNetwork("192.0.2.0/24"))
	assertRequestBytesBuffer(t, "putRequestNetworkValue(IPv4)", err, b[:6], n,
		byte(requestWireTypeIPv4Network), 24, 192, 0, 2, 0,
	)

	n, err = putRequestNetworkValue(b[:18], makeTestNetwork("2001:db8::/32"))
	assertRequestBytesBuffer(t, "putRequestNetworkValue(IPv6)", err, b[:18], n,
		byte(requestWireTypeIPv6Network), 32, 32, 1, 13, 184, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	)

	n, err = putRequestNetworkValue(b[:], nil)
	if err == nil {
		t.Errorf("expected no data put to buffer for invalid IP network but got %d", n)
	} else if _, ok := err.(*requestInvalidNetworkValueError); !ok {
		t.Errorf("expected *requestInvalidNetworkValueError but got %T (%s)", err, err)
	}

	n, err = putRequestNetworkValue(b[:], &net.IPNet{IP: nil, Mask: net.CIDRMask(24, 32)})
	if err == nil {
		t.Errorf("expected no data put to buffer for invalid IP network but got %d", n)
	} else if _, ok := err.(*requestInvalidNetworkValueError); !ok {
		t.Errorf("expected *requestInvalidNetworkValueError but got %T (%s)", err, err)
	}

	n, err = putRequestNetworkValue(b[:], &net.IPNet{IP: net.ParseIP("192.0.2.1").To4(), Mask: nil})
	if err == nil {
		t.Errorf("expected no data put to buffer for invalid IP network but got %d", n)
	} else if _, ok := err.(*requestInvalidNetworkValueError); !ok {
		t.Errorf("expected *requestInvalidNetworkValueError but got %T (%s)", err, err)
	}

	n, err = putRequestNetworkValue(nil, makeTestNetwork("192.0.2.0/24"))
	assertRequestBufferOverflow(t, "putRequestNetworkValue", err, n)

	n, err = putRequestNetworkValue(b[:4], makeTestNetwork("192.0.2.0/24"))
	if err == nil {
		t.Errorf("expected no data put to small buffer but got %d", n)
	} else if _, ok := err.(*requestBufferOverflowError); !ok {
		t.Errorf("expected *requestBufferOverflowError but got %T (%s)", err, err)
	}
}

func TestPutRequestDomainValue(t *testing.T) {
	var b [18]byte

	n, err := putRequestDomainValue(b[:], makeTestDomain("www.example.com"))
	assertRequestBytesBuffer(t, "putRequestDomainValue", err, b[:], n,
		byte(requestWireTypeDomain), 15, 0, 'w', 'w', 'w', '.', 'e', 'x', 'a', 'm', 'p', 'l', 'e', '.', 'c', 'o', 'm',
	)

	n, err = putRequestDomainValue(nil, makeTestDomain("www.example.com"))
	assertRequestBufferOverflow(t, "putRequestDomainValue", err, n)

	n, err = putRequestDomainValue(b[:7], makeTestDomain("www.example.com"))
	if err == nil {
		t.Errorf("expected no data put to small buffer but got %d", n)
	} else if _, ok := err.(*requestBufferOverflowError); !ok {
		t.Errorf("expected *requestBufferOverflowError but got %T (%s)", err, err)
	}
}

func assertRequestBytesBuffer(t *testing.T, desc string, err error, b []byte, n int, e ...byte) {
	if err != nil {
		t.Errorf("expected no error for %s but got: %s", desc, err)
	} else if n != len(b) {
		t.Errorf("expected exactly all buffer used (%d bytes) for %s but got %d bytes", len(b), desc, n)
	} else {
		if bytes.Compare(b[:], e) != 0 {
			t.Errorf("expected [% x] for %s but got [% x]", e, desc, b)
		}
	}
}

func assertRequestBufferOverflow(t *testing.T, desc string, err error, n int) {
	if err == nil {
		t.Errorf("expected no data put to nil buffer for %s but got %d", desc, n)
	} else if _, ok := err.(*requestBufferOverflowError); !ok {
		t.Errorf("expected *requestBufferOverflowError for %s but got %T (%s)", desc, err, err)
	}
}

func assertRequestAssignmentExpressions(t *testing.T, desc string, err error, a []AttributeAssignmentExpression, n int, e ...AttributeAssignmentExpression) {
	if err != nil {
		t.Errorf("expected no error for %s but got: %s", desc, err)
	} else if n != len(a) {
		t.Errorf("expected exactly all buffer used (%d assignments) for %s but got %d assignments", len(a), desc, n)
	} else {
		aStrs, err := serializeAssignmentExpressions(a)
		if err != nil {
			t.Errorf("can't serialize assignment %d for %s: %s", len(aStrs)+1, desc, err)
			return
		}

		eStrs, err := serializeAssignmentExpressions(e)
		if err != nil {
			t.Errorf("can't serialize expected assignment %d for %s: %s", len(aStrs)+1, desc, err)
			return
		}

		assertStrings(aStrs, eStrs, desc, t)
	}
}

func serializeAssignmentExpressions(a []AttributeAssignmentExpression) ([]string, error) {
	out := make([]string, len(a))
	for i, a := range a {
		id, t, v, err := a.Serialize(nil)
		if err != nil {
			return out[:i], err
		}

		out[i] = fmt.Sprintf("id: %q, type: %q, value: %q", id, t, v)
	}

	return out, nil
}
