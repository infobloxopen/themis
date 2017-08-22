package jcon

import (
	"encoding/json"
	"io"

	"github.com/google/uuid"

	"github.com/infobloxopen/themis/pdp"
)

func Unmarshal(r io.Reader, tag *uuid.UUID) (*pdp.LocalContent, error) {
	c := &content{}
	err := c.unmarshal(json.NewDecoder(r))
	if err != nil {
		return nil, err
	}

	return pdp.NewLocalContent(c.id, tag, c.items), nil
}

func UnmarshalUpdate(r io.Reader, cID string, oldTag, newTag uuid.UUID) (*pdp.ContentUpdate, error) {
	u := pdp.NewContentUpdate(cID, oldTag, newTag)
	err := unmarshalCommands(json.NewDecoder(r), u)
	if err != nil {
		return nil, err
	}

	return u, nil
}
