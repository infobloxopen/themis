package pdp

import (
	"fmt"
	"math"
	"reflect"
)

func setInt(v reflect.Value, i int64) error {
	k := v.Kind()
	switch k {
	case reflect.Int:
		if i < math.MinInt32 || i > math.MaxInt32 {
			return newRequestUnmarshalIntegerOverflowError(i, k)
		}

		v.SetInt(i)
		return nil

	case reflect.Int8:
		if i < math.MinInt8 || i > math.MaxInt8 {
			return newRequestUnmarshalIntegerOverflowError(i, k)
		}

		v.SetInt(i)
		return nil

	case reflect.Int16:
		if i < math.MinInt16 || i > math.MaxInt16 {
			return newRequestUnmarshalIntegerOverflowError(i, k)
		}

		v.SetInt(i)
		return nil

	case reflect.Int32:
		if i < math.MinInt32 || i > math.MaxInt32 {
			return newRequestUnmarshalIntegerOverflowError(i, k)
		}

		v.SetInt(i)
		return nil

	case reflect.Int64:
		v.SetInt(i)
		return nil

	case reflect.Uint:
		if i < 0 {
			return newRequestUnmarshalIntegerUnderflowError(i)
		}

		if i > math.MaxUint32 {
			return newRequestUnmarshalIntegerOverflowError(i, k)
		}

		v.SetUint(uint64(i))
		return nil

	case reflect.Uint8:
		if i < 0 {
			return newRequestUnmarshalIntegerUnderflowError(i)
		}

		if i > math.MaxUint8 {
			return newRequestUnmarshalIntegerOverflowError(i, k)
		}

		v.SetUint(uint64(i))
		return nil

	case reflect.Uint16:
		if i < 0 {
			return newRequestUnmarshalIntegerUnderflowError(i)
		}

		if i > math.MaxUint16 {
			return newRequestUnmarshalIntegerOverflowError(i, k)
		}

		v.SetUint(uint64(i))
		return nil

	case reflect.Uint32:
		if i < 0 {
			return newRequestUnmarshalIntegerUnderflowError(i)
		}

		if i > math.MaxUint32 {
			return newRequestUnmarshalIntegerOverflowError(i, k)
		}

		v.SetUint(uint64(i))
		return nil

	case reflect.Uint64:
		if i < 0 {
			return newRequestUnmarshalIntegerUnderflowError(i)
		}

		v.SetUint(uint64(i))
		return nil
	}

	panic(fmt.Errorf("setting int value to %s", k))
}
