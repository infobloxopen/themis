package main

import (
	"errors"
	"fmt"

	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"

	"github.com/infobloxopen/themis/pdp"
	"github.com/infobloxopen/themis/pdp/jcon"
)

var (
	errUpdateIdxOverflow = errors.New("update index overflow")
	errNoActiveUpdate    = errors.New("no active update")
)

type update struct {
	id    string
	inUse bool

	fromTag *uuid.UUID
	toTag   *uuid.UUID

	c *pdp.LocalContent
	t *pdp.LocalContentStorageTransaction
}

func newUpdate(id string, fromTag, toTag *uuid.UUID, t *pdp.LocalContentStorageTransaction) *update {
	return &update{
		id:      id,
		fromTag: fromTag,
		toTag:   toTag,
		t:       t,
	}
}

func (s *srv) getUpdate(id int32) (*update, error) {
	s.Lock()
	defer s.Unlock()

	u := s.u
	if u == nil {
		return nil, errNoActiveUpdate
	}

	if s.uIdx != id {
		return nil, fmt.Errorf("no update with id %d", id)
	}

	if u.inUse {
		return nil, fmt.Errorf("update %d is already in use", id)
	}

	u.inUse = true
	return u, nil
}

func (u *update) upload(r *streamReader) error {
	if u.fromTag == nil {
		return u.uploadSnapshot(r)
	}

	return u.uploadDiff(r)
}

func (u *update) uploadSnapshot(r *streamReader) error {
	c, err := jcon.Unmarshal(r, u.toTag)
	if err != nil {
		r.skip()
		return err
	}

	u.c = c
	log.WithField("size", r.size).Info("stream has been read and parsed as snapshot")
	return nil
}

func (u *update) uploadDiff(r *streamReader) error {
	d, err := jcon.UnmarshalUpdate(r, u.id, *u.fromTag, *u.toTag, u.t.Symbols())
	if err != nil {
		r.skip()
		return err
	}

	if err = u.t.Apply(d); err != nil {
		return err
	}

	log.WithField("size", r.size).Info("stream has been read and parsed as update")
	return nil
}
