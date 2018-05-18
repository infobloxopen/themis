package pdp

import (
	"errors"
	"fmt"
	"math"
	"net"
	"reflect"
	"testing"

	"github.com/infobloxopen/go-trees/domain"
)

var (
	testWireAttributes = []byte{
		3, 0,
		6, 's', 't', 'r', 'i', 'n', 'g', byte(requestWireTypeString), 4, 0, 't', 'e', 's', 't',
		7, 'b', 'o', 'o', 'l', 'e', 'a', 'n', byte(requestWireTypeBooleanTrue),
		7, 'i', 'n', 't', 'e', 'g', 'e', 'r', byte(requestWireTypeInteger),
		0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x7f,
	}
)

func TestPutAssignmentExpressions(t *testing.T) {
	var b [42]byte
	n, err := putAssignmentExpressions(b[:], testRequestAssignments)
	assertRequestBytesBuffer(t, "putAssignmentExpressions", err, b[:], n, testWireAttributes...)

	n, err = putAssignmentExpressions([]byte{}, testRequestAssignments)
	assertRequestBufferOverflow(t, "putAssignmentExpressions", err, n)

	n, err = putAssignmentExpressions(b[:], []AttributeAssignmentExpression{
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

	n, err = putAssignmentExpressions(b[:12], testRequestAssignments)
	if err == nil {
		t.Errorf("expected no data put to small buffer but got %d", n)
	} else if _, ok := err.(*requestBufferOverflowError); !ok {
		t.Errorf("expected *requestBufferOverflowError but got %T (%s)", err, err)
	}
}

func TestPutAttributesFromReflection(t *testing.T) {
	var b [109]byte

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
	n, err := putAttributesFromReflection(b[:], 7, f)
	assertRequestBytesBuffer(t, "putAttributesFromReflection", err, b[:], n,
		7, 0,
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

	n, err = putAttributesFromReflection([]byte{}, 1, f)
	assertRequestBufferOverflow(t, "putAttributesFromReflection", err, n)

	testFuncErr := errors.New("test function error")
	n, err = putAttributesFromReflection(b[:], 1, func(i int) (string, Type, reflect.Value, error) {
		return "", TypeUndefined, reflect.ValueOf(nil), testFuncErr
	})
	if err == nil {
		t.Errorf("expected no data put to buffer for broken function but got %d", n)
	} else if err != testFuncErr {
		t.Errorf("expected testFuncErr but got %T (%s)", err, err)
	}

	n, err = putAttributesFromReflection(b[:], 1, func(i int) (string, Type, reflect.Value, error) {
		return "undefined", TypeUndefined, reflect.ValueOf(nil), nil
	})
	if err == nil {
		t.Errorf("expected no data put to buffer for undefined value but got %d", n)
	} else if _, ok := err.(*requestAttributeMarshallingNotImplemented); !ok {
		t.Errorf("expected *requestAttributeMarshallingNotImplemented but got %T (%s)", err, err)
	}

	n, err = putAttributesFromReflection(b[:10], 1, f)
	if err == nil {
		t.Errorf("expected no data put to small buffer but got %d", n)
	} else if _, ok := err.(*requestBufferOverflowError); !ok {
		t.Errorf("expected *requestBufferOverflowError but got %T (%s)", err, err)
	}
}

func TestGetAssignmentExpressions(t *testing.T) {
	var a [3]AttributeAssignmentExpression

	n, err := getAssignmentExpressions(testWireAttributes, a[:])
	assertRequestAssignmentExpressions(t, "getAssignmentExpressions", err, a[:], n, testRequestAssignments...)

	n, err = getAssignmentExpressions([]byte{}, a[:])
	if err == nil {
		t.Errorf("expected *requestBufferUnderflowError but got %d bytes", n)
	} else if _, ok := err.(*requestBufferUnderflowError); !ok {
		t.Errorf("expected *requestBufferUnderflowError but got %T (%s)", err, err)
	}

	n, err = getAssignmentExpressions([]byte{255, 255}, a[:])
	if err == nil {
		t.Errorf("expected *requestAssignmentsOverflowError but got %d bytes", n)
	} else if _, ok := err.(*requestAssignmentsOverflowError); !ok {
		t.Errorf("expected *requestAssignmentsOverflowError but got %T (%s)", err, err)
	}

	n, err = getAssignmentExpressions([]byte{1, 0}, a[:])
	if err == nil {
		t.Errorf("expected *requestBufferUnderflowError but got got %d bytes", n)
	} else if _, ok := err.(*requestBufferUnderflowError); !ok {
		t.Errorf("expected *requestBufferUnderflowError but got %T (%s)", err, err)
	}
}

func TestGetAttributesToReflection(t *testing.T) {
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

	err := getAttributesToReflection([]byte{
		byte(len(names)), 0,
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

	assertRequestAssignmentExpressions(t, "getAttributesToReflection", err, a, i,
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

	err = getAttributesToReflection([]byte{}, func(id string, t Type) (reflect.Value, bool, error) {
		return reflect.ValueOf(nil), false, fmt.Errorf("in unreacheable place with id %q and type %q", id, t)
	})
	if err == nil {
		t.Error("expected *requestBufferUnderflowError but got nothing")
	} else if _, ok := err.(*requestBufferUnderflowError); !ok {
		t.Errorf("expected *requestBufferUnderflowError but got %T (%s)", err, err)
	}

	err = getAttributesToReflection([]byte{
		1, 0,
	}, func(id string, t Type) (reflect.Value, bool, error) {
		return reflect.ValueOf(nil), false, fmt.Errorf("in unreacheable place with id %q and type %q", id, t)
	})
	if err == nil {
		t.Error("expected *requestBufferUnderflowError but got nothing")
	} else if _, ok := err.(*requestBufferUnderflowError); !ok {
		t.Errorf("expected *requestBufferUnderflowError but got %T (%s)", err, err)
	}

	err = getAttributesToReflection([]byte{
		1, 0,
		7, 's', 't', 'r', 'i', 'n', 'g', 's',
	}, func(id string, t Type) (reflect.Value, bool, error) {
		return reflect.ValueOf(nil), false, fmt.Errorf("in unreacheable place with id %q and type %q", id, t)
	})
	if err == nil {
		t.Error("expected *requestBufferUnderflowError but got nothing")
	} else if _, ok := err.(*requestBufferUnderflowError); !ok {
		t.Errorf("expected *requestBufferUnderflowError but got %T (%s)", err, err)
	}

	err = getAttributesToReflection([]byte{
		1, 0,
		7, 's', 't', 'r', 'i', 'n', 'g', 's', byte(requestWireTypeSetOfStrings),
	}, func(id string, t Type) (reflect.Value, bool, error) {
		return reflect.ValueOf(nil), false, fmt.Errorf("in unreacheable place with id %q and type %q", id, t)
	})
	if err == nil {
		t.Error("expected *requestAttributeUnmarshallingNotImplemented but got nothing")
	} else if _, ok := err.(*requestAttributeUnmarshallingNotImplemented); !ok {
		t.Errorf("expected *requestAttributeUnmarshallingNotImplemented but got %T (%s)", err, err)
	}

	err = getAttributesToReflection([]byte{
		1, 0,
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
	err = getAttributesToReflection(testWireAttributes, func(id string, t Type) (reflect.Value, bool, error) {
		return reflect.ValueOf(nil), false, testFuncErr
	})
	if err == nil {
		t.Error("expected testFuncErr but got nothing")
	} else if err != testFuncErr {
		t.Errorf("expected testFuncErr but got %T (%s)", err, err)
	}

	err = getAttributesToReflection([]byte{
		1, 0,
		7, 's', 't', 'r', 'i', 'n', 'g', 's', byte(requestWireTypeString), 4, 0, 't', 'e',
	}, func(id string, t Type) (reflect.Value, bool, error) {
		return values[2], true, nil
	})
	if err == nil {
		t.Errorf("expected *requestBufferUnderflowError but got %s", str)
	} else if _, ok := err.(*requestBufferUnderflowError); !ok {
		t.Errorf("expected *requestBufferUnderflowError but got %T (%s)", err, err)
	}

	err = getAttributesToReflection([]byte{
		1, 0,
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

	err = getAttributesToReflection([]byte{
		1, 0,
		7, 'i', 'n', 't', 'e', 'g', 'e', 'r', byte(requestWireTypeInteger), 128, 0, 0, 0, 0, 0, 0, 0,
	}, func(id string, t Type) (reflect.Value, bool, error) {
		return v, true, nil
	})
	if err == nil {
		t.Errorf("expected *requestUnmarshalIntegerOverflowError but got %s", str)
	} else if _, ok := err.(*requestUnmarshalIntegerOverflowError); !ok {
		t.Errorf("expected *requestUnmarshalIntegerOverflowError but got %T (%s)", err, err)
	}

	err = getAttributesToReflection([]byte{
		1, 0,
		5, 'f', 'l', 'o', 'a', 't', byte(requestWireTypeFloat), 24, 45, 68, 84,
	}, func(id string, t Type) (reflect.Value, bool, error) {
		return values[4], true, nil
	})
	if err == nil {
		t.Errorf("expected *requestBufferUnderflowError but got %s", str)
	} else if _, ok := err.(*requestBufferUnderflowError); !ok {
		t.Errorf("expected *requestBufferUnderflowError but got %T (%s)", err, err)
	}

	err = getAttributesToReflection([]byte{
		1, 0,
		8, 'a', 'd', 'd', 'r', 'e', 's', 's', '4', byte(requestWireTypeIPv4Address), 192, 0,
	}, func(id string, t Type) (reflect.Value, bool, error) {
		return values[5], true, nil
	})
	if err == nil {
		t.Errorf("expected *requestBufferUnderflowError but got %s", str)
	} else if _, ok := err.(*requestBufferUnderflowError); !ok {
		t.Errorf("expected *requestBufferUnderflowError but got %T (%s)", err, err)
	}

	err = getAttributesToReflection([]byte{
		1, 0,
		8, 'a', 'd', 'd', 'r', 'e', 's', 's', '6', byte(requestWireTypeIPv6Address), 32, 1, 13, 184, 0, 0, 0, 0,
	}, func(id string, t Type) (reflect.Value, bool, error) {
		return values[6], true, nil
	})
	if err == nil {
		t.Errorf("expected *requestBufferUnderflowError but got %s", str)
	} else if _, ok := err.(*requestBufferUnderflowError); !ok {
		t.Errorf("expected *requestBufferUnderflowError but got %T (%s)", err, err)
	}

	err = getAttributesToReflection([]byte{
		1, 0,
		8, 'n', 'e', 't', 'w', 'o', 'r', 'k', '4', byte(requestWireTypeIPv4Network), 24, 192, 0,
	}, func(id string, t Type) (reflect.Value, bool, error) {
		return values[7], true, nil
	})
	if err == nil {
		t.Errorf("expected *requestBufferUnderflowError but got %s", str)
	} else if _, ok := err.(*requestBufferUnderflowError); !ok {
		t.Errorf("expected *requestBufferUnderflowError but got %T (%s)", err, err)
	}

	err = getAttributesToReflection([]byte{
		1, 0,
		8, 'n', 'e', 't', 'w', 'o', 'r', 'k', '6', byte(requestWireTypeIPv6Network), 32, 32, 1, 13, 184, 0, 0, 0, 0,
	}, func(id string, t Type) (reflect.Value, bool, error) {
		return values[8], true, nil
	})
	if err == nil {
		t.Errorf("expected *requestBufferUnderflowError but got %s", str)
	} else if _, ok := err.(*requestBufferUnderflowError); !ok {
		t.Errorf("expected *requestBufferUnderflowError but got %T (%s)", err, err)
	}

	err = getAttributesToReflection([]byte{
		1, 0,
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
