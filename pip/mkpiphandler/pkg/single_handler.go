package pkg

import (
	"io"
	"strings"
	"text/template"
)

func (s *Schema) genSingleHandler(dir string, p *Endpoint) error {
	return toFile(dir, handlerDst, s.makeSingleHandler(p).execute)
}

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

func (s *Schema) makeSingleHandler(p *Endpoint) singleHandler {
	return singleHandler{
		Package:    s.Package,
		Imports:    strings.Join(makeImports(p.goArgPkgs|p.goResultPkg, singleHandlerImports...), "\n\t"),
		ArgCount:   len(p.goArgs),
		Types:      strings.Join(p.goArgs, ", "),
		Args:       p.goArgList,
		ArgParsers: p.goParsers,
		ResultType: p.goResult,
		ResultZero: p.goResultZero,
		Marshaller: p.goMarshaller,
	}
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
	reqIDSize         = 4
	reqVersionSize    = 2
	reqVersion        = uint16(1)
	reqArgs           = uint16({{.ArgCount}})
	reqBigCounterSize = 2
)

var (
	errFragment          = errors.New("fragment")
	errInvalidReqVersion = errors.New("invalid request version")
	errInvalidArgCount   = errors.New("invalid count of request arguments")
)

// WrapHandler converts custom Handler to generic PIP ServiceHandler.
func WrapHandler(f Handler) server.ServiceHandler {
	return func(b []byte) []byte {
		if len(b) < reqIDSize {
			panic("missing request id")
		}
		in := b[reqIDSize:]

		r, err := handler(in, f)
		if err != nil {
			n, err := pdp.MarshalInfoError(in[:cap(in)], err)
			if err != nil {
				panic(err)
			}

			return b[:reqIDSize+n]
		}

		n, err := {{.Marshaller}}(in[:cap(in)], r)
		if err != nil {
			panic(err)
		}

		return b[:reqIDSize+n]
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

	skip := binary.LittleEndian.Uint16(in)
	in = in[reqBigCounterSize:]

	if len(in) < int(skip)+reqBigCounterSize {
		return {{.ResultZero}}, errFragment
	}
	in = in[skip:]

	if c := binary.LittleEndian.Uint16(in); c != reqArgs {
		return {{.ResultZero}}, errInvalidArgCount
	}
	in = in[reqBigCounterSize:]

{{.ArgParsers}}	return f({{.Args}})
}
`))
)
