package pkg

import (
	"errors"
	"io"
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
		if assert.DirExists(t, path.Join(tmp, "test")) {
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

func TestSchemaGenerateWithNoEndpoints(t *testing.T) {
	tmp, err := ioutil.TempDir("", "")
	if err != nil {
		assert.FailNow(t, "ioutil.TempDir(\"\", \"\"): %q", err)
	}

	defer func() {
		assert.NoError(t, os.RemoveAll(tmp))
	}()

	s := &Schema{
		Package:   "null",
		Endpoints: map[string]*Endpoint{},
	}

	err = s.Generate(tmp)
	assert.Equal(t, errNoEndpoints, err)
}

func TestInSubDirectory(t *testing.T) {
	tmp, err := ioutil.TempDir("", "")
	if err != nil {
		assert.FailNow(t, "ioutil.TempDir(\"\", \"\"): %q", err)
	}

	defer func() {
		assert.NoError(t, os.RemoveAll(tmp))
	}()

	err = inSubDirectory(tmp, "test", func(s string) error {
		assert.Contains(t, s, "test")
		return nil
	})
	assert.NoError(t, err)
	assert.DirExists(t, path.Join(tmp, "test"))
}

func TestInSubDirectoryWithError(t *testing.T) {
	tmp, err := ioutil.TempDir("", "")
	if err != nil {
		assert.FailNow(t, "ioutil.TempDir(\"\", \"\"): %q", err)
	}

	defer func() {
		assert.NoError(t, os.RemoveAll(tmp))
	}()

	tErr := errors.New("test error")
	err = inSubDirectory(tmp, "test", func(s string) error {
		return tErr
	})
	assert.Equal(t, tErr, err)

	_, err = os.Stat(path.Join(tmp, "test"))
	assert.Error(t, err)
}

func TestInSubDirectoryWithMkDirError(t *testing.T) {
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

	err = inSubDirectory(tmp, "test", func(s string) error {
		return nil
	})
	assert.Error(t, err)
}

func TestInSubDirectoryWithInitialCleanupError(t *testing.T) {
	tmp, err := ioutil.TempDir("", "")
	if err != nil {
		assert.FailNow(t, "ioutil.TempDir(\"\", \"\"): %q", err)
	}

	defer func() {
		if assert.NoError(t, os.Chmod(tmp, 0750)) {
			assert.NoError(t, os.RemoveAll(tmp))
		}
	}()

	dir := path.Join(tmp, "test")
	if err = os.MkdirAll(dir, 0750); !assert.NoError(t, err) {
		assert.Failf(t, "os.MkdirAll(%q): %q", dir, err)
	}

	if err = os.Chmod(tmp, 0550); !assert.NoError(t, err) {
		assert.Failf(t, "os.Chmod(%q): %q", tmp, err)
	}

	err = inSubDirectory(tmp, "test", func(s string) error {
		return nil
	})
	assert.Error(t, err)
}

func TestInSubDirectoryWithCleanupError(t *testing.T) {
	tmp, err := ioutil.TempDir("", "")
	if err != nil {
		assert.FailNow(t, "ioutil.TempDir(\"\", \"\"): %q", err)
	}

	defer func() {
		if assert.NoError(t, os.Chmod(tmp, 0750)) {
			assert.NoError(t, os.RemoveAll(tmp))
		}
	}()

	tErr := errors.New("test error")
	err = inSubDirectory(tmp, "test", func(s string) error {
		if err = os.Chmod(tmp, 0550); !assert.NoError(t, err) {
			assert.Failf(t, "os.Chmod(%q): %q", tmp, err)
		}

		return tErr
	})
	if assert.Error(t, err) {
		assert.Contains(t, err.Error(), tErr.Error())
	}
}

func TestToFile(t *testing.T) {
	tmp, err := ioutil.TempDir("", "")
	if err != nil {
		assert.FailNow(t, "ioutil.TempDir(\"\", \"\"): %q", err)
	}

	defer func() {
		assert.NoError(t, os.RemoveAll(tmp))
	}()

	err = toFile(tmp, "test", func(w io.Writer) error {
		assert.NotZero(t, w)
		return nil
	})
	assert.NoError(t, err)
	assert.FileExists(t, path.Join(tmp, "test"))
}

func TestToFileWithCreationError(t *testing.T) {
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

	err = toFile(tmp, "test", func(w io.Writer) error {
		return nil
	})
	assert.Error(t, err)

	_, err = os.Stat(path.Join(tmp, "test"))
	assert.Error(t, err)
}

func TestToFileWithClosingError(t *testing.T) {
	tmp, err := ioutil.TempDir("", "")
	if err != nil {
		assert.FailNow(t, "ioutil.TempDir(\"\", \"\"): %q", err)
	}

	defer func() {
		assert.NoError(t, os.RemoveAll(tmp))
	}()

	err = toFile(tmp, "test", func(w io.Writer) error {
		c, ok := w.(io.Closer)
		if assert.True(t, ok) {
			assert.NoError(t, c.Close())
		}

		return nil
	})
	assert.Error(t, err)
}

func TestToFileWithFuncAndClosingError(t *testing.T) {
	tmp, err := ioutil.TempDir("", "")
	if err != nil {
		assert.FailNow(t, "ioutil.TempDir(\"\", \"\"): %q", err)
	}

	defer func() {
		assert.NoError(t, os.RemoveAll(tmp))
	}()

	tErr := errors.New("test error")
	err = toFile(tmp, "test", func(w io.Writer) error {
		c, ok := w.(io.Closer)
		if assert.True(t, ok) {
			assert.NoError(t, c.Close())
		}

		return tErr
	})
	if assert.Error(t, err) {
		assert.Contains(t, err.Error(), tErr.Error())
	}
}
