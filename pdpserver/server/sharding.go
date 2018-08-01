package server

import (
	"fmt"
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
