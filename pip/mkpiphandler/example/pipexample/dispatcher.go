// Package pipexample is a generated PIP server handler package. DO NOT EDIT.
package pipexample

import (
	"encoding/binary"
	"errors"
)

const (
	reqVersionSize    = 2
	reqVersion        = uint16(1)
	reqBigCounterSize = 2
)

var (
	errFragment          = errors.New("fragment")
	errInvalidReqVersion = errors.New("invalid request version")
)

func dispatch(b []byte, e Endpoints) (int, error) {
	in := b
	if len(in) < reqVersionSize+reqBigCounterSize {
		return 0, errFragment
	}

	if v := binary.LittleEndian.Uint16(in); v != reqVersion {
		return 0, errInvalidReqVersion
	}
	in = in[reqVersionSize:]

	size := int(binary.LittleEndian.Uint16(in))
	in = in[reqBigCounterSize:]
	if len(in) < size+reqBigCounterSize {
		return 0, errFragment
	}

	path := in[:size]
	in = in[size:]

	c := int(binary.LittleEndian.Uint16(in))
	in = in[reqBigCounterSize:]

	var (
		n   int
		err error
	)

	switch string(path) {
	default:
		n, err = handleDefault(c, in, b, e)

	case "list":
		n, err = handleList(c, in, b, e)

	case "set":
		n, err = handleSet(c, in, b, e)
	}

	return n, err
}
