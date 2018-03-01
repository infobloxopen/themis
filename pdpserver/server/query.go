package server

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/julienschmidt/httprouter"
)

const (
	storageQueryCommands = `PDP Storage Query API

GET /storage/exists/<ruleId> [string of ruleId]
GET /storage/rules/<policyId> [string of policyId] 
GET /storage/rules/<policyId>/<limit> [string of policyId] [0 .. n # of rules to show]
GET /storage/policies
`
	maximumRetries = 100
	maxUint32      = 2147483647
)

func queryIndexHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	fmt.Fprint(w, storageQueryCommands)
}

func (s *Server) listenQuery() error {
	if len(s.opts.queryEP) <= 0 {
		return nil
	}
	storageQueryRouter := httprouter.New()
	storageQueryRouter.GET("/", queryIndexHandler)
	storageQueryRouter.GET("/storage/exists/:ruleId",
		func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
			ruleID := ps.ByName("ruleId")
			if s.p == nil {
				fmt.Fprint(w, "policy storage doesn't exist yet")
			} else {
				_, err := s.p.GetRule(ruleID)
				if err == nil {
					fmt.Fprintf(w, "rule %s exists", ruleID)
				} else {
					fmt.Fprint(w, err)
				}
			}
		})
	storageQueryRouter.GET("/storage/rules/:policyId",
		func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
			policyID := ps.ByName("policyId")
			ruleStr, err := s.getPolicyRules(policyID, maxUint32)
			if err != nil {
				fmt.Fprint(w, err)
			} else {
				fmt.Fprint(w, ruleStr)
			}
		})
	storageQueryRouter.GET("/storage/rules/:policyId/:limit",
		func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
			policyID := ps.ByName("policyId")
			limitStr := ps.ByName("limit")
			limit, err := strconv.ParseUint(limitStr, 10, 32)
			if err != nil {
				fmt.Fprintf(w, "invalid limit %s", limitStr)
			} else {
				ruleStr, err := s.getPolicyRules(policyID, uint(limit))
				if err != nil {
					fmt.Fprint(w, err)
				} else {
					fmt.Fprint(w, ruleStr)
				}
			}
		})
	storageQueryRouter.GET("/storage/policies",
		func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
			if s.p == nil {
				fmt.Fprint(w, "policy storage doesn't exist yet")
			} else {
				policies := s.p.GetAllPolicies()
				pubIdx := 0
				policyIDs := make([]string, len(policies))
				for _, policy := range policies {
					if policyID, ok := policy.GetID(); ok {
						policyIDs[pubIdx] = strconv.Quote(policyID)
						pubIdx++
					}
				}
				fmt.Fprint(w, strings.Join(policyIDs[:pubIdx], ", "))
			}
		})
	var err error
	go func() {
		err = http.ListenAndServe(s.opts.queryEP, storageQueryRouter)
	}()
	if err := checkEP(s.opts.queryEP, func() error { return err }); err != nil {
		return fmt.Errorf("Serving storage query failed: %s", err)
	}
	return nil
}

func checkEP(endpoint string, check func() error) error {
	for retries := 0; retries < maximumRetries; retries-- {
		time.Sleep(100 * time.Millisecond)
		if err := check(); err != nil {
			return err
		}
		response, err := http.Get(endpoint)
		if err == nil {
			if response.StatusCode != 200 {
				response.Body.Close()
				return fmt.Errorf("Debug http API returns %d", response.StatusCode)
			}
			response.Body.Close()
			return nil
		}
		if response != nil {
			response.Body.Close()
		}
	}
	return fmt.Errorf("Cannot reach endpoint %s", endpoint)
}

func (s *Server) getPolicyRules(policyID string, limit uint) (string, error) {
	if s.p == nil {
		return "policy storage doesn't exist yet", nil
	}
	policy, err := s.p.GetPolicy(policyID)
	if err != nil {
		return "", err
	}
	rules, hidden := policy.GetRules()
	if hidden {
		return "<hidden policy>", nil
	}
	var (
		iRule   int
		ruleStr string
		pubIdx  uint
		nRules  = len(rules)
		ruleIDs = make([]string, limit)
	)

	// find the first limit-1 visible rules
	for iRule = 0; iRule < nRules && pubIdx < limit-1; iRule++ {
		if ruleID, ok := rules[iRule].GetID(); ok {
			ruleIDs[pubIdx] = strconv.Quote(ruleID)
			pubIdx++
		}
	}
	// look for the last visible ruleID
	for j := nRules - 1; j > iRule && pubIdx < limit; j++ {
		if ruleID, ok := rules[j].GetID(); ok {
			ruleIDs[pubIdx] = strconv.Quote(ruleID)
			pubIdx++
		}
	}
	// assert pubIdx <= limit
	if pubIdx == limit {
		ruleStr = strings.Join(ruleIDs[:limit-1], ", ") +
			", ..., " + ruleIDs[limit-1]
	} else {
		ruleStr = strings.Join(ruleIDs[:pubIdx], ", ")
	}
	return ruleStr, nil
}
