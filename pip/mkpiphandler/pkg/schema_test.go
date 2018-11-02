package pkg

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewSchemaFromFile(t *testing.T) {
	tmp := writeTestSchema(t, testYAMLSchema)
	defer func() {
		assert.NoError(t, os.Remove(tmp))
	}()

	s, err := NewSchemaFromFile(tmp)
	assert.NoError(t, err)
	assert.Equal(t, &Schema{
		Package: "test",
		Endpoints: map[string]*Endpoint{
			"*": {
				Args: []string{
					"boolean",
					"integer",
					"network",
				},
				goArgs: []string{
					goTypeBool,
					goTypeInt64,
					goTypeNetIPNet,
				},
				goArgPkgs: goPkgNetMask,

				Result:       "set of domains",
				goResult:     goTypeDomainTree,
				goResultZero: "nil",
				goResultPkg:  goPkgDomainTreeMask,
			},
		},
	}, s)
}

func TestNewSchemaFromFileWithNoFile(t *testing.T) {
	tmp := writeTestSchema(t, "")
	if err := os.Remove(tmp); err != nil {
		assert.FailNow(t, "os.Remove(tmp): %q", err)
	}

	s, err := NewSchemaFromFile(tmp)
	assert.Error(t, err, "schema: %#v", s)
}

func TestNewSchemaFromFileWithInvalidYAML(t *testing.T) {
	tmp := writeTestSchema(t, "\t")
	defer func() {
		assert.NoError(t, os.Remove(tmp))
	}()

	s, err := NewSchemaFromFile(tmp)
	assert.Error(t, err, "schema: %#v", s)
}

func TestSchemaPostProcess(t *testing.T) {
	s := &Schema{
		Package: "test",
		Endpoints: map[string]*Endpoint{
			"*": {
				Args: []string{
					"boolean",
					"integer",
					"network",
				},
				Result: "set of domains",
			},
		},
	}

	err := s.postProcess()
	assert.NoError(t, err)
	assert.Equal(t, &Schema{
		Package: "test",
		Endpoints: map[string]*Endpoint{
			"*": {
				Args: []string{
					"boolean",
					"integer",
					"network",
				},
				goArgs: []string{
					goTypeBool,
					goTypeInt64,
					goTypeNetIPNet,
				},
				goArgPkgs: goPkgNetMask,

				Result:       "set of domains",
				goResult:     goTypeDomainTree,
				goResultZero: "nil",
				goResultPkg:  goPkgDomainTreeMask,
			},
		},
	}, s)
}

func TestSchemaPostProcessWithSeveralEndpoints(t *testing.T) {
	s := &Schema{
		Package: "test",
		Endpoints: map[string]*Endpoint{
			"test": {
				Args: []string{
					"set of domains",
				},
				Result: "list of strings",
			},
			"*": {
				Args: []string{
					"boolean",
					"integer",
					"network",
				},
				Result: "set of domains",
			},
		},
	}

	err := s.postProcess()
	assert.Error(t, err, "schema: %#v", s)
}

func TestSchemaPostProcessWithInvalidPath(t *testing.T) {
	s := &Schema{
		Package: "test",
		Endpoints: map[string]*Endpoint{
			"test": {
				Args: []string{
					"set of domains",
				},
				Result: "list of strings",
			},
		},
	}

	err := s.postProcess()
	assert.Error(t, err, "schema: %#v", s)
}

func TestSchemaPostProcessWithInvalidResult(t *testing.T) {
	s := &Schema{
		Package: "test",
		Endpoints: map[string]*Endpoint{
			"*": {
				Result: "unknown",
			},
		},
	}

	err := s.postProcess()
	assert.Error(t, err, "schema: %#v", s)
}

const testYAMLSchema = `# Test schema for PIP server handler
package: test
endpoints:
  "*":
    args:
    - boolean
    - integer
    - network
    result: set of domains
`

func writeTestSchema(t *testing.T, s string) string {
	tmp, err := ioutil.TempFile("", "")
	if err != nil {
		assert.FailNow(t, "ioutil.TempFile(\"\", \"\"): %q", err)
	}
	defer func() {
		if err = tmp.Close(); err != nil {
			assert.FailNow(t, "tmp.Close(): %q", err)
		}
	}()

	if _, err = tmp.Write([]byte(s)); err != nil {
		assert.FailNow(t, "tmp.Write([]byte{s}): %q", err)
	}

	return tmp.Name()
}
