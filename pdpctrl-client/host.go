package pdpctrl_client

//go:generate bash -c "mkdir -p $GOPATH/src/github.com/infobloxopen/themis/pdp-control && protoc -I $GOPATH/src/github.com/infobloxopen/themis/proto/ $GOPATH/src/github.com/infobloxopen/themis/proto/control.proto --go_out=plugins=grpc:$GOPATH/src/github.com/infobloxopen/themis/pdp-control && ls $GOPATH/src/github.com/infobloxopen/themis/pdp-control"

import (
	"errors"
	"fmt"
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

	ready  bool
	policy int32

	log Logger
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
		policy:  -1,
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

func (h *Host) Process(patch bool, version string, includes map[string][]byte, policy []byte) error {
	if err := h.upload(patch, version, includes, policy); err != nil {
		return err
	}

	if err := h.apply(); err != nil {
		return err
	}

	return nil
}

func (h *Host) uploadIncludes(patch bool, version string, includes map[string][]byte) ([]int32, error) {
	if h.client == nil {
		return nil, nil
	}

	IDs := []int32{}
	for ID, b := range includes {
		h.log.Infof("Uploading %q content to %s...", ID, h.address)

		stream, err := h.client.Upload(context.Background())
		if err != nil {
			h.log.Errorf("Can't start content uploading: %v. Skipping the host...", err)
			return nil, err
		}

		for offset := 0; offset < len(b); offset += h.chunk {
			var c []byte
			if len(b) < offset+h.chunk {
				c = b[offset:]
			} else {
				c = b[offset : offset+h.chunk]
			}

			if err := stream.Send(&pb.Chunk{string(c)}); err != nil {
				h.log.Errorf("Can't upload %d bytes starting %d: %v. Skipping the host...", len(c), offset, err)
				return nil, err
			}
		}

		r, err := stream.CloseAndRecv()
		if err != nil {
			h.log.Errorf("Can't finish content uploading: %v. Skipping the host...", err)
			return nil, err
		}

		if r.Status != pb.Response_ACK {
			h.log.Errorf("Error while loading content to the host: %s. Skipping the host...", r.Details)
			return nil, errors.New(r.Details)
		}

		var dtype pb.Item_DataType
		if patch {
			dtype = pb.Item_CONTENT_PATCH
		} else {
			dtype = pb.Item_CONTENT
		}

		item := &pb.Item{
			Type:    dtype,
			Version: version,
			DataId:  r.Id,
			Id:      ID,
		}

		r, err = h.client.Parse(context.Background(), item)
		if err != nil {
			h.log.Errorf("Can't parse uploaded content: %v. Skipping the host...", err)
			return nil, err
		}

		switch r.Status {
		case pb.Response_ACK:
			IDs = append(IDs, r.Id)
		case pb.Response_VERSION_ERROR:
			h.log.Errorf("Incorrect %s data version. Current version is %s", version, r.Details)
			return nil, &VersionError{version: r.Details}
		case pb.Response_ERROR:
			h.log.Errorf("Error while parsing uploaded content: %s. Skipping the host...", r.Details)
			return nil, errors.New(r.Details)
		default:
			return nil, fmt.Errorf("Unexpected %s response status", r.Details)
		}
	}

	return IDs, nil
}

func (h *Host) uploadPolicy(patch bool, version string, policy []byte, IDs []int32) error {
	if IDs == nil {
		return nil
	}

	h.log.Infof("Uploading policy to %s...", h.address)

	stream, err := h.client.Upload(context.Background())
	if err != nil {
		h.log.Errorf("Can't start policy uploading: %v. Skipping the host...", err)
		return err
	}

	for offset := 0; offset < len(policy); offset += h.chunk {
		var c []byte
		if len(policy) < offset+h.chunk {
			c = policy[offset:]
		} else {
			c = policy[offset : offset+h.chunk]
		}

		if err := stream.Send(&pb.Chunk{string(c)}); err != nil {
			h.log.Errorf("Can't upload %d bytes starting %d: %v. Skipping the host...", len(c), offset, err)
			return err
		}
	}

	r, err := stream.CloseAndRecv()
	if err != nil {
		h.log.Errorf("Can't finish policy uploading: %v. Skipping the host...", err)
		return err
	}

	if r.Status != pb.Response_ACK {
		h.log.Errorf("Error while loading policy to the host: %s. Skipping the host...", r.Details)
		return errors.New(r.Details)
	}

	var dtype pb.Item_DataType
	if patch {
		dtype = pb.Item_POLICIES_PATCH
	} else {
		dtype = pb.Item_POLICIES
	}

	item := &pb.Item{
		Type:     dtype,
		Version:  version,
		DataId:   r.Id,
		Includes: IDs,
	}

	r, err = h.client.Parse(context.Background(), item)
	if err != nil {
		h.log.Errorf("Can't parse policy: %v. Skipping host...", err)
		return err
	}

	if r.Status != pb.Response_ACK {
		h.log.Errorf("Error while parsing policy at the host %s. Skipping the host...", r.Details)
		return errors.New(r.Details)
	}

	h.policy = r.Id
	h.ready = true

	return nil
}

func (h *Host) upload(patch bool, version string, includes map[string][]byte, policy []byte) error {
	ids, err := h.uploadIncludes(patch, version, includes)
	if err != nil {
		return err
	}

	if err := h.uploadPolicy(patch, version, policy, ids); err != nil {
		return err
	}

	return nil
}

func (h *Host) apply() error {
	if !h.ready {
		return nil
	}

	h.log.Infof("Applying policy to host %s...", h.address)

	update := pb.Update{h.policy}
	r, err := h.client.Apply(context.Background(), &update)
	if err != nil {
		h.log.Errorf("Can't apply policy: %v", err)
		return err
	}

	if r.Status != pb.Response_ACK {
		h.log.Errorf("Error while applying policy to the host %s", r.Details)
		return errors.New(r.Details)
	}

	h.log.Infof("Policy has been applied.")

	return nil
}
