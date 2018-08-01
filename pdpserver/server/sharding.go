package server

import (
	"time"

	"github.com/infobloxopen/themis/pdp"
	"github.com/infobloxopen/themis/pep"
	"github.com/infobloxopen/themis/pipclient"

	log "github.com/sirupsen/logrus"
)

func (s *Server) getShard(name string) pep.Client {
	if c, ok := s.shardClients[name]; ok {
		return c
	}

	return nil
}

func (s *Server) updateShardClients() {
	for name, c := range s.shardClients {
		s.shardClients[name] = nil
		c.Close()
	}

	m := s.p.GetShards().Map()
	if len(m) > 0 {
		s.shardClients = make(map[string]pep.Client, len(m))
		for name, addrs := range m {
			opts := []pep.Option{
				pep.WithCustomData(name),
				pep.WithRoundRobinBalancer(addrs...),
				pep.WithConnectionTimeout(5 * time.Second),
			}
			if s.opts.shardingStreams > 0 {
				opts = append(opts,
					pep.WithStreams(s.opts.shardingStreams),
				)
			}
			s.shardClients[name] = pep.NewClient(opts...)
		}
	}
}

func (s *Server) updatePIPShardClients() pdp.Router {
	for name, c := range s.pipShardClients {
		s.pipShardClients[name] = nil
		c.Close()
	}

	m := s.c.GetShards().Map()
	if len(m) > 0 {
		s.pipShardClients = make(map[string]pipclient.Client, len(m))
		for name, addrs := range m {
			s.opts.logger.WithField("name", name).Debug("Creating content sharding client")
			opts := []pipclient.Option{
				pipclient.WithCustomData(name),
				pipclient.WithRoundRobinBalancer(addrs...),
				pipclient.WithConnectionTimeout(5 * time.Second),
			}
			if s.opts.shardingStreams > 0 {
				opts = append(opts,
					pipclient.WithStreams(s.opts.shardingStreams),
				)
			}
			s.pipShardClients[name] = pipclient.NewClient(opts...)
		}
	}

	return s.newLocalContentRouter()
}

type localContentRouter struct {
	c      map[string]pipclient.Client
	logger *log.Logger
}

func (s *Server) newLocalContentRouter() pdp.Router {
	return &localContentRouter{
		c:      s.pipShardClients,
		logger: s.opts.logger,
	}
}

func (r *localContentRouter) Call(err *pdp.ContentShardingError) (pdp.AttributeValue, error) {
	name := err.Shard
	key := err.Key
	if r.logger.Level >= log.DebugLevel {
		r.logger.WithFields(log.Fields{
			"name": name,
			"key":  key,
		}).Debug("Content sharding redirect")
	}

	if sc, ok := r.c[err.Shard]; ok {
		var res pdp.Response
		if err := sc.Map([]pdp.AttributeAssignment{
			pdp.MakeStringAssignment("key", err.Key),
		}, &res); err != nil {
			if r.logger.Level >= log.DebugLevel {
				r.logger.WithFields(log.Fields{
					"name": name,
					"key":  key,
					"err":  err,
				}).Debug("Content sharding redirect failure")
			}

			return pdp.UndefinedValue, err
		}

		if res.Status != nil {
			if r.logger.Level >= log.DebugLevel {
				r.logger.WithFields(log.Fields{
					"name":   name,
					"key":    key,
					"status": res.Status,
				}).Debug("Content sharding redirect failure")
			}

			return pdp.UndefinedValue, err
		}

		if len(res.Obligations) != 1 {
			if r.logger.Level >= log.DebugLevel {
				r.logger.WithFields(log.Fields{
					"name":        name,
					"key":         key,
					"obligations": len(res.Obligations),
				}).Debug("Invalid response. Expected exactly one value in content sharding response")
			}

			return pdp.UndefinedValue, err
		}

		v, err := res.Obligations[0].GetValue()
		if err != nil {
			if r.logger.Level >= log.DebugLevel {
				r.logger.WithFields(log.Fields{
					"name": name,
					"key":  key,
					"err":  err,
				}).Debug("Invalid expression in content sharding response")
			}

			return pdp.UndefinedValue, err
		}

		return v, err

	}

	if r.logger.Level >= log.DebugLevel {
		r.logger.WithFields(log.Fields{
			"name": name,
			"key":  key,
		}).Debug("Shard not found")
	}

	return pdp.UndefinedValue, newMissingContentShardError(name)
}
