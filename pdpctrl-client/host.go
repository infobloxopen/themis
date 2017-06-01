package pdpctrl_client

//go:generate bash -c "mkdir -p $GOPATH/src/github.com/infobloxopen/themis/pdp-control && protoc -I $GOPATH/src/github.com/infobloxopen/themis/proto/ $GOPATH/src/github.com/infobloxopen/themis/proto/control.proto --go_out=plugins=grpc:$GOPATH/src/github.com/infobloxopen/themis/pdp-control && ls $GOPATH/src/github.com/infobloxopen/themis/pdp-control"

import (
	"errors"
	"time"

	"golang.org/x/net/context"
	"google.golang.org/grpc"

	pb "github.com/infobloxopen/themis/pdp-control"
)

type Logger interface {
	Infof(format string, a ...interface{})
	Errorf(format string, a ...interface{})
}

type Host struct {
	address string
	chunk   int

	conn   *grpc.ClientConn
	client pb.PDPControlClient

	log Logger
}

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

func NewHost(addr string, chunk int, log Logger) *Host {
	return &Host{
		address: addr,
		chunk:   chunk,
		log:     log,
	}
}

func (h *Host) Connect(timeout time.Duration) error {
	h.log.Infof("Connecting to %s...", h.address)
	conn, err := grpc.Dial(h.address, grpc.WithInsecure(), grpc.WithBlock(), grpc.WithTimeout(timeout))
	if err != nil {
		h.log.Errorf("Can't connect to %s: %v. Skipping...", h.address, err)
		return err
	}

	h.conn = conn
	h.client = pb.NewPDPControlClient(h.conn)

	return nil
}

func (h *Host) Close() {
	if h.conn != nil {
		h.conn.Close()
	}

	h.client = nil
}

func (h *Host) Upload(bucket *DataBucket) error {
	h.log.Infof("Uploading PDP data to host %s...", h.address)

	iids, err := h.uploadIncludes(bucket)
	if err != nil {
		return err
	}

	if err := h.uploadPolicies(bucket, iids); err != nil {
		return err
	}

	return nil
}

func (h *Host) Apply(bucketId int32) error {
	h.log.Infof("Applying policies to host %s...", h.address)

	r, err := h.client.Apply(context.Background(), &pb.Update{bucketId})
	if err != nil {
		return err
	}

	if r.Status != pb.Response_ACK {
		return errors.New(r.Details)
	}

	return nil
}

func (h *Host) uploadData(data []byte) (int32, error) {
	stream, err := h.client.Upload(context.Background())
	if err != nil {
		return 0, nil
	}

	for offset := 0; offset < len(data); offset += h.chunk {
		var c []byte
		if len(data) < offset+h.chunk {
			c = data[offset:]
		} else {
			c = data[offset : offset+h.chunk]
		}

		if err := stream.Send(&pb.Chunk{string(c)}); err != nil {
			h.log.Errorf("Can't upload %d bytes starting %d: %v", len(c), offset, err)
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

func (h *Host) uploadIncludes(bucket *DataBucket) ([]int32, error) {
	rids := []int32{}

	for id, data := range bucket.Includes {
		h.log.Infof("Uploading %q content to %s...", id, h.address)

		rid, err := h.uploadData(data)
		if err != nil {
			h.log.Errorf("Failed to upload %q content: %v", id, err)
			return nil, err
		}

		item := &pb.Item{
			Type:        pb.Item_CONTENT,
			FromVersion: bucket.FromVersion,
			ToVersion:   bucket.ToVersion,
			DataId:      rid,
			Id:          id,
		}

		r, err := h.client.Parse(context.Background(), item)
		if err != nil {
			h.log.Errorf("Failed to parse uploaded content %q: %v", id, err)
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

func (h *Host) uploadPolicies(bucket *DataBucket, iids []int32) error {
	h.log.Infof("Uploading policies to %s...", h.address)

	rid, err := h.uploadData(bucket.Policies)
	if err != nil {
		h.log.Errorf("Failed to upload policies: %v", err)
		return err
	}

	item := &pb.Item{
		Type:        pb.Item_POLICIES,
		FromVersion: bucket.FromVersion,
		ToVersion:   bucket.ToVersion,
		DataId:      rid,
		Includes:    iids,
	}

	r, err := h.client.Parse(context.Background(), item)
	if err != nil {
		h.log.Errorf("Can't parse policy: %v. Skipping host...", err)
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
