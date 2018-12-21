package pkg

import (
	"io"
	"text/template"
)

const handlerDst = "handler.go"

func (s *Schema) genHandler(dir string) error {
	return toFile(dir, handlerDst, s.makeHandler().execute)
}

type handler struct {
	Package string
}

func (s *Schema) makeHandler() handler {
	return handler{
		Package: s.Package,
	}
}

func (t handler) execute(w io.Writer) error {
	return handlerTemplate.Execute(w, t)
}

var handlerTemplate = template.Must(template.New("handler").Parse(
	`// Package {{.Package}} is a generated PIP server handler package. DO NOT EDIT.
package {{.Package}}

import (
	"github.com/infobloxopen/themis/pdp"
	"github.com/infobloxopen/themis/pip/server"
)

const reqIDSize = 4

// MakeHandler creates PIP service handler for given Endpoints.
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
`))
