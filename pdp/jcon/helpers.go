package jcon

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

func checkRootObjectStart(d *json.Decoder) (bool, error) {
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

func checkRootArrayStart(d *json.Decoder) (bool, error) {
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

func checkObjectStart(d *json.Decoder, desc string) error {
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

func checkArrayStart(d *json.Decoder, desc string) error {
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

func checkObjectArrayStart(d *json.Decoder, desc string) (bool, error) {
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

func checkEOF(d *json.Decoder) error {
	t, err := d.Token()
	if err == io.EOF {
		return nil
	}

	if err != nil {
		return err
	}

	return newMissingEOFError(t)
}

func skipValue(d *json.Decoder, desc string) error {
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
			return skipObject(d, desc)

		case delimArrayStart:
			return skipArray(d, desc)
		}
	}

	return nil
}

func skipObject(d *json.Decoder, desc string) error {
	for {
		t, err := d.Token()
		if err != nil {
			return err
		}

		switch t := t.(type) {
		default:
			return newObjectTokenError(t, delimObjectEnd, desc)

		case string:
			err := skipValue(d, desc)
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

func skipArray(d *json.Decoder, desc string) error {
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
				err := skipObject(d, desc)
				if err != nil {
					return bindError(err, src)
				}

			case delimArrayStart:
				err := skipArray(d, desc)
				if err != nil {
					return bindError(err, src)
				}
			}
		}

		i++
	}
}

type pair struct {
	k string
	v interface{}
}

func getUndefined(d *json.Decoder, desc string) (interface{}, error) {
	t, err := d.Token()
	if err != nil {
		return nil, err
	}

	switch t := t.(type) {
	case json.Delim:
		s := t.String()
		switch s {
		case delimObjectStart:
			return getObject(d, desc)

		case delimArrayStart:
			return getArray(d, desc)
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

func getObject(d *json.Decoder, desc string) ([]pair, error) {
	obj := []pair{}

	for {
		t, err := d.Token()
		if err != nil {
			return nil, err
		}

		switch t := t.(type) {
		default:
			return nil, newObjectTokenError(t, delimObjectEnd, desc)

		case string:
			v, err := getUndefined(d, desc)
			if err != nil {
				return nil, bindError(err, t)
			}

			obj = append(obj, pair{k: t, v: v})

		case json.Delim:
			if t.String() != delimObjectEnd {
				return nil, newObjectEndDelimiterError(t, delimObjectEnd, desc)
			}

			return obj, nil
		}
	}
}

func getArray(d *json.Decoder, desc string) ([]interface{}, error) {
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
				v, err := getObject(d, desc)
				if err != nil {
					return nil, bindError(err, src)
				}

				arr = append(arr, v)

			case delimArrayStart:
				v, err := getArray(d, desc)
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

func getBoolean(d *json.Decoder, desc string) (bool, error) {
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

func getString(d *json.Decoder, desc string) (string, error) {
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

func getStringSequence(d *json.Decoder, desc string, f func(s string) error) error {
	obj, err := checkObjectArrayStart(d, desc)
	if err != nil {
		return err
	}

	if obj {
		return getStringSequenceFromObject(d, desc, f)
	}

	return getStringSequenceFromArray(d, desc, f)
}

func getStringSequenceFromObject(d *json.Decoder, desc string, f func(s string) error) error {
	i := 1
	for {
		t, err := d.Token()
		if err != nil {
			return bindError(err, fmt.Sprintf("%d", i))
		}

		switch t := t.(type) {
		default:
			return bindError(newObjectTokenError(t, delimObjectEnd, desc), fmt.Sprintf("%d", i))

		case string:
			err := f(t)
			if err != nil {
				return bindError(err, fmt.Sprintf("%d", i))
			}

			err = skipValue(d, desc)
			if err != nil {
				return bindError(err, fmt.Sprintf("%d", i))
			}

		case json.Delim:
			if t.String() != delimObjectEnd {
				return bindError(newObjectEndDelimiterError(t, delimObjectEnd, desc), fmt.Sprintf("%d", i))
			}

			return nil
		}

		i++
	}
}

func getStringSequenceFromArray(d *json.Decoder, desc string, f func(s string) error) error {
	i := 1
	for {
		t, err := d.Token()
		if err != nil {
			return bindError(err, fmt.Sprintf("%d", i))
		}

		switch t := t.(type) {
		default:
			return bindError(newStringArrayTokenError(t, delimArrayEnd, desc), fmt.Sprintf("%d", i))

		case string:
			err := f(t)
			if err != nil {
				return bindError(err, fmt.Sprintf("%d", i))
			}

		case json.Delim:
			if t.String() != delimArrayEnd {
				return bindError(newArrayEndDelimiterError(t, delimArrayEnd, desc), fmt.Sprintf("%d", i))
			}

			return nil
		}

		i++
	}
}

func unmarshalObject(d *json.Decoder, u func(string, *json.Decoder) error, desc string) error {
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

func unmarshalObjectArray(d *json.Decoder, u func(*json.Decoder) error, desc string) error {
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
			err := u(d)
			if err != nil {
				return bindError(err, src)
			}
		}

		i++
	}
}
