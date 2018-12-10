package pkg

import (
	"bytes"
	"io/ioutil"
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSchemaGenDispatcher(t *testing.T) {
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

	err = s.genDispatcher(tmp)
	assert.NoError(t, err)
	assert.FileExists(t, path.Join(tmp, dispatcherDst))
}

func TestSchemaMakeDispatcher(t *testing.T) {
	pt := &Endpoint{
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
	pe := &Endpoint{
		goName: "Example",
		Args: []string{
			pipTypeAddress,
		},
		goArgs: []string{
			goTypeNetIP,
		},
		goArgPkgs: goPkgNetMask,

		Result:       pipTypeNetwork,
		goResult:     goTypeNetIPNet,
		goResultZero: "nil",
		goResultPkg:  goPkgNetMask,
	}
	s := &Schema{
		Package: "test",
		Endpoints: map[string]*Endpoint{
			"test":    pt,
			"example": pe,
		},
	}

	d := s.makeDispatcher()
	assert.Equal(t, "test", d.Package)
	assert.Equal(t, testDispatcherHandlersSnippet, d.Handlers)
}

func TestDispatcherExecute(t *testing.T) {
	h := dispatcher{
		Package:  "test",
		Handlers: testDispatcherHandlersSnippet,
	}

	b := new(bytes.Buffer)
	err := h.execute(b)
	assert.NoError(t, err)
	assert.Equal(t, testDispatcherSource, b.String())
}

const (
	testDispatcherHandlersSnippet = `
	case "example":
		n, err = handleExample(c, in, b, e)

	case "test":
		n, err = handleTest(c, in, b, e)
`

	testDispatcherSource = `// Package test is a generated PIP server handler package. DO NOT EDIT.
package test

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

	case "example":
		n, err = handleExample(c, in, b, e)

	case "test":
		n, err = handleTest(c, in, b, e)
	}

	return n, err
}
`
)
