package pdp

import (
	"encoding/binary"
	"math"
	"net"
	"reflect"
	"strconv"
	"strings"

	"github.com/infobloxopen/go-trees/domain"
)

const (
	responseStatusTooLong            = "status too long"
	responseStatusObligationsTooLong = "obligations too long"
)

func MarshalResponse(b []byte, effect int, obligations []AttributeAssignmentExpression, errs ...error) (int, error) {
	off, err := putRequestVersion(b)
	if err != nil {
		return off, err
	}

	n, err := putResponseEffect(b[off:], effect)
	if err != nil {
		return off, err
	}
	off += n

	n, err = putResponseStatus(b[off:], errs...)
	if err != nil {
		n, err = putResponseStatusTooLong(b[off:])
		if err != nil {
			return off, err
		}

		return off + n, nil
	}

	off += n

	n, err = putAssignmentExpressions(b[off:], obligations)
	if err != nil {
		if _, ok := err.(*requestBufferOverflowError); ok {
			off, _ := putRequestVersion(b)
			n, _ := putResponseEffect(b[off:], effect)
			off += n

			n, err := putResponseObligationsTooLong(b[off:])
			if err != nil {
				return off, err
			}

			return off + n, nil
		}

		return off, err
	}

	return off + n, nil
}

func UnmarshalResponse(b []byte, out []AttributeAssignmentExpression) (int, int, error) {
	off, err := checkRequestVersion(b)
	if err != nil {
		return EffectIndeterminate, 0, err
	}

	effect, n, err := getResponseEffect(b[off:])
	if err != nil {
		return EffectIndeterminate, 0, err
	}
	off += n

	s, n, err := getRequestStringValue(b[off:])
	if err != nil {
		return EffectIndeterminate, 0, err
	}
	off += n

	n, err = getAssignmentExpressions(b[off:], out)
	if err != nil {
		return EffectIndeterminate, 0, err
	}

	if len(s) > 0 {
		return effect, n, newResponseServerError(s)
	}

	return effect, n, nil
}

func putResponseEffect(b []byte, effect int) (int, error) {
	if effect < 0 || effect >= effectTotalCount {
		return 0, newResponseEffectError(effect)
	}

	if len(b) < 1 {
		return 0, newRequestBufferOverflowError()
	}

	b[0] = byte(effect)
	return 1, nil
}

func getResponseEffect(b []byte) (int, int, error) {
	if len(b) < 1 {
		return EffectIndeterminate, 0, newRequestBufferUnderflowError()
	}

	effect := int(b[0])
	if effect < 0 || effect >= effectTotalCount {
		return EffectIndeterminate, 0, newResponseEffectError(effect)
	}

	return effect, 1, nil
}

func putResponseStatus(b []byte, err ...error) (int, error) {
	if len(err) < 1 || len(err) == 1 && err[0] == nil {
		if len(b) < 2 {
			return 0, newRequestBufferOverflowError()
		}

		binary.LittleEndian.PutUint16(b, 0)
		return 2, nil
	}

	var msg string
	if len(err) == 1 {
		msg = err[0].Error()
	} else {
		msgs := make([]string, len(err))
		for i, err := range err {
			msgs[i] = strconv.Quote(err.Error())
		}

		msg = "multiple errors: " + strings.Join(msgs, ", ")
	}

	if len(msg) > math.MaxUint16 {
		i := 0
		for j := range msg {
			if j > math.MaxUint16 {
				break
			}

			i = j
		}

		msg = msg[:i]
	}

	size := len(msg) + 2
	if len(b) < size {
		return 0, newRequestBufferOverflowError()
	}

	binary.LittleEndian.PutUint16(b, uint16(len(msg)))
	copy(b[2:], msg)

	return size, nil
}

func putResponseStatusTooLong(b []byte) (int, error) {
	size := len(responseStatusTooLong) + 2
	if len(b) < size {
		return 0, newRequestBufferOverflowError()
	}

	binary.LittleEndian.PutUint16(b, uint16(len(responseStatusTooLong)))
	copy(b[2:], responseStatusTooLong)

	return size, nil
}

func putResponseObligationsTooLong(b []byte) (int, error) {
	size := len(responseStatusObligationsTooLong) + 2
	if len(b) < size {
		return 0, newRequestBufferOverflowError()
	}

	binary.LittleEndian.PutUint16(b, uint16(len(responseStatusObligationsTooLong)))
	copy(b[2:], responseStatusObligationsTooLong)

	return size, nil
}

func putAssignmentExpressions(b []byte, in []AttributeAssignmentExpression) (int, error) {
	off, err := putRequestAttributeCount(b, len(in))
	if err != nil {
		return off, err
	}

	for _, a := range in {
		id := a.a.id
		v, ok := a.e.(AttributeValue)
		if !ok {
			return off, newRequestInvalidExpressionError(a)
		}

		n, err := putRequestAttribute(b[off:], id, v)
		if err != nil {
			return off, bindError(err, id)
		}
		off += n
	}

	return off, nil
}

