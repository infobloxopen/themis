package pep

import (
	"fmt"
	"net"
	"reflect"
	"strings"
	"testing"

	pb "github.com/infobloxopen/themis/pdp-service"
)

type TestResponseStruct struct {
	Effect  bool
	Int     int
	Bool    bool
	String  string
	Address net.IP
	Network net.IPNet
}

type TestTaggedResponseStruct struct {
	Result  string `pdp:"Effect"`
	Error   string `pdp:"Reason"`
	Bool1   bool
	Bool2   bool      `pdp`
	Bool3   bool      `pdp:"flag"`
	Domain  string    `pdp:"d,domain"`
	Address net.IP    `pdp`
	Network net.IPNet `pdp:"net,network"`
}

type TestInvalidResponseStruct1 struct {
	Effect bool `pdp:"Effect,string"`
}

type TestInvalidResponseStruct2 struct {
	Reason string `pdp:"Reason,string"`
}

type TestInvalidResponseStruct3 struct {
	Attribute string `pdp:",unknown"`
}

type TestInvalidResponseStruct4 struct {
	Attribute bool `pdp:",address"`
}

type TestInvalidResponseStruct5 struct {
	flag bool `pdp:"flag"`
}

var (
	TestObligations = []*pb.Attribute{
		{"Bool", "boolean", "true"},
		{"String", "string", "test"},
		{"Address", "address", "1.2.3.4"},
		{"Network", "network", "1.2.3.4/32"}}

	TestTaggedObligations = []*pb.Attribute{
		{"Bool2", "boolean", "false"},
		{"flag", "boolean", "true"},
		{"d", "domain", "example.com"},
		{"Address", "address", "1.2.3.4"},
		{"net", "network", "1.2.3.4/32"}}

	TestInvalidObligations1 = []*pb.Attribute{
		{"Bool", "boolean", "unknown"},
		{"String", "string", "test"},
		{"Address", "address", "1.2.3.4"},
		{"Network", "network", "1.2.3.4/32"}}

	TestInvalidObligations2 = []*pb.Attribute{
		{"Bool", "boolean", "false"},
		{"String", "string", "test"},
		{"Address", "address", "1.2:3.4"},
		{"Network", "network", "1.2.3.4/32"}}

	TestInvalidObligations3 = []*pb.Attribute{
		{"Bool", "boolean", "false"},
		{"String", "string", "test"},
		{"Address", "address", "1.2.3.4"},
		{"Network", "network", "1.2.3.4/77"}}

	TestInvalidObligations4 = []*pb.Attribute{
		{"Bool", "boolean", "false"},
		{"String", "long", "test"},
		{"Address", "address", "1.2.3.4"},
		{"Network", "network", "1.2.3.4/77"}}

	TestInvalidObligations5 = []*pb.Attribute{
		{"Bool2", "boolean", "false"},
		{"flag", "long", "true"},
		{"d", "domain", "example.com"},
		{"Address", "address", "1.2.3.4"},
		{"net", "network", "1.2.3.4/32"}}

	TestInvalidObligations6 = []*pb.Attribute{
		{"Bool2", "boolean", "false"},
		{"flag", "network", "true"},
		{"d", "domain", "example.com"},
		{"Address", "address", "1.2.3.4"},
		{"net", "network", "1.2.3.4/32"}}
)

func TestUnmarshalUntaggedStruct(t *testing.T) {
	r := &pb.Response{pb.Response_PERMIT, "", TestObligations}
	v := TestResponseStruct{}

	err := unmarshalToValue(r, reflect.ValueOf(&v))
	if err != nil {
		t.Errorf("Expected no error but got: %s", err)
	} else {
		_, n, _ := net.ParseCIDR("1.2.3.4/32")
		CompareTestResponseStruct(v, TestResponseStruct{true, 0, true, "test", net.ParseIP("1.2.3.4"), *n}, t)
	}
}

func TestUnmarshalTaggedStruct(t *testing.T) {
	r := &pb.Response{pb.Response_INDETERMINATED, "Test Error!", TestTaggedObligations}
	v := TestTaggedResponseStruct{}

	err := unmarshalToValue(r, reflect.ValueOf(&v))
	if err != nil {
		t.Errorf("Expected no error but got: %s", err)
	} else {
		_, n, _ := net.ParseCIDR("1.2.3.4/32")
		e := TestTaggedResponseStruct{
			pb.Response_Effect_name[int32(pb.Response_INDETERMINATED)],
			"Test Error!",
			false, false, true,
			"example.com",
			net.ParseIP("1.2.3.4"), *n}
		CompareTestTaggedStruct(v, e, t)
	}
}

