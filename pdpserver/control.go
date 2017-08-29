package main

import (
	"io"
	"io/ioutil"

	log "github.com/Sirupsen/logrus"
	"github.com/google/uuid"
	"golang.org/x/net/context"

	"github.com/infobloxopen/themis/pdp"
	pb "github.com/infobloxopen/themis/pdp-control"
	"github.com/infobloxopen/themis/pdp/jcon"
	"github.com/infobloxopen/themis/pdp/yast"
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

func (s *Server) Request(ctx context.Context, in *pb.Item) (*pb.Response, error) {
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

	switch in.Type {
	case pb.Item_POLICIES:
		if fromTag != nil {
			s.RLock()
			p := s.p
			s.RUnlock()

			err := p.CheckTag(fromTag)
			if err != nil {
				return controlFail(newTagCheckError(err)), nil
			}
		}

		id, err := s.queue.Push(NewPolicyItem(fromTag, toTag))
		if err != nil {
			return controlFail(err), nil
		}

		return &pb.Response{
			Status: pb.Response_ACK,
			Id:     id}, nil

	case pb.Item_CONTENT:
		if fromTag != nil {
			s.RLock()
			c := s.c
			s.RUnlock()

			_, err := c.GetLocalContent(in.Id, fromTag)
			if err != nil {
				return controlFail(newTagCheckError(err)), nil
			}
		}

		id, err := s.queue.Push(NewContentItem(in.Id, fromTag, toTag))
		if err != nil {
			return controlFail(err), nil
		}

		return &pb.Response{
			Status: pb.Response_ACK,
			Id:     id}, nil
	}

	return controlFail(newUnknownUploadRequestError(in.Type)), nil
}

type streamReader struct {
	id     int32
	stream pb.PDPControl_UploadServer
	chunk  []byte
	offset int
	eof    bool
}

func newStreamReader(id int32, head string, stream pb.PDPControl_UploadServer) *streamReader {
	return &streamReader{
		id:     id,
		stream: stream,
		chunk:  []byte(head)}
}

func (r *streamReader) skip() error {
	if r.eof {
		return nil
	}

	for {
		_, err := r.stream.Recv()
		if err == io.EOF {
			r.eof = true
			break
		}

		if err != nil {
			log.WithFields(log.Fields{
				"id":    r.id,
				"error": err}).Error("failed to read data stream")
			return err
		}
	}

	return nil
}

func (r *streamReader) Read(p []byte) (n int, err error) {
	if r.eof {
		return 0, io.EOF
	}

	if len(p) <= 0 {
		return 0, nil
	}

	offset := 0
	req := len(p) - offset
	rem := len(r.chunk) - r.offset
	for req > rem {
		for i := 0; i < rem; i++ {
			p[offset+i] = r.chunk[r.offset+i]
		}

		offset += rem
		req -= rem
		r.offset = 0

		chunk, err := r.stream.Recv()
		if err == io.EOF {
			r.eof = true
			return offset, io.EOF
		}

		if err != nil {
			log.WithFields(log.Fields{
				"id":    r.id,
				"error": err}).Error("failed to read data stream")
			return offset, err
		}

		r.chunk = []byte(chunk.Data)

		rem = len(r.chunk)
	}

	for i := 0; i < req; i++ {
		p[offset+i] = r.chunk[r.offset+i]
	}

	r.offset += req

	return offset + req, nil
}

func (s *Server) getHead(stream pb.PDPControl_UploadServer) (int32, *streamReader, error) {
	chunk, err := stream.Recv()
	if err == io.EOF {
		return 0, nil, stream.SendAndClose(controlFail(newEmptyUploadError()))
	}

	if err != nil {
		return 0, nil, err
	}

	return chunk.Id, newStreamReader(chunk.Id, chunk.Data, stream), nil
}

func (s *Server) dispatchUpload(id int32, r *streamReader, stream pb.PDPControl_UploadServer) (*Item, error) {
	req, ok := s.queue.Pop(id)
	if !ok {
		log.WithField("id", id).Error("no such request")
		err := r.skip()
		if err != nil {
			return nil, err
		}

		return nil, stream.SendAndClose(controlFail(newUnknownUploadError(id)))
	}

	return req, nil
}

func (s *Server) uploadPolicy(id int32, r *streamReader, req *Item, stream pb.PDPControl_UploadServer) error {
	b, err := ioutil.ReadAll(r)
	if err != nil {
		return err
	}

	p, err := yast.Unmarshal(b, req.toTag)
	if err != nil {
		return stream.SendAndClose(controlFail(newPolicyUploadParseError(id, err)))
	}

	req.p = p
	nid, err := s.queue.Push(req)
	if err != nil {
		return stream.SendAndClose(controlFail(newPolicyUploadStoreError(id, err)))
	}

	return stream.SendAndClose(&pb.Response{Status: pb.Response_ACK, Id: nid})
}

func (s *Server) uploadPolicyUpdate(id int32, r *streamReader, req *Item, stream pb.PDPControl_UploadServer) error {
	s.RLock()
	if s.p == nil {
		s.RUnlock()
		r.skip()
		return stream.SendAndClose(controlFail(newMissingPolicyStorageError()))
	}

	t, err := s.p.NewTransaction(req.fromTag)
	if err != nil {
		s.RUnlock()
		r.skip()
		return stream.SendAndClose(controlFail(newPolicyTransactionCreationError(id, req, err)))
	}
	s.RUnlock()

	b, err := ioutil.ReadAll(r)
	if err != nil {
		return err
	}

	u, err := yast.UnmarshalUpdate(b, t.Attributes(), *req.fromTag, *req.toTag)
	if err != nil {
		return stream.SendAndClose(controlFail(newPolicyUpdateParseError(id, req, err)))
	}

	log.WithField("update", u).Debug("Policy update")

	err = t.Apply(u)
	if err != nil {
		return stream.SendAndClose(controlFail(newPolicyUpdateApplicationError(id, req, err)))
	}

	req.pt = t
	nid, err := s.queue.Push(req)
	if err != nil {
		return stream.SendAndClose(controlFail(newPolicyUpdateUploadStoreError(id, req, err)))
	}

	return stream.SendAndClose(&pb.Response{Status: pb.Response_ACK, Id: nid})
}

func (s *Server) applyPolicy(id int32, req *Item) (*pb.Response, error) {
	if req.p != nil {
		s.Lock()
		s.p = req.p
		s.Unlock()

		if req.toTag == nil {
			log.WithField("id", id).Info("New policy has been applied")
		} else {
			log.WithFields(log.Fields{
				"id":  id,
				"tag": req.toTag.String()}).Info("New policy has been applied")
		}

		return &pb.Response{Status: pb.Response_ACK, Id: id}, nil
	}

	if req.pt != nil {
		p, err := req.pt.Commit()
		if err != nil {
			return controlFail(newPolicyTransactionCommitError(id, req, err)), nil
		}

		s.Lock()
		s.p = p
		s.Unlock()

		log.WithFields(log.Fields{
			"id":       id,
			"prev-tag": req.fromTag,
			"curr-tag": req.toTag}).Info("Policy update has been applied")

		return &pb.Response{Status: pb.Response_ACK, Id: id}, nil
	}

	return controlFail(newMissingPolicyDataApplyError(id)), nil
}

func (s *Server) uploadContent(id int32, r *streamReader, req *Item, stream pb.PDPControl_UploadServer) error {
	c, err := jcon.Unmarshal(r, req.toTag)
	if err != nil {
		r.skip()
		return stream.SendAndClose(controlFail(newContentUploadParseError(id, err)))
	}

	req.c = c
	nid, err := s.queue.Push(req)
	if err != nil {
		return stream.SendAndClose(controlFail(newContentUploadStoreError(id, err)))
	}

	return stream.SendAndClose(&pb.Response{Status: pb.Response_ACK, Id: nid})
}

func (s *Server) uploadContentUpdate(id int32, r *streamReader, req *Item, stream pb.PDPControl_UploadServer) error {
	s.RLock()
	t, err := s.c.NewTransaction(req.id, req.fromTag)
	if err != nil {
		s.RUnlock()
		r.skip()
		return stream.SendAndClose(controlFail(newContentTransactionCreationError(id, req, err)))
	}
	s.RUnlock()

	u, err := jcon.UnmarshalUpdate(r, req.id, *req.fromTag, *req.toTag)
	if err != nil {
		r.skip()
		return stream.SendAndClose(controlFail(newContentUpdateParseError(id, req, err)))
	}

	log.WithField("update", u).Debug("Content update")

	err = t.Apply(u)
	if err != nil {
		return stream.SendAndClose(controlFail(newContentUpdateApplicationError(id, req, err)))
	}

	req.ct = t
	nid, err := s.queue.Push(req)
	if err != nil {
		return stream.SendAndClose(controlFail(newContentUpdateUploadStoreError(id, req, err)))
	}

	return stream.SendAndClose(&pb.Response{Status: pb.Response_ACK, Id: nid})
}

func (s *Server) applyContent(id int32, req *Item) (*pb.Response, error) {
	if req.c != nil {
		s.Lock()
		s.c = s.c.Add(req.c)
		s.Unlock()

		if req.toTag == nil {
			log.WithField("id", id).Info("New content has been applied")
		} else {
			log.WithFields(log.Fields{
				"id":  id,
				"tag": req.toTag.String()}).Info("New content has been applied")
		}

		return &pb.Response{Status: pb.Response_ACK, Id: id}, nil
	}

	if req.ct != nil {
		s.Lock()
		c, err := req.ct.Commit(s.c)
		if err != nil {
			s.Unlock()

			return controlFail(newContentTransactionCommitError(id, req, err)), nil
		}

		s.c = c
		s.Unlock()

		log.WithFields(log.Fields{
			"id":       id,
			"cid":      req.id,
			"prev-tag": req.fromTag,
			"curr-tag": req.toTag}).Info("Content update has been applied")

		return &pb.Response{Status: pb.Response_ACK, Id: id}, nil
	}

	return controlFail(newMissingContentDataApplyError(id, req.id)), nil
}

func (s *Server) Upload(stream pb.PDPControl_UploadServer) error {
	log.Info("Got new data stream")

	id, r, err := s.getHead(stream)
	if r == nil {
		return err
	}

	req, err := s.dispatchUpload(id, r, stream)
	if req == nil {
		return err
	}

	if req.fromTag == nil {
		if req.policy {
			return s.uploadPolicy(id, r, req, stream)
		}

		return s.uploadContent(id, r, req, stream)
	}

	if req.policy {
		return s.uploadPolicyUpdate(id, r, req, stream)
	}

	return s.uploadContentUpdate(id, r, req, stream)
}

func (s *Server) Apply(ctx context.Context, in *pb.Update) (*pb.Response, error) {
	log.Info("Got apply command")

	req, ok := s.queue.Pop(in.Id)
	if !ok {
		log.WithField("id", in.Id).Error("no such request")
		return controlFail(newUnknownUploadedRequestError(in.Id)), nil
	}

	if req.policy {
		return s.applyPolicy(in.Id, req)
	}

	return s.applyContent(in.Id, req)
}
