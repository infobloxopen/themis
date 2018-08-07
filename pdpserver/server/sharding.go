package server

import (
	"fmt"
	"time"

	"github.com/infobloxopen/themis/pdp"
	"github.com/infobloxopen/themis/pep"
	"github.com/infobloxopen/themis/pipclient"

	log "github.com/sirupsen/logrus"
)

type localContentRouter struct {
	c      map[string]pipclient.Client
	logger *log.Logger
}

func (s *Server) newLocalContentRouter() pdp.Router {
	return &localContentRouter{
		c:      s.pipShardClients.getClientMap(),
		logger: s.opts.logger,
	}
}

func (r *localContentRouter) Call(sErr *pdp.ContentShardingError) (pdp.AttributeValue, error) {
	if r.logger.Level >= log.DebugLevel {
		sKey, sKeyErr := sErr.Key.Serialize()
		if sKeyErr != nil {
			return pdp.UndefinedValue, fmt.Errorf("can't serialize key for debug message: %s", sKeyErr)
		}

		r.logger.WithFields(log.Fields{
			"name":    sErr.Shard,
			"content": sErr.Content,
			"item":    sErr.Item,
			"key":     sKey,
		}).Debug("Content sharding redirect")
	}

	if sc, ok := r.c[sErr.Shard]; ok {
		var res pdp.Response
		if err := sc.Map([]pdp.AttributeAssignment{
			pdp.MakeStringAssignment("content", sErr.Content),
			pdp.MakeStringAssignment("item", sErr.Item),
			pdp.MakeExpressionAssignment("key", sErr.Key),
		}, &res); err != nil {
			if r.logger.Level >= log.DebugLevel {
				sKey, sKeyErr := sErr.Key.Serialize()
				if sKeyErr != nil {
					return pdp.UndefinedValue, fmt.Errorf("can't serialize key for debug message: %s", sKeyErr)
				}

				r.logger.WithFields(log.Fields{
					"name":    sErr.Shard,
					"content": sErr.Content,
					"item":    sErr.Item,
					"key":     sKey,
					"err":     err,
				}).Debug("Content sharding redirect failure")
			}

			return pdp.UndefinedValue, err
		}

		if res.Status != nil {
			if r.logger.Level >= log.DebugLevel {
				sKey, sKeyErr := sErr.Key.Serialize()
				if sKeyErr != nil {
					return pdp.UndefinedValue, fmt.Errorf("can't serialize key for debug message: %s", sKeyErr)
				}

				r.logger.WithFields(log.Fields{
					"name":    sErr.Shard,
					"content": sErr.Content,
					"item":    sErr.Item,
					"key":     sKey,
					"status":  res.Status,
				}).Debug("Content sharding redirect failure")
			}

			return pdp.UndefinedValue, res.Status
		}

		if len(res.Obligations) != 1 {
			if r.logger.Level >= log.DebugLevel {
				sKey, sKeyErr := sErr.Key.Serialize()
				if sKeyErr != nil {
					return pdp.UndefinedValue, fmt.Errorf("can't serialize key for debug message: %s", sKeyErr)
				}

				r.logger.WithFields(log.Fields{
					"name":        sErr.Shard,
					"content":     sErr.Content,
					"item":        sErr.Item,
					"key":         sKey,
					"obligations": len(res.Obligations),
				}).Debug("Invalid response. Expected exactly one value in content sharding response")
			}

			return pdp.UndefinedValue, fmt.Errorf("expected exactly one obligation but got %d", len(res.Obligations))
		}

		v, err := res.Obligations[0].GetValue()
		if err != nil {
			if r.logger.Level >= log.DebugLevel {
				sKey, sKeyErr := sErr.Key.Serialize()
				if sKeyErr != nil {
					return pdp.UndefinedValue, fmt.Errorf("can't serialize key for debug message: %s", sKeyErr)
				}

				r.logger.WithFields(log.Fields{
					"name":    sErr.Shard,
					"content": sErr.Content,
					"item":    sErr.Item,
					"key":     sKey,
					"err":     err,
				}).Debug("Invalid expression in content sharding response")
			}

			return pdp.UndefinedValue, err
		}

		return v, nil

	}

	if r.logger.Level >= log.DebugLevel {
		sKey, sKeyErr := sErr.Key.Serialize()
		if sKeyErr != nil {
			return pdp.UndefinedValue, fmt.Errorf("can't serialize key for debug message: %s", sKeyErr)
		}

		r.logger.WithFields(log.Fields{
			"name":    sErr.Shard,
			"content": sErr.Content,
			"item":    sErr.Item,
			"key":     sKey,
		}).Debug("Shard not found")
	}

	return pdp.UndefinedValue, newMissingContentShardError(sErr.Shard)
}

