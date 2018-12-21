package pkg

import (
	"fmt"
	"io"
	"strings"
	"text/template"
)

type singleHandler struct {
	Package string
	Imports string

	ArgCount   int
	Types      string
	Args       string
	ArgParsers string
	ResultType string
	ResultZero string
	Marshaller string
}

func (s *Schema) makeSingleHandler(p *Endpoint) (singleHandler, error) {
	args := make([]string, 0, len(p.Args))
	for i := range p.Args {
		args = append(args, fmt.Sprintf("v%d", i))
	}

	parsers, err := makeArgParsers(p.Args, p.goResultZero)
	if err != nil {
		return singleHandler{}, err
	}

	argParsers := strings.Join(parsers, "\n")
	if len(argParsers) > 0 {
		argParsers += "\n"
	}

	marshaller, ok := marshallerMap[strings.ToLower(p.Result)]
	if !ok {
		return singleHandler{}, fmt.Errorf("result: unknown type %q", p.Result)
	}

	return singleHandler{
		Package:    s.Package,
		Imports:    strings.Join(makeImports(p.goArgPkgs|p.goResultPkg, singleHandlerImports...), "\n\t"),
		ArgCount:   len(p.goArgs),
		Types:      strings.Join(p.goArgs, ", "),
		Args:       strings.Join(args, ", "),
		ArgParsers: argParsers,
		ResultType: p.goResult,
		ResultZero: p.goResultZero,
		Marshaller: marshaller,
	}, nil
}

func (t singleHandler) execute(w io.Writer) error {
	return singleHandlerTemplate.Execute(w, t)
}

var (
	singleHandlerImports = []string{
		"\"encoding/binary\"",
		"\"errors\"",
		"\"github.com/infobloxopen/themis/pdp\"",
		"\"github.com/infobloxopen/themis/pip/server\"",
	}

	singleHandlerTemplate = template.Must(template.New("handler").Parse(
		`// Package {{.Package}} is a generated PIP server handler package. DO NOT EDIT.
package {{.Package}}

import (
	{{.Imports}}
)

// Handler is a customized PIP handler for given input and output.
type Handler func({{.Types}}) ({{.ResultType}}, error)

const (
	reqIdSize         = 4
	reqVersionSize    = 2
	reqVersion        = uint16(1)
	reqArgs           = uint16({{.ArgCount}})
	reqBigCounterSize = 2
	reqTypeSize       = 1
)

var (
	errFragment          = errors.New("fragment")
	errInvalidReqVersion = errors.New("invalid request version")
	errInvalidArgCount   = errors.New("invalid count of request arguments")
)

// WrapHandler converts custom Handler to generic PIP ServiceHandler.
func WrapHandler(f Handler) server.ServiceHandler {
	return func(b []byte) []byte {
		if len(b) < reqIdSize {
			panic("missing request id")
		}
		in := b[reqIdSize:]

		r, err := handler(in, f)
		if err != nil {
			n, err := pdp.MarshalInfoError(in[:cap(in)], err)
			if err != nil {
				panic(err)
			}

			return b[:reqIdSize+n]
		}

		n, err := {{.Marshaller}}(in[:cap(in)], r)
		if err != nil {
			panic(err)
		}

		return b[:reqIdSize+n]
	}
}

func handler(in []byte, f Handler) ({{.ResultType}}, error) {
	if len(in) < reqVersionSize+reqBigCounterSize {
		return {{.ResultZero}}, errFragment
	}

	if v := binary.LittleEndian.Uint16(in); v != reqVersion {
		return {{.ResultZero}}, errInvalidReqVersion
	}
	in = in[reqVersionSize:]

	if c := binary.LittleEndian.Uint16(in); c != reqArgs {
		return {{.ResultZero}}, errInvalidArgCount
	}
	in = in[reqBigCounterSize:]

{{.ArgParsers}}	return f({{.Args}})
}
`))
)
