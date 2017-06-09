package main

import (
	"github.com/infobloxopen/themis/pdp"
)

func (s *Server) makePatchedPolicies(data []byte, content map[string]interface{}) (pdp.EvaluableType, error) {
	return s.Policy, nil
}

func (s *Server) makePatchedContent(data []byte, id string) (interface{}, error) {
	return s.Includes, nil
}
