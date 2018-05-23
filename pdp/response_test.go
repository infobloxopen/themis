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

func TestMarshalResponse(t *testing.T) {
	var b [90]byte

	n, err := MarshalResponse(b[:], EffectIndeterminate, testRequestAssignments,
		fmt.Errorf("testError1"),
		fmt.Errorf("testError2"),
	)
	assertRequestBytesBuffer(t, "MarshalResponse", err, b[:], n, append(
		[]byte{
			1, 0, 3,
			43, 0, 'm', 'u', 'l', 't', 'i', 'p', 'l', 'e', ' ', 'e', 'r', 'r', 'o', 'r', 's', ':', ' ',
			'"', 't', 'e', 's', 't', 'E', 'r', 'r', 'o', 'r', '1', '"', ',', ' ',
			'"', 't', 'e', 's', 't', 'E', 'r', 'r', 'o', 'r', '2', '"',
		},
		testWireAttributes...)...,
	)

	n, err = MarshalResponse([]byte{}, EffectIndeterminate, testRequestAssignments,
		fmt.Errorf("testError1"),
		fmt.Errorf("testError2"),
	)
	assertRequestBufferOverflow(t, "MarshalResponse", err, n)

	n, err = MarshalResponse(b[:2], EffectIndeterminate, testRequestAssignments,
		fmt.Errorf("testError1"),
		fmt.Errorf("testError2"),
	)
	if err == nil {
		t.Errorf("expected no data put to small buffer but got %d", n)
	} else if _, ok := err.(*requestBufferOverflowError); !ok {
		t.Errorf("expected *requestBufferOverflowError but got %T (%s)", err, err)
	}

	n, err = MarshalResponse(b[:5], EffectIndeterminate, testRequestAssignments,
		fmt.Errorf("testError1"),
		fmt.Errorf("testError2"),
	)
	if err == nil {
		t.Errorf("expected no data put to small buffer but got %d", n)
	} else if _, ok := err.(*requestBufferOverflowError); !ok {
		t.Errorf("expected *requestBufferOverflowError but got %T (%s)", err, err)
	}

	n, err = MarshalResponse(b[:20], EffectIndeterminate, testRequestAssignments,
		fmt.Errorf("testError1"),
		fmt.Errorf("testError2"),
	)
	assertRequestBytesBuffer(t, "MarshalResponse(longStatus)", err, b[:20], n,
		1, 0, 3,
		15, 0, 's', 't', 'a', 't', 'u', 's', ' ', 't', 'o', 'o', ' ', 'l', 'o', 'n', 'g',
	)

	n, err = MarshalResponse(b[:25], EffectIndeterminate, testRequestAssignments, fmt.Errorf("testError"))
	assertRequestBytesBuffer(t, "MarshalResponse(longStatus)", err, b[:25], n,
		1, 0, 3,
		20, 0, 'o', 'b', 'l', 'i', 'g', 'a', 't', 'i', 'o', 'n', 's', ' ', 't', 'o', 'o', ' ', 'l', 'o', 'n', 'g',
	)

	n, err = MarshalResponse(b[:14], EffectIndeterminate, testRequestAssignments,
		fmt.Errorf("testError"),
	)
	if err == nil {
		t.Errorf("expected no data put to small buffer but got %d", n)
	} else if _, ok := err.(*requestBufferOverflowError); !ok {
		t.Errorf("expected *requestBufferOverflowError but got %T (%s)", err, err)
	}

	n, err = MarshalResponse(b[:], EffectIndeterminate, []AttributeAssignmentExpression{
		MakeAttributeAssignmentExpression(
			MakeAttribute("address", TypeAddress),
			MakeAddressValue(net.IP{1, 2, 3, 4, 5, 6}),
		),
	})
	if err == nil {
		t.Errorf("expected no data put to buffer for response with invalid network but got %d", n)
	} else if _, ok := err.(*requestAddressValueError); !ok {
		t.Errorf("expected *requestAddressValueError but got %T (%s)", err, err)
	}
}

