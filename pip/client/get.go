package client

import (
	"errors"

	"github.com/infobloxopen/themis/pdp"
)

var errNotImplemented = errors.New("not implemented")

func (c *client) Get(args ...pdp.AttributeAssignment) (pdp.AttributeValue, error) {
	return pdp.UndefinedValue, errNotImplemented
}
