// Package jparser provides helper methods to parse JSON from stream.
package jparser

import (
	"encoding/json"
	"fmt"
	"io"
)

const (
	delimObjectStart = "{"
	delimObjectEnd   = "}"

	delimArrayStart = "["
	delimArrayEnd   = "]"
)

func CheckRootObjectStart(d *json.Decoder) (bool, error) {
	t, err := d.Token()
	if err == io.EOF {
		return false, nil
	}

	if err != nil {
		return false, err
	}

	delim, ok := t.(json.Delim)
	if !ok {
		return false, newRootObjectStartTokenError(t, delimObjectStart)
	}

	if delim.String() != delimObjectStart {
		return false, newRootObjectStartDelimiterError(delim, delimObjectStart)
	}

	return true, nil
}

func CheckRootArrayStart(d *json.Decoder) (bool, error) {
	t, err := d.Token()
	if err == io.EOF {
		return false, nil
	}

	if err != nil {
		return false, err
	}

	delim, ok := t.(json.Delim)
	if !ok {
		return false, newRootArrayStartTokenError(t, delimArrayStart)
	}

	if delim.String() != delimArrayStart {
		return false, newRootArrayStartDelimiterError(delim, delimArrayStart)
	}

	return true, nil
}

func CheckObjectStart(d *json.Decoder, desc string) error {
	t, err := d.Token()
	if err != nil {
		return err
	}

	delim, ok := t.(json.Delim)
	if !ok {
		return newObjectStartTokenError(t, delimObjectStart, desc)
	}

	if delim.String() != delimObjectStart {
		return newObjectStartDelimiterError(delim, delimObjectStart, desc)
	}

	return nil
}

func CheckArrayStart(d *json.Decoder, desc string) error {
	t, err := d.Token()
	if err != nil {
		return err
	}

	delim, ok := t.(json.Delim)
	if !ok {
		return newArrayStartTokenError(t, delimArrayStart, desc)
	}

	if delim.String() != delimArrayStart {
		return newArrayStartDelimiterError(delim, delimArrayStart, desc)
	}

	return nil
}

func CheckObjectArrayStart(d *json.Decoder, desc string) (bool, error) {
	t, err := d.Token()
	if err != nil {
		return false, err
	}

	delim, ok := t.(json.Delim)
	if !ok {
		return false, newObjectArrayStartTokenError(t, delimObjectStart, delimArrayStart, desc)
	}

	switch delim.String() {
	case delimObjectStart:
		return true, nil

	case delimArrayStart:
		return false, nil
	}

	return false, newObjectArrayStartDelimiterError(delim, delimObjectStart, delimArrayStart, desc)
}

func CheckEOF(d *json.Decoder) error {
	t, err := d.Token()
	if err == io.EOF {
		return nil
	}

	if err != nil {
		return err
	}

	return newMissingEOFError(t)
}

func SkipValue(d *json.Decoder, desc string) error {
	t, err := d.Token()
	if err != nil {
		return err
	}

	if delim, ok := t.(json.Delim); ok {
		s := delim.String()
		switch s {
		default:
			return newUnexpectedDelimiterError(s, desc)

		case delimObjectStart:
			return SkipObject(d, desc)

		case delimArrayStart:
			return SkipArray(d, desc)
		}
	}

	return nil
}

func SkipObject(d *json.Decoder, desc string) error {
	for {
		t, err := d.Token()
		if err != nil {
			return err
		}

		switch t := t.(type) {
		default:
			return newObjectTokenError(t, delimObjectEnd, desc)

		case string:
			err := SkipValue(d, desc)
			if err != nil {
				return bindError(err, t)
			}

		case json.Delim:
			if t.String() != delimObjectEnd {
				return newObjectEndDelimiterError(t, delimObjectEnd, desc)
			}

			return nil
		}
	}
}

func SkipArray(d *json.Decoder, desc string) error {
	i := 1
	for {
		src := fmt.Sprintf("%d", i)

		t, err := d.Token()
		if err != nil {
			return bindError(err, src)
		}

		if delim, ok := t.(json.Delim); ok {
			s := delim.String()
			switch s {
			default:
				return bindError(newUnexpectedDelimiterError(s, desc), src)

			case delimArrayEnd:
				return nil

			case delimObjectStart:
				err := SkipObject(d, desc)
				if err != nil {
					return bindError(err, src)
				}

			case delimArrayStart:
				err := SkipArray(d, desc)
				if err != nil {
					return bindError(err, src)
				}
			}
		}

		i++
	}
}

type Pair struct {
	K string
	V interface{}
}

func GetUndefined(d *json.Decoder, desc string) (interface{}, error) {
	t, err := d.Token()
	if err != nil {
		return nil, err
	}

	switch t := t.(type) {
	case json.Delim:
		s := t.String()
		switch s {
		case delimObjectStart:
			return GetObject(d, desc)

		case delimArrayStart:
			return GetArray(d, desc)
		}

		return nil, newUnexpectedDelimiterError(s, desc)

	case bool:
		return t, nil

	case float64:
		return t, nil

	case json.Number:
		return t, nil

	case string:
		return t, nil
	}

	return t, nil
}

