package server

import (
	"math/rand"

	"github.com/infobloxopen/themis/pep"
)

func (s *Server) getShard() pep.Client {
	if len(s.shardClients) > 0 {
		if i := rand.Intn(len(s.shardClients) + 1); i < len(s.shardClients) {
			return s.shardClients[i]
		}
	}

	return nil
}