type TestIntResponse struct {
	Effect int8
}

type TestUintResponse struct {
	Effect uint8
}

func TestUnmarshalEffectTypes(t *testing.T) {
	r := &pb.Response{pb.Response_INDETERMINATEDP, "", nil}

	v1 := TestIntResponse{}
	err := unmarshalToValue(r, reflect.ValueOf(&v1))
	if err != nil {
		t.Errorf("Expected no error but got: %s", err)
	} else if v1.Effect != int8(pb.Response_INDETERMINATEDP) {
		t.Errorf("Expected %d effect but got: %d", pb.Response_INDETERMINATEDP, v1.Effect)
	}

	v2 := TestUintResponse{}
	err = unmarshalToValue(r, reflect.ValueOf(&v2))
	if err != nil {
		t.Errorf("Expected no error but got: %s", err)
	} else if v2.Effect != uint8(pb.Response_INDETERMINATEDP) {
		t.Errorf("Expected %d effect but got: %d", pb.Response_INDETERMINATEDP, v2.Effect)
	}
}

func TestUnmarshalInvalidObligations(t *testing.T) {
	r := &pb.Response{pb.Response_INDETERMINATEDP, "", TestInvalidObligations1}
	v := TestResponseStruct{}
	err := unmarshalToValue(r, reflect.ValueOf(&v))
	if err != nil {
		if !strings.Contains(err.Error(), "Can't treat") {
			t.Errorf("Expected \"Can't treat\" error but got: %s", err)
		}
	} else {
		t.Errorf("Expected \"Can't treat\" error")
	}

	r = &pb.Response{pb.Response_INDETERMINATEDP, "", TestInvalidObligations2}
	v = TestResponseStruct{}
	err = unmarshalToValue(r, reflect.ValueOf(&v))
	if err != nil {
		if !strings.Contains(err.Error(), "Can't treat") {
			t.Errorf("Expected \"Can't treat\" error but got: %s", err)
		}
	} else {
		t.Errorf("Expected \"Can't treat\" error")
	}

	r = &pb.Response{pb.Response_INDETERMINATEDP, "", TestInvalidObligations3}
	v = TestResponseStruct{}
	err = unmarshalToValue(r, reflect.ValueOf(&v))
	if err != nil {
		if !strings.Contains(err.Error(), "Can't treat") {
			t.Errorf("Expected \"Can't treat\" error but got: %s", err)
		}
	} else {
		t.Errorf("Expected \"Can't treat\" error")
	}

	r = &pb.Response{pb.Response_INDETERMINATEDP, "", TestInvalidObligations4}
	v = TestResponseStruct{}
	err = unmarshalToValue(r, reflect.ValueOf(&v))
	if err != nil {
		if !strings.Contains(err.Error(), "Can't unmarshal") {
			t.Errorf("Expected \"Can't unmarshal\" error but got: %s", err)
		}
	} else {
		t.Errorf("Expected \"Can't unmarshal\" error")
	}

	r = &pb.Response{pb.Response_INDETERMINATEDP, "", TestInvalidObligations5}
	v1 := TestTaggedResponseStruct{}
	err = unmarshalToValue(r, reflect.ValueOf(&v1))
	if err != nil {
		if !strings.Contains(err.Error(), "Can't unmarshal") {
			t.Errorf("Expected \"Can't unmarshal\" error but got: %s", err)
		}
	} else {
		t.Errorf("Expected \"Can't unmarshal\" error")
	}

	r = &pb.Response{pb.Response_INDETERMINATEDP, "", TestInvalidObligations6}
	v1 = TestTaggedResponseStruct{}
	err = unmarshalToValue(r, reflect.ValueOf(&v1))
	if err != nil {
		if !strings.Contains(err.Error(), "Can't unmarshal") {
			t.Errorf("Expected \"Can't unmarshal\" error but got: %s", err)
		}
	} else {
		t.Errorf("Expected \"Can't unmarshal\" error")
	}
}

