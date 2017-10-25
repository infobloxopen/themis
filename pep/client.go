package pep

import (
	"errors"
	"io"
	"log"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"

	pdp "github.com/infobloxopen/themis/pdp-service"
	"github.com/valyala/gorpc"
)

var (
	errPdpNotReady = errors.New("no PDP server in ready state")
)

type atomBool struct{ flag int32 }

func (b *atomBool) Set(value bool) {
	var i int32 = 0
	if value {
		i = 1
	}
	atomic.StoreInt32(&(b.flag), int32(i))
}

func (b *atomBool) Get() bool {
	if atomic.LoadInt32(&(b.flag)) != 0 {
		return true
	}
	return false
}

type Client interface {
	// Connect establishes connection to PDP server.
	Connect() error
	// Close terminates previously established connection if any.
	// Close should silently return if connection hasn't been established yet or
	// if it has been already closed.
	Close()
	// Validate sends decision request to PDP server and fills out response.
	Validate(request *pdp.Request) (*pdp.Response, error)
}

type rpc struct {
	ready    *atomBool
	endpoint string
	client   *gorpc.Client
	batch    *gorpc.Batch
	delay    time.Duration
	pending  uint
	count    uint
	sync.Mutex
}

func newRpc(endpoint string, delay, pending uint) *rpc {
	return &rpc{
		endpoint: endpoint,
		ready:    new(atomBool),
		delay:    time.Duration(delay) * time.Microsecond,
		pending:  pending,
	}
}

func (r *rpc) addToBatch(request *pdp.Request) *gorpc.BatchResult {
	r.Lock()
	defer r.Unlock()
	ret := r.batch.Add(request)
	r.count++
	if r.count > r.pending {
		go r.batch.Call()
		r.batch = r.client.NewBatch()
		r.count = 0
	}
	return ret
}

func (r *rpc) callBatch() {
	r.Lock()
	defer r.Unlock()
	if r.count > 0 {
		go r.batch.Call()
		r.batch = r.client.NewBatch()
		r.count = 0
	}
}

type client struct {
	rpcs []*rpc
}

// NewBalancedClient creates client instance bound to several PDP servers with random balancing.
func NewBalancedClient(endpoints []string, delay, pending uint) Client {
	rpcs := make([]*rpc, len(endpoints))
	for i, endpoint := range endpoints {
		rpcs[i] = newRpc(endpoint, delay, pending)
	}
	return &client{rpcs}
}

func (r *rpc) newOnConnectFunc() gorpc.OnConnectFunc {
	return func(remoteAddr string, rwc io.ReadWriteCloser) (io.ReadWriteCloser, error) {
		log.Printf("[DEBUG] Connected to %s", remoteAddr)
		r.ready.Set(true)
		return rwc, nil
	}
}

func (c *client) Connect() error {
	for _, r := range c.rpcs {
		r.client = &gorpc.Client{
			Addr:               r.endpoint,
			OnConnect:          r.newOnConnectFunc(),
			DisableCompression: true,
		}
		r.client.Start()
		if r.delay > 0 {
			go func(r *rpc) {
				r.batch = r.client.NewBatch()
				timeout := time.After(r.delay)
				for {
					select {
					case <-timeout:
						r.callBatch()
						timeout = time.After(r.delay)
					}
				}
			}(r)
		}
	}
	return nil
}

func (c *client) Close() {
	for _, r := range c.rpcs {
		r.client.Stop()
	}
}

func (c *client) Validate(request *pdp.Request) (*pdp.Response, error) {
	var index []int
	for i, r := range c.rpcs {
		if r.ready.Get() {
			index = append(index, i)
		}
	}
	l := len(index)
	if l == 0 {
		return nil, errPdpNotReady
	}
	i := 0
	if l > 1 {
		r := rand.New(rand.NewSource(time.Now().UnixNano()))
		i = index[r.Intn(l)]
	}
	r := c.rpcs[i]
	if r.delay > 0 {
		batch := r.addToBatch(request)
		<-batch.Done
		if batch.Error != nil {
			return nil, batch.Error
		}
		return batch.Response.(*pdp.Response), nil
	} else {
		ret, err := r.client.Call(request)
		if err != nil {
			return nil, err
		}
		return ret.(*pdp.Response), nil
	}
}
