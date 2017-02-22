package main

import (
	"fmt"
	"golang.org/x/net/context"

	pb "github.com/infobloxopen/policy-box/pdp-control"
)

func (s *Server) Upload(stream pb.PDPControl_UploadServer) error {
	return stream.SendAndClose(&pb.Response{pb.Response_ERROR, -1, "Not implemented yet"})
}

func (s *Server) Parse(server_ctx context.Context, in *pb.Item) (*pb.Response, error) {
	return nil, fmt.Errorf("Not implemented yet")
}

func (s *Server) Apply(server_ctx context.Context, in *pb.Update) (*pb.Response, error) {
	return nil, fmt.Errorf("Not implemented yet")
}
