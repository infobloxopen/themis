// Package pkg of mkpipsrv utility provides data schema for generator input and
// generation logic itself.
package pkg

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v2"
)

// Schema is a root data structure for generator input. It holds name of package
// to generate and set of endpoints for data handlers. Currently, only single
// endpoint with "*" key accepted. The endpoint is responsible for processing of
// any request.
type Schema struct {
	Package   string
	Endpoints map[string]*Endpoint
}

// Endpoint defines input arguments for request and result of response.
// All arguments are required.
type Endpoint struct {
	goName string

	Args      []string
	goArgs    []string
	goArgList string
	goArgPkgs int

	goParsers    string
	goMarshaller string

	Result       string
	goResult     string
	goResultZero string
	goResultPkg  int
}

// NewSchemaFromFile reads schema from given YAML file.
func NewSchemaFromFile(path string) (s *Schema, err error) {
	f, err := os.Open(filepath.Clean(path))
	if err != nil {
		return
	}

	defer func() {
		if cErr := f.Close(); err == nil {
			err = cErr
		}
	}()

	s = new(Schema)

	err = yaml.NewDecoder(f).Decode(s)
	if err != nil {
		return
	}

	err = s.postProcess()

	return
}

func (s *Schema) postProcess() error {
	k, v, err := s.getFirstEndpoint()
	if err != nil {
		return err
	}

	if len(s.Endpoints) == 1 && isDefaultEndpoint(k) {
		if err := v.postProcess(k, true); err != nil {
			return fmt.Errorf("endpoint %q: %s", k, err)
		}

		return nil
	}

	for k, v := range s.Endpoints {
		if err := v.postProcess(k, false); err != nil {
			return fmt.Errorf("endpoint %q: %s", k, err)
		}
	}

	return nil
}
