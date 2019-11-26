package main

import (
  "context"
  "log"
  "net"

  "google.golang.org/grpc"
  "google.golang.org/grpc/codes"
  "google.golang.org/grpc/status"

  "github.com/infobloxopen/themis/pdp"
  pb "github.com/infobloxopen/themis/pdp-service"
)

const (
  port = ":5555"
)

type Server struct {}

func (s *Server) Validate(ctx context.Context, in *pb.Msg) (*pb.Msg, error) {
  log.Printf("Receiving Validate")

  resp := &pdp.Response{}
  resp.Effect = pdp.EffectPermit
  resp.Status = nil
  resp.Obligations = []pdp.AttributeAssignment{}

  body, _ := resp.Marshal(nil)

  return &pb.Msg{Body: body}, nil
}

func (s *Server) NewValidationStream(stream pb.PDP_NewValidationStreamServer) error {
  log.Printf("Receiving NewValidationStream")
  return status.Errorf(codes.Unimplemented, "NewValidationStream is not supported")
}

func main(){
  log.Printf("Starting on port: %v", port)

  lis, err := net.Listen("tcp", port)
  if err != nil {
    log.Fatalf("failed to listen: %v", err)
  }

  s := grpc.NewServer()
  pb.RegisterPDPServer(s, &Server{})

  if err := s.Serve(lis); err != nil {
    log.Fatal("failed to serve: %v", err)
  }
}
