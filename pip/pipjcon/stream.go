package main

import (
	"io"

	log "github.com/sirupsen/logrus"

	pb "github.com/infobloxopen/themis/pdp-control"
)

type streamReader struct {
	stream pb.PDPControl_UploadServer
	id     int32
	chunk  []byte
	eof    bool
	size   int
}

func newStreamReader(stream pb.PDPControl_UploadServer, chunk *pb.Chunk) *streamReader {
	return &streamReader{
		stream: stream,
		id:     chunk.Id,
		chunk:  []byte(chunk.Data),
		size:   len(chunk.Data),
	}
}

func (r *streamReader) skip() {
	if r.eof {
		return
	}

	for {
		_, err := r.stream.Recv()
		if err == io.EOF {
			r.eof = true
			break
		}

		if err != nil {
			log.WithFields(log.Fields{
				"id":  r.id,
				"err": err,
			}).Error("failed to read data stream")

			break
		}
	}
}

func (r *streamReader) Read(p []byte) (int, error) {
	if r.eof {
		return 0, io.EOF
	}

	if len(p) <= 0 {
		return 0, nil
	}

	n := 0
	for len(p) > len(r.chunk) {
		m := copy(p, r.chunk)
		n += m
		p = p[m:]
		r.chunk = r.chunk[m:]

		chunk, err := r.stream.Recv()
		if err == io.EOF {
			r.eof = true
			return n, io.EOF
		}

		if err != nil {
			log.WithFields(log.Fields{
				"id":  r.id,
				"err": err,
			}).Error("failed to read data stream")

			return n, err
		}

		r.chunk = []byte(chunk.Data)
		r.size += len(r.chunk)
	}

	m := copy(p, r.chunk)
	r.chunk = r.chunk[m:]

	return n + m, nil
}
