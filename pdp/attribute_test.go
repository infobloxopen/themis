package pdp

import (
	"net"
	"testing"

	"github.com/infobloxopen/go-trees/domaintree"
	"github.com/infobloxopen/go-trees/iptree"
	"github.com/infobloxopen/go-trees/strtree"
)

func TestTypes(t *testing.T) {
	if typesTotal != len(TypeNames) {
		t.Errorf("Expected total number of types to be equal to number of type names "+
			"but got typesTotal = %d and len(TypeNames) = %d", typesTotal, len(TypeNames))
	}
}

func TestAttribute(t *testing.T) {
	ctx, err := NewContext(nil, 0, nil)
	if err != nil {
		t.Fatalf("Expected context but got error %s", err)
	}

	a := MakeAttribute("test", TypeString)
	if a.id != "test" {
		t.Errorf("Expected \"test\" as attribute id but got %q", a.id)
	}

	at := a.GetType()
	if at != TypeString {
		t.Errorf("Expected %q as attribute type but got %q", TypeNames[TypeString], TypeNames[at])
	}

	v := AttributeValue{t: -1, v: nil}
	expDesc := "val(unknown type)"
	d := v.describe()
	if d != expDesc {
		t.Errorf("Expected %q as value description but got %q", expDesc, d)
	}

	r, err := v.Calculate(ctx)
	if err != nil {
		t.Errorf("Expected no error but got %s", err)
	} else if r.t != v.t || r.v != v.v {
		t.Errorf("Expected the same attribute with type %d and value %T (%#v) but got %d and %T (%#v)",
			v.t, v.v, v.v, r.t, r.v, r.v)
	}

	v = undefinedValue
	vt := v.GetResultType()
	if vt != TypeUndefined {
		t.Errorf("Expected %q as value type but got %q", TypeNames[TypeUndefined], TypeNames[vt])
	}

	expDesc = "val(undefined)"
	d = v.describe()
	if d != expDesc {
		t.Errorf("Expected %q as value description but got %q", expDesc, d)
	}

	v = MakeBooleanValue(true)
	vt = v.GetResultType()
	if vt != TypeBoolean {
		t.Errorf("Expected %q as value type but got %q", TypeNames[TypeBoolean], TypeNames[vt])
	}

	expDesc = "true"
	d = v.describe()
	if d != expDesc {
		t.Errorf("Expected %q as value description but got %q", expDesc, d)
	}

	v = MakeStringValue("test")
	vt = v.GetResultType()
	if vt != TypeString {
		t.Errorf("Expected %q as value type but got %q", TypeNames[TypeString], TypeNames[vt])
	}

	expDesc = "\"test\""
	d = v.describe()
	if d != expDesc {
		t.Errorf("Expected %q as value description but got %q", expDesc, d)
	}

	v = MakeIntegerValue(123)
	vt = v.GetResultType()
	if vt != TypeInteger {
		t.Errorf("Expected %q as value type but got %q", TypeNames[TypeInteger], TypeNames[vt])
	}

	expDesc = "123"
	d = v.describe()
	if d != expDesc {
		t.Errorf("Expected %q as value description but got %q", expDesc, d)
	}

	v = MakeFloatValue(123.456)
	vt = v.GetResultType()
	if vt != TypeFloat {
		t.Errorf("Expected %q as value type but got %q", TypeNames[TypeFloat], TypeNames[vt])
	}

	expDesc = "123.456"
	d = v.describe()
	if d != expDesc {
		t.Errorf("Expected %q as value description but got %q", expDesc, d)
	}

	v = MakeAddressValue(net.ParseIP("192.0.2.1"))
	vt = v.GetResultType()
	if vt != TypeAddress {
		t.Errorf("Expected %q as value type but got %q", TypeNames[TypeAddress], TypeNames[vt])
	}

	expDesc = "192.0.2.1"
	d = v.describe()
	if d != expDesc {
		t.Errorf("Expected %q as value description but got %q", expDesc, d)
	}

	_, n, err := net.ParseCIDR("192.0.2.0/24")
	if err != nil {
		t.Fatalf("Can't parse 192.0.2.0/24 as network: %s", err)
	}
	v = MakeNetworkValue(n)
	vt = v.GetResultType()
	if vt != TypeNetwork {
		t.Errorf("Expected %q as value type but got %q", TypeNames[TypeNetwork], TypeNames[vt])
	}

	expDesc = "192.0.2.0/24"
	d = v.describe()
	if d != expDesc {
		t.Errorf("Expected %q as value description but got %q", expDesc, d)
	}

	v = MakeDomainValue(domaintree.WireDomainNameLower("\x07example\x03com\x00"))
	vt = v.GetResultType()
	if vt != TypeDomain {
		t.Errorf("Expected %q as value type but got %q", TypeNames[TypeDomain], TypeNames[vt])
	}

	expDesc = "domain(example.com)"
	d = v.describe()
	if d != expDesc {
		t.Errorf("Expected %q as value description but got %q", expDesc, d)
	}

	sTree := strtree.NewTree()
	sTree.InplaceInsert("1 - one", 1)
	sTree.InplaceInsert("2 - two", 2)
	sTree.InplaceInsert("3 - three", 3)
	sTree.InplaceInsert("4 - four", 4)
	v = MakeSetOfStringsValue(sTree)
	vt = v.GetResultType()
	if vt != TypeSetOfStrings {
		t.Errorf("Expected %q as value type but got %q", TypeNames[TypeSetOfStrings], TypeNames[vt])
	}

	expDesc = "set(\"1 - one\", \"2 - two\", ...)"
	d = v.describe()
	if d != expDesc {
		t.Errorf("Expected %q as value description but got %q", expDesc, d)
	}

	nTree := iptree.NewTree()
	_, n, err = net.ParseCIDR("192.0.2.16/28")
	if err != nil {
		t.Fatalf("Can't parse 192.0.2.16/28 as network: %s", err)
	}
	nTree.InplaceInsertNet(n, 1)
	_, n, err = net.ParseCIDR("192.0.2.32/28")
	if err != nil {
		t.Fatalf("Can't parse 192.0.2.32/28 as network: %s", err)
	}
	nTree.InplaceInsertNet(n, 2)
	_, n, err = net.ParseCIDR("192.0.2.48/28")
	if err != nil {
		t.Fatalf("Can't parse 192.0.2.48/28 as network: %s", err)
	}
	nTree.InplaceInsertNet(n, 3)
	_, n, err = net.ParseCIDR("192.0.2.64/28")
	if err != nil {
		t.Fatalf("Can't parse 192.0.2.64/28 as network: %s", err)
	}
	nTree.InplaceInsertNet(n, 4)
	v = MakeSetOfNetworksValue(nTree)
	vt = v.GetResultType()
	if vt != TypeSetOfNetworks {
		t.Errorf("Expected %q as value type but got %q", TypeNames[TypeSetOfNetworks], TypeNames[vt])
	}

	expDesc = "set(192.0.2.16/28, 192.0.2.32/28, ...)"
	d = v.describe()
	if d != expDesc {
		t.Errorf("Expected %q as value description but got %q", expDesc, d)
	}

	dTree := &domaintree.Node{}
	dTree.InplaceInsert("example.com", 1)
	dTree.InplaceInsert("example.gov", 2)
	dTree.InplaceInsert("example.net", 3)
	dTree.InplaceInsert("example.org", 4)
	v = MakeSetOfDomainsValue(dTree)
	vt = v.GetResultType()
	if vt != TypeSetOfDomains {
		t.Errorf("Expected %q as value type but got %q", TypeNames[TypeSetOfDomains], TypeNames[vt])
	}

	expDesc = "domains(\"example.com\", \"example.gov\", ...)"
	d = v.describe()
	if d != expDesc {
		t.Errorf("Expected %q as value description but got %q", expDesc, d)
	}

	v = MakeListOfStringsValue([]string{"one", "two", "three", "four"})
	vt = v.GetResultType()
	if vt != TypeListOfStrings {
		t.Errorf("Expected %q as value type but got %q", TypeNames[TypeListOfStrings], TypeNames[vt])
	}

	expDesc = "[\"one\", \"two\", ...]"
	d = v.describe()
	if d != expDesc {
		t.Errorf("Expected %q as value description but got %q", expDesc, d)
	}
}

