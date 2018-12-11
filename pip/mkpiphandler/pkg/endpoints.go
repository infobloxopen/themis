package pkg

import (
	"errors"
	"fmt"
	"io"
	"strings"
	"text/template"
	"unicode"
	"unicode/utf8"
)

const (
	endpointsDst = "endpoints.go"

	defaultEndpointAlias = "*"
	defaultEndpointName  = "default"
)

var (
	errNoEndpoints       = errors.New("no endpoints provided")
	errEmptyEndpointName = errors.New("name is empty")
)

func (s *Schema) getFirstEndpoint() (string, *Endpoint, error) {
	for k, v := range s.Endpoints {
		return k, v, nil
	}

	return "", nil, errNoEndpoints
}

func (p *Endpoint) postProcess(k string, single bool) error {
	if k == defaultEndpointAlias {
		k = defaultEndpointName
	}

	name, err := makeGoName(k)
	if err != nil {
		return err
	}
	p.goName = name

	goArgs, err := makeGoTypeList(p.Args)
	if err != nil {
		return err
	}

	p.goArgs = goArgs
	p.goArgPkgs = collectImports(goArgs...)

	argNames := make([]string, 0, len(p.Args))
	for i := range p.Args {
		argNames = append(argNames, fmt.Sprintf("v%d", i))
	}
	p.goArgList = strings.Join(argNames, ", ")

	goResult, goResultZero, err := getGoType(p.Result)
	if err != nil {
		return err
	}

	p.goResult = goResult
	p.goResultZero = goResultZero
	p.goResultPkg = collectImports(goResult)

	if !single {
		p.goResultZero = "0"
	}
	p.goParsers = strings.Join(makeArgParsers(p.Args, p.goResultZero), "\n")
	if len(p.goParsers) > 0 {
		p.goParsers += "\n"
	}

	p.goMarshaller = marshallerMap[strings.ToLower(p.Result)]

	return nil
}

func isDefaultEndpoint(s string) bool {
	return s == defaultEndpointAlias || strings.ToLower(s) == defaultEndpointName
}

func (s *Schema) genEndpointsInterface(dir string) error {
	return toFile(dir, endpointsDst, s.makeEndpointsInterface().execute)
}

type endpointsInterface struct {
	Package string
	Imports string
	Methods string
}

func (s *Schema) makeEndpointsInterface() endpointsInterface {
	goPkgs := 0
	methods := make([]string, 0, len(s.Endpoints))
	for _, p := range s.Endpoints {
		goPkgs |= p.goArgPkgs | p.goResultPkg
		methods = append(methods,
			fmt.Sprintf("%s(%s) (%s, error)", p.goName, strings.Join(p.goArgs, ", "), p.goResult),
		)
	}

	return endpointsInterface{
		Package: s.Package,
		Imports: strings.Join(makeImports(goPkgs), "\n\t"),
		Methods: strings.Join(methods, "\n\t"),
	}
}

func (t endpointsInterface) execute(w io.Writer) error {
	return endpointsInterfaceTemplate.Execute(w, t)
}

var endpointsInterfaceTemplate = template.Must(template.New("endpoints").Parse(
	`// Package {{.Package}} is a generated PIP server handler package. DO NOT EDIT.
package {{.Package}}

import (
	{{.Imports}}
)

// Endpoints is the interface that wraps PIP handlers.
type Endpoints interface {
	{{.Methods}}
}
`))

func makeGoName(s string) (string, error) {
	r, n, err := getFirstRuneForExport(s)
	if err != nil {
		return "", err
	}

	b := make([]byte, 0, len(s))
	b = appendRune(b, r)

	toUpper := false
	for _, r := range s[n:] {
		if unicode.Is(unicode.Letter, r) || unicode.Is(unicode.Digit, r) {
			if toUpper {
				toUpper = false
				r = unicode.ToUpper(r)
			}

			b = appendRune(b, r)
		} else {
			toUpper = true
		}
	}

	return string(b), nil
}

func getFirstRuneForExport(s string) (rune, int, error) {
	for _, r := range s {
		u := unicode.ToUpper(r)
		if !unicode.Is(unicode.Letter, u) || !unicode.Is(unicode.Upper, u) {
			return utf8.RuneError, 0, fmt.Errorf("failed to make exportable name from %q", s)
		}

		return u, utf8.RuneLen(r), nil
	}

	return utf8.RuneError, 0, errEmptyEndpointName
}

func appendRune(b []byte, r rune) []byte {
	rLen := utf8.RuneLen(r)
	if rLen > 0 {
		if cap(b)-len(b) < rLen {
			size := cap(b) + cap(b)/2
			if size-len(b) < rLen {
				size = len(b) + rLen
			}

			tmp := make([]byte, len(b), size)
			copy(tmp, b)
			b = tmp
		}

		start := len(b)
		b = b[:len(b)+rLen]
		utf8.EncodeRune(b[start:], r)
	}

	return b
}
