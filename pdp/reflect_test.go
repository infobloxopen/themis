package pdp

import (
	"math"
	"reflect"
	"testing"
)

func TestSetInt(t *testing.T) {
	var i int
	v := reflect.Indirect(reflect.ValueOf(&i))

	if err := setInt(v, math.MinInt32); err != nil {
		t.Error(err)
	} else if i != math.MinInt32 {
		t.Errorf("Expected %d value but got %d", math.MinInt32, i)
	}

	var i8 int8
	v = reflect.Indirect(reflect.ValueOf(&i8))

	if err := setInt(v, math.MinInt8); err != nil {
		t.Error(err)
	} else if i8 != math.MinInt8 {
		t.Errorf("Expected %d value but got %d", math.MinInt8, i8)
	}

	var i16 int16
	v = reflect.Indirect(reflect.ValueOf(&i16))

	if err := setInt(v, math.MinInt16); err != nil {
		t.Error(err)
	} else if i16 != math.MinInt16 {
		t.Errorf("Expected %d value but got %d", math.MinInt16, i16)
	}

	var i32 int32
	v = reflect.Indirect(reflect.ValueOf(&i32))

	if err := setInt(v, math.MinInt32); err != nil {
		t.Error(err)
	} else if i32 != math.MinInt32 {
		t.Errorf("Expected %d value but got %d", math.MinInt32, i32)
	}

	var i64 int64
	v = reflect.Indirect(reflect.ValueOf(&i64))

	if err := setInt(v, math.MinInt64); err != nil {
		t.Error(err)
	} else if i64 != math.MinInt64 {
		t.Errorf("Expected %d value but got %d", math.MinInt64, i64)
	}

	var u uint
	v = reflect.Indirect(reflect.ValueOf(&u))

	if err := setInt(v, math.MaxInt32); err != nil {
		t.Error(err)
	} else if u != math.MaxInt32 {
		t.Errorf("Expected %d value but got %d", math.MaxInt32, u)
	}

	var u8 uint8
	v = reflect.Indirect(reflect.ValueOf(&u8))

	if err := setInt(v, math.MaxInt8); err != nil {
		t.Error(err)
	} else if u8 != math.MaxInt8 {
		t.Errorf("Expected %d value but got %d", math.MaxInt8, u8)
	}

	var u16 uint16
	v = reflect.Indirect(reflect.ValueOf(&u16))

	if err := setInt(v, math.MaxInt16); err != nil {
		t.Error(err)
	} else if u16 != math.MaxInt16 {
		t.Errorf("Expected %d value but got %d", math.MaxInt16, u16)
	}

	var u32 uint32
	v = reflect.Indirect(reflect.ValueOf(&u32))

	if err := setInt(v, math.MaxInt32); err != nil {
		t.Error(err)
	} else if u32 != math.MaxInt32 {
		t.Errorf("Expected %d value but got %d", math.MaxInt32, u32)
	}

	var u64 uint64
	v = reflect.Indirect(reflect.ValueOf(&u64))

	if err := setInt(v, math.MaxInt64); err != nil {
		t.Error(err)
	} else if u64 != math.MaxInt64 {
		t.Errorf("Expected %d value but got %d", math.MaxInt64, u64)
	}
}

func TestSetIntOverflow(t *testing.T) {
	var i int
	v := reflect.Indirect(reflect.ValueOf(&i))

	assertUnmarshalIntegerOverflowError(t, setInt(v, math.MaxInt64), v)

	var i8 int8
	v = reflect.Indirect(reflect.ValueOf(&i8))

	assertUnmarshalIntegerOverflowError(t, setInt(v, math.MaxInt64), v)

	var i16 int16
	v = reflect.Indirect(reflect.ValueOf(&i16))

	assertUnmarshalIntegerOverflowError(t, setInt(v, math.MaxInt64), v)

	var i32 int32
	v = reflect.Indirect(reflect.ValueOf(&i32))

	assertUnmarshalIntegerOverflowError(t, setInt(v, math.MaxInt64), v)

	var u uint
	v = reflect.Indirect(reflect.ValueOf(&u))

	assertUnmarshalIntegerOverflowError(t, setInt(v, math.MaxInt64), v)

	var u8 uint8
	v = reflect.Indirect(reflect.ValueOf(&u8))

	assertUnmarshalIntegerOverflowError(t, setInt(v, math.MaxInt64), v)

	var u16 uint16
	v = reflect.Indirect(reflect.ValueOf(&u16))

	assertUnmarshalIntegerOverflowError(t, setInt(v, math.MaxInt64), v)

	var u32 uint32
	v = reflect.Indirect(reflect.ValueOf(&u32))

	assertUnmarshalIntegerOverflowError(t, setInt(v, math.MaxInt64), v)
}

func TestSetIntUnderflow(t *testing.T) {
	var u uint
	v := reflect.Indirect(reflect.ValueOf(&u))

	assertUnmarshalIntegerUnderflowError(t, setInt(v, -1), v)

	var u8 uint8
	v = reflect.Indirect(reflect.ValueOf(&u8))

	assertUnmarshalIntegerUnderflowError(t, setInt(v, -1), v)

	var u16 uint16
	v = reflect.Indirect(reflect.ValueOf(&u16))

	assertUnmarshalIntegerUnderflowError(t, setInt(v, -1), v)

	var u32 uint32
	v = reflect.Indirect(reflect.ValueOf(&u32))

	assertUnmarshalIntegerUnderflowError(t, setInt(v, -1), v)

	var u64 uint64
	v = reflect.Indirect(reflect.ValueOf(&u64))

	assertUnmarshalIntegerUnderflowError(t, setInt(v, -1), v)
}

func TestSetIntInvalidKind(t *testing.T) {
	var s string
	v := reflect.Indirect(reflect.ValueOf(&s))

	assertPanicWithError(t, func() { setInt(v, 0) }, "setting int value to %s", v.Kind())
}

func assertUnmarshalIntegerOverflowError(t *testing.T, err error, v reflect.Value) {
	if err == nil {
		t.Errorf("Expected *requestUnmarshalIntegerOverflowError but got value %#v", v)
	} else if _, ok := err.(*requestUnmarshalIntegerOverflowError); !ok {
		t.Errorf("Expected *requestUnmarshalIntegerOverflowError but got %T (%s)", err, err)
	}
}

func assertUnmarshalIntegerUnderflowError(t *testing.T, err error, v reflect.Value) {
	if err == nil {
		t.Errorf("Expected *requestUnmarshalIntegerUnderflowError but got value %#v", v)
	} else if _, ok := err.(*requestUnmarshalIntegerUnderflowError); !ok {
		t.Errorf("Expected *requestUnmarshalIntegerUnderflowError but got %T (%s)", err, err)
	}
}
