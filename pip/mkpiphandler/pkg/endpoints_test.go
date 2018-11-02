package pkg

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSchemaGetFirstEndpoint(t *testing.T) {
	p := new(Endpoint)
	s := &Schema{
		Package: "test",
		Endpoints: map[string]*Endpoint{
			"*": p,
		},
	}
	k, v, err := s.getFirstEndpoint()
	assert.NoError(t, err)
	assert.Equal(t, "*", k)
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

	err := p.postProcess()
	assert.NoError(t, err)
	assert.Equal(t, []string{goTypeNetIP}, p.goArgs)
	assert.Equal(t, goPkgNetMask, p.goArgPkgs)
	assert.Equal(t, goTypeNetIPNet, p.goResult)
	assert.Equal(t, "nil", p.goResultZero)
	assert.Equal(t, goPkgNetMask, p.goResultPkg)
}

func TestEndpointPostProcessWithInvalidArg(t *testing.T) {
	p := &Endpoint{
		Args:   []string{"unknown"},
		Result: pipTypeNetwork,
	}

	err := p.postProcess()
	assert.EqualError(t, err, "argument 0: unknown type \"unknown\"", "endpoint: %#v", p)
}

func TestEndpointPostProcessWithInvalidResult(t *testing.T) {
	p := &Endpoint{
		Args:   []string{pipTypeAddress},
		Result: "unknown",
	}

	err := p.postProcess()
	assert.EqualError(t, err, "result: unknown type \"unknown\"", "endpoint: %#v", p)
}
