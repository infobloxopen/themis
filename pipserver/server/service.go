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

func (s *Server) Map(ctx context.Context, in *pb.Msg) (*pb.Msg, error) {
	b, err := pdp.MakeIndeterminateResponse(fmt.Errorf("not implemented"))
	if err != nil {
		panic(err)
	}

	return &pb.Msg{
		Body: b,
	}, nil
}

var streamAutoIncrement uint64

func (s *Server) NewMappingStream(stream pb.PIP_NewMappingStreamServer) error {
	ctx := stream.Context()

	sID := atomic.AddUint64(&streamAutoIncrement, 1)
	s.opts.logger.WithField("id", sID).Debug("Got new stream")

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
			}).Error("Failed to read next request from stream. Dropping stream...")

			return err
		}

		b, err := pdp.MakeIndeterminateResponse(fmt.Errorf("not implemented"))
		if err != nil {
			panic(err)
		}
		out.Body = b

		if err = stream.Send(&out); err != nil {
			s.opts.logger.WithFields(log.Fields{
				"id":  sID,
				"err": err,
			}).Error("Failed to send response. Dropping stream...")

			return err
		}
	}

	s.opts.logger.WithField("id", sID).Debug("Stream deleted")
	return nil
}
