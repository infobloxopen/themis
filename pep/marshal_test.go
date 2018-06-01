package pep

import (
	"bytes"
	"math"
	"net"
	"reflect"
	"strings"
	"testing"

	"github.com/infobloxopen/go-trees/domain"

	"github.com/infobloxopen/themis/pdp"
	pb "github.com/infobloxopen/themis/pdp-service"
)

type DummyStruct struct {
}

type TestStruct struct {
	Bool    bool
	Int     int
	Float   float64
	String  string
	If      interface{}
	Address net.IP
	hidden  string
	Network *net.IPNet
	Slice   []int
	Struct  DummyStruct
	Domain  domain.Name
}

type TestTaggedStruct struct {
	Bool1   bool
	Bool2   bool        `pdp:""`
	Bool3   bool        `pdp:"flag"`
	Int     int         `pdp:"i,integer"`
	Float   float64     `pdp:"f,float"`
	Domain  domain.Name `pdp:"d,domain"`
	Address net.IP      `pdp:""`
	network *net.IPNet  `pdp:"net,network"`
}

type TestInvalidStruct1 struct {
	String string `pdp:",address"`
}

type TestInvalidStruct2 struct {
	If interface{} `pdp:""`
}

var (
	testStruct = TestStruct{
		Bool:    true,
		Int:     5,
		Float:   555.5,
		String:  "test",
		If:      "interface",
		Address: net.ParseIP("1.2.3.4"),
		hidden:  "hide",
		Network: makeTestNetwork("1.2.3.4/32"),
		Slice:   []int{1, 2, 3, 4},
		Struct:  DummyStruct{},
		Domain:  makeTestDomain("example.com"),
	}

	testRequestBuffer = []byte{
		1, 0,
		7, 0,
		4, 'B', 'o', 'o', 'l', 1,
		3, 'I', 'n', 't', 3, 5, 0, 0, 0, 0, 0, 0, 0,
		5, 'F', 'l', 'o', 'a', 't', 4, 0, 0, 0, 0, 0, 92, 129, 64,
		6, 'S', 't', 'r', 'i', 'n', 'g', 2, 4, 0, 't', 'e', 's', 't',
		7, 'A', 'd', 'd', 'r', 'e', 's', 's', 5, 1, 2, 3, 4,
		7, 'N', 'e', 't', 'w', 'o', 'r', 'k', 7, 32, 1, 2, 3, 4,
		6, 'D', 'o', 'm', 'a', 'i', 'n', 9, 11, 0, 'e', 'x', 'a', 'm', 'p', 'l', 'e', '.', 'c', 'o', 'm',
	}
)

func TestMarshalUntaggedStruct(t *testing.T) {
	var b [100]byte

	n, err := marshalValue(reflect.ValueOf(testStruct), b[:])
	assertBytesBuffer(t, "marshalValue(TestStruct)", err, b[:], n, testRequestBuffer...)
}

func TestMarshalTaggedStruct(t *testing.T) {
	var b [78]byte

	v := TestTaggedStruct{
		Bool1:   true,
		Bool2:   false,
		Bool3:   true,
		Int:     math.MaxInt32,
		Float:   12345.6789,
		Domain:  makeTestDomain("example.com"),
		Address: net.ParseIP("1.2.3.4"),
		network: makeTestNetwork("1.2.3.4/32"),
	}

	n, err := marshalValue(reflect.ValueOf(v), b[:])
	assertBytesBuffer(t, "marshalValue(TestTaggedStruct)", err, b[:], n,
		1, 0,
		7, 0,
		5, 'B', 'o', 'o', 'l', '2', 0,
		4, 'f', 'l', 'a', 'g', 1,
		1, 'i', 3, 255, 255, 255, 127, 00, 00, 00, 00,
		1, 'f', 4, 161, 248, 49, 230, 214, 28, 200, 64,
		1, 'd', 9, 11, 0, 'e', 'x', 'a', 'm', 'p', 'l', 'e', '.', 'c', 'o', 'm',
		7, 'A', 'd', 'd', 'r', 'e', 's', 's', 5, 1, 2, 3, 4,
		3, 'n', 'e', 't', 7, 32, 1, 2, 3, 4,
	)
}

func TestMarshalInvalidStructs(t *testing.T) {
	var b [100]byte

	n, err := marshalValue(reflect.ValueOf(TestInvalidStruct1{}), b[:])
	if err == nil {
		t.Errorf("exepcted \"can't marshal\" error but got %d bytes in buffer: [% x]", n, b[:n])
	} else if !strings.Contains(err.Error(), "can't marshal") {
		t.Errorf("Exepcted \"can't marshal\" error but got:\n%s", err)
	}

	n, err = marshalValue(reflect.ValueOf(TestInvalidStruct2{}), b[:])
	if err == nil {
		t.Errorf("exepcted \"can't marshal\" error but got %d bytes in buffer: [% x]", n, b[:n])
	} else if !strings.Contains(err.Error(), "can't marshal") {
		t.Errorf("Exepcted \"can't marshal\" error but got:\n%s", err)
	}
}

func TestMakeRequest(t *testing.T) {
	var b [100]byte

	m, err := makeRequest(pb.Msg{Body: testRequestBuffer}, b[:])
	assertBytesBuffer(t, "makeRequest(pb.Msg)", err, m.Body, len(m.Body), testRequestBuffer...)

	m, err = makeRequest(&pb.Msg{Body: testRequestBuffer}, b[:])
	assertBytesBuffer(t, "makeRequest(&pb.Msg)", err, m.Body, len(m.Body), testRequestBuffer...)

	m, err = makeRequest(testRequestBuffer, b[:])
	assertBytesBuffer(t, "makeRequest(testRequestBuffer)", err, m.Body, len(m.Body), testRequestBuffer...)

	m, err = makeRequest(testStruct, b[:])
	assertBytesBuffer(t, "makeRequest(testStruct)", err, m.Body, len(m.Body), testRequestBuffer...)

	m, err = makeRequest([]pdp.AttributeAssignment{
		pdp.MakeBooleanAssignment("Bool", true),
		pdp.MakeIntegerAssignment("Int", 5),
		pdp.MakeFloatAssignment("Float", 555.5),
		pdp.MakeStringAssignment("String", "test"),
		pdp.MakeAddressAssignment("Address", net.ParseIP("1.2.3.4")),
		pdp.MakeNetworkAssignment("Network", makeTestNetwork("1.2.3.4/32")),
		pdp.MakeDomainAssignment("Domain", makeTestDomain("example.com")),
	}, b[:])
	assertBytesBuffer(t, "makeRequest(assignments)", err, m.Body, len(m.Body), testRequestBuffer...)
}

func makeTestNetwork(s string) *net.IPNet {
	_, n, err := net.ParseCIDR(s)
	if err != nil {
		panic(err)
	}

	return n
}

func makeTestDomain(s string) domain.Name {
	d, err := domain.MakeNameFromString(s)
	if err != nil {
		panic(err)
	}

	return d
}

func assertBytesBuffer(t *testing.T, desc string, err error, b []byte, n int, e ...byte) {
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
