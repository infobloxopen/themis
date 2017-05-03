package main

//go:generate bash -c "mkdir -p $GOPATH/src/github.com/infobloxopen/themis/pdp-service && protoc -I $GOPATH/src/github.com/infobloxopen/themis/proto/ $GOPATH/src/github.com/infobloxopen/themis/proto/service.proto --go_out=plugins=grpc:$GOPATH/src/github.com/infobloxopen/themis/pdp-service && ls $GOPATH/src/github.com/infobloxopen/themis/pdp-service"

import (
	"fmt"
	"os"
	"time"

	"golang.org/x/net/context"
	"google.golang.org/grpc"

	pb "github.com/infobloxopen/themis/pdp-service"
)

var (
	ErrorConnected    = fmt.Errorf("Connection has been already established")
	ErrorNotConnected = fmt.Errorf("No connection")
)

type Client struct {
	Address string

	conn   *grpc.ClientConn
	client *pb.PDPClient
}

func NewClient() *Client {
	return &Client{}
}

func (c *Client) Connect(addr string, timeout time.Duration) error {
	if c.conn != nil {
		return ErrorConnected
	}

	conn, err := grpc.Dial(addr, grpc.WithInsecure(), grpc.WithBlock(), grpc.WithTimeout(timeout))
	if err != nil {
		return err
	}

	c.Address = addr
	c.conn = conn

	client := pb.NewPDPClient(c.conn)
	c.client = &client
	return nil
}

func (c *Client) Close() {
	if c.conn != nil {
		c.conn.Close()
		c.conn = nil
	}

	c.client = nil
}

func (c *Client) Send(reqs *Requests, count int, name string) error {
	f := os.Stdout
	var err error
	if len(name) > 0 {
		f, err = os.Create(name)
		if err != nil {
			return err
		}
		defer f.Close()
	}

	for req := range reqs.Parse(count) {
		if req.Error != nil {
			return fmt.Errorf("don't understand request %d: %s", req.Position, req.Error)
		}

		res, err := (*c.client).Validate(context.Background(), req.Request)
		if err != nil {
			return fmt.Errorf("can't send request %d (%d): %s", req.Index, req.Position, err)
		}

		err = DumpResponse(res, f)
		if err != nil {
			return fmt.Errorf("can't dump response for reqiest %d (%d): %s", req.Index, req.Position, err)
		}
	}

	return nil
}
