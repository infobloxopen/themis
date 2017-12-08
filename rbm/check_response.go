package main

import (
	"fmt"
	"strings"

	pb "github.com/infobloxopen/themis/pdp-service"
)

func checkResponse(r *pb.Response) error {
	if r.Effect == pb.Response_PERMIT &&
		r.Reason == "Ok" &&
		len(r.Obligation) == 1 &&
		r.Obligation[0].Id == "c" &&
		strings.ToLower(r.Obligation[0].Type) == "string" &&
		r.Obligation[0].Value == "rule-match" {

		return nil
	}

	attrs := []string{}
	for _, a := range r.Obligation {
		attrs = append(attrs, fmt.Sprintf("%q.(%s): %q", a.Id, a.Type, a.Value))
	}
	var obligations string
	if len(attrs) > 0 {
		obligations = fmt.Sprintf("  obligations:\n    - %s\n", strings.Join(attrs, "\n    - "))
	}

	return fmt.Errorf("Unexpected response:\n"+
		"  effect: %s\n"+
		"  reason: %q\n"+
		"%s\n",
		pb.Response_Effect_name[int32(r.Effect)],
		r.Reason,
		obligations)
}