func putAttributesFromReflection(b []byte, c int, f func(i int) (string, Type, reflect.Value, error)) (int, error) {
	off, err := putRequestAttributeCount(b, c)
	if err != nil {
		return off, err
	}

	for i := 0; i < c; i++ {
		id, t, v, err := f(i)
		if err != nil {
			return off, err
		}

		var n int
		switch t {
		default:
			return off, bindError(newRequestAttributeMarshallingNotImplemented(t), id)

		case TypeBoolean:
			n, err = putRequestAttributeBoolean(b[off:], id, v.Bool())

		case TypeString:
			n, err = putRequestAttributeString(b[off:], id, v.String())

		case TypeInteger:
			n, err = putRequestAttributeInteger(b[off:], id, v.Int())

		case TypeFloat:
			n, err = putRequestAttributeFloat(b[off:], id, v.Float())

		case TypeAddress:
			n, err = putRequestAttributeAddress(b[off:], id, net.IP(v.Bytes()))

		case TypeNetwork:
			n, err = putRequestAttributeNetwork(b[off:], id, v.Interface().(*net.IPNet))

		case TypeDomain:
			n, err = putRequestAttributeDomain(b[off:], id, v.Interface().(domain.Name))
		}

		if err != nil {
			return off, bindError(err, id)
		}
		off += n
	}

	return off, nil
}

func getAssignmentExpressions(b []byte, out []AttributeAssignmentExpression) (int, error) {
	c, n, err := getRequestAttributeCount(b)
	if err != nil {
		return 0, err
	}
	b = b[n:]

	if len(out) < c {
		return 0, newRequestAssignmentsOverflowError(c, len(out))
	}

	for i := 0; i < c; i++ {
		id, v, n, err := getRequestAttribute(b)
		if err != nil {
			return 0, bindErrorf(err, "%d", i+1)
		}
		b = b[n:]

		out[i] = MakeAttributeAssignmentExpression(MakeAttribute(id, v.GetResultType()), v)
	}

	return c, nil
}

func getAttributesToReflection(b []byte, f func(string, Type) (reflect.Value, bool, error)) error {
	c, n, err := getRequestAttributeCount(b)
	if err != nil {
		return err
	}
	b = b[n:]

	for i := 0; i < c; i++ {
		id, n, err := getRequestAttributeName(b)
		if err != nil {
			return bindErrorf(err, "%d", i+1)
		}
		b = b[n:]

		t, n, err := getRequestAttributeType(b)
		if err != nil {
			return bindError(err, id)
		}
		b = b[n:]

		if t == requestWireTypeSetOfStrings || t == requestWireTypeSetOfNetworks || t == requestWireTypeSetOfDomains || t == requestWireTypeListOfStrings {
			return bindError(newRequestAttributeUnmarshallingNotImplemented(t), id)
		}

		if t < 0 || t >= len(builtinTypeByWire) {
			return bindError(newRequestAttributeUnmarshallingTypeError(t), id)
		}

		v, ok, err := f(id, builtinTypeByWire[t])
		if err != nil {
			return err
		}

		switch t {
		case requestWireTypeBooleanFalse:
			if ok {
				v.SetBool(false)
			}

		case requestWireTypeBooleanTrue:
			if ok {
				v.SetBool(true)
			}

		case requestWireTypeString:
			s, n, err := getRequestStringValue(b)
			if err != nil {
				return bindError(err, id)
			}
			b = b[n:]

			if ok {
				v.SetString(s)
			}

		case requestWireTypeInteger:
			i, n, err := getRequestIntegerValue(b)
			if err != nil {
				return bindError(err, id)
			}
			b = b[n:]

			if ok {
				if err := setInt(v, i); err != nil {
					return bindError(err, id)
				}
			}

		case requestWireTypeFloat:
			f, n, err := getRequestFloatValue(b)
			if err != nil {
				return bindError(err, id)
			}
			b = b[n:]

			if ok {
				v.SetFloat(f)
			}

		case requestWireTypeIPv4Address:
			a, n, err := getRequestIPv4AddressValue(b)
			if err != nil {
				return bindError(err, id)
			}
			b = b[n:]

			if ok {
				v.Set(reflect.ValueOf(a))
			}

		case requestWireTypeIPv6Address:
			a, n, err := getRequestIPv6AddressValue(b)
			if err != nil {
				return bindError(err, id)
			}
			b = b[n:]

			if ok {
				v.Set(reflect.ValueOf(a))
			}

		case requestWireTypeIPv4Network:
			ip, n, err := getRequestIPv4NetworkValue(b)
			if err != nil {
				return bindError(err, id)
			}
			b = b[n:]

			if ok {
				v.Set(reflect.ValueOf(ip))
			}

		case requestWireTypeIPv6Network:
			ip, n, err := getRequestIPv6NetworkValue(b)
			if err != nil {
				return bindError(err, id)
			}
			b = b[n:]

			if ok {
				v.Set(reflect.ValueOf(ip))
			}

		case requestWireTypeDomain:
			d, n, err := getRequestDomainValue(b)
			if err != nil {
				return bindError(err, id)
			}
			b = b[n:]

			if ok {
				v.Set(reflect.ValueOf(d))
			}
		}
	}

	return nil
}
