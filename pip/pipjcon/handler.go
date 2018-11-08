package main

import (
	"errors"

	log "github.com/sirupsen/logrus"

	"github.com/infobloxopen/themis/pdp"
)

var errNotImplemented = errors.New("not implemented")

const reqIdSize = 4

func handler(b []byte) []byte {
	if len(b) < reqIdSize {
		log.WithField("len(b)", len(b)).Fatal("missing request id")
	}

	out := b[reqIdSize:]
	n, err := pdp.MarshalInfoError(out[:cap(out)], errNotImplemented)
	if err != nil {
		log.WithError(err).Fatal("failed to marshal response")
	}

	return b[:reqIdSize+n]
}
