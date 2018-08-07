package server

import (
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"

	pb "github.com/infobloxopen/themis/pdp-control"
	"github.com/infobloxopen/themis/pdp/jcon"
)

func (s *Server) contentRequest(id string, fromTag, toTag *uuid.UUID) (int32, error) {
	if fromTag != nil {
		s.RLock()
		c := s.c
		s.RUnlock()

		_, err := c.GetLocalContent(id, fromTag)
		if err != nil {
			return 0, newTagCheckError(err)
		}
	}

	return s.q.push(newContentItem(id, fromTag, toTag))
}

func (s *Server) uploadContent(id int32, r *streamReader, req *item, stream pb.PDPControl_UploadServer) error {
	c, err := jcon.Unmarshal(r, req.toTag)
	if err != nil {
		r.skip()
		return stream.SendAndClose(controlFail(newContentUploadParseError(id, err)))
	}

	req.c = c

	s.RLock()
	req.scc = s.pipShardClients
	s.RUnlock()

	nid, err := s.q.push(req)
	if err != nil {
		return stream.SendAndClose(controlFail(newContentUploadStoreError(id, err)))
	}

	return stream.SendAndClose(&pb.Response{Status: pb.Response_ACK, Id: nid})
}

func (s *Server) uploadContentUpdate(id int32, r *streamReader, req *item, stream pb.PDPControl_UploadServer) error {
	s.RLock()
	t, err := s.c.NewTransaction(req.id, req.fromTag)
	if err != nil {
		s.RUnlock()
		r.skip()
		return stream.SendAndClose(controlFail(newContentTransactionCreationError(id, req, err)))
	}
	scc := s.pipShardClients
	s.RUnlock()

	u, err := jcon.UnmarshalUpdate(r, req.id, *req.fromTag, *req.toTag, t.Symbols())
	if err != nil {
		r.skip()
		return stream.SendAndClose(controlFail(newContentUpdateParseError(id, req, err)))
	}

	s.opts.logger.WithField("update", u).Debug("Content update")

	err = t.Apply(u)
	if err != nil {
		return stream.SendAndClose(controlFail(newContentUpdateApplicationError(id, req, err)))
	}

	req.ct = t
	req.scc = scc
	nid, err := s.q.push(req)
	if err != nil {
		return stream.SendAndClose(controlFail(newContentUpdateUploadStoreError(id, req, err)))
	}

	return stream.SendAndClose(&pb.Response{Status: pb.Response_ACK, Id: nid})
}

func (s *Server) applyContent(id int32, req *item) (*pb.Response, error) {
	if req.c != nil {
		scc, newC, oldC := req.scc.update(req.c.GetShards().Map(), s.opts.shardingStreams)
		for _, c := range newC {
			c.Connect("")
		}

		s.Lock()
		s.c = s.c.Add(req.c)
		s.pipShardClients = scc
		if s.p != nil {
			s.p.Event(s.newLocalContentRouter())
		}
		s.Unlock()

		for _, c := range oldC {
			c.Close()
		}

		if req.toTag == nil {
			s.opts.logger.WithField("id", id).Info("New content has been applied")
		} else {
			s.opts.logger.WithFields(log.Fields{
				"id":  id,
				"tag": req.toTag.String()}).Info("New content has been applied")
		}

		return &pb.Response{Status: pb.Response_ACK, Id: id}, nil
	}

	if req.ct != nil {
		c, err := req.ct.Commit(s.c)
		if err != nil {
			return controlFail(newContentTransactionCommitError(id, req, err)), nil
		}

		scc, newC, oldC := req.scc.update(req.c.GetShards().Map(), s.opts.shardingStreams)
		for _, c := range newC {
			c.Connect("")
		}

		s.Lock()
		s.c = c
		s.pipShardClients = scc
		if s.p != nil {
			s.p.Event(s.newLocalContentRouter())
		}
		s.Unlock()

		for _, c := range oldC {
			c.Close()
		}

		s.opts.logger.WithFields(log.Fields{
			"id":       id,
			"cid":      req.id,
			"prev-tag": req.fromTag,
			"curr-tag": req.toTag}).Info("Content update has been applied")

		return &pb.Response{Status: pb.Response_ACK, Id: id}, nil
	}

	return controlFail(newMissingContentDataApplyError(id, req.id)), nil
}
