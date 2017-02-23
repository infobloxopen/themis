package main

import (
	"encoding/json"
	"fmt"
	"io"
	"golang.org/x/net/context"
	log "github.com/Sirupsen/logrus"

	pb "github.com/infobloxopen/policy-box/pdp-control"

	"github.com/infobloxopen/policy-box/pdp"
)

func collect(stream pb.PDPControl_UploadServer) ([]byte, error) {
	data := make([]byte, 0)

	for {
		chunk, err := stream.Recv()
		if err == io.EOF {
			break
		}

		if err != nil {
			return nil, err
		}

		data = append(data, []byte(chunk.Data)...)
	}

	return data, nil
}

func controlAck(id int32) *pb.Response {
	return &pb.Response{pb.Response_ACK, id, ""}
}

func controlFail(format string, args ...interface{}) *pb.Response {
	return &pb.Response{pb.Response_ERROR, -1, fmt.Sprintf(format, args...)}
}

func (s *Server) Upload(stream pb.PDPControl_UploadServer) error {
	log.Info("Got new data stream")

	data, err := collect(stream)
	if err != nil {
		return err
	}

	id, err := s.Updates.Put(data)
	if err != nil {
		log.WithField("error", err).Error("Error on enqueueing data")
		return stream.SendAndClose(controlFail("Can't accept new data while all previous haven't been processed"))
	}

	log.WithFields(log.Fields{
		"size": len(data),
		"id": id}).Info("Data enqueued")

	return stream.SendAndClose(controlAck(id))
}

func (s *Server) DispatchPolicies(in *pb.Item) (interface{}, *pb.Response) {
	data, err := s.Updates.Get(in.DataId)
	if err != nil {
		log.WithField("id", in.DataId).Error("Failed to get specified data")
		return nil, controlFail("%v", err)
	}

	ext, err := s.Updates.GetIncludes(in.Includes)
	if err != nil {
		log.WithField("id", in.DataId).Error("Failed to collect specified includes")
		return nil, controlFail("%v", err)
	}

	item, err := pdp.UnmarshalYAST(data, s.Path, ext)
	if err != nil {
		log.WithFields(log.Fields{
			"id":   in.DataId,
			"type": pb.Item_DataType_name[int32(in.Type)],
		}).Error("Failed to parse the uploaded data as the desired type")
		return nil, controlFail("%v", err)
	}

	return item, nil
}

func (s *Server) DispatchContent(in *pb.Item) (interface{}, *pb.Response) {
	data, err := s.Updates.Get(in.DataId)
	if err != nil {
		log.WithField("id", in.DataId).Error("Failed to get specified data")
		return nil, controlFail("%v", err)
	}

	var item interface{}
	err = json.Unmarshal(data, &item)
	if err != nil {
		log.WithFields(log.Fields{
			"id":   in.DataId,
			"type": pb.Item_DataType_name[int32(in.Type)],
		}).Error("Failed to parse the uploaded data as the desired type")
		return nil, controlFail("%v", err)
	}

	return Content{in.Id, item}, nil
}

func (s *Server) DispatchUpdate(in *pb.Item) (interface{}, *pb.Response) {
	switch in.Type {
	case pb.Item_POLICIES:
		return s.DispatchPolicies(in)

	case pb.Item_CONTENT:
		return s.DispatchContent(in)
	}

	log.WithField("type", in.Type).Error("Unexpected policies or content")
	return nil, controlFail("Unknown upload type %d", in.Type)
}

func (s *Server) Parse(server_ctx context.Context, in *pb.Item) (*pb.Response, error) {
	log.Info("Parsing data")

	item, response := s.DispatchUpdate(in)
	if response != nil {
		return response, nil
	}

	err := s.Updates.Replace(in.DataId, item)
	if err != nil {
		log.Error("Failed to return parsed data to queue")
		return controlFail("%v", err), nil
	}

	log.WithFields(log.Fields{
		"id":   in.DataId,
		"type": pb.Item_DataType_name[int32(in.Type)],
	}).Info("Parsed the uploaded data as the type")

	count := s.Updates.PopIncludes(in.Includes)
	if count > 0 {
		log.WithField("count", count).Info("Deleted content items")
	}

	return controlAck(in.DataId), nil
}

func (s *Server) Apply(server_ctx context.Context, in *pb.Update) (*pb.Response, error) {
	s.Updates.Lock.Lock()
	defer s.Updates.Lock.Unlock()

	p, err := s.Updates.rawGetPolicies(in.Id)
	if err != nil {
		log.WithField("id", in.Id).Error("Can't get policies with specified id")
		return controlFail("%v", err), nil
	}

	s.Lock.Lock()
	s.Policy = p
	s.Lock.Unlock()
	log.WithField("id", in.Id).Info("Policies with the id has been applied")

	delete(s.Updates.Items, in.Id)
	if s.Updates.rawResetAutoIncrement() {
		log.Info("Autoincrement has been reseted")
	}

	return controlAck(in.Id), nil
}
