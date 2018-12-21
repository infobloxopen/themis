package pkg

import (
	"os"
	"path"
)

const handlerDst = "handler.go"

func (s *Schema) genHandler(dir string) (err error) {
	f, err := os.Create(path.Join(dir, handlerDst))
	if err != nil {
		return
	}

	defer func() {
		if cErr := f.Close(); err == nil {
			err = cErr
		}
	}()

	_, p, err := s.getFirstEndpoint()
	if err != nil {
		return
	}

	sh, err := s.makeSingleHandler(p)
	if err != nil {
		return
	}

	err = sh.execute(f)

	return
}
