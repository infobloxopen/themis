package pkg

import (
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

	s := &Schema{
		Package: "test",
		Endpoints: map[string]*Endpoint{
			"*": {
				Args: []string{
					"Boolean",
					"Address",
					"Domain",
				},
				Result: "List of Strings",
			},
		},
	}
	err = s.postProcess()
	if err != nil {
		assert.FailNow(t, "s.(*Schema).postProcess(): %q", err)
	}

	err = s.genHandler(tmp)
	if assert.NoError(t, err) {
		f, err := os.Open(path.Join(tmp, handlerDst))
		if assert.NoError(t, err) {
			defer func() {
				assert.NoError(t, f.Close())
			}()
		}
	}
}

func TestSchemaGenHandlerWithInvalidDirectory(t *testing.T) {
	s := &Schema{
		Package: "test",
		Endpoints: map[string]*Endpoint{
			"*": {
				Args: []string{
					"Boolean",
					"Address",
					"Domain",
				},
				Result: "List of Strings",
			},
		},
	}
	err := s.postProcess()
	if err != nil {
		assert.FailNow(t, "s.(*Schema).postProcess(): %q", err)
	}

	err = s.genHandler("/dev/null")
	if !assert.Error(t, err) {
		if err = os.Remove(path.Join("/dev/null", handlerDst)); err != nil {
			assert.FailNow(t, "os.Remove(path.Join(\"/dev/null\", handlerDst)): %q", err)
		}
	}
}

func TestSchemaGenHandlerWithNoEndpoints(t *testing.T) {
	tmp, err := ioutil.TempDir("", "")
	if err != nil {
		assert.FailNow(t, "ioutil.TempDir(\"\", \"\"): %q", err)
	}

	defer func() {
		assert.NoError(t, os.RemoveAll(tmp))
	}()

	s := &Schema{
		Package: "test",
	}

	err = s.genHandler(tmp)
	assert.Equal(t, errNoEndpoints, err)
}

func TestSchemaGenHandlerWithInvalidEndpoint(t *testing.T) {
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
			"*": {
				Args:   []string{"unknown"},
				Result: "unknown",
			},
		},
	}

	err = s.genHandler(tmp)
	assert.EqualError(t, err, "argument 0: unknown type \"unknown\"")
}