func cmpAddrs(first []string, second []string) bool {
	if len(first) != len(second) {
		return false
	}

	for i, a := range first {
		if a != second[i] {
			return false
		}
	}

	return true
}

type pepClient struct {
	c pep.Client
	s []string
}

func makePepClient(addrs []string, streams int) pepClient {
	opts := []pep.Option{
		pep.WithRoundRobinBalancer(addrs...),
		pep.WithConnectionTimeout(5 * time.Second),
	}

	if streams > 0 {
		opts = append(opts,
			pep.WithStreams(streams),
		)
	}

	return pepClient{
		c: pep.NewClient(opts...),
		s: addrs,
	}
}

type pepClients struct {
	c map[string]pepClient
}

func newPepClients() *pepClients {
	return &pepClients{
		c: make(map[string]pepClient),
	}
}

func (c *pepClients) connect() error {
	for _, cc := range c.c {
		if err := cc.c.Connect(""); err != nil {
			return err
		}
	}

	return nil
}

func (c *pepClients) disconnect() {
	for _, cc := range c.c {
		cc.c.Close()
	}
}

func (c *pepClients) get(name string) pep.Client {
	if cc, ok := c.c[name]; ok {
		return cc.c
	}

	return nil
}

func (c *pepClients) update(m map[string][]string, streams int) (*pepClients, []pep.Client, []pep.Client) {
	out := &pepClients{
		c: make(map[string]pepClient),
	}

	newC := []pep.Client{}
	oldC := []pep.Client{}

	for name, addrs := range m {
		if cc, ok := c.c[name]; ok {
			if cmpAddrs(cc.s, addrs) {
				out.c[name] = cc
				continue
			}

			oldC = append(oldC, cc.c)
		}

		cc := makePepClient(addrs, streams)

		out.c[name] = cc
		newC = append(newC, cc.c)
	}

	for name, cc := range c.c {
		if _, ok := m[name]; !ok {
			oldC = append(oldC, cc.c)
		}
	}

	return out, newC, oldC
}

type pipClient struct {
	c pipclient.Client
	s []string
}

func makePipClient(addrs []string, streams int) pipClient {
	opts := []pipclient.Option{
		pipclient.WithRoundRobinBalancer(addrs...),
		pipclient.WithConnectionTimeout(5 * time.Second),
	}

	if streams > 0 {
		opts = append(opts,
			pipclient.WithStreams(streams),
		)
	}

	return pipClient{
		c: pipclient.NewClient(opts...),
		s: addrs,
	}
}

type pipClients struct {
	c map[string]pipClient
}

func newPipClients() *pipClients {
	return &pipClients{
		c: make(map[string]pipClient),
	}
}

func (c *pipClients) connect() error {
	for _, cc := range c.c {
		if err := cc.c.Connect(""); err != nil {
			return err
		}
	}

	return nil
}

func (c *pipClients) disconnect() {
	for _, cc := range c.c {
		cc.c.Close()
	}
}

func (c *pipClients) getClientMap() map[string]pipclient.Client {
	out := make(map[string]pipclient.Client, len(c.c))
	for name, cc := range c.c {
		out[name] = cc.c
	}

	return out
}

func (c *pipClients) update(m map[string][]string, streams int) (*pipClients, []pipclient.Client, []pipclient.Client) {
	out := &pipClients{
		c: make(map[string]pipClient),
	}

	newC := []pipclient.Client{}
	oldC := []pipclient.Client{}

	for name, addrs := range m {
		if cc, ok := c.c[name]; ok {
			if cmpAddrs(cc.s, addrs) {
				out.c[name] = cc
				continue
			}

			oldC = append(oldC, cc.c)
		}

		cc := makePipClient(addrs, streams)

		out.c[name] = cc
		newC = append(newC, cc.c)
	}

	for name, cc := range c.c {
		if _, ok := m[name]; !ok {
			oldC = append(oldC, cc.c)
		}
	}

	return out, newC, oldC
}
