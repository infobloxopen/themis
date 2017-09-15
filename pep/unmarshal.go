package pep

import (
	"errors"
	"fmt"
	"net"
	"reflect"
	"strconv"
	"strings"

	"github.com/infobloxopen/themis/pdp"
	pb "github.com/infobloxopen/themis/pdp-service"
)

const (
	effectFieldName = "Effect"
	reasonFieldName = "Reason"
)

var (
	// ErrorInvalidDestination indicates that output value of validate method is
	// not a structure.
	ErrorInvalidDestination = errors.New("given value is not a pointer to structure")
)

func unmarshalToValue(res *pb.Response, v reflect.Value) error {
	if v.Kind() != reflect.Ptr {
		return ErrorInvalidDestination
	}

	v = v.Elem()
	if v.Kind() != reflect.Struct {
		return ErrorInvalidDestination
	}

	fields, err := makeFieldMap(v.Type())
	if err != nil {
		return err
	}

	if len(fields) > 0 {
		return unmarshalToTaggedStruct(res, v, fields)
	}

	return unmarshalToUntaggedStruct(res, v)
}

func parseTag(tag string, f reflect.StructField, t reflect.Type) (string, error) {
	items := strings.Split(tag, ",")
	if len(items) > 1 {
		tag = items[0]
		taggedTypeName := items[1]

		if tag == effectFieldName || tag == reasonFieldName {
			return "", fmt.Errorf("don't support type definition for \"%s\" and \"%s\" fields (%s.%s)",
				effectFieldName, reasonFieldName, t.Name(), f.Name)
		}

		taggedType, ok := typeByTag[strings.ToLower(taggedTypeName)]
		if !ok {
			return "", fmt.Errorf("unknown type \"%s\" (%s.%s)", taggedTypeName, t.Name(), f.Name)
		}

		if taggedType != f.Type {
			return "", fmt.Errorf("tagged type \"%s\" doesn't match field type \"%s\" (%s.%s)",
				taggedTypeName, f.Type.Name(), t.Name(), f.Name)
		}

		return tag, nil
	}

	if tag == effectFieldName {
		return effectFieldName, nil
	}

	if tag == reasonFieldName {
		return reasonFieldName, nil
	}

	return tag, nil
}

func makeFieldMap(t reflect.Type) (map[string]string, error) {
	m := make(map[string]string)

	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)

		tag, ok := getTag(f)
		if !ok {
			continue
		}

		if len(tag) <= 0 {
			tag, ok = getName(f)
			if !ok {
				continue
			}
		}

		tag, err := parseTag(tag, f, t)
		if err != nil {
			return nil, err
		}

		m[tag] = f.Name
	}

	return m, nil
}

type fieldUnmarshaller func(attr *pb.Attribute, v reflect.Value) error

var unmarshallersByType = map[string]fieldUnmarshaller{
	pdp.TypeKeys[pdp.TypeBoolean]: boolUnmarshaller,
	pdp.TypeKeys[pdp.TypeString]:  stringUnmarshaller,
	pdp.TypeKeys[pdp.TypeAddress]: addressUnmarshaller,
	pdp.TypeKeys[pdp.TypeNetwork]: networkUnmarshaller,
	pdp.TypeKeys[pdp.TypeDomain]:  domainUnmarshaller}

func unmarshalToTaggedStruct(res *pb.Response, v reflect.Value, fields map[string]string) error {
	name, ok := fields[effectFieldName]
	if ok {
		setToUntaggedEffect(res, v, name)
	}

	name, ok = fields[reasonFieldName]
	if ok {
		setToUntaggedReason(res, v, name)
	}

	for _, attr := range res.Obligation {
		name, ok := fields[attr.Id]
		if !ok {
			continue
		}

		f := v.FieldByName(name)
		if !f.CanSet() {
			return fmt.Errorf("field %s.%s is tagged but can't be set", v.Type().Name(), name)
		}

		unmarshaller, ok := unmarshallersByType[attr.Type]
		if !ok {
			return fmt.Errorf("can't unmarshal \"%s\" of \"%s\" type", attr.Id, attr.Type)
		}

		if t, ok := typeByTag[attr.Type]; ok {
			if t != f.Type() {
				return fmt.Errorf("can't unmarshal \"%s\" of \"%s\" type to field %s.%s",
					attr.Id, attr.Type, v.Type().Name(), name)
			}
		} else {
			return fmt.Errorf("can't unmarshal \"%s\" of \"%s\" type", attr.Id, attr.Type)
		}

		err := unmarshaller(attr, f)
		if err != nil {
			return err
		}
	}

	return nil
}