func TestMakeValueFromSting(t *testing.T) {
	v, err := MakeValueFromString(-1, "test")
	if err == nil {
		t.Errorf("Expected error but got value: %s", v.describe())
	} else if _, ok := err.(*unknownTypeStringCastError); !ok {
		t.Errorf("Expected *unknownTypeStringCastError but got %T (%s)", err, err)
	}

	v, err = MakeValueFromString(TypeUndefined, "test")
	if err == nil {
		t.Errorf("Expected error but got value: %s", v.describe())
	} else if _, ok := err.(*invalidTypeStringCastError); !ok {
		t.Errorf("Expected *invalidTypeStringCastError but got %T (%s)", err, err)
	}

	v, err = MakeValueFromString(TypeSetOfStrings, "test")
	if err == nil {
		t.Errorf("Expected error but got value: %s", v.describe())
	} else if _, ok := err.(*notImplementedStringCastError); !ok {
		t.Errorf("Expected *notImplementedStringCastError but got %T (%s)", err, err)
	}

	v, err = MakeValueFromString(TypeSetOfNetworks, "test")
	if err == nil {
		t.Errorf("Expected error but got value: %s", v.describe())
	} else if _, ok := err.(*notImplementedStringCastError); !ok {
		t.Errorf("Expected *notImplementedStringCastError but got %T (%s)", err, err)
	}

	v, err = MakeValueFromString(TypeSetOfDomains, "test")
	if err == nil {
		t.Errorf("Expected error but got value: %s", v.describe())
	} else if _, ok := err.(*notImplementedStringCastError); !ok {
		t.Errorf("Expected *notImplementedStringCastError but got %T (%s)", err, err)
	}

	v, err = MakeValueFromString(TypeListOfStrings, "test")
	if err == nil {
		t.Errorf("Expected error but got value: %s", v.describe())
	} else if _, ok := err.(*notImplementedStringCastError); !ok {
		t.Errorf("Expected *notImplementedStringCastError but got %T (%s)", err, err)
	}

	v, err = MakeValueFromString(TypeBoolean, "true")
	if err != nil {
		t.Errorf("Expected boolean attribute value but got error: %s", err)
	} else {
		expDesc := "true"
		d := v.describe()
		if d != expDesc {
			t.Errorf("Expected %q as value description but got %q", expDesc, d)
		}
	}

	v, err = MakeValueFromString(TypeBoolean, "not boolean value")
	if err == nil {
		t.Errorf("Expected error but got value: %s", v.describe())
	} else if _, ok := err.(*invalidBooleanStringCastError); !ok {
		t.Errorf("Expected *invalidBooleanStringCastError but got %T (%s)", err, err)
	}

	v, err = MakeValueFromString(TypeString, "test")
	if err != nil {
		t.Errorf("Expected string attribute value but got error: %s", err)
	} else {
		expDesc := "\"test\""
		d := v.describe()
		if d != expDesc {
			t.Errorf("Expected %q as value description but got %q", expDesc, d)
		}
	}

	v, err = MakeValueFromString(TypeInteger, "654321")
	if err != nil {
		t.Errorf("Expected integer attribute value but got error: %s", err)
	} else {
		expDesc := "654321"
		d := v.describe()
		if d != expDesc {
			t.Errorf("Expected %s as value description but got %s", expDesc, d)
		}
	}

	v, err = MakeValueFromString(TypeInteger, "not integer value")
	if err == nil {
		t.Errorf("Expected error but got value: %s", v.describe())
	} else if _, ok := err.(*invalidIntegerStringCastError); !ok {
		t.Errorf("Expected *invalidIntegerStringCastError but got %T (%s)", err, err)
	}

	v, err = MakeValueFromString(TypeFloat, "654.321")
	if err != nil {
		t.Errorf("Expected integer attribute value but got error: %s", err)
	} else {
		expDesc := "654.321"
		d := v.describe()
		if d != expDesc {
			t.Errorf("Expected %s as value description but got %s", expDesc, d)
		}
	}

	v, err = MakeValueFromString(TypeFloat, "0.00000000000654321")
	if err != nil {
		t.Errorf("Expected integer attribute value but got error: %s", err)
	} else {
		expDesc := "6.54321E-12"
		d := v.describe()
		if d != expDesc {
			t.Errorf("Expected %s as value description but got %s", expDesc, d)
		}
	}

	v, err = MakeValueFromString(TypeFloat, "not float value")
	if err == nil {
		t.Errorf("Expected error but got value: %s", v.describe())
	} else if _, ok := err.(*invalidFloatStringCastError); !ok {
		t.Errorf("Expected *invalidFloatStringCastError but got %T (%s)", err, err)
	}

	v, err = MakeValueFromString(TypeAddress, "192.0.2.1")
	if err != nil {
		t.Errorf("Expected address attribute value but got error: %s", err)
	} else {
		expDesc := "192.0.2.1"
		d := v.describe()
		if d != expDesc {
			t.Errorf("Expected %q as value description but got %q", expDesc, d)
		}
	}

	v, err = MakeValueFromString(TypeAddress, "999.999.999.999")
	if err == nil {
		t.Errorf("Expected error but got value: %s", v.describe())
	} else if _, ok := err.(*invalidAddressStringCastError); !ok {
		t.Errorf("Expected *invalidAddressStringCastError but got %T (%s)", err, err)
	}

	v, err = MakeValueFromString(TypeNetwork, "192.0.2.0/24")
	if err != nil {
		t.Errorf("Expected network attribute value but got error: %s", err)
	} else {
		expDesc := "192.0.2.0/24"
		d := v.describe()
		if d != expDesc {
			t.Errorf("Expected %q as value description but got %q", expDesc, d)
		}
	}

	v, err = MakeValueFromString(TypeNetwork, "999.999.999.999/999")
	if err == nil {
		t.Errorf("Expected error but got value: %s", v.describe())
	} else if _, ok := err.(*invalidNetworkStringCastError); !ok {
		t.Errorf("Expected *invalidNetworkStringCastError but got %T (%s)", err, err)
	}

	v, err = MakeValueFromString(TypeDomain, "example.com")
	if err != nil {
		t.Errorf("Expected domain attribute value but got error: %s", err)
	} else {
		expDesc := "domain(example.com)"
		d := v.describe()
		if d != expDesc {
			t.Errorf("Expected %q as value description but got %q", expDesc, d)
		}
	}
}

