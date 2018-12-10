package pkg

import (
	"fmt"
	"io"
	"sort"
	"strings"
	"text/template"
)

const dispatcherDst = "dispatcher.go"

func (s *Schema) genDispatcher(dir string) error {
	return toFile(dir, dispatcherDst, s.makeDispatcher().execute)
}

type dispatcher struct {
	Package  string
	Handlers string
}

func (s *Schema) makeDispatcher() dispatcher {
	keys := make([]string, 0, len(s.Endpoints))
	for k := range s.Endpoints {
		if !isDefaultEndpoint(k) {
			keys = append(keys, k)
		}
	}
	sort.Strings(keys)

	handlers := make([]string, 0, len(keys))
	for _, k := range keys {
		handlers = append(handlers,
			fmt.Sprintf("\tcase %q:\n\t\tn, err = handle%s(c, in, b, e)", k, s.Endpoints[k].goName),
		)
	}

	var ss string
	if len(handlers) > 0 {
		ss += "\n" + strings.Join(handlers, "\n\n") + "\n"
	}

	return dispatcher{
		Package:  s.Package,
		Handlers: ss,
	}
}

func (t dispatcher) execute(w io.Writer) error {
	return dispatcherTemplate.Execute(w, t)
}

var dispatcherTemplate = template.Must(template.New("dispatcher").Parse(
	`// Package {{.Package}} is a generated PIP server handler package. DO NOT EDIT.
package {{.Package}}

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
{{.Handlers}}	}

	return n, err
}
`))