func setToUntaggedEffect(res *pb.Response, v reflect.Value, name string) bool {
	f := v.FieldByName(name)
	if !f.CanSet() {
		return false
	}

	k := f.Kind()
	if k == reflect.Bool {
		f.SetBool(res.Effect == pb.Response_PERMIT)
		return true
	}

	if k == reflect.Int || k == reflect.Int8 || k == reflect.Int16 || k == reflect.Int32 || k == reflect.Int64 {
		f.SetInt(int64(res.Effect))
		return true
	}

	if k == reflect.Uint || k == reflect.Uint8 || k == reflect.Uint16 || k == reflect.Uint32 || k == reflect.Uint64 {
		f.SetUint(uint64(res.Effect))
		return true
	}

	if k == reflect.String {
		f.SetString(pb.Response_Effect_name[int32(res.Effect)])
		return true
	}

	return false
}

func setToUntaggedReason(res *pb.Response, v reflect.Value, name string) bool {
	f := v.FieldByName(name)
	if !f.CanSet() {
		return false
	}

	if f.Kind() == reflect.String {
		f.SetString(res.Reason)
		return true
	}

	return false
}

func unmarshalToUntaggedStruct(res *pb.Response, v reflect.Value) error {
	skipEffect := setToUntaggedEffect(res, v, effectFieldName)
	skipReason := setToUntaggedReason(res, v, reasonFieldName)

	for _, attr := range res.Obligation {
		if attr.Id == effectFieldName && skipEffect {
			continue
		}

		if attr.Id == reasonFieldName && skipReason {
			continue
		}

		f := v.FieldByName(attr.Id)
		if !f.CanSet() {
			continue
		}

		unmarshaller, ok := unmarshallersByType[attr.Type]
		if !ok {
			return fmt.Errorf("can't unmarshal \"%s\" of \"%s\" type", attr.Id, attr.Type)
		}

		if t, ok := typeByTag[attr.Type]; !ok || t != f.Type() {
			continue
		}

		err := unmarshaller(attr, f)
		if err != nil {
			return err
		}
	}

	return nil
}

func boolUnmarshaller(attr *pb.Attribute, v reflect.Value) error {
	b, err := strconv.ParseBool(attr.Value)
	if err != nil {
		return fmt.Errorf("can't treat \"%s\" value (%s) as boolean: %s", attr.Id, attr.Value, err)
	}

	v.SetBool(b)
	return nil
}

func stringUnmarshaller(attr *pb.Attribute, v reflect.Value) error {
	v.SetString(attr.Value)
	return nil
}

func addressUnmarshaller(attr *pb.Attribute, v reflect.Value) error {
	s := attr.Value
	if strings.Contains(s, ":") {
		if strings.Contains(s, "]") {
			s = strings.Split(s, "]")[0][1:]
		} else if strings.Contains(s, ".") {
			s = strings.Split(s, ":")[0]
		}
	}

	ip := net.ParseIP(s)
	if ip == nil {
		return fmt.Errorf("can't treat \"%s\" value (%s) as address", attr.Id, attr.Value)
	}

	v.Set(reflect.ValueOf(ip))
	return nil
}

func networkUnmarshaller(attr *pb.Attribute, v reflect.Value) error {
	_, n, err := net.ParseCIDR(attr.Value)
	if err != nil {
		return fmt.Errorf("can't treat \"%s\" value (%s) as network: %s", attr.Id, attr.Value, err)
	}

	v.Set(reflect.ValueOf(*n))
	return nil
}

func domainUnmarshaller(attr *pb.Attribute, v reflect.Value) error {
	v.SetString(attr.Value)
	return nil
}