func TestAttributeValueTypeCast(t *testing.T) {
	v, err := MakeValueFromString(TypeBoolean, "true")
	if err != nil {
		t.Errorf("Expected boolean attribute value but got error: %s", err)
	} else {
		b, err := v.boolean()
		if err != nil {
			t.Errorf("Expected boolean value but got error: %s", err)
		} else if !b {
			t.Errorf("Expected true as attribute value but got %#v", b)
		}

		s, err := v.str()
		if err == nil {
			t.Errorf("Expected error but got string %q", s)
		} else if _, ok := err.(*attributeValueTypeError); !ok {
			t.Errorf("Expected *attributeValueTypeError but got %T (%s)", err, err)
		}
	}

	v, err = MakeValueFromString(TypeInteger, "34567")
	if err != nil {
		t.Errorf("Expected integer attribute value but got error: %s", err)
	} else {
		i, err := v.integer()
		if err != nil {
			t.Errorf("Expected integer value but got error: %s", err)
		} else if i != 34567 {
			t.Errorf("Expected %d as attribute value but got %#v", 34567, i)
		}

		s, err := v.str()
		if err == nil {
			t.Errorf("Expected error but got string %q", s)
		} else if _, ok := err.(*attributeValueTypeError); !ok {
			t.Errorf("Expected *attributeValueTypeError but got %T (%s)", err, err)
		}
	}

	v, err = MakeValueFromString(TypeFloat, "12345.678")
	if err != nil {
		t.Errorf("Expected float attribute value but got error: %s", err)
	} else {
		f, err := v.float()
		if err != nil {
			t.Errorf("Expected float value but got error: %s", err)
		} else if f != 12345.678 {
			t.Errorf("Expected %g as attribute value but got %#v", 12345.678, f)
		}

		s, err := v.str()
		if err == nil {
			t.Errorf("Expected error but got string %q", s)
		} else if _, ok := err.(*attributeValueTypeError); !ok {
			t.Errorf("Expected *attributeValueTypeError but got %T (%s)", err, err)
		}
	}

	v, err = MakeValueFromString(TypeString, "test")
	if err != nil {
		t.Errorf("Expected string attribute value but got error: %s", err)
	} else {
		s, err := v.str()
		if err != nil {
			t.Errorf("Expected string value but got error: %s", err)
		} else if s != "test" {
			t.Errorf("Expected \"test\" as attribute value but got %q", s)
		}

		a, err := v.address()
		if err == nil {
			t.Errorf("Expected error but got address %s", a)
		} else if _, ok := err.(*attributeValueTypeError); !ok {
			t.Errorf("Expected *attributeValueTypeError but got %T (%s)", err, err)
		}
	}

	v, err = MakeValueFromString(TypeAddress, "192.0.2.1")
	if err != nil {
		t.Errorf("Expected address attribute value but got error: %s", err)
	} else {
		_, err := v.address()
		if err != nil {
			t.Errorf("Expected address value but got error: %s", err)
		} else {
			expDesc := "192.0.2.1"
			d := v.describe()
			if d != expDesc {
				t.Errorf("Expected %q as value description but got %q", expDesc, d)
			}
		}

		n, err := v.network()
		if err == nil {
			t.Errorf("Expected error but got network %s", n)
		} else if _, ok := err.(*attributeValueTypeError); !ok {
			t.Errorf("Expected *attributeValueTypeError but got %T (%s)", err, err)
		}
	}

	v, err = MakeValueFromString(TypeNetwork, "192.0.2.0/24")
	if err != nil {
		t.Errorf("Expected network attribute value but got error: %s", err)
	} else {
		_, err := v.network()
		if err != nil {
			t.Errorf("Expected network value but got error: %s", err)
		} else {
			expDesc := "192.0.2.0/24"
			d := v.describe()
			if d != expDesc {
				t.Errorf("Expected %q as value description but got %q", expDesc, d)
			}
		}

		d, err := v.domain()
		if err == nil {
			t.Errorf("Expected error but got domain %s", d)
		} else if _, ok := err.(*attributeValueTypeError); !ok {
			t.Errorf("Expected *attributeValueTypeError but got %T (%s)", err, err)
		}
	}

	v, err = MakeValueFromString(TypeDomain, "example.com")
	if err != nil {
		t.Errorf("Expected domain attribute value but got error: %s", err)
	} else {
		d, err := v.domain()
		if err != nil {
			t.Errorf("Expected domain value but got error: %s", err)
		} else if string(d) != "\x07example\x03com\x00" {
			t.Errorf("Expected \"\\aexample\\x03com\\x00\" as attribute value but got %s", d)
		}

		_, err = v.setOfStrings()
		if err == nil {
			t.Errorf("Expected error but got set of strings %s", v.describe())
		} else if _, ok := err.(*attributeValueTypeError); !ok {
			t.Errorf("Expected *attributeValueTypeError but got %T (%s)", err, err)
		}
	}

	sTree := strtree.NewTree()
	sTree.InplaceInsert("1 - one", 1)
	sTree.InplaceInsert("2 - two", 2)
	sTree.InplaceInsert("3 - three", 3)
	sTree.InplaceInsert("4 - four", 4)
	v = MakeSetOfStringsValue(sTree)

	_, err = v.setOfStrings()
	if err != nil {
		t.Errorf("Expected set of strings value but got error: %s", err)
	} else {
		expDesc := "set(\"1 - one\", \"2 - two\", ...)"
		d := v.describe()
		if d != expDesc {
			t.Errorf("Expected %q as value description but got %q", expDesc, d)
		}
	}

	_, err = v.setOfNetworks()
	if err == nil {
		t.Errorf("Expected error but got set of networks %s", v.describe())
	} else if _, ok := err.(*attributeValueTypeError); !ok {
		t.Errorf("Expected *attributeValueTypeError but got %T (%s)", err, err)
	}

	nTree := iptree.NewTree()
	_, n, err := net.ParseCIDR("192.0.2.16/28")
	if err != nil {
		t.Fatalf("Can't parse 192.0.2.16/28 as network: %s", err)
	}
	nTree.InplaceInsertNet(n, 1)
	_, n, err = net.ParseCIDR("192.0.2.32/28")
	if err != nil {
		t.Fatalf("Can't parse 192.0.2.32/28 as network: %s", err)
	}
	nTree.InplaceInsertNet(n, 2)
	_, n, err = net.ParseCIDR("192.0.2.48/28")
	if err != nil {
		t.Fatalf("Can't parse 192.0.2.48/28 as network: %s", err)
	}
	nTree.InplaceInsertNet(n, 3)
	_, n, err = net.ParseCIDR("192.0.2.64/28")
	if err != nil {
		t.Fatalf("Can't parse 192.0.2.64/28 as network: %s", err)
	}
	nTree.InplaceInsertNet(n, 4)
	v = MakeSetOfNetworksValue(nTree)

	_, err = v.setOfNetworks()
	if err != nil {
		t.Errorf("Expected set of networks value but got error: %s", err)
	} else {
		expDesc := "set(192.0.2.16/28, 192.0.2.32/28, ...)"
		d := v.describe()
		if d != expDesc {
			t.Errorf("Expected %q as value description but got %q", expDesc, d)
		}
	}

	_, err = v.setOfDomains()
	if err == nil {
		t.Errorf("Expected error but got set of domains %s", v.describe())
	} else if _, ok := err.(*attributeValueTypeError); !ok {
		t.Errorf("Expected *attributeValueTypeError but got %T (%s)", err, err)
	}

	dTree := &domaintree.Node{}
	dTree.InplaceInsert("example.com", 1)
	dTree.InplaceInsert("example.gov", 2)
	dTree.InplaceInsert("example.net", 3)
	dTree.InplaceInsert("example.org", 4)
	v = MakeSetOfDomainsValue(dTree)
	_, err = v.setOfDomains()
	if err != nil {
		t.Errorf("Expected set of domains value but got error: %s", err)
	} else {
		expDesc := "domains(\"example.com\", \"example.gov\", ...)"
		d := v.describe()
		if d != expDesc {
			t.Errorf("Expected %q as value description but got %q", expDesc, d)
		}
	}

	_, err = v.listOfStrings()
	if err == nil {
		t.Errorf("Expected error but got list of strings %s", v.describe())
	} else if _, ok := err.(*attributeValueTypeError); !ok {
		t.Errorf("Expected *attributeValueTypeError but got %T (%s)", err, err)
	}

	v = MakeListOfStringsValue([]string{"one", "two", "three", "four"})
	_, err = v.listOfStrings()
	if err != nil {
		t.Errorf("Expected list of strings value but got error: %s", err)
	} else {
		expDesc := "[\"one\", \"two\", ...]"
		d := v.describe()
		if d != expDesc {
			t.Errorf("Expected %q as value description but got %q", expDesc, d)
		}
	}

	b, err := v.boolean()
	if err == nil {
		t.Errorf("Expected error but got boolean %#v", b)
	} else if _, ok := err.(*attributeValueTypeError); !ok {
		t.Errorf("Expected *attributeValueTypeError but got %T (%s)", err, err)
	}
}

