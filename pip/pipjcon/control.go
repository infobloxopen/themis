package main

import (
	"context"

	pb "github.com/infobloxopen/themis/pdp-control"
)

func (s *srv) Request(context.Context, *pb.Item) (*pb.Response, error) {
	return nil, errNotImplemented
}

func (s *srv) Upload(pb.PDPControl_UploadServer) error {
	return errNotImplemented
}

func (s *srv) Apply(context.Context, *pb.Update) (*pb.Response, error) {
	return nil, errNotImplemented
}

func (s *srv) NotifyReady(context.Context, *pb.Empty) (*pb.Response, error) {
	return nil, errNotImplemented
}
