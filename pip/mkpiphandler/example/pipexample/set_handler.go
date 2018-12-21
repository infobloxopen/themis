// Package pipexample is a generated PIP server handler package. DO NOT EDIT.
package pipexample

import (
	"errors"
	"github.com/infobloxopen/themis/pdp"
)

const reqSetArgs = 2

var errInvalidSetArgCount = errors.New("invalid count of request arguments for set endpoint")

func handleSet(c int, in, b []byte, e Endpoints) (int, error) {
	if c != reqSetArgs {
		return 0, errInvalidSetArgCount
	}

	v0, in, err := pdp.GetInfoRequestIntegerValue(in)
	if err != nil {
		return 0, err
	}

	v1, _, err := pdp.GetInfoRequestDomainValue(in)
	if err != nil {
		return 0, err
	}

	v, err := e.Set(v0, v1)
	if err != nil {
		return 0, err
	}

	n, err := pdp.MarshalInfoResponseSetOfStrings(b[:cap(b)], v)
	if err != nil {
		panic(err)
	}

	return n, nil
}
