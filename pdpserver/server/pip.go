package server

import (
	"fmt"
	"io"
	"sync/atomic"

	log "github.com/sirupsen/logrus"
	"golang.org/x/net/context"

	pb "github.com/infobloxopen/themis/pip-service"
)

func newFailureMsgf(f string, args ...interface{}) (*pb.Msg, error) {
	return &pb.Msg{
		Body: makeFailureResponse(fmt.Errorf(f, args...)),
	}, nil
}

func (s *Server) Map(ctx context.Context, in *pb.Msg) (*pb.Msg, error) {
	return newFailureMsgf("not implemented")
}

var pipStreamAutoIncrement uint64

func (s *Server) NewMappingStream(stream pb.PIP_NewMappingStreamServer) error {
	ctx := stream.Context()

	sID := atomic.AddUint64(&pipStreamAutoIncrement, 1)
	s.opts.logger.WithField("id", sID).Debug("Got new content stream")

	var out pb.Msg

	for {
		_, err := stream.Recv()
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

		out.Body = makeFailureResponse(fmt.Errorf("not implemented"))

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
