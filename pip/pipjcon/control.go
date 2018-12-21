package main

import (
	"context"
	"fmt"
	"io"
	"math"

	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"

	"github.com/infobloxopen/themis/pdp"
	pb "github.com/infobloxopen/themis/pdp-control"
)

func (s *srv) Request(ctx context.Context, in *pb.Item) (*pb.Response, error) {
	log.Info("control request")

	fromTag, err := newTag(in.FromTag)
	if err != nil {
		return ctrlTagErrorf("can't treat %q as current tag: %s", in.FromTag, err), nil
	}

	toTag, err := newTag(in.ToTag)
	if err != nil {
		return ctrlTagErrorf("can't treat %q as new tag: %s", in.ToTag, err), nil
	}

	if fromTag != nil && toTag == nil {
		return ctrlTagErrorf("can't update from %q to no tag", in.FromTag), nil
	}

	switch in.Type {
	case pb.Item_POLICIES:
		return ctrlError("can't accept policy update request"), nil

	case pb.Item_CONTENT:
		id, err := s.contentRequest(in.Id, fromTag, toTag)
		if err == errUpdateIdxOverflow {
			return ctrlError(err.Error()), nil
		}

		if err != nil {
			return ctrlTagError(err.Error()), nil
		}

		log.WithField("req-id", id).Info("request has been registered")

		return ctrlAckID(id), nil
	}

	return ctrlErrorf("unknown upload request type: %d", in.Type), nil
}

func (s *srv) Upload(stream pb.PDPControl_UploadServer) error {
	log.Info("data stream")

	chunk, err := stream.Recv()
	if err == io.EOF {
		return stream.SendAndClose(ctrlError("empty upload"))
	}

	if err != nil {
		return err
	}

	id := chunk.Id
	log.WithField("req-id", id).Info("uploading data for request")

	u, err := s.getUpdate(id)
	if err != nil {
		return stream.SendAndClose(ctrlError(err.Error()))
	}
	defer func() {
		s.Lock()
		defer s.Unlock()

		u.inUse = false
	}()

	if err = u.upload(newStreamReader(stream, chunk)); err != nil {
		return stream.SendAndClose(ctrlError(err.Error()))
	}

	return stream.SendAndClose(ctrlAckID(id))
}

func (s *srv) Apply(ctx context.Context, in *pb.Update) (*pb.Response, error) {
	log.WithField("req-id", in.Id).Info("apply command")

	u, err := s.getUpdate(in.Id)
	if err != nil {
		return ctrlError(err.Error()), nil
	}
	defer func() {
		s.Lock()
		defer s.Unlock()

		u.inUse = false
	}()

	if u.c != nil {
		s.Lock()
		s.c = s.c.Add(u.c)
		s.u = nil
		s.Unlock()

		if u.toTag == nil {
			log.WithField("ctn-id", u.id).Info("new content has been applied")
		} else {
			log.WithFields(log.Fields{
				"ctn-id": u.id,
				"tag":    u.toTag,
			}).Info("new content has been applied")
		}

		return ctrlAckID(in.Id), nil
	}

	if u.t != nil {
		s.Lock()
		c, err := u.t.Commit(s.c)
		if err != nil {
			s.Unlock()

			return ctrlErrorf("failed to commit content %q transaction %d from tag %q to %q: %s",
				u.id, in.Id, u.fromTag, u.toTag, err), nil
		}

		s.c = c
		s.u = nil
		s.Unlock()

		log.WithFields(log.Fields{
			"req-id":   in.Id,
			"ctn-id":   u.id,
			"prev-tag": u.fromTag,
			"curr-tag": u.toTag,
		}).Info("content update has been applied")

		return ctrlAckID(in.Id), nil
	}

	return ctrlErrorf("request %d doesn't contain parsed content %q or parsed content update", in.Id, u.id), nil
}

func (s *srv) NotifyReady(context.Context, *pb.Empty) (*pb.Response, error) {
	s.once.Do(s.startSrv)
	return ctrlAck(), nil
}

func (s *srv) contentRequest(id string, fromTag, toTag *uuid.UUID) (int32, error) {
	s.Lock()
	defer s.Unlock()

	if s.uIdx >= math.MaxInt32 {
		return 0, errUpdateIdxOverflow
	}

	t, err := s.newTransaction(id, fromTag)
	if err != nil {
		return 0, err
	}

	s.uIdx++
	s.u = newUpdate(id, fromTag, toTag, t)

	return s.uIdx, nil
}

func (s *srv) newTransaction(id string, fromTag *uuid.UUID) (*pdp.LocalContentStorageTransaction, error) {
	if fromTag != nil {
		return s.c.NewTransaction(id, fromTag)
	}

	return nil, nil
}

func newTag(s string) (*uuid.UUID, error) {
	if len(s) > 0 {
		t, err := uuid.Parse(s)
		if err != nil {
			return nil, err
		}

		return &t, nil
	}

	return nil, nil
}

func ctrlAck() *pb.Response {
	return ctrlAckID(0)
}

func ctrlAckID(id int32) *pb.Response {
	return &pb.Response{Status: pb.Response_ACK, Id: id}
}

func ctrlTagErrorf(f string, args ...interface{}) *pb.Response {
	return ctrlTagError(fmt.Sprintf(f, args...))
}

func ctrlTagError(s string) *pb.Response {
	return ctrlStatusError(pb.Response_TAG_ERROR, s)
}

func ctrlErrorf(f string, args ...interface{}) *pb.Response {
	return ctrlError(fmt.Sprintf(f, args...))
}

func ctrlError(s string) *pb.Response {
	return ctrlStatusError(pb.Response_ERROR, s)
}

func ctrlStatusError(status pb.Response_Status, s string) *pb.Response {
	return &pb.Response{
		Status:  status,
		Id:      -1,
		Details: s,
	}
}
