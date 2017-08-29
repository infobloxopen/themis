package main

import (
	"encoding/json"
	log "github.com/Sirupsen/logrus"
	"net"
	"os"

	pb "github.com/infobloxopen/themis/pip-service"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

// server is used to implement helloworld.GreeterServer.
type server struct {
	CategoryMap map[string]string
}

const supportedQueryType = "domain-category"


func getCategoryMap(categoryFile string) (map[string]string, error) {
	categoryMap := make(map[string]string)

	f, err := os.Open(categoryFile)
	if err != nil {
		log.Fatalf("cannot open category file '%s': '%s'\n", categoryFile, err)
	}
	defer f.Close()

	dec := json.NewDecoder(f)

	if err = dec.Decode(&categoryMap); err != nil {
		log.Fatalf("cannot decode category file '%s': '%s\n", categoryFile, err)
	}

	return categoryMap, err
}

// SayHello implements helloworld.GreeterServer
func (s *server) GetAttribute(ctx context.Context, in *pb.Request) (*pb.Response, error) {
	responseStatus := pb.Response_OK

	queryType := in.GetQueryType()
	if queryType != supportedQueryType {
		log.Errorf("Query type '%s' is not supported", queryType)
		responseStatus = pb.Response_SERVICEERROR
	}
	inAttrs := in.GetAttributes()
	domainStr := inAttrs[0].GetValue()
	category, ok := s.CategoryMap[domainStr]
	if !ok {
		category = "unknown category"
		responseStatus = pb.Response_NOTFOUND
	}

	respAttr := &pb.Attribute{Value: category}
	values := []*pb.Attribute{respAttr}
	return &pb.Response{Status: responseStatus, Values: values}, nil
}

func main() {
	pipServiceName := "mcafee-ts"
	conn, err := pb.GetPIPConnection(pipServiceName)
	if err != nil {
		log.Fatalf("Cannot get PIP connection for PIP service '%s': %s", pipServiceName, err)
	}

	lis, err := net.Listen("tcp", conn)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	categoryMap, err := getCategoryMap("category-map.json")
	if err != nil {
		log.Fatalf("getCategoryMap error: %s\n", err)
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
