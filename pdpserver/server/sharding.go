package server

import "github.com/infobloxopen/themis/pep"

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
			s.shardClients[name] = pep.NewClient(
				pep.WithCustomData(name),
				pep.WithRoundRobinBalancer(addrs...),
			)
		}
	}
}