func TestUnmarshalResponse(t *testing.T) {
	var a [3]AttributeAssignmentExpression

	effect, n, err := UnmarshalResponse(append([]byte{1, 0, 1, 0, 0}, testWireAttributes...), a[:])
	assertRequestAssignmentExpressions(t, "UnmarshalResponse", err, a[:], n, testRequestAssignments...)
	if effect != EffectPermit {
		t.Errorf("expected %q effect but got %q", EffectNameFromEnum(EffectPermit), EffectNameFromEnum(effect))
	}

	effect, n, err = UnmarshalResponse(append([]byte{
		1, 0, 3,
		9, 0, 't', 'e', 's', 't', 'E', 'r', 'r', 'o', 'r',
	}, testWireAttributes...), a[:])
	if err == nil {
		t.Errorf("expected *ResponseServerError but got no error")
	} else if _, ok := err.(*ResponseServerError); !ok {
		t.Errorf("expected *ResponseServerError but got %T (%s)", err, err)
	}

	assertRequestAssignmentExpressions(t, "UnmarshalResponse", nil, a[:], n, testRequestAssignments...)
	if effect != EffectIndeterminate {
		t.Errorf("expected %q effect but got %q", EffectNameFromEnum(EffectIndeterminate), EffectNameFromEnum(effect))
	}

	effect, n, err = UnmarshalResponse([]byte{}, a[:])
	if err == nil {
		t.Errorf("expected *requestBufferUnderflowError but got effect %q and %d attributes",
			EffectNameFromEnum(effect), n)
	} else if _, ok := err.(*requestBufferUnderflowError); !ok {
		t.Errorf("expected *requestBufferUnderflowError but got %T (%s)", err, err)
	}

	effect, n, err = UnmarshalResponse([]byte{
		1, 0,
	}, a[:])
	if err == nil {
		t.Errorf("expected *requestBufferUnderflowError but got effect %q and %d attributes",
			EffectNameFromEnum(effect), n)
	} else if _, ok := err.(*requestBufferUnderflowError); !ok {
		t.Errorf("expected *requestBufferUnderflowError but got %T (%s)", err, err)
	}

	effect, n, err = UnmarshalResponse([]byte{
		1, 0, 3,
	}, a[:])
	if err == nil {
		t.Errorf("expected *requestBufferUnderflowError but got effect %q and %d attributes",
			EffectNameFromEnum(effect), n)
	} else if _, ok := err.(*requestBufferUnderflowError); !ok {
		t.Errorf("expected *requestBufferUnderflowError but got %T (%s)", err, err)
	}

	effect, n, err = UnmarshalResponse([]byte{
		1, 0, 3, 0, 0,
	}, a[:])
	if err == nil {
		t.Errorf("expected *requestBufferUnderflowError but got effect %q and %d attributes",
			EffectNameFromEnum(effect), n)
	} else if _, ok := err.(*requestBufferUnderflowError); !ok {
		t.Errorf("expected *requestBufferUnderflowError but got %T (%s)", err, err)
	}
}

func TestPutResponseEffect(t *testing.T) {
	var b [1]byte

	n, err := putResponseEffect(b[:], EffectPermit)
	assertRequestBytesBuffer(t, "putResponseEffect", err, b[:], n, 1)

	n, err = putResponseEffect([]byte{}, EffectPermit)
	assertRequestBufferOverflow(t, "putResponseEffect", err, n)

	n, err = putResponseEffect(b[:], -1)
	if err == nil {
		t.Errorf("expected no data put to buffer for invalid effect but got %d", n)
	} else if _, ok := err.(*responseEffectError); !ok {
		t.Errorf("expected *responseEffectError but got %T (%s)", err, err)
	}
}

