package pdpctrl_client

//go:generate bash -c "mkdir -p $GOPATH/src/github.com/infobloxopen/themis/pdp-control && protoc -I $GOPATH/src/github.com/infobloxopen/themis/proto/ $GOPATH/src/github.com/infobloxopen/themis/proto/control.proto --go_out=plugins=grpc:$GOPATH/src/github.com/infobloxopen/themis/pdp-control && ls $GOPATH/src/github.com/infobloxopen/themis/pdp-control"

import (
	"errors"
	"time"

	"golang.org/x/net/context"
	"google.golang.org/grpc"

	pb "github.com/infobloxopen/themis/pdp-control"
)

type DataBucket struct {
	ID int32

	FromVersion, ToVersion string

	Policies []byte
	Includes map[string][]byte
}

type VersionError struct {
	version string
}

func (e *VersionError) Error() string {
	return e.version
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

func (c *Client) Upload(bucket *DataBucket) error {
	iids, err := c.uploadIncludes(bucket)
	if err != nil {
		return err
	}

	if err := c.uploadPolicies(bucket, iids); err != nil {
		return err
	}

	return nil
}

func (c *Client) Apply(bucketId int32) error {
	r, err := c.client.Apply(context.Background(), &pb.Update{bucketId})
	if err != nil {
		return err
	}

	if r.Status != pb.Response_ACK {
		return errors.New(r.Details)
	}

	return nil
}

func (c *Client) uploadData(data []byte) (int32, error) {
	stream, err := c.client.Upload(context.Background())
	if err != nil {
		return 0, nil
	}

	for offset := 0; offset < len(data); offset += c.chunkSize {
		var chunk []byte
		if len(data) < offset+c.chunkSize {
			chunk = data[offset:]
		} else {
			chunk = data[offset : offset+c.chunkSize]
		}

		if err := stream.Send(&pb.Chunk{string(chunk)}); err != nil {
			return 0, err
		}
	}

	r, err := stream.CloseAndRecv()
	if err != nil {
		return 0, err
	}

	if r.Status != pb.Response_ACK {
		return 0, errors.New(r.Details)
	}

	return r.Id, nil
}

func (c *Client) uploadIncludes(bucket *DataBucket) ([]int32, error) {
	rids := []int32{}

	for id, data := range bucket.Includes {
		rid, err := c.uploadData(data)
		if err != nil {
			return nil, err
		}

		item := &pb.Item{
			Type:        pb.Item_CONTENT,
			FromVersion: bucket.FromVersion,
			ToVersion:   bucket.ToVersion,
			DataId:      rid,
			Id:          id,
		}

		r, err := c.client.Parse(context.Background(), item)
		if err != nil {
			return nil, err
		}

		if r.Status == pb.Response_ACK {
			rids = append(rids, r.Id)
		} else if r.Status == pb.Response_VERSION_ERROR {
			return nil, &VersionError{version: r.Details}
		} else {
			return nil, errors.New(r.Details)
		}
	}

	return rids, nil
}

func (c *Client) uploadPolicies(bucket *DataBucket, iids []int32) error {
	rid, err := c.uploadData(bucket.Policies)
	if err != nil {
		return err
	}

	item := &pb.Item{
		Type:        pb.Item_POLICIES,
		FromVersion: bucket.FromVersion,
		ToVersion:   bucket.ToVersion,
		DataId:      rid,
		Includes:    iids,
	}

	r, err := c.client.Parse(context.Background(), item)
	if err != nil {
		return err
	}

	if r.Status == pb.Response_VERSION_ERROR {
		return &VersionError{version: r.Details}
	} else if r.Status != pb.Response_ACK {
		return errors.New(r.Details)
	}

	bucket.ID = r.Id

	return nil
}