func TestAttributeValueSerialize(t *testing.T) {
	v := AttributeValue{t: -1, v: nil}
	s, err := v.Serialize()
	if err == nil {
		t.Errorf("Expected error but got string %q", s)
	} else if _, ok := err.(*unknownTypeSerializationError); !ok {
		t.Errorf("Expected *unknownTypeSerializationError but got %T (%s)", err, err)
	}

	v = undefinedValue
	s, err = v.Serialize()
	if err == nil {
		t.Errorf("Expected error but got string %q", s)
	} else if _, ok := err.(*invalidTypeSerializationError); !ok {
		t.Errorf("Expected *invalidTypeSerializationError but got %T (%s)", err, err)
	}

	v, err = MakeValueFromString(TypeBoolean, "true")
	if err != nil {
		t.Errorf("Expected no error but got %s", err)
	} else {
		s, err := v.Serialize()
		if err != nil {
			t.Errorf("Expected no error but got %s", err)
		} else if s != "true" {
			t.Errorf("Expected \"true\" but got %q", s)
		}
	}

	v, err = MakeValueFromString(TypeInteger, "47238")
	if err != nil {
		t.Errorf("Expected no error but got %s", err)
	} else {
		s, err := v.Serialize()
		if err != nil {
			t.Errorf("Expected no error but got %s", err)
		} else if s != "47238" {
			t.Errorf("Expected \"true\" but got %q", s)
		}
	}

	v, err = MakeValueFromString(TypeFloat, "3.1415927")
	if err != nil {
		t.Errorf("Expected no error but got %s", err)
	} else {
		s, err := v.Serialize()
		if err != nil {
			t.Errorf("Expected no error but got %s", err)
		} else if s != "3.1415927" {
			t.Errorf("Expected \"true\" but got %q", s)
		}
	}

	v, err = MakeValueFromString(TypeString, "test")
	if err != nil {
		t.Errorf("Expected no error but got %s", err)
	} else {
		s, err := v.Serialize()
		if err != nil {
			t.Errorf("Expected no error but got %s", err)
		} else if s != "test" {
			t.Errorf("Expected \"test\" but got %q", s)
		}
	}

	v, err = MakeValueFromString(TypeAddress, "192.0.2.1")
	if err != nil {
		t.Errorf("Expected no error but got %s", err)
	} else {
		s, err := v.Serialize()
		if err != nil {
			t.Errorf("Expected no error but got %s", err)
		} else if s != "192.0.2.1" {
			t.Errorf("Expected \"192.0.2.1\" but got %q", s)
		}
	}

	v, err = MakeValueFromString(TypeNetwork, "192.0.2.0/24")
	if err != nil {
		t.Errorf("Expected no error but got %s", err)
	} else {
		s, err := v.Serialize()
		if err != nil {
			t.Errorf("Expected no error but got %s", err)
		} else if s != "192.0.2.0/24" {
			t.Errorf("Expected \"192.0.2.0/24\" but got %q", s)
		}
	}

	v, err = MakeValueFromString(TypeDomain, "example.com")
	if err != nil {
		t.Errorf("Expected no error but got %s", err)
	} else {
		s, err := v.Serialize()
		if err != nil {
			t.Errorf("Expected no error but got %s", err)
		} else if s != "example.com" {
			t.Errorf("Expected \"example.com\" but got %q", s)
		}
	}

	sTree := strtree.NewTree()
	sTree.InplaceInsert("1 - one", 1)
	sTree.InplaceInsert("2 - two", 2)
	sTree.InplaceInsert("3 - three", 3)
	sTree.InplaceInsert("4 - four", 4)
	v = MakeSetOfStringsValue(sTree)
	s, err = v.Serialize()
	if err != nil {
		t.Errorf("Expected no error but got %s", err)
	} else {
		e := "\"1 - one\",\"2 - two\",\"3 - three\",\"4 - four\""
		if s != e {
			t.Errorf("Expected %q but got %q", e, s)
		}
	}

	nTree := iptree.NewTree()
	_, n, err := net.ParseCIDR("192.0.2.16/28")
	if err != nil {
		t.Fatalf("Can't parse 192.0.2.16/28 as network: %s", err)
	}
	nTree.InplaceInsertNet(n, 1)
	_, n, err = net.ParseCIDR("192.0.2.32/28")
	if err != nil {
		t.Fatalf("Can't parse 192.0.2.32/28 as network: %s", err)
	}
	nTree.InplaceInsertNet(n, 2)
	_, n, err = net.ParseCIDR("192.0.2.48/28")
	if err != nil {
		t.Fatalf("Can't parse 192.0.2.48/28 as network: %s", err)
	}
	nTree.InplaceInsertNet(n, 3)
	_, n, err = net.ParseCIDR("192.0.2.64/28")
	if err != nil {
		t.Fatalf("Can't parse 192.0.2.64/28 as network: %s", err)
	}
	nTree.InplaceInsertNet(n, 4)
	v = MakeSetOfNetworksValue(nTree)
	s, err = v.Serialize()
	if err != nil {
		t.Errorf("Expected no error but got %s", err)
	} else {
		e := "\"192.0.2.16/28\",\"192.0.2.32/28\",\"192.0.2.48/28\",\"192.0.2.64/28\""
		if s != e {
			t.Errorf("Expected %q but got %q", e, s)
		}
	}

	dTree := &domaintree.Node{}
	dTree.InplaceInsert("example.com", 1)
	dTree.InplaceInsert("example.gov", 2)
	dTree.InplaceInsert("example.net", 3)
	dTree.InplaceInsert("example.org", 4)
	v = MakeSetOfDomainsValue(dTree)
	s, err = v.Serialize()
	if err != nil {
		t.Errorf("Expected no error but got %s", err)
	} else {
		e := "\"example.com\",\"example.gov\",\"example.net\",\"example.org\""
		if s != e {
			t.Errorf("Expected %q but got %q", e, s)
		}
	}

	v = MakeListOfStringsValue([]string{"one", "two", "three", "four"})
	s, err = v.Serialize()
	if err != nil {
		t.Errorf("Expected no error but got %s", err)
	} else {
		e := "\"one\",\"two\",\"three\",\"four\""
		if s != e {
			t.Errorf("Expected %q but got %q", e, s)
		}
	}
}

