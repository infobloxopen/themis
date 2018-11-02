package pkg

import (
	"os"
	"path"
)

// Generate creates a package according to the schema inside given directory.
// It makes a subdirectory with name of the package in the given directory and
// places code to the subdirectory.
func (s *Schema) Generate(root string) error {
	dir := path.Join(root, s.Package)
	if err := os.RemoveAll(dir); err != nil {
		return err
	}

	if err := os.MkdirAll(dir, 0750); err != nil {
		return err
	}

	if err := s.genHandler(dir); err != nil {
		return err
	}

	return fixImports(dir, handlerDst)
}
