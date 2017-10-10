package pep

import (
	"math/rand"
	"sync"
	"time"

	"github.com/valyala/gorpc"
)

// Effect field values
const (
	DENY            = 0
	PERMIT          = 1
	NOTAPPLICABLE   = 2
	INDETERMINATE   = 3
	INDETERMINATED  = 4
	INDETERMINATEP  = 5
	INDETERMINATEDP = 6
)

func EffectName(effect byte) string {
	switch effect {
	case 0:
		return "DENY"
	case 1:
		return "PERMIT"
	case 2:
		return "NOTAPPLICABLE"
	case 3:
		return "INDETERMINATE"
	case 4:
		return "INDETERMINATED"
	case 5:
		return "INDETERMINATEP"
	case 6:
		return "INDETERMINATEDP"
	}
	return "INVALID EFFECT"
}

type Attribute struct {
	Id    string
	Type  string
	Value string
}

type Request struct {
	Attributes []*Attribute
}

type Response struct {
	Effect      byte
	Reason      string
	Obligations []*Attribute
}

const (
	batchInterval = 100 * time.Microsecond
)

type Client interface {
	// Connect establishes connection to PDP server.
	Connect() error
	// Close terminates previously established connection if any.
	// Close should silently return if connection hasn't been established yet or
	// if it has been already closed.
	Close()
	// Validate sends decision request to PDP server and fills out response.
	Validate(request *Request) (*Response, error)
}

type rpc struct {
	client *gorpc.Client
	batch  *gorpc.Batch
	sync.RWMutex
}

type client struct {
	endpoints []string
	rpcs      []rpc
}

// NewBalancedClient creates client instance bound to several PDP servers with random balancing.
func NewBalancedClient(endpoints []string) Client {
	gorpc.RegisterType(&Request{})
	gorpc.RegisterType(&Response{})
	return &client{
		endpoints: endpoints,
		rpcs:      make([]rpc, len(endpoints)),
	}
}

func (c *client) Connect() error {
	for i, endpoint := range c.endpoints {
		client := &gorpc.Client{Addr: endpoint}
		client.Start()
		c.rpcs[i].client = client
		c.rpcs[i].batch = client.NewBatch()
		ticker := time.NewTicker(batchInterval)
		go func(i int) {
			for range ticker.C {
				c.rpcs[i].Lock()
				batch := c.rpcs[i].batch
				c.rpcs[i].batch = client.NewBatch()
				c.rpcs[i].Unlock()
				go batch.Call()
			}
		}(i)
	}
	return nil
}

func (c *client) Close() {
	for i := 0; i < len(c.rpcs); i++ {
		c.rpcs[i].client.Stop()
	}
}

func (c *client) addCall(request *Request) *gorpc.BatchResult {
	l := len(c.rpcs)
	i := 0
	if l > 1 {
		r := rand.New(rand.NewSource(time.Now().UnixNano()))
		i = r.Intn(l)
	}
	c.rpcs[i].RLock()
	defer c.rpcs[i].RUnlock()
	return c.rpcs[i].batch.Add(request)
}

func (c *client) Validate(request *Request) (*Response, error) {
	batch := c.addCall(request)
	<-batch.Done
	return batch.Response.(*Response), batch.Error
}
