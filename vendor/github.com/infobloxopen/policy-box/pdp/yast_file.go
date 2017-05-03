package pdp

import (
	"os"
	"path/filepath"
)

func findAndOpenFile(path string, dir string) (*os.File, error) {
	if len(dir) < 1 || filepath.IsAbs(path) {
		return os.Open(path)
	}

	f, err := os.Open(path)
	if err == nil {
		return f, nil
	}

	if !os.IsNotExist(err) {
		return nil, err
	}

	return os.Open(filepath.Join(dir, path))
}
