package pkg

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"testing"
	"unicode/utf8"

	"github.com/stretchr/testify/assert"
)

func TestSchemaGetFirstEndpoint(t *testing.T) {
	p := new(Endpoint)
	s := &Schema{
		Package: "test",
		Endpoints: map[string]*Endpoint{
			defaultEndpointAlias: p,
		},
	}
	k, v, err := s.getFirstEndpoint()
	assert.NoError(t, err)
	assert.Equal(t, defaultEndpointAlias, k)
	assert.Equal(t, p, v)

	s = &Schema{
		Package: "test",
	}
	k, v, err = s.getFirstEndpoint()
	assert.Equal(t, errNoEndpoints, err, "k: %q, v: %#v", k, v)
}

func TestEndpointPostProcess(t *testing.T) {
	p := &Endpoint{
		Args: []string{
			pipTypeAddress,
		},
		Result: pipTypeNetwork,
	}

	err := p.postProcess(defaultEndpointAlias, false)
	assert.NoError(t, err)
	assert.Equal(t, []string{goTypeNetIP}, p.goArgs)
	assert.Equal(t, goPkgNetMask, p.goArgPkgs)
	assert.Equal(t, goTypeNetIPNet, p.goResult)
	assert.Equal(t, "0", p.goResultZero)
	assert.Equal(t, goPkgNetMask, p.goResultPkg)
}

func TestEndpointPostProcessSingle(t *testing.T) {
	p := &Endpoint{
		Args: []string{
			pipTypeAddress,
		},
		Result: pipTypeNetwork,
	}

	err := p.postProcess(defaultEndpointAlias, true)
	assert.NoError(t, err)
	assert.Equal(t, []string{goTypeNetIP}, p.goArgs)
	assert.Equal(t, goPkgNetMask, p.goArgPkgs)
	assert.Equal(t, goTypeNetIPNet, p.goResult)
	assert.Equal(t, "nil", p.goResultZero)
	assert.Equal(t, goPkgNetMask, p.goResultPkg)
}

func TestEndpointPostProcessWithInvalidName(t *testing.T) {
	p := &Endpoint{
		Args: []string{
			pipTypeAddress,
		},
		Result: pipTypeNetwork,
	}

	err := p.postProcess("", false)
	assert.Equal(t, errEmptyEndpointName, err)
}

func TestEndpointPostProcessWithInvalidArg(t *testing.T) {
	p := &Endpoint{
		Args:   []string{"unknown"},
		Result: pipTypeNetwork,
	}

	err := p.postProcess(defaultEndpointAlias, false)
	assert.EqualError(t, err, "argument 0: unknown type \"unknown\"", "endpoint: %#v", p)
}

func TestEndpointPostProcessWithInvalidResult(t *testing.T) {
	p := &Endpoint{
		Args:   []string{pipTypeAddress},
		Result: "unknown",
	}

	err := p.postProcess(defaultEndpointAlias, false)
	assert.EqualError(t, err, "result: unknown type \"unknown\"", "endpoint: %#v", p)
}

func TestSchemaGenEndpointsInterface(t *testing.T) {
	tmp, err := ioutil.TempDir("", "")
	if err != nil {
		assert.FailNow(t, "ioutil.TempDir(\"\", \"\"): %q", err)
	}

	defer func() {
		assert.NoError(t, os.RemoveAll(tmp))
	}()

	s := &Schema{
		Package: "test",
		Endpoints: map[string]*Endpoint{
			"test": {
				Args:   []string{"String"},
				Result: "String",
			},
		},
	}

	if err = s.postProcess(); err != nil {
		assert.FailNow(t, "s.(*Schema).postProcess(): %q", err)
	}

	err = s.genEndpointsInterface(tmp)
	assert.NoError(t, err)
	assert.FileExists(t, path.Join(tmp, endpointsDst))
}

func TestMakeEndpointsInterface(t *testing.T) {
	p := &Endpoint{
		goName: "Test",
		Args: []string{
			pipTypeNetwork,
		},
		goArgs: []string{
			goTypeNetIPNet,
		},
		goArgPkgs: goPkgNetMask,

		Result:       pipTypeString,
		goResult:     goTypeString,
		goResultZero: "\"\"",
		goResultPkg:  0,
	}
	s := &Schema{
		Package: "test",
		Endpoints: map[string]*Endpoint{
			"test": p,
		},
	}

	ei := s.makeEndpointsInterface()
	assert.Equal(t, "test", ei.Package)
	assert.Contains(t, ei.Imports, goPkgNetName)
	assert.Contains(t, ei.Methods, fmt.Sprintf("Test(%s) (%s, error)", goTypeNetIPNet, goTypeString))
}

func TestEndpointsInterfaceExecute(t *testing.T) {
	ei := endpointsInterface{
		Package: "test",
		Imports: goPkgNetName,
		Methods: fmt.Sprintf("Test(%s) (%s, error)", goTypeNetIPNet, goTypeString),
	}

	b := new(bytes.Buffer)
	err := ei.execute(b)
	assert.NoError(t, err)
	assert.Equal(t, testEndpointsInterfaceSource, b.String())
}

var testEndpointsInterfaceSource = `// Package test is a generated PIP server handler package. DO NOT EDIT.
package test

import (
	"net"
)

type Endpoints interface {
	Test(*net.IPNet) (string, error)
}
`

func TestMakeGoName(t *testing.T) {
	s, err := makeGoName("test-endpoint-number-1")
	assert.NoError(t, err)
	assert.Equal(t, "TestEndpointNumber1", s)

	_, err = makeGoName("")
	assert.Equal(t, errEmptyEndpointName, err)
}

func TestGetFirstRuneForExport(t *testing.T) {
	r, n, err := getFirstRuneForExport("test")
	assert.NoError(t, err)
	assert.Equal(t, 'T', r)
	assert.Equal(t, n, 1)

	_, _, err = getFirstRuneForExport("_test")
	assert.EqualError(t, err, "failed to make exportable name from \"_test\"")

	_, _, err = getFirstRuneForExport("")
	assert.Equal(t, errEmptyEndpointName, err)
}

func TestAppendRune(t *testing.T) {
	b := appendRune(make([]byte, 0, 1), 'X')
	assert.Equal(t, []byte("X"), b)

	s := "test"
	b = make([]byte, len(s), len(s)+2)
	copy(b, []byte(s))
	b = appendRune(b, '符')
	assert.Equal(t, []byte(s+"符"), b)

	b = []byte{}
	b = appendRune(b, '符')
	assert.Equal(t, []byte("符"), b)

	b = appendRune(b, utf8.MaxRune+1)
	assert.Equal(t, []byte("符"), b)
}
