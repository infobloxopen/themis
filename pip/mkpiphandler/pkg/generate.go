package pkg

import (
	"fmt"
	"io"
	"os"
	"path"
)

// Generate creates a package according to the schema inside given directory.
// It makes a subdirectory with name of the package in the given directory and
// places code to the subdirectory.
func (s *Schema) Generate(root string) error {
	return inSubDirectory(root, s.Package, func(dir string) error {
		_, p, err := s.getFirstEndpoint()
		if err != nil {
			return err
		}

		if err = s.genHandler(dir, p); err != nil {
			return err
		}

		return fixImports(dir, handlerDst)
	})
}

func inSubDirectory(root, name string, f func(string) error) (err error) {
	dir := path.Join(root, name)

	err = os.RemoveAll(dir)
	if err != nil {
		return
	}

	err = os.MkdirAll(dir, 0750)
	if err != nil {
		return
	}
	defer func() {
		if err != nil {
			if rErr := os.RemoveAll(dir); rErr != nil {
				err = fmt.Errorf("%s; %s", err, rErr)
			}
		}
	}()

	err = f(dir)
	return
}

func toFile(dir, name string, f func(io.Writer) error) (err error) {
	w, err := os.Create(path.Join(dir, name))
	if err != nil {
		return
	}
	defer func() {
		cErr := w.Close()
		if cErr != nil {
			if err != nil {
				err = fmt.Errorf("%s; %s", err, cErr)
			} else {
				err = cErr
			}
		}
	}()

	err = f(w)
	return
}
