package pkg

import (
	"bytes"
	"io/ioutil"
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSchemaGenHandler(t *testing.T) {
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

	err = s.genHandler(tmp)
	assert.NoError(t, err)
	assert.FileExists(t, path.Join(tmp, handlerDst))
}

func TestSchemaMakeHandler(t *testing.T) {
	p := &Endpoint{
		goName: "Test",
		Args: []string{
			pipTypeString,
		},
		goArgs: []string{
			goTypeString,
		},

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

	h := s.makeHandler()
	assert.Equal(t, "test", h.Package)
}

func TestHandlerExecute(t *testing.T) {
	h := handler{
		Package: "test",
	}

	b := new(bytes.Buffer)
	err := h.execute(b)
	assert.NoError(t, err)
	assert.Equal(t, testHandlerSource, b.String())
}

const testHandlerSource = `// Package test is a generated PIP server handler package. DO NOT EDIT.
package test

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
`
