// Package pipexample is a generated PIP server handler package. DO NOT EDIT.
package pipexample

import (
	"errors"
	"github.com/infobloxopen/themis/pdp"
)

const reqDefaultArgs = 2

var errInvalidDefaultArgCount = errors.New("invalid count of request arguments for * endpoint")

func handleDefault(c int, in, b []byte, e Endpoints) (int, error) {
	if c != reqDefaultArgs {
		return 0, errInvalidDefaultArgCount
	}

	v0, in, err := pdp.GetInfoRequestStringValue(in)
	if err != nil {
		return 0, err
	}

	v1, _, err := pdp.GetInfoRequestAddressValue(in)
	if err != nil {
		return 0, err
	}

	v, err := e.Default(v0, v1)
	if err != nil {
		return 0, err
	}

	n, err := pdp.MarshalInfoResponseNetwork(b[:cap(b)], v)
	if err != nil {
		panic(err)
	}

	return n, nil
}
