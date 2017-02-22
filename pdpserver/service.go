package main

import (
	"fmt"
	"golang.org/x/net/context"

	pb "github.com/infobloxopen/policy-box/pdp-service"
)

func (s *Server) Validate(server_ctx context.Context, in *pb.Request) (*pb.Response, error) {
	return nil, fmt.Errorf("Not implemented yet")
}
