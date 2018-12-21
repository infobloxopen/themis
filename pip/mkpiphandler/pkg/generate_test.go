package pkg

import (
	"io/ioutil"
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSchemaGenerate(t *testing.T) {
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
				Result: "Set of Networks",
			},
		},
	}
	err = s.postProcess()
	if err != nil {
		assert.FailNow(t, "s.(*Schema).postProcess(): %q", err)
	}

	err = s.Generate(tmp)
	if assert.NoError(t, err) {
		fi, err := os.Stat(path.Join(tmp, "test"))
		assert.NoError(t, err)
		if assert.True(t, fi.Mode().IsDir()) {
			f, err := os.Open(path.Join(tmp, "test", handlerDst))
			if assert.NoError(t, err) {
				defer func(f *os.File) {
					assert.NoError(t, f.Close())
				}(f)

				/*b, err := ioutil.ReadAll(f)
				if assert.NoError(t, err) {
					t.Logf("%s:\n%s", handlerDst, string(b))
				}*/
			}
		}
	}
}

func TestSchemaGenerateWithInvalidDirectory(t *testing.T) {
	s := &Schema{
		Package: "null",
		Endpoints: map[string]*Endpoint{
			"*": {
				Args: []string{
					"Boolean",
					"Address",
					"Domain",
				},
				Result: "Set of Networks",
			},
		},
	}
	err := s.postProcess()
	if err != nil {
		assert.FailNow(t, "s.(*Schema).postProcess(): %q", err)
	}

	err = s.Generate("/dev")
	assert.Error(t, err)

	s.Package = "test"
	err = s.Generate("/dev/null")
	if !assert.Error(t, err) {
		assert.NoError(t, os.RemoveAll(path.Join("/dev/null", "test")))
	}
}

func TestSchemaGenerateWithInvalidSchema(t *testing.T) {
	tmp, err := ioutil.TempDir("", "")
	if err != nil {
		assert.FailNow(t, "ioutil.TempDir(\"\", \"\"): %q", err)
	}

	defer func() {
		assert.NoError(t, os.RemoveAll(tmp))
	}()

	s := &Schema{
		Package: "null",
		Endpoints: map[string]*Endpoint{
			"*": {
				Args:   []string{"unknown"},
				Result: "unknown",
			},
		},
	}

	err = s.Generate(tmp)
	assert.Error(t, err)
}
