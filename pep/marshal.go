package pep

import (
	"errors"
	"fmt"
	"net"
	"reflect"
	"strconv"
	"strings"
	"sync"

	pb "github.com/infobloxopen/themis/pdp-service"
)

var (
	// ErrorInvalidSource indicates that input value of validate method is not
	// a structure.
	ErrorInvalidSource = errors.New("given value is not a structure")
	// ErrorInvalidSlice indicates that input structure has slice field
	// (client can't marshal slices).
	ErrorInvalidSlice = errors.New("marshalling for the slice hasn't been implemented")
	// ErrorInvalidStruct indicates that input structure has struct field
	// (client can't marshal nested structures).
	ErrorInvalidStruct = errors.New("marshalling for the struct hasn't been implemented")
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

type reqFieldInfo struct {
	idx        int
	tag        string
	marshaller fieldMarshaller
}

type reqFieldsInfo struct {
	fields []reqFieldInfo
	err    error
}

func makeTaggetFieldsInfo(fields []reflect.StructField, typeName string) reqFieldsInfo {
	out := []reqFieldInfo{}
	for i, f := range fields {
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
				return reqFieldsInfo{err: fmt.Errorf("unknown type \"%s\" (%s.%s)", t, typeName, f.Name)}
			}

			if typeByTag[strings.ToLower(t)] != f.Type {
				return reqFieldsInfo{
					err: fmt.Errorf("can't marshal \"%s\" as \"%s\" (%s.%s)", f.Type.String(), t, typeName, f.Name),
				}
			}

		} else {
			marshaller, ok = marshallersByKind[f.Type.Kind()]
			if !ok {
				return reqFieldsInfo{err: fmt.Errorf("can't marshal \"%s\" (%s.%s)", f.Type.String(), typeName, f.Name)}
			}
		}

		if len(tag) <= 0 {
			tag, ok = getName(f)
			if !ok {
				continue
			}
		}

		out = append(out, reqFieldInfo{
			idx:        i,
			tag:        tag,
			marshaller: marshaller,
		})
	}

	return reqFieldsInfo{fields: out}
}

func makeUntaggedFieldsInfo(fields []reflect.StructField, typeName string) reqFieldsInfo {
	out := []reqFieldInfo{}
	for i, f := range fields {
		name, ok := getName(f)
		if !ok {
			continue
		}

		marshaller, ok := marshallersByKind[f.Type.Kind()]
		if !ok {
			continue
		}

		out = append(out, reqFieldInfo{
			idx:        i,
			tag:        name,
			marshaller: marshaller,
		})
	}

	return reqFieldsInfo{fields: out}
}

var (
	typeCache     = map[string]reqFieldsInfo{}
	typeCacheLock = sync.RWMutex{}
)

func makeRequest(v interface{}) (pb.Request, error) {
	if req, ok := v.(pb.Request); ok {
		return req, nil
	}
	attrs, err := marshalValue(reflect.ValueOf(v))
	if err != nil {
		return pb.Request{}, err
	}

	return pb.Request{Attributes: attrs}, nil
}

func marshalValue(v reflect.Value) ([]*pb.Attribute, error) {
	if v.Kind() != reflect.Struct {
		return nil, ErrorInvalidSource
	}

	return marshalStruct(v, getFields(v.Type()))
}

func getFields(t reflect.Type) reqFieldsInfo {
	key := t.PkgPath() + "." + t.Name()
	typeCacheLock.RLock()
	if info, ok := typeCache[key]; ok {
		typeCacheLock.RUnlock()
		return info
	}
	typeCacheLock.RUnlock()

	fields := make([]reflect.StructField, 0)
	tagged := false
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)

		_, ok := getTag(f)
		tagged = tagged || ok

		fields = append(fields, f)
	}

	typeCacheLock.Lock()
	var info reqFieldsInfo
	if tagged {
		info = makeTaggetFieldsInfo(fields, t.Name())
	} else {
		info = makeUntaggedFieldsInfo(fields, t.Name())
	}
	typeCache[key] = info
	typeCacheLock.Unlock()

	return info
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

func marshalStruct(v reflect.Value, info reqFieldsInfo) ([]*pb.Attribute, error) {
	if info.err != nil {
		return nil, info.err
	}

	attrs := make([]*pb.Attribute, len(info.fields))
	i := 0
	for _, f := range info.fields {
		s, t, err := f.marshaller(v.Field(f.idx))
		if err != nil {
			if err == ErrorInvalidStruct || err == ErrorInvalidSlice {
				continue
			}

			return nil, err
		}

		attrs[i] = &pb.Attribute{Id: f.tag, Type: t, Value: s}
		i++
	}

	return attrs[:i], nil
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
	return (&net.IPNet{IP: v.Field(0).Bytes(), Mask: v.Field(1).Bytes()}).String(), "network", nil
}

func domainMarshaller(v reflect.Value) (string, string, error) {
	return v.String(), "domain", nil
}
