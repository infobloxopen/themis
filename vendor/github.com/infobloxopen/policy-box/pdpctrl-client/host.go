package pdpctrl_client

//go:generate bash -c "mkdir -p $GOPATH/src/github.com/infobloxopen/themis/pdp-control && protoc -I $GOPATH/src/github.com/infobloxopen/themis/proto/ $GOPATH/src/github.com/infobloxopen/themis/proto/control.proto --go_out=plugins=grpc:$GOPATH/src/github.com/infobloxopen/themis/pdp-control && ls $GOPATH/src/github.com/infobloxopen/themis/pdp-control"

import (
	"time"

	"golang.org/x/net/context"
	"google.golang.org/grpc"

	pb "github.com/infobloxopen/themis/pdp-control"
)

type Host struct {
	address string
	chunk   int

	conn   *grpc.ClientConn
	client *pb.PDPControlClient

	ready  bool
	policy int32
}

func NewHost(addr string, chunk int) *Host {
	return &Host{addr, chunk, nil, nil, false, -1}
}

func (h *Host) connect(timeout time.Duration, log Logger) error {
	log.Infof("Connecting to %s...", h.address)
	conn, err := grpc.Dial(h.address, grpc.WithInsecure(), grpc.WithBlock(), grpc.WithTimeout(timeout))
	if err != nil {
		log.Errorf("Can't connect to %s: %v. Skipping...", h.address, err)
		return err
	}

	h.conn = conn

	client := pb.NewPDPControlClient(h.conn)
	h.client = &client
	return nil
}

func (h *Host) close() {
	if h.conn != nil {
		h.conn.Close()
	}

	h.client = nil
}

func (h *Host) uploadIncludes(includes map[string][]byte, log Logger) []int32 {
	if h.client == nil {
		return nil
	}

	IDs := []int32{}
	for ID, b := range includes {
		log.Infof("Uploading %q content to %s...", ID, h.address)

		stream, err := (*h.client).Upload(context.Background())
		if err != nil {
			log.Errorf("Can't start content uploading: %v. Skipping the host...", err)
			return nil
		}

		for offset := 0; offset < len(b); offset += h.chunk {
			var c []byte
			if len(b) < offset+h.chunk {
				c = b[offset:]
			} else {
				c = b[offset : offset+h.chunk]
			}

			if err := stream.Send(&pb.Chunk{string(c)}); err != nil {
				log.Errorf("Can't upload %d bytes starting %d: %v. Skipping the host...", len(c), offset, err)
				return nil
			}
		}

		r, err := stream.CloseAndRecv()
		if err != nil {
			log.Errorf("Can't finish content uploading: %v. Skipping the host...", err)
			return nil
		}

		if r.Status != pb.Response_ACK {
			log.Errorf("Error while loading content to the host: %s. Skipping the host...", r.Details)
			return nil
		}

		item := pb.Item{pb.Item_CONTENT, r.Id, ID, nil}
		r, err = (*h.client).Parse(context.Background(), &item)
		if err != nil {
			log.Errorf("Can't parse uploaded content: %v. Skipping the host...", err)
			return nil
		}

		if r.Status != pb.Response_ACK {
			log.Errorf("Error while parsing uploaded content: %s. Skipping the host...", r.Details)
			return nil
		}

		IDs = append(IDs, r.Id)
	}

	return IDs
}

func (h *Host) uploadPolicy(policy []byte, IDs []int32, log Logger) {
	if IDs == nil {
		return
	}

	log.Infof("Uploading policy to %s...", h.address)

	stream, err := (*h.client).Upload(context.Background())
	if err != nil {
		log.Errorf("Can't start policy uploading: %v. Skipping the host...", err)
		return
	}

	for offset := 0; offset < len(policy); offset += h.chunk {
		var c []byte
		if len(policy) < offset+h.chunk {
			c = policy[offset:]
		} else {
			c = policy[offset : offset+h.chunk]
		}

		if err := stream.Send(&pb.Chunk{string(c)}); err != nil {
			log.Errorf("Can't upload %d bytes starting %d: %v. Skipping the host...", len(c), offset, err)
			return
		}
	}

	r, err := stream.CloseAndRecv()
	if err != nil {
		log.Errorf("Can't finish policy uploading: %v. Skipping the host...", err)
		return
	}

	if r.Status != pb.Response_ACK {
		log.Errorf("Error while loading policy to the host: %s. Skipping the host...", r.Details)
		return
	}

	item := pb.Item{pb.Item_POLICIES, r.Id, "", IDs}
	r, err = (*h.client).Parse(context.Background(), &item)
	if err != nil {
		log.Errorf("Can't parse policy: %v. Skipping host...", err)
		return
	}

	if r.Status != pb.Response_ACK {
		log.Errorf("Error while parsing policy at the host %s. Skipping the host...", r.Details)
		return
	}

	h.policy = r.Id
	h.ready = true
}

func (h *Host) upload(includes map[string][]byte, policy []byte, log Logger) {
	h.uploadPolicy(policy, h.uploadIncludes(includes, log), log)
}

func (h *Host) apply(log Logger) {
	if !h.ready {
		return
	}

	log.Infof("Applying policy to host %s...", h.address)

	update := pb.Update{h.policy}
	r, err := (*h.client).Apply(context.Background(), &update)
	if err != nil {
		log.Errorf("Can't apply policy: %v", err)
		return
	}

	if r.Status != pb.Response_ACK {
		log.Errorf("Error while applying policy to the host %s", r.Details)
		return
	}

	log.Infof("Policy has been applied.")
}
