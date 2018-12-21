package pkg

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCollectImports(t *testing.T) {
	pkgs := collectImports(
		goTypeBool,
		goTypeString,
		goTypeInt64,
		goTypeFloat64,
		goTypeNetIP,
		goTypeNetIPNet,
		goTypeDomainName,
		goTypeStrtree,
		goTypeIPTree,
		goTypeDomainTree,
		goTypeSliceOfStrings,
	)
	assert.Equal(t,
		goPkgNetMask|
			goPkgDomainMask|
			goPkgStrtreeMask|
			goPkgIPTreeMask|
			goPkgDomainTreeMask,
		pkgs,
	)
}

func TestMakeImports(t *testing.T) {
	imports := makeImports(goPkgNetMask | goPkgStrtreeMask | goPkgDomainTreeMask)
	assert.ElementsMatch(t, []string{
		goPkgNetName,
		goPkgStrtreeName,
		goPkgDomainTreeName,
	}, imports)

	imports = makeImports(goPkgDomainMask|goPkgIPTreeMask, "fmt", "log")
	assert.ElementsMatch(t, []string{
		"fmt",
		"log",
		goPkgDomainName,
		goPkgIPTreeName,
	}, imports)
}

func TestFixImports(t *testing.T) {
	tmp := writeTestGoFile(t, testGoFile)
	defer func() {
		if err := os.Remove(tmp); err != nil {
			assert.FailNow(t, "os.Remove(tmp): %q", err)
		}
	}()

	dir, name := filepath.Split(tmp)
	assert.NoError(t, fixImports(dir, name), "dir: %q, file: %q", dir, name)
}

func TestFixImportsWithInvaildFile(t *testing.T) {
	tmp := writeTestGoFile(t, testWrongGoFile)
	defer func() {
		if err := os.Remove(tmp); err != nil {
			assert.FailNow(t, "os.Remove(tmp): %q", err)
		}
	}()

	dir, name := filepath.Split(tmp)
	err := fixImports(dir, name)
	if assert.Error(t, err) {
		assert.Contains(t, err.Error(), name)
	}
}

const (
	testGoFile = `package main

import (
	"strings"
	"fmt"
)

func main() {
	log.Printf("%q", strings.Joint("one", "two", "three"))
}
`

	testWrongGoFile = `package main

import (
	"strings"
	"fmt"
)

func main() {
`
)

func writeTestGoFile(t *testing.T, s string) string {
	tmp, err := ioutil.TempFile("", "*.go")
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
