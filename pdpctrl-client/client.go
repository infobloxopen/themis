package pdpcc

//go:generate bash -c "mkdir -p $GOPATH/src/github.com/infobloxopen/themis/pdp-control && protoc -I $GOPATH/src/github.com/infobloxopen/themis/proto/ $GOPATH/src/github.com/infobloxopen/themis/proto/control.proto --go_out=plugins=grpc:$GOPATH/src/github.com/infobloxopen/themis/pdp-control && ls $GOPATH/src/github.com/infobloxopen/themis/pdp-control"

import (
	"errors"
	"fmt"
	"io"
	"time"

	"golang.org/x/net/context"
	"google.golang.org/grpc"

	pb "github.com/infobloxopen/themis/pdp-control"
)

type TagError struct {
	tag string
}

func (e *TagError) Error() string {
	return e.tag
}

type Client struct {
	address   string
	chunkSize int

	conn   *grpc.ClientConn
	client pb.PDPControlClient
}

func NewClient(addr string, chunkSize int) *Client {
	return &Client{
		address:   addr,
		chunkSize: chunkSize,
	}
}

func (c *Client) Connect(timeout time.Duration) error {
	conn, err := grpc.Dial(c.address, grpc.WithInsecure(), grpc.WithBlock(), grpc.WithTimeout(timeout))
	if err != nil {
		return err
	}

	c.conn = conn
	c.client = pb.NewPDPControlClient(c.conn)

	return nil
}

func (c *Client) Close() {
	if c.conn != nil {
		c.conn.Close()
	}

	c.client = nil
}

func (c *Client) RequestPoliciesUpload(fromTag, toTag string) (int32, error) {
	return c.request(&pb.Item{
		Type:    pb.Item_POLICIES,
		FromTag: fromTag,
		ToTag:   toTag})
}

func (c *Client) RequestContentUpload(id, fromTag, toTag string) (int32, error) {
	return c.request(&pb.Item{
		Type:    pb.Item_CONTENT,
		FromTag: fromTag,
		ToTag:   toTag,
		Id:      id})
}

func (c *Client) Upload(id int32, r io.Reader) (int32, error) {
	u, err := c.client.Upload(context.Background())
	if err != nil {
		return -1, err
	}

	p := make([]byte, c.chunkSize)
	for {
		n, err := r.Read(p)
		if n > 0 {
			chunk := &pb.Chunk{
				Id:   id,
				Data: string(p[:n])}
			if err := u.Send(chunk); err != nil {
				return -1, err
			}
		}

		if err == io.EOF {
			break
		}

		if err != nil {
			c.closeUpload(u)
			return -1, err
		}
	}

	return c.closeUpload(u)
}

func (c *Client) Apply(id int32) error {
	r, err := c.client.Apply(context.Background(), &pb.Update{Id: id})
	if err != nil {
		return err
	}

	if r.Status != pb.Response_ACK {
		return errors.New(r.Details)
	}

	return nil
}

func (c *Client) request(item *pb.Item) (int32, error) {
	r, err := c.client.Request(context.Background(), item)
	if err != nil {
		return -1, err
	}

	switch r.Status {
	case pb.Response_ACK:
		return r.Id, nil

	case pb.Response_ERROR:
		return -1, errors.New(r.Details)

	case pb.Response_TAG_ERROR:
		return -1, &TagError{tag: r.Details}
	}

	return -1, fmt.Errorf("Unknown response statue: %d", r.Status)
}

func (c *Client) closeUpload(u pb.PDPControl_UploadClient) (int32, error) {
	r, err := u.CloseAndRecv()
	if err != nil {
		return -1, err
	}

	if r.Status != pb.Response_ACK {
		return -1, errors.New(r.Details)
	}

	return r.Id, nil
}