func TestAttributeAssignmentExpression(t *testing.T) {
	ctx, err := NewContext(nil, 0, nil)
	if err != nil {
		t.Fatalf("Expected context but got error %s", err)
	}

	expect := "test-value"
	v := MakeStringValue(expect)
	a := Attribute{
		id: "test-id",
		t:  TypeString}

	ae := MakeAttributeAssignmentExpression(a, v)
	id, tName, s, err := ae.Serialize(ctx)
	if err != nil {
		t.Errorf("Expected no error but got %s", err)
	} else if id != a.id || tName != TypeKeys[a.t] || s != expect {
		t.Errorf("Expected %q, %q, %q but got %q, %q, %q", a.id, TypeKeys[a.t], expect, id, tName, s)
	}

	dv := MakeDomainValue(domaintree.WireDomainNameLower("\x07example\x03com\x00"))
	v = MakeStringValue(expect)
	e := makeFunctionStringEqual(v, dv)
	a = Attribute{
		id: "test-id",
		t:  TypeBoolean}

	ae = MakeAttributeAssignmentExpression(a, e)
	id, tName, s, err = ae.Serialize(ctx)
	if err == nil {
		t.Errorf("Expected error but got %q, %q, %q", id, tName, s)
	} else if _, ok := err.(*attributeValueTypeError); !ok {
		t.Errorf("Expected *attributeValueTypeError error but got %T (%s)", err, err)
	}

	expect = "test-value"
	v = MakeStringValue(expect)
	a = Attribute{
		id: "test-id",
		t:  TypeBoolean}
	ae = MakeAttributeAssignmentExpression(a, v)
	id, tName, s, err = ae.Serialize(ctx)
	if err == nil {
		t.Errorf("Expected error but got %q, %q, %q", id, tName, s)
	} else if _, ok := err.(*assignmentTypeMismatch); !ok {
		t.Errorf("Expected *ssignmentTypeMismatch error but got %T (%s)", err, err)
	}

	fv := MakeFloatValue(2.718282)
	v = MakeStringValue(expect)
	e = makeFunctionStringEqual(v, fv)
	a = Attribute{
		id: "test-id",
		t:  TypeInteger}

	ae = MakeAttributeAssignmentExpression(a, e)
	id, tName, s, err = ae.Serialize(ctx)
	if err == nil {
		t.Errorf("Expected error but got %q, %q, %q", id, tName, s)
	} else if _, ok := err.(*attributeValueTypeError); !ok {
		t.Errorf("Expected *attributeValueTypeError error but got %T (%s)", err, err)
	}

	v = MakeFloatValue(1234.567)
	a = Attribute{
		id: "test-id",
		t:  TypeInteger}
	ae = MakeAttributeAssignmentExpression(a, v)
	id, tName, s, err = ae.Serialize(ctx)
	if err == nil {
		t.Errorf("Expected error but got %q, %q, %q", id, tName, s)
	} else if _, ok := err.(*assignmentTypeMismatch); !ok {
		t.Errorf("Expected *ssignmentTypeMismatch error but got %T (%s)", err, err)
	}

	iv := MakeIntegerValue(45678)
	v = MakeStringValue(expect)
	e = makeFunctionStringEqual(v, iv)
	a = Attribute{
		id: "test-id",
		t:  TypeFloat}

	ae = MakeAttributeAssignmentExpression(a, e)
	id, tName, s, err = ae.Serialize(ctx)
	if err == nil {
		t.Errorf("Expected error but got %q, %q, %q", id, tName, s)
	} else if _, ok := err.(*attributeValueTypeError); !ok {
		t.Errorf("Expected *attributeValueTypeError error but got %T (%s)", err, err)
	}

	expect = "45679.23"
	iv = MakeIntegerValue(45678)
	fv = MakeFloatValue(1.23)
	e = makeFunctionFloatAdd(fv, iv)
	a = Attribute{
		id: "test-id",
		t:  TypeFloat}

	ae = MakeAttributeAssignmentExpression(a, e)
	id, tName, s, err = ae.Serialize(ctx)
	if err != nil {
		t.Errorf("Expected no error but got %s", err)
	} else if id != a.id || tName != TypeKeys[a.t] || s != expect {
		t.Errorf("Expected %q, %q, %q but got %q, %q, %q", a.id, TypeKeys[a.t], expect, id, tName, s)
	}

	v = MakeIntegerValue(12345)
	a = Attribute{
		id: "test-id",
		t:  TypeFloat}
	ae = MakeAttributeAssignmentExpression(a, v)
	id, tName, s, err = ae.Serialize(ctx)
	if err == nil {
		t.Errorf("Expected error but got %q, %q, %q", id, tName, s)
	} else if _, ok := err.(*assignmentTypeMismatch); !ok {
		t.Errorf("Expected *ssignmentTypeMismatch error but got %T (%s)", err, err)
	}

	v = undefinedValue
	a = Attribute{
		id: "test-id",
		t:  TypeUndefined}
	ae = MakeAttributeAssignmentExpression(a, v)
	id, tName, s, err = ae.Serialize(ctx)
	if err == nil {
		t.Errorf("Expected error but got %q, %q, %q", id, tName, s)
	} else if _, ok := err.(*invalidTypeSerializationError); !ok {
		t.Errorf("Expected *invalidTypeSerializationError error but got %T (%s)", err, err)
	}
}