func TestUnmarshalInvalidStructures(t *testing.T) {
	r := &pb.Response{pb.Response_INDETERMINATEDP, "", nil}
	v1 := TestInvalidResponseStruct1{}
	err := unmarshalToValue(r, reflect.ValueOf(&v1))
	if err != nil {
		if !strings.Contains(err.Error(), "Don't support type definition") {
			t.Errorf("Expected \"Don't support type definition\" error but got: %s", err)
		}
	} else {
		t.Errorf("Expected \"Don't support type definition\" error")
	}

	r = &pb.Response{pb.Response_INDETERMINATEDP, "", nil}
	v2 := TestInvalidResponseStruct2{}
	err = unmarshalToValue(r, reflect.ValueOf(&v2))
	if err != nil {
		if !strings.Contains(err.Error(), "Don't support type definition") {
			t.Errorf("Expected \"Don't support type definition\" error but got: %s", err)
		}
	} else {
		t.Errorf("Expected \"Don't support type definition\" error")
	}

	r = &pb.Response{pb.Response_INDETERMINATEDP, "", nil}
	v3 := TestInvalidResponseStruct3{}
	err = unmarshalToValue(r, reflect.ValueOf(&v3))
	if err != nil {
		if !strings.Contains(err.Error(), "Unknown type") {
			t.Errorf("Expected \"Unknown type\" error but got: %s", err)
		}
	} else {
		t.Errorf("Expected \"Unknown type\" error")
	}

	r = &pb.Response{pb.Response_INDETERMINATEDP, "", nil}
	v4 := TestInvalidResponseStruct4{}
	err = unmarshalToValue(r, reflect.ValueOf(&v4))
	if err != nil {
		if !strings.Contains(err.Error(), "Tagged type") {
			t.Errorf("Expected \"Tagged type\" error but got: %s", err)
		}
	} else {
		t.Errorf("Expected \"Tagged type\" error")
	}

	r = &pb.Response{pb.Response_INDETERMINATED, "Test Error!", TestTaggedObligations}
	v5 := TestInvalidResponseStruct5{}
	err = unmarshalToValue(r, reflect.ValueOf(&v5))
	if err != nil {
		if !strings.Contains(err.Error(), "can't be set") {
			t.Errorf("Expected \"can't be set\" error but got: %s", err)
		}
	} else {
		t.Errorf("Expected \"can't be set\" error")
	}
}

func CompareTestResponseStruct(v, e TestResponseStruct, t *testing.T) {
	if v.Effect != e.Effect ||
		v.Int != e.Int ||
		v.Bool != e.Bool ||
		v.String != e.String ||
		v.Address.String() != e.Address.String() ||
		v.Network.String() != e.Network.String() {
		t.Errorf("Expected:\n%v\nbut got:\n%v\n", SprintfTestResponseStruct(e), SprintfTestResponseStruct(v))
	}
}

func CompareTestTaggedStruct(v, e TestTaggedResponseStruct, t *testing.T) {
	if v.Result != e.Result ||
		v.Error != e.Error ||
		v.Bool1 != e.Bool1 ||
		v.Bool2 != e.Bool2 ||
		v.Bool3 != e.Bool3 ||
		v.Domain != e.Domain ||
		v.Address.String() != e.Address.String() ||
		v.Network.String() != e.Network.String() {
		t.Errorf("Expected:\n%v\nbut got:\n%v\n", SprintfTestTaggedStruct(e), SprintfTestTaggedStruct(v))
	}
}

func SprintfTestResponseStruct(v TestResponseStruct) string {
	return fmt.Sprintf("\tEffect: %v\n\tInt: %v\n\tBool: %v\n\tString: %v\n\tAddress: %s\n\tNetwork: %s\n",
		v.Effect, v.Int, v.Bool, v.String, v.Address.String(), v.Network.String())
}

func SprintfTestTaggedStruct(v TestTaggedResponseStruct) string {
	return fmt.Sprintf("\tResult: %v\n\tError: %v\n"+
		"\tBool1: %v\n\tBool2: %v\n\tBool3: %v\n"+
		"\tDomain: %v\n\tAddress: %v\n\tNetwork: %v\n",
		v.Result, v.Error, v.Bool1, v.Bool2, v.Bool3, v.Domain, v.Address.String(), v.Network.String())
}