func TestGetResponseEffect(t *testing.T) {
	effect, n, err := getResponseEffect([]byte{1})
	if err != nil {
		t.Error(err)
	} else if n != 1 {
		t.Errorf("expected one byte consumed but got %d", n)
	} else if effect != EffectPermit {
		t.Errorf("expected %q effect but got %q",
			EffectNameFromEnum(EffectPermit), EffectNameFromEnum(effect),
		)
	}

	effect, n, err = getResponseEffect([]byte{})
	if err == nil {
		t.Errorf("expected *requestBufferUnderflowError but consumed %d bytes and got %q effect",
			n, EffectNameFromEnum(effect),
		)
	} else if _, ok := err.(*requestBufferUnderflowError); !ok {
		t.Errorf("expected *requestBufferUnderflowError but got %T (%s)", err, err)
	}

	effect, n, err = getResponseEffect([]byte{255})
	if err == nil {
		t.Errorf("expected *responseEffectError but consumed %d bytes and got %q effect",
			n, EffectNameFromEnum(effect),
		)
	} else if _, ok := err.(*responseEffectError); !ok {
		t.Errorf("expected *responseEffectError but got %T (%s)", err, err)
	}
}

func TestPutResponseStatus(t *testing.T) {
	var b [65536]byte

	n, err := putResponseStatus(b[:])
	assertRequestBytesBuffer(t, "putResponseStatus", err, b[:2], n,
		0, 0,
	)

	n, err = putResponseStatus(b[:], fmt.Errorf("test"))
	assertRequestBytesBuffer(t, "putResponseStatus(1)", err, b[:6], n,
		4, 0, 't', 'e', 's', 't',
	)

	n, err = putResponseStatus(b[:], fmt.Errorf("test1"), fmt.Errorf("test2"))
	assertRequestBytesBuffer(t, "putResponseStatus(2)", err, b[:35], n,
		33, 0, 'm', 'u', 'l', 't', 'i', 'p', 'l', 'e', ' ', 'e', 'r', 'r', 'o', 'r', 's', ':', ' ',
		'"', 't', 'e', 's', 't', '1', '"', ',', ' ', '"', 't', 'e', 's', 't', '2', '"',
	)

	n, err = putResponseStatus([]byte{})
	assertRequestBufferOverflow(t, "putResponseStatus", err, n)

	n, err = putResponseStatus([]byte{}, fmt.Errorf("test"))
	assertRequestBufferOverflow(t, "putResponseStatus(1)", err, n)

	s := ""
	for i := 0; i < 6553; i++ {
		s += "0123456789"
	}
	s += "0123\u56db56789"

	e := make([]byte, 65536)
	e[0] = 254
	e[1] = 255
	for i := 0; i < 6553; i++ {
		copy(e[10*i+2:], "0123456789")
	}
	e[65532] = '0'
	e[65533] = '1'
	e[65534] = '2'
	e[65535] = '3'

	n, err = putResponseStatus(b[:], fmt.Errorf(s))
	assertRequestBytesBuffer(t, "putResponseStatus(long)", err, b[:], n, e...)
}

func TestPutResponseStatusTooLong(t *testing.T) {
	if len(responseStatusTooLong) > math.MaxUint16 {
		t.Errorf("expected no more than %d bytes for responseStatusTooLong but got %d",
			math.MaxUint16, len(responseStatusTooLong),
		)
	}

	var b [17]byte

	n, err := putResponseStatusTooLong(b[:])
	assertRequestBytesBuffer(t, "putResponseStatusTooLong", err, b[:], n,
		15, 0, 's', 't', 'a', 't', 'u', 's', ' ', 't', 'o', 'o', ' ', 'l', 'o', 'n', 'g',
	)

	n, err = putResponseStatusTooLong([]byte{})
	assertRequestBufferOverflow(t, "putResponseStatusTooLong", err, n)
}

func TestPutResponseObligationsTooLong(t *testing.T) {
	if len(responseStatusObligationsTooLong) > math.MaxUint16 {
		t.Errorf("expected no more than %d bytes for responseStatusObligationsTooLong but got %d",
			math.MaxUint16, len(responseStatusObligationsTooLong),
		)
	}

	var b [22]byte

	n, err := putResponseObligationsTooLong(b[:])
	assertRequestBytesBuffer(t, "putResponseObligationsTooLong", err, b[:], n,
		20, 0, 'o', 'b', 'l', 'i', 'g', 'a', 't', 'i', 'o', 'n', 's', ' ', 't', 'o', 'o', ' ', 'l', 'o', 'n', 'g',
	)

	n, err = putResponseObligationsTooLong([]byte{})
	assertRequestBufferOverflow(t, "putResponseObligationsTooLong", err, n)
}

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
