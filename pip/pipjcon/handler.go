package main

import (
	"fmt"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/infobloxopen/themis/pdp"
)

const reqIDSize = 4

func (s *srv) handler(b []byte) []byte {
	if len(b) < reqIDSize {
		log.WithField("len(b)", len(b)).Fatal("missing request id")
	}

	in := b[reqIDSize:]

	v, err := s.process(in)
	if err != nil {
		n, err := pdp.MarshalInfoError(in[:cap(in)], err)
		if err != nil {
			log.WithError(err).Fatal("failed to marshal response")
		}

		return b[:reqIDSize+n]
	}

	n, err := pdp.MarshalInfoResponse(in[:cap(in)], v)
	if err != nil {
		log.WithError(err).Fatal("failed to marshal response")
	}

	return b[:reqIDSize+n]
}

func (s *srv) process(b []byte) (pdp.AttributeValue, error) {
	ab := s.a.Get()
	defer s.a.Put(ab)

	args := ab.a
	path, n, err := pdp.UnmarshalInfoRequest(b, args)
	if err != nil {
		return pdp.UndefinedValue, err
	}

	if path[0] == '/' {
		path = path[1:]
	}

	loc := strings.Split(path, "/")
	if len(loc) != 2 {
		return pdp.UndefinedValue, fmt.Errorf("expected path in form of <Content-ID>/<Item-ID> got %s", path)
	}

	s.RLock()
	c := s.c
	s.RUnlock()

	item, err := c.Get(loc[0], loc[1])
	if err != nil {
		return pdp.UndefinedValue, err
	}

	return item.GetByValues(args[:n], pdp.AggTypeDisable)
}
