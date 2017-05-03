package pep

import (
	"fmt"
	"net"
	"reflect"
	"strconv"
	"strings"

	pb "github.com/infobloxopen/themis/pdp-service"
)

var (
	ErrorInvalidSource = fmt.Errorf("Given value is not a structure")
	ErrorInvalidSlice  = fmt.Errorf("Marshalling for the slice hasn't been implemented")
	ErrorInvalidStruct = fmt.Errorf("Marshalling for the struct hasn't been implemented")
)

type fieldMarshaller func(v reflect.Value) (string, string, error)

var (
	marshallersByKind = map[reflect.Kind]fieldMarshaller{
		reflect.Bool:   boolMarshaller,
		reflect.String: stringMarshaller,
		reflect.Slice:  sliceMarshaller,
		reflect.Struct: structMarshaller}

	marshallersByTag = map[string]fieldMarshaller{
		"boolean": boolMarshaller,
		"string":  stringMarshaller,
		"address": addressMarshaller,
		"network": networkMarshaller,
		"domain":  domainMarshaller}

	netIPType    = reflect.TypeOf(net.IP{})
	netIPNetType = reflect.TypeOf(net.IPNet{})

	typeByTag = map[string]reflect.Type{
		"boolean": reflect.TypeOf(true),
		"string":  reflect.TypeOf("string"),
		"address": netIPType,
		"network": netIPNetType,
		"domain":  reflect.TypeOf("string")}
)

func marshalValue(v reflect.Value) ([]*pb.Attribute, error) {
	if v.Kind() != reflect.Struct {
		return nil, ErrorInvalidSource
	}

	fields, tagged := getFields(v.Type())
	if tagged {
		return marshalTaggedStruct(v, fields)
	}

	return marshalUntaggedStruct(v, fields)
}

func getFields(t reflect.Type) ([]reflect.StructField, bool) {
	fields := make([]reflect.StructField, 0)
	tagged := false
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)

		_, ok := getTag(f)
		tagged = tagged || ok

		fields = append(fields, f)
	}

	return fields, tagged
}

func getName(f reflect.StructField) (string, bool) {
	name := f.Name
	if len(name) <= 0 {
		return "", false
	}

	c := name[0:1]
	if c != strings.ToUpper(c) {
		return "", false
	}

	return name, true
}

func getTag(f reflect.StructField) (string, bool) {
	if f.Tag == "pdp" {
		return "", true
	}

	return f.Tag.Lookup("pdp")
}

func marshalTaggedStruct(v reflect.Value, fields []reflect.StructField) ([]*pb.Attribute, error) {
	attrs := make([]*pb.Attribute, 0)
	for _, f := range fields {
		tag, ok := getTag(f)
		if !ok {
			continue
		}

		var marshaller fieldMarshaller
		items := strings.Split(tag, ",")
		if len(items) > 1 {
			tag = items[0]
			t := items[1]

			marshaller, ok = marshallersByTag[strings.ToLower(t)]
			if !ok {
				return nil, fmt.Errorf("Unknown type \"%s\" (%s.%s)", t, v.Type().Name(), f.Name)
			}

			if typeByTag[strings.ToLower(t)] != f.Type {
				return nil,
					fmt.Errorf("Can't marshal \"%s\" as \"%s\" (%s.%s)", f.Type.String(), t, v.Type().Name(), f.Name)
			}

		} else {
			marshaller, ok = marshallersByKind[f.Type.Kind()]
			if !ok {
				return nil, fmt.Errorf("Can't marshal \"%s\" (%s.%s)", f.Type.String(), v.Type().Name(), f.Name)
			}
		}

		if len(tag) <= 0 {
			tag, ok = getName(f)
			if !ok {
				continue
			}
		}

		s, t, err := marshaller(v.FieldByName(f.Name))
		if err != nil {
			if err == ErrorInvalidStruct || err == ErrorInvalidSlice {
				continue
			}

			return nil, err
		}

		attrs = append(attrs, &pb.Attribute{tag, t, s})
	}

	return attrs, nil
}

func marshalUntaggedStruct(v reflect.Value, fields []reflect.StructField) ([]*pb.Attribute, error) {
	attrs := make([]*pb.Attribute, 0)
	for _, f := range fields {
		marshaller, ok := marshallersByKind[f.Type.Kind()]
		if !ok {
			continue
		}

		name, ok := getName(f)
		if !ok {
			continue
		}

		s, t, err := marshaller(v.FieldByName(name))
		if err != nil {
			if err == ErrorInvalidStruct || err == ErrorInvalidSlice {
				continue
			}

			return nil, err
		}

		attrs = append(attrs, &pb.Attribute{name, t, s})
	}

	return attrs, nil
}

func boolMarshaller(v reflect.Value) (string, string, error) {
	return strconv.FormatBool(v.Bool()), "boolean", nil
}

func stringMarshaller(v reflect.Value) (string, string, error) {
	return v.String(), "string", nil
}

func sliceMarshaller(v reflect.Value) (string, string, error) {
	if v.Type() != netIPType {
		return "", "", ErrorInvalidSlice
	}

	return addressMarshaller(v)
}

func structMarshaller(v reflect.Value) (string, string, error) {
	if v.Type() != netIPNetType {
		return "", "", ErrorInvalidStruct
	}

	return networkMarshaller(v)
}

func addressMarshaller(v reflect.Value) (string, string, error) {
	return net.IP(v.Bytes()).String(), "address", nil
}

func networkMarshaller(v reflect.Value) (string, string, error) {
	return (&net.IPNet{v.Field(0).Bytes(), v.Field(1).Bytes()}).String(), "network", nil
}

func domainMarshaller(v reflect.Value) (string, string, error) {
	return v.String(), "domain", nil
}
