package jcon

import (
	"encoding/json"
	"strings"

	"github.com/infobloxopen/themis/pdp"
)

type content struct {
	id    string
	items []*pdp.ContentItem
}

func (c *content) bindError(err error) error {
	if len(c.id) > 0 {
		return bindError(err, c.id)
	}

	return bindError(err, "content")
}

func (c *content) unmarshal(d *json.Decoder) error {
	ok, err := checkRootObjectStart(d)
	if err != nil {
		return c.bindError(err)
	}

	if !ok {
		return nil
	}

	err = unmarshalObject(d, func(k string, d *json.Decoder) error {
		switch strings.ToLower(k) {
		case "id":
			return c.unmarshalIDField(d)

		case "items":
			return c.unmarshalItemsField(d)
		}

		return newUnknownContentFieldError(k)
	}, "root")
	if err != nil {
		return c.bindError(err)
	}

	err = checkEOF(d)
	if err != nil {
		return c.bindError(err)
	}

	return nil
}

func (c *content) unmarshalIDField(d *json.Decoder) error {
	id, err := getString(d, "content id")
	if err != nil {
		return err
	}

	c.id = id
	return nil
}

func (c *content) unmarshalItemsField(d *json.Decoder) error {
	err := checkObjectStart(d, "content items")
	if err != nil {
		return err
	}

	items := []*pdp.ContentItem{}
	err = unmarshalObject(d, func(k string, d *json.Decoder) error {
		v, err := unmarshalContentItem(k, d)
		if err != nil {
			return bindError(err, k)
		}

		items = append(items, v)

		return nil
	}, "content items")
	if err != nil {
		return err
	}

	c.items = items
	return nil
}
