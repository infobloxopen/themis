package pkg

import (
	"io"
	"strings"
	"text/template"
)

const (
	handlersDst       = "_handler.go"
	defaultHandlerDst = "default_handler.go"
)

func (s *Schema) genHandlers(dir string) error {
	hasDefault := false
	for k, p := range s.Endpoints {
		name := defaultHandlerDst
		if isDefaultEndpoint(k) {
			hasDefault = true
		} else {
			name = k + handlersDst
		}

		if err := toFile(dir, name, s.makeEndpointHandler(k, p).execute); err != nil {
			return err
		}
	}

	if !hasDefault {
		return toFile(dir, defaultHandlerDst, s.makeDefaultHandler().execute)
	}

	return nil
}

type endpointHandler struct {
	Package string
	Imports string

	Name       string
	GoName     string
	ArgCount   int
	Args       string
	ArgParsers string
	Marshaller string
}

func (s *Schema) makeEndpointHandler(k string, p *Endpoint) endpointHandler {
	return endpointHandler{
		Package:    s.Package,
		Imports:    strings.Join(endpointHandlerImports, "\n\t"),
		Name:       k,
		GoName:     p.goName,
		ArgCount:   len(p.goArgs),
		Args:       p.goArgList,
		ArgParsers: p.goParsers,
		Marshaller: p.goMarshaller,
	}
}

func (t endpointHandler) execute(w io.Writer) error {
	return endpointHandlerTemplate.Execute(w, t)
}

var (
	endpointHandlerImports = []string{
		"\"errors\"",
		"\"github.com/infobloxopen/themis/pdp\"",
	}

	endpointHandlerTemplate = template.Must(template.New("handler").Parse(
		`// Package {{.Package}} is a generated PIP server handler package. DO NOT EDIT.
package {{.Package}}

import (
	{{.Imports}}
)

const req{{.GoName}}Args = {{.ArgCount}}

var errInvalid{{.GoName}}ArgCount = errors.New("invalid count of request arguments for {{.Name}} endpoint")

func handle{{.GoName}}(c int, in, b []byte, e Endpoints) (int, error) {
	if c != req{{.GoName}}Args {
		return 0, errInvalid{{.GoName}}ArgCount
	}

{{.ArgParsers}}	v, err := e.{{.GoName}}({{.Args}})
	if err != nil {
		return 0, err
	}

	n, err := {{.Marshaller}}(b[:cap(b)], v)
	if err != nil {
		panic(err)
	}

	return n, nil
}
`))
)

type defaultHandler struct {
	Package string
}

func (s *Schema) makeDefaultHandler() defaultHandler {
	return defaultHandler{
		Package: s.Package,
	}
}

func (t defaultHandler) execute(w io.Writer) error {
	return defaultHandlerTemplate.Execute(w, t)
}

var defaultHandlerTemplate = template.Must(template.New("handler").Parse(
	`// Package {{.Package}} is a generated PIP server handler package. DO NOT EDIT.
package {{.Package}}

import "errors"

var errUnknownEndpoint = errors.New("unknown endpoint")

func handleDefault(c int, in, b []byte, e Endpoints) (int, error) {
	return 0, errUnknownEndpoint
}
`))