func TestAttributeDesignator(t *testing.T) {
	testAttributes := []struct {
		id  string
		val AttributeValue
	}{
		{
			id:  "test-id",
			val: MakeStringValue("test-value"),
		},
		{
			id:  "test-id-i",
			val: MakeIntegerValue(12345),
		},
		{
			id:  "test-id-f",
			val: MakeFloatValue(67.89),
		},
	}

	ctx, err := NewContext(nil, len(testAttributes), func(i int) (string, AttributeValue, error) {
		return testAttributes[i].id, testAttributes[i].val, nil
	})
	if err != nil {
		t.Fatalf("Expected context but got error %s", err)
	}

	a := Attribute{
		id: "test-id",
		t:  TypeString}
	d := MakeAttributeDesignator(a)
	dat := d.GetResultType()
	if dat != TypeString {
		t.Errorf("Expected %q type but got %q", TypeNames[TypeString], TypeNames[dat])
	}

	_, err = d.Calculate(ctx)
	if err != nil {
		t.Errorf("Expected no error but got %s", err)
	}

	i := Attribute{
		id: "test-id-i",
		t:  TypeInteger}
	d = MakeAttributeDesignator(i)
	dat = d.GetResultType()
	if dat != TypeInteger {
		t.Errorf("Expected %q type but got %q", TypeNames[TypeInteger], TypeNames[dat])
	}

	_, err = d.Calculate(ctx)
	if err != nil {
		t.Errorf("Expected no error but got %s", err)
	}

	f := Attribute{
		id: "test-id-f",
		t:  TypeFloat}
	d = MakeAttributeDesignator(f)
	dat = d.GetResultType()
	if dat != TypeFloat {
		t.Errorf("Expected %q type but got %q", TypeNames[TypeFloat], TypeNames[dat])
	}

	_, err = d.Calculate(ctx)
	if err != nil {
		t.Errorf("Expected no error but got %s", err)
	}
}

func assertStringValue(v AttributeValue, err error, e string, desc string, t *testing.T) {
	if err != nil {
		t.Errorf("Expected no error for %s but got %s", desc, err)
		return
	}

	s, err := v.str()
	if err != nil {
		t.Errorf("Expected string value for %s but got error \"%s\"", desc, err)
	}

	if s != e {
		t.Errorf("Expected %q for %s but got %q", e, desc, s)
	}
}
