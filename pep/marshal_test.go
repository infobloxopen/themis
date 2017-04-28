package pep

import (
	"fmt"
	"net"
	"reflect"
	"strings"
	"testing"

	pb "github.com/infobloxopen/themis/pdp-service"
)

type DummyStruct struct {
}

type TestStruct struct {
	Bool    bool
	Int     int
	String  string
	If      interface{}
	Address net.IP
	hidden  string
	Network net.IPNet
	Slice   []int
	Struct  DummyStruct
}

type TestTaggedStruct struct {
	Bool1   bool
	Bool2   bool      `pdp`
	Bool3   bool      `pdp:"flag"`
	Domain  string    `pdp:"d,domain"`
	Address net.IP    `pdp`
	network net.IPNet `pdp:"net,network"`
}

type TestInvalidStruct1 struct {
	Int int `pdp:"number,integer"`
}

type TestInvalidStruct2 struct {
	String string `pdp:",address"`
}

type TestInvalidStruct3 struct {
	If interface{} `pdp`
}

var (
	TestAttributes = []*pb.Attribute{
		{"Bool", "boolean", "true"},
		{"String", "string", "test"},
		{"Address", "address", "1.2.3.4"},
		{"Network", "network", "1.2.3.4/32"}}

	TestTaggedAttributes = []*pb.Attribute{
		{"Bool2", "boolean", "false"},
		{"flag", "boolean", "true"},
		{"d", "domain", "example.com"},
		{"Address", "address", "1.2.3.4"},
		{"net", "network", "1.2.3.4/32"}}
)

func TestMarshalUntaggedStruct(t *testing.T) {
	_, n, _ := net.ParseCIDR("1.2.3.4/32")
	v := TestStruct{true, 5, "test", "interface", net.ParseIP("1.2.3.4"), "hide", *n, []int{1, 2, 3, 4}, DummyStruct{}}
	attrs, err := marshalValue(reflect.ValueOf(v))
	if err != nil {
		t.Errorf("Expected no error but got: %s", err)
	} else {
		CompareAttributes(attrs, TestAttributes, t)
	}
}

func TestMarshalTaggedStruct(t *testing.T) {
	_, n, _ := net.ParseCIDR("1.2.3.4/32")
	v := TestTaggedStruct{true, false, true, "example.com", net.ParseIP("1.2.3.4"), *n}
	attrs, err := marshalValue(reflect.ValueOf(v))
	if err != nil {
		t.Errorf("Expected no error but got: %s", err)
	} else {
		CompareAttributes(attrs, TestTaggedAttributes, t)
	}
}

func TestMarshalInvalidStructs(t *testing.T) {
	v1 := TestInvalidStruct1{}
	_, err := marshalValue(reflect.ValueOf(v1))
	if err != nil {
		if !strings.Contains(err.Error(), "Unknown type") {
			t.Errorf("Exepcted \"Unknown type\" error but got:\n%s", err)
		}
	} else {
		t.Errorf("Exepcted \"Unknown type\" error")
	}

	v2 := TestInvalidStruct2{}
	_, err = marshalValue(reflect.ValueOf(v2))
	if err != nil {
		if !strings.Contains(err.Error(), "Can't marshal") {
			t.Errorf("Exepcted \"Can't marshal\" error but got:\n%s", err)
		}
	} else {
		t.Errorf("Exepcted \"Can't marshal\" error")
	}

	v3 := TestInvalidStruct3{}
	_, err = marshalValue(reflect.ValueOf(v3))
	if err != nil {
		if !strings.Contains(err.Error(), "Can't marshal") {
			t.Errorf("Exepcted \"Can't marshal\" error but got:\n%s", err)
		}
	} else {
		t.Errorf("Exepcted \"Can't marshal\" error")
	}
}

func CompareAttributes(v, e []*pb.Attribute, t *testing.T) {
	isVEmpty := v == nil || len(v) <= 0
	isEEmpty := e == nil || len(e) <= 0
	if isVEmpty != isEEmpty {
		t.Errorf("Expected:\n%s\nbut got:\n%s\n", SprintfAttributes(e), SprintfAttributes(v))
		return
	}

	if len(v) != len(e) {
		t.Errorf("Expected:\n%s\nbut got:\n%s\n", SprintfAttributes(e), SprintfAttributes(v))
		return
	}

	for i, a := range v {
		if *a != *e[i] {
			t.Errorf("Expected:\n%s\nbut got:\n%s\n", SprintfAttributes(e), SprintfAttributes(v))
			return
		}
	}
}

func SprintfAttributes(v []*pb.Attribute) string {
	if v == nil {
		return "<nil>"
	}

	if len(v) <= 0 {
		return "<empty>"
	}

	strs := make([]string, len(v))
	for i, attr := range v {
		strs[i] = fmt.Sprintf("\t%s: %s (%s)", attr.Id, attr.Value, attr.Type)
	}

	return strings.Join(strs, "\n")
}
