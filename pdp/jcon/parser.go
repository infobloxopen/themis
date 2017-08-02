package jcon

import (
	"encoding/json"
	"io"

	"github.com/infobloxopen/go-trees/strtree"
)

func Unmarshal(r io.Reader) (string, *strtree.Tree, error) {
	c := &content{}
	err := c.unmarshal(json.NewDecoder(r))
	if err != nil {
		return "", nil, err
	}

	return c.ID, c.items, nil
}
