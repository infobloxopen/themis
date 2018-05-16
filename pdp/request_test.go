package pdp

import (
	"bytes"
	"math"
	"net"
	"testing"

	"github.com/infobloxopen/go-trees/domain"
)

var testWireRequest = []byte{
	1, 0, 1, 0,
	4, 't', 'e', 's', 't', byte(requestWireTypeString), 10, 0, 't', 'e', 's', 't', ' ', 'v', 'a', 'l', 'u', 'e',
}

func TestRequestWireTypesTotal(t *testing.T) {
	if requestWireTypesTotal != len(requestWireTypeNames) {
		t.Errorf("Expected number of wire type names %d to be equal to total number of wire types %d",
			len(requestWireTypeNames), requestWireTypesTotal)
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
		t.Error("expected *requestBufferUnderflowError error but got nothing")
	} else if _, ok := err.(*requestBufferUnderflowError); !ok {
		t.Errorf("expected *requestBufferUnderflowError error but got %T (%s)", err, err)
	}

	_, err = checkRequestVersion([]byte{2, 0})
	if err == nil {
		t.Error("expected *requestVersionError error but got nothing")
	} else if _, ok := err.(*requestVersionError); !ok {
		t.Errorf("expected *requestVersionError error but got %T (%s)", err, err)
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
	} else if c != 1 {
		t.Errorf("expected %d as attribute count but got %d", 1, c)
	}

	c, _, err = getRequestAttributeCount([]byte{})
	if err == nil {
		t.Errorf("expected *requestBufferUnderflowError error but got count %d", c)
	} else if _, ok := err.(*requestBufferUnderflowError); !ok {
		t.Errorf("expected *requestBufferUnderflowError error but got %T (%s)", err, err)
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
	} else if c != 1 {
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
	} else if name != "test" {
		t.Errorf("expected %q as attribute name but got %q", "test", name)
	}

	name, _, err = getRequestAttributeName([]byte{})
	if err == nil {
		t.Errorf("expected *requestBufferUnderflowError error but got name %q", name)
	} else if _, ok := err.(*requestBufferUnderflowError); !ok {
		t.Errorf("expected *requestBufferUnderflowError error but got %T (%s)", err, err)
	}

	name, _, err = getRequestAttributeName([]byte{4, 't', 'e'})
	if err == nil {
		t.Errorf("expected *requestBufferUnderflowError error but got name %q", name)
	} else if _, ok := err.(*requestBufferUnderflowError); !ok {
		t.Errorf("expected *requestBufferUnderflowError error but got %T (%s)", err, err)
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
	} else if c != 1 {
		t.Fatalf("expected %d as attribute count but got %d", 1, c)
	}

	off += n

	name, n, err := getRequestAttributeName(testWireRequest[off:])
	if err != nil {
		t.Fatal(err)
	} else if name != "test" {
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

		t.Errorf("expected *requestBufferUnderflowError error but got type %q (%d)", tn, at)
	} else if _, ok := err.(*requestBufferUnderflowError); !ok {
		t.Errorf("expected *requestBufferUnderflowError error but got %T (%s)", err, err)
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
	} else if c != 1 {
		t.Fatalf("expected %d as attribute count but got %d", 1, c)
	}

	off += n

	name, n, err := getRequestAttributeName(testWireRequest[off:])
	if err != nil {
		t.Fatal(err)
	} else if name != "test" {
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
	} else if off+n != len(testWireRequest) {
		t.Errorf("expected whole buffer consumed (%d) but got (%d)", len(testWireRequest), off+n)
	} else if v != "test value" {
		t.Errorf("expected string %q as attribute value but got %q", "test value", v)
	}

	v, _, err = getRequestStringValue([]byte{})
	if err == nil {
		t.Errorf("expected *requestBufferUnderflowError error but got string %q", v)
	} else if _, ok := err.(*requestBufferUnderflowError); !ok {
		t.Errorf("expected *requestBufferUnderflowError error but got %T (%s)", err, err)
	}

	v, _, err = getRequestStringValue([]byte{10, 0, 't', 'e', 's', 't'})
	if err == nil {
		t.Errorf("expected *requestBufferUnderflowError error but got string %q", v)
	} else if _, ok := err.(*requestBufferUnderflowError); !ok {
		t.Errorf("expected *requestBufferUnderflowError error but got %T (%s)", err, err)
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
		t.Errorf("expected *requestBufferUnderflowError error but got integer %d", v)
	} else if _, ok := err.(*requestBufferUnderflowError); !ok {
		t.Errorf("expected *requestBufferUnderflowError error but got %T (%s)", err, err)
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
		t.Errorf("expected *requestBufferUnderflowError error but got float %g", v)
	} else if _, ok := err.(*requestBufferUnderflowError); !ok {
		t.Errorf("expected *requestBufferUnderflowError error but got %T (%s)", err, err)
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
		t.Errorf("expected *requestBufferUnderflowError error but got IPv4 address %q", v)
	} else if _, ok := err.(*requestBufferUnderflowError); !ok {
		t.Errorf("expected *requestBufferUnderflowError error but got %T (%s)", err, err)
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
		t.Errorf("expected *requestBufferUnderflowError error but got IPv6 address %q", v)
	} else if _, ok := err.(*requestBufferUnderflowError); !ok {
		t.Errorf("expected *requestBufferUnderflowError error but got %T (%s)", err, err)
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
		t.Errorf("expected *requestBufferUnderflowError error but got IPv4 network %q", v)
	} else if _, ok := err.(*requestBufferUnderflowError); !ok {
		t.Errorf("expected *requestBufferUnderflowError error but got %T (%s)", err, err)
	}

	v, _, err = getRequestIPv4NetworkValue([]byte{
		255, 192, 0, 2, 1,
	})
	if err == nil {
		t.Errorf("expected *requestIPv4InvalidMaskError error but got IPv4 network %q", v)
	} else if _, ok := err.(*requestIPv4InvalidMaskError); !ok {
		t.Errorf("expected *requestIPv4InvalidMaskError error but got %T (%s)", err, err)
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
		t.Errorf("expected *requestBufferUnderflowError error but got IPv6 network %q", v)
	} else if _, ok := err.(*requestBufferUnderflowError); !ok {
		t.Errorf("expected *requestBufferUnderflowError error but got %T (%s)", err, err)
	}

	v, _, err = getRequestIPv6NetworkValue([]byte{
		255, 32, 1, 13, 184, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1,
	})
	if err == nil {
		t.Errorf("expected *requestIPv6InvalidMaskError error but got IPv6 network %q", v)
	} else if _, ok := err.(*requestIPv6InvalidMaskError); !ok {
		t.Errorf("expected *requestIPv6InvalidMaskError error but got %T (%s)", err, err)
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
		t.Errorf("expected *requestBufferUnderflowError error but got domain %q", v)
	} else if _, ok := err.(*requestBufferUnderflowError); !ok {
		t.Errorf("expected *requestBufferUnderflowError error but got %T (%s)", err, err)
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
		t.Errorf("expected *requestBufferUnderflowError error but got attribute %q = %s", name, v.describe())
	} else if _, ok := err.(*requestBufferUnderflowError); !ok {
		t.Errorf("expected *requestBufferUnderflowError error but got %T (%s)", err, err)
	}

	name, v, _, err = getRequestAttribute([]byte{
		6, 'n', 'o', 't', 'y', 'p', 'e',
	})
	if err == nil {
		t.Errorf("expected *requestBufferUnderflowError error but got attribute %q = %s", name, v.describe())
	} else if _, ok := err.(*requestBufferUnderflowError); !ok {
		t.Errorf("expected *requestBufferUnderflowError error but got %T (%s)", err, err)
	}

	name, v, _, err = getRequestAttribute([]byte{
		7, 'u', 'n', 'k', 'n', 'o', 'w', 'n', 255,
	})
	if err == nil {
		t.Errorf("expected *requestAttributeUnmarshallingTypeError error but got attribute %q = %s", name, v.describe())
	} else if _, ok := err.(*requestAttributeUnmarshallingTypeError); !ok {
		t.Errorf("expected *requestAttributeUnmarshallingTypeError error but got %T (%s)", err, err)
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
		t.Errorf("expected *requestBufferUnderflowError error but got attribute %q = %s", name, v.describe())
	} else if _, ok := err.(*requestBufferUnderflowError); !ok {
		t.Errorf("expected *requestBufferUnderflowError error but got %T (%s)", err, err)
	}

	name, v, _, err = getRequestAttribute([]byte{
		7, 'i', 'n', 't', 'e', 'g', 'e', 'r', byte(requestWireTypeInteger), 0, 0, 0, 0,
	})
	if err == nil {
		t.Errorf("expected *requestBufferUnderflowError error but got attribute %q = %s", name, v.describe())
	} else if _, ok := err.(*requestBufferUnderflowError); !ok {
		t.Errorf("expected *requestBufferUnderflowError error but got %T (%s)", err, err)
	}

	name, v, _, err = getRequestAttribute([]byte{
		5, 'f', 'l', 'o', 'a', 't', byte(requestWireTypeFloat), 24, 45, 68, 84,
	})
	if err == nil {
		t.Errorf("expected *requestBufferUnderflowError error but got attribute %q = %s", name, v.describe())
	} else if _, ok := err.(*requestBufferUnderflowError); !ok {
		t.Errorf("expected *requestBufferUnderflowError error but got %T (%s)", err, err)
	}

	name, v, _, err = getRequestAttribute([]byte{
		4, 'I', 'P', 'v', '4', byte(requestWireTypeIPv4Address), 192, 0,
	})
	if err == nil {
		t.Errorf("expected *requestBufferUnderflowError error but got attribute %q = %s", name, v.describe())
	} else if _, ok := err.(*requestBufferUnderflowError); !ok {
		t.Errorf("expected *requestBufferUnderflowError error but got %T (%s)", err, err)
	}

	name, v, _, err = getRequestAttribute([]byte{
		4, 'I', 'P', 'v', '6', byte(requestWireTypeIPv6Address), 32, 1, 13, 184, 0, 0, 0, 0,
	})
	if err == nil {
		t.Errorf("expected *requestBufferUnderflowError error but got attribute %q = %s", name, v.describe())
	} else if _, ok := err.(*requestBufferUnderflowError); !ok {
		t.Errorf("expected *requestBufferUnderflowError error but got %T (%s)", err, err)
	}

	name, v, _, err = getRequestAttribute([]byte{
		11, 'I', 'P', 'v', '4', 'N', 'e', 't', 'w', 'o', 'r', 'k', byte(requestWireTypeIPv4Network), 192, 0, 2, 1,
	})
	if err == nil {
		t.Errorf("expected *requestBufferUnderflowError error but got attribute %q = %s", name, v.describe())
	} else if _, ok := err.(*requestBufferUnderflowError); !ok {
		t.Errorf("expected *requestBufferUnderflowError error but got %T (%s)", err, err)
	}

	name, v, _, err = getRequestAttribute([]byte{
		11, 'I', 'P', 'v', '6', 'N', 'e', 't', 'w', 'o', 'r', 'k', byte(requestWireTypeIPv6Network),
		32, 1, 13, 184, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1,
	})
	if err == nil {
		t.Errorf("expected *requestBufferUnderflowError error but got attribute %q = %s", name, v.describe())
	} else if _, ok := err.(*requestBufferUnderflowError); !ok {
		t.Errorf("expected *requestBufferUnderflowError error but got %T (%s)", err, err)
	}

	name, v, _, err = getRequestAttribute([]byte{
		6, 'd', 'o', 'm', 'a', 'i', 'n', byte(requestWireTypeDomain), 8, 0, 't', 'e', 's', 't',
	})
	if err == nil {
		t.Errorf("expected *requestBufferUnderflowError error but got attribute %q = %s", name, v.describe())
	} else if _, ok := err.(*requestBufferUnderflowError); !ok {
		t.Errorf("expected *requestBufferUnderflowError error but got %T (%s)", err, err)
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

	_, ip, _ := net.ParseCIDR("192.0.2.0/24")
	n, err = putRequestAttribute(b[:14], "network", MakeNetworkValue(ip))
	assertRequestBytesBuffer(t, "putRequestAttribute(network)", err, b[:14], n,
		7, 'n', 'e', 't', 'w', 'o', 'r', 'k', byte(requestWireTypeIPv4Network), 24, 192, 0, 2, 0,
	)

	d, _ := domain.MakeNameFromString("www.example.com")
	n, err = putRequestAttribute(b[:25], "domain", MakeDomainValue(d))
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

	_, ip, _ := net.ParseCIDR("192.0.2.0/24")
	n, err := putRequestAttributeNetwork(b[:], "network", ip)
	assertRequestBytesBuffer(t, "putRequestAttributeFloat", err, b[:], n,
		7, 'n', 'e', 't', 'w', 'o', 'r', 'k', byte(requestWireTypeIPv4Network), 24, 192, 0, 2, 0,
	)

	n, err = putRequestAttributeNetwork(b[:4], "network", ip)
	if err == nil {
		t.Errorf("expected no data put to small buffer but got %d", n)
	} else if _, ok := err.(*requestBufferOverflowError); !ok {
		t.Errorf("expected *requestBufferOverflowError but got %T (%s)", err, err)
	}

	n, err = putRequestAttributeNetwork(b[:10], "network", ip)
	if err == nil {
		t.Errorf("expected no data put to small buffer but got %d", n)
	} else if _, ok := err.(*requestBufferOverflowError); !ok {
		t.Errorf("expected *requestBufferOverflowError but got %T (%s)", err, err)
	}
}

func TestPutRequestAttributeDomain(t *testing.T) {
	var b [25]byte

	d, _ := domain.MakeNameFromString("www.example.com")
	n, err := putRequestAttributeDomain(b[:], "domain", d)
	assertRequestBytesBuffer(t, "putRequestAttributeFloat", err, b[:], n,
		6, 'd', 'o', 'm', 'a', 'i', 'n', byte(requestWireTypeDomain),
		15, 0, 'w', 'w', 'w', '.', 'e', 'x', 'a', 'm', 'p', 'l', 'e', '.', 'c', 'o', 'm',
	)

	n, err = putRequestAttributeDomain(b[:4], "domain", d)
	if err == nil {
		t.Errorf("expected no data put to small buffer but got %d", n)
	} else if _, ok := err.(*requestBufferOverflowError); !ok {
		t.Errorf("expected *requestBufferOverflowError but got %T (%s)", err, err)
	}

	n, err = putRequestAttributeDomain(b[:10], "domain", d)
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

	_, ip, _ := net.ParseCIDR("192.0.2.0/24")
	n, err := putRequestNetworkValue(b[:6], ip)
	assertRequestBytesBuffer(t, "putRequestNetworkValue(IPv4)", err, b[:6], n,
		byte(requestWireTypeIPv4Network), 24, 192, 0, 2, 0,
	)

	_, ip, _ = net.ParseCIDR("2001:db8::/32")
	n, err = putRequestNetworkValue(b[:18], ip)
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

	n, err = putRequestNetworkValue(nil, ip)
	assertRequestBufferOverflow(t, "putRequestNetworkValue", err, n)

	n, err = putRequestNetworkValue(b[:9], ip)
	if err == nil {
		t.Errorf("expected no data put to small buffer but got %d", n)
	} else if _, ok := err.(*requestBufferOverflowError); !ok {
		t.Errorf("expected *requestBufferOverflowError but got %T (%s)", err, err)
	}
}

func TestPutRequestDomainValue(t *testing.T) {
	var b [18]byte

	d, _ := domain.MakeNameFromString("www.example.com")
	n, err := putRequestDomainValue(b[:], d)
	assertRequestBytesBuffer(t, "putRequestDomainValue", err, b[:], n,
		byte(requestWireTypeDomain), 15, 0, 'w', 'w', 'w', '.', 'e', 'x', 'a', 'm', 'p', 'l', 'e', '.', 'c', 'o', 'm',
	)

	n, err = putRequestDomainValue(nil, d)
	assertRequestBufferOverflow(t, "putRequestDomainValue", err, n)

	n, err = putRequestDomainValue(b[:7], d)
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
		t.Errorf("expected *requestBufferOverflowError error for %s but got %T (%s)", desc, err, err)
	}
}
