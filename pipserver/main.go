package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"

	pb "github.com/infobloxopen/themis/pip-service"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

const (
	port = ":5356"
)

// server is used to implement helloworld.GreeterServer.
type server struct {
	CategoryMap map[string]string
}

func getCategoryMap(categoryFile string) (map[string]string, error) {
	categoryMap := make(map[string]string)

	f, err := os.Open(categoryFile)
	if err != nil {
		fmt.Printf("cannot open category file '%s': '%s'\n", categoryFile, err)
	}
	defer f.Close()

	dec := json.NewDecoder(f)

	if err = dec.Decode(&categoryMap); err != nil {
		fmt.Printf("cannot decode category file '%s': '%s\n", categoryFile, err)
	}

	return categoryMap, err
}

// SayHello implements helloworld.GreeterServer
func (s *server) GetAttribute(ctx context.Context, in *pb.Request) (*pb.Response, error) {
	inAttrs := in.GetAttributes()
	domainStr := inAttrs[0].GetValue()
	category, ok := s.CategoryMap[domainStr]
	if !ok {
		category = "unknown category"
	}

	respAttr := &pb.Attribute{Value: category}
	values := []*pb.Attribute{respAttr}
	return &pb.Response{Values: values}, nil
}

func main() {
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	categoryMap, err := getCategoryMap("category-map.json")
	if err != nil {
		fmt.Printf("getCategoryMap error: %s\n", err)
		os.Exit(1)
	}

	s := grpc.NewServer()
	pb.RegisterPIPServer(s, &server{CategoryMap: categoryMap})
	// Register reflection service on gRPC server.
	reflection.Register(s)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
