// Package pipexample is a generated PIP server handler package. DO NOT EDIT.
package pipexample

import (
	"github.com/infobloxopen/themis/pdp"
	"github.com/infobloxopen/themis/pip/server"
)

const reqIDSize = 4

func MakeHandler(e Endpoints) server.ServiceHandler {
	return func(b []byte) []byte {
		if len(b) < reqIDSize {
			panic("missing request id")
		}
		in := b[reqIDSize:]

		n, err := dispatch(in, e)
		if err != nil {
			n, err = pdp.MarshalInfoError(in[:cap(in)], err)
			if err != nil {
				panic(err)
			}
		}

		return b[:reqIDSize+n]
	}
}
