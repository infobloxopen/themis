package server

import (
	"fmt"
	"io"
	"sync/atomic"

	log "github.com/sirupsen/logrus"
	"golang.org/x/net/context"

	"github.com/infobloxopen/themis/pdp"
	pb "github.com/infobloxopen/themis/pip-service"
)

func newFailureMsg(err error) (*pb.Msg, error) {
	return &pb.Msg{
		Body: makeFailureResponse(err),
	}, nil
}

func newFailureMsgf(f string, args ...interface{}) (*pb.Msg, error) {
	return newFailureMsg(fmt.Errorf(f, args...))
}

func unmarshalContentMsg(in *pb.Msg) (string, string, pdp.AttributeValue, error) {
	a, err := pdp.UnmarshalRequestAssignments(in.Body)
	if err != nil {
		return "", "", pdp.UndefinedValue, err
	}

	var (
		content   string
		contentOk bool

		item   string
		itemOk bool

		key   pdp.AttributeValue
		keyOk bool
	)

	for _, a := range a {
		switch a.GetID() {
		case "content":
			s, err := a.GetString(nil)
			if err != nil {
				return "", "", pdp.UndefinedValue, fmt.Errorf("can't get content name: %s", err)
			}

			content = s
			contentOk = true

		case "item":
			s, err := a.GetString(nil)
			if err != nil {
				return "", "", pdp.UndefinedValue, fmt.Errorf("can't get content item name: %s", err)
			}

			item = s
			itemOk = true

		case "key":
			v, err := a.GetValue()
			if err != nil {
				return "", "", pdp.UndefinedValue, fmt.Errorf("can't get key: %s", err)
			}

			key = v
			keyOk = true
		}
	}

	if !contentOk {
		return "", "", pdp.UndefinedValue, fmt.Errorf("missing content name")
	}

	if !itemOk {
		return "", "", pdp.UndefinedValue, fmt.Errorf("missing content item name")
	}

	if !keyOk {
		return "", "", pdp.UndefinedValue, fmt.Errorf("missing key")
	}

	return content, item, key, nil
}

func (s *Server) mapByContent(c *pdp.LocalContentStorage, in *pb.Msg) (pdp.AttributeValue, error) {
	content, item, key, err := unmarshalContentMsg(in)
	if err != nil {
		return pdp.UndefinedValue, err
	}

	if s.opts.logger.Level >= log.DebugLevel {
		sKey, err := key.Serialize()
		if err != nil {
			s.opts.logger.WithFields(log.Fields{
				"content": content,
				"item":    item,
				"key-err": err,
			}).Debug("Content request")
		} else {
			s.opts.logger.WithFields(log.Fields{
				"content": content,
				"item":    item,
				"key":     sKey,
			}).Debug("Content request")
		}
	}

	ci, err := c.Get(content, item)
	if err != nil {
		return pdp.UndefinedValue, err
	}

	r, err := ci.Get([]pdp.Expression{key}, nil)
	if err != nil {
		return pdp.UndefinedValue, err
	}

	if s.opts.logger.Level >= log.DebugLevel {
		sR, err := r.Serialize()
		if err != nil {
			s.opts.logger.WithField("err", err).Debug("Content response")
		} else {
			s.opts.logger.WithField("response", sR).Debug("Content response")
		}
	}

	return r, nil
}

func (s *Server) Map(ctx context.Context, in *pb.Msg) (*pb.Msg, error) {
	s.RLock()
	c := s.c
	s.RUnlock()

	if c == nil {
		return newFailureMsgf("no local content")
	}

	r, err := s.mapByContent(c, in)
	if err != nil {
		return newFailureMsg(err)
	}

	b, err := pdp.Response{
		Obligations: []pdp.AttributeAssignment{
			pdp.MakeExpressionAssignment("r", r),
		},
	}.Marshal(nil)
	if err != nil {
		return newFailureMsgf("can't marshal response: %s", err)
	}

	return &pb.Msg{
		Body: b,
	}, nil
}

var pipStreamAutoIncrement uint64

func (s *Server) NewMappingStream(stream pb.PIP_NewMappingStreamServer) error {
	ctx := stream.Context()

	sID := atomic.AddUint64(&pipStreamAutoIncrement, 1)
	s.opts.logger.WithField("id", sID).Debug("Got new content stream")

	res := pdp.Response{
		Obligations: make([]pdp.AttributeAssignment, 1),
	}
	var out pb.Msg

	for {
		in, err := stream.Recv()
		if err == io.EOF {
			break
		}

		if err != nil {
			if err := ctx.Err(); err != nil && (err == context.Canceled || err == context.DeadlineExceeded) {
				break
			}

			s.opts.logger.WithFields(log.Fields{
				"id":  sID,
				"err": err,
			}).Error("Failed to read next request from content stream. Dropping stream...")

			return err
		}

		s.RLock()
		c := s.c
		s.RUnlock()

		if c == nil {
			out.Body = makeFailureResponse(fmt.Errorf("no local content"))
		} else {
			r, err := s.mapByContent(c, in)
			if err != nil {
				out.Body = makeFailureResponse(err)
			} else {
				res.Obligations[0] = pdp.MakeExpressionAssignment("r", r)
				b, err := res.Marshal(nil)
				if err != nil {
					out.Body = makeFailureResponse(fmt.Errorf("can't marshal response: %s", err))
				} else {
					out.Body = b
				}
			}
		}

		if err = stream.Send(&out); err != nil {
			s.opts.logger.WithFields(log.Fields{
				"id":  sID,
				"err": err,
			}).Error("Failed to send response to content stream. Dropping stream...")

			return err
		}
	}

	s.opts.logger.WithField("id", sID).Debug("Content stream depleted")
	return nil
}
