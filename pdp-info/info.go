package info

import (
	"fmt"
	"math/rand"
	"strings"
	"sync"
	"time"

	log "github.com/Sirupsen/logrus"
)

type streamObj struct {
	id     int
	stream PipService_GetRequestsStreamServer
	err    chan error
}

type streamsPool struct {
	uid     int
	streams map[string][]streamObj
	sync.RWMutex
}

func newStreamsPool() *streamsPool {
	return &streamsPool{
		streams: make(map[string][]streamObj),
	}
}

func (p *streamsPool) RegisterStream(stream PipService_GetRequestsStreamServer) chan error {
	cerr := make(chan error, 1)

	resp, err := stream.Recv()
	if err != nil {
		cerr <- err
		return cerr
	}

	inAttrs := resp.GetAttributes()
	if len(inAttrs) != 1 {
		cerr <- fmt.Errorf("Expected exactly one attribute, got: %d", len(inAttrs))
		return cerr
	}

	name := inAttrs[0].GetValue()
	s := streamObj{stream: stream, err: cerr}

	log.Infof("New PIP [%s] stream registered", name)

	p.Lock()
	defer p.Unlock()

	s.id = p.uid
	if _, ok := p.streams[name]; !ok {
		p.streams[name] = []streamObj{s}
	} else {
		p.streams[name] = append(p.streams[name], s)
	}
	p.uid++

	return cerr
}

var random = rand.New(rand.NewSource(time.Now().UnixNano()))

func (p *streamsPool) GetStream(name string) (stream PipService_GetRequestsStreamServer,
	id int, err error) {

	p.RLock()
	defer p.RUnlock()

	list := p.streams[name]
	if len(list) == 0 {
		err = fmt.Errorf("PIP [%s] streams not registered", name)
		log.Errorf(err.Error())
		return
	}

	index := random.Intn(len(list))
	id = list[index].id
	stream = list[index].stream

	return
}

func (p *streamsPool) DeleteStream(name string, id int) {
	p.Lock()
	defer p.Unlock()

	if list, ok := p.streams[name]; ok {
		for i, s := range list {
			if s.id == id {
				err := fmt.Errorf("PIP [%s] stream deleted", name)
				p.streams[name][i].err <- err
				p.streams[name] = append(list[:i], list[i+1:]...)
				log.Errorf(err.Error())
				return
			}
		}
	}
}

type PipService interface {
	Get(uri string, attrs []*Attribute) ([]*Attribute, error)
	PipServiceServer
}

type pipService struct {
	pool *streamsPool
}

func (p *pipService) Get(uri string, attrs []*Attribute) ([]*Attribute, error) {
	var service, query string
	if i := strings.Index(uri, "/"); i >= 0 {
		service = uri[:i]
		query = uri[i+1:]
	} else {
		service = uri
	}

	var streamErr error
	for {
		stream, id, err := p.pool.GetStream(service)
		if err != nil {
			streamErr = err
			break
		}

		req := &Request{Query: query, Attributes: attrs}
		if err := stream.Send(req); err != nil {
			p.pool.DeleteStream(service, id)
			continue
		}

		resp, err := stream.Recv()
		if err != nil {
			p.pool.DeleteStream(service, id)
			continue
		}

		return resp.GetAttributes(), nil
	}
	return nil, streamErr
}

func (p *pipService) GetRequestsStream(stream PipService_GetRequestsStreamServer) error {
	return <-p.pool.RegisterStream(stream)
}

var psOnce sync.Once
var psInstance *pipService

func GetPipService() PipService {
	psOnce.Do(func() {
		psInstance = &pipService{pool: newStreamsPool()}
	})
	return psInstance
}
