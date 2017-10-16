package main

import (
	"io"
	"sync/atomic"

	log "github.com/Sirupsen/logrus"

	pb "github.com/infobloxopen/themis/pdp-service"
)

var streamAutoIncrement uint64

func (s *server) NewValidationStream(stream pb.PDP_NewValidationStreamServer) error {
	streamID := atomic.AddUint64(&streamAutoIncrement, 1)
	log.WithField("id", streamID).Info("Got new stream")

	for {
		in, err := stream.Recv()
		if err == io.EOF {
			break
		}

		if err != nil {
			log.WithFields(log.Fields{
				"id":  streamID,
				"err": err,
			}).Error("Failed to read next request from stream. Closing stream...")

			return err
		}

		out, err := s.Validate(stream.Context(), in)
		if err != nil {
			log.WithFields(log.Fields{
				"id":  streamID,
				"err": err,
			}).Panic("Failed to validate request")
		}

		err = stream.Send(out)
		if err != nil {
			log.WithFields(log.Fields{
				"id":  streamID,
				"err": err,
			}).Error("Failed to send response. Closing stream...")

			return err
		}
	}

	log.WithField("id", streamID).Info("Stream depleted")
	return nil
}
