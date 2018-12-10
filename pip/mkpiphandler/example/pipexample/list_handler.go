// Package pipexample is a generated PIP server handler package. DO NOT EDIT.
package pipexample

import (
	"errors"
	"github.com/infobloxopen/themis/pdp"
)

const reqListArgs = 2

var errInvalidListArgCount = errors.New("invalid count of request arguments for list endpoint")

func handleList(c int, in, b []byte, e Endpoints) (int, error) {
	if c != reqListArgs {
		return 0, errInvalidListArgCount
	}

	v0, in, err := pdp.GetInfoRequestIntegerValue(in)
	if err != nil {
		return 0, err
	}

	v1, _, err := pdp.GetInfoRequestDomainValue(in)
	if err != nil {
		return 0, err
	}

	v, err := e.List(v0, v1)
	if err != nil {
		return 0, err
	}

	n, err := pdp.MarshalInfoResponseListOfStrings(b[:cap(b)], v)
	if err != nil {
		panic(err)
	}

	return n, nil
}
