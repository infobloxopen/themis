package main

import (
	"io"
	"runtime/debug"

	log "github.com/Sirupsen/logrus"
	"github.com/google/uuid"
	"golang.org/x/net/context"

	"github.com/infobloxopen/themis/pdp"
	pb "github.com/infobloxopen/themis/pdp-control"
)

func controlFail(err error) *pb.Response {
	status := pb.Response_ERROR
	switch e := err.(type) {
	case *tagCheckError:
		switch e.err.(type) {
		case *pdp.UntaggedPolicyModificationError, *pdp.MissingPolicyTagError, *pdp.PolicyTagsNotMatchError, *pdp.UntaggedContentModificationError, *pdp.MissingContentTagError, *pdp.ContentTagsNotMatchError:
			status = pb.Response_TAG_ERROR
		}

	case *policyTransactionCreationError:
		switch e.err.(type) {
		case *pdp.UntaggedPolicyModificationError, *pdp.MissingPolicyTagError, *pdp.PolicyTagsNotMatchError:
			status = pb.Response_TAG_ERROR
		}
	case *contentTransactionCreationError:
		switch e.err.(type) {
		case *pdp.UntaggedContentModificationError, *pdp.MissingContentTagError, *pdp.ContentTagsNotMatchError:
			status = pb.Response_TAG_ERROR
		}
	}

	return &pb.Response{
		Status:  status,
		Id:      -1,
		Details: err.Error()}
}

func newTag(s string) (*uuid.UUID, error) {
	if len(s) <= 0 {
		return nil, nil
	}

	t, err := uuid.Parse(s)
	if err != nil {
		return nil, err
	}

	return &t, nil
}

func (s *server) Request(ctx context.Context, in *pb.Item) (*pb.Response, error) {
	log.Info("Got new control request")

	fromTag, err := newTag(in.FromTag)
	if err != nil {
		return controlFail(newInvalidFromTagError(in.FromTag, err)), nil
	}

	toTag, err := newTag(in.ToTag)
	if err != nil {
		return controlFail(newInvalidToTagError(in.ToTag, err)), nil
	}

	if fromTag != nil && toTag == nil {
		return controlFail(newInvalidTagsError(in.FromTag)), nil
	}

	var id int32
	switch in.Type {
	default:
		return controlFail(newUnknownUploadRequestError(in.Type)), nil

	case pb.Item_POLICIES:
		id, err = s.policyRequest(fromTag, toTag)

	case pb.Item_CONTENT:
		id, err = s.contentRequest(in.Id, fromTag, toTag)
	}

	if err != nil {
		return controlFail(err), nil
	}

	return &pb.Response{Status: pb.Response_ACK, Id: id}, nil
}

func (s *server) getHead(stream pb.PDPControl_UploadServer) (int32, *streamReader, error) {
	chunk, err := stream.Recv()
	if err == io.EOF {
		return 0, nil, stream.SendAndClose(controlFail(newEmptyUploadError()))
	}

	if err != nil {
		return 0, nil, err
	}

	return chunk.Id, newStreamReader(chunk.Id, chunk.Data, stream), nil
}

func (s *server) Upload(stream pb.PDPControl_UploadServer) error {
	log.Info("Got new data stream")

	id, r, err := s.getHead(stream)
	if r == nil {
		return err
	}

	req, ok := s.q.pop(id)
	if !ok {
		log.WithField("id", id).Error("no such request")
		err := r.skip()
		if err != nil {
			return err
		}

		return stream.SendAndClose(controlFail(newUnknownUploadError(id)))
	}

	if req.fromTag == nil {
		if req.policy {
			err = s.uploadPolicy(id, r, req, stream)
		} else {
			err = s.uploadContent(id, r, req, stream)
		}
	} else {
		if req.policy {
			err = s.uploadPolicyUpdate(id, r, req, stream)
		} else {
			err = s.uploadContentUpdate(id, r, req, stream)
		}
	}

	debug.FreeOSMemory()
	s.checkMemory(&conf.mem)

	return err
}

func (s *server) Apply(ctx context.Context, in *pb.Update) (*pb.Response, error) {
	log.Info("Got apply command")

	req, ok := s.q.pop(in.Id)
	if !ok {
		log.WithField("id", in.Id).Error("no such request")
		return controlFail(newUnknownUploadedRequestError(in.Id)), nil
	}

	var (
		res *pb.Response
		err error
	)
	if req.policy {
		res, err = s.applyPolicy(in.Id, req)
	} else {
		res, err = s.applyContent(in.Id, req)
	}

	debug.FreeOSMemory()
	s.checkMemory(&conf.mem)

	return res, err
}

func (s *server) NotifyReady(ctx context.Context, m *pb.Empty) (*pb.Response, error) {
	log.Info("Got notified about readiness")
	s.start<-true

	return &pb.Response{Status: pb.Response_ACK}, nil
}