func GetObject(d *json.Decoder, desc string) ([]Pair, error) {
	obj := []Pair{}

	for {
		t, err := d.Token()
		if err != nil {
			return nil, err
		}

		switch t := t.(type) {
		default:
			return nil, newObjectTokenError(t, delimObjectEnd, desc)

		case string:
			v, err := GetUndefined(d, desc)
			if err != nil {
				return nil, bindError(err, t)
			}

			obj = append(obj, Pair{K: t, V: v})

		case json.Delim:
			if t.String() != delimObjectEnd {
				return nil, newObjectEndDelimiterError(t, delimObjectEnd, desc)
			}

			return obj, nil
		}
	}
}

func GetArray(d *json.Decoder, desc string) ([]interface{}, error) {
	arr := []interface{}{}
	i := 1
	for {
		src := fmt.Sprintf("%d", i)

		t, err := d.Token()
		if err != nil {
			return nil, bindError(err, src)
		}

		if delim, ok := t.(json.Delim); ok {
			s := delim.String()
			switch s {
			default:
				return nil, bindError(newUnexpectedDelimiterError(s, desc), src)

			case delimArrayEnd:
				return arr, nil

			case delimObjectStart:
				v, err := GetObject(d, desc)
				if err != nil {
					return nil, bindError(err, src)
				}

				arr = append(arr, v)

			case delimArrayStart:
				v, err := GetArray(d, desc)
				if err != nil {
					return nil, bindError(err, src)
				}

				arr = append(arr, v)
			}
		} else {
			arr = append(arr, t)
		}

		i++
	}
}

func GetBoolean(d *json.Decoder, desc string) (bool, error) {
	t, err := d.Token()
	if err != nil {
		return false, err
	}

	b, ok := t.(bool)
	if !ok {
		return false, newBooleanCastError(t, desc)
	}

	return b, nil
}

func GetString(d *json.Decoder, desc string) (string, error) {
	t, err := d.Token()
	if err != nil {
		return "", err
	}

	s, ok := t.(string)
	if !ok {
		return "", newStringCastError(t, desc)
	}

	return s, nil
}

func GetStringSequence(d *json.Decoder, desc string, f func(idx int, s string) error) error {
	ok, err := CheckObjectArrayStart(d, desc)
	if err != nil {
		return err
	}

	if ok {
		return GetStringSequenceFromObject(d, desc, f)
	}

	return GetStringSequenceFromArray(d, desc, f)
}

func GetStringSequenceFromObject(d *json.Decoder, desc string, f func(idx int, s string) error) error {
	i := 1
	for {
		t, err := d.Token()
		if err != nil {
			return bindErrorf(err, "%d", i)
		}

		switch t := t.(type) {
		default:
			return bindErrorf(newObjectTokenError(t, delimObjectEnd, desc), "%d", i)

		case string:
			err := f(i, t)
			if err != nil {
				return err
			}

			err = SkipValue(d, desc)
			if err != nil {
				return bindErrorf(err, "%d", i)
			}

		case json.Delim:
			if t.String() != delimObjectEnd {
				return bindErrorf(newObjectEndDelimiterError(t, delimObjectEnd, desc), "%d", i)
			}

			return nil
		}

		i++
	}
}

func GetStringSequenceFromArray(d *json.Decoder, desc string, f func(idx int, s string) error) error {
	i := 1
	for {
		t, err := d.Token()
		if err != nil {
			return bindErrorf(err, "%d", i)
		}

		switch t := t.(type) {
		default:
			return bindErrorf(newStringArrayTokenError(t, delimArrayEnd, desc), "%d", i)

		case string:
			err := f(i, t)
			if err != nil {
				return err
			}

		case json.Delim:
			if t.String() != delimArrayEnd {
				return bindErrorf(newArrayEndDelimiterError(t, delimArrayEnd, desc), "%d", i)
			}

			return nil
		}

		i++
	}
}

func UnmarshalObject(d *json.Decoder, u func(key string, d *json.Decoder) error, desc string) error {
	for {
		t, err := d.Token()
		if err != nil {
			return err
		}

		switch t := t.(type) {
		default:
			return newObjectTokenError(t, delimObjectEnd, desc)

		case string:
			err = u(t, d)
			if err != nil {
				return err
			}

		case json.Delim:
			if t.String() != delimObjectEnd {
				return newObjectEndDelimiterError(t, delimObjectEnd, desc)
			}

			return nil
		}
	}
}

func UnmarshalObjectArray(d *json.Decoder, u func(idx int, d *json.Decoder) error, desc string) error {
	i := 1
	for {
		src := fmt.Sprintf("%d", i)

		t, err := d.Token()
		if err != nil {
			return bindError(err, src)
		}

		delim, ok := t.(json.Delim)
		if !ok {
			return bindError(newObjectArrayTokenError(t, delimArrayEnd, desc), src)
		}

		s := delim.String()
		switch s {
		default:
			return bindError(newUnexpectedObjectArrayDelimiterError(s, desc), src)

		case delimArrayEnd:
			return nil

		case delimObjectStart:
			err := u(i, d)
			if err != nil {
				return err
			}
		}

		i++
	}
}
