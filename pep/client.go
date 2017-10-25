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
	ready  atomic.Value // bool
	client *gorpc.Client
	batch  *gorpc.Batch
	limit  uint
	count  uint
	sync.Mutex
}

func (r *rpc) setReady() {
	r.ready.Store(true)
}

func (r *rpc) isReady() bool {
	return r.ready.Load().(bool)
}

func (r *rpc) addToBatch(request *pdp.Request) *gorpc.BatchResult {
	r.Lock()
	defer r.Unlock()
	ret := r.batch.Add(request)
	r.count++
	if r.count > r.limit {
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
	endpoints []string
	rpcs      []rpc
	delay     uint
	pending   uint
}

// NewBalancedClient creates client instance bound to several PDP servers with random balancing.
func NewBalancedClient(endpoints []string, delay, pending uint) Client {
	return &client{
		endpoints: endpoints,
		rpcs:      make([]rpc, len(endpoints)),
		delay:     delay,
		pending:   pending,
	}
}

func newOnConnectFunc(r *rpc) gorpc.OnConnectFunc {
	return func(remoteAddr string, rwc io.ReadWriteCloser) (io.ReadWriteCloser, error) {
		log.Printf("[DEBUG] Connected to %s", remoteAddr)
		r.setReady()
		return rwc, nil
	}
}

func (c *client) Connect() error {
	for i, endpoint := range c.endpoints {
		client := &gorpc.Client{Addr: endpoint}
		c.rpcs[i].ready.Store(false)
		client.OnConnect = newOnConnectFunc(&c.rpcs[i])
		client.DisableCompression = true
		client.Start()
		c.rpcs[i].client = client
		c.rpcs[i].limit = c.pending
		c.rpcs[i].batch = client.NewBatch()
		if c.delay > 0 {
			timeout := time.After(time.Duration(c.delay) * time.Microsecond)
			go func(i int) {
				for {
					select {
					case <-timeout:
						c.rpcs[i].callBatch()
						timeout = time.After(time.Duration(c.delay) * time.Microsecond)
					}
				}
			}(i)
		}
	}
	return nil
}

func (c *client) Close() {
	for i := 0; i < len(c.rpcs); i++ {
		c.rpcs[i].client.Stop()
	}
}

func (c *client) Validate(request *pdp.Request) (*pdp.Response, error) {
	var index []int
	for i, r := range c.rpcs {
		if r.isReady() {
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
	if c.delay > 0 {
		batch := c.rpcs[i].addToBatch(request)
		<-batch.Done
		return batch.Response.(*pdp.Response), batch.Error
	} else {
		ret, err := c.rpcs[i].client.Call(request)
		return ret.(*pdp.Response), err
	}
}
