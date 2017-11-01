package main

import (
	"io"
	"sync/atomic"

	log "github.com/Sirupsen/logrus"
	"golang.org/x/net/context"

	pb "github.com/infobloxopen/themis/pdp-service"
)

var streamAutoIncrement uint64

func (s *server) NewValidationStream(stream pb.PDP_NewValidationStreamServer) error {
	sID := atomic.AddUint64(&streamAutoIncrement, 1)
	log.WithField("id", sID).Info("Got new stream")

	for {
		in, err := stream.Recv()
		if err == io.EOF {
			break
		}

		if err != nil {
			log.WithFields(log.Fields{
				"id":  sID,
				"err": err,
			}).Error("Failed to read next request from stream. Dropping stream...")

			return err
		}

		out, err := s.Validate(context.Background(), in)
		if err != nil {
			log.WithFields(log.Fields{
				"id":  sID,
				"err": err,
			}).Panic("Failed to validate request")
		}

		err = stream.Send(out)
		if err != nil {
			log.WithFields(log.Fields{
				"id":  sID,
				"err": err,
			}).Error("Failed to send response. Dropping stream...")

			return err
		}
	}

	log.WithField("id", sID).Info("Stream depleted")
	return nil
}
