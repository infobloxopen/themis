package pkg

import (
	"bytes"
	"io/ioutil"
	"os"
	"path"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSchemaGenHandlers(t *testing.T) {
	tmp, err := ioutil.TempDir("", "")
	if err != nil {
		assert.FailNow(t, "ioutil.TempDir(\"\", \"\"): %q", err)
	}

	defer func() {
		assert.NoError(t, os.RemoveAll(tmp))
	}()

	p := &Endpoint{
		Args:   []string{"String"},
		Result: "String",
	}
	d := &Endpoint{
		Args:   []string{"String"},
		Result: "String",
	}

	s := &Schema{
		Package: "test",
		Endpoints: map[string]*Endpoint{
			"test":               p,
			defaultEndpointAlias: d,
		},
	}
	err = s.postProcess()
	if err != nil {
		assert.FailNow(t, "s.(*Schema).postProcess(): %q", err)
	}

	err = s.genHandlers(tmp)
	assert.NoError(t, err)
	assert.FileExists(t, path.Join(tmp, "test"+handlersDst))
	assert.FileExists(t, path.Join(tmp, defaultHandlerDst))
}

func TestSchemaGenHandlersWithAutoDefault(t *testing.T) {
	tmp, err := ioutil.TempDir("", "")
	if err != nil {
		assert.FailNow(t, "ioutil.TempDir(\"\", \"\"): %q", err)
	}

	defer func() {
		assert.NoError(t, os.RemoveAll(tmp))
	}()

	p := &Endpoint{
		Args:   []string{"String"},
		Result: "String",
	}

	s := &Schema{
		Package: "test",
		Endpoints: map[string]*Endpoint{
			"test": p,
		},
	}
	err = s.postProcess()
	if err != nil {
		assert.FailNow(t, "s.(*Schema).postProcess(): %q", err)
	}

	err = s.genHandlers(tmp)
	assert.NoError(t, err)
	assert.FileExists(t, path.Join(tmp, "test"+handlersDst))
	assert.FileExists(t, path.Join(tmp, defaultHandlerDst))
}

func TestSchemaGenHandlersWithError(t *testing.T) {
	tmp, err := ioutil.TempDir("", "")
	if err != nil {
		assert.FailNow(t, "ioutil.TempDir(\"\", \"\"): %q", err)
	}

	defer func() {
		if assert.NoError(t, os.Chmod(tmp, 0750)) {
			assert.NoError(t, os.RemoveAll(tmp))
		}
	}()

	if err = os.Chmod(tmp, 0550); !assert.NoError(t, err) {
		assert.Failf(t, "os.Chmod(%q): %q", tmp, err)
	}

	p := &Endpoint{
		Args:   []string{"String"},
		Result: "String",
	}

	s := &Schema{
		Package: "test",
		Endpoints: map[string]*Endpoint{
			"test": p,
		},
	}
	err = s.postProcess()
	if err != nil {
		assert.FailNow(t, "s.(*Schema).postProcess(): %q", err)
	}

	err = s.genHandlers(tmp)
	assert.Error(t, err)

	_, err = os.Stat(path.Join(tmp, "test"+handlersDst))
	assert.Error(t, err)

	_, err = os.Stat(path.Join(tmp, defaultHandlerDst))
	assert.Error(t, err)
}

func TestSchemaMakeEndpointHandler(t *testing.T) {
	p := &Endpoint{
		goName: "Test",
		Args: []string{
			pipTypeString,
		},
		goArgs: []string{
			goTypeString,
		},
		goArgList: "v0",

		goParsers:    testEndpointHandlerSnippet,
		goMarshaller: pdpMarshallerString,

		Result:       pipTypeString,
		goResult:     goTypeString,
		goResultZero: "\"\"",
	}
	s := &Schema{
		Package: "test",
		Endpoints: map[string]*Endpoint{
			"test": p,
		},
	}

	h := s.makeEndpointHandler("test", p)
	assert.Equal(t, "test", h.Package)
	assert.NotEmpty(t, h.Imports)

	assert.Equal(t, "test", h.Name)
	assert.Equal(t, "Test", h.GoName)

	assert.Equal(t, 1, h.ArgCount)
	assert.NotEmpty(t, h.Args)
	assert.NotEmpty(t, h.Marshaller)
}

func TestEndpointHandlerExecute(t *testing.T) {
	h := endpointHandler{
		Package:    "test",
		Imports:    strings.Join(endpointHandlerImports, "\n\t"),
		Name:       "test",
		GoName:     "Test",
		ArgCount:   1,
		Args:       "v0",
		ArgParsers: testEndpointHandlerSnippet,
		Marshaller: pdpMarshallerString,
	}

	b := new(bytes.Buffer)
	err := h.execute(b)
	assert.NoError(t, err)
	assert.Equal(t, testEndpointHandlerSource, b.String())
}

func TestSchemaMakeDefaultHandler(t *testing.T) {
	p := &Endpoint{
		goName: "Test",
		Args: []string{
			pipTypeString,
		},
		goArgs: []string{
			goTypeString,
		},
		goArgList: "v0",

		goParsers:    testEndpointHandlerSnippet,
		goMarshaller: pdpMarshallerString,

		Result:       pipTypeString,
		goResult:     goTypeString,
		goResultZero: "\"\"",
	}
	s := &Schema{
		Package: "test",
		Endpoints: map[string]*Endpoint{
			"test": p,
		},
	}

	h := s.makeDefaultHandler()
	assert.Equal(t, "test", h.Package)
}

func TestDefaultHandlerExecute(t *testing.T) {
	h := defaultHandler{
		Package: "test",
	}

	b := new(bytes.Buffer)
	err := h.execute(b)
	assert.NoError(t, err)
	assert.Equal(t, testDefaultHandlerSource, b.String())
}

const (
	testEndpointHandlerSnippet = `	v0, in, err := pdp.GetInfoRequestBooleanValue(in)
	if err != nil {
		return 0, err
	}
`
	testEndpointHandlerSource = `// Package test is a generated PIP server handler package. DO NOT EDIT.
package test

import (
	"errors"
	"github.com/infobloxopen/themis/pdp"
)

const reqTestArgs = 1

var errInvalidTestArgCount = errors.New("invalid count of request arguments for test endpoint")

func handleTest(c int, in, b []byte, e Endpoints) (int, error) {
	if c != reqTestArgs {
		return 0, errInvalidTestArgCount
	}

	v0, in, err := pdp.GetInfoRequestBooleanValue(in)
	if err != nil {
		return 0, err
	}
	v, err := e.Test(v0)
	if err != nil {
		return 0, err
	}

	n, err := pdp.MarshalInfoResponseString(b[:cap(b)], v)
	if err != nil {
		panic(err)
	}

	return n, nil
}
`

	testDefaultHandlerSource = `// Package test is a generated PIP server handler package. DO NOT EDIT.
package test

import "errors"

var errUnknownEndpoint = errors.New("unknown endpoint")

func handleDefault(c int, in, b []byte, e Endpoints) (int, error) {
	return 0, errUnknownEndpoint
}
`
)
